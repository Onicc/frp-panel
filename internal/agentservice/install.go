package agentservice

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

type InstallOptions struct {
	GOOS       string
	Root       string
	Executable string
	Config     []byte
	Start      bool
}

func Install(opts InstallOptions) (Layout, error) {
	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	layout, err := LayoutFor(goos, opts.Root, os.Getenv)
	if err != nil {
		return Layout{}, err
	}
	if opts.Executable == "" {
		opts.Executable, err = os.Executable()
		if err != nil {
			return Layout{}, fmt.Errorf("locate current executable: %w", err)
		}
	}
	if len(opts.Config) == 0 {
		return Layout{}, errors.New("agent configuration is required")
	}
	if err := atomicCopy(opts.Executable, layout.Binary, 0o755); err != nil {
		return Layout{}, fmt.Errorf("install binary: %w", err)
	}
	if err := atomicWrite(layout.Config, opts.Config, 0o600); err != nil {
		return Layout{}, fmt.Errorf("install config: %w", err)
	}
	if err := os.MkdirAll(layout.Data, 0o750); err != nil {
		return Layout{}, fmt.Errorf("create data directory: %w", err)
	}
	if layout.GOOS != "windows" {
		if err := os.Chmod(layout.Data, 0o750); err != nil {
			return Layout{}, fmt.Errorf("secure data directory: %w", err)
		}
	}
	if (opts.Root == "" || opts.Root == string(filepath.Separator)) && layout.GOOS != "windows" {
		if err := ensureServiceIdentity(layout); err != nil {
			return Layout{}, err
		}
		if err := secureOwnership(layout); err != nil {
			return Layout{}, err
		}
	}
	if (opts.Root == "" || opts.Root == string(filepath.Separator)) && layout.GOOS == "windows" {
		if err := secureWindowsACL(layout); err != nil {
			return Layout{}, err
		}
	}
	if layout.ServiceFile != "" {
		definition, err := RenderService(layout)
		if err != nil {
			return Layout{}, err
		}
		if err := atomicWrite(layout.ServiceFile, definition, 0o644); err != nil {
			return Layout{}, fmt.Errorf("install service definition: %w", err)
		}
	}
	// A non-system root is intentionally staging-only. This makes native package
	// verification safe and deterministic without touching the host service manager.
	if opts.Root != "" && opts.Root != string(filepath.Separator) {
		return layout, nil
	}
	if err := installNative(layout, opts.Start); err != nil {
		return Layout{}, err
	}
	return layout, nil
}

func Uninstall(goos, root string, purge bool) (Layout, error) {
	if goos == "" {
		goos = runtime.GOOS
	}
	layout, err := LayoutFor(goos, root, os.Getenv)
	if err != nil {
		return Layout{}, err
	}
	if root == "" || root == string(filepath.Separator) {
		_ = uninstallNative(layout)
	}
	for _, target := range []string{layout.ServiceFile, layout.Binary, layout.Binary + ".previous"} {
		if target != "" {
			if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
				return layout, err
			}
		}
	}
	if purge {
		for _, target := range []string{layout.Config, layout.Data} {
			if target == "" {
				continue
			}
			if err := os.RemoveAll(target); err != nil {
				return layout, err
			}
		}
	}
	return layout, nil
}

