# Design: Phase 05 — Adapter Status

## Command Tree

```
hf cluster adapter post-status <adapter_name> <True|False|Unknown> [generation]
hf nodepool adapter post-status <adapter_name> <True|False|Unknown> [generation] [nodepool_id]
```

## New File: cmd/adapter.go

Contains the entire phase: `clusterAdapterCmd`, `nodepoolAdapterCmd`, and their `post-status` sub-subcommands. All registered in a single `init()`.

Registration:
```go
func init() {
    clusterCmd.AddCommand(clusterAdapterCmd)
    clusterAdapterCmd.AddCommand(clusterAdapterPostStatusCmd)

    nodepoolCmd.AddCommand(nodepoolAdapterCmd)   // nodepoolCmd defined in phase-04 cmd/nodepool.go
    nodepoolAdapterCmd.AddCommand(nodepoolAdapterPostStatusCmd)
}
```

`clusterCmd` is package-level in `cmd/cluster.go`; `nodepoolCmd` will be package-level in `cmd/nodepool.go` (phase 4). Both are in the same `cmd` package, so cross-file reference compiles without imports.

## Resource Type Updates (internal/resource)

### condition.go — extend ConditionRequest

Add `LastTransitionTime` to match the verified API contract:

```go
type ConditionRequest struct {
    Type               string `json:"type"`
    Status             string `json:"status"`
    Reason             string `json:"reason,omitempty"`
    Message            string `json:"message,omitempty"`
    LastTransitionTime string `json:"last_transition_time,omitempty"`
}
```

### adapter.go — extend AdapterStatusCreateRequest

Replace `observed_time` (bash-era field) with the fields the API actually expects, keeping `ObservedTime` for backward compatibility with any callers that may exist:

```go
type AdapterStatusCreateRequest struct {
    Adapter            string                 `json:"adapter"`
    ObservedGeneration int32                  `json:"observed_generation"`
    ObservedTime       string                 `json:"observed_time,omitempty"`   // kept, unused in phase 5
    Conditions         []ConditionRequest     `json:"conditions"`
    Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"`
    Data               map[string]any         `json:"data,omitempty"`
    CreatedTime        string                 `json:"created_time,omitempty"`
    LastReportTime     string                 `json:"last_report_time,omitempty"`
}
```

## Shared Helper: buildAdapterPostPayload

```go
func buildAdapterPostPayload(adapter, status string, generation int32) resource.AdapterStatusCreateRequest {
    now := time.Now().UTC().Format(time.RFC3339)
    msg := "Status posted via hf adapter post-status"
    conditions := []resource.ConditionRequest{
        {Type: "Available", Status: status, Reason: "ManualStatusPost", Message: msg, LastTransitionTime: now},
        {Type: "Applied",   Status: status, Reason: "ManualStatusPost", Message: msg, LastTransitionTime: now},
        {Type: "Health",    Status: status, Reason: "ManualStatusPost", Message: msg, LastTransitionTime: now},
    }
    return resource.AdapterStatusCreateRequest{
        Adapter:            adapter,
        ObservedGeneration: generation,
        Conditions:         conditions,
        Data:               map[string]any{},
        CreatedTime:        now,
        LastReportTime:     now,
    }
}
```

This is a package-private function in `cmd/adapter.go`, testable directly from `cmd/adapter_test.go`.

## Argument Validation

Validation runs before any API call or ID resolution:

```
status must be one of: True | False | Unknown
```

If `adapter` arg is missing or `status` is invalid: print usage to stderr, return error (non-zero exit). Use `cmd.SilenceUsage = false` so Cobra prints help on error.

Generation defaults to `1` if the optional third arg is absent. Parse with `strconv.Atoi`; if non-numeric, return a clear error.

## ID Resolution

| Command | Cluster ID | NodePool ID |
|---|---|---|
| cluster adapter post-status | `config.ClusterID(cfgStore, "")` — always from state | n/a |
| nodepool adapter post-status | `config.ClusterID(cfgStore, "")` — always from state | `config.NodePoolID(cfgStore, arg)` — arg = 4th positional or "" |

No explicit cluster-id override in either command (matches bash script behavior).

## API Endpoints

| Target | Method | Path |
|---|---|---|
| Cluster | POST | `clusters/{cluster_id}/adapter-statuses` |
| NodePool | POST | `clusters/{cluster_id}/nodepools/{nodepool_id}/adapter-statuses` |

Both use `api.Post[resource.AdapterStatus](c, ctx, path, payload)`. The response is the created `AdapterStatus` object, printed via `printer().Print(response)`.

## Output Behavior

```
[INFO] Posting adapter status: cl-deployment = True (gen: 3) for cluster: <cluster_id>
{ ...response JSON... }
```

- `[INFO]` line written to stderr via `out.Info()`
- Response printed to stdout via `printer().Print()` (respects `--output` flag)
- Default output format: JSON (same as other create-style commands)

## Unit Test Design (cmd/adapter_test.go)

Use `httptest.NewServer` to capture the outbound request:

```
TestClusterAdapterPostStatus_PostsCorrectURL
  - start httptest.Server
  - configure cfgStore with api-url pointing at test server, cluster-id set
  - run clusterAdapterPostStatusCmd with args ["cl-deployment", "True", "3"]
  - assert POST to /api/hyperfleet/{version}/clusters/{id}/adapter-statuses
  - assert body: adapter="cl-deployment", 3 conditions all status="True", data={}

TestNodePoolAdapterPostStatus_PostsCorrectURL
  - same pattern, with nodepool-id also set in cfgStore
  - assert POST to .../nodepools/{nid}/adapter-statuses

TestAdapterPostStatus_InvalidStatus_ReturnsError
  - run command with status="maybe"
  - assert error returned, no HTTP request reached the server

TestAdapterPostStatus_MissingArgs_ReturnsError
  - run command with zero args
  - assert error returned
```

## Cobra Command Signatures

```go
var clusterAdapterCmd = &cobra.Command{
    Use:   "adapter",
    Short: "Adapter operations for a cluster",
}

var clusterAdapterPostStatusCmd = &cobra.Command{
    Use:          "post-status <adapter> <True|False|Unknown> [generation]",
    Short:        "Post adapter status conditions for the current cluster",
    Args:         cobra.RangeArgs(2, 3),
    SilenceUsage: true,
    RunE:         runClusterAdapterPostStatus,
}

var nodepoolAdapterCmd = &cobra.Command{
    Use:   "adapter",
    Short: "Adapter operations for a nodepool",
}

var nodepoolAdapterPostStatusCmd = &cobra.Command{
    Use:          "post-status <adapter> <True|False|Unknown> [generation] [nodepool_id]",
    Short:        "Post adapter status conditions for the current nodepool",
    Args:         cobra.RangeArgs(2, 4),
    SilenceUsage: true,
    RunE:         runNodePoolAdapterPostStatus,
}
```

Using `RangeArgs` keeps validation explicit and lets the `RunE` body handle optional args cleanly.

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| File placement | `cmd/adapter.go` (new file) | Keeps cluster.go clean; both cluster+nodepool adapter logic in one place |
| `buildAdapterPostPayload` visibility | package-private (`cmd`) | Only used within the cmd package; no need to export |
| Status validation | In `RunE` before any IO | Fail fast with clear error before hitting the network |
| Struct fields | Extend existing structs (additive) | Phase 2 types had incomplete fields vs. actual API; additive change is safe for existing callers |
| `ObservedTime` retention | Keep with `omitempty` | Zero backward-compat risk; no callers currently use it |
