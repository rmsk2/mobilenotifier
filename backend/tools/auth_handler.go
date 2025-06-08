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

// Alternative usage with a higher portion of syntactic sugar.
//
//	Given the handler function 'handleFunc', the AuthSecret 'as' and a logger you can add authentication functionality to the handler by calling
//
// WithApiKeyAuthentication(handleFunc).UsingParameters(as, logger)
type WithApiKeyAuthentication func(http.ResponseWriter, *http.Request)

func (h WithApiKeyAuthentication) UsingParameters(authSecret AuthSecret, logger *log.Logger) func(http.ResponseWriter, *http.Request) {
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