func RenderService(layout Layout) ([]byte, error) {
	var source string
	switch layout.GOOS {
	case "linux":
		source = `[Unit]
Description=frp-panel agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User={{.ServiceUser}}
Group={{.ServiceUser}}
ExecStart={{quote .Binary}} agent run --config {{quote .Config}}
WorkingDirectory={{quote .Data}}
Restart=on-failure
RestartSec=5s
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths={{quote .Data}}

[Install]
WantedBy=multi-user.target
`
	case "darwin":
		source = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>io.github.onicc.frp-panel.agent</string>
<key>ProgramArguments</key><array>
<string>{{xml .Binary}}</string><string>agent</string><string>run</string>
<string>--config</string><string>{{xml .Config}}</string>
</array>
<key>WorkingDirectory</key><string>{{xml .Data}}</string>
<key>UserName</key><string>{{xml .ServiceUser}}</string>
<key>RunAtLoad</key><true/><key>KeepAlive</key><true/>
<key>ProcessType</key><string>Background</string>
</dict></plist>
`
	default:
		return nil, fmt.Errorf("%s uses the native service manager and has no definition file", layout.GOOS)
	}
	tmpl, err := template.New("service").Funcs(template.FuncMap{
		"quote": func(v string) string { return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"` },
		"xml": func(v string) string {
			var b bytes.Buffer
			template.HTMLEscape(&b, []byte(v))
			return b.String()
		},
	}).Parse(source)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, layout); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func atomicCopy(source, destination string, mode os.FileMode) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	return atomicFromReader(destination, src, mode)
}

func atomicWrite(destination string, data []byte, mode os.FileMode) error {
	return atomicFromReader(destination, bytes.NewReader(data), mode)
}

func atomicFromReader(destination string, source io.Reader, mode os.FileMode) error {
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".frp-panel-agent-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := io.Copy(tmp, io.LimitReader(source, 512<<20)); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	backup := destination + ".previous"
	if _, err := os.Stat(destination); err == nil {
		_ = os.Remove(backup)
		if err := os.Rename(destination, backup); err != nil {
			return err
		}
	}
	if err := os.Rename(tmpName, destination); err != nil {
		_ = os.Rename(backup, destination)
		return err
	}
	return nil
}

