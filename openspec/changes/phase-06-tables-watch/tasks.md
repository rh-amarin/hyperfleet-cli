# Tasks: Phase 06 — Tables & Watch Mode

## 1. internal/watch Package

- [x] 1.1 Create `internal/watch/watch.go` with `Watch(interval, fn)` — ANSI clear, timestamp footer, SIGINT/SIGTERM handling
- [x] 1.2 Create `internal/watch/watch_test.go` — `TestWatch_CallsFnMultipleTimes`, `TestWatch_PropagatesFnError`

## 2. Wire Watch into Existing Commands

- [x] 2.1 Update `cmd/cluster.go`: remove `watchLoop`; add `watch` import; add `--interval`/`-i` flags to `clusterConditionsCmd` and `clusterStatusesCmd`; replace `watchLoop(fetch)` with `watch.Watch(interval, fetch)`
- [x] 2.2 Update `cmd/nodepool.go`: add `watch` import; add `--interval`/`-i` flags to `nodepoolConditionsCmd` and `nodepoolStatusesCmd`; add `--watch`/`-w` and `--interval`/`-i` flags to `nodepoolTableCmd`; replace `watchLoop(fetch)` calls; add watch support to `nodepoolTableCmd`

## 3. cmd/table.go

- [x] 3.1 Create `cmd/table.go` with `clusterTableCmd` (`hf cluster table`) — cluster list with dynamic condition columns, `--watch`/`--interval` support
- [x] 3.2 Add `tableCmd` (`hf table`) — combined clusters + nodepools table, `--watch`/`--interval` support
- [x] 3.3 Register commands in `init()`: `clusterCmd.AddCommand(clusterTableCmd)`, `rootCmd.AddCommand(tableCmd)`

## 4. Unit Tests

- [x] 4.1 Create `cmd/table_test.go` — `TestClusterTable_RendersWithDynamicColumns`, `TestTable_RendersClusterAndNodePoolRows`, `TestTable_EmptyCluster_NoNodePoolRows`, `TestTable_AdapterConditionsAlphabetical`

## Verify

- [x] (a) `go build ./...` succeeds
- [x] (b) `go vet ./...` reports no issues
- [x] (c) `go test ./...` passes — captured to `verification_proof/tests.txt` (67 tests, 0 failures)
- [x] (d) Live cluster verification
  - [x] d.1 `hf cluster table` renders cluster table — saved to `verification_proof/d1-cluster-table.txt`
  - [x] d.2 `hf nodepool table` renders nodepool table — saved to `verification_proof/d2-nodepool-table.txt`
  - [x] d.3 `hf table` renders combined table — saved to `verification_proof/d3-table.txt`
  - [x] d.4 `hf cluster conditions -w --interval 2s` watch mode refreshes — saved to `verification_proof/d4-watch.txt`
