package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// adapterStatusJSON returns a minimal AdapterStatus JSON response.
func adapterStatusJSON(adapter, status string) []byte {
	as := resource.AdapterStatus{
		Adapter:            adapter,
		ObservedGeneration: 1,
		Conditions: []resource.AdapterCondition{
			{Type: "Available", Status: status, Reason: "ManualStatusPost"},
			{Type: "Applied", Status: status, Reason: "ManualStatusPost"},
			{Type: "Health", Status: status, Reason: "ManualStatusPost"},
		},
		Data:           map[string]any{},
		CreatedTime:    "2026-04-25T00:00:00Z",
		LastReportTime: "2026-04-25T00:00:00Z",
	}
	b, _ := json.Marshal(as)
	return b
}

// ── cluster adapter post-status ────────────────────────────────────────────────

func TestClusterAdapterPostStatus_PostsCorrectURL(t *testing.T) {
	var capturedMethod, capturedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("cl-deployment", "True"))
	}))
	defer srv.Close()

	_, _, err := runCmdWithCluster(t, srv, "c-001", "cluster", "adapter", "post-status", "cl-deployment", "True", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", capturedMethod)
	}
	if !strings.HasSuffix(capturedPath, "/clusters/c-001/statuses") {
		t.Errorf("path = %q, want suffix /clusters/c-001/statuses", capturedPath)
	}
}

func TestClusterAdapterPostStatus_PayloadShape(t *testing.T) {
	var body resource.AdapterStatusCreateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("cl-job", "False"))
	}))
	defer srv.Close()

	_, _, err := runCmdWithCluster(t, srv, "c-001", "cluster", "adapter", "post-status", "cl-job", "False", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body.Adapter != "cl-job" {
		t.Errorf("adapter = %q, want cl-job", body.Adapter)
	}
	if body.ObservedGeneration != 2 {
		t.Errorf("observed_generation = %d, want 2", body.ObservedGeneration)
	}
	if len(body.Conditions) != 3 {
		t.Fatalf("conditions len = %d, want 3", len(body.Conditions))
	}
	for _, c := range body.Conditions {
		if c.Status != "False" {
			t.Errorf("condition %q status = %q, want False", c.Type, c.Status)
		}
		if c.Reason != "ManualStatusPost" {
			t.Errorf("condition %q reason = %q, want ManualStatusPost", c.Type, c.Reason)
		}
		if c.LastTransitionTime == "" {
			t.Errorf("condition %q last_transition_time is empty", c.Type)
		}
	}
	if body.ObservedTime == "" {
		t.Error("observed_time must be set")
	}
}

func TestClusterAdapterPostStatus_DefaultGeneration(t *testing.T) {
	var body resource.AdapterStatusCreateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("cl-namespace", "Unknown"))
	}))
	defer srv.Close()

	// No generation arg — should default to 1
	_, _, err := runCmdWithCluster(t, srv, "c-001", "cluster", "adapter", "post-status", "cl-namespace", "Unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body.ObservedGeneration != 1 {
		t.Errorf("observed_generation = %d, want 1 (default)", body.ObservedGeneration)
	}
}

func TestClusterAdapterPostStatus_AllThreeConditionTypes(t *testing.T) {
	var body resource.AdapterStatusCreateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("cl-maestro", "True"))
	}))
	defer srv.Close()

	runCmdWithCluster(t, srv, "c-001", "cluster", "adapter", "post-status", "cl-maestro", "True") //nolint:errcheck

	types := map[string]bool{}
	for _, c := range body.Conditions {
		types[c.Type] = true
	}
	for _, want := range []string{"Available", "Applied", "Health"} {
		if !types[want] {
			t.Errorf("condition type %q missing from payload", want)
		}
	}
}

// ── nodepool adapter post-status ───────────────────────────────────────────────

func TestNodePoolAdapterPostStatus_PostsCorrectURL(t *testing.T) {
	var capturedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("np-configmap", "True"))
	}))
	defer srv.Close()

	_, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001",
		"nodepool", "adapter", "post-status", "np-configmap", "True", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantSuffix := "/clusters/c-001/nodepools/np-001/statuses"
	if !strings.HasSuffix(capturedPath, wantSuffix) {
		t.Errorf("path = %q, want suffix %q", capturedPath, wantSuffix)
	}
}

func TestNodePoolAdapterPostStatus_ExplicitNodePoolID(t *testing.T) {
	var capturedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("np-configmap", "True"))
	}))
	defer srv.Close()

	// Pass nodepool-id as explicit 4th arg (overrides state)
	_, _, err := runCmdWithCluster(t, srv, "c-001",
		"nodepool", "adapter", "post-status", "np-configmap", "True", "1", "np-explicit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantSuffix := "/clusters/c-001/nodepools/np-explicit/statuses"
	if !strings.HasSuffix(capturedPath, wantSuffix) {
		t.Errorf("path = %q, want suffix %q", capturedPath, wantSuffix)
	}
}

func TestNodePoolAdapterPostStatus_PayloadShape(t *testing.T) {
	var body resource.AdapterStatusCreateRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(adapterStatusJSON("np-configmap", "True"))
	}))
	defer srv.Close()

	runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", //nolint:errcheck
		"nodepool", "adapter", "post-status", "np-configmap", "True", "2")

	if body.Adapter != "np-configmap" {
		t.Errorf("adapter = %q, want np-configmap", body.Adapter)
	}
	if len(body.Conditions) != 3 {
		t.Fatalf("conditions len = %d, want 3", len(body.Conditions))
	}
	for _, c := range body.Conditions {
		if c.Status != "True" {
			t.Errorf("condition %q status = %q, want True", c.Type, c.Status)
		}
	}
}

// ── validation ─────────────────────────────────────────────────────────────────

func TestAdapterPostStatus_InvalidStatus_ReturnsError(t *testing.T) {
	serverCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()

	_, _, err := runCmdWithCluster(t, srv, "c-001", "cluster", "adapter", "post-status", "cl-deployment", "maybe")
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "True, False, Unknown") {
		t.Errorf("error message should list valid statuses, got: %v", err)
	}
	if serverCalled {
		t.Error("server must not be called when status is invalid")
	}
}

func TestAdapterPostStatus_InvalidGeneration_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()

	_, _, err := runCmdWithCluster(t, srv, "c-001", "cluster", "adapter", "post-status", "cl-deployment", "True", "notanint")
	if err == nil {
		t.Fatal("expected error for non-integer generation")
	}
	if !strings.Contains(err.Error(), "generation must be an integer") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdapterPostStatus_NoClusterID_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()

	// runCmd uses empty config dir → no cluster-id in state
	_, _, err := runCmd(t, srv, "cluster", "adapter", "post-status", "cl-deployment", "True")
	if err == nil {
		t.Fatal("expected error when no cluster-id set")
	}
	if !strings.Contains(err.Error(), "cluster-id") {
		t.Errorf("expected cluster-id error, got: %v", err)
	}
}

func TestNodePoolAdapterPostStatus_NoNodePoolID_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()

	// runCmdWithCluster sets cluster-id but not nodepool-id
	_, _, err := runCmdWithCluster(t, srv, "c-001", "nodepool", "adapter", "post-status", "np-configmap", "True")
	if err == nil {
		t.Fatal("expected error when no nodepool-id set")
	}
	if !strings.Contains(err.Error(), "nodepool-id") {
		t.Errorf("expected nodepool-id error, got: %v", err)
	}
}
