package tools

import (
	"log"
	"net/http"
)

type AuthSecret struct {
	Secret     string
	HeaderName string
}

type AuthWrapperFunc func(func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)
type AuthenticatorFunc func(AuthSecret, *log.Logger, func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)

func ApiKeyAuthenticator(authSecret AuthSecret, logger *log.Logger, originalHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(authSecret.HeaderName)
		if apiKey != authSecret.Secret {
			logger.Printf("Unable to authenticate client")
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		originalHandler(w, r)
	}
}

func MakeWrapper(authSecret AuthSecret, logger *log.Logger, authenticator AuthenticatorFunc) AuthWrapperFunc {
	return func(handler func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
		return authenticator(authSecret, logger, handler)
	}
}
