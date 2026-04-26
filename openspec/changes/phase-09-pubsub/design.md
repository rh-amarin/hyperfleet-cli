# Phase 9 Design — Pub/Sub & RabbitMQ

## Package layout

```
internal/pubsub/
  interfaces.go   — GCPPublisher + RabbitPublisher interfaces
  gcp.go          — GCPClient (wraps cloud.google.com/go/pubsub)
  rabbitmq.go     — RabbitClient (wraps github.com/rabbitmq/amqp091-go)
  gcp_test.go     — compile-time interface compliance check
  rabbitmq_test.go
cmd/
  pubsub.go       — hf pubsub {list, publish cluster, publish nodepool}
  pubsub_test.go
  rabbitmq.go     — hf rabbitmq publish cluster
  rabbitmq_test.go
```

## Interfaces

```go
// GCPPublisher can list subscriptions and publish to GCP Pub/Sub.
type GCPPublisher interface {
    ListSubscriptions(ctx context.Context, filter string) ([]string, error)
    Publish(ctx context.Context, topicID string, data []byte) (string, error)
    Close() error
}

// RabbitPublisher publishes messages to a RabbitMQ exchange.
type RabbitPublisher interface {
    Publish(ctx context.Context, exchange, routingKey string, body []byte) error
    Close() error
}
```

Both concrete types satisfy these interfaces at compile time (checked by `var _ = ...`
lines in test files).

## GCPClient

- Wraps `*pubsub.Client` from `cloud.google.com/go/pubsub`.
- Auth: `oauth2.StaticTokenSource` from `hyperfleet.token` config value.
- `ListSubscriptions` iterates `client.Subscriptions(ctx)`, returns full resource names
  (`projects/<project>/subscriptions/<id>`), optionally filtered by substring.
- `Publish` calls `topic.Publish(ctx, msg).Get(ctx)` and returns the server message ID.

## RabbitClient

- Wraps `*amqp.Connection` from `github.com/rabbitmq/amqp091-go`.
- `NewRabbit(url)` dials `amqp://<user>:<pass>@<host>/<vhost>`.
- `Publish` opens a channel per call (channel is cheap; avoids channel reuse race).

## CloudEvent envelope

```json
{
  "specversion": "1.0",
  "type": "com.hyperfleet.cluster.changed",
  "source": "/hyperfleet/cli",
  "id": "<random-uuid>",
  "data": <cluster|nodepool object>
}
```

Type strings:
- cluster publish → `com.hyperfleet.cluster.changed`
- nodepool publish → `com.hyperfleet.nodepool.changed`

## Command wiring

### `hf pubsub list [filter]`
1. Read `hyperfleet.gcp-project` and `hyperfleet.token` from cfgStore.
2. Call `gcpFactory(ctx, project, token)` → GCPPublisher.
3. Call `client.ListSubscriptions(ctx, filter)`.
4. Render as table with header `SUBSCRIPTION`.

### `hf pubsub publish cluster <topic>`
1. Resolve cluster-id from config (same as all cluster commands).
2. Fetch cluster from HyperFleet API.
3. Build CloudEvent JSON.
4. Publish via `gcpFactory`.
5. Print `[INFO]` line with cluster-id, topic, server msg-id.

### `hf pubsub publish nodepool <topic>`
Same as above but fetches nodepool; event type `com.hyperfleet.nodepool.changed`.

### `hf rabbitmq publish cluster <exchange> [routing-key]`
1. Resolve cluster-id.
2. Fetch cluster.
3. Build CloudEvent JSON.
4. Build AMQP URL from `rabbitmq.{host,user,password,vhost}` config.
5. Publish via `rabbitFactory`.
6. Print `[INFO]` line.

## Testability

Package-level factory vars in cmd:
```go
var gcpFactory    = func(ctx, project, token) (ps.GCPPublisher, error)  { ... }
var rabbitFactory = func(amqpURL string) (ps.RabbitPublisher, error)     { ... }
```

Tests override these vars with fakes, mock HyperFleet API with `httptest.Server`.
No live GCP or RabbitMQ required for unit tests.

## Config keys used

| Key | Source |
|---|---|
| `hyperfleet.gcp-project` | GCP project for Pub/Sub |
| `hyperfleet.token` | OAuth2 access token |
| `rabbitmq.host` | AMQP server hostname |
| `rabbitmq.user` | AMQP username |
| `rabbitmq.password` | AMQP password |
| `rabbitmq.vhost` | AMQP virtual host |
