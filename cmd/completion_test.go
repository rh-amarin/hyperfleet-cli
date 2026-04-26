package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newNoopServer creates a test HTTP server that returns 200 OK for all requests.
// Used by tests that exercise commands which don't make HTTP calls.
func newNoopServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// runCompletion exercises the completion command's RunE directly, redirecting
// output to a buffer. This avoids the os.Pipe deadlock that occurs when the
// generated script (≥65 KB for bash) exceeds the OS pipe buffer.
func runCompletion(t *testing.T, shell string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	completionCmd.SetOut(&buf)
	defer completionCmd.SetOut(nil)
	err := completionCmd.RunE(completionCmd, []string{shell})
	return buf.String(), err
}

func TestCompletion_Bash(t *testing.T) {
	stdout, err := runCompletion(t, "bash")
	if err != nil {
		t.Fatalf("completion bash: %v", err)
	}
	if len(stdout) == 0 {
		t.Error("expected non-empty bash completion script")
	}
	if !strings.Contains(stdout, "bash") {
		t.Errorf("expected bash-specific content in completion script")
	}
}

func TestCompletion_Zsh(t *testing.T) {
	stdout, err := runCompletion(t, "zsh")
	if err != nil {
		t.Fatalf("completion zsh: %v", err)
	}
	if len(stdout) == 0 {
		t.Error("expected non-empty zsh completion output")
	}
}

func TestCompletion_Fish(t *testing.T) {
	stdout, err := runCompletion(t, "fish")
	if err != nil {
		t.Fatalf("completion fish: %v", err)
	}
	if len(stdout) == 0 {
		t.Error("expected non-empty fish completion output")
	}
}

func TestCompletion_Powershell(t *testing.T) {
	stdout, err := runCompletion(t, "powershell")
	if err != nil {
		t.Fatalf("completion powershell: %v", err)
	}
	if len(stdout) == 0 {
		t.Error("expected non-empty powershell completion output")
	}
}

func TestCompletion_UnknownShell(t *testing.T) {
	_, err := runCompletion(t, "nushell")
	if err == nil {
		t.Fatal("expected error for unknown shell, got nil")
	}
	if !strings.Contains(err.Error(), "unknown shell") {
		t.Errorf("expected 'unknown shell' in error, got: %v", err)
	}
}

func TestCompletion_NoArgs(t *testing.T) {
	srv := newNoopServer(t)
	_, _, err := runCmd(t, srv, "completion")
	if err == nil {
		t.Fatal("expected error with no args, got nil")
	}
}
