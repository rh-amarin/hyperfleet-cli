#!/bin/bash
# Delete a cluster (uses current cluster if no ID provided)
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id

hf_require_jq
CLUSTER_ID=$(hf_cluster_id "${1:-}")
hf_info "Deleting cluster: $CLUSTER_ID"
hf_delete "/clusters/${CLUSTER_ID}" | jq
