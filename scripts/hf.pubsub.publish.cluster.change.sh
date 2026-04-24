#!/bin/bash
# Publish cluster change message to Pub/Sub topic
# Usage: hf.pubsub.publish.sh <topic>
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config gcp-project cluster-id

TOPIC="${1:-}"
[[ -z "$TOPIC" ]] && hf_die "Usage: hf.pubsub.publish.sh <topic>"

hf_require_gcloud
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

hf_info "Publishing change message to topic: $TOPIC"
echo "$MESSAGE" | jq .

gcloud pubsub topics publish "$TOPIC" --project "$HF_GCP_PROJECT" --message "$MESSAGE"
