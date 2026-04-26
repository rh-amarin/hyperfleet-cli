# Design: Phase 03 — Cluster Lifecycle

## API Base Path

All cluster endpoints live under:
```
{api-url}/api/hyperfleet/{api-version}/clusters[/{id}]
```

## ID Resolution

```
hf cluster <cmd> [cluster_id]
  explicit arg > ~/.config/hf/state.yaml cluster-id > error
```

Implemented via `config.ClusterID(store, arg)`.

## Shared Helper: api.FindClusterByName

Both `create` and `search` need to query clusters filtered by name and select exact matches that are not soft-deleted. This logic lives in `internal/api` as:

```go
func FindClusterByName(c *Client, ctx context.Context, name string) ([]resource.Cluster, error)
```

Calls `GET /clusters?search=name='<name>'` and filters client-side for exact name match + `deleted_at == null` (represented as empty `DeletedTime` field on the struct).

## Subcommand Design

### create
- Args: `[name] [region] [version]` — all optional, defaults applied if absent
- Defaults: `name=my-cluster`, `region=us-east-1`, `version=4.15.0`
- SilenceUsage on the subcommand so no-arg invocation proceeds silently
- Duplicate guard: call `FindClusterByName` first; warn + exit 0 if found
- After create: call `FindClusterByName` + `config.SetClusterID` to persist the ID

### search
- Args: `<name>` (required)
- Output: JSON array of matching clusters to stdout
- On single match: `config.SetClusterID`; on multiple: warn + set first; on zero: warn + output `[]`

### get
- Args: `[cluster_id]`
- Output: full cluster JSON

### list
- No args
- Calls `GET /clusters`
- Output: JSON (respects `--output` flag)

### patch
- Args: `spec|labels [cluster_id]` — first arg required
- Fetches current cluster, reads `<section>.counter`, increments, patches
- `[INFO]` line to stderr before patching

### delete
- Args: `[cluster_id]`
- Output: deleted cluster JSON (includes `deleted_by`, `deleted_time`)

### id
- No args
- Prints the configured cluster-id (from state.yaml) or errors

### conditions
- Args: `[cluster_id]`; flag: `-w` (watch, 2s interval)
- Output: `{"generation": N, "status": {"conditions": [...]}}`

### conditions table (subcommand `conditions table`)
- Args: `[cluster_id]`
- Output: tabwriter table with columns TYPE | STATUS | LAST TRANSITION | REASON | MESSAGE
- Status column: colored dot via `output.Printer.Dot()`

### statuses
- Args: `[cluster_id]`; flag: `-w` (watch, 2s interval)
- Calls `GET /clusters/{id}/adapter-statuses`
- Output: full `AdapterStatusList` JSON

## Watch Mode

A minimal inline watch loop (no separate package in this phase — that's phase-06):
- Clear terminal with `\033[H\033[2J`, call fetch+print fn, sleep 2s
- Loop until SIGINT via `os/signal`

## Output Defaults

| Command            | Default format |
|--------------------|---------------|
| create, get, patch | JSON          |
| list, search       | JSON (table in phase-06) |
| conditions         | JSON          |
| conditions table   | tabwriter     |
| statuses           | JSON          |

## Client Construction

All subcommands build the API client from `cfgStore`:
```go
cfg := cfgStore.Cfg()
client := api.NewClient(cfg.Hyperfleet.APIURL, cfg.Hyperfleet.APIVersion, cfg.Hyperfleet.Token, verbose)
```
