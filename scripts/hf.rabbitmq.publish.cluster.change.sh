#!/bin/bash
# Publish cluster change message to a RabbitMQ exchange via the HTTP management API
# Usage: hf.rabbitmq.publish.cluster.change.sh <exchange> [routing-key]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config cluster-id rabbitmq-host rabbitmq-mgmt-port rabbitmq-user rabbitmq-password rabbitmq-vhost
hf_require_jq

EXCHANGE="${1:-}"
ROUTING_KEY="${2:-}"
[[ -z "$EXCHANGE" ]] && hf_die "Usage: hf.rabbitmq.publish.cluster.change.sh <exchange> [routing-key]"

CLUSTER_ID=$(hf_cluster_id)

MESSAGE=$(cat <<EOF
{
  "specversion": "1.0",
  "type": "com.redhat.hyperfleet.cluster.reconcile.v1",
  "source": "/hyperfleet/service/sentinel",
  "id": "${CLUSTER_ID}",
  "time": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "datacontenttype": "application/json",
  "data": {
    "id": "${CLUSTER_ID}",
    "kind": "Cluster",
    "href": "https://api.hyperfleet.com/v1/clusters/${CLUSTER_ID}",
    "generation": 1
  }
}
EOF
)

# URL-encode the vhost (e.g. "/" -> "%2F")
VHOST_ENC="${HF_RABBITMQ_VHOST//\//%2F}"

BODY=$(jq -n \
  --arg routing_key "$ROUTING_KEY" \
  --arg payload "$MESSAGE" \
  '{properties: {}, routing_key: $routing_key, payload: $payload, payload_encoding: "string"}')

hf_info "Publishing change message to exchange: $EXCHANGE (routing-key: ${ROUTING_KEY:-<empty>})"
echo "$MESSAGE" | jq .

TMPFILE=$(mktemp)
HTTP_CODE=$(curl -s -o "$TMPFILE" -w "%{http_code}" \
  -u "${HF_RABBITMQ_USER}:${HF_RABBITMQ_PASSWORD}" \
  -H "Content-Type: application/json" \
  -X POST \
  "http://${HF_RABBITMQ_HOST}:${HF_RABBITMQ_MGMT_PORT}/api/exchanges/${VHOST_ENC}/${EXCHANGE}/publish" \
  -d "$BODY")
BODY_RESP=$(cat "$TMPFILE")
rm -f "$TMPFILE"

if [[ "$HTTP_CODE" != "200" ]]; then
  hf_die "Failed to publish (HTTP $HTTP_CODE): $BODY_RESP"
fi

hf_info "Published successfully: $BODY_RESP"
