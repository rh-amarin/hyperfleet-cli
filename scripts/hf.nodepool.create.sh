#!/bin/bash
# Create one or more NodePools under the current cluster
# Usage: hf.nodepool.create.sh <name> [count] [instance-type]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id

NAME="${1:-my-nodepool}"
COUNT="${2:-1}"
INSTANCE_TYPE="${3:-m4}"

[[ -z "$NAME" ]] && {
  hf_usage "<name> [count] [instance-type]"
  echo "Arguments:"
  echo "  name            NodePool name prefix (required)"
  echo "  count           Number of nodepools to create (default: 1)"
  echo "  instance-type   Instance type (default: m4)"
  exit 1
}

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)

hf_info "Creating $COUNT nodepool(s) with prefix '$NAME' in cluster: $CLUSTER_ID"
hf_info "  Instance type: $INSTANCE_TYPE"

for ((i = 1; i <= COUNT; i++)); do
  POOL_NAME="${NAME}-${i}"
  hf_info "Creating nodepool '$POOL_NAME' ($i/$COUNT)..."

  PAYLOAD=$(
    cat <<EOF
{
  "kind": "NodePool",
  "name": "$POOL_NAME",
  "labels": {
    "counter":"$i"
  },
  "spec": {
    "replicas": 1,
    "counter":"$i",
    "platform": {
      "type": "$INSTANCE_TYPE"
    }
  }
}
EOF
  )

  RESPONSE=$(hf_post "/clusters/${CLUSTER_ID}/nodepools" "$PAYLOAD")
  echo "$RESPONSE" | jq

  NODEPOOL_ID=$(echo "$RESPONSE" | jq -r '.id // empty')
  if [[ -n "$NODEPOOL_ID" ]]; then
    hf_set_nodepool_id "$NODEPOOL_ID"
  fi
done
