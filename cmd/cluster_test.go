package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"gopkg.in/yaml.v3"
)

// runCmd executes the root command with the given args, capturing stdout.
// The test server URL and a per-test config dir are injected automatically.
func runCmd(t *testing.T, srv *httptest.Server, args ...string) (stdout string, stderr string, exitErr error) {
	t.Helper()

	cfgDir := t.TempDir()
	fullArgs := append(
		[]string{"--config", cfgDir, "--api-url", srv.URL},
		args...,
	)

	// Capture stdout via os.Pipe (Printer writes to os.Stdout directly).
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs(fullArgs)
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	var bufOut, bufErr bytes.Buffer
	bufOut.ReadFrom(rOut)
	bufErr.ReadFrom(rErr)

	return bufOut.String(), bufErr.String(), err
}

// clusterJSON returns a minimal cluster JSON response body.
func clusterJSON(id, name string, generation int32, deletedTime string) []byte {
	c := resource.Cluster{
		ID:          id,
		Kind:        "Cluster",
		Name:        name,
		Generation:  generation,
		Labels:      map[string]string{"counter": "1"},
		Spec:        map[string]any{"counter": "1", "region": "us-east-1", "version": "4.15.0"},
		Status:      resource.ClusterStatus{Conditions: []resource.ResourceCondition{{Type: "Ready", Status: "False", Reason: "MissingRequiredAdapters"}, {Type: "Available", Status: "False", Reason: "AdaptersNotAtSameGeneration"}}},
		CreatedBy:   "system@hyperfleet.local",
		CreatedTime: "2026-04-25T00:00:00Z",
		UpdatedBy:   "system@hyperfleet.local",
		UpdatedTime: "2026-04-25T00:00:00Z",
		DeletedTime: deletedTime,
	}
	b, _ := json.Marshal(c)
	return b
}

func clusterListJSON(clusters []resource.Cluster) []byte {
	l := resource.ListResponse[resource.Cluster]{
		Items: clusters,
		Kind:  "ClusterList",
		Page:  1,
		Size:  int32(len(clusters)),
		Total: int32(len(clusters)),
	}
	b, _ := json.Marshal(l)
	return b
}

// ── create ─────────────────────────────────────────────────────────────────────

