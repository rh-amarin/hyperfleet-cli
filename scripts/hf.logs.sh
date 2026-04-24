#!/bin/bash
# Tail logs from pods matching a partial name pattern

set -euo pipefail

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <pod-name-pattern> [kubectl-logs-options...]"
    echo "Example: $0 api"
    echo "         $0 api -c container-name"
    echo "         $0 api --since=1h"
    exit 1
fi

PATTERN="$1"
shift

# Get matching pods
PODS=$(kubectl get pods --no-headers -o custom-columns=":metadata.name" 2>/dev/null | grep -i "$PATTERN" || true)

if [[ -z "$PODS" ]]; then
    echo "No pods found matching pattern: $PATTERN"
    echo "Current context: $(kubectl config current-context)"
    echo "Current namespace: $(kubectl config view --minify -o jsonpath='{..namespace}')"
    exit 1
fi

POD_COUNT=$(echo "$PODS" | wc -l | tr -d ' ')

if [[ "$POD_COUNT" -eq 1 ]]; then
    POD_NAME=$(echo "$PODS" | head -1)
    echo "Tailing logs for pod: $POD_NAME"
    kubectl logs -f "$POD_NAME" "$@"
else
    echo "Found $POD_COUNT pods matching '$PATTERN':"
    echo "$PODS" | nl -w2 -s') '
    echo ""
    read -rp "Select pod number (or 'a' for all with stern if available): " CHOICE

    if [[ "$CHOICE" == "a" ]]; then
        if command -v stern &>/dev/null; then
            echo "Using stern to tail all matching pods..."
            stern "$PATTERN" "$@"
        else
            echo "stern not installed. Install with: brew install stern"
            echo "Falling back to first matching pod..."
            POD_NAME=$(echo "$PODS" | head -1)
            kubectl logs -f "$POD_NAME" "$@"
        fi
    elif [[ "$CHOICE" =~ ^[0-9]+$ ]] && [[ "$CHOICE" -ge 1 ]] && [[ "$CHOICE" -le "$POD_COUNT" ]]; then
        POD_NAME=$(echo "$PODS" | sed -n "${CHOICE}p")
        echo "Tailing logs for pod: $POD_NAME"
        kubectl logs -f "$POD_NAME" "$@"
    else
        echo "Invalid selection"
        exit 1
    fi
fi
