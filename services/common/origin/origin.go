package origin

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const EnvName = "WEBSOCKET_ALLOWED_ORIGINS"

// Validator performs exact scheme-and-host checks for browser WebSocket origins.
type Validator struct {
	allowed map[string]struct{}
}

// FromEnvironment loads a comma-separated origin allowlist and fails closed.
func FromEnvironment() (*Validator, error) {
	return Parse(os.Getenv(EnvName))
}

// Parse validates a comma-separated list such as
// "https://social.example.com,http://localhost:8080".
func Parse(raw string) (*Validator, error) {
	allowed := make(map[string]struct{})
	for _, candidate := range strings.Split(raw, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		normalized, err := normalize(candidate)
		if err != nil {
			return nil, err
		}
		allowed[normalized] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, errors.New("WEBSOCKET_ALLOWED_ORIGINS must contain at least one absolute http(s) origin")
	}
	return &Validator{allowed: allowed}, nil
}

// Check permits non-browser clients without Origin and browser clients whose
// exact origin is configured.
func (v *Validator) Check(r *http.Request) bool {
	originHeader := strings.TrimSpace(r.Header.Get("Origin"))
	if originHeader == "" {
		return true
	}
	normalized, err := normalize(originHeader)
	if err != nil {
		return false
	}
	_, ok := v.allowed[normalized]
	return ok
}

func normalize(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("WebSocket origins must be absolute http(s) URLs")
	}
	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("WebSocket origins may contain only scheme and host")
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), nil
}
