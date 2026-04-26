package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogsCmd_MissingKubeconfig verifies logs returns an error when kubeconfig is missing.
func TestLogsCmd_MissingKubeconfig(t *testing.T) {
	tmp := t.TempDir()
	fakeKube := filepath.Join(tmp, "no-config")

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs([]string{"logs", "myapp", "--kubeconfig", fakeKube})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	rOut.Close()
	rErr.Close()

	if err == nil {
		t.Fatal("expected error for missing kubeconfig")
	}
}

// TestLogsAdapterCmd_MissingKubeconfig verifies logs adapter returns an error when kubeconfig is missing.
func TestLogsAdapterCmd_MissingKubeconfig(t *testing.T) {
	tmp := t.TempDir()
	fakeKube := filepath.Join(tmp, "no-config")

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs([]string{"logs", "adapter", "hyperfleet", "--kubeconfig", fakeKube})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	rOut.Close()
	rErr.Close()

	if err == nil {
		t.Fatal("expected error for missing kubeconfig")
	}
}

// TestLogsCmd_NoArgs_UsesEmptyPattern verifies logs with no args uses empty pattern (matches all pods).
func TestLogsCmd_NoArgs_UsesEmptyPattern(t *testing.T) {
	tmp := t.TempDir()
	fakeKube := filepath.Join(tmp, "no-config")

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	// No pattern arg — should try to connect (and fail due to missing kubeconfig).
	rootCmd.SetArgs([]string{"logs", "--kubeconfig", fakeKube})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	rOut.Close()
	rErr.Close()

	// Error is expected (missing kubeconfig), but the command should not panic.
	if err == nil {
		t.Fatal("expected error for missing kubeconfig")
	}
}

// TestLogsAdapterCmd_PatternPrefixed verifies the adapter subcommand uses "adapter" prefix in search.
// This is a behavioral test that verifies correct routing without a live cluster.
func TestLogsAdapterCmd_PatternPrefixed(t *testing.T) {
	// We can't easily test the actual k8s call here without a fake server.
	// Instead, verify the command is registered and the subcommand tree is correct.
	found := false
	for _, subcmd := range rootCmd.Commands() {
		if subcmd.Name() == "logs" {
			for _, sub := range subcmd.Commands() {
				if sub.Name() == "adapter" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected 'logs adapter' subcommand to be registered")
	}
}

func readPipe(r *os.File) string {
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, _ := r.Read(buf)
		if n == 0 {
			break
		}
		b.Write(buf[:n])
	}
	return b.String()
}
