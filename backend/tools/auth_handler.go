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

type ParamUser interface {
	UsingParameters(authSecret AuthSecret, logger *log.Logger) func(http.ResponseWriter, *http.Request)
}

type WrapperConstraint interface {
	~func(http.ResponseWriter, *http.Request)
	ParamUser
}

func NewAuthProvider[T WrapperConstraint](as *AuthSecret, l *log.Logger) *AuthProvider[T] {
	return &AuthProvider[T]{
		authSecret: as,
		logger:     l,
	}
}

type AuthProvider[T WrapperConstraint] struct {
	authSecret *AuthSecret
	logger     *log.Logger
}

func (a *AuthProvider[T]) Wrap(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return T(handler).UsingParameters(*a.authSecret, a.logger)
}

// Create a new type based on func(http.ResponseWriter, *http.Request) and a UsingParameters method for that type to add additional authentication
// methods
type ApiKey func(http.ResponseWriter, *http.Request)

func (h ApiKey) UsingParameters(authSecret AuthSecret, logger *log.Logger) func(http.ResponseWriter, *http.Request) {
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
