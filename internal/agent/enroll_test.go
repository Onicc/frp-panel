package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnroll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/agent/enroll" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["token"] != "01234567890123456789012345678901" {
			t.Fatalf("unexpected body: %#v, %v", body, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"clientId":"owner.c.mac","secret":"permanent-secret"}`))
	}))
	defer server.Close()

	result, err := Enroll(server.URL+"/", "01234567890123456789012345678901", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.ClientID != "owner.c.mac" || result.Secret != "permanent-secret" {
		t.Fatalf("unexpected enrollment: %+v", result)
	}
}

func TestEnrollRejectsInvalidURL(t *testing.T) {
	if _, err := Enroll("file:///tmp/master", "token", false); err == nil {
		t.Fatal("unsafe Master URL accepted")
	}
}
