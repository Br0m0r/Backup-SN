package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"social-network/services/common/observability"
	"social-network/services/common/serviceauth"
	"social-network/services/gateway/edge"
)

// Targets contains the private upstream addresses used by the edge gateway.
type Targets struct {
	Auth          *url.URL
	Users         *url.URL
	Posts         *url.URL
	Groups        *url.URL
	Chat          *url.URL
	Notifications *url.URL
	Media         *url.URL
	Frontend      *url.URL
}

// TargetsFromEnvironment loads and validates all upstream service URLs.
func TargetsFromEnvironment() (Targets, error) {
	definitions := []struct {
		name         string
		defaultValue string
		destination  **url.URL
	}{
		{name: "AUTH_SERVICE_URL", defaultValue: "http://localhost:8081"},
		{name: "USERS_SERVICE_URL", defaultValue: "http://localhost:8082"},
		{name: "POSTS_SERVICE_URL", defaultValue: "http://localhost:8083"},
		{name: "GROUPS_SERVICE_URL", defaultValue: "http://localhost:8084"},
		{name: "CHAT_SERVICE_URL", defaultValue: "http://localhost:8085"},
		{name: "NOTIFICATIONS_SERVICE_URL", defaultValue: "http://localhost:8086"},
		{name: "MEDIA_SERVICE_URL", defaultValue: "http://localhost:9000"},
		{name: "FRONTEND_SERVICE_URL", defaultValue: "http://localhost:3000"},
	}

	targets := Targets{}
	definitions[0].destination = &targets.Auth
	definitions[1].destination = &targets.Users
	definitions[2].destination = &targets.Posts
	definitions[3].destination = &targets.Groups
	definitions[4].destination = &targets.Chat
	definitions[5].destination = &targets.Notifications
	definitions[6].destination = &targets.Media
	definitions[7].destination = &targets.Frontend

	for _, definition := range definitions {
		rawURL := strings.TrimSpace(os.Getenv(definition.name))
		if rawURL == "" {
			rawURL = definition.defaultValue
		}
		parsedURL, err := url.Parse(rawURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return Targets{}, fmt.Errorf("invalid %s: %q", definition.name, rawURL)
		}
		*definition.destination = parsedURL
	}

	return targets, nil
}

// NewRouter creates the single public HTTP and WebSocket entry point.
func NewRouter(targets Targets, controls *edge.Controls) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "gateway"})
	})

	handleService(mux, "/api/auth", targets.Auth)
	handleService(mux, "/api/users", targets.Users)
	handleService(mux, "/api/posts", targets.Posts)
	handleService(mux, "/api/groups", targets.Groups)
	handleService(mux, "/api/chat", targets.Chat)
	handleService(mux, "/api/notifications", targets.Notifications)
	handleReadOnlyService(mux, "/media", targets.Media)
	mux.Handle("/", newReverseProxy(targets.Frontend))

	return observability.HTTPLogging("gateway", controls.Wrap(stripInternalCredentials(mux)))
}

func handleReadOnlyService(mux *http.ServeMux, prefix string, target *url.URL) {
	proxy := newReverseProxy(target)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		proxy.ServeHTTP(w, r)
	})
	mux.Handle(prefix, handler)
	mux.Handle(prefix+"/", handler)
}

func handleService(mux *http.ServeMux, prefix string, target *url.URL) {
	proxy := newReverseProxy(target)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		proxy.ServeHTTP(w, r)
	})
	mux.Handle(prefix, handler)
	mux.Handle(prefix+"/", handler)
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(request *http.Request) {
		originalDirector(request)
		request.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Upstream service unavailable",
		})
	}
	return proxy
}

func stripInternalCredentials(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(serviceauth.HeaderName)
		next.ServeHTTP(w, r)
	})
}
