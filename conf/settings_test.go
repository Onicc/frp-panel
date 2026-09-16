package conf

import "testing"

func TestPublicURLConfiguresAllManagedEndpoints(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PublicURL = "https://panel.example.com:8443/"
	if err := cfg.applyPublicURL(); err != nil {
		t.Fatal(err)
	}
	if cfg.PublicURL != "https://panel.example.com:8443" || cfg.Client.APIUrl != cfg.PublicURL {
		t.Fatalf("unexpected API URL: public=%q client=%q", cfg.PublicURL, cfg.Client.APIUrl)
	}
	if cfg.Client.RPCUrl != "wss://panel.example.com:8443" {
		t.Fatalf("unexpected RPC URL: %q", cfg.Client.RPCUrl)
	}
	if cfg.Master.APIHost != "panel.example.com" || cfg.Master.RPCHost != "panel.example.com" || cfg.Master.APIScheme != "https" {
		t.Fatalf("unexpected master addressing: %#v", cfg.Master)
	}
}

func TestPublicURLRejectsAmbiguousValues(t *testing.T) {
	for _, value := range []string{
		"panel.example.com",
		"ftp://panel.example.com",
		"https://user:password@panel.example.com",
		"https://panel.example.com/admin",
		"https://panel.example.com?debug=true",
	} {
		cfg := DefaultConfig()
		cfg.PublicURL = value
		if err := cfg.applyPublicURL(); err == nil {
			t.Fatalf("invalid PUBLIC_URL accepted: %q", value)
		}
	}
}

func TestHTTPPublicURLUsesWebSocketRPC(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PublicURL = "http://127.0.0.1:19000"
	if err := cfg.applyPublicURL(); err != nil {
		t.Fatal(err)
	}
	if cfg.Client.RPCUrl != "ws://127.0.0.1:19000" {
		t.Fatalf("unexpected local RPC URL: %q", cfg.Client.RPCUrl)
	}
}
