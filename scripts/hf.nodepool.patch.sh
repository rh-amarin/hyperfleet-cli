#!/bin/bash
# Increment the counter field in nodepool spec or labels
# Usage: hf.nodepool.patch.sh spec|labels [nodepool_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id nodepool-id

hf_require_jq

TARGET="${1:-}"
if [[ "$TARGET" != "spec" && "$TARGET" != "labels" ]]; then
  hf_usage "spec|labels [nodepool_id]"
  echo "Arguments:"
  echo "  spec|labels   Which section to increment the counter field in (required)"
  echo "  nodepool_id   NodePool ID (default: current nodepool)"
  exit 1
fi

CLUSTER_ID=$(hf_cluster_id)
NODEPOOL_ID=$(hf_nodepool_id "${2:-}")
hf_info "Fetching nodepool: $NODEPOOL_ID (cluster: $CLUSTER_ID)"

CURRENT=$(hf_get "/clusters/${CLUSTER_ID}/nodepools/${NODEPOOL_ID}")
COUNTER=$(echo "$CURRENT" | jq -r ".${TARGET}.counter // \"0\"")
NEW_COUNTER=$((COUNTER + 1))

hf_info "Incrementing ${TARGET}.counter: $COUNTER → $NEW_COUNTER"

PAYLOAD=$(echo "$CURRENT" | jq --arg t "$TARGET" --arg v "$NEW_COUNTER" '{($t): (.[$t] + {counter: $v})}')
hf_patch "/clusters/${CLUSTER_ID}/nodepools/${NODEPOOL_ID}" "$PAYLOAD" | jq
