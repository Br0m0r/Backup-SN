package middleware

import (
	"net/http"

	"social-network/services/common/observability"
)

// Logging middleware to log HTTP requests
func Logging(next http.Handler) http.Handler {
	return observability.HTTPLogging("auth", next)
}
