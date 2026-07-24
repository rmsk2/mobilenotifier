/*
 * Copyright (c) 2024 Contributors to the Eclipse Foundation
 *
 *  All rights reserved. This program and the accompanying materials
 *  are made available under the terms of the Eclipse Public License v2.0
 *  and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 *  and the Eclipse Distribution License is available at
 *    http://www.eclipse.org/org/documents/edl-v10.php.
 *
 *  SPDX-License-Identifier: EPL-2.0 OR BSD-3-Clause
 */
// Package mqtt provides a self contained way to connect to an MQTT broker and to publish messages to the
// broker. Heavily modified from the basics.go autopaho example.
package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"notifier/tools"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

const EnvMqttBrokerUrl = "MN_MQTT_BROKER_URL"
const EnvMqttClientID = "MN_MQTT_CLIENT_ID"
const EnvMqttUser = "MN_MQTT_USER"
const EnvMqttPassword = "MN_MQTT_PASSWORD"
const EnvRootCertFile = tools.EnvAdditionalRootCerts
const EnvSessionExpiry = "MN_MQTT_SESSION_EXPIRY"

// WaitForever is intended to be used when calling StartAndWaitForConnection with the intention to wait forever
const WaitForever = -1

type MqttConfig struct {
	BrokerUrl string
	ClientId  string
	Options   []SenderFuncOption
}

func NewConfigFromEnvironment() (*MqttConfig, error) {
	brokerUrl, ok := os.LookupEnv(EnvMqttBrokerUrl)
	if !ok {
		return nil, fmt.Errorf("environment variable '%s' not set", EnvMqttBrokerUrl)
	}

	clientId, ok := os.LookupEnv(EnvMqttClientID)
	if !ok {
		return nil, fmt.Errorf("environment variable '%s' not set", EnvMqttClientID)
	}

	res := &MqttConfig{
		BrokerUrl: brokerUrl,
		ClientId:  clientId,
		Options:   []SenderFuncOption{},
	}

	// Optional: credentials. Applied when a user name is provided; the password
	// is read alongside and may be empty.
	if user, ok := os.LookupEnv(EnvMqttUser); ok {
		password, _ := os.LookupEnv(EnvMqttPassword)
		res.Options = append(res.Options, WithCredentials(user, password))
	}

	// Optional: custom root certificates for TLS.
	if rootCertFile, ok := os.LookupEnv(EnvRootCertFile); ok {
		res.Options = append(res.Options, WithTlsCustomRoots(rootCertFile))
	}

	// Optional: session expiry interval in seconds.
	if expiryStr, ok := os.LookupEnv(EnvSessionExpiry); ok {
		expiry, err := strconv.ParseUint(expiryStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("illegal value '%s' for %s: %w", expiryStr, EnvSessionExpiry, err)
		}
		res.Options = append(res.Options, WithSessionExpiry(uint32(expiry)))
	}

	return res, nil
}

var ErrNoConnManager = errors.New("no connection manager found; please call Start() first")
var ErrInvalidTimeoutValue = errors.New("a timeout of zero makes no sense")
var ErrNotConnected = errors.New("connection to message broker currently unavailable")

type MqttSender interface {
	PublishWhenConnected(topic string, payload []byte, qos byte, timeout time.Duration) error
}

func getTlsConfig(rootFileName string) (*tls.Config, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	certs, err := os.ReadFile(rootFileName)
	if err != nil {
		return nil, fmt.Errorf("unable to load additional root certs from '%s': %w", rootFileName, err)
	}

	ok := rootCAs.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("no usable certs found in '%s'", rootFileName)
	}

	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	return tlsConfig, nil
}

type Sender struct {
	cliCfg      autopaho.ClientConfig
	rootCtx     context.Context
	cancel      context.CancelFunc
	clientId    string
	isConnected atomic.Bool
	connManager atomic.Pointer[autopaho.ConnectionManager]
}

type SenderFuncOption func(*Sender) error

func WithTlsCustomRoots(rootFileName string) SenderFuncOption {
	return func(m *Sender) error {
		tlsConfig, err := getTlsConfig((rootFileName))
		if err != nil {
			return fmt.Errorf("unable to determine custom roots: %w", err)
		}

		m.cliCfg.TlsCfg = tlsConfig

		return nil
	}
}

func WithCredentials(userId string, password string) SenderFuncOption {
	return func(m *Sender) error {
		m.cliCfg.ConnectUsername = userId
		m.cliCfg.ConnectPassword = []byte(password)
		return nil
	}
}

func WithSessionExpiry(interval uint32) SenderFuncOption {
	return func(m *Sender) error {
		m.cliCfg.SessionExpiryInterval = interval
		return nil
	}
}

func WithCleanStart() SenderFuncOption {
	return func(m *Sender) error {
		m.cliCfg.CleanStartOnInitialConnection = true
		return nil
	}
}

