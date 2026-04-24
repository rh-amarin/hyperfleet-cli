#!/bin/bash

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <partial-deployment-name> [namespace]"
    echo "Example: $0 my-app default"
    exit 1
fi

PARTIAL_NAME="$1"
NAMESPACE="${2:-$(kubectl config view --minify -o jsonpath='{..namespace}' 2>/dev/null || echo 'default')}"

# Find the deployment matching the partial name
DEPLOYMENT=$(kubectl get deployments -n "$NAMESPACE" -o name 2>/dev/null | grep -i "$PARTIAL_NAME" | head -1)

if [ -z "$DEPLOYMENT" ]; then
    echo "Error: No deployment found matching '$PARTIAL_NAME' in namespace '$NAMESPACE'"
    echo "Available deployments:"
    kubectl get deployments -n "$NAMESPACE" -o name
    exit 1
fi

DEPLOYMENT_NAME=$(echo "$DEPLOYMENT" | sed 's|deployment.apps/||')
echo "Found deployment: $DEPLOYMENT_NAME"

# Generate a unique debug pod name
DEBUG_POD_NAME="${DEPLOYMENT_NAME}-debug-$(date +%s)"

# Get the pod template from the deployment and create a debug pod
echo "Creating debug pod: $DEBUG_POD_NAME"

kubectl get deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" -o json | \
    jq --arg name "$DEBUG_POD_NAME" '
    .spec.template |
    .metadata.name = $name |
    .metadata.labels["debug-pod"] = "true" |
    del(.metadata.labels["pod-template-hash"]) |
    .spec.restartPolicy = "Never" |
    .spec.containers[0].command = ["/bin/sh", "-c", "echo Debug pod ready; sleep infinity"] |
    del(.spec.containers[0].args) |
    del(.spec.containers[0].livenessProbe) |
    del(.spec.containers[0].readinessProbe) |
    del(.spec.containers[0].startupProbe) |
    .kind = "Pod" |
    .apiVersion = "v1"
    ' | kubectl apply -n "$NAMESPACE" -f -

# Wait for the pod to be ready
echo "Waiting for pod to be ready..."
kubectl wait --for=condition=Ready pod/"$DEBUG_POD_NAME" -n "$NAMESPACE" --timeout=120s

# Get the container name
CONTAINER_NAME=$(kubectl get pod "$DEBUG_POD_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.containers[0].name}')

echo ""
echo "Debug pod '$DEBUG_POD_NAME' is ready."
echo "Exec'ing into container '$CONTAINER_NAME'..."
echo "To delete this pod later: kubectl delete pod $DEBUG_POD_NAME -n $NAMESPACE"
echo ""

# Exec into the pod - try bash first, fall back to sh
kubectl exec -it "$DEBUG_POD_NAME" -n "$NAMESPACE" -c "$CONTAINER_NAME" -- /bin/bash 2>/dev/null || \
    kubectl exec -it "$DEBUG_POD_NAME" -n "$NAMESPACE" -c "$CONTAINER_NAME" -- /bin/sh
