#!/bin/bash
# Increment the counter field in cluster spec or labels
# Usage: hf.cluster.patch.sh spec|labels [cluster_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id

hf_require_jq

TARGET="${1:-}"
if [[ "$TARGET" != "spec" && "$TARGET" != "labels" ]]; then
  hf_usage "spec|labels [cluster_id]"
  echo "Arguments:"
  echo "  spec|labels   Which section to increment the counter field in (required)"
  echo "  cluster_id    Cluster ID (default: current cluster)"
  exit 1
fi

CLUSTER_ID=$(hf_cluster_id "${2:-}")
hf_info "Fetching cluster: $CLUSTER_ID"

CURRENT=$(hf_get "/clusters/${CLUSTER_ID}")
COUNTER=$(echo "$CURRENT" | jq -r ".${TARGET}.counter // \"0\"")
NEW_COUNTER=$((COUNTER + 1))

hf_info "Incrementing ${TARGET}.counter: $COUNTER → $NEW_COUNTER"

PAYLOAD=$(echo "$CURRENT" | jq --arg t "$TARGET" --arg v "$NEW_COUNTER" '{($t): (.[$t] + {counter: $v})}')
hf_patch "/clusters/${CLUSTER_ID}" "$PAYLOAD" | jq
