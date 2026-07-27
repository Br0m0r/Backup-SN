package middleware

import (
	"net/http"

	"social-network/services/common/observability"
)

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return observability.HTTPLogging("notifications", next)
}
