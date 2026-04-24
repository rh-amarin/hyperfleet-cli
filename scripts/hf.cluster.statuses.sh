#!/bin/bash
# Show cluster statuses (use -w for watch mode)
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)
URL="${HF_API_URL}/api/hyperfleet/${HF_API_VERSION}/clusters/${CLUSTER_ID}/statuses"

if [[ "${1:-}" == "-w" ]]; then
    hf_require_viddy
    hf_info "Watching statuses for cluster: $CLUSTER_ID"
    viddy -d "curl -s '$URL' | jq"
else
    hf_info "Getting statuses for cluster: $CLUSTER_ID"
    curl -s "$URL" | jq
fi
