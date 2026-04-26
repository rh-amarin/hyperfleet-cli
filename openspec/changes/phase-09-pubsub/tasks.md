# Phase 9 Tasks — Pub/Sub & RabbitMQ

## 1. internal/pubsub package

- [x] 1.1 Define `GCPPublisher` and `RabbitPublisher` interfaces in `interfaces.go`
- [x] 1.2 Implement `GCPClient` with `NewGCP`, `ListSubscriptions`, `Publish`, `Close`
- [x] 1.3 Implement `RabbitClient` with `NewRabbit`, `Publish`, `Close`
- [x] 1.4 Compile-time interface compliance checks in test files

## 2. cmd/pubsub.go

- [x] 2.1 `hf pubsub list [filter]` — list GCP subscriptions, print table
- [x] 2.2 `hf pubsub publish cluster <topic>` — fetch cluster state, publish CloudEvent
- [x] 2.3 `hf pubsub publish nodepool <topic>` — fetch nodepool state, publish CloudEvent
- [x] 2.4 `gcpFactory` var for testability

## 3. cmd/rabbitmq.go

- [x] 3.1 `hf rabbitmq publish cluster <exchange> [routing-key]` — fetch cluster, publish
- [x] 3.2 `rabbitFactory` var for testability
- [x] 3.3 `buildRabbitURL` helper from config fields

## 4. Tests

- [x] 4.1 `cmd/pubsub_test.go` — fakeGCPClient, list + publish cluster + publish nodepool
- [x] 4.2 `cmd/rabbitmq_test.go` — fakeRabbitClient, publish cluster + routing key

## 5. Dependencies

- [x] 5.1 `go get cloud.google.com/go/pubsub`
- [x] 5.2 `go get github.com/rabbitmq/amqp091-go`

## Verify

- [x] (a) `go build ./...` succeeds
- [x] (b) `go vet ./...` no issues
- [x] (c) `go test ./...` passes → `verification_proof/tests.txt`
- [x] (d) Live verification → `verification_proof/connectivity.txt`, `3.1-pubsub-list.txt`, `3.2-pubsub-publish-cluster.txt`, `3.3-rabbitmq-publish-cluster.txt`
  - GCP pubsub list: topics+subscriptions nested output verified against real `hcm-hyperfleet` project
  - GCP pubsub publish cluster: published CloudEvent to `amarin-ns1-clusters` (msg-id returned)
  - RabbitMQ: port-forwarded rabbitmq-5dcb9cdf95-5pmxj (v3.13.7), published to `amq.fanout` and `amq.topic` exchanges via HTTP Management API
