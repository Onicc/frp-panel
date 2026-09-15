package utils

import (
	"net/http/httptest"
	"testing"
)

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed string
		want    bool
	}{
		{"non browser", "", "", true},
		{"same origin", "https://panel.example.com", "", true},
		{"explicit", "https://admin.example.com", "https://admin.example.com", true},
		{"reject foreign", "https://evil.example", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "https://panel.example.com/ws", nil)
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := IsOriginAllowed(r, tt.allowed); got != tt.want {
				t.Fatalf("got %t want %t", got, tt.want)
			}
		})
	}
}
