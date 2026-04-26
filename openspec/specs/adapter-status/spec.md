# Adapter Status Specification

## Purpose

Provide CLI commands to simulate adapter status reporting for clusters and nodepools. These commands allow manual posting of adapter conditions to the HyperFleet API, enabling testing of the convergence logic without running real adapters.

## Requirements

### Requirement: Post Cluster Adapter Status

The CLI SHALL post adapter status conditions for the current cluster.

#### Scenario: Post status with True

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster adapter post-status <adapter_name> True [generation]`
- THEN the CLI MUST send POST to `/api/hyperfleet/v1/clusters/{cluster_id}/statuses`
- AND the request payload MUST include:
  - `adapter`: the adapter name (e.g., `cl-deployment`, `cl-namespace`)
  - `conditions`: an array of 3 conditions with types `Available`, `Applied`, `Health`, all with status `True`
  - `observed_generation`: the provided generation (default: 1)
  - `observed_time`: current ISO8601 timestamp
  - `last_transition_time` per condition: current ISO8601 timestamp
- AND each condition MUST have `reason: "ManualStatusPost"` and `message: "Status posted via hf adapter post-status"`

#### Scenario: Post status with False

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster adapter post-status <adapter_name> False [generation]`
- THEN all 3 condition statuses MUST be set to `False`

#### Scenario: Post status with Unknown

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster adapter post-status <adapter_name> Unknown [generation]`
- THEN all 3 condition statuses MUST be set to `Unknown`
- AND the API returns HTTP 204 No Content; the CLI MUST handle this gracefully (exit 0, print empty object)

#### Scenario: Missing required arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf cluster adapter post-status`
- THEN the CLI MUST display usage information
- AND exit with code 1

#### Scenario: Invalid status value

- GIVEN an invalid status value is provided (not `True`, `False`, or `Unknown`)
- WHEN the user runs `hf cluster adapter post-status <adapter> <invalid>`
- THEN the CLI MUST return an error before making any HTTP request
- AND exit with non-zero code

### Requirement: Post NodePool Adapter Status

The CLI SHALL post adapter status conditions for a nodepool.

#### Scenario: Post nodepool adapter status

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool adapter post-status <adapter_name> <True|False|Unknown> [generation] [nodepool_id]`
- THEN the CLI MUST send POST to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}/statuses`
- AND the payload MUST follow the same structure as cluster adapter status posting
- AND the adapter name for nodepools is typically `np-configmap`
- AND the optional `nodepool_id` 4th arg overrides the nodepool-id from state

#### Scenario: Nodepool convergence after all adapters report

- GIVEN a nodepool's only required adapter is `np-configmap`
- WHEN `np-configmap` reports `Available=True` at the nodepool's current generation
- THEN the nodepool's `Ready` condition MUST flip to `True`
- AND the `Available` condition MUST flip to `True`

### Requirement: Adapter Status Model

The system SHALL follow a defined convergence model for adapter statuses.

#### Scenario: Cluster convergence

- GIVEN a cluster with required adapters: `cl-deployment`, `cl-invalid-resource`, `cl-job`, `cl-maestro`, `cl-namespace`, `cl-precondition-error`
- WHEN ALL required adapters report `Available=True` at the cluster's current generation
- THEN the cluster's `Ready` condition MUST become `True`
- AND each adapter MUST generate a per-adapter condition named `<AdapterName>Successful` (e.g., `ClDeploymentSuccessful`)

#### Scenario: Nodepool convergence

- GIVEN a nodepool with required adapter: `np-configmap`
- WHEN ALL required adapters report `Available=True` at the nodepool's current generation
- THEN the nodepool's `Ready` condition MUST become `True`

#### Scenario: Partial adapter reporting

- GIVEN some but not all required adapters have reported
- WHEN the user queries conditions
- THEN `Ready` MUST remain `False` with reason `MissingRequiredAdapters`
- AND the message MUST list which adapters are missing

---

## Go Command Interface (added in phase-05-adapter-status)

### hf cluster adapter post-status

```
hf cluster adapter post-status <adapter_name> <True|False|Unknown> [generation]
```

- `adapter_name` — required
- `True|False|Unknown` — required; case-sensitive
- `generation` — optional integer, default `1`

Requires `cluster-id` in state (`~/.config/hf/state.yaml`). No explicit cluster-id override arg.

### hf nodepool adapter post-status

```
hf nodepool adapter post-status <adapter_name> <True|False|Unknown> [generation] [nodepool_id]
```

- `nodepool_id` — optional 4th arg; overrides the nodepool-id from state

Requires both `cluster-id` and `nodepool-id` in state (or explicit `nodepool_id` arg).

## API Endpoints

| Target | Method | Path |
|---|---|---|
| Cluster | POST | `/api/hyperfleet/{version}/clusters/{cluster_id}/statuses` |
| NodePool | POST | `/api/hyperfleet/{version}/clusters/{cluster_id}/nodepools/{nodepool_id}/statuses` |

Note: the POST path is `/statuses` (not `/adapter-statuses`). The GET path for reading all adapter statuses remains `/adapter-statuses` as used by `hf cluster statuses`.

## Request Payload

```json
{
  "adapter": "<adapter_name>",
  "observed_generation": <generation>,
  "observed_time": "<ISO8601 UTC>",
  "conditions": [
    {"type": "Available", "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status", "last_transition_time": "<ISO8601 UTC>"},
    {"type": "Applied",   "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status", "last_transition_time": "<ISO8601 UTC>"},
    {"type": "Health",    "status": "<status>", "reason": "ManualStatusPost", "message": "Status posted via hf adapter post-status", "last_transition_time": "<ISO8601 UTC>"}
  ]
}
```

`data`, `created_time`, and `last_report_time` are server-computed; they appear in the response only.

HTTP 204 is returned by the API for `Unknown` status — the CLI handles this gracefully via a 204 guard in `internal/api/methods.go decode[T]`.

## Go Struct Changes

- `resource.ConditionRequest` — added `LastTransitionTime string \`json:"last_transition_time,omitempty"\``
- `resource.AdapterStatusCreateRequest` — added `CreatedTime`, `LastReportTime` (response-side); `ObservedTime` changed to `omitempty`
- `internal/api/methods.go decode[T]` — returns zero value of T on HTTP 204 instead of EOF error
