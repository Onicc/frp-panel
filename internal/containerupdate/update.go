package containerupdate

import (
	"context"
	"crypto/sha256"
	"debug/elf"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Onicc/frp-panel/internal/release"
	"github.com/Onicc/frp-panel/utils"
)

type Manifest struct {
	OperationID string    `json:"operationId"`
	Commit      string    `json:"commit"`
	Version     string    `json:"version"`
	ImageCommit string    `json:"imageCommit"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Status struct {
	OperationID string    `json:"operationId"`
	Commit      string    `json:"commit"`
	Version     string    `json:"version"`
	State       string    `json:"state"`
	Error       string    `json:"error,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func Enabled() bool {
	if os.Getenv("FRP_PANEL_CONTAINER_UPDATES") != "1" || !validCommit(os.Getenv("FRP_PANEL_IMAGE_COMMIT")) {
		return false
	}
	launcher, err := os.Stat("/usr/local/bin/frp-panel-launcher")
	return err == nil && launcher.Mode().IsRegular() && launcher.Mode().Perm()&0o111 != 0
}

func Root() string {
	if value := strings.TrimSpace(os.Getenv("FRP_PANEL_UPDATE_DIR")); value != "" {
		return value
	}
	return "/data/update"
}

func BinaryPath(commit string) string { return filepath.Join(Root(), "bin", commit, "frp-panel") }
func PendingPath() string             { return filepath.Join(Root(), "pending.json") }
func ActivePath() string              { return filepath.Join(Root(), "active.json") }
func PreviousPath() string            { return filepath.Join(Root(), "previous.json") }
func StatusPath() string              { return filepath.Join(Root(), "status.json") }

func ReadManifest(path string) (Manifest, error) {
	var value Manifest
	raw, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if err = json.Unmarshal(raw, &value); err != nil {
		return value, err
	}
	if !validCommit(value.Commit) {
		return value, errors.New("invalid update commit")
	}
	return value, nil
}

func ReadStatus() (Status, error) {
	var value Status
	raw, err := os.ReadFile(StatusPath())
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(raw, &value)
	return value, err
}

func WriteStatus(value Status) error {
	value.UpdatedAt = time.Now().UTC()
	return writeJSON(StatusPath(), value)
}

func WriteManifest(path string, value Manifest) error {
	if !validCommit(value.Commit) {
		return errors.New("invalid update commit")
	}
	return writeJSON(path, value)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".update-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(raw); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func validCommit(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, ch := range value {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}

// Stage downloads only the official controller asset for one resolved release.
// The checked binary is retained on the persistent data volume; the image's
// executable and Docker daemon are never written by the application.
func Stage(ctx context.Context, rel release.Release, operationID string) error {
	if !Enabled() {
		return errors.New("this image does not support in-container updates; pull and redeploy the new image first")
	}
	return stage(ctx, rel, operationID, utils.DownloadFile)
}

func stage(ctx context.Context, rel release.Release, operationID string, download func(context.Context, string, string) (string, error)) error {
	if !validCommit(rel.Commit) || rel.TagName == "" || operationID == "" {
		return errors.New("invalid release or operation")
	}
	if _, err := os.Stat(PendingPath()); err == nil {
		return errors.New("an update is already pending")
	}
	assetName := "frp-panel-linux-" + runtime.GOARCH
	var binaryAsset, checksumAsset release.Asset
	for _, asset := range rel.Assets {
		if asset.Name == assetName {
			binaryAsset = asset
		}
		if asset.Name == "checksums.txt" {
			checksumAsset = asset
		}
	}
	if binaryAsset.ID == 0 || checksumAsset.ID == 0 || binaryAsset.Size > 200<<20 {
		return errors.New("release is missing a compatible controller or checksums")
	}
	for _, asset := range []release.Asset{binaryAsset, checksumAsset} {
		prefix := "https://github.com/Onicc/frp-panel/releases/download/" + rel.TagName + "/"
		if asset.BrowserDownloadURL != prefix+asset.Name {
			return errors.New("release asset URL is not official")
		}
	}
	_ = WriteStatus(Status{OperationID: operationID, Commit: rel.Commit, Version: rel.TagName, State: "downloading"})
	binaryTmp, err := download(ctx, binaryAsset.BrowserDownloadURL, "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(binaryTmp))
	info, err := os.Stat(binaryTmp)
	if err != nil || info.Size() == 0 || info.Size() > 200<<20 || (binaryAsset.Size > 0 && info.Size() != binaryAsset.Size) {
		return errors.New("release asset size mismatch or exceeds 200 MiB")
	}
	checksumsTmp, err := download(ctx, checksumAsset.BrowserDownloadURL, "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(checksumsTmp))
	checksumInfo, err := os.Stat(checksumsTmp)
	if err != nil || checksumInfo.Size() == 0 || checksumInfo.Size() > 1<<20 {
		return errors.New("release checksums size is invalid")
	}
	checksums, err := os.ReadFile(checksumsTmp)
	if err != nil {
		return err
	}
	var expected string
	for _, line := range strings.Split(string(checksums), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && strings.TrimPrefix(parts[1], "*") == assetName {
			expected = parts[0]
			break
		}
	}
	if len(expected) != 64 {
		return errors.New("asset missing from checksums.txt")
	}
	file, err := os.Open(binaryTmp)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(h, io.LimitReader(file, 200<<20+1))
	file.Close()
	if err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), expected) {
		return errors.New("release asset checksum mismatch")
	}
	elfFile, err := elf.Open(binaryTmp)
	if err != nil {
		return err
	}
	machine := elfFile.FileHeader.Machine
	elfFile.Close()
	if (runtime.GOARCH == "amd64" && machine != elf.EM_X86_64) || (runtime.GOARCH == "arm64" && machine != elf.EM_AARCH64) {
		return errors.New("release asset architecture mismatch")
	}
	versionCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := os.Chmod(binaryTmp, 0o755); err != nil {
		return err
	}
	output, err := exec.CommandContext(versionCtx, binaryTmp, "version").CombinedOutput()
	if err != nil || !strings.Contains(string(output), "GitCommit: "+rel.Commit) {
		return fmt.Errorf("release executable commit does not match resolved release: %w", err)
	}
	target := BinaryPath(rel.Commit)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	staged := target + ".new"
	defer os.Remove(staged)
	in, err := os.Open(binaryTmp)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(staged, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		in.Close()
		return err
	}
	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		out.Close()
		return err
	}
	if err = out.Sync(); err != nil {
		out.Close()
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	if err = os.Rename(staged, target); err != nil {
		return err
	}
	manifest := Manifest{OperationID: operationID, Commit: rel.Commit, Version: rel.TagName, ImageCommit: os.Getenv("FRP_PANEL_IMAGE_COMMIT"), CreatedAt: time.Now().UTC()}
	if err = WriteManifest(PendingPath(), manifest); err != nil {
		return err
	}
	if err := WriteStatus(Status{OperationID: operationID, Commit: rel.Commit, Version: rel.TagName, State: "staged"}); err != nil {
		_ = os.Remove(PendingPath())
		return err
	}
	return nil
}

func RestartAfterResponse() {
	go func() {
		time.Sleep(750 * time.Millisecond)
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()
}
