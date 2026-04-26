# Phase 9 — Pub/Sub & RabbitMQ

## Intent

Add `hf pubsub` and `hf rabbitmq` command groups that publish CloudEvents to GCP
Pub/Sub and RabbitMQ, replacing ad-hoc `gcloud` and `curl` scripts used by HyperFleet
operators. The commands fetch live cluster/nodepool state from the HyperFleet API and
wrap it in a CloudEvents 1.0 envelope before publishing.

## Why

- Operators currently rely on shell scripts that call `gcloud pubsub topics publish` or
  custom `curl` commands to push cluster-state events. These scripts are fragile,
  untyped, and not reusable.
- The `hf` binary already holds the HyperFleet API credentials and config; it is the
  right place to encode this publishing logic.

## Scope in

- `internal/pubsub/` — GCPClient and RabbitClient with interface definitions
- `cmd/pubsub.go` — `hf pubsub list` and `hf pubsub publish {cluster,nodepool}`
- `cmd/rabbitmq.go` — `hf rabbitmq publish cluster`
- Dependencies: `cloud.google.com/go/pubsub`, `github.com/rabbitmq/amqp091-go`

## Scope out

- RabbitMQ management API (queues, bindings)
- GCP topic/subscription creation or deletion
- Any non-publishing Pub/Sub operations (pull, seek, ack)
