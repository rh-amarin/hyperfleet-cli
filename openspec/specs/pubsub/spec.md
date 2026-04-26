# Pub/Sub and Messaging Specification

## Purpose

Provide CLI commands for publishing CloudEvent messages to GCP Pub/Sub topics and RabbitMQ exchanges, and for listing Pub/Sub topics and subscriptions. These events trigger adapter reconciliation in the HyperFleet system.

## Requirements

### Requirement: List Pub/Sub Topics and Subscriptions

The CLI SHALL list GCP Pub/Sub topics and their subscriptions.

#### Scenario: List all topics

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list`
- THEN the CLI MUST list all Pub/Sub topics in the configured GCP project
- AND for each topic, list its subscriptions indented beneath it
- AND topics MUST appear at the left margin
- AND subscriptions MUST be indented with 4 spaces

#### Scenario: List with filter

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list <filter_term>`
- THEN the CLI MUST filter both topics AND subscriptions by the provided term
- AND display an `[INFO]` message about the active filter
- AND only show topics and subscriptions matching the filter

### Requirement: Publish Cluster Change Event to Pub/Sub

The CLI SHALL publish a cluster reconcile event to a GCP Pub/Sub topic.

#### Scenario: Publish cluster event

- GIVEN gcp-project and cluster-id are configured
- WHEN the user runs `hf pubsub publish cluster <topic>`
- THEN the CLI MUST publish a CloudEvent 1.0 message to the specified topic via `gcloud pubsub topics publish`
- AND the event payload MUST be:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.cluster.reconcile.v1",
    "source": "/hyperfleet/api",
    "id": "<cluster_id>",
    "data": {"cluster_id": "<cluster_id>"}
  }
  ```

### Requirement: Publish NodePool Change Event to Pub/Sub

The CLI SHALL publish a nodepool reconcile event to a GCP Pub/Sub topic.

#### Scenario: Publish nodepool event

- GIVEN gcp-project, cluster-id, and nodepool-id are configured
- WHEN the user runs `hf pubsub publish nodepool <topic>`
- THEN the CLI MUST publish a CloudEvent 1.0 message to the specified topic
- AND the event payload MUST be:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.nodepool.reconcile.v1",
    "source": "/hyperfleet/api",
    "id": "<nodepool_id>",
    "data": {"cluster_id": "<cluster_id>", "nodepool_id": "<nodepool_id>"}
  }
  ```

### Requirement: Publish Cluster Change Event to RabbitMQ

The CLI SHALL publish a cluster reconcile event to a RabbitMQ exchange.

#### Scenario: Publish cluster event to RabbitMQ

- GIVEN rabbitmq-host, rabbitmq-mgmt-port, rabbitmq-user, rabbitmq-password, rabbitmq-vhost, and cluster-id are configured
- WHEN the user runs `hf rabbitmq publish cluster <exchange> [routing-key]`
- THEN the CLI MUST send a POST request to the RabbitMQ Management API at `http://{host}:{mgmt-port}/api/exchanges/{vhost}/{exchange}/publish`
- AND the vhost MUST be URL-encoded (`/` becomes `%2F`)
- AND the message MUST use CloudEvent 1.0 format with:
  - `content_type`: `application/json`
  - CloudEvent headers in `headers` field: `ce_specversion`, `ce_type`, `ce_source`, `ce_id`
  - `ce_type`: `com.redhat.hyperfleet.cluster.reconcile.v1`
  - `payload`: `{"cluster_id": "<cluster_id>"}`
