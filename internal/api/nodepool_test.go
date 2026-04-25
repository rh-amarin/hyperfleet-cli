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

func nodepoolList(nodepools []resource.NodePool) resource.ListResponse[resource.NodePool] {
	return resource.ListResponse[resource.NodePool]{
		Items: nodepools,
		Kind:  "NodePoolList",
		Page:  1,
		Size:  int32(len(nodepools)),
		Total: int32(len(nodepools)),
	}
}

func makeNodePool(id, name, deletedTime string) resource.NodePool {
	return resource.NodePool{
		ID:          id,
		Kind:        "NodePool",
		Name:        name,
		Generation:  1,
		DeletedTime: deletedTime,
	}
}

func TestFindNodePoolByName_ExactMatch(t *testing.T) {
	want := makeNodePool("np-001", "workers-1", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodepoolList([]resource.NodePool{want}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindNodePoolByName(c, context.Background(), "c-001", "workers-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}
	if got[0].ID != "np-001" {
		t.Errorf("id = %q, want np-001", got[0].ID)
	}
}

func TestFindNodePoolByName_FiltersDeletedNodePools(t *testing.T) {
	active := makeNodePool("np-001", "workers-1", "")
	deleted := makeNodePool("np-002", "workers-1", "2026-04-01T00:00:00Z")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodepoolList([]resource.NodePool{active, deleted}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindNodePoolByName(c, context.Background(), "c-001", "workers-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 non-deleted match, got %d", len(got))
	}
	if got[0].ID != "np-001" {
		t.Errorf("returned deleted nodepool instead of active one")
	}
}

func TestFindNodePoolByName_FiltersPartialNameMatches(t *testing.T) {
	exact := makeNodePool("np-001", "workers-1", "")
	partial := makeNodePool("np-002", "workers-10", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodepoolList([]resource.NodePool{exact, partial}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindNodePoolByName(c, context.Background(), "c-001", "workers-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 exact match, got %d: %v", len(got), got)
	}
	if got[0].ID != "np-001" {
		t.Errorf("returned partial-match nodepool instead of exact")
	}
}

func TestFindNodePoolByName_NoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodepoolList([]resource.NodePool{}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindNodePoolByName(c, context.Background(), "c-001", "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 matches, got %d", len(got))
	}
}

func TestFindNodePoolByName_MultipleActiveMatches(t *testing.T) {
	np1 := makeNodePool("np-001", "shared", "")
	np2 := makeNodePool("np-002", "shared", "")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodepoolList([]resource.NodePool{np1, np2}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	got, err := FindNodePoolByName(c, context.Background(), "c-001", "shared")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 matches, got %d", len(got))
	}
}

func TestFindNodePoolByName_SearchQueryEncodesClusterID(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path + "?" + r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodepoolList([]resource.NodePool{}))
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	_, _ = FindNodePoolByName(c, context.Background(), "c-001", "workers-1")

	if !strings.Contains(capturedPath, "c-001") {
		t.Errorf("path %q does not contain cluster ID c-001", capturedPath)
	}
	decoded, _ := url.QueryUnescape(capturedPath)
	if !strings.Contains(decoded, "name=") {
		t.Errorf("query %q does not contain name= filter", decoded)
	}
}

func TestFindNodePoolByName_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"status":401,"title":"Unauthorized","code":"AUTH-001"}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, "", false)
	_, err := FindNodePoolByName(c, context.Background(), "c-001", "any")
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
