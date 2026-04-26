package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	ps "github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
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
	rabbitFactory = func(_ string, _ int, _, _, _ string) (ps.RabbitPublisher, error) { return fake, nil }
	defer func() { rabbitFactory = orig }()

	srv := minSrv()
	defer srv.Close()

	stdout, stderr, err := runCmdWithCluster(t, srv, "c-001", "rabbitmq", "publish", "cluster", "my-exchange")
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
	// stdout should contain the pretty-printed JSON
	if !strings.Contains(stdout, "specversion") {
		t.Errorf("expected JSON in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "[INFO]") {
		t.Errorf("expected [INFO] line in stderr, got: %s", stderr)
	}
}

func TestRabbitMQPublishCluster_WithRoutingKey(t *testing.T) {
	fake := &fakeRabbitClient{}
	orig := rabbitFactory
	rabbitFactory = func(_ string, _ int, _, _, _ string) (ps.RabbitPublisher, error) { return fake, nil }
	defer func() { rabbitFactory = orig }()

	srv := minSrv()
	defer srv.Close()

	_, _, err := runCmdWithCluster(t, srv, "c-001", "rabbitmq", "publish", "cluster", "my-exchange", "cluster.changed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fake.publishedRoutingKey != "cluster.changed" {
		t.Errorf("routing-key = %q, want cluster.changed", fake.publishedRoutingKey)
	}
}
