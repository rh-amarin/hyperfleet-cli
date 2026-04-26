# Pub/Sub Spec — Delta (phase-09-pubsub)

## ADDED: GCP Pub/Sub integration (`hf pubsub`)

`hf pubsub` commands list topics/subscriptions and publish CloudEvents 1.0
messages to GCP Pub/Sub topics.
Authentication uses the `hyperfleet.token` OAuth2 access token via
`oauth2.StaticTokenSource` (Go SDK `cloud.google.com/go/pubsub`).

### Commands

| Command | Description |
|---|---|
| `hf pubsub list [filter]` | List topics+subscriptions for `hyperfleet.gcp-project`, optional substring filter |
| `hf pubsub publish cluster <topic>` | Build CloudEvent from stored cluster-id, publish to topic |
| `hf pubsub publish nodepool <topic>` | Build CloudEvent from stored nodepool-id, publish to topic |

### Output format for `hf pubsub list`

Topics are printed at left margin; subscriptions are indented 4 spaces:

```
topic-name
    subscription-name
    subscription-name
next-topic
```

An `[INFO]` line is printed listing the project on every invocation.
When a filter is provided, an additional `[INFO] Filtering by: <term>` line appears.
The filter applies to both topic names and subscription names.

### CloudEvent envelope (cluster)

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

### CloudEvent envelope (nodepool)

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

The JSON is printed to stdout before publishing (mirrors `echo "$MESSAGE" | jq .` in bash).
No HyperFleet API fetch is performed — only stored IDs from state are used.

## ADDED: RabbitMQ integration (`hf rabbitmq`)

`hf rabbitmq` commands publish CloudEvents to RabbitMQ exchanges via the
**HTTP Management API** (`net/http` + BasicAuth). Not AMQP.

### Commands

| Command | Description |
|---|---|
| `hf rabbitmq publish cluster <exchange> [routing-key]` | Build CloudEvent from stored cluster-id, publish to exchange |

### Config keys

| Key | Default | Purpose |
|---|---|---|
| `rabbitmq.host` | `rabbitmq` | Management API hostname |
| `rabbitmq.mgmt-port` | `15672` | Management API port |
| `rabbitmq.user` | `guest` | HTTP BasicAuth username |
| `rabbitmq.password` | `guest` | HTTP BasicAuth password |
| `rabbitmq.vhost` | `/` | Virtual host (URL-encoded; `/` → `%2F`) |

### HTTP Management API call

```
POST http://{host}:{mgmt-port}/api/exchanges/{vhost-encoded}/{exchange}/publish
Content-Type: application/json
Authorization: Basic <base64(user:password)>

{
  "properties": {},
  "routing_key": "<routing-key>",
  "payload": "<CloudEvent JSON string>",
  "payload_encoding": "string"
}
```

The same CloudEvent cluster envelope (see above) is used as the `payload` value.
The JSON is printed to stdout before the HTTP call.
