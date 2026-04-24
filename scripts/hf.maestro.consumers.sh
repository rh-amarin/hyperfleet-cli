#!/bin/bash
# List maestro consumers
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config maestro-http-endpoint

hf_require_jq
curl -s "${HF_MAESTRO_HTTP_ENDPOINT}/api/maestro/v1/consumers" | jq
