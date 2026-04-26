package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ps "github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// fakeGCPClient implements ps.GCPPublisher for testing without a live GCP connection.
type fakeGCPClient struct {
	subs           []string
	publishedTopic string
	publishedData  []byte
	publishErr     error
	listErr        error
}

func (f *fakeGCPClient) ListSubscriptions(_ context.Context, filter string) ([]string, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if filter == "" {
		return f.subs, nil
	}
	var filtered []string
	for _, s := range f.subs {
		if strings.Contains(s, filter) {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func (f *fakeGCPClient) Publish(_ context.Context, topicID string, data []byte) (string, error) {
	if f.publishErr != nil {
		return "", f.publishErr
	}
	f.publishedTopic = topicID
	f.publishedData = data
	return "msg-fake-id", nil
}

func (f *fakeGCPClient) Close() error { return nil }

var _ ps.GCPPublisher = (*fakeGCPClient)(nil)

// ── list ──────────────────────────────────────────────────────────────────────

func TestPubSubList_PrintsTable(t *testing.T) {
	fake := &fakeGCPClient{
		subs: []string{
			"projects/my-project/subscriptions/sub-alpha",
			"projects/my-project/subscriptions/sub-beta",
		},
	}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "pubsub", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "SUBSCRIPTION") {
		t.Errorf("expected SUBSCRIPTION header, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "sub-alpha") {
		t.Errorf("expected sub-alpha in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "sub-beta") {
		t.Errorf("expected sub-beta in output, got:\n%s", stdout)
	}
}

func TestPubSubList_WithFilter(t *testing.T) {
	fake := &fakeGCPClient{
		subs: []string{
			"projects/my-project/subscriptions/sub-alpha",
			"projects/my-project/subscriptions/sub-beta",
		},
	}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stdout, _, err := runCmd(t, srv, "--no-color", "pubsub", "list", "alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "sub-alpha") {
		t.Errorf("expected sub-alpha in output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "sub-beta") {
		t.Errorf("sub-beta should be filtered out, got:\n%s", stdout)
	}
}

// ── publish cluster ───────────────────────────────────────────────────────────

func TestPubSubPublishCluster_PublishesCloudEvent(t *testing.T) {
	fake := &fakeGCPClient{}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	cluster := resource.Cluster{ID: "c-001", Name: "prod", Generation: 3}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cluster)
	}))
	defer srv.Close()

	_, stderr, err := runCmdWithCluster(t, srv, "c-001", "pubsub", "publish", "cluster", "my-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.publishedTopic != "my-topic" {
		t.Errorf("topic = %q, want my-topic", fake.publishedTopic)
	}

	var event map[string]any
	if err := json.Unmarshal(fake.publishedData, &event); err != nil {
		t.Fatalf("published data is not valid JSON: %v", err)
	}
	if event["type"] != "com.hyperfleet.cluster.changed" {
		t.Errorf("event.type = %v, want com.hyperfleet.cluster.changed", event["type"])
	}
	if event["specversion"] != "1.0" {
		t.Errorf("event.specversion = %v, want 1.0", event["specversion"])
	}
	if event["source"] != "/hyperfleet/cli" {
		t.Errorf("event.source = %v, want /hyperfleet/cli", event["source"])
	}
	if event["id"] == "" {
		t.Error("event.id must not be empty")
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got: %s", stderr)
	}
}

// ── publish nodepool ──────────────────────────────────────────────────────────

func TestPubSubPublishNodepool_PublishesCloudEvent(t *testing.T) {
	fake := &fakeGCPClient{}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	np := resource.NodePool{ID: "np-001", Name: "workers-1", Generation: 2}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(np)
	}))
	defer srv.Close()

	_, stderr, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "pubsub", "publish", "nodepool", "np-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.publishedTopic != "np-topic" {
		t.Errorf("topic = %q, want np-topic", fake.publishedTopic)
	}

	var event map[string]any
	if err := json.Unmarshal(fake.publishedData, &event); err != nil {
		t.Fatalf("published data is not valid JSON: %v", err)
	}
	if event["type"] != "com.hyperfleet.nodepool.changed" {
		t.Errorf("event.type = %v, want com.hyperfleet.nodepool.changed", event["type"])
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got: %s", stderr)
	}
}
