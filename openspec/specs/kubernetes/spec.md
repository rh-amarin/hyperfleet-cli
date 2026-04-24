# Kubernetes Utilities Specification

## Purpose

Provide CLI commands for Kubernetes-related operations including port-forwarding to HyperFleet services, in-cluster curl execution, debug pod creation, and log tailing for pods and adapters.

## Requirements

### Requirement: Port Forward Management

The CLI SHALL manage port forwards to HyperFleet services running in Kubernetes.

#### Scenario: Start port forwards

- GIVEN context, namespace, and port-forward config properties are set
- WHEN the user runs `hf kube port-forward start`
- THEN the CLI MUST start background port-forward processes for:
  - `hyperfleet-api`: localhost:{pf-api-port} to the hyperfleet-api service in the configured namespace
  - `postgresql`: localhost:{pf-pg-port} to the postgres service in the configured namespace
  - `maestro-http`: localhost:{pf-maestro-http-port} to the maestro service (remote port {pf-maestro-http-remote-port}) in the maestro-namespace
  - `maestro-grpc`: localhost:{pf-maestro-grpc-port} to the maestro service in the maestro-namespace

#### Scenario: Stop port forwards

- GIVEN port forwards are running
- WHEN the user runs `hf kube port-forward stop`
- THEN the CLI MUST terminate all running port-forward processes

#### Scenario: Check port forward status

- GIVEN the CLI is running
- WHEN the user runs `hf kube port-forward status`
- THEN the CLI MUST display the status of each port forward with:
  - A colored indicator: green = running, red = not running
  - Service name
  - Local port
  - PID (if running)

### Requirement: Kubectl Context Selection

The CLI SHALL provide interactive selection of kubectl context and namespace.

#### Scenario: Select context and namespace

- GIVEN kubectl is installed
- WHEN the user runs `hf kube context`
- THEN the CLI MUST list available kubectl contexts
- AND prompt the user to select one
- AND list namespaces in the selected context
- AND prompt the user to select one
- AND save both selections to `~/.config/hf/context` and `~/.config/hf/namespace`

### Requirement: In-Cluster Curl

The CLI SHALL execute curl commands from inside the Kubernetes cluster.

#### Scenario: Run curl from in-cluster pod

- GIVEN kubectl context is configured
- WHEN the user runs `hf kube curl [options] <url>`
- THEN the CLI MUST create or reuse an ephemeral curl pod in the cluster
- AND execute the curl command with the provided options (-H headers, -d data, -F file, -v verbose)
- AND display the curl output

### Requirement: Debug Pod Creation

The CLI SHALL create debug pods from existing deployment templates.

#### Scenario: Create debug pod

- GIVEN a deployment exists in the cluster
- WHEN the user runs `hf kube debug <partial-deployment-name> [namespace]`
- THEN the CLI MUST find a deployment matching the partial name
- AND create a pod using the same spec but with `restartPolicy: Never`
- AND exec into the pod for interactive debugging

### Requirement: Pod Log Tailing

The CLI SHALL tail logs from pods matching a name pattern.

#### Scenario: Tail logs for matching pods

- GIVEN pods exist matching the pattern
- WHEN the user runs `hf logs <pod-name-pattern> [kubectl-logs-options...]`
- THEN the CLI MUST find pods matching the pattern
- AND if stern is available and multiple pods match, use stern for multi-pod tailing
- AND if stern is not available, prompt the user to select a pod
- AND pass through kubectl logs options (-f, --since, --tail, etc.)

### Requirement: Adapter Log Tailing

The CLI SHALL tail adapter logs filtered by the current cluster ID.

#### Scenario: Tail adapter logs

- GIVEN cluster-id is configured
- WHEN the user runs `hf logs adapter <adapter-name-pattern> [--since=<duration>] [--tail=<lines>] [-f]`
- THEN the CLI MUST find pods matching the adapter name pattern
- AND filter log output to show only entries matching the current cluster-id
- AND display only timestamp and message fields
- AND default to `--since=1h` if not specified
