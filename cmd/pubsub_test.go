package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ps "github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
)

// fakeGCPClient implements ps.GCPPublisher for testing without a live GCP connection.
type fakeGCPClient struct {
	groups         []ps.TopicGroup
	publishedTopic string
	publishedData  []byte
	publishErr     error
	listErr        error
}

func (f *fakeGCPClient) ListTopics(_ context.Context, _ string) ([]ps.TopicGroup, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.groups, nil
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

// minSrv returns a minimal httptest.Server (returns 200 for all requests).
func minSrv() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestPubSubList_PrintsTopicsAndSubscriptions(t *testing.T) {
	fake := &fakeGCPClient{
		groups: []ps.TopicGroup{
			{Name: "topic-alpha", Subscriptions: []string{"sub-a", "sub-b"}},
			{Name: "topic-beta", Subscriptions: []string{"sub-c"}},
		},
	}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	srv := minSrv()
	defer srv.Close()

	stdout, stderr, err := runCmd(t, srv, "pubsub", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "topic-alpha") {
		t.Errorf("expected topic-alpha in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "    sub-a") {
		t.Errorf("expected indented sub-a in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "    sub-b") {
		t.Errorf("expected indented sub-b in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "topic-beta") {
		t.Errorf("expected topic-beta in output, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got:\n%s", stderr)
	}
}

func TestPubSubList_WithFilter_ShowsInfoAndFiltered(t *testing.T) {
	fake := &fakeGCPClient{
		groups: []ps.TopicGroup{
			{Name: "amarin-ns1-clusters", Subscriptions: []string{"amarin-ns1-clusters-cl-job", "amarin-ns1-clusters-cl-precondition-error"}},
		},
	}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	srv := minSrv()
	defer srv.Close()

	stdout, stderr, err := runCmd(t, srv, "pubsub", "list", "amarin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "amarin-ns1-clusters") {
		t.Errorf("expected topic in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "    amarin-ns1-clusters-cl-job") {
		t.Errorf("expected indented subscription in output, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "Filtering by") {
		t.Errorf("expected Filtering by in stderr, got:\n%s", stderr)
	}
}

// ── publish cluster ───────────────────────────────────────────────────────────

func TestPubSubPublishCluster_PublishesCloudEvent(t *testing.T) {
	fake := &fakeGCPClient{}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	srv := minSrv()
	defer srv.Close()

	stdout, stderr, err := runCmdWithCluster(t, srv, "c-001", "pubsub", "publish", "cluster", "my-topic")
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
	if event["type"] != "com.redhat.hyperfleet.cluster.reconcile.v1" {
		t.Errorf("event.type = %v, want com.redhat.hyperfleet.cluster.reconcile.v1", event["type"])
	}
	if event["specversion"] != "1.0" {
		t.Errorf("event.specversion = %v, want 1.0", event["specversion"])
	}
	if event["source"] != "/hyperfleet/service/sentinel" {
		t.Errorf("event.source = %v, want /hyperfleet/service/sentinel", event["source"])
	}
	if event["id"] != "c-001" {
		t.Errorf("event.id = %v, want c-001", event["id"])
	}
	if event["datacontenttype"] != "application/json" {
		t.Errorf("event.datacontenttype = %v, want application/json", event["datacontenttype"])
	}
	// stdout should contain the pretty-printed JSON
	if !strings.Contains(stdout, "specversion") {
		t.Errorf("expected JSON in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got:\n%s", stderr)
	}
}

// ── publish nodepool ──────────────────────────────────────────────────────────

func TestPubSubPublishNodepool_PublishesCloudEvent(t *testing.T) {
	fake := &fakeGCPClient{}
	orig := gcpFactory
	gcpFactory = func(_ context.Context, _, _ string) (ps.GCPPublisher, error) { return fake, nil }
	defer func() { gcpFactory = orig }()

	srv := minSrv()
	defer srv.Close()

	stdout, stderr, err := runCmdWithClusterAndNodePool(t, srv, "c-001", "np-001", "pubsub", "publish", "nodepool", "np-topic")
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
	if event["type"] != "com.redhat.hyperfleet.nodepool.reconcile.v1" {
		t.Errorf("event.type = %v, want com.redhat.hyperfleet.nodepool.reconcile.v1", event["type"])
	}
	if event["id"] != "np-001" {
		t.Errorf("event.id = %v, want np-001", event["id"])
	}
	if event["source"] != "/hyperfleet/service/sentinel" {
		t.Errorf("event.source = %v, want /hyperfleet/service/sentinel", event["source"])
	}
	// stdout should contain the pretty-printed JSON
	if !strings.Contains(stdout, "owner_references") {
		t.Errorf("expected owner_references in stdout JSON, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got: %s", stderr)
	}
}
