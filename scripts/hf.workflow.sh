#!/bin/bash
# Real adapter workflow: exercises the full cluster/nodepool lifecycle and waits
# for real adapters to report status instead of simulating them via API posts.
#
# Steps:
# - generate a random name to identify the test
# - create a cluster
# - show cluster table
# - wait for Ready=True  (adapters converge)
# - patch cluster labels
# - show cluster table
# - wait for Ready=False (generation bump makes it not-ready)
# - wait for Ready=True  (adapters reconverge)
# - patch cluster spec
# - show cluster table
# - wait for Ready=False
# - wait for Ready=True
# - create a nodepool
# - show nodepool table
# - wait for nodepool Ready=True
# - patch nodepool labels
# - show nodepool table
# - wait for nodepool Ready=False
# - wait for nodepool Ready=True
# - patch nodepool spec
# - show nodepool table
# - wait for nodepool Ready=False
# - wait for nodepool Ready=True
# - delete the cluster
# - show cluster table
# - wait for cluster Ready=True  (deletion acknowledged)
# - show nodepool table
# - wait for nodepool Ready=True (deletion acknowledged)

source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

hf_require_jq

DIR="$(dirname "$(realpath "$0")")"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_step() {
  echo ""
  echo -e "${BOLD}${CYAN}==> $*${NC}"
}

_expect() {
  echo -e "${BOLD}${YELLOW}>>> Expected: $*${NC}"
}

# Poll until status.conditions[type=Ready].status equals $3.
# Usage: _wait_ready <cluster|nodepool> <id> <True|False> [timeout_secs]
_wait_ready() {
  local resource="$1"
  local id="$2"
  local expected="$3"
  local timeout="${4:-120}"
  local interval=3
  local elapsed=0
  local url

  case "$resource" in
    cluster)
      url="${HF_API_URL}/api/hyperfleet/${HF_API_VERSION}/clusters/${id}"
      ;;
    nodepool)
      local cluster_id
      cluster_id=$(hf_cluster_id)
      url="${HF_API_URL}/api/hyperfleet/${HF_API_VERSION}/clusters/${cluster_id}/nodepools/${id}"
      ;;
    *)
      hf_die "Unknown resource type: $resource"
      ;;
  esac

  hf_info "Waiting for $resource Ready=$expected (timeout ${timeout}s)..."

  while true; do
    local status
    status=$(curl -s "$url" | jq -r '
      .status.conditions // [] |
      map(select(.type == "Ready")) |
      .[0].status // "Unknown"
    ')

    if [[ "$status" == "$expected" ]]; then
      hf_info "  $resource Ready=$status  ✓"
      return 0
    fi

    if (( elapsed >= timeout )); then
      hf_error "Timed out after ${timeout}s waiting for $resource Ready=$expected (current: $status)"
      return 1
    fi

    printf "\r  %s Ready=%-7s  waiting... %ds" "$resource" "$status" "$elapsed"
    sleep "$interval"
    elapsed=$(( elapsed + interval ))
  done
}

# ---------------------------------------------------------------------------
# Setup
# ---------------------------------------------------------------------------

TEST_ID="$(openssl rand -hex 4)"
CLUSTER_NAME="test-${TEST_ID}"
NODEPOOL_NAME="test-${TEST_ID}"

hf_info "Starting workflow: $TEST_ID"
hf_info "  Cluster:  $CLUSTER_NAME"
hf_info "  NodePool: $NODEPOOL_NAME"

# ---------------------------------------------------------------------------
# Cluster lifecycle
# ---------------------------------------------------------------------------

_step "Creating cluster: $CLUSTER_NAME"
"$DIR/hf.cluster.create.sh" "$CLUSTER_NAME"
CLUSTER_ID=$(hf_cluster_id)

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=False (just created)"

_step "Waiting for cluster to become Ready"
_wait_ready cluster "$CLUSTER_ID" True

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=True"

_step "Patching cluster labels"
"$DIR/hf.cluster.patch.sh" labels

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=False (generation bumped)"

_step "Waiting for cluster to become Ready again"
_wait_ready cluster "$CLUSTER_ID" True

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=True"

_step "Patching cluster spec"
"$DIR/hf.cluster.patch.sh" spec

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=False (generation bumped)"

_step "Waiting for cluster to become Ready again"
_wait_ready cluster "$CLUSTER_ID" True

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=True"

# ---------------------------------------------------------------------------
# NodePool lifecycle
# ---------------------------------------------------------------------------

_step "Creating nodepool: $NODEPOOL_NAME"
"$DIR/hf.nodepool.create.sh" "$NODEPOOL_NAME"
NODEPOOL_ID=$(hf_nodepool_id)

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=False (just created)"

_step "Waiting for nodepool to become Ready"
_wait_ready nodepool "$NODEPOOL_ID" True

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=True"

_step "Patching nodepool labels"
"$DIR/hf.nodepool.patch.sh" labels

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=False (generation bumped)"

_step "Waiting for nodepool to become Ready again"
_wait_ready nodepool "$NODEPOOL_ID" True

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=True"

_step "Patching nodepool spec"
"$DIR/hf.nodepool.patch.sh" spec

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=False (generation bumped)"

_step "Waiting for nodepool to become Ready again"
_wait_ready nodepool "$NODEPOOL_ID" True

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=True"

# ---------------------------------------------------------------------------
# Cluster deletion
# ---------------------------------------------------------------------------

_step "Deleting cluster: $CLUSTER_NAME"
"$DIR/hf.cluster.delete.sh"

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=False (deletion in progress)"

_step "Waiting for cluster deletion to be acknowledged"
_wait_ready cluster "$CLUSTER_ID" True

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready=True"

# ---------------------------------------------------------------------------
# NodePool post-deletion
# ---------------------------------------------------------------------------

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=False (cluster deleted)"

_step "Waiting for nodepool deletion to be acknowledged"
_wait_ready nodepool "$NODEPOOL_ID" True

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready=True"
