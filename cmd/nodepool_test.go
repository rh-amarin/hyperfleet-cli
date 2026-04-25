package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"gopkg.in/yaml.v3"
	"os"
)

// nodepoolJSON returns a minimal NodePool JSON response.
func nodepoolJSON(id, name string, generation int32, deletedTime string) []byte {
	np := resource.NodePool{
		ID:          id,
		Kind:        "NodePool",
		Name:        name,
		Generation:  generation,
		Labels:      map[string]string{"counter": "1"},
		Spec:        map[string]any{"counter": "1", "platform": map[string]any{"type": "m4"}, "replicas": float64(1)},
		Status:      resource.NodePoolStatus{Conditions: []resource.ResourceCondition{{Type: "Ready", Status: "False", Reason: "MissingRequiredAdapters"}, {Type: "Available", Status: "False", Reason: "AdaptersNotAtSameGeneration"}}},
		DeletedTime: deletedTime,
		OwnerReferences: resource.ObjectReference{ID: "c-001", Kind: "Cluster", Href: "/api/hyperfleet/v1/clusters/c-001"},
	}
	b, _ := json.Marshal(np)
	return b
}

func nodepoolListJSON(nodepools []resource.NodePool) []byte {
	l := resource.ListResponse[resource.NodePool]{
		Items: nodepools,
		Kind:  "NodePoolList",
		Page:  1,
		Size:  int32(len(nodepools)),
		Total: int32(len(nodepools)),
	}
	b, _ := json.Marshal(l)
	return b
}

// runCmdWithCluster runs a command with a pre-configured cluster-id in state.
func runCmdWithCluster(t *testing.T, srv *httptest.Server, clusterID string, args ...string) (string, string, error) {
	t.Helper()
	cfgDir := t.TempDir()

	// Write state.yaml with cluster-id pre-set
	stateContent := fmt.Sprintf("cluster-id: %s\ncluster-name: test-cluster\n", clusterID)
	if err := os.WriteFile(cfgDir+"/state.yaml", []byte(stateContent), 0600); err != nil {
		t.Fatalf("could not write state.yaml: %v", err)
	}

	fullArgs := append([]string{"--config", cfgDir, "--api-url", srv.URL}, args...)
	return runCmdRaw(t, fullArgs)
}

// runCmdWithClusterAndNodePool runs a command with both cluster-id and nodepool-id pre-set.
func runCmdWithClusterAndNodePool(t *testing.T, srv *httptest.Server, clusterID, nodepoolID string, args ...string) (string, string, error) {
	t.Helper()
	cfgDir := t.TempDir()

	stateContent := fmt.Sprintf("cluster-id: %s\ncluster-name: test-cluster\nnodepool-id: %s\n", clusterID, nodepoolID)
	if err := os.WriteFile(cfgDir+"/state.yaml", []byte(stateContent), 0600); err != nil {
		t.Fatalf("could not write state.yaml: %v", err)
	}

	fullArgs := append([]string{"--config", cfgDir, "--api-url", srv.URL}, args...)
	return runCmdRaw(t, fullArgs)
}

// runCmdRaw executes rootCmd with the given full args list, capturing stdout/stderr.
func runCmdRaw(t *testing.T, fullArgs []string) (string, string, error) {
	t.Helper()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	rootCmd.SetArgs(fullArgs)
	err := rootCmd.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr

	var bufOut, bufErr strings.Builder
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
		bufErr.Write(buf[:n])
	}
	return bufOut.String(), bufErr.String(), err
}

// ── create ─────────────────────────────────────────────────────────────────────

func TestNodePoolCreate_DefaultArgs(t *testing.T) {
	var capturedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&capturedBody)
			w.WriteHeader(http.StatusCreated)
			w.Write(nodepoolJSON("np-001", "my-nodepool-1", 1, ""))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithCluster(t, srv, "c-001", "nodepool", "create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody["name"] != "my-nodepool-1" {
		t.Errorf("name = %v, want my-nodepool-1", capturedBody["name"])
	}
	spec, _ := capturedBody["spec"].(map[string]any)
	platform, _ := spec["platform"].(map[string]any)
	if platform["type"] != "m4" {
		t.Errorf("platform.type = %v, want m4", platform["type"])
	}
	if capturedBody["kind"] != "NodePool" {
		t.Errorf("kind = %v, want NodePool", capturedBody["kind"])
	}
	if !strings.Contains(stdout, `"kind"`) {
		t.Errorf("expected JSON nodepool in stdout, got: %s", stdout)
	}
}

