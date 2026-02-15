//go:build !noauth
// +build !noauth

package main

import (
	"notifier/tools"
	"notifier/tools/jwt"
	"os"
)

const envNotifierVerificationSecret = "MN_NOTIFIER_VERIFICATION_SECRET"
const envNotifierUseEcdsa = "MN_NOTIFIER_USE_ECDSA"

func createAuthSecret() *tools.AuthSecret {
	return &tools.AuthSecret{
		Secret:     os.Getenv(envNotifierVerificationSecret),
		HeaderName: authHeaderName,
	}
}

func checkEcdsaPublicKey(raw []byte) error {
	_, err := jwt.LoadEcdsaPublicKey(raw)
	return err
}

func createAuthWrapper() (tools.AuthWrapperFunc, error) {
	authSecret := createAuthSecret()
	//authWrapper := tools.MakeWrapper(*authSecret, createLogger(), tools.ApiKeyAuthenticator), nil

	_, useEcdsa := os.LookupEnv(envNotifierUseEcdsa)
	if useEcdsa {
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtEs256Authenticator), checkEcdsaPublicKey([]byte(authSecret.Secret))
	} else {
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtHs256Authenticator), nil
	}
}