func TestClusterCreate_DefaultArgs(t *testing.T) {
	var capturedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.RawQuery, "search="):
			// duplicate guard + post-create search: return empty first, then the created cluster
			if capturedBody == nil {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"items":[],"kind":"ClusterList","page":1,"size":0,"total":0}`)
			} else {
				c := resource.Cluster{ID: "c-new", Name: "my-cluster"}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, string(clusterListJSON([]resource.Cluster{c})))
			}
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			w.Write(clusterJSON("c-new", "my-cluster", 1, ""))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Defaults applied
	if capturedBody["name"] != "my-cluster" {
		t.Errorf("name = %v, want my-cluster", capturedBody["name"])
	}
	spec, _ := capturedBody["spec"].(map[string]any)
	if spec["region"] != "us-east-1" {
		t.Errorf("spec.region = %v, want us-east-1", spec["region"])
	}
	if spec["version"] != "4.15.0" {
		t.Errorf("spec.version = %v, want 4.15.0", spec["version"])
	}

	// kind field included
	if capturedBody["kind"] != "Cluster" {
		t.Errorf("kind = %v, want Cluster", capturedBody["kind"])
	}

	// Output is JSON with the cluster
	if !strings.Contains(stdout, `"kind"`) {
		t.Errorf("expected JSON cluster in stdout, got: %s", stdout)
	}
}

func TestClusterCreate_ExplicitArgs(t *testing.T) {
	var capturedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			fmt.Fprint(w, `{"items":[],"kind":"ClusterList","page":1,"size":0,"total":0}`)
		case r.Method == http.MethodPost:
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.WriteHeader(http.StatusCreated)
			w.Write(clusterJSON("c-001", "named", 1, ""))
		}
		w.Header().Set("Content-Type", "application/json")
	}))
	defer srv.Close()

	_, _, err := runCmd(t, srv, "cluster", "create", "named", "eu-west-1", "4.16.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spec, _ := capturedBody["spec"].(map[string]any)
	if spec["region"] != "eu-west-1" {
		t.Errorf("spec.region = %v, want eu-west-1", spec["region"])
	}
	if spec["version"] != "4.16.0" {
		t.Errorf("spec.version = %v, want 4.16.0", spec["version"])
	}
}

func TestClusterCreate_DuplicateGuard(t *testing.T) {
	existing := resource.Cluster{ID: "c-001", Name: "my-cluster"}
	postCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			fmt.Fprint(w, string(clusterListJSON([]resource.Cluster{existing})))
		case http.MethodPost:
			postCalled = true
			w.WriteHeader(http.StatusCreated)
			w.Write(clusterJSON("c-002", "my-cluster", 1, ""))
		}
	}))
	defer srv.Close()

	_, stderr, err := runCmd(t, srv, "cluster", "create", "my-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if postCalled {
		t.Error("POST should not have been called when cluster already exists")
	}
	if !strings.Contains(stderr, "[WARN]") || !strings.Contains(stderr, "already exists") {
		t.Errorf("expected [WARN] already exists, got: %s", stderr)
	}
}

func TestClusterCreate_PersistsClusterID(t *testing.T) {
	posted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			if !posted {
				// duplicate guard: no existing cluster
				fmt.Fprint(w, `{"items":[],"kind":"ClusterList","page":1,"size":0,"total":0}`)
			} else {
				// post-create search returns the created cluster
				c := resource.Cluster{ID: "c-persisted", Name: "my-cluster"}
				fmt.Fprint(w, string(clusterListJSON([]resource.Cluster{c})))
			}
		case r.Method == http.MethodPost:
			posted = true
			w.WriteHeader(http.StatusCreated)
			w.Write(clusterJSON("c-persisted", "my-cluster", 1, ""))
		}
	}))
	defer srv.Close()

	cfgDir := t.TempDir()
	fullArgs := []string{"--config", cfgDir, "--api-url", srv.URL, "cluster", "create"}

	rOut, wOut, _ := os.Pipe()
	oldOut := os.Stdout
	os.Stdout = wOut
	rootCmd.SetArgs(fullArgs)
	rootCmd.Execute() //nolint:errcheck
	wOut.Close()
	os.Stdout = oldOut
	rOut.Close()

	// Verify state.yaml was written with cluster-id
	stateData, err := os.ReadFile(cfgDir + "/state.yaml")
	if err != nil {
		t.Fatalf("state.yaml not created: %v", err)
	}
	var state map[string]string
	if err := yaml.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("parse state.yaml: %v", err)
	}
	if state["cluster-id"] != "c-persisted" {
		t.Errorf("cluster-id = %q, want c-persisted", state["cluster-id"])
	}
	if state["cluster-name"] != "my-cluster" {
		t.Errorf("cluster-name = %q, want my-cluster", state["cluster-name"])
	}
}

// ── get ────────────────────────────────────────────────────────────────────────

func TestClusterGet_OutputsJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(clusterJSON("c-001", "prod", 5, ""))
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "get", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got resource.Cluster
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\nout: %s", err, stdout)
	}
	if got.ID != "c-001" {
		t.Errorf("id = %q, want c-001", got.ID)
	}
	if got.Generation != 5 {
		t.Errorf("generation = %d, want 5", got.Generation)
	}
}

func TestClusterGet_NotFound_RFC7807(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"about:blank","title":"Resource Not Found","status":404,"code":"HYPERFLEET-NTF-001","detail":"Cluster not found"}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "get", "bad-id")
	// Should exit 0 even on 404 (passes RFC 7807 body to stdout)
	if err != nil {
		t.Fatalf("expected exit 0 on 404, got: %v", err)
	}
	if !strings.Contains(stdout, "HYPERFLEET-NTF-001") {
		t.Errorf("expected RFC 7807 code in stdout, got: %s", stdout)
	}
}

// ── search ─────────────────────────────────────────────────────────────────────

func TestClusterSearch_SetsClusterID(t *testing.T) {
	c := resource.Cluster{ID: "c-found", Name: "target"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(clusterListJSON([]resource.Cluster{c})))
	}))
	defer srv.Close()

	cfgDir := t.TempDir()
	rOut, wOut, _ := os.Pipe()
	oldOut := os.Stdout
	os.Stdout = wOut
	rootCmd.SetArgs([]string{"--config", cfgDir, "--api-url", srv.URL, "cluster", "search", "target"})
	rootCmd.Execute() //nolint:errcheck
	wOut.Close()
	os.Stdout = oldOut
	rOut.Close()

	stateData, _ := os.ReadFile(cfgDir + "/state.yaml")
	var state map[string]string
	yaml.Unmarshal(stateData, &state) //nolint:errcheck
	if state["cluster-id"] != "c-found" {
		t.Errorf("cluster-id = %q, want c-found", state["cluster-id"])
	}
}

func TestClusterSearch_NoMatch_WarnAndEmptyArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"kind":"ClusterList","page":1,"size":0,"total":0}`)
	}))
	defer srv.Close()

	stdout, stderr, err := runCmd(t, srv, "cluster", "search", "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "[WARN]") || !strings.Contains(stderr, "ghost") {
		t.Errorf("expected [WARN] with name, got: %s", stderr)
	}
	stdout = strings.TrimSpace(stdout)
	if stdout != "[]" {
		t.Errorf("expected empty array, got: %s", stdout)
	}
}

// ── patch ──────────────────────────────────────────────────────────────────────

func TestClusterPatch_NoArgs_ExitsNonZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, stderr, err := runCmd(t, srv, "cluster", "patch")
	if err == nil {
		t.Fatal("expected non-nil error when no target arg given")
	}
	if !strings.Contains(stderr, "Usage: hf cluster patch spec|labels") {
		t.Errorf("expected usage string in stderr, got: %s", stderr)
	}
}

