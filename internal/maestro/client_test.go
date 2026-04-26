package maestro

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resourceListJSON(items []Resource) []byte {
	b, _ := json.Marshal(listResponse[Resource]{Items: items, Kind: "ResourceBundleList", Total: len(items)})
	return b
}

func consumerListJSON(items []Consumer) []byte {
	b, _ := json.Marshal(listResponse[Consumer]{Items: items, Kind: "ConsumerList", Total: len(items)})
	return b
}

func TestClientList_FiltersConsumer(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		items := []Resource{{ID: "r-001", Metadata: map[string]string{"name": "mw-abc"}, ConsumerName: "cluster1"}}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resourceListJSON(items))
	}))
	defer srv.Close()

	c := New(srv.URL, "cluster1", "tok")
	resources, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedQuery, "search=") || !strings.Contains(capturedQuery, "consumer_name") {
		t.Errorf("expected consumer_name search filter in query, got: %s", capturedQuery)
	}
	if len(resources) != 1 || resources[0].Metadata["name"] != "mw-abc" {
		t.Errorf("unexpected resources: %+v", resources)
	}
}

func TestClientList_NoConsumer_OmitsQueryParam(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write(resourceListJSON(nil))
	}))
	defer srv.Close()

	c := New(srv.URL, "", "tok")
	_, err := c.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedQuery != "" {
		t.Errorf("expected no query params when consumer is empty, got: %s", capturedQuery)
	}
}

func TestClientGet_ReturnsResource(t *testing.T) {
	expected := Resource{ID: "r-002", Metadata: map[string]string{"name": "mw-xyz"}, ConsumerName: "cluster1", Version: 3}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/resource-bundles/mw-xyz") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(expected)
		w.Write(b)
	}))
	defer srv.Close()

	c := New(srv.URL, "cluster1", "")
	got, err := c.Get(context.Background(), "mw-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "r-002" || got.Version != 3 {
		t.Errorf("unexpected resource: %+v", got)
	}
}

func TestClientDelete_SendsDELETE(t *testing.T) {
	var capturedMethod, capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, "cluster1", "tok")
	err := c.Delete(context.Background(), "mw-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", capturedMethod)
	}
	if !strings.HasSuffix(capturedPath, "/resource-bundles/mw-abc") {
		t.Errorf("unexpected path: %s", capturedPath)
	}
}

func TestClientListBundles_ReturnsItems(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		items := []Resource{{ID: "b-001", Metadata: map[string]string{"name": "bundle-1"}}}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resourceListJSON(items))
	}))
	defer srv.Close()

	c := New(srv.URL, "cluster1", "")
	bundles, err := c.ListBundles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(capturedPath, "/resource-bundles") {
		t.Errorf("expected /resource-bundles path, got: %s", capturedPath)
	}
	if len(bundles) != 1 || bundles[0].Metadata["name"] != "bundle-1" {
		t.Errorf("unexpected bundles: %+v", bundles)
	}
}

func TestClientListConsumers_ReturnsItems(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		items := []Consumer{{ID: "c-001", Kind: "Consumer", Name: "cluster1"}}
		w.Header().Set("Content-Type", "application/json")
		w.Write(consumerListJSON(items))
	}))
	defer srv.Close()

	c := New(srv.URL, "", "")
	consumers, err := c.ListConsumers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(capturedPath, "/consumers") {
		t.Errorf("expected /consumers path, got: %s", capturedPath)
	}
	if len(consumers) != 1 || consumers[0].Name != "cluster1" {
		t.Errorf("unexpected consumers: %+v", consumers)
	}
}

func TestClientDo_NonSuccessStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"internal server error"}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "cluster1", "")
	_, err := c.ListConsumers(context.Background())
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("expected HTTP 500 in error, got: %v", err)
	}
}

func TestClientDo_AuthHeader_SendsBearer(t *testing.T) {
	var capturedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write(consumerListJSON(nil))
	}))
	defer srv.Close()

	c := New(srv.URL, "", "my-token")
	_, err := c.ListConsumers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAuth != "Bearer my-token" {
		t.Errorf("expected Authorization header 'Bearer my-token', got: %s", capturedAuth)
	}
}
