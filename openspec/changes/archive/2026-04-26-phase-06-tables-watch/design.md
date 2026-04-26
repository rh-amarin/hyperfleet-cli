# Design: Phase 06 — Tables & Watch Mode

## Command Tree (additions)

```
hf table                              # combined cluster + nodepool table
hf cluster table                      # cluster list as dot table
hf cluster conditions   [-w] [-i 2s]  # existing; adds --interval
hf cluster statuses     [-w] [-i 2s]  # existing; adds --interval
hf nodepool table       [-w] [-i 2s]  # existing; adds --watch/--interval
hf nodepool conditions  [-w] [-i 2s]  # existing; adds --interval
hf nodepool statuses    [-w] [-i 2s]  # existing; adds --interval
```

## New Package: `internal/watch`

### `watch.go`

```go
package watch

func Watch(interval time.Duration, fn func() error) error
```

Behavior:
1. Register SIGINT/SIGTERM via `signal.Notify`; defer `signal.Stop`
2. Loop: clear terminal (`\033[H\033[2J`), call `fn`, print footer to stderr, wait `interval` or quit signal
3. Returns `nil` on interrupt, propagates `fn` error

The existing `watchLoop` function in `cmd/cluster.go` is removed and replaced by calls to `watch.Watch`.

### Test Plan (`watch_test.go`)

| Test | Assertion |
|---|---|
| `TestWatch_CallsFnMultipleTimes` | fn called ≥2 times; use short interval (50ms); send SIGINT after 2nd call |
| `TestWatch_PropagatesFnError` | fn returns sentinel error on first call; Watch returns that error immediately |

## New File: `cmd/table.go`

### `hf cluster table` (`clusterTableCmd`)

- GET `/clusters` → `ListResponse[Cluster]`
- Columns: `NAME`, `GENERATION`, then `DynamicColumns` from cluster.Status.Conditions
- Dynamic column ordering: AVAILABLE first, alphabetical middle, READY last (via `output.DynamicColumns`)
- Supports `--watch`/`-w` and `--interval`/`-i` (default 2s)
- Registered: `clusterCmd.AddCommand(clusterTableCmd)`

### `hf table` (`tableCmd`)

- GET `/clusters` → all clusters
- For each cluster: GET `/clusters/{id}/nodepools` → nodepools
- Columns: `NAME`, `KIND`, `CLUSTER`, `GENERATION`, then `DynamicColumns` from all conditions combined
- Cluster rows: KIND=Cluster, CLUSTER=""
- NodePool rows: KIND=NodePool, CLUSTER=parent cluster name
- Supports `--watch`/`-w` and `--interval`/`-i` (default 2s)
- Registered: `rootCmd.AddCommand(tableCmd)`

## Modified Files

### `cmd/cluster.go`

- Remove `watchLoop` function (and its `os/signal`, `syscall` imports)
- Add import `"github.com/rh-amarin/hyperfleet-cli/internal/watch"`
- In `init()`: add `DurationP("interval", "i", 2*time.Second, ...)` flags to `clusterConditionsCmd` and `clusterStatusesCmd`
- In RunE of both commands: read interval, call `watch.Watch(interval, fetch)` instead of `watchLoop(fetch)`

### `cmd/nodepool.go`

- Add import `"github.com/rh-amarin/hyperfleet-cli/internal/watch"`
- In `init()`: add `DurationP("interval", "i", 2*time.Second, ...)` to `nodepoolConditionsCmd` and `nodepoolStatusesCmd`
- In `init()`: add `BoolP("watch", "w", ...)` and `DurationP("interval", "i", ...)` to `nodepoolTableCmd`
- In RunE of conditions and statuses: call `watch.Watch(interval, fetch)` instead of `watchLoop(fetch)`
- In RunE of `nodepoolTableCmd`: support watch/interval

## Dynamic Column Algorithm

Uses the existing `output.DynamicColumns([][]output.Condition)`:
1. Collect all condition types from all items
2. `AVAILABLE` first (if present)
3. Middle conditions sorted alphabetically
4. `READY` last (if present)

The `output.DynamicColumns` function already implements this; we just collect the input correctly.

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| `watchLoop` removal | Replace with `watch.Watch` | Eliminates duplication; `watchLoop` was hardcoded to 2s |
| `nodepoolTableCmd` location | Keep in `nodepool.go` | Already registered there; moving would require cleanup |
| `hf table` API calls | N+1 (clusters + per-cluster nodepools) | Simple; acceptable for the small cluster counts in HyperFleet |
| Combined table KIND column | "Cluster" / "NodePool" | Distinguishes rows in single flat table |
| Watch signal handling | SIGINT + SIGTERM | Matches phase 3 `watchLoop` behavior |

## Unit Test Plan (`cmd/table_test.go`)

| Test | Setup | Assertion |
|---|---|---|
| `TestClusterTable_RendersWithDynamicColumns` | 2 clusters with Available+Ready conditions | Headers contain NAME, GEN, AVAILABLE, READY; Available before Ready |
| `TestTable_RendersClusterAndNodePoolRows` | 1 cluster + 2 nodepools | Both cluster name and nodepool names appear in output; KIND col present |
| `TestTable_EmptyCluster_NoNodePoolRows` | 1 cluster, 0 nodepools | Cluster row present; no extra rows |
