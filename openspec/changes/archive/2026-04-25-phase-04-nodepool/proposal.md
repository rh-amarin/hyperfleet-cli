# Proposal: Phase 04 — NodePool Lifecycle

## Summary

Implement all `hf nodepool` subcommands for full CRUD lifecycle management of HyperFleet nodepools. NodePools are always scoped to a parent cluster via `cluster-id` in state. Mirrors the cluster lifecycle phase with two additions: `create` creates N nodepools in a loop, and `table` renders a dynamic condition-column table.

## Commands

| Command | API | Notes |
|---|---|---|
| `hf nodepool create [name] [count] [instance-type]` | POST ×N | Defaults: my-nodepool, 1, m4. Names: `<name>-1..N`. Persists last ID. |
| `hf nodepool list [cluster_id]` | GET | NodePoolList response |
| `hf nodepool search <name>` | GET+filter | Exact name, non-deleted. Sets nodepool-id. |
| `hf nodepool get [nodepool_id]` | GET | Full NodePool JSON |
| `hf nodepool patch spec\|labels [nodepool_id]` | GET+PATCH | Increments counter |
| `hf nodepool delete [nodepool_id]` | DELETE | Returns deleted NodePool |
| `hf nodepool id` | — | Prints configured nodepool-id |
| `hf nodepool conditions [-w] [nodepool_id]` | GET | `{generation, status.conditions}` JSON |
| `hf nodepool conditions table [nodepool_id]` | GET | tabwriter table: TYPE STATUS LAST-TRANSITION REASON MESSAGE |
| `hf nodepool statuses [-w] [nodepool_id]` | GET | AdapterStatusList; 404 → empty list |
| `hf nodepool table [cluster_id]` | GET | Table: ID NAME REPLICAS TYPE GEN + dynamic condition columns |

## Testing Scope

- `internal/api/nodepool_test.go` — `FindNodePoolByName`: exact match, filters deleted, filters partial, no match, multiple active, URL encoding (includes cluster-id in path), API error
- `cmd/nodepool_test.go` — command tests: create (defaults, multi-nodepool loop, persists last ID), list, search (sets ID, no match), get (JSON output), patch (spec counter, labels counter, no args), delete, conditions (JSON, table), statuses (empty, 404→empty), table (renders columns)

## Dependencies

- `config.NodePoolID` / `config.SetNodePoolID` — already implemented
- `resource.NodePool`, `resource.NodePoolStatus`, `resource.ObjectReference` — already implemented
- `output.DynamicColumns` — already implemented; used by `nodepool table`
- `watchLoop` helper in `cmd/cluster.go` — reuse pattern, define in nodepool.go locally or extract to shared file

## Risks

- `nodepool table` requires extracting `spec.replicas` (int) and `spec.platform.type` (nested map) — type assertions needed
- `create` loop: each POST is independent; if one fails mid-loop the earlier nodepools are already created
