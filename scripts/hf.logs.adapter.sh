#!/bin/bash
# Get adapter pod logs filtered by cluster ID, displaying time and message

source "$(dirname "$(realpath "$0")")/hf.lib.sh"

hf_require_config cluster-id
hf_require_kubectl

if [[ $# -lt 1 ]]; then
  hf_usage "<adapter-name-pattern> [--since=<duration>] [--tail=<lines>] [-f]"
  echo "Arguments:"
  echo "  adapter-name-pattern  Partial pod/deployment name to match (e.g., 'validator', 'dns')"
  echo ""
  echo "Options:"
  echo "  -f, --follow         Follow log output (stream)"
  echo "  --since=<duration>   Show logs newer than duration (e.g., 1h, 30m). Default: 1h"
  echo "  --tail=<lines>       Limit initial lines returned"
  echo ""
  echo "Active config:"
  echo "  cluster-id: $(hf_config_value cluster-id)"
  echo "  context:    ${HF_KUBE_CONTEXT:-<default>}"
  echo "  namespace:  ${HF_KUBE_NAMESPACE:-<default>}"
  echo ""
  echo "Examples:"
  echo "  hf.logs.adapter.sh validator"
  echo "  hf.logs.adapter.sh dns --since=30m -f"
  echo "  hf.logs.adapter.sh provisioner --tail=200"
  exit 1
fi

CLUSTER_ID=$(hf_cluster_id)
PATTERN="$1"
shift

# Parse options — separate follow flag and kubectl pass-through args
KUBECTL_ARGS=()
HAS_SINCE=false
for arg in "$@"; do
  case "$arg" in
    -f | --follow) KUBECTL_ARGS+=("--follow") ;;
    --since=*) HAS_SINCE=true; KUBECTL_ARGS+=("$arg") ;;
    *) KUBECTL_ARGS+=("$arg") ;;
  esac
done
[[ "$HAS_SINCE" == false ]] && KUBECTL_ARGS+=("--since=1h")

hf_info "Searching pods matching: ${BOLD}${PATTERN}${NC}"
hf_info "Cluster ID filter:       ${CYAN}${CLUSTER_ID}${NC}"

# Find matching pods
PODS=$(hf_kubectl_ns get pods --no-headers -o custom-columns=":metadata.name" 2>/dev/null \
  | grep -i "$PATTERN" || true)

if [[ -z "$PODS" ]]; then
  hf_error "No pods found matching: $PATTERN"
  echo "  Context:   ${HF_KUBE_CONTEXT:-<default>}"
  echo "  Namespace: ${HF_KUBE_NAMESPACE:-<default>}"
  exit 1
fi

POD_COUNT=$(echo "$PODS" | wc -l | tr -d ' ')

if [[ "$POD_COUNT" -eq 1 ]]; then
  POD_NAME=$(echo "$PODS" | head -1)
else
  echo -e "\nFound ${BOLD}${POD_COUNT}${NC} pods matching '${BOLD}${PATTERN}${NC}':"
  echo "$PODS" | nl -w2 -s') '
  echo ""
  read -rp "Select pod number [1]: " CHOICE
  CHOICE="${CHOICE:-1}"
  if [[ "$CHOICE" =~ ^[0-9]+$ ]] && [[ "$CHOICE" -ge 1 ]] && [[ "$CHOICE" -le "$POD_COUNT" ]]; then
    POD_NAME=$(echo "$PODS" | sed -n "${CHOICE}p")
  else
    hf_die "Invalid selection: $CHOICE"
  fi
fi

hf_info "Pod: ${BOLD}${POD_NAME}${NC}"
echo ""
echo -e "${BOLD}--- pod: ${POD_NAME} | cluster: ${CLUSTER_ID} ---${NC}"
echo ""

# Extract time and msg from a single log line.
# Supports three formats (tried in order):
#   1. JSON object  — field names: time/ts/timestamp/@timestamp, msg/message/log
#   2. logfmt       — key=value or key="quoted value" (e.g. time=... msg="...")
#   3. Plain text   — printed as-is
format_log_line() {
  local line="$1"
  local time="" msg=""

  if [[ "$line" == "{"* ]]; then
    # ── JSON ──────────────────────────────────────────────────────────────────
    if [[ "$line" =~ \"time\":\"([^\"]+)\" ]];      then time="${BASH_REMATCH[1]}"
    elif [[ "$line" =~ \"ts\":\"([^\"]+)\" ]];       then time="${BASH_REMATCH[1]}"
    elif [[ "$line" =~ \"timestamp\":\"([^\"]+)\" ]]; then time="${BASH_REMATCH[1]}"
    elif [[ "$line" =~ \"@timestamp\":\"([^\"]+)\" ]]; then time="${BASH_REMATCH[1]}"
    fi

    if [[ "$line" =~ \"msg\":\"([^\"]+)\" ]];     then msg="${BASH_REMATCH[1]}"
    elif [[ "$line" =~ \"message\":\"([^\"]+)\" ]]; then msg="${BASH_REMATCH[1]}"
    elif [[ "$line" =~ \"log\":\"([^\"]+)\" ]];   then msg="${BASH_REMATCH[1]}"
    fi

  elif [[ "$line" =~ [[:space:]]msg=|^msg= ]]; then
    # ── logfmt ────────────────────────────────────────────────────────────────
    # time= (always unquoted in logfmt)
    if [[ "$line" =~ (^|[[:space:]])time=([^[:space:]]+) ]]; then
      time="${BASH_REMATCH[2]}"
    fi

    # msg= quoted or unquoted
    if [[ "$line" =~ (^|[[:space:]])msg=\"([^\"]+)\" ]]; then
      msg="${BASH_REMATCH[2]}"
    elif [[ "$line" =~ (^|[[:space:]])msg=([^[:space:]]+) ]]; then
      msg="${BASH_REMATCH[2]}"
    fi
  fi

  if [[ -n "$time" || -n "$msg" ]]; then
    echo -e "${CYAN}${time}${NC}  ${msg}"
  else
    echo "$line"
  fi
}

while IFS= read -r line; do
  format_log_line "$line"
done < <(hf_kubectl_ns logs "${KUBECTL_ARGS[@]}" "$POD_NAME" 2>/dev/null \
  | grep --line-buffered "$CLUSTER_ID")
