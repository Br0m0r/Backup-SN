package middleware

import (
	"net/http"

	"social-network/services/common/observability"
)

// Logging middleware logs all HTTP requests
func Logging(next http.Handler) http.Handler {
	return observability.HTTPLogging("groups", next)
}
