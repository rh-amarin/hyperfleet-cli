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

// fakeRabbitClient implements ps.RabbitPublisher for testing without a live RabbitMQ.
type fakeRabbitClient struct {
	publishedExchange   string
	publishedRoutingKey string
	publishedData       []byte
	publishErr          error
}

func (f *fakeRabbitClient) Publish(_ context.Context, exchange, routingKey string, body []byte) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	f.publishedExchange = exchange
	f.publishedRoutingKey = routingKey
	f.publishedData = body
	return nil
}

func (f *fakeRabbitClient) Close() error { return nil }

var _ ps.RabbitPublisher = (*fakeRabbitClient)(nil)

// ── publish cluster ───────────────────────────────────────────────────────────

func TestRabbitMQPublishCluster_PublishesCloudEvent(t *testing.T) {
	fake := &fakeRabbitClient{}
	orig := rabbitFactory
	rabbitFactory = func(_ string) (ps.RabbitPublisher, error) { return fake, nil }
	defer func() { rabbitFactory = orig }()

	cluster := resource.Cluster{ID: "c-001", Name: "prod", Generation: 3}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cluster)
	}))
	defer srv.Close()

	_, stderr, err := runCmdWithCluster(t, srv, "c-001", "rabbitmq", "publish", "cluster", "my-exchange")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.publishedExchange != "my-exchange" {
		t.Errorf("exchange = %q, want my-exchange", fake.publishedExchange)
	}
	if fake.publishedRoutingKey != "" {
		t.Errorf("routing-key = %q, want empty (not provided)", fake.publishedRoutingKey)
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
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got: %s", stderr)
	}
}

func TestRabbitMQPublishCluster_WithRoutingKey(t *testing.T) {
	fake := &fakeRabbitClient{}
	orig := rabbitFactory
	rabbitFactory = func(_ string) (ps.RabbitPublisher, error) { return fake, nil }
	defer func() { rabbitFactory = orig }()

	cluster := resource.Cluster{ID: "c-001", Name: "prod"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cluster)
	}))
	defer srv.Close()

	_, _, err := runCmdWithCluster(t, srv, "c-001", "rabbitmq", "publish", "cluster", "my-exchange", "cluster.changed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.publishedRoutingKey != "cluster.changed" {
		t.Errorf("routing-key = %q, want cluster.changed", fake.publishedRoutingKey)
	}
}
