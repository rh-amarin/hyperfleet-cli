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

// adapterStatusListJSON builds a JSON body for a list of AdapterStatus objects.
func adapterStatusListJSON(statuses []resource.AdapterStatus) []byte {
	l := resource.ListResponse[resource.AdapterStatus]{
		Items: statuses,
		Kind:  "AdapterStatusList",
		Page:  1,
		Size:  int32(len(statuses)),
		Total: int32(len(statuses)),
	}
	b, _ := json.Marshal(l)
	return b
}

var emptyAdapterStatusList = `{"items":[],"kind":"AdapterStatusList","page":1,"size":0,"total":0}`

// ── hf cluster table ──────────────────────────────────────────────────────────

func TestClusterTable_RendersWithDynamicColumns(t *testing.T) {
	clusters := []resource.Cluster{
		{
			ID: "c-001", Name: "prod", Generation: 3,
			Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "True"},
				{Type: "Reconciled", Status: "True"},
			}},
		},
		{
			ID: "c-002", Name: "staging", Generation: 1,
			Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "False"},
				{Type: "Reconciled", Status: "False"},
			}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(clusterListJSON(clusters))
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "cluster", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, h := range []string{"NAME", "GEN", "AVAILABLE", "RECONCILED"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q, got:\n%s", h, stdout)
		}
	}

	// Available must appear before Reconciled
	availIdx := strings.Index(stdout, "AVAILABLE")
	reconciledIdx := strings.Index(stdout, "RECONCILED")
	if availIdx == -1 || reconciledIdx == -1 || availIdx >= reconciledIdx {
		t.Errorf("expected AVAILABLE before RECONCILED, got:\n%s", stdout)
	}

	// Both cluster names must appear
	if !strings.Contains(stdout, "prod") || !strings.Contains(stdout, "staging") {
		t.Errorf("expected cluster names in output, got:\n%s", stdout)
	}
}