func TestNodePoolCreate_MultipleNodePools(t *testing.T) {
	var postCount int
	var lastBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			postCount++
			json.NewDecoder(r.Body).Decode(&lastBody)
			name, _ := lastBody["name"].(string)
			w.WriteHeader(http.StatusCreated)
			w.Write(nodepoolJSON(fmt.Sprintf("np-%03d", postCount), name, 1, ""))
		}
	}))
	defer srv.Close()

	cfgDir := t.TempDir()
	stateContent := "cluster-id: c-001\ncluster-name: test-cluster\n"
	os.WriteFile(cfgDir+"/state.yaml", []byte(stateContent), 0600)

	_, _, err := runCmdRaw(t, []string{"--config", cfgDir, "--api-url", srv.URL, "nodepool", "create", "workers", "2", "n2-standard-4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if postCount != 2 {
		t.Errorf("expected 2 POST calls, got %d", postCount)
	}
	if lastBody["name"] != "workers-2" {
		t.Errorf("last nodepool name = %v, want workers-2", lastBody["name"])
	}

	// Verify last nodepool-id was persisted
	stateData, err := os.ReadFile(cfgDir + "/state.yaml")
	if err != nil {
		t.Fatalf("state.yaml not readable: %v", err)
	}
	var state map[string]string
	yaml.Unmarshal(stateData, &state)
	if state["nodepool-id"] != "np-002" {
		t.Errorf("nodepool-id = %q, want np-002", state["nodepool-id"])
	}
}

// ── list ───────────────────────────────────────────────────────────────────────

func TestNodePoolList_OutputsJSON(t *testing.T) {
	nps := []resource.NodePool{
		{ID: "np-001", Kind: "NodePool", Name: "workers-1"},
		{ID: "np-002", Kind: "NodePool", Name: "workers-2"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nodepoolListJSON(nps))
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithCluster(t, srv, "c-001", "nodepool", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\nout: %s", err, stdout)
	}
	if got["kind"] != "NodePoolList" {
		t.Errorf("kind = %v, want NodePoolList", got["kind"])
	}
}

// ── search ─────────────────────────────────────────────────────────────────────

func TestNodePoolSearch_SetsNodePoolID(t *testing.T) {
	np := resource.NodePool{ID: "np-found", Name: "workers-1"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(nodepoolListJSON([]resource.NodePool{np})))
	}))
	defer srv.Close()

	cfgDir := t.TempDir()
	stateContent := "cluster-id: c-001\ncluster-name: test-cluster\n"
	os.WriteFile(cfgDir+"/state.yaml", []byte(stateContent), 0600)

	_, _, err := runCmdRaw(t, []string{"--config", cfgDir, "--api-url", srv.URL, "nodepool", "search", "workers-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stateData, _ := os.ReadFile(cfgDir + "/state.yaml")
	var state map[string]string
	yaml.Unmarshal(stateData, &state)
	if state["nodepool-id"] != "np-found" {
		t.Errorf("nodepool-id = %q, want np-found", state["nodepool-id"])
	}
}

func TestNodePoolSearch_NoMatch_WarnAndEmptyArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
	}))
	defer srv.Close()

	stdout, stderr, err := runCmdWithCluster(t, srv, "c-001", "nodepool", "search", "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "[WARN]") || !strings.Contains(stderr, "ghost") {
		t.Errorf("expected [WARN] with name, got: %s", stderr)
	}
	if strings.TrimSpace(stdout) != "[]" {
		t.Errorf("expected empty array, got: %s", stdout)
	}
}

// ── get ────────────────────────────────────────────────────────────────────────

func TestNodePoolGet_OutputsJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nodepoolJSON("np-001", "workers-1", 3, ""))
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got resource.NodePool
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\nout: %s", err, stdout)
	}
	if got.ID != "np-001" {
		t.Errorf("id = %q, want np-001", got.ID)
	}
	if got.Generation != 3 {
		t.Errorf("generation = %d, want 3", got.Generation)
	}
}

// ── patch ──────────────────────────────────────────────────────────────────────

func TestNodePoolPatch_NoArgs_ExitsNonZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, stderr, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "patch")
	if err == nil {
		t.Fatal("expected non-nil error when no target arg given")
	}
	if !strings.Contains(stderr, "Usage: hf nodepool patch spec|labels") {
		t.Errorf("expected usage string in stderr, got: %s", stderr)
	}
}

func TestNodePoolPatch_SpecCounter_IncrementsAndPatches(t *testing.T) {
	var patchBody map[string]any
	current := resource.NodePool{
		ID: "np-001", Name: "workers-1", Generation: 2,
		Spec: map[string]any{"counter": "5", "platform": map[string]any{"type": "m4"}, "replicas": float64(1)},
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

	_, stderr, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "patch", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "5 -> 6") {
		t.Errorf("expected [INFO] 5 -> 6, got: %s", stderr)
	}
	spec, _ := patchBody["spec"].(map[string]any)
	if spec["counter"] != "6" {
		t.Errorf("patch spec.counter = %v, want 6", spec["counter"])
	}
}

