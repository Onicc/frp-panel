package agentservice

import (
	"os"
	"path/filepath"
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
	}
}
