#!/bin/bash
# List all NodePools for the current cluster
# Usage: hf.nodepool.list.sh [cluster_id]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id

hf_require_jq
CLUSTER_ID=$(hf_cluster_id "${1:-}")
hf_get "/clusters/${CLUSTER_ID}/nodepools" | jq
