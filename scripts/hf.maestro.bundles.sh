#!/bin/bash
# List maestro resource bundles
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config maestro-http-endpoint

curl -s "${HF_MAESTRO_HTTP_ENDPOINT}/api/maestro/v1/resource-bundles" | jq
