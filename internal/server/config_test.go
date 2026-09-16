package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "server.yaml")
	expected := Config{
		Version:     ConfigVersion,
		Master:      Master{APIURL: "https://panel.example.test", RPCURL: "wss://panel.example.test"},
		Credentials: Credentials{ServerID: "owner.s.edge", Secret: "permanent-secret"},
	}
	if err := WriteConfig(path, expected); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config mode = %o, want 600", info.Mode().Perm())
	}
	actual, err := ReadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("config mismatch: %#v", actual)
	}
}
