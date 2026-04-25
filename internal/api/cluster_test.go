package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func clusterList(clusters []resource.Cluster) resource.ListResponse[resource.Cluster] {
	return resource.ListResponse[resource.Cluster]{
		Items: clusters,
		Kind:  "ClusterList",
		Page:  1,
		Size:  int32(len(clusters)),
		Total: int32(len(clusters)),
	}
}

func makeCluster(id, name, deletedTime string) resource.Cluster {
	return resource.Cluster{
		ID:          id,
		Kind:        "Cluster",
		Name:        name,
		Generation:  1,
		DeletedTime: deletedTime,
	}
}

func TestFindClusterByName_ExactMatch(t *testing.T) {
	want := makeCluster("c-001", "my-cluster", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{want}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindClusterByName(c, context.Background(), "my-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
	if got[0].ID != "c-001" {
		t.Errorf("id = %q, want c-001", got[0].ID)
	}
}

func TestFindClusterByName_FiltersDeletedClusters(t *testing.T) {
	active := makeCluster("c-001", "my-cluster", "")
	deleted := makeCluster("c-002", "my-cluster", "2026-04-01T00:00:00Z")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{active, deleted}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindClusterByName(c, context.Background(), "my-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 non-deleted match, got %d", len(got))
	}
	if got[0].ID != "c-001" {
		t.Errorf("returned deleted cluster instead of active one")
	}
}

func TestFindClusterByName_FiltersPartialNameMatches(t *testing.T) {
	// API may return clusters whose name contains the search term;
	// FindClusterByName must apply exact match client-side.
	exact := makeCluster("c-001", "my-cluster", "")
	partial := makeCluster("c-002", "my-cluster-extra", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{exact, partial}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindClusterByName(c, context.Background(), "my-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 exact match, got %d: %v", len(got), got)
	}
	if got[0].ID != "c-001" {
		t.Errorf("returned partial-match cluster instead of exact")
	}
}

func TestFindClusterByName_NoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindClusterByName(c, context.Background(), "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 matches, got %d", len(got))
	}
}

func TestFindClusterByName_MultipleActiveMatches(t *testing.T) {
	// Two active clusters with the same name (shouldn't happen in practice,
	// but the helper must return all of them and let the caller decide).
	c1 := makeCluster("c-001", "shared-name", "")
	c2 := makeCluster("c-002", "shared-name", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{c1, c2}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindClusterByName(c, context.Background(), "shared-name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 matches, got %d", len(got))
	}
}

func TestFindClusterByName_SearchQueryEncoded(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	_, _ = FindClusterByName(c, context.Background(), "my cluster")

	// The query should contain a URL-encoded search parameter.
	decoded, _ := url.QueryUnescape(capturedQuery)
	if !strings.Contains(decoded, "name=") {
		t.Errorf("search query %q does not contain name= filter", decoded)
	}
	if !strings.Contains(decoded, "my cluster") {
		t.Errorf("search query %q does not contain the cluster name", decoded)
	}
}

func TestFindClusterByName_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"status":401,"title":"Unauthorized","code":"AUTH-001"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	_, err := FindClusterByName(c, context.Background(), "any")
	if err == nil {
		t.Fatal("expected error from 401 response")
	}
	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 401 {
		t.Errorf("status = %d, want 401", apiErr.Status)
	}
}

func TestFindClusterByName_AllDeletedReturnsEmpty(t *testing.T) {
	// All items are soft-deleted; result should be empty.
	d1 := makeCluster("c-001", "stale", "2026-01-01T00:00:00Z")
	d2 := makeCluster("c-002", "stale", "2026-02-01T00:00:00Z")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(clusterList([]resource.Cluster{d1, d2}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindClusterByName(c, context.Background(), "stale")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 matches (all deleted), got %d", len(got))
	}
}
