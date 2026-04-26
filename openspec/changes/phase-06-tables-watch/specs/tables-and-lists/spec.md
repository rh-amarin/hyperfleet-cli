# Delta Spec: Tables & Lists

## Source

This delta spec documents corrections and confirmations made during live verification of phase-06-tables-watch against the HyperFleet API at `34.175.27.173:8000`.

## Confirmed Facts

### API Behavior
- The HyperFleet API at `34.175.27.173:8000` is publicly accessible without authentication during development.
- Clusters respond with conditions including adapter-specific types (e.g. `ClDeploymentSuccessful`, `ClJobSuccessful`, `ClMaestroSuccessful`, `ClNamespaceSuccessful`) in addition to the standard `Ready` and `Available`.

### Dynamic Column Ordering (confirmed correct)
The `output.DynamicColumns` algorithm correctly orders:
1. `Available` first (if present)
2. Any other condition types alphabetically (e.g. `ClDeploymentSuccessful`, `ClJobSuccessful`, ...)
3. `Ready` last (if present)

Live output from `hf cluster table` confirms:
```
NAME        GEN  AVAILABLE  CLDEPLOYMENTSUCCESSFUL  CLJOBSUCCESSFUL  CLMAESTROSUCCESSFUL  CLNAMESPACESUCCESSFUL  READY
my-cluster  1    True       True                    True             True                 True                   True
```

### Combined Table (`hf table`)
Columns: `NAME`, `KIND`, `CLUSTER`, `GEN`, then dynamic condition columns.
- Cluster rows have empty `CLUSTER` field.
- NodePool rows (when present) show parent cluster name in `CLUSTER`.
- Both clusters and nodepools share the same dynamic column set (union of all condition types).

### Watch Mode
- `--interval` / `-i` flag (default `2s`) is available on all watch-capable commands.
- `watch.Watch` clears terminal with ANSI escape `\033[H\033[2J`, prints footer `Last updated: HH:MM:SS  (Ctrl+C to stop)` to stderr.
- `timeout` or SIGTERM triggers clean exit (return nil).

## Corrections Found During Verification

None. Implementation matched the spec from `IMPLEMENTATION_PLAN.md` and phase-05 design patterns.

## Commands Added

| Command | Description |
|---|---|
| `hf cluster table` | Cluster list as dot table with dynamic condition columns |
| `hf table` | Combined clusters + nodepools dot table |
| `hf cluster conditions -w -i <dur>` | Watch cluster conditions with configurable interval |
| `hf cluster statuses -w -i <dur>` | Watch cluster adapter statuses with configurable interval |
| `hf nodepool conditions -w -i <dur>` | Watch nodepool conditions with configurable interval |
| `hf nodepool statuses -w -i <dur>` | Watch nodepool adapter statuses with configurable interval |
| `hf nodepool table -w -i <dur>` | Nodepool table with watch support |
