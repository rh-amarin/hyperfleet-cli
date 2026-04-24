#!/bin/bash
# Get details of a specific NodePool
# Usage: hf.nodepool.get.sh [nodepool_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id nodepool-id

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)
NODEPOOL_ID=$(hf_nodepool_id "${1:-}")
[ -t 1 ] && hf_info "Getting nodepool: $NODEPOOL_ID (cluster: $CLUSTER_ID)"
hf_get "/clusters/${CLUSTER_ID}/nodepools/${NODEPOOL_ID}" | jq
