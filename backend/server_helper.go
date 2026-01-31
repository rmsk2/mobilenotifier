package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const envListenerPort = "MN_PORT_LISTEN"
const envListenLocalHostOnly = "MN_LISTEN_LOCALHOST_ONLY"
const envTlsCertFile = "MN_CERT_FILE"
const envTlsKeyFile = "MN_KEY_FILE"
const DefaultPort = 5100

type SimpleServer struct {
	port          uint16
	localHostOnly bool
}

func newSimpleServer(p uint16, localOnly bool) *SimpleServer {
	res := SimpleServer{
		port:          p,
		localHostOnly: localOnly,
	}

	return &res
}

func (s *SimpleServer) Serve() error {
	addr := fmt.Sprintf(":%d", s.port)
	if s.localHostOnly {
		addr = "localhost" + addr
	}
	return http.ListenAndServe(addr, nil)
}

type TlsServer struct {
	tlsConf  *tls.Config
	server   *http.Server
	certFile string
	keyFile  string
}

func (t *TlsServer) Serve() error {
	return t.server.ListenAndServeTLS(t.certFile, t.keyFile)
}

func newTlsServer(p uint16, crtFile string, keyFile string) (*TlsServer, error) {
	tlsConfig := &tls.Config{
		ClientAuth: tls.NoClientCert,
		MinVersion: tls.VersionTLS12,
	}

	httpServer := &http.Server{
		Addr:      fmt.Sprintf(":%d", p),
		TLSConfig: tlsConfig,
	}

	srv := TlsServer{
		tlsConf:  tlsConfig,
		server:   httpServer,
		certFile: crtFile,
		keyFile:  keyFile,
	}

	return &srv, nil
}

func createWebServer() (WebServer, error) {
	var listenerPort uint16 = DefaultPort
	var certFileName string
	var keyFileName string
	var res WebServer
	var err error
	var localHostOnly bool

	_, localHostOnly = os.LookupEnv(envListenLocalHostOnly)

	temp, ok := os.LookupEnv(envListenerPort)
	if ok {
		p, err := strconv.ParseUint(temp, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("Illegal port number: %v", err)
		}

		listenerPort = uint16(p)
	}

	temp, okCert := os.LookupEnv(envTlsCertFile)
	if okCert {
		certFileName = temp
	}

	temp, okKey := os.LookupEnv(envTlsKeyFile)
	if okKey {
		keyFileName = temp
	}

	if okKey && okCert && (!localHostOnly) {
		res, err = newTlsServer(listenerPort, certFileName, keyFileName)
		if err != nil {
			return nil, fmt.Errorf("Unable to build TLS server object: %v", err)
		}

		log.Println("Using TLS")
	} else {
		res = newSimpleServer(listenerPort, localHostOnly)
		log.Println("Using plain HTTP")
		if localHostOnly {
			log.Println("Only listening on localhost")
		}
	}

	log.Printf("Listening on port %d", listenerPort)

	return res, nil
}
