# Proposal: Fix adapter-status endpoint path and missing Finalized condition

## What

Two bugs in the adapter status commands:

1. **Wrong GET endpoint**: `hf cluster statuses` and `hf nodepool statuses` call
   `/adapter-statuses` which returns 404 on the live API. The correct path is `/statuses`.

2. **Missing Finalized condition**: `hf cluster adapter post-status` and
   `hf nodepool adapter post-status` send only 3 conditions (Available, Applied, Health).
   The spec and the API require 4 — `Finalized` must be included.

## Why

Both issues were already documented in `openspec/specs/adapter-status/spec.md` under
"API Endpoints" (note: no `/adapter-statuses` endpoint exists) and the 4-condition
request payload example. The implementation drifted from the spec.

## Scope

- `cmd/cluster.go` — fix GET path
- `cmd/nodepool.go` — fix GET path
- `cmd/adapter.go` — add Finalized condition to `buildAdapterPostPayload`
- `cmd/adapter_test.go` — update condition-count assertions and type checklist
