package serviceauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTokenFromEnvironmentFailsClosed(t *testing.T) {
	t.Setenv(EnvName, "")
	if _, err := TokenFromEnvironment(); err == nil {
		t.Fatal("expected missing token to be rejected")
	}

	t.Setenv(EnvName, strings.Repeat("a", minLength))
	if _, err := TokenFromEnvironment(); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
}

func TestAuthenticate(t *testing.T) {
	expectedToken := strings.Repeat("s", minLength)
	handler := Authenticate(expectedToken, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "missing", wantStatus: http.StatusUnauthorized},
		{name: "incorrect", token: strings.Repeat("x", minLength), wantStatus: http.StatusUnauthorized},
		{name: "valid", token: expectedToken, wantStatus: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/notifications", nil)
			if test.token != "" {
				request.Header.Set(HeaderName, test.token)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
