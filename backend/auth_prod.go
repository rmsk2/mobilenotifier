//go:build !noauth
// +build !noauth

package main

import (
	"notifier/tools"
	"os"
)

const envNotifierHmacKey = "NOTIFIER_HMAC_KEY"

func createAuthSecret() *tools.AuthSecret {
	return &tools.AuthSecret{
		Secret:     os.Getenv(envNotifierHmacKey),
		HeaderName: authHeaderName,
	}
}

func createAuthWrapper() tools.AuthWrapperFunc {
	authSecret := createAuthSecret()
	//authWrapper := tools.MakeWrapper(*authSecret, createLogger(), tools.ApiKeyAuthenticator)
	return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtHs256Authenticator)

}
