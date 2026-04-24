#!/bin/bash
# Show current cluster ID
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config cluster-id

if [[ -f "$HF_CLUSTER_ID_FILE" ]]; then
    cat "$HF_CLUSTER_ID_FILE"
else
    hf_die "No cluster ID set. Use hf.cluster.search.sh to set one."
fi
