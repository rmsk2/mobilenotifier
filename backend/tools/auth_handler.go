package tools

import (
	"log"
	"net/http"
	"notifier/tools/jwt"
	"time"
)

var ExpectedJwtAudience = "gschmarri"
var ExpectedJwtIssuer string = "daheim_token_issuer"
var TokenTtl int64 = 3600

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
	return JwtAuthenticator(authSecret, logger, originalHandler, jwt.NewHs256JwtVerifier)
}

func JwtHs384Authenticator(authSecret AuthSecret, logger *log.Logger, originalHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return JwtAuthenticator(authSecret, logger, originalHandler, jwt.NewHs384JwtVerifier)
}

func JwtEs256Authenticator(authSecret AuthSecret, logger *log.Logger, originalHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return JwtAuthenticator(authSecret, logger, originalHandler, jwt.NewEs256JwtVerifier)
}

func JwtEs384Authenticator(authSecret AuthSecret, logger *log.Logger, originalHandler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return JwtAuthenticator(authSecret, logger, originalHandler, jwt.NewEs384JwtVerifier)
}

func JwtAuthenticator(authSecret AuthSecret, logger *log.Logger, originalHandler func(http.ResponseWriter, *http.Request), gen func([]byte) *jwt.JwtVerifier) func(http.ResponseWriter, *http.Request) {
	jwtVerifier := gen([]byte(authSecret.Secret))

	res := func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(authSecret.HeaderName)
		parsedClaims, err := jwtVerifier.VerifyJwt(token)
		if err != nil {
			logger.Printf("Unable to authenticate client: %v", err)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		claims, err := jwt.NewFromVerifiedClaims(parsedClaims)
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

		tokenAge := time.Now().UTC().Unix() - claims.IssuedAt
		if tokenAge < 0 {
			logger.Printf("Unable to authenticate client. Token for '%s' issued in the future: '%d' ", claims.Issuer, claims.IssuedAt)
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		if tokenAge > TokenTtl {
			logger.Printf("Unable to authenticate client. Token for '%s' is too old: '%d' seconds", claims.Issuer, tokenAge)
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
