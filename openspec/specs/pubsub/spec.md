# Pub/Sub and Messaging Specification

## Purpose

Provide CLI commands for publishing CloudEvent messages to GCP Pub/Sub topics and RabbitMQ exchanges, and for listing Pub/Sub topics and subscriptions. These events trigger adapter reconciliation in the HyperFleet system.

## Requirements

### Requirement: List Pub/Sub Topics and Subscriptions

The CLI SHALL list GCP Pub/Sub topics and their subscriptions.

#### Scenario: List all topics

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list`
- THEN the CLI MUST print an `[INFO]` line identifying the project
- AND list all Pub/Sub topics in the configured GCP project
- AND for each topic, list its subscriptions indented beneath it
- AND topics MUST appear at the left margin
- AND subscriptions MUST be indented with 4 spaces

#### Scenario: List with filter

- GIVEN gcp-project is configured
- WHEN the user runs `hf pubsub list <filter_term>`
- THEN the CLI MUST print `[INFO] Filtering by: <filter_term>`
- AND filter both topics AND subscriptions by the provided substring
- AND only show topics/subscriptions whose name contains the term

### Requirement: Publish Cluster Change Event to Pub/Sub

The CLI SHALL publish a cluster reconcile event to a GCP Pub/Sub topic.

#### Scenario: Publish cluster event

- GIVEN gcp-project and cluster-id are configured
- WHEN the user runs `hf pubsub publish cluster <topic>`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND publish the following CloudEvent 1.0 message to the specified topic via the GCP Pub/Sub SDK:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.cluster.reconcile.v1",
    "source": "/hyperfleet/service/sentinel",
    "id": "<cluster_id>",
    "time": "<UTC ISO8601>",
    "datacontenttype": "application/json",
    "data": {
      "id": "<cluster_id>",
      "kind": "Cluster",
      "href": "https://api.hyperfleet.com/v1/clusters/<cluster_id>",
      "generation": 1
    }
  }
  ```
- AND the cluster-id MUST be read from state (no HyperFleet API fetch)
- AND print `[INFO] Published cluster <id> to topic <topic> (msg-id: <id>)` on success

### Requirement: Publish NodePool Change Event to Pub/Sub

The CLI SHALL publish a nodepool reconcile event to a GCP Pub/Sub topic.

#### Scenario: Publish nodepool event

- GIVEN gcp-project, cluster-id, and nodepool-id are configured
- WHEN the user runs `hf pubsub publish nodepool <topic>`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND publish the following CloudEvent 1.0 message to the specified topic:
  ```json
  {
    "specversion": "1.0",
    "type": "com.redhat.hyperfleet.nodepool.reconcile.v1",
    "source": "/hyperfleet/service/sentinel",
    "id": "<nodepool_id>",
    "time": "<UTC ISO8601>",
    "datacontenttype": "application/json",
    "data": {
      "id": "<nodepool_id>",
      "kind": "NodePool",
      "href": "http://localhost:8000/api/hyperfleet/v1/clusters/<cluster_id>/node_pools/<nodepool_id>",
      "generation": 1,
      "owner_references": {
        "id": "<cluster_id>",
        "kind": "NodePool",
        "href": "http://localhost:8000/api/hyperfleet/v1/clusters/<cluster_id>",
        "generation": 1
      }
    }
  }
  ```
- AND both cluster-id and nodepool-id MUST be read from state (no HyperFleet API fetch)

### Requirement: Publish Cluster Change Event to RabbitMQ

The CLI SHALL publish a cluster reconcile event to a RabbitMQ exchange via the HTTP Management API.

#### Scenario: Publish cluster event to RabbitMQ

- GIVEN rabbitmq-host, rabbitmq-mgmt-port, rabbitmq-user, rabbitmq-password, rabbitmq-vhost, and cluster-id are configured
- WHEN the user runs `hf rabbitmq publish cluster <exchange> [routing-key]`
- THEN the CLI MUST print the CloudEvent JSON to stdout
- AND send a POST to `http://{host}:{mgmt-port}/api/exchanges/{vhost-encoded}/{exchange}/publish`
- AND the vhost MUST be URL-encoded (`/` becomes `%2F`)
- AND the request body MUST be:
  ```json
  {
    "properties": {},
    "routing_key": "<routing-key>",
    "payload": "<CloudEvent JSON string>",
    "payload_encoding": "string"
  }
  ```
- AND the CloudEvent payload MUST use the same cluster envelope as `hf pubsub publish cluster`
- AND cluster-id MUST be read from state (no HyperFleet API fetch)
- AND routing-key defaults to empty string when not provided

---

## Go Implementation (added in phase-09-pubsub)

### Package: `internal/pubsub`

| File | Contents |
|---|---|
| `interfaces.go` | `GCPPublisher`, `RabbitPublisher` interfaces; `TopicGroup` struct |
| `gcp.go` | `GCPClient` — wraps `cloud.google.com/go/pubsub`, auth via `oauth2.StaticTokenSource` using the `hyperfleet.token` config value |
| `rabbitmq.go` | `RabbitClient` — HTTP Management API via `net/http` + BasicAuth; vhost `/` is URL-encoded to `%2F` |

### Commands

| Command | File | Factory var |
|---|---|---|
| `hf pubsub list/publish` | `cmd/pubsub.go` | `gcpFactory` |
| `hf rabbitmq publish` | `cmd/rabbitmq.go` | `rabbitFactory` |

### Config keys

| Key | Default | Purpose |
|---|---|---|
| `hyperfleet.token` | — | OAuth2 access token passed to `oauth2.StaticTokenSource` for GCP Pub/Sub |
| `hyperfleet.gcp-project` | `hcm-hyperfleet` | GCP project ID |
| `rabbitmq.host` | `rabbitmq` | RabbitMQ management hostname |
| `rabbitmq.mgmt-port` | `15672` | RabbitMQ HTTP management port |
| `rabbitmq.user` | `guest` | HTTP BasicAuth username |
| `rabbitmq.password` | `guest` | HTTP BasicAuth password |
| `rabbitmq.vhost` | `/` | Virtual host (URL-encoded as `%2F` in requests when set to `/`) |

### CloudEvent builders

`buildClusterEvent(clusterID string) ([]byte, error)` and
`buildNodePoolEvent(clusterID, nodepoolID string) ([]byte, error)` in `cmd/pubsub.go`
construct the canonical CloudEvent envelopes. Both are used by both `hf pubsub` and
`hf rabbitmq` commands.
