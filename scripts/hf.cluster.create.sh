#!/bin/bash
# Create a new cluster
# Usage: hf.cluster.create.sh <name> [region] [version]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

NAME="${1:-my-cluster}"
REGION="${2:-us-east-1}"
VERSION="${3:-4.15.0}"

hf_require_jq

# Check if cluster exists
hf_info "Checking if cluster '$NAME' exists..."
EXISTING=$(hf_get "/clusters?search=name='$NAME'" | jq -r ".items[]? | select(.name == \"$NAME\") | .name")

if [[ -n "$EXISTING" ]]; then
  hf_warn "Cluster '$NAME' already exists, skipping creation"
  exit 0
fi

hf_info "Creating cluster: $NAME (region: $REGION, version: $VERSION)"

PAYLOAD=$(
  cat <<EOF
{
  "kind": "Cluster",
  "name": "$NAME",
  "labels": {
    "environment": "development",
    "shard": "1",
    "team": "core",
    "counter": "1"
  },
  "spec": {
    "region": "$REGION",
    "version": "$VERSION",
    "counter": "1"
  }
}
EOF
)

hf_post "/clusters" "$PAYLOAD" | jq

# Set as current cluster
"$(dirname "$(realpath "$0")")/hf.cluster.search.sh" "$NAME"
