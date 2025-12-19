package tools

import (
	"log"
	"net/http"
)

const ExpectedJwtAudience = "gschmarri"

var ExpectedJwtIssuer string = "daheim_token_issuer"

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

func JwtHs256Authenticator(authSecret AuthSecret, logger *log.Logger, originalHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	jwtVerifier := NewHs256Jwt([]byte(authSecret.Secret))

	res := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(authSecret.HeaderName)
		parsedClaims, err := jwtVerifier.VerifyJwt(token)
		if err != nil {
			logger.Printf("Unable to authenticate client: %v", err)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		claims, err := NewFromVerifiedClaims(parsedClaims)
		if err != nil {
			logger.Printf("Unable to authenticate client: %v", err)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		if claims.Audience != ExpectedJwtAudience {
			logger.Printf("Unable to authenticate client. Audience mismatch: '%s'", claims.Audience)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		if claims.Issuer != ExpectedJwtIssuer {
			logger.Printf("Unable to authenticate client. Issuer mismatch: '%s' ", claims.Issuer)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		logger.Printf("User '%s' successfully authenticated", claims.Subject)

		originalHandler(w, r)
	}

	return res
}

func NullAuthenticator(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return f
}

func MakeWrapper(authSecret AuthSecret, logger *log.Logger, authenticator AuthenticatorFunc) AuthWrapperFunc {
	return func(handler func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
		return authenticator(authSecret, logger, handler)
	}
}
