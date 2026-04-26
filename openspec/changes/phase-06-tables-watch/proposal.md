# Proposal: Phase 06 — Tables & Watch Mode

## Intent

Implement the dynamic dot-table combined view (`hf table`, `hf cluster table`, `hf nodepool table`) and wire configurable watch mode (`-w`/`--watch` + `-i`/`--interval`) into all listing commands. Extract watch loop logic into a reusable `internal/watch` package.

## Scope In

- `internal/watch` package: `Watch(interval, fn)` with ANSI terminal clear, timestamp footer, SIGINT handling
- `cmd/table.go`: `hf table` (combined clusters + nodepools), `hf cluster table` (clusters with condition columns)
- Wire `--interval` flag into existing `hf cluster conditions`, `hf cluster statuses`, `hf nodepool conditions`, `hf nodepool statuses` (already have `-w` flag; add `-i`)
- Wire `--watch` and `--interval` flags into the existing `hf nodepool table` command (in `nodepool.go`)
- Replace the hardcoded 2s `watchLoop` function in `cmd/cluster.go` with calls to `watch.Watch`

## Scope Out

- No changes to `go.mod` / `go.sum` / `cmd/root.go` / `internal/api/methods.go`
- No changes to `internal/output` existing files
- No database, maestro, pubsub, kubernetes, or repos commands

## Testing Scope

### `internal/watch` (new package — `watch_test.go`)
- `TestWatch_CallsFnMultipleTimes`: Watch with 50ms interval; assert fn called ≥2 times before SIGINT
- `TestWatch_ReturnsNilOnInterrupt`: after fn called twice, send SIGINT; assert return value is nil
- `TestWatch_PropagatesFnError`: fn returns an error on first call; assert Watch returns that error

### `cmd` (existing test package — `table_test.go`)
- `TestClusterTable_RendersWithDynamicColumns`: httptest serving cluster list; assert headers NAME, GENERATION, AVAILABLE, READY; assert Available before Ready
- `TestTable_RendersClusterAndNodePoolRows`: httptest serving cluster list + nodepool list; assert rows for both cluster and nodepool names; assert KIND column
- `TestTable_EmptyCluster_NoNodePoolRows`: cluster list with one cluster, nodepool list empty; assert no nodepool rows

## Which Steps Require Live Cluster

Verification steps (d.1)–(d.4) require the live HyperFleet API at `34.175.27.173:8000`.
