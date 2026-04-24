#!/bin/bash
# Launch maestro-cli TUI
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config maestro-http-endpoint

maestro-cli tui --http-endpoint "$HF_MAESTRO_HTTP_ENDPOINT"