func TestNodePoolPatch_LabelsCounter_IncrementsAndPatches(t *testing.T) {
	var patchBody map[string]any
	current := resource.NodePool{
		ID: "np-001", Name: "workers-1", Generation: 1,
		Labels: map[string]string{"counter": "3", "env": "dev"},
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

	_, stderr, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "patch", "labels")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "3 -> 4") {
		t.Errorf("expected [INFO] 3 -> 4, got: %s", stderr)
	}
	labels, _ := patchBody["labels"].(map[string]any)
	if labels["counter"] != "4" {
		t.Errorf("patch labels.counter = %v, want 4", labels["counter"])
	}
	if labels["env"] != "dev" {
		t.Errorf("patch labels.env = %v, want dev", labels["env"])
	}
}

// ── delete ─────────────────────────────────────────────────────────────────────

func TestNodePoolDelete_OutputsDeletedNodePool(t *testing.T) {
	deleted := nodepoolJSON("np-001", "workers-1", 5, "2026-04-25T10:00:00Z")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(deleted)
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got resource.NodePool
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if got.DeletedTime == "" {
		t.Error("expected deleted_time in response")
	}
	if got.Generation != 5 {
		t.Errorf("generation = %d, want 5", got.Generation)
	}
}

// ── conditions ─────────────────────────────────────────────────────────────────

func TestNodePoolConditions_OutputsGenerationAndConditions(t *testing.T) {
	np := resource.NodePool{
		ID: "np-001", Generation: 4,
		Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
			{Type: "Ready", Status: "False", Reason: "MissingRequiredAdapters"},
		}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(np)
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "conditions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("output not valid JSON: %v\nout: %s", err, stdout)
	}
	if got["generation"] != float64(4) {
		t.Errorf("generation = %v, want 4", got["generation"])
	}
	if _, ok := got["name"]; ok {
		t.Error("conditions output should not include name field")
	}
}

func TestNodePoolConditionsTable_RendersRows(t *testing.T) {
	np := resource.NodePool{
		ID: "np-001", Generation: 1,
		Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
			{Type: "Ready", Status: "False", Reason: "MissingRequiredAdapters", LastTransitionTime: "2026-04-25T00:00:00Z"},
			{Type: "Available", Status: "True", Reason: "AllGood", LastTransitionTime: "2026-04-25T00:00:00Z"},
		}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(np)
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "--no-color", "nodepool", "conditions", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "TYPE") || !strings.Contains(stdout, "STATUS") {
		t.Errorf("expected table headers, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Ready") || !strings.Contains(stdout, "Available") {
		t.Errorf("expected condition rows, got:\n%s", stdout)
	}
}

// ── statuses ───────────────────────────────────────────────────────────────────

func TestNodePoolStatuses_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"kind":"AdapterStatusList","page":1,"size":0,"total":0}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "statuses")
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

func TestNodePoolStatuses_404_TreatedAsEmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"status":404,"title":"Resource Not Found","code":"HYPERFLEET-NTF-001"}`)
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "nodepool", "statuses")
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

// ── table ──────────────────────────────────────────────────────────────────────

func TestNodePoolTable_RendersWithDynamicColumns(t *testing.T) {
	nps := []resource.NodePool{
		{
			ID: "np-001", Name: "workers-1", Generation: 2,
			Spec: map[string]any{"platform": map[string]any{"type": "m4"}, "replicas": float64(1)},
			Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "True"},
				{Type: "Ready", Status: "True"},
			}},
		},
		{
			ID: "np-002", Name: "workers-2", Generation: 1,
			Spec: map[string]any{"platform": map[string]any{"type": "m4"}, "replicas": float64(1)},
			Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "False"},
				{Type: "Ready", Status: "False"},
			}},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(nodepoolListJSON(nps))
	}))
	defer srv.Close()

	stdout, _, err := runCmdWithCluster(t, srv, "c-001", "--no-color", "nodepool", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Headers (PrintTable uppercases all headers)
	for _, h := range []string{"ID", "NAME", "REPLICAS", "TYPE", "GEN", "AVAILABLE", "READY"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q in output, got:\n%s", h, stdout)
		}
	}
	// Available comes before Ready in dynamic columns
	availIdx := strings.Index(stdout, "AVAILABLE")
	readyIdx := strings.Index(stdout, "READY")
	if availIdx == -1 || readyIdx == -1 || availIdx >= readyIdx {
		t.Errorf("expected Available before Ready in headers, got:\n%s", stdout)
	}
	// Row data
	if !strings.Contains(stdout, "workers-1") || !strings.Contains(stdout, "workers-2") {
		t.Errorf("expected nodepool names in output, got:\n%s", stdout)
	}
}
