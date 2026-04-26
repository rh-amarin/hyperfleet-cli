# Proposal: Phase 05 — Adapter Status

## Intent

Implement `hf cluster adapter post-status` and `hf nodepool adapter post-status` — commands that simulate adapters reporting back to the HyperFleet API. These are the primary testing lever for exercising the convergence logic: once all required adapters report `Available=True` at the current generation, the cluster or nodepool `Ready` condition flips to `True`.

## Scope In

- `cmd/adapter.go` — new file containing both cluster and nodepool adapter subtrees
- `hf cluster adapter post-status <adapter> <True|False|Unknown> [generation]`
- `hf nodepool adapter post-status <adapter> <True|False|Unknown> [generation] [nodepool_id]`
- Shared `buildAdapterPostPayload` helper that constructs the 3-condition payload
- Struct updates to `internal/resource` to align with the verified API contract (add `last_transition_time` to `ConditionRequest`; add `created_time` and `last_report_time` to `AdapterStatusCreateRequest`)

## Scope Out

- Reading/querying adapter statuses (`hf cluster statuses` already done in phase 3)
- The dynamic dot-table combined view (phase 6)
- Any other adapter operations (listing, deleting)

## Background

The bash equivalents are:

- `scripts/hf.cluster.adapter.post.status.sh` → posts 3 conditions to `/clusters/{id}/adapter-statuses`  
  (note: the bash script uses `/clusters/{id}/statuses` — this is a bash-era naming mismatch; the correct Go endpoint is `/adapter-statuses` as confirmed by the verified output spec and the existing `hf cluster statuses` GET path)
- `scripts/hf.nodepool.adapter.post.status.sh` → posts 3 conditions to `/clusters/{cid}/nodepools/{nid}/adapter-statuses`

Required cluster adapters: `cl-deployment`, `cl-invalid-resource`, `cl-job`, `cl-maestro`, `cl-namespace`, `cl-precondition-error`  
Required nodepool adapter: `np-configmap`

## Dependency Note

This phase `dependsOn: phase-04-nodepool`. Phase 4 creates `cmd/nodepool.go` which defines `nodepoolCmd` — a package-level variable that `cmd/adapter.go` registers `nodepoolAdapterCmd` under. Phase 5 must not be implemented until phase 4 is archived.

## Resource Type Discrepancy (Delta Required)

The existing `resource.ConditionRequest` struct (defined in phase 2) is missing `last_transition_time`. The existing `resource.AdapterStatusCreateRequest` uses `observed_time` rather than `created_time` + `last_report_time`. The verified API execution logs in `scripts/output/04-adapter-status.json` show the API actually expects:

- `conditions[*].last_transition_time` — per condition
- `created_time` — on the root payload
- `last_report_time` — on the root payload

This phase extends those structs as a delta to the phase-2 spec.

## Testing Scope

| Package | Test cases |
|---|---|
| `cmd` (adapter_test.go) | POST cluster adapter-status: correct URL + payload shape |
| `cmd` (adapter_test.go) | POST nodepool adapter-status: correct URL + payload shape |
| `cmd` (adapter_test.go) | Invalid status value (`"maybe"`) → error returned, no HTTP request made |
| `cmd` (adapter_test.go) | Missing adapter arg → usage error |

All tests use `httptest.NewServer`. No mocks.

## Verification Steps Requiring Live Cluster Access

- 5.4 through 5.8 require a running HyperFleet API at the configured `api-url`.
- 5.7 requires an active nodepool-id to be set (depends on phase 4 being functional).
- 5.8 (conditions table after posting) requires at least `hf cluster conditions table` to work (phase 3 ✓).
