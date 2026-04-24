#!/bin/bash
# Delete a maestro resource by name, or interactively select one from the list
# Usage: hf.maestro.delete.sh [name]
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config maestro-consumer maestro-http-endpoint maestro-grpc-endpoint

SCRIPT_DIR="$(dirname "$(realpath "$0")")"

delete_resource() {
  local name="$1"
  hf_info "Deleting maestro resource: $name"
  maestro-cli delete --consumer "$HF_MAESTRO_CONSUMER" --name "$name" --grpc-endpoint "$HF_MAESTRO_GRPC_ENDPOINT" --http-endpoint "$HF_MAESTRO_HTTP_ENDPOINT"
}

if [[ -n "${1:-}" ]]; then
  delete_resource "$1"
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

  echo "Select a maestro resource to delete:"
  echo ""
  for i in "${!name_array[@]}"; do
    echo "  $((i + 1))) ${name_array[$i]}"
  done
  echo ""

  read -rp "Enter number (1-${#name_array[@]}): " selection

  if ! [[ "$selection" =~ ^[0-9]+$ ]] || (( selection < 1 || selection > ${#name_array[@]} )); then
    echo "Invalid selection." >&2
    exit 1
  fi

  selected_name="${name_array[$((selection - 1))]}"
  delete_resource "$selected_name"
fi
