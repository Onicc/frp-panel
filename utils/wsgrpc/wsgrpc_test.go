package wsgrpc

import (
	"net/http/httptest"
	"testing"
)

func TestRequestClientIPPrefersReverseProxyHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header string
		value  string
		want   string
	}{
		{name: "cloudflare", header: "CF-Connecting-IP", value: "118.25.94.27", want: "118.25.94.27"},
		{name: "forwarded chain", header: "X-Forwarded-For", value: "43.134.184.42, 127.0.0.1", want: "43.134.184.42"},
		{name: "forwarded syntax", header: "Forwarded", value: "for=203.0.113.8;proto=wss", want: "203.0.113.8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "https://panel.example.test/wsgrpc", nil)
			request.RemoteAddr = "127.0.0.1:8080"
			request.Header.Set(tt.header, tt.value)
			if got := requestClientIP(request).String(); got != tt.want {
				t.Fatalf("requestClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequestClientIPFallsBackToSocketAddress(t *testing.T) {
	request := httptest.NewRequest("GET", "https://panel.example.test/wsgrpc", nil)
	request.RemoteAddr = "43.134.184.42:443"
	if got := requestClientIP(request).String(); got != "43.134.184.42" {
		t.Fatalf("requestClientIP() = %q, want socket address", got)
	}
}
