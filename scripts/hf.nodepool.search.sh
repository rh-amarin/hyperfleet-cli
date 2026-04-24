#!/bin/bash
# Search for nodepool by name and set as current
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version cluster-id

# Get name from argument or saved file
if [[ -n "${1:-}" ]]; then
    SEARCH_NAME="$1"
    hf_set_nodepool_name "$SEARCH_NAME"
else
    SEARCH_NAME=$(hf_nodepool_name) || exit 1
fi

hf_require_jq
CLUSTER_ID=$(hf_cluster_id)
hf_info "Searching for nodepool: $SEARCH_NAME (cluster: $CLUSTER_ID)"

RESULT=$(hf_get "/clusters/${CLUSTER_ID}/nodepools?search=name='$SEARCH_NAME'")
EXACT=$(echo "$RESULT" | jq --arg name "$SEARCH_NAME" '[.items[] | select(.name == $name and .deleted_at == null)]')
echo "$EXACT" | jq
COUNT=$(echo "$EXACT" | jq 'length')
if [[ "$COUNT" -eq 1 ]]; then
    NODEPOOL_ID=$(echo "$EXACT" | jq -r '.[0].id')
    hf_set_nodepool_id "$NODEPOOL_ID"
elif [[ "$COUNT" -gt 1 ]]; then
    hf_warn "Multiple nodepools found with name '$SEARCH_NAME'. Use a more specific search."
else
    hf_warn "No nodepools found matching '$SEARCH_NAME'"
fi
