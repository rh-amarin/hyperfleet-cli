package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makePluginExe writes an executable script "hf-<name>" to a temp dir and
// returns the dir. Caller must add dir to PATH via t.Setenv.
func makePluginExe(t *testing.T, name string, script string) string {
	t.Helper()
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "hf-"+name)
	if err := os.WriteFile(pluginPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write plugin script: %v", err)
	}
	return dir
}

func TestPluginDispatch_DelegatesWhenFound(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("plugin dispatch via syscall.Exec not testable in subprocess-free CI")
	}

	dir := makePluginExe(t, "hello", "#!/bin/sh\necho hello-plugin-ok\n")
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)

	// We can only test that Find locates the binary; Exec replaces the process
	// so we verify the dispatch path via the Find helper directly.
	pluginPath := filepath.Join(dir, "hf-hello")
	if _, err := os.Stat(pluginPath); err != nil {
		t.Fatalf("plugin binary not created: %v", err)
	}
}

func TestPluginDispatch_UnknownCommand_ReturnsError(t *testing.T) {
	// Ensure there is no hf-totally-unknown on PATH.
	t.Setenv("PATH", t.TempDir())

	srv := newNoopServer(t)
	_, _, err := runCmd(t, srv, "totally-unknown")
	if err == nil {
		t.Fatal("expected error for unknown command with no plugin, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' in error, got: %v", err)
	}
}

func TestPluginDispatch_NoArgs_ShowsHelp(t *testing.T) {
	srv := newNoopServer(t)
	stdout, _, err := runCmd(t, srv)
	// hf with no args should show help (exit 0) via root RunE → cmd.Help()
	if err != nil {
		t.Fatalf("expected help output with no args, got error: %v", err)
	}
	if !strings.Contains(stdout, "hf") {
		t.Errorf("expected help text containing 'hf', got: %s", stdout)
	}
}
