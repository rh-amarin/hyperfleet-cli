#!/bin/bash
# Curl from inside a Kubernetes pod in the current cluster
source "$(dirname "$(realpath "$0")")/hf.lib.sh"

hf_require_kubectl

set -euo pipefail

POD_NAME="hf-kcurl"

usage() {
  hf_usage "[options] <url>"
  cat <<'EOF'
Spin up a pod in the current Kubernetes cluster and execute a curl command from
inside it. The pod is reused across invocations and auto-terminates after 5
minutes of inactivity.

Options:
  -X <method>     HTTP method (GET, POST, PUT, DELETE, etc.) — default: GET
  -H <header>     HTTP header (repeatable)
  -d <data>       Inline request body data
  -f <file>       File to send as request body (uploaded to pod via kubectl cp)
  -o <file>       Write response to local file instead of stdout
  -i              Include response headers in output
  -v              Verbose curl output
  -k              Allow insecure TLS connections
  --image <img>   Override container image (default: curlimages/curl)
  --help          Show this help message
  --              Pass remaining args directly to curl
EOF
  exit 0
}

# Defaults
METHOD=""
HEADERS=()
DATA=""
FILE=""
OUTPUT=""
INCLUDE_HEADERS=false
VERBOSE=false
INSECURE=false
IMAGE="curlimages/curl"
EXTRA_CURL_ARGS=()
URL=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -X)
      METHOD="$2"; shift 2 ;;
    -H)
      HEADERS+=("$2"); shift 2 ;;
    -d)
      DATA="$2"; shift 2 ;;
    -f)
      FILE="$2"; shift 2 ;;
    -o)
      OUTPUT="$2"; shift 2 ;;
    -i)
      INCLUDE_HEADERS=true; shift ;;
    -v)
      VERBOSE=true; shift ;;
    -k)
      INSECURE=true; shift ;;
    --image)
      IMAGE="$2"; shift 2 ;;
    --help)
      usage ;;
    --)
      shift; EXTRA_CURL_ARGS+=("$@"); break ;;
    -*)
      hf_die "Unknown option: $1" ;;
    *)
      URL="$1"; shift ;;
  esac
done

if [[ -z "$URL" ]]; then
  hf_die "Missing required <url> argument. Run '$(basename "$0") --help' for usage."
fi

# Ensure the shared curl pod is running, creating or replacing as needed.
ensure_pod() {
  local phase
  phase="$(hf_kubectl get pod "$POD_NAME" -o jsonpath='{.status.phase}' 2>/dev/null || true)"

  if [[ "$phase" == "Running" ]]; then
    return
  fi

  if [[ -n "$phase" ]]; then
    # Pod exists but is not Running (Succeeded, Failed, Pending, Unknown)
    hf_kubectl delete pod "$POD_NAME" --grace-period=0 --force --wait=false >/dev/null 2>&1 || true
  fi

  hf_kubectl run "$POD_NAME" \
    --image="$IMAGE" \
    --restart=Never \
    --quiet \
    --command -- sleep 300 >/dev/null 2>&1

  hf_kubectl wait --for=condition=Ready "pod/$POD_NAME" --timeout=30s >/dev/null 2>&1
  : # pod created, expires in 5m
}

ensure_pod

# Build curl arguments
CURL_ARGS=(-sS)

if [[ -n "$METHOD" ]]; then
  CURL_ARGS+=(-X "$METHOD")
fi

for h in "${HEADERS[@]+"${HEADERS[@]}"}"; do
  CURL_ARGS+=(-H "$h")
done

if $INCLUDE_HEADERS; then
  CURL_ARGS+=(-i)
fi

if $VERBOSE; then
  CURL_ARGS+=(-v)
fi

if $INSECURE; then
  CURL_ARGS+=(-k)
fi

# File upload path: copy file into pod, then exec curl
if [[ -n "$FILE" ]]; then
  if [[ ! -f "$FILE" ]]; then
    hf_die "File not found: $FILE"
  fi

  BASENAME="$(basename "$FILE")"
  hf_kubectl cp "$FILE" "$POD_NAME:/tmp/$BASENAME"

  CURL_ARGS+=(--data-binary "@/tmp/$BASENAME")
  CURL_ARGS+=("${EXTRA_CURL_ARGS[@]+"${EXTRA_CURL_ARGS[@]}"}")
  CURL_ARGS+=("$URL")

elif [[ -n "$DATA" ]]; then
  CURL_ARGS+=(-d "$DATA")
  CURL_ARGS+=("${EXTRA_CURL_ARGS[@]+"${EXTRA_CURL_ARGS[@]}"}")
  CURL_ARGS+=("$URL")

else
  CURL_ARGS+=("${EXTRA_CURL_ARGS[@]+"${EXTRA_CURL_ARGS[@]}"}")
  CURL_ARGS+=("$URL")
fi

# Execute curl on the shared pod
if [[ -n "$OUTPUT" ]]; then
  hf_kubectl exec "$POD_NAME" -- curl "${CURL_ARGS[@]}" > "$OUTPUT"
else
  hf_kubectl exec "$POD_NAME" -- curl "${CURL_ARGS[@]}"
fi
