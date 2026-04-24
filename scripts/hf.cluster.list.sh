#!/bin/bash
# List all clusters
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

hf_require_jq
hf_get "/clusters" | jq
