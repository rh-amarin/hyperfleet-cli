package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// ── hf cluster table ──────────────────────────────────────────────────────────

func TestClusterTable_RendersWithDynamicColumns(t *testing.T) {
	clusters := []resource.Cluster{
		{
			ID: "c-001", Name: "prod", Generation: 3,
			Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "True"},
				{Type: "Ready", Status: "True"},
			}},
		},
		{
			ID: "c-002", Name: "staging", Generation: 1,
			Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "False"},
				{Type: "Ready", Status: "False"},
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

	for _, h := range []string{"NAME", "GEN", "AVAILABLE", "READY"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q, got:\n%s", h, stdout)
		}
	}

	// Available must appear before Ready
	availIdx := strings.Index(stdout, "AVAILABLE")
	readyIdx := strings.Index(stdout, "READY")
	if availIdx == -1 || readyIdx == -1 || availIdx >= readyIdx {
		t.Errorf("expected AVAILABLE before READY, got:\n%s", stdout)
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
			{Type: "Ready", Status: "True"},
		}},
	}
	nps := []resource.NodePool{
		{
			ID: "np-001", Name: "workers-1", Generation: 1,
			Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "False"},
				{Type: "Ready", Status: "False"},
			}},
		},
		{
			ID: "np-002", Name: "workers-2", Generation: 1,
			Status: resource.NodePoolStatus{Conditions: []resource.ResourceCondition{
				{Type: "Available", Status: "True"},
				{Type: "Ready", Status: "True"},
			}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "" && strings.Contains(r.URL.Path, "/nodepools") {
			w.Write(nodepoolListJSON(nps))
		} else {
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Headers
	for _, h := range []string{"NAME", "KIND", "CLUSTER", "GEN"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q, got:\n%s", h, stdout)
		}
	}

	// Cluster and nodepool rows
	if !strings.Contains(stdout, "prod") {
		t.Errorf("expected cluster name 'prod' in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "workers-1") || !strings.Contains(stdout, "workers-2") {
		t.Errorf("expected nodepool names in output, got:\n%s", stdout)
	}

	// KIND column values
	if !strings.Contains(stdout, "Cluster") || !strings.Contains(stdout, "NodePool") {
		t.Errorf("expected KIND values in output, got:\n%s", stdout)
	}

	// Available before Ready
	availIdx := strings.Index(stdout, "AVAILABLE")
	readyIdx := strings.Index(stdout, "READY")
	if availIdx == -1 || readyIdx == -1 || availIdx >= readyIdx {
		t.Errorf("expected AVAILABLE before READY, got:\n%s", stdout)
	}
}

func TestTable_EmptyCluster_NoNodePoolRows(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "empty-cluster", Generation: 1,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Available", Status: "False"},
			{Type: "Ready", Status: "False"},
		}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/nodepools") {
			w.Write(nodepoolListJSON(nil))
		} else {
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
	if strings.Contains(stdout, "NodePool") {
		t.Errorf("expected no NodePool rows, got:\n%s", stdout)
	}
}

func TestTable_AdapterConditionsAlphabetical(t *testing.T) {
	cluster := resource.Cluster{
		ID: "c-001", Name: "cl", Generation: 1,
		Status: resource.ClusterStatus{Conditions: []resource.ResourceCondition{
			{Type: "Zebra", Status: "True"},
			{Type: "Available", Status: "True"},
			{Type: "Alpha", Status: "True"},
			{Type: "Ready", Status: "True"},
		}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/nodepools") {
			fmt.Fprint(w, `{"items":[],"kind":"NodePoolList","page":1,"size":0,"total":0}`)
		} else {
			w.Write(clusterListJSON([]resource.Cluster{cluster}))
		}
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "table")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Column order in header line: AVAILABLE, then alpha middle, then READY
	header := strings.SplitN(stdout, "\n", 2)[0]
	availIdx := strings.Index(header, "AVAILABLE")
	alphaIdx := strings.Index(header, "ALPHA")
	zebraIdx := strings.Index(header, "ZEBRA")
	readyIdx := strings.Index(header, "READY")

	if availIdx == -1 || alphaIdx == -1 || zebraIdx == -1 || readyIdx == -1 {
		t.Fatalf("missing column in header: %s", header)
	}
	if !(availIdx < alphaIdx && alphaIdx < zebraIdx && zebraIdx < readyIdx) {
		t.Errorf("wrong column order in header: %s", header)
	}
}

