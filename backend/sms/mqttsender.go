package sms

import (
	"context"
	"fmt"
	"notifier/mqtt"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

type MqttMessageSender struct {
	sender  mqtt.MqttSender
	timeOut time.Duration
}

func NewMqttMessageSender(s mqtt.MqttSender, senderTimeOut time.Duration) *MqttMessageSender {
	return &MqttMessageSender{
		sender:  s,
		timeOut: senderTimeOut,
	}
}

func (m *MqttMessageSender) Send(recipientAddress string, message string) error {
	if !m.sender.IsConnected() {
		return fmt.Errorf("connection to message broker currently unavailable")
	}

	ctxWithTimeout, cancel := context.WithTimeout(m.sender.GetRootCtx(), m.timeOut)
	defer cancel()

	if _, err := m.sender.PublishWithCtx(ctxWithTimeout, &paho.Publish{
		QoS:     1,
		Topic:   recipientAddress,
		Payload: []byte(message),
	}); err != nil {
		return err
	}

	return nil
}

func (m *MqttMessageSender) GetName() string {
	return "MQTT"
}
