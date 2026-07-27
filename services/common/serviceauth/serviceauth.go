package serviceauth

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"
)

const (
	HeaderName = "X-Internal-Service-Token"
	EnvName    = "INTERNAL_SERVICE_TOKEN"
	minLength  = 32
)

// TokenFromEnvironment loads the shared interim service credential.
// This is intentionally fail-closed; production should later replace the
// shared token with workload identity or mTLS.
func TokenFromEnvironment() (string, error) {
	token := strings.TrimSpace(os.Getenv(EnvName))
	if len(token) < minLength {
		return "", errors.New("INTERNAL_SERVICE_TOKEN must be configured with at least 32 characters")
	}
	return token, nil
}

// Authenticate permits requests carrying the expected internal credential.
func Authenticate(expectedToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedToken := r.Header.Get(HeaderName)
		if len(providedToken) != len(expectedToken) ||
			subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
