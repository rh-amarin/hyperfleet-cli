#!/bin/bash
# Publish nodepool change message to Pub/Sub topic
# Usage: hf.pubsub.publish.nodepool.change.sh <topic>
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config gcp-project cluster-id nodepool-id

TOPIC="${1:-}"
[[ -z "$TOPIC" ]] && hf_die "Usage: hf.pubsub.publish.nodepool.change.sh <topic>"

hf_require_gcloud
CLUSTER_ID=$(hf_cluster_id)
NODEPOOL_ID=$(hf_nodepool_id)

MESSAGE=$(
  cat <<EOF
{
  "specversion": "1.0",
  "type": "com.redhat.hyperfleet.nodepool.reconcile.v1",
  "source": "/hyperfleet/service/sentinel",
  "id": "${NODEPOOL_ID}",
  "time": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "datacontenttype": "application/json",
  "data": {
    "id": "${NODEPOOL_ID}",
    "kind": "NodePool",
    "href": "http://localhost:8000/api/hyperfleet/v1/clusters/${CLUSTER_ID}/node_pools/${NODEPOOL_ID}",
    "generation": 1,
    "owner_references": {
      "id": "${CLUSTER_ID}",
      "kind": "NodePool",
      "href": "http://localhost:8000/api/hyperfleet/v1/clusters/${CLUSTER_ID}",
      "generation": 1
    }

  }
}
EOF
)

hf_info "Publishing change message to topic: $TOPIC"
echo "$MESSAGE" | jq .

gcloud pubsub topics publish "$TOPIC" --project "$HF_GCP_PROJECT" --message "$MESSAGE"
