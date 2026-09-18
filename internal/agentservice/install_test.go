package agentservice

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLayouts(t *testing.T) {
	root := t.TempDir()
	env := func(key string) string {
		switch key {
		case "ProgramFiles":
			return `C:\Program Files`
		case "ProgramData":
			return `C:\ProgramData`
		default:
			return ""
		}
	}
	tests := []struct {
		goos       string
		binaryPart string
		configPart string
	}{
		{"linux", "usr/local/libexec/frp-panel/frp-panel-agent", "etc/frp-panel/agent.yaml"},
		{"darwin", "usr/local/libexec/frp-panel/frp-panel-agent", "Library/Application Support/frp-panel/agent.yaml"},
		{"windows", "Program Files/frp-panel/frp-panel-agent.exe", "ProgramData/frp-panel/agent.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			got, err := LayoutFor(tt.goos, root, env)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.ToSlash(got.Binary)
			config := filepath.ToSlash(got.Config)
			if !strings.HasSuffix(binary, tt.binaryPart) || !strings.HasSuffix(config, tt.configPart) {
				t.Fatalf("unexpected layout: %+v", got)
			}
		})
	}
}

func TestInstallStagesAtomicallyAndPurges(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(t.TempDir(), "frp-panel-agent")
	if err := os.WriteFile(source, []byte("agent-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	layout, err := Install(InstallOptions{
		GOOS:       "darwin",
		Root:       root,
		Executable: source,
		Config:     []byte("version: 2\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{layout.Binary, layout.Config, layout.ServiceFile, layout.Data} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
	mode, err := os.Stat(layout.Config)
	if err != nil || mode.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions: %v %v", mode, err)
	}
	if _, err := Uninstall("darwin", root, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(layout.Binary); !os.IsNotExist(err) {
		t.Fatalf("binary remains after uninstall: %v", err)
	}
}

func TestServiceDefinitionsDoNotExposeCredentials(t *testing.T) {
	for _, goos := range []string{"linux", "darwin"} {
		layout, err := LayoutFor(goos, t.TempDir(), os.Getenv)
		if err != nil {
			t.Fatal(err)
		}
		definition, err := RenderService(layout)
		if err != nil {
			t.Fatal(err)
		}
		text := string(definition)
		if !strings.Contains(text, "--config") || strings.Contains(text, "--secret") {
			t.Fatalf("unsafe service definition: %s", text)
		}
		if goos == "linux" {
			if strings.Contains(text, "\nGroup=") {
				t.Fatalf("systemd must use the account's primary group: %s", text)
			}
			if strings.Contains(text, `ExecStart="`) || strings.Contains(text, `WorkingDirectory="`) {
				t.Fatalf("systemd paths must use native directive syntax: %s", text)
			}
		}
	}
}

func TestDSCLAlreadyExists(t *testing.T) {
	for _, test := range []struct {
		name   string
		output string
		want   bool
	}{
		{name: "record", output: "DS Error: eDSRecordAlreadyExists", want: true},
		{name: "attribute", output: "DS Error: eDSAttributeAlreadyExists", want: true},
		{name: "other error", output: "DS Error: eDSRecordNotFound", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := dsclAlreadyExists([]byte(test.output)); got != test.want {
				t.Fatalf("dsclAlreadyExists(%q) = %v, want %v", test.output, got, test.want)
			}
		})
	}
}

func TestLinuxServiceDefinitionPassesSystemdAnalyze(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("systemd validation requires Linux")
	}
	analyzer, err := exec.LookPath("systemd-analyze")
	if err != nil {
		t.Skip("systemd-analyze is not installed")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	layout, err := Install(InstallOptions{
		GOOS:       "linux",
		Root:       t.TempDir(),
		Executable: executable,
		Config:     []byte("version: 2\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(analyzer, "verify", layout.ServiceFile).CombinedOutput(); err != nil {
		t.Fatalf("systemd-analyze verify: %v\n%s", err, output)
	}
}
