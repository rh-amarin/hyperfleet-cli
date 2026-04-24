#!/bin/bash
# Show NodePool adapter statuses (use -w for watch mode)
# Usage: hf.nodepool.statuses.sh [-w] [nodepool_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id nodepool-id

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)

WATCH=false
NODEPOOL_ARG=""

for arg in "$@"; do
  if [[ "$arg" == "-w" ]]; then
    WATCH=true
  else
    NODEPOOL_ARG="$arg"
  fi
done

NODEPOOL_ID=$(hf_nodepool_id "$NODEPOOL_ARG")
URL="${HF_API_URL}/api/hyperfleet/${HF_API_VERSION}/clusters/${CLUSTER_ID}/nodepools/${NODEPOOL_ID}/statuses"

if [[ "$WATCH" == true ]]; then
    hf_require_viddy
    hf_info "Watching statuses for nodepool: $NODEPOOL_ID"
    viddy -d "curl -s '$URL' | jq"
else
    hf_info "Getting statuses for nodepool: $NODEPOOL_ID"
    curl -s "$URL" | jq
fi
