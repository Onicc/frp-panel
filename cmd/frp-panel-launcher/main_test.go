package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Onicc/frp-panel/internal/containerupdate"
)

func TestSelectPromoteAndRollback(t *testing.T) {
	t.Setenv("FRP_PANEL_UPDATE_DIR", t.TempDir())
	imageCommit := strings.Repeat("a", 40)
	newCommit := strings.Repeat("b", 40)
	manifest := containerupdate.Manifest{OperationID: "test-op", Commit: newCommit, Version: "edge", ImageCommit: imageCommit}
	path := containerupdate.BinaryPath(newCommit)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("test executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := containerupdate.WriteManifest(containerupdate.PendingPath(), manifest); err != nil {
		t.Fatal(err)
	}
	selected, pending := selectBinary(imageCommit)
	if selected != path || pending == nil {
		t.Fatalf("pending selection: %q, %+v", selected, pending)
	}
	if err := promote(*pending); err != nil {
		t.Fatal(err)
	}
	selected, pending = selectBinary(imageCommit)
	if selected != path || pending != nil {
		t.Fatalf("active selection: %q, %+v", selected, pending)
	}
	if selected, _ := selectBinary(strings.Repeat("c", 40)); selected != imageBinary {
		t.Fatalf("new image must win over an overlay: %q", selected)
	}
	status, err := containerupdate.ReadStatus()
	if err != nil || status.State != "succeeded" {
		t.Fatalf("promotion status: %+v, %v", status, err)
	}
	if err := containerupdate.WriteManifest(containerupdate.PendingPath(), manifest); err != nil {
		t.Fatal(err)
	}
	rollback(manifest, errors.New("readiness timeout"))
	if _, err := os.Stat(containerupdate.PendingPath()); !os.IsNotExist(err) {
		t.Fatalf("pending still exists: %v", err)
	}
	status, err = containerupdate.ReadStatus()
	if err != nil || status.State != "rolled_back" {
		t.Fatalf("rollback status: %+v, %v", status, err)
	}
	if err := containerupdate.WriteManifest(containerupdate.PreviousPath(), manifest); err != nil {
		t.Fatal(err)
	}
	bad := containerupdate.Manifest{OperationID: "failed-active", Commit: strings.Repeat("c", 40), Version: "edge", ImageCommit: imageCommit}
	if err := containerupdate.WriteManifest(containerupdate.ActivePath(), bad); err != nil {
		t.Fatal(err)
	}
	recoverActive(errors.New("active process exited"))
	active, err := containerupdate.ReadManifest(containerupdate.ActivePath())
	if err != nil || active.Commit != newCommit {
		t.Fatalf("previous binary not restored: %+v, %v", active, err)
	}
	status, err = containerupdate.ReadStatus()
	if err != nil || status.OperationID != "failed-active" || status.State != "rolled_back" {
		t.Fatalf("active rollback status: %+v, %v", status, err)
	}
}

func TestInterruptedStageIsMarkedFailed(t *testing.T) {
	t.Setenv("FRP_PANEL_UPDATE_DIR", t.TempDir())
	status := containerupdate.Status{OperationID: "interrupted", Commit: strings.Repeat("d", 40), Version: "edge", State: "downloading"}
	if err := containerupdate.WriteStatus(status); err != nil {
		t.Fatal(err)
	}
	reconcileInterruptedStage()
	got, err := containerupdate.ReadStatus()
	if err != nil || got.State != "failed" || got.OperationID != "interrupted" {
		t.Fatalf("unexpected interrupted status: %+v, %v", got, err)
	}
}
