#!/bin/bash
# create a new script hf.workflow.sh that does the following steps using the scripts:
#
# - first generate a random name to identify the test
# - create a cluster
# - show cluster table
# - echo a message that says "Ready should be False"
# - post cluster adapter status for adapter1 True at generation 1
# - post cluster adapter status for adapter2 True at generation 1
# - show cluster table
# - echo a message that says "Ready should be True"
# - patch cluster labels
# - show cluster table
# - echo a message that says "Ready should be False"
# - post cluster adapter status for adapter1 True at generation 2
# - post cluster adapter status for adapter2 True at generation 2
# - show cluster table
# - echo a message that says "Ready should be True"
# - patch cluster spec
# - show cluster table
# - echo a message that says "Ready should be False"
# - post cluster adapter status for adapter1 True at generation 3
# - post cluster adapter status for adapter2 True at generation 3
# - show cluster table
# - echo a message that says "Ready should be True"
#
# - create a nodepool
# - show nodepool table
# - echo a message that says "Ready should be False"
# - post nodepool adapter status for adapter3 True at generation 1
# - show nodepool table
# - echo a message that says "Ready should be True"
# - patch nodepool labels
# - show nodepool table
# - echo a message that says "Ready should be False"
# - post nodepool adapter status for adapter3 True at generation 2
# - show nodepool table
# - echo a message that says "Ready should be True"
# - patch nodepool spec
# - show nodepool table
# - echo a message that says "Ready should be False"
# - post nodepool adapter status for adapter3 True at generation 3
# - show nodepool table
# - echo a message that says "Ready should be True"
#
# - delete the cluster
# - show cluster table
# - echo a message that says "Ready should be False"
# - post cluster adapter status for adapter1 True at generation 4
# - post cluster adapter status for adapter2 True at generation 4
# - show cluster table
# - echo a message that says "Ready should be True"
#
# - show nodpool table
# - echo a message that says "Ready should be False"
# - post nodepool adapter status for adapter3 True at generation 4
# - show nodepool table
# - echo a message that says "Ready should be True"
#
# copy this prompt verbatim as a comment to the beginning of that script in case I want to regenerate this command

source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config api-url api-version

DIR="$(dirname "$(realpath "$0")")"

_step() {
  echo ""
  echo -e "${BOLD}${CYAN}==> $*${NC}"
}

_expect() {
  echo -e "${BOLD}${YELLOW}>>> $*${NC}"
}

# Generate a random name for this test run
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

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be False"

_step "Posting adapter1 status True (gen 1)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter1 True 1
_step "Posting adapter2 status True (gen 1)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter2 True 1

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be True"

_step "Patching cluster labels"
"$DIR/hf.cluster.patch.sh" labels

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be False"

_step "Posting adapter1 status True (gen 2)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter1 True 2
_step "Posting adapter2 status True (gen 2)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter2 True 2

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be True"

_step "Patching cluster spec"
"$DIR/hf.cluster.patch.sh" spec

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be False"

_step "Posting adapter1 status True (gen 3)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter1 True 3
_step "Posting adapter2 status True (gen 3)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter2 True 3

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be True"

# ---------------------------------------------------------------------------
# NodePool lifecycle
# ---------------------------------------------------------------------------

_step "Creating nodepool: $NODEPOOL_NAME"
"$DIR/hf.nodepool.create.sh" "$NODEPOOL_NAME"

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be False"

_step "Posting adapter3 status True (gen 1)"
"$DIR/hf.nodepool.adapter.post.status.sh" adapter3 True 1

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be True"

_step "Patching nodepool labels"
"$DIR/hf.nodepool.patch.sh" labels

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be False"

_step "Posting adapter3 status True (gen 2)"
"$DIR/hf.nodepool.adapter.post.status.sh" adapter3 True 2

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be True"

_step "Patching nodepool spec"
"$DIR/hf.nodepool.patch.sh" spec

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be False"

_step "Posting adapter3 status True (gen 3)"
"$DIR/hf.nodepool.adapter.post.status.sh" adapter3 True 3

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be True"

# ---------------------------------------------------------------------------
# Cluster deletion
# ---------------------------------------------------------------------------

_step "Deleting cluster: $CLUSTER_NAME"
"$DIR/hf.cluster.delete.sh"

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be False"

_step "Posting adapter1 status True (gen 4)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter1 True 4
_step "Posting adapter2 status True (gen 4)"
"$DIR/hf.cluster.adapter.post.status.sh" adapter2 True 4

_step "Cluster table"
"$DIR/hf.cluster.table.sh"
_expect "Ready should be True"

# ---------------------------------------------------------------------------
# NodePool post-deletion
# ---------------------------------------------------------------------------

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be False"

_step "Posting adapter3 status True (gen 4)"
"$DIR/hf.nodepool.adapter.post.status.sh" adapter3 True 4

_step "NodePool table"
"$DIR/hf.nodepool.table.sh"
_expect "Ready should be True"
