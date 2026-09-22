package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Onicc/frp-panel/internal/containerupdate"
)

const imageBinary = "/usr/local/bin/frp-panel"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "launcher:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		args = []string{"master"}
	}
	reconcileInterruptedStage()
	imageCommit := os.Getenv("FRP_PANEL_IMAGE_COMMIT")
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)
	for {
		candidate, pending := selectBinary(imageCommit)
		child := exec.Command(candidate, args...) // #nosec G702 -- exec.Command does not invoke a shell; candidate is the image binary or a validated staged path.
		child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			if pending != nil {
				rollback(*pending, err)
				continue
			}
			if candidate != imageBinary {
				recoverActive(err)
				continue
			}
			return err
		}
		done := make(chan error, 1)
		go func() { done <- child.Wait() }()
		if pending != nil {
			if err := waitHealthy(args[0], child, done, stop); err != nil {
				rollback(*pending, err)
				if errors.Is(err, errStopped) {
					return nil
				}
				continue
			}
			if err := promote(*pending); err != nil {
				_ = child.Process.Signal(syscall.SIGTERM)
				<-done
				rollback(*pending, err)
				continue
			}
		} else if candidate != imageBinary {
			if err := waitHealthy(args[0], child, done, stop); err != nil {
				if errors.Is(err, errStopped) {
					return nil
				}
				recoverActive(err)
				continue
			}
		}
		select {
		case err := <-done:
			if _, statErr := os.Stat(containerupdate.PendingPath()); statErr == nil {
				continue
			}
			if err == nil {
				return nil
			}
			return err
		case <-stop:
			_ = child.Process.Signal(syscall.SIGTERM)
			select {
			case <-done:
			case <-time.After(25 * time.Second):
				_ = child.Process.Kill()
				<-done
			}
			return nil
		}
	}
}

func reconcileInterruptedStage() {
	status, err := containerupdate.ReadStatus()
	if err != nil || (status.State != "queued" && status.State != "downloading" && status.State != "staged") {
		return
	}
	if _, err := os.Stat(containerupdate.PendingPath()); err == nil {
		return
	}
	_ = containerupdate.WriteStatus(containerupdate.Status{OperationID: status.OperationID, Commit: status.Commit, Version: status.Version, State: "failed", Error: "update was interrupted before a staged binary was ready"})
}

func selectBinary(imageCommit string) (string, *containerupdate.Manifest) {
	if candidate, err := containerupdate.ReadManifest(containerupdate.PendingPath()); err == nil {
		if candidate.ImageCommit != imageCommit {
			rollback(candidate, errors.New("new Docker image superseded pending update"))
		} else if validBinary(candidate.Commit) {
			return containerupdate.BinaryPath(candidate.Commit), &candidate
		} else {
			rollback(candidate, errors.New("pending executable is missing or unsafe"))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "launcher: discarding invalid pending update:", err)
		_ = os.Remove(containerupdate.PendingPath())
	}
	if active, err := containerupdate.ReadManifest(containerupdate.ActivePath()); err == nil && active.ImageCommit == imageCommit {
		if validBinary(active.Commit) {
			return containerupdate.BinaryPath(active.Commit), nil
		}
	}
	return imageBinary, nil
}

func validBinary(commit string) bool {
	path := containerupdate.BinaryPath(commit)
	if !strings.HasPrefix(path, containerupdate.Root()+string(os.PathSeparator)) {
		return false
	}
	stat, err := os.Lstat(path)
	return err == nil && stat.Mode().IsRegular() && stat.Mode().Perm()&0o111 != 0
}

var errStopped = errors.New("launcher stopped")

func waitHealthy(role string, child *exec.Cmd, done <-chan error, stop <-chan os.Signal) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	readySince := time.Time{}
	for {
		select {
		case err := <-done:
			return fmt.Errorf("updated process exited before ready: %v", err)
		case <-stop:
			_ = child.Process.Signal(syscall.SIGTERM)
			<-done
			return errStopped
		case <-ctx.Done():
			_ = child.Process.Signal(syscall.SIGTERM)
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				_ = child.Process.Kill()
				<-done
			}
			return errors.New("updated process did not become healthy within 60 seconds")
		case <-ticker.C:
			if !probe(role) {
				readySince = time.Time{}
				continue
			}
			if readySince.IsZero() {
				readySince = time.Now()
				continue
			}
			if time.Since(readySince) >= 2*time.Second {
				return nil
			}
		}
	}
}

func probe(role string) bool {
	if role == "master" {
		port := os.Getenv("MASTER_API_PORT")
		if port == "" {
			port = "9000"
		}
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return false
		}
		client := &http.Client{Timeout: time.Second}
		response, err := client.Get("http://127.0.0.1:" + port + "/api/v2/health") // #nosec G107 G704 -- fixed loopback host and validated numeric local port.
		if err != nil {
			return false
		}
		response.Body.Close()
		return response.StatusCode == http.StatusOK
	}
	if role == "server" {
		port := os.Getenv("SERVER_API_PORT")
		if port == "" {
			port = "8999"
		}
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return false
		}
		client := &http.Client{Timeout: time.Second}
		response, err := client.Get("http://127.0.0.1:" + port + "/health") // #nosec G107 G704 -- fixed loopback host and validated numeric local port.
		if err != nil {
			return false
		}
		response.Body.Close()
		return response.StatusCode == http.StatusOK
	}
	return false
}

func promote(pending containerupdate.Manifest) error {
	if active, err := containerupdate.ReadManifest(containerupdate.ActivePath()); err == nil {
		if err := containerupdate.WriteManifest(containerupdate.PreviousPath(), active); err != nil {
			return err
		}
	}
	if err := containerupdate.WriteManifest(containerupdate.ActivePath(), pending); err != nil {
		return err
	}
	if err := containerupdate.WriteStatus(containerupdate.Status{OperationID: pending.OperationID, Commit: pending.Commit, Version: pending.Version, State: "succeeded"}); err != nil {
		return err
	}
	return os.Remove(containerupdate.PendingPath())
}

func rollback(pending containerupdate.Manifest, reason error) {
	if active, err := containerupdate.ReadManifest(containerupdate.ActivePath()); err == nil && active.OperationID == pending.OperationID && active.Commit == pending.Commit {
		previous, previousErr := containerupdate.ReadManifest(containerupdate.PreviousPath())
		if previousErr == nil && previous.ImageCommit == pending.ImageCommit && previous.Commit != pending.Commit && validBinary(previous.Commit) {
			_ = containerupdate.WriteManifest(containerupdate.ActivePath(), previous)
		} else {
			_ = os.Remove(containerupdate.ActivePath())
		}
	}
	_ = os.Remove(containerupdate.PendingPath())
	_ = containerupdate.WriteStatus(containerupdate.Status{OperationID: pending.OperationID, Commit: pending.Commit, Version: pending.Version, State: "rolled_back", Error: reason.Error()})
}

func recoverActive(reason error) {
	active, err := containerupdate.ReadManifest(containerupdate.ActivePath())
	if err != nil {
		return
	}
	previous, err := containerupdate.ReadManifest(containerupdate.PreviousPath())
	if err == nil && previous.ImageCommit == active.ImageCommit && previous.Commit != active.Commit && validBinary(previous.Commit) {
		_ = containerupdate.WriteManifest(containerupdate.ActivePath(), previous)
	} else {
		_ = os.Remove(containerupdate.ActivePath())
	}
	_ = containerupdate.WriteStatus(containerupdate.Status{OperationID: active.OperationID, Commit: active.Commit, Version: active.Version, State: "rolled_back", Error: reason.Error()})
}
