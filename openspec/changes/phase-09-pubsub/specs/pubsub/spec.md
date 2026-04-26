# Pub/Sub Spec — Delta

## ADDED: GCP Pub/Sub integration

`hf pubsub` commands publish CloudEvents 1.0 messages to GCP Pub/Sub topics.
Authentication uses the `hyperfleet.token` OAuth2 access token via
`oauth2.StaticTokenSource`.

### Commands

| Command | Description |
|---|---|
| `hf pubsub list [filter]` | List subscriptions for `hyperfleet.gcp-project`, optional substring filter |
| `hf pubsub publish cluster <topic>` | Fetch current cluster, publish CloudEvent |
| `hf pubsub publish nodepool <topic>` | Fetch current nodepool, publish CloudEvent |

### CloudEvent envelope

```json
{
  "specversion": "1.0",
  "type": "com.hyperfleet.{cluster|nodepool}.changed",
  "source": "/hyperfleet/cli",
  "id": "<random-uuid>",
  "data": <cluster|nodepool JSON object>
}
```

## ADDED: RabbitMQ integration

`hf rabbitmq` commands publish CloudEvents to RabbitMQ exchanges via AMQP 0-9-1.

### Commands

| Command | Description |
|---|---|
| `hf rabbitmq publish cluster <exchange> [routing-key]` | Fetch cluster, publish CloudEvent to exchange |

### Config keys

| Key | Purpose |
|---|---|
| `rabbitmq.host` | AMQP server hostname |
| `rabbitmq.user` | AMQP username |
| `rabbitmq.password` | AMQP password |
| `rabbitmq.vhost` | AMQP virtual host (default `/`) |

### AMQP URL construction

`amqp://<user>:<pass>@<host>/<vhost>` — vhost `/` maps to path component empty string
so the URL is `amqp://user:pass@host/`.
