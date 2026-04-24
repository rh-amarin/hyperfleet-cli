#!/bin/bash
# Search for cluster by name and set as current
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

# Get name from argument or saved file
if [[ -n "${1:-}" ]]; then
    SEARCH_NAME="$1"
    hf_set_cluster_name "$SEARCH_NAME"
else
    SEARCH_NAME=$(hf_cluster_name) || exit 1
fi

hf_require_jq
hf_info "Searching for cluster: $SEARCH_NAME"

RESULT=$(hf_get "/clusters?search=name='$SEARCH_NAME'")
EXACT=$(echo "$RESULT" | jq --arg name "$SEARCH_NAME" '[.items[] | select(.name == $name and .deleted_at == null)]')
echo "$EXACT" | jq
COUNT=$(echo "$EXACT" | jq 'length')
if [[ "$COUNT" -eq 1 ]]; then
    CLUSTER_ID=$(echo "$EXACT" | jq -r '.[0].id')
    hf_set_cluster_id "$CLUSTER_ID"
elif [[ "$COUNT" -gt 1 ]]; then
    hf_warn "Multiple clusters found with name '$SEARCH_NAME'. Use a more specific search."
else
    hf_warn "No clusters found matching '$SEARCH_NAME'"
fi
