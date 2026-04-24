#!/bin/bash
# Get a maestro resource by name, or interactively select one from the list
# Usage: hf.maestro.get.sh [name]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config maestro-consumer maestro-http-endpoint maestro-grpc-endpoint

SCRIPT_DIR="$(dirname "$(realpath "$0")")"

get_resource() {
  local name="$1"
  hf_info "Getting maestro resource: $name"
  maestro-cli get --output json --consumer "$HF_MAESTRO_CONSUMER" --name "$name" --grpc-endpoint "$HF_MAESTRO_GRPC_ENDPOINT" --http-endpoint "$HF_MAESTRO_HTTP_ENDPOINT"
}

if [[ -n "${1:-}" ]]; then
  get_resource "$1"
else
  # Get the list of resources
  output=$(bash "$SCRIPT_DIR/hf.maestro.list.sh")

  # Extract names from JSON output
  names=$(echo "$output" | jq -r '.[].name' 2>/dev/null)

  if [[ -z "$names" ]]; then
    echo "No maestro resources found." >&2
    exit 1
  fi

  # Build array from names (compatible with bash 3)
  name_array=()
  while IFS= read -r line; do
    name_array+=("$line")
  done <<< "$names"

  # If only one result, use it directly
  if [[ ${#name_array[@]} -eq 1 ]]; then
    get_resource "${name_array[0]}"
    exit 0
  fi

  echo "Select a maestro resource to get:"
  echo ""
  for i in "${!name_array[@]}"; do
    echo "  $((i + 1))) ${name_array[$i]}"
  done
  echo ""

  read -rp "Enter number (1-${#name_array[@]}): " selection

  if ! [[ "$selection" =~ ^[0-9]+$ ]] || ((selection < 1 || selection > ${#name_array[@]})); then
    echo "Invalid selection." >&2
    exit 1
  fi

  selected_name="${name_array[$((selection - 1))]}"
  get_resource "$selected_name"
fi
