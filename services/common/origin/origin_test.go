package origin

import (
	"net/http/httptest"
	"testing"
)

func TestValidatorCheck(t *testing.T) {
	validator, err := Parse("https://social.example.com, http://localhost:8080")
	if err != nil {
		t.Fatalf("parse origins: %v", err)
	}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "configured production origin", origin: "https://social.example.com", want: true},
		{name: "case insensitive host", origin: "https://SOCIAL.EXAMPLE.COM", want: true},
		{name: "configured local origin", origin: "http://localhost:8080", want: true},
		{name: "wrong scheme", origin: "http://social.example.com", want: false},
		{name: "wrong port", origin: "http://localhost:5173", want: false},
		{name: "malformed", origin: "null", want: false},
		{name: "non-browser client", origin: "", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "http://chat.internal/ws", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			if got := validator.Check(request); got != test.want {
				t.Fatalf("Check() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestParseFailsClosed(t *testing.T) {
	for _, value := range []string{"", "*", "https://example.com/path", "javascript://example.com"} {
		if _, err := Parse(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}
