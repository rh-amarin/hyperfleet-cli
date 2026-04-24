#!/bin/bash
# Delete a NodePool (uses current nodepool if no ID provided)
# Usage: hf.nodepool.delete.sh [nodepool_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id nodepool-id

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)
NODEPOOL_ID=$(hf_nodepool_id "${1:-}")
hf_info "Deleting nodepool: $NODEPOOL_ID (cluster: $CLUSTER_ID)"
hf_delete "/clusters/${CLUSTER_ID}/nodepools/${NODEPOOL_ID}" | jq