func TestClusterTable_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[],"kind":"ClusterList","page":1,"size":0,"total":0}`)
	}))
	defer srv.Close()

	_, _, err := runCmd(t, srv, "--no-color", "cluster", "table")
	if err != nil {
		t.Fatalf("unexpected error on empty cluster list: %v", err)
	}
}

// ── hf table (combined) ───────────────────────────────────────────────────────

func TestTable_RendersClusterAndNodePoolRows(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "prod", Generation: 2,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "True"},
			{Type: "Reconciled", Status: "True"},
		}},
	}
	nps := []resource.NodePool{
		{
			ID: "np-001", Name: "workers-1", Generation: 1,
			Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "False"},
				{Type: "Reconciled", Status: "False"},
			}},
		},
		{
			ID: "np-002", Name: "workers-2", Generation: 1,
			Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "True"},
				{Type: "Reconciled", Status: "True"},
			}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			fmt.Fprint(w, emptyAdapterStatusList)
		case strings.Contains(r.URL.Path, "/nodepools"):
			w.Write(nodepoolListJSON(nps))
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Required headers — ID must be present, KIND and CLUSTER must not appear
	for _, h := range []string{"ID", "NAME", "GEN"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q, got:\n%s", h, stdout)
		}
	}
	for _, banned := range []string{"KIND", "CLUSTER"} {
		if strings.Contains(strings.SplitN(stdout, "\n", 2)[0], banned) {
			t.Errorf("unexpected header %q in output:\n%s", banned, stdout)
		}
	}

	// Cluster row: full ID and name
	if !strings.Contains(stdout, "c-001") {
		t.Errorf("expected cluster ID 'c-001' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "prod") {
		t.Errorf("expected cluster name 'prod' in output, got:\n%s", stdout)
	}

	// Nodepool rows: names must appear indented (prefixed with spaces)
	if !strings.Contains(stdout, "  np-001") {
		t.Errorf("expected indented nodepool ID '  np-001' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "  workers-1") || !strings.Contains(stdout, "  workers-2") {
		t.Errorf("expected indented nodepool names in output, got:\n%s", stdout)
	}

	// Available before Reconciled
	availIdx := strings.Index(stdout, "AVAILABLE")
	reconciledIdx := strings.Index(stdout, "RECONCILED")
	if availIdx == -1 || reconciledIdx == -1 || availIdx >= reconciledIdx {
		t.Errorf("expected AVAILABLE before RECONCILED, got:\n%s", stdout)
	}
}

func TestTable_EmptyCluster_NoNodePoolRows(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "empty-cluster", Generation: 1,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "False"},
			{Type: "Reconciled", Status: "False"},
		}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			fmt.Fprint(w, emptyAdapterStatusList)
		case strings.Contains(r.URL.Path, "/nodepools"):
			w.Write(nodepoolListJSON(nil))
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "empty-cluster") {
		t.Errorf("expected cluster row in output, got:\n%s", stdout)
	}
	// No indented rows (nodepool list is empty)
	if strings.Contains(stdout, "  ") {
		lines := strings.Split(stdout, "\n")
		for _, l := range lines[2:] { // skip header + separator
			if strings.HasPrefix(l, "  ") && strings.TrimSpace(l) != "" {
				t.Errorf("unexpected indented row in output:\n%s", stdout)
				break
			}
		}
	}
}

func TestTable_ExcludesSuccessfulColumns(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "cl", Generation: 2,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "True", ObservedGeneration: 2},
			{Type: "ClDeploymentSuccessful", Status: "True", ObservedGeneration: 2},
			{Type: "NpConfigmapSuccessful", Status: "True", ObservedGeneration: 2},
			{Type: "Reconciled", Status: "True", ObservedGeneration: 2},
		}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			fmt.Fprint(w, emptyAdapterStatusList)
		case strings.Contains(r.URL.Path, "/nodepools"):
			fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := strings.SplitN(stdout, "\n", 2)[0]
	if strings.Contains(header, "SUCCESSFUL") {
		t.Errorf("expected no *Successful columns in header, got:\n%s", header)
	}
	if !strings.Contains(header, "AVAILABLE") {
		t.Errorf("expected AVAILABLE column, got:\n%s", header)
	}
	if !strings.Contains(header, "RECONCILED") {
		t.Errorf("expected RECONCILED column, got:\n%s", header)
	}
}

func TestTable_DotRendersWithGenerationSuffix(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "cl", Generation: 3,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "True", ObservedGeneration: 3},
			{Type: "Reconciled", Status: "False", ObservedGeneration: 3},
		}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			fmt.Fprint(w, emptyAdapterStatusList)
		case strings.Contains(r.URL.Path, "/nodepools"):
			fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With --no-color the dot is rendered as text + gen suffix e.g. "True 3"
	if !strings.Contains(stdout, "True 3") {
		t.Errorf("expected dot cell 'True 3' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "False 3") {
		t.Errorf("expected dot cell 'False 3' in output, got:\n%s", stdout)
	}
}

func TestTable_AdapterConditionsAlphabetical(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "cl", Generation: 1,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Zebra", Status: "True"},
			{Type: "Available", Status: "True"},
			{Type: "Alpha", Status: "True"},
			{Type: "Reconciled", Status: "True"},
		}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			fmt.Fprint(w, emptyAdapterStatusList)
		case strings.Contains(r.URL.Path, "/nodepools"):
			fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Column order in header line: AVAILABLE, then alpha middle, then RECONCILED
	header := strings.SplitN(stdout, "\n", 2)[0]
	availIdx := strings.Index(header, "AVAILABLE")
	alphaIdx := strings.Index(header, "ALPHA")
	zebraIdx := strings.Index(header, "ZEBRA")
	reconciledIdx := strings.Index(header, "RECONCILED")

	if availIdx == -1 || alphaIdx == -1 || zebraIdx == -1 || reconciledIdx == -1 {
		t.Fatalf("missing column in header: %s", header)
	}
	if !(availIdx < alphaIdx && alphaIdx < zebraIdx && zebraIdx < reconciledIdx) {
		t.Errorf("wrong column order in header: %s", header)
	}
}

// TestTable_AdapterColumns verifies that adapter names appear as columns and that
// the Available condition value is shown for non-deleted resources.
func TestTable_AdapterColumns(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "prod", Generation: 2,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "True", ObservedGeneration: 2},
		}},
	}

	statuses := []resource.AdapterStatus{
		{
			Adapter:            "cl-deployment",
			ObservedGeneration: 2,
			Conditions: []resource.AdapterCondition{
				{Type: "Available", Status: "True"},
				{Type: "Finalized", Status: "False"},
			},
		},
		{
			Adapter:            "cl-namespace",
			ObservedGeneration: 2,
			Conditions: []resource.AdapterCondition{
				{Type: "Available", Status: "False"},
				{Type: "Finalized", Status: "True"},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			w.Write(adapterStatusListJSON(statuses))
		case strings.Contains(r.URL.Path, "/nodepools"):
			fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := strings.SplitN(stdout, "\n", 2)[0]

	// Adapter names must appear as column headers (uppercased by PrintTable)
	if !strings.Contains(header, "CL-DEPLOYMENT") {
		t.Errorf("expected CL-DEPLOYMENT column in header, got:\n%s", header)
	}
	if !strings.Contains(header, "CL-NAMESPACE") {
		t.Errorf("expected CL-NAMESPACE column in header, got:\n%s", header)
	}

	// cl-deployment Available=True → "True 2"; cl-namespace Available=False → "False 2"
	if !strings.Contains(stdout, "True 2") {
		t.Errorf("expected adapter Available=True rendered as 'True 2', got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "False 2") {
		t.Errorf("expected adapter Available=False rendered as 'False 2', got:\n%s", stdout)
	}
}

// TestTable_DeletionMarker verifies that when deleted_time is set on a cluster:
//   - the GEN cell contains ❌
//   - the adapter columns use the Finalized condition (not Available)
func TestTable_DeletionMarker(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-del", Name: "dying", Generation: 4,
		DeletedTime: "2026-05-01T00:00:00Z",
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "False", ObservedGeneration: 4},
		}},
	}

	statuses := []resource.AdapterStatus{
		{
			Adapter:            "cl-deployment",
			ObservedGeneration: 4,
			Conditions: []resource.AdapterCondition{
				{Type: "Available", Status: "False"},
				{Type: "Finalized", Status: "True"},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/statuses"):
			w.Write(adapterStatusListJSON(statuses))
		case strings.Contains(r.URL.Path, "/nodepools"):
			fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
		default:
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// GEN cell must contain ❌
	if !strings.Contains(stdout, "❌") {
		t.Errorf("expected deletion marker ❌ in GEN cell, got:\n%s", stdout)
	}

	// Adapter must use Finalized=True → "True 4", not Available=False → "False 4"
	// But "False 4" could appear from the condition column — check that Finalized is used
	// by verifying the row has "True 4" for the adapter (Finalized=True).
	if !strings.Contains(stdout, "True 4") {
		t.Errorf("expected Finalized=True rendered as 'True 4' for deleted cluster, got:\n%s", stdout)
	}
}