func TestClusterPatch_SpecCounter_IncrementsAndPatches(t *testing.T) {
	var patchBody map[string]any
	current := resource.Cluster{
		ID: "c-001", Name: "cl", Generation: 2,
		Spec: map[string]any{"counter": "3", "region": "us-east-1"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(current)
		case http.MethodPatch:
			json.NewDecoder(r.Body).Decode(&patchBody)
			current.Generation++
			json.NewEncoder(w).Encode(current)
		}
	}))
	defer srv.Close()

	_, stderr, err := runCmd(t, srv, "cluster", "patch", "spec", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// [INFO] line reports the old and new counter values
	if !strings.Contains(stderr, "3 -> 4") {
		t.Errorf("expected [INFO] 3 -> 4, got: %s", stderr)
	}
	// PATCH body contains the incremented counter
	spec, _ := patchBody["spec"].(map[string]any)
	if spec["counter"] != "4" {
		t.Errorf("patch spec.counter = %v, want 4", spec["counter"])
	}
}

func TestClusterPatch_LabelsCounter_IncrementsAndPatches(t *testing.T) {
	var patchBody map[string]any
	current := resource.Cluster{
		ID: "c-001", Name: "cl", Generation: 1,
		Labels: map[string]string{"counter": "7", "env": "dev"},
		Spec:   map[string]any{"counter": "1"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(current)
		case http.MethodPatch:
			json.NewDecoder(r.Body).Decode(&patchBody)
			json.NewEncoder(w).Encode(current)
		}
	}))
	defer srv.Close()

	_, stderr, err := runCmd(t, srv, "cluster", "patch", "labels", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "7 -> 8") {
		t.Errorf("expected [INFO] 7 -> 8, got: %s", stderr)
	}
	labels, _ := patchBody["labels"].(map[string]any)
	if labels["counter"] != "8" {
		t.Errorf("patch labels.counter = %v, want 8", labels["counter"])
	}
	// Other label keys must be preserved
	if labels["env"] != "dev" {
		t.Errorf("patch labels.env = %v, want dev", labels["env"])
	}
}

// ── conditions ─────────────────────────────────────────────────────────────────

func TestClusterConditions_OutputsGenerationAndConditions(t *testing.T) {
	cl := resource.Cluster{
		ID: "c-001", Generation: 3,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Ready", Status: "False", Reason: "MissingRequiredAdapters"},
		}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cl)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "conditions", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\nout: %s", err, stdout)
	}
	if got["generation"] != float64(3) {
		t.Errorf("generation = %v, want 3", got["generation"])
	}
	status, _ := got["status"].(map[string]any)
	conds, _ := status["conditions"].([]any)
	if len(conds) != 1 {
		t.Errorf("conditions len = %d, want 1", len(conds))
	}
	// Cluster-level fields (name, labels, spec) must NOT appear in output
	if _, ok := got["name"]; ok {
		t.Error("conditions output should not include name field")
	}
}

func TestClusterConditionsTable_RendersRows(t *testing.T) {
	cl := resource.Cluster{
		ID: "c-001", Generation: 2,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Ready", Status: "True", Reason: "AllGood", LastTransitionTime: "2026-04-25T00:00:00Z"},
			{Type: "Available", Status: "False", Reason: "Waiting", LastTransitionTime: "2026-04-25T00:00:00Z"},
		}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cl)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "cluster", "conditions", "table", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Headers
	if !strings.Contains(stdout, "TYPE") || !strings.Contains(stdout, "STATUS") {
		t.Errorf("expected table headers, got:\n%s", stdout)
	}
	// Rows
	if !strings.Contains(stdout, "Ready") || !strings.Contains(stdout, "True") {
		t.Errorf("expected Ready/True row, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Available") || !strings.Contains(stdout, "False") {
		t.Errorf("expected Available/False row, got:\n%s", stdout)
	}
}

// ── statuses ───────────────────────────────────────────────────────────────────

func TestClusterStatuses_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"kind":"AdapterStatusList","page":1,"size":0,"total":0}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "statuses", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if got["kind"] != "AdapterStatusList" {
		t.Errorf("kind = %v, want AdapterStatusList", got["kind"])
	}
}

func TestClusterStatuses_404_TreatedAsEmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"status":404,"title":"Resource Not Found","code":"HYPERFLEET-NTF-001"}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "statuses", "c-001")
	if err != nil {
		t.Fatalf("expected exit 0 when statuses returns 404, got: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\nout: %s", err, stdout)
	}
	if got["kind"] != "AdapterStatusList" {
		t.Errorf("kind = %v, want AdapterStatusList", got["kind"])
	}
	items, _ := got["items"].([]any)
	if len(items) != 0 {
		t.Errorf("expected empty items, got %d", len(items))
	}
}

// ── delete ─────────────────────────────────────────────────────────────────────

func TestClusterDelete_OutputsDeletedCluster(t *testing.T) {
	deleted := clusterJSON("c-001", "bye-cluster", 4, "2026-04-25T10:00:00Z")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(deleted)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "cluster", "delete", "c-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got resource.Cluster
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if got.DeletedTime == "" {
		t.Error("expected deleted_time in response")
	}
	if got.Generation != 4 {
		t.Errorf("generation = %d, want 4", got.Generation)
	}
}
