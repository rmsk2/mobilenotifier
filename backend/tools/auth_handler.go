package tools

import (
	"log"
	"net/http"
)

type AuthSecret struct {
	Secret     string
	HeaderName string
}

type Wrapper interface {
	Wrap(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)
}

func NewApiKeyProvider(as *AuthSecret, l *log.Logger) *ApiKeyProvider {
	return &ApiKeyProvider{
		authSecret: as,
		logger:     l,
	}
}

type ApiKeyProvider struct {
	authSecret *AuthSecret
	logger     *log.Logger
}

func (a *ApiKeyProvider) Wrap(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(a.authSecret.HeaderName)
		if apiKey != a.authSecret.Secret {
			a.logger.Printf("Unable to authenticate client")
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}

// Alternative usage with a higher portion of syntactic sugar.
//
//	Given the handler function 'handleFunc', the AuthSecret 'as' and a logger you can add authentication functionality to the handler by calling
//
// WithAuthentication(handleFunc).UsingParameters(as, logger)
type WithAuthentication func(http.ResponseWriter, *http.Request)

func (h WithAuthentication) UsingParameters(authSecret AuthSecret, logger *log.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(authSecret.HeaderName)
		if apiKey != authSecret.Secret {
			logger.Printf("Unable to authenticate client")
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		h(w, r)
	}
}
