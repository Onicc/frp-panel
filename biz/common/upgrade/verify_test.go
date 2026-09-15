package upgrade

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestChecksumVerification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent")
	content := []byte("verified agent")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	expected := fmt.Sprintf("%x", sha256.Sum256(content))
	if err := verifySHA256(path, expected); err != nil {
		t.Fatal(err)
	}
	if err := verifySHA256(path, fmt.Sprintf("%064d", 0)); err == nil {
		t.Fatal("mismatched checksum accepted")
	}
}

func TestChecksumManifestLookup(t *testing.T) {
	want := fmt.Sprintf("%064d", 1)
	got, err := checksumForAsset([]byte(want+"  frp-panel-agent-darwin-arm64\n"), "frp-panel-agent-darwin-arm64")
	if err != nil || got != want {
		t.Fatalf("got %q, %v", got, err)
	}
}
