package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKubeContext_MissingKubeconfig checks that a missing kubeconfig returns an error.
func TestKubeContext_MissingKubeconfig(t *testing.T) {
	// Use a temp dir that has no kubeconfig.
	tmp := t.TempDir()
	fakeKube := filepath.Join(tmp, "no-config")

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs([]string{"kube", "context", "--kubeconfig", fakeKube})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	var bufErr strings.Builder
	buf := make([]byte, 4096)
	for {
		n, _ := rErr.Read(buf)
		if n == 0 {
			break
		}
		bufErr.Write(buf[:n])
	}
	rOut.Close()

	if err == nil {
		t.Fatal("expected error for missing kubeconfig")
	}
}

// TestKubePortForwardStatus_Empty verifies status prints a no-forwards message when no PID files exist.
func TestKubePortForwardStatus_Empty(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	os.MkdirAll(filepath.Join(home, ".config", "hf"), 0755)

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", home)

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs([]string{"kube", "port-forward", "status"})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	var bufOut strings.Builder
	buf := make([]byte, 4096)
	for {
		n, _ := rOut.Read(buf)
		if n == 0 {
			break
		}
		bufOut.Write(buf[:n])
	}
	for {
		n, _ := rErr.Read(buf)
		if n == 0 {
			break
		}
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := bufOut.String()
	if !strings.Contains(out, "SERVICE") {
		t.Errorf("expected SERVICE header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "No port-forwards") {
		t.Errorf("expected 'No port-forwards' message, got:\n%s", out)
	}
}

// TestKubePFStartCmd_InvalidPorts verifies malformed port arg returns an error.
func TestKubePFStartCmd_InvalidPorts(t *testing.T) {
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs([]string{"kube", "port-forward", "start", "api", "not-valid-ports"})
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	rOut.Close()
	rErr.Close()

	if err == nil {
		t.Fatal("expected error for invalid ports format")
	}
}
