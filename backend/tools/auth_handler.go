package tools

import (
	"log"
	"net/http"
)

type AuthHandler interface {
	WithAuthentication(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)
}

func NewApiKeyProvider(s string, h string, l *log.Logger) *ApiKeyProvider {
	return &ApiKeyProvider{
		secret:     s,
		headerName: h,
		logger:     l,
	}
}

type ApiKeyProvider struct {
	secret     string
	headerName string
	logger     *log.Logger
}

func (a *ApiKeyProvider) WithAuthentication(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(a.headerName)
		if apiKey != a.secret {
			a.logger.Printf("Unable to authenticate client")
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}
