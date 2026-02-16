//go:build !noauth
// +build !noauth

package main

import (
	"notifier/tools"
	"notifier/tools/jwt"
	"os"
)

const envNotifierVerificationSecret = "MN_VERIFICATION_SECRET"
const envTokenType = "MN_TOKEN_TYPE"

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

	tokenType, ok := os.LookupEnv(envTokenType)
	if !ok {
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtHs256Authenticator), nil
	}

	switch tokenType {
	case jwt.AlgHs384:
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtHs384Authenticator), nil
	case jwt.AlgEs256:
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtEs256Authenticator), checkEcdsaPublicKey([]byte(authSecret.Secret))
	case jwt.AlgEs384:
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtEs384Authenticator), checkEcdsaPublicKey([]byte(authSecret.Secret))
	default:
		return tools.MakeWrapper(*authSecret, createLogger(), tools.JwtHs256Authenticator), nil
	}
}
