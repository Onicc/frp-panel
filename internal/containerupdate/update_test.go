package containerupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Onicc/frp-panel/internal/release"
)

func TestStageVerifiesOfficialAssetAndCommit(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF execution is verified on Linux CI and Docker")
	}
	t.Setenv("FRP_PANEL_UPDATE_DIR", t.TempDir())
	t.Setenv("FRP_PANEL_IMAGE_COMMIT", strings.Repeat("a", 40))
	commit := strings.Repeat("b", 40)
	name := "frp-panel-linux-" + runtime.GOARCH
	temp := t.TempDir()
	source := filepath.Join(temp, "main.go")
	if err := os.WriteFile(source, []byte(fmt.Sprintf("package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"GitCommit: %s\")}\n", commit)), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryDir := filepath.Join(temp, "binary")
	checksumDir := filepath.Join(temp, "checksum")
	for _, dir := range []string{binaryDir, checksumDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(binaryDir, name)
	command := exec.Command("go", "build", "-o", binary, source)
	command.Env = append(os.Environ(), "CGO_ENABLED=0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fixture: %v: %s", err, output)
	}
	raw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	checksums := filepath.Join(checksumDir, "checksums.txt")
	if err := os.WriteFile(checksums, []byte(hex.EncodeToString(sum[:])+"  "+name+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prefix := "https://github.com/Onicc/frp-panel/releases/download/edge/"
	rel := release.Release{TagName: "edge", Commit: commit, Assets: []release.Asset{
		{ID: 1, Name: name, BrowserDownloadURL: prefix + name, Size: int64(len(raw))},
		{ID: 2, Name: "checksums.txt", BrowserDownloadURL: prefix + "checksums.txt", Size: int64(len(sum))},
	}}
	download := func(_ context.Context, url, _ string) (string, error) {
		if url == prefix+name {
			return binary, nil
		}
		if url == prefix+"checksums.txt" {
			return checksums, nil
		}
		return "", fmt.Errorf("unexpected URL %s", url)
	}
	if err := stage(context.Background(), rel, "local-test", download); err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadManifest(PendingPath())
	if err != nil || manifest.Commit != commit || manifest.ImageCommit != strings.Repeat("a", 40) {
		t.Fatalf("unexpected manifest: %+v, %v", manifest, err)
	}
	if _, err := os.Stat(BinaryPath(commit)); err != nil {
		t.Fatalf("staged executable: %v", err)
	}
}

func TestStageRejectsUnofficialAssetURL(t *testing.T) {
	t.Setenv("FRP_PANEL_UPDATE_DIR", t.TempDir())
	name := "frp-panel-linux-" + runtime.GOARCH
	rel := release.Release{TagName: "edge", Commit: strings.Repeat("b", 40), Assets: []release.Asset{
		{ID: 1, Name: name, BrowserDownloadURL: "https://evil.example/" + name},
		{ID: 2, Name: "checksums.txt", BrowserDownloadURL: "https://github.com/Onicc/frp-panel/releases/download/edge/checksums.txt"},
	}}
	called := false
	err := stage(context.Background(), rel, "local-test", func(context.Context, string, string) (string, error) { called = true; return "", nil })
	if err == nil || called {
		t.Fatal("unofficial asset URL reached downloader")
	}
}

func TestStageRejectsChecksumMismatchWithoutPendingUpdate(t *testing.T) {
	t.Setenv("FRP_PANEL_UPDATE_DIR", t.TempDir())
	temp := t.TempDir()
	name := "frp-panel-linux-" + runtime.GOARCH
	binaryDir, checksumDir := filepath.Join(temp, "binary"), filepath.Join(temp, "checksum")
	for _, dir := range []string{binaryDir, checksumDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	binary, checksums := filepath.Join(binaryDir, name), filepath.Join(checksumDir, "checksums.txt")
	if err := os.WriteFile(binary, []byte("not a release binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(checksums, []byte(strings.Repeat("0", 64)+"  "+name+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prefix := "https://github.com/Onicc/frp-panel/releases/download/edge/"
	rel := release.Release{TagName: "edge", Commit: strings.Repeat("b", 40), Assets: []release.Asset{
		{ID: 1, Name: name, BrowserDownloadURL: prefix + name, Size: int64(len("not a release binary"))},
		{ID: 2, Name: "checksums.txt", BrowserDownloadURL: prefix + "checksums.txt"},
	}}
	err := stage(context.Background(), rel, "checksum-test", func(_ context.Context, url, _ string) (string, error) {
		if url == prefix+name {
			return binary, nil
		}
		return checksums, nil
	})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum rejection, got %v", err)
	}
	if _, err := os.Stat(PendingPath()); !os.IsNotExist(err) {
		t.Fatalf("unverified update became pending: %v", err)
	}
}
