package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/maestro"
	"os"
)

// runMaestroCmd runs a maestro command with the Maestro endpoint set to srv.URL.
// It writes a config.yaml to a temp dir with the maestro http-endpoint configured.
func runMaestroCmd(t *testing.T, srv *httptest.Server, args ...string) (string, string, error) {
	t.Helper()
	cfgDir := t.TempDir()

	// Write config.yaml with maestro endpoint pointing at the test server.
	cfgContent := fmt.Sprintf("maestro:\n  http-endpoint: %q\n  consumer: test-consumer\n", srv.URL)
	if err := os.WriteFile(cfgDir+"/config.yaml", []byte(cfgContent), 0600); err != nil {
		t.Fatalf("could not write config.yaml: %v", err)
	}

	fullArgs := append([]string{"--config", cfgDir}, args...)
	return runCmdRaw(t, fullArgs)
}

// maestroResourceJSON returns a minimal Maestro resource as JSON bytes.
func maestroResourceJSON(id, name, consumer string) []byte {
	r := maestro.Resource{
		ID:            id,
		Kind:          "Resource",
		Name:          name,
		ConsumerName:  consumer,
		Version:       1,
		ManifestCount: 2,
		Manifests:     []maestro.Manifest{{Kind: "Deployment", Name: name + "-deploy", Namespace: "default"}},
		Conditions:    []maestro.Condition{{Type: "Applied", Status: "True", Reason: "Applied"}},
	}
	b, _ := json.Marshal(r)
	return b
}

func maestroResourceListJSON(items []maestro.Resource) []byte {
	type listResp struct {
		Items []maestro.Resource `json:"items"`
		Kind  string             `json:"kind"`
		Total int                `json:"total"`
	}
	b, _ := json.Marshal(listResp{Items: items, Kind: "ResourceList", Total: len(items)})
	return b
}

func maestroBundleListJSON(items []maestro.Bundle) []byte {
	type listResp struct {
		Items []maestro.Bundle `json:"items"`
		Kind  string           `json:"kind"`
		Total int              `json:"total"`
	}
	b, _ := json.Marshal(listResp{Items: items, Kind: "ResourceBundleList", Total: len(items)})
	return b
}

func maestroConsumerListJSON(items []maestro.Consumer) []byte {
	type listResp struct {
		Items []maestro.Consumer `json:"items"`
		Kind  string             `json:"kind"`
		Total int                `json:"total"`
	}
	b, _ := json.Marshal(listResp{Items: items, Kind: "ConsumerList", Total: len(items)})
	return b
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestMaestroList_RendersTable(t *testing.T) {
	resources := []maestro.Resource{
		{ID: "r-001", Name: "mw-abc", ConsumerName: "cluster1", Version: 3, ManifestCount: 4},
		{ID: "r-002", Name: "mw-def", ConsumerName: "cluster1", Version: 1, ManifestCount: 2},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(maestroResourceListJSON(resources))
	}))
	defer srv.Close()

	stdout, _, err := runMaestroCmd(t, srv, "--no-color", "maestro", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, h := range []string{"NAME", "CONSUMER", "VERSION", "MANIFESTS"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q in output, got:\n%s", h, stdout)
		}
	}
	if !strings.Contains(stdout, "mw-abc") || !strings.Contains(stdout, "mw-def") {
		t.Errorf("expected resource names in output, got:\n%s", stdout)
	}
}

// ── get ───────────────────────────────────────────────────────────────────────

func TestMaestroGet_PrintsJSON(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write(maestroResourceJSON("r-001", "mw-abc", "cluster1"))
	}))
	defer srv.Close()

	stdout, _, err := runMaestroCmd(t, srv, "maestro", "get", "mw-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(strings.TrimRight(capturedPath, "/"), "/resources/mw-abc") {
		t.Errorf("expected GET /resources/mw-abc, got path: %s", capturedPath)
	}

	var got maestro.Resource
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\nout: %s", err, stdout)
	}
	if got.Name != "mw-abc" {
		t.Errorf("name = %q, want mw-abc", got.Name)
	}
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestMaestroDelete_WithYesFlag_CallsDELETE(t *testing.T) {
	var capturedMethod, capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	_, stderr, err := runMaestroCmd(t, srv, "maestro", "delete", "mw-abc", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", capturedMethod)
	}
	if !strings.HasSuffix(capturedPath, "/resources/mw-abc") {
		t.Errorf("unexpected path: %s", capturedPath)
	}
	if !strings.Contains(stderr, "Deleted") {
		t.Errorf("expected deletion confirmation in stderr, got: %s", stderr)
	}
}

func TestMaestroDelete_CancelledByUser_NoRequest(t *testing.T) {
	// Reset --yes flag to false in case a previous test left it set.
	maestroDeleteCmd.Flags().Set("yes", "false") //nolint:errcheck

	requestReceived := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	// Provide "n" on stdin to cancel
	oldStdin := os.Stdin
	pipeR, pipeW, _ := os.Pipe()
	os.Stdin = pipeR
	fmt.Fprint(pipeW, "n\n")
	pipeW.Close()

	_, stderr, err := runMaestroCmd(t, srv, "maestro", "delete", "mw-abc")
	os.Stdin = oldStdin
	pipeR.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestReceived {
		t.Error("DELETE request should not have been sent when user cancels")
	}
	if !strings.Contains(stderr, "cancelled") {
		t.Errorf("expected cancellation message, got: %s", stderr)
	}
}

// ── bundles ───────────────────────────────────────────────────────────────────

func TestMaestroBundles_PrintsJSON(t *testing.T) {
	var capturedPath string
	bundles := []maestro.Bundle{
		{ID: "b-001", Name: "bundle-1", Kind: "ResourceBundle"},
		{ID: "b-002", Name: "bundle-2", Kind: "ResourceBundle"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write(maestroBundleListJSON(bundles))
	}))
	defer srv.Close()

	stdout, _, err := runMaestroCmd(t, srv, "maestro", "bundles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(strings.TrimRight(capturedPath, "/"), "/resource-bundles") {
		t.Errorf("expected /resource-bundles path, got: %s", capturedPath)
	}
	if !strings.Contains(stdout, "bundle-1") || !strings.Contains(stdout, "bundle-2") {
		t.Errorf("expected bundle names in output, got:\n%s", stdout)
	}
}

// ── consumers ─────────────────────────────────────────────────────────────────

func TestMaestroConsumers_RendersTable(t *testing.T) {
	consumers := []maestro.Consumer{
		{ID: "c-001", Kind: "Consumer", Name: "cluster1"},
		{ID: "c-002", Kind: "Consumer", Name: "cluster2"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(maestroConsumerListJSON(consumers))
	}))
	defer srv.Close()

	stdout, _, err := runMaestroCmd(t, srv, "--no-color", "maestro", "consumers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, h := range []string{"ID", "NAME"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q in output, got:\n%s", h, stdout)
		}
	}
	if !strings.Contains(stdout, "cluster1") || !strings.Contains(stdout, "cluster2") {
		t.Errorf("expected consumer names in output, got:\n%s", stdout)
	}
}

// ── tui ───────────────────────────────────────────────────────────────────────

func TestMaestroTUI_MissingBinary_ReturnsError(t *testing.T) {
	// Set PATH to empty so maestro-cli cannot be found.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()

	_, _, err := runMaestroCmd(t, srv, "maestro", "tui")
	if err == nil {
		t.Fatal("expected error when maestro-cli not in PATH")
	}
	if !strings.Contains(err.Error(), "maestro-cli") {
		t.Errorf("expected error mentioning maestro-cli, got: %v", err)
	}
}
