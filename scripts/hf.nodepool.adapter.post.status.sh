#!/bin/bash
# Post adapter status for current nodepool
# Usage: hf.nodepool.adapter.post.status.sh <adapter_name> <available> [generation] [nodepool_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id nodepool-id

ADAPTER_NAME="${1:-}"
AVAILABLE="${2:-}"
GENERATION="${3:-1}"
NODEPOOL_ARG="${4:-}"

if [[ -z "$ADAPTER_NAME" ]] || [[ -z "$AVAILABLE" ]]; then
  hf_usage "<adapter_name> <available> [generation] [nodepool_id]"
  echo "Arguments:"
  echo "  adapter_name  Name of the adapter (e.g., validator, dns, provisioner)"
  echo "  available     Status: True, False, or Unknown"
  echo "  generation    Observed generation (default: 1)"
  echo "  nodepool_id   NodePool ID (default: configured nodepool-id)"
  echo ""
  echo "Example: hf.nodepool.adapter.post.status.sh validator True 1"
  exit 1
fi

# Validate available status
if [[ ! "$AVAILABLE" =~ ^(True|False|Unknown)$ ]]; then
  hf_die "available must be 'True', 'False', or 'Unknown'"
fi

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)
NODEPOOL_ID=$(hf_nodepool_id "$NODEPOOL_ARG")
OBSERVED_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

hf_info "Posting adapter status: $ADAPTER_NAME = $AVAILABLE (gen: $GENERATION) for nodepool: $NODEPOOL_ID"

PAYLOAD=$(
  cat <<EOF
{
    "adapter": "${ADAPTER_NAME}",
    "observed_generation": ${GENERATION},
    "observed_time": "${OBSERVED_TIME}",
    "conditions": [
        {
            "type": "Available",
            "status": "${AVAILABLE}",
            "reason": "ManualStatusPost",
            "message": "Status posted via hf.nodepool.adapter.post.status.sh"
        },
        {
            "type": "Applied",
            "status": "${AVAILABLE}",
            "reason": "ManualStatusPost",
            "message": "Status posted via hf.nodepool.adapter.post.status.sh"
        },
        {
            "type": "Health",
            "status": "${AVAILABLE}",
            "reason": "ManualStatusPost",
            "message": "Status posted via hf.nodepool.adapter.post.status.sh"
        }
    ]
}
EOF
)

hf_post "/clusters/${CLUSTER_ID}/nodepools/${NODEPOOL_ID}/statuses" "$PAYLOAD" | jq
