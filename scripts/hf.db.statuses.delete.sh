#!/bin/bash
# Delete adapter statuses from PostgreSQL
source "$(dirname "$(realpath "$0")")/hf.lib.sh"
hf_require_config context namespace cluster-id

CLUSTER_ID=$(hf_cluster_id)
POD=$(hf_find_postgres_pod)

hf_info "Deleting adapter statuses from PostgreSQL"
echo "  Context: $HF_KUBE_CONTEXT"
echo "  Namespace: $HF_KUBE_NAMESPACE"
echo "  Pod: $POD"
echo "  Cluster ID: $CLUSTER_ID"
echo ""

hf_kubectl_ns exec -it "$POD" -- \
    psql -U hyperfleet -d hyperfleet -c \
    "DELETE FROM adapter_statuses WHERE resource_type = 'cluster' AND resource_id = '${CLUSTER_ID}';"