func installNative(layout Layout, start bool) error {
	var commands [][]string
	switch layout.GOOS {
	case "linux":
		commands = append(commands, []string{"systemctl", "daemon-reload"}, []string{"systemctl", "enable", ServiceName})
		if start {
			commands = append(commands, []string{"systemctl", "restart", ServiceName})
		}
	case "darwin":
		_ = exec.Command("launchctl", "bootout", "system/io.github.onicc.frp-panel.agent").Run()
		commands = append(commands, []string{"launchctl", "bootstrap", "system", layout.ServiceFile})
		if start {
			commands = append(commands, []string{"launchctl", "kickstart", "-k", "system/io.github.onicc.frp-panel.agent"})
		}
	case "windows":
		_ = exec.Command("sc.exe", "stop", ServiceName).Run()
		_ = exec.Command("sc.exe", "delete", ServiceName).Run()
		binPath := fmt.Sprintf(`"%s" agent run --config "%s"`, layout.Binary, layout.Config)
		commands = append(commands, []string{"sc.exe", "create", ServiceName, "binPath=", binPath, "start=", "auto", "obj=", layout.ServiceUser})
		if start {
			commands = append(commands, []string{"sc.exe", "start", ServiceName})
		}
	}
	for _, command := range commands {
		if output, err := exec.Command(command[0], command[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %w: %s", command[0], err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func secureWindowsACL(layout Layout) error {
	commands := [][]string{
		{"icacls.exe", layout.Config, "/inheritance:r", "/grant:r", "*S-1-5-18:F", "*S-1-5-32-544:F", "*S-1-5-19:R"},
		{"icacls.exe", layout.Data, "/inheritance:r", "/grant:r", "*S-1-5-18:(OI)(CI)F", "*S-1-5-32-544:(OI)(CI)F", "*S-1-5-19:(OI)(CI)M"},
	}
	for _, command := range commands {
		if output, err := exec.Command(command[0], command[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("protect agent data with Windows ACLs: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func ensureServiceIdentity(layout Layout) error {
	if _, err := user.Lookup(layout.ServiceUser); err == nil {
		return nil
	}
	switch layout.GOOS {
	case "linux":
		output, err := exec.Command("useradd", "--system", "--user-group", "--home-dir", layout.Data, "--shell", "/usr/sbin/nologin", layout.ServiceUser).CombinedOutput()
		if err != nil {
			return fmt.Errorf("create service user: %w: %s", err, strings.TrimSpace(string(output)))
		}
	case "darwin":
		uid, err := unusedDarwinSystemID()
		if err != nil {
			return err
		}
		group := "/Groups/" + layout.ServiceUser
		account := "/Users/" + layout.ServiceUser
		commands := [][]string{
			{"dscl", ".", "-create", group},
			{"dscl", ".", "-create", group, "PrimaryGroupID", strconv.Itoa(uid)},
			{"dscl", ".", "-create", account},
			{"dscl", ".", "-create", account, "UniqueID", strconv.Itoa(uid)},
			{"dscl", ".", "-create", account, "PrimaryGroupID", strconv.Itoa(uid)},
			{"dscl", ".", "-create", account, "UserShell", "/usr/bin/false"},
			{"dscl", ".", "-create", account, "NFSHomeDirectory", layout.Data},
			{"dscl", ".", "-create", account, "IsHidden", "1"},
			{"dscl", ".", "-append", group, "GroupMembership", layout.ServiceUser},
		}
		for _, command := range commands {
			if output, err := exec.Command(command[0], command[1:]...).CombinedOutput(); err != nil {
				return fmt.Errorf("create service user: %w: %s", err, strings.TrimSpace(string(output)))
			}
		}
	}
	return nil
}

func unusedDarwinSystemID() (int, error) {
	for id := 399; id >= 350; id-- {
		if _, err := user.LookupId(strconv.Itoa(id)); err != nil {
			return id, nil
		}
	}
	return 0, errors.New("no available macOS system UID in range 350-399")
}

func secureOwnership(layout Layout) error {
	account, err := user.Lookup(layout.ServiceUser)
	if err != nil {
		return fmt.Errorf("look up service user: %w", err)
	}
	uid, err := strconv.Atoi(account.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(account.Gid)
	if err != nil {
		return err
	}
	if err := os.Chown(layout.Config, uid, gid); err != nil {
		return fmt.Errorf("secure config ownership: %w", err)
	}
	if err := filepath.Walk(layout.Data, func(path string, _ os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		return os.Chown(path, uid, gid)
	}); err != nil {
		return fmt.Errorf("secure data ownership: %w", err)
	}
	return nil
}

func uninstallNative(layout Layout) error {
	switch layout.GOOS {
	case "linux":
		_ = exec.Command("systemctl", "disable", "--now", ServiceName).Run()
	case "darwin":
		_ = exec.Command("launchctl", "bootout", "system/io.github.onicc.frp-panel.agent").Run()
	case "windows":
		_ = exec.Command("sc.exe", "stop", ServiceName).Run()
		_ = exec.Command("sc.exe", "delete", ServiceName).Run()
	}
	return nil
}

func Control(action string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "linux":
		name, args = "systemctl", []string{action, ServiceName}
	case "darwin":
		name = "launchctl"
		switch action {
		case "start", "restart":
			args = []string{"kickstart", "-k", "system/io.github.onicc.frp-panel.agent"}
		case "stop":
			args = []string{"kill", "SIGTERM", "system/io.github.onicc.frp-panel.agent"}
		case "status":
			args = []string{"print", "system/io.github.onicc.frp-panel.agent"}
		default:
			return fmt.Errorf("unsupported action %q", action)
		}
	case "windows":
		name = "sc.exe"
		switch action {
		case "status":
			args = []string{"query", ServiceName}
		case "start", "stop":
			args = []string{action, ServiceName}
		case "restart":
			if err := exec.Command(name, "stop", ServiceName).Run(); err != nil {
				return err
			}
			args = []string{"start", ServiceName}
		default:
			return fmt.Errorf("unsupported action %q", action)
		}
	default:
		return fmt.Errorf("unsupported operating system %q", runtime.GOOS)
	}
	command := exec.Command(name, args...)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	return command.Run()
}
