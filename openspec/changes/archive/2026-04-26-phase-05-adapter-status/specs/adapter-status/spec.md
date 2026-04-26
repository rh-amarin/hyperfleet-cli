# Delta Spec: Adapter Status (Phase 05)

**Status:** MODIFIED — extends `openspec/specs/adapter-status/spec.md`

---

## Go Command Interface

### hf cluster adapter post-status

```
hf cluster adapter post-status <adapter_name> <True|False|Unknown> [generation]
```

- `adapter_name` — required; name of the adapter to report for (e.g., `cl-deployment`, `cl-namespace`)
- `True|False|Unknown` — required; status value for all 3 condition types
- `generation` — optional; observed generation as integer (default: `1`)

Requires `cluster-id` in state (`~/.config/hf/state.yaml`). No explicit cluster-id override arg.

### hf nodepool adapter post-status

```
hf nodepool adapter post-status <adapter_name> <True|False|Unknown> [generation] [nodepool_id]
```

- `adapter_name`, `True|False|Unknown`, `generation` — same as cluster variant
- `nodepool_id` — optional 4th arg; overrides the nodepool-id from state

Requires both `cluster-id` and `nodepool-id` in state (or explicit `nodepool_id` arg).

---

## Payload Contract

Both commands POST the following JSON body (field values substituted at runtime):

```json
{
  "adapter": "<adapter_name>",
  "observed_generation": <generation>,
  "observed_time": "<ISO8601 UTC timestamp>",
  "conditions": [
    {
      "type": "Available",
      "status": "<True|False|Unknown>",
      "reason": "ManualStatusPost",
      "message": "Status posted via hf adapter post-status",
      "last_transition_time": "<ISO8601 UTC timestamp>"
    },
    {
      "type": "Applied",
      "status": "<True|False|Unknown>",
      "reason": "ManualStatusPost",
      "message": "Status posted via hf adapter post-status",
      "last_transition_time": "<ISO8601 UTC timestamp>"
    },
    {
      "type": "Health",
      "status": "<True|False|Unknown>",
      "reason": "ManualStatusPost",
      "message": "Status posted via hf adapter post-status",
      "last_transition_time": "<ISO8601 UTC timestamp>"
    }
  ]
}
```

`observed_time` and `last_transition_time` per condition are set to the same `time.Now().UTC().Format(time.RFC3339)` value computed once at invocation. `data`, `created_time`, and `last_report_time` are server-computed and appear only in the response.

**HTTP 204 behaviour:** When posting `Unknown` status the API returns 204 No Content with an empty body. The CLI handles this gracefully (prints a zero-value JSON object) via the `decode` helper's 204 guard.

---

## API Endpoints

| Target | Method | Path |
|---|---|---|
| Cluster | POST | `/api/hyperfleet/{version}/clusters/{cluster_id}/statuses` |
| NodePool | POST | `/api/hyperfleet/{version}/clusters/{cluster_id}/nodepools/{nodepool_id}/statuses` |

**Note:** The implementation plan originally stated `/adapter-statuses` for the POST path. Live verification showed the correct POST path is `/statuses` (without the `adapter-` prefix). The GET path for reading statuses remains `adapter-statuses` as used by `hf cluster statuses`.

Response: the created `AdapterStatus` object (JSON), or empty object on HTTP 204 (returned by the API for `Unknown` status). Printed to stdout via `printer().Print()`.

---

## Validation Rules

- `status` MUST be exactly one of: `True`, `False`, `Unknown` (case-sensitive)
- If `status` is invalid: print error to stderr, exit non-zero, no HTTP request made
- If `adapter_name` is missing (zero args): Cobra `RangeArgs` rejects the call before `RunE`

---

## Go Struct Changes (MODIFIED vs. phase-02-api-client)

### resource.ConditionRequest (internal/resource/condition.go)

Added field:

```go
LastTransitionTime string `json:"last_transition_time,omitempty"`
```

**Rationale:** The verified API execution logs show conditions in the POST body include `last_transition_time`. The original phase-2 definition was incomplete.

### resource.AdapterStatusCreateRequest (internal/resource/adapter.go)

Added fields (response-side, received from API):

```go
CreatedTime    string `json:"created_time,omitempty"`
LastReportTime string `json:"last_report_time,omitempty"`
```

`ObservedTime` tag changed from required to `omitempty`. Live verification confirmed the API expects `observed_time` in the request body (not `created_time`/`last_report_time` — those are server-computed and only appear in responses). `Data` tag kept as `omitempty`; the field is not sent in the request.

Also added to `internal/api/methods.go`: 204 No Content guard in `decode[T]` — returns zero value of `T` instead of an EOF error when the API responds with no body.

---

## Convergence Reference (unchanged from base spec)

Required cluster adapters (must all report `Available=True` at current generation for cluster `Ready=True`):
`cl-deployment`, `cl-invalid-resource`, `cl-job`, `cl-maestro`, `cl-namespace`, `cl-precondition-error`

Required nodepool adapter:
`np-configmap`