func NewSender(brokerUrl string, ctx context.Context, clientId string, options ...SenderFuncOption) (*Sender, error) {
	u, err := url.Parse(brokerUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create Paho client: %w", err)
	}

	cancellableContext, cancel := context.WithCancel(ctx)

	cliCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{u},
		KeepAlive:  20, // Keepalive message should be sent every 20 seconds
		// CleanStartOnInitialConnection defaults to false. Setting this to true will clear the session on the first connection.
		CleanStartOnInitialConnection: false,
		// SessionExpiryInterval - Seconds that a session will survive after disconnection.
		// It is important to set this because otherwise, any queued messages will be lost if the connection drops and
		// the server will not queue messages while it is down. The specific setting will depend upon your needs
		// (60 = 1 minute, 3600 = 1 hour, 86400 = one day, 0xFFFFFFFE = 136 years, 0xFFFFFFFF = don't expire)
		SessionExpiryInterval: 60,
		// eclipse/paho.golang/paho provides base mqtt functionality, the below config will be passed in for each connection
		ClientConfig: paho.ClientConfig{
			// If you are using QOS 1/2, then it's important to specify a client id (which must be unique)
			ClientID: clientId,
		},
	}

	res := &Sender{
		cliCfg:      cliCfg,
		rootCtx:     cancellableContext,
		cancel:      cancel,
		clientId:    clientId,
		isConnected: atomic.Bool{},
		connManager: atomic.Pointer[autopaho.ConnectionManager]{},
	}

	res.cliCfg.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		res.SetConnectionStatus(true)
	}

	res.cliCfg.OnConnectionDown = func() bool {
		res.SetConnectionStatus(false)
		return true
	}

	res.cliCfg.ClientConfig.OnServerDisconnect = func(d *paho.Disconnect) {
		res.SetConnectionStatus(false)
	}

	for _, o := range options {
		err := o(res)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (m *Sender) SetConnectionStatus(val bool) {
	m.isConnected.Store(val)
}

func (m *Sender) IsConnected() bool {
	return m.isConnected.Load()
}

func (m *Sender) Start() error {
	c, err := autopaho.NewConnection(m.rootCtx, m.cliCfg) // starts process; will reconnect until context cancelled
	if err != nil {
		return fmt.Errorf("unable to start Sender: %w", err)
	}

	m.connManager.Store(c)

	return nil
}

func (m *Sender) StartAndWaitForConnection(timeOutInSeconds int) (*autopaho.ConnectionManager, error) {
	err := m.Start()
	if err != nil {
		return nil, err
	}

	var waitCtx context.Context = m.GetRootCtx()
	var cancelFunc context.CancelFunc

	if timeOutInSeconds > 0 {
		waitCtx, cancelFunc = context.WithTimeout(m.GetRootCtx(), time.Duration(timeOutInSeconds)*time.Second)
		defer func() { cancelFunc() }()
	} else {
		if timeOutInSeconds != WaitForever {
			return nil, ErrInvalidTimeoutValue
		}
	}

	cm := m.connManager.Load()

	// Just wait. Ignore result.
	_ = cm.AwaitConnection(waitCtx)

	return cm, nil
}

func (m *Sender) GetRootCtx() context.Context {
	return m.rootCtx
}

func (m *Sender) GetConnManager() *autopaho.ConnectionManager {
	return m.connManager.Load()
}

func (m *Sender) Stop() {
	m.cancel()
}

// PublishWhenConnected sends payload to topic with the given QoS. The publish is
// bound to a context derived from the sender's root context and the given
// timeout, so it is cancelled either when the timeout elapses or when the sender
// is stopped. It returns ErrNotConnected if the broker connection is currently
// down.
func (m *Sender) PublishWhenConnected(topic string, payload []byte, qos byte, timeout time.Duration) error {
	if !m.IsConnected() {
		return ErrNotConnected
	}

	ctx, cancel := context.WithTimeout(m.rootCtx, timeout)
	defer cancel()

	_, err := m.PublishWithCtx(ctx, &paho.Publish{
		QoS:     qos,
		Topic:   topic,
		Payload: payload,
	})

	return err
}

func (m *Sender) PublishWithCtx(ctx context.Context, p *paho.Publish) (*paho.PublishResponse, error) {
	cm := m.connManager.Load()

	if cm == nil {
		return nil, ErrNoConnManager
	}

	return cm.Publish(ctx, p)
}

func (m *Sender) WasAlreadyStopped() bool {
	return m.GetRootCtx().Err() != nil
}

func (m *Sender) WaitForShutdown() {
	cm := m.connManager.Load()

	if cm != nil {
		<-cm.Done()
	}
}

func (m *Sender) StoppedChannel() <-chan struct{} {
	return m.GetRootCtx().Done()
}
