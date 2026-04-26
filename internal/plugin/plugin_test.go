package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

// makeTempPlugin writes a minimal executable script named "hf-<name>" to a temp dir
// and returns the directory path. The caller must add the dir to PATH.
func makeTempPlugin(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "hf-"+name)
	content := "#!/bin/sh\necho \"plugin-$*\"\n"
	if err := os.WriteFile(pluginPath, []byte(content), 0o755); err != nil {
		t.Fatalf("write plugin: %v", err)
	}
	return dir
}

func TestFind_PluginOnPath(t *testing.T) {
	dir := makeTempPlugin(t, "testplugin")

	orig := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)

	path, ok := Find("testplugin")
	if !ok {
		t.Fatal("Find returned false; expected plugin to be found on PATH")
	}
	if path == "" {
		t.Error("Find returned empty path when plugin found")
	}
}

func TestFind_PluginNotOnPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty dir — no plugins

	_, ok := Find("notexist")
	if ok {
		t.Error("Find returned true for non-existent plugin")
	}
}

func TestFind_NameLowercased(t *testing.T) {
	dir := makeTempPlugin(t, "myplugin")

	orig := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)

	// Pass uppercase name; Find should lowercase it.
	_, ok := Find("MyPlugin")
	if !ok {
		t.Error("Find should lowercase the name before searching")
	}
}

func TestFind_ReturnsAbsolutePath(t *testing.T) {
	dir := makeTempPlugin(t, "absplugin")

	orig := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)

	path, ok := Find("absplugin")
	if !ok {
		t.Fatal("plugin not found")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got: %s", path)
	}
}
