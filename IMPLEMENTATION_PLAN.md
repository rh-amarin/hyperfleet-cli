# HyperFleet CLI — Go Implementation Plan

## Overview

A Go reimplementation of ~50 bash scripts into a single `hf` binary. The tool manages HyperFleet clusters, node pools, databases, Kubernetes resources, and Maestro resources via a REST API. Each phase below is independently testable against the real cluster once you provide the kubeconfig.

**Target repo:** `rh-amarin/hyperfleet-cli`  
**Language:** Go 1.22+  
**Framework:** Cobra (command tree), Viper (config), pgx (postgres), client-go (k8s)

---

## Repository Layout (final target)

```
hf/
├── cmd/
│   ├── root.go         # global flags, plugin discovery
│   ├── cluster.go
│   ├── nodepool.go
│   ├── adapter.go
│   ├── config.go
│   ├── db.go
│   ├── maestro.go
│   ├── pubsub.go
│   ├── rabbitmq.go
│   ├── kube.go
│   ├── logs.go
│   ├── repos.go
│   ├── table.go
│   └── workflow.go
├── internal/
│   ├── api/            # HyperFleet HTTP client
│   ├── config/         # file-based config registry (~/.config/hf/)
│   ├── output/         # json/table/yaml + colored-dot renderer
│   ├── resource/       # Cluster, NodePool, Condition, AdapterStatus types
│   ├── watch/          # watch mode (-w flag)
│   ├── kube/           # client-go wrapper
│   ├── maestro/        # Maestro HTTP client
│   ├── pubsub/         # GCP Pub/Sub + RabbitMQ
│   ├── db/             # pgx wrapper
│   ├── plugin/         # hf-<name> discovery
│   └── version/
├── plugins/            # example plugin
├── main.go
├── go.mod
└── go.sum
```

---

## Phase 0 — Project Bootstrap

**Goal:** Compilable binary with `hf version` and `hf help` working. No config, no API calls.

### Tasks

1. `go mod init github.com/rh-amarin/hyperfleet-cli`
2. Add dependencies: `cobra`, `pgx/v5`, `k8s.io/client-go`, `cloud.google.com/go/pubsub`, `google/go-github/v60`, `gopkg.in/yaml.v3`
3. `main.go` → creates root Cobra command, calls `Execute()`
4. `cmd/root.go` — persistent flags: `--config`, `--output`, `--no-color`, `--verbose/-v`, `--api-url`, `--api-token`
5. `internal/version/version.go` — `Version`, `Commit`, `Date` injected at build time via `-ldflags`
6. `cmd/root.go` — `hf version` subcommand
7. `Makefile` — `build`, `install`, `lint`, `test` targets
8. `GoReleaser` config (`.goreleaser.yaml`) — linux/mac/windows × amd64/arm64

**Test:** `hf version` → prints version string. `hf --help` → full command tree visible.

---

## Phase 1 — Configuration System

**Goal:** `hf config show|set|clear|doctor|bootstrap|env` fully working. No API required.

### Internal: `internal/config`

- **Registry** — slice of `Entry{Section, Key, Default, Flags}` mirroring `HF_CONFIG_REGISTRY` in `hf.lib.sh`
- **Store** — reads/writes `~/.config/hf/<key>` files (one file per property)
- **Env profiles** — `<env>.<key>` files; `env activate` copies them over base files
- **Precedence** — flag > env var (`HF_*`) > env profile override > base config file > default
- **Secrets** — token, db-password, rabbitmq-password shown as `<set>`/`<not set>`

### Config registry properties (from `hf.lib.sh`)

```
Section        | Key                         | Default
hyperfleet     | api-url                     | http://localhost:8000
hyperfleet     | api-version                 | v1
hyperfleet     | token                       | (secret)
hyperfleet     | context                     |
hyperfleet     | namespace                   |
hyperfleet     | gcp-project                 | hcm-hyperfleet
hyperfleet     | cluster-id                  |
hyperfleet     | cluster-name                |
hyperfleet     | nodepool-id                 |
maestro        | maestro-consumer            | cluster1
maestro        | maestro-http-endpoint       | http://localhost:8100
maestro        | maestro-grpc-endpoint       | localhost:8090
maestro        | maestro-namespace           | maestro
portforward    | pf-api-port                 | 8000
portforward    | pf-pg-port                  | 5432
portforward    | pf-maestro-http-port        | 8100
portforward    | pf-maestro-http-remote-port | 8000
portforward    | pf-maestro-grpc-port        | 8090
database       | db-host                     | localhost
database       | db-port                     | 5432
database       | db-name                     |
database       | db-user                     |
database       | db-password                 | (secret)
rabbitmq       | rabbitmq-host               | localhost
rabbitmq       | rabbitmq-mgmt-port          | 15672
rabbitmq       | rabbitmq-user               | guest
rabbitmq       | rabbitmq-password           | (secret)
rabbitmq       | rabbitmq-vhost              | /
registry       | registry                    | $USER
```

### Commands

| Command | Behavior |
|---|---|
| `hf config` / `hf config show` | Print help header + env list + all sections |
| `hf config show <env>` | Same but highlight env-override properties in cyan with `[env]` suffix |
| `hf config set <key> <value>` | Write to `~/.config/hf/<key>` |
| `hf config clear <key>` | Delete `~/.config/hf/<key>` |
| `hf config clear all` | Delete all files in `~/.config/hf/` |
| `hf config doctor` | Scan all command requirements, show ●/○ per command |
| `hf config bootstrap [env]` | Interactive prompts for all properties in a section |
| `hf config env list` | List env profiles with property counts; active marked `●` |
| `hf config env show <name>` | Show full config with env overrides annotated |
| `hf config env activate <name>` | Copy `<env>.<key>` → `<key>` for each property |

**Test:** `hf config show` renders all sections. `hf config set api-url http://x && hf config show` reflects change. `hf config doctor` runs without error.

---

## Phase 2 — Shared API Client + Resource Types

**Goal:** All subsequent API-calling commands share the same HTTP client.

### Internal: `internal/api`

- `Client` struct — `BaseURL`, `Token`, `Timeout`, `Verbose`
- `NewClient(cfg *config.Store)` — builds from loaded config
- Methods: `Get[T]`, `Post[T]`, `Patch[T]`, `Delete` — generic typed round-trips
- Base path: `{api-url}/api/hyperfleet/{api-version}/`
- Auth: `Authorization: Bearer <token>` if token set
- RFC 7807 error parsing → structured `APIError{Code, Detail, Status, Title, TraceID, Type, Timestamp}`
- `APIError` implements `error`; commands choose json/text rendering
- `--verbose` logs: method, URL, status code, duration to stderr

### Internal: `internal/resource`

```go
type Cluster struct {
    ID          string         `json:"id"`
    Kind        string         `json:"kind"`
    Name        string         `json:"name"`
    Generation  int64          `json:"generation"`
    Labels      map[string]any `json:"labels"`
    Spec        map[string]any `json:"spec"`
    Status      ClusterStatus  `json:"status"`
    CreatedBy   string         `json:"created_by"`
    CreatedTime time.Time      `json:"created_time"`
    UpdatedBy   string         `json:"updated_by"`
    UpdatedTime time.Time      `json:"updated_time"`
    DeletedBy   string         `json:"deleted_by,omitempty"`
    DeletedTime *time.Time     `json:"deleted_time,omitempty"`
    Href        string         `json:"href"`
}

type NodePool struct { /* same shape, plus OwnerReferences */ }

type Condition struct {
    Type               string    `json:"type"`
    Status             string    `json:"status"`
    Reason             string    `json:"reason"`
    Message            string    `json:"message"`
    LastTransitionTime time.Time `json:"last_transition_time"`
    ObservedGeneration int64     `json:"observed_generation"`
}

type AdapterStatus struct {
    Adapter            string      `json:"adapter"`
    Conditions         []Condition `json:"conditions"`
    ObservedGeneration int64       `json:"observed_generation"`
    LastReportTime     time.Time   `json:"last_report_time"`
    Data               any         `json:"data"`
    CreatedTime        time.Time   `json:"created_time"`
}

type ListResponse[T any] struct {
    Items []T    `json:"items"`
    Kind  string `json:"kind"`
    Page  int    `json:"page"`
    Size  int    `json:"size"`
    Total int    `json:"total"`
}
```

### Internal: `internal/output`

- `Printer` — dispatches to json/table/yaml based on `--output` flag
- **JSON** — `encoding/json` with `json.MarshalIndent`
- **YAML** — `gopkg.in/yaml.v3`
- **Table** — `text/tabwriter` with aligned columns
- **Colored dot renderer** for conditions:
  - `True` → green `●`
  - `False` → red `●`
  - `Unknown` → yellow `●`
  - absent → `-`
  - Respects `--no-color`
- Dynamic column ordering: fixed columns → `Available` → alphabetical adapter conditions → `Ready` last

**Test:** Build succeeds. Unit test the API error parser against RFC 7807 JSON.

---

## Phase 3 — Cluster Lifecycle

**Goal:** Full cluster CRUD testable against the real API.

### Commands

| Command | API Call | Notes |
|---|---|---|
| `hf cluster create <name> [region] [version]` | `POST /clusters` | Check duplicate first; auto-set cluster-id on success |
| `hf cluster get [cluster_id]` | `GET /clusters/{id}` | Uses config cluster-id if not given |
| `hf cluster list` | `GET /clusters` | Returns `ClusterList`; default table output |
| `hf cluster search <name>` | `GET /clusters?search=name` | Filters deleted; sets cluster-id in config |
| `hf cluster patch spec\|labels [cluster_id]` | `PATCH /clusters/{id}` | Reads JSON patch from stdin or flag |
| `hf cluster delete [cluster_id]` | `DELETE /clusters/{id}` | Confirms with cluster-id from config |
| `hf cluster id` | (local) | Reads `~/.config/hf/cluster-id` |
| `hf cluster conditions [-w] [cluster_id]` | `GET /clusters/{id}` | Shows conditions; `-w` for watch mode |
| `hf cluster conditions-table [cluster_id]` | `GET /clusters/{id}/adapter-statuses` | Dynamic dot-table |
| `hf cluster statuses [-w] [cluster_id]` | `GET /clusters/{id}/adapter-statuses` | List all adapter statuses |
| `hf cluster table` | `GET /clusters` + statuses | Dynamic column table (see Phase 5) |

**Create payload:**
```json
{
  "name": "<name>",
  "labels": {"counter": "1", "environment": "development", "shard": "1", "team": "core"},
  "spec": {"counter": "1", "region": "<region>", "version": "<version>"}
}
```
Defaults: name=`my-cluster`, region=`us-east-1`, version=`4.15.0`. **No args → creates with defaults (no usage text).**

**Test:** Create cluster, get it, search by name, patch labels, view conditions, delete.

---

## Phase 4 — NodePool Lifecycle

**Goal:** Full nodepool CRUD, always scoped to current cluster-id.

### Commands

| Command | API Call | Notes |
|---|---|---|
| `hf nodepool create <name> [count] [instance-type]` | `POST /clusters/{cid}/nodepools` (×count) | Creates `count` pools with `-1`,`-2` suffix; sets nodepool-id to last |
| `hf nodepool get [nodepool_id]` | `GET /clusters/{cid}/nodepools/{id}` | |
| `hf nodepool list [cluster_id]` | `GET /clusters/{cid}/nodepools` | |
| `hf nodepool search <name>` | `GET /clusters/{cid}/nodepools` (filter) | Sets nodepool-id |
| `hf nodepool patch spec\|labels [nodepool_id]` | `PATCH /clusters/{cid}/nodepools/{id}` | |
| `hf nodepool delete [nodepool_id]` | `DELETE /clusters/{cid}/nodepools/{id}` | |
| `hf nodepool id` | (local) | Reads `~/.config/hf/nodepool-id` |
| `hf nodepool conditions [-w] [nodepool_id]` | `GET /clusters/{cid}/nodepools/{id}` | |
| `hf nodepool conditions-table [nodepool_id]` | adapter-statuses | Dynamic dot-table |
| `hf nodepool statuses [-w] [nodepool_id]` | adapter-statuses | |
| `hf nodepool table` | all nodepools + statuses | Dynamic dot-table |

**Create payload per pool:**
```json
{
  "name": "<name>-<N>",
  "labels": {"counter": "<N>"},
  "spec": {"counter": "<N>", "platform": {"type": "<instance-type>"}, "replicas": 1}
}
```
Defaults: name=`my-nodepool`, count=`1`, instance-type=`m4`.

**Test:** Create 2 nodepools, list, search, patch, view statuses, delete.

---

## Phase 5 — Adapter Status

**Goal:** `hf cluster adapter post-status` and `hf nodepool adapter post-status`.

### Commands

| Command | API Call |
|---|---|
| `hf cluster adapter post-status <adapter> <True\|False\|Unknown> [generation]` | `POST /clusters/{cid}/adapter-statuses` |
| `hf nodepool adapter post-status <adapter> <True\|False\|Unknown> [generation] [nodepool_id]` | `POST /clusters/{cid}/nodepools/{nid}/adapter-statuses` |

**Payload:** Posts 3 conditions (`Available`, `Applied`, `Health`) all set to the given status value.

```json
{
  "adapter": "<adapter>",
  "conditions": [
    {"type": "Available", "status": "<status>", "reason": "ManualStatusPost", "message": "...", "last_transition_time": "<now>"},
    {"type": "Applied",   "status": "<status>", "reason": "ManualStatusPost", "message": "...", "last_transition_time": "<now>"},
    {"type": "Health",    "status": "<status>", "reason": "ManualStatusPost", "message": "...", "last_transition_time": "<now>"}
  ],
  "observed_generation": <generation>,
  "data": {},
  "created_time": "<now>",
  "last_report_time": "<now>"
}
```

**Test:** Post True/False/Unknown statuses; verify conditions-table reflects the colored dots correctly.

---

## Phase 6 — Tables & Combined View

**Goal:** `hf table` — combined cluster+nodepool dynamic dot-table.

### `internal/output` dynamic table

Column ordering algorithm:
1. Fixed: `NAME`, `GENERATION`
2. `AVAILABLE` 
3. All adapter-specific condition types, alphabetically
4. `READY` last

One row per cluster (or nodepool), one column per unique condition type across all resources. Missing conditions render as `-`.

### Commands

- `hf table` — shows clusters at top, nodepools below (grouped by cluster)
- `hf cluster table` — clusters only
- `hf nodepool table` — nodepools for current cluster

**Test:** Create cluster + 2 nodepools + post statuses → `hf table` shows correct colored dots.

---

## Phase 7 — Watch Mode

**Goal:** `-w` flag on `conditions` and `statuses` commands.

### `internal/watch`

- `Watch(interval time.Duration, fn func() error)` — clears terminal on each tick, calls `fn`, shows "Last updated: <timestamp>" footer
- Default interval: 2 seconds
- ANSI clear: `\033[H\033[2J`
- Stops on Ctrl+C (SIGINT)

**Test:** `hf cluster conditions -w` refreshes every 2s; Ctrl+C exits cleanly.

---

## Phase 8 — Database Commands

**Goal:** Direct PostgreSQL via `jackc/pgx/v5` — no `psql` needed.

### `internal/db`

- `Client` — wraps `pgxpool.Pool`
- `NewClient(cfg *config.Store)` — connects using db-host/port/name/user/password
- `Query(sql string) ([][]string, []string, error)` — returns headers + rows
- `Exec(sql string) error`

### Commands

| Command | Behavior |
|---|---|
| `hf db query <sql>` or `hf db query -f <file>` | Run SQL, display as table |
| `hf db delete <table> [id]` | `DELETE FROM <table> WHERE id='<id>'` or all rows |
| `hf db delete-all` | Truncate all HyperFleet tables |
| `hf db statuses` | `SELECT * FROM adapter_statuses` (formatted) |
| `hf db statuses-delete` | Delete all rows from adapter_statuses |
| `hf db config` | Show current DB connection info |

**Test:** `hf db query "SELECT 1"`, `hf db statuses`, `hf db config`.

---

## Phase 9 — Maestro

**Goal:** Maestro resource management via HTTP API; TUI via shell-out.

### `internal/maestro`

- `Client` — `net/http` calls to `maestro-http-endpoint`
- Endpoints under `/api/maestro/v1/`
- Consumer scoped by `maestro-consumer` config

### Commands

| Command | API |
|---|---|
| `hf maestro list` | `GET /resourcebundles` (consumer filter) |
| `hf maestro get [name]` | `GET /resourcebundles/{name}` |
| `hf maestro delete [name]` | `DELETE /resourcebundles/{name}` |
| `hf maestro bundles` | `GET /resourcebundles` (all) |
| `hf maestro consumers` | `GET /consumers` |
| `hf maestro tui` | Shell out to `maestro-cli`; error if not found |

**Test:** `hf maestro list`, `hf maestro consumers`. `hf maestro tui` → clear error if binary missing.

---

## Phase 10 — Pub/Sub & RabbitMQ

**Goal:** Publish events to GCP Pub/Sub and RabbitMQ.

### `internal/pubsub`

- GCP: `cloud.google.com/go/pubsub` — `PublishClusterChange`, `PublishNodePoolChange`
- RabbitMQ: `github.com/rabbitmq/amqp091-go` — `PublishClusterChange`
- Message payload: CloudEvent JSON `{specversion, type, source, id, data}`

### Commands

| Command | Behavior |
|---|---|
| `hf pubsub list [filter]` | List GCP Pub/Sub topics in gcp-project |
| `hf pubsub publish cluster <topic>` | Publish cluster CloudEvent to topic |
| `hf pubsub publish nodepool <topic>` | Publish nodepool CloudEvent to topic |
| `hf rabbitmq publish cluster <exchange> [routing-key]` | Publish to RabbitMQ exchange |

**Test (requires GCP credentials):** `hf pubsub list`. RabbitMQ: `hf rabbitmq publish cluster hyperfleet`.

---

## Phase 11 — Kubernetes

**Goal:** Port-forward lifecycle, log streaming, debug pod — no `kubectl` needed.

### `internal/kube`

- `NewClient(ctx, kubeconfig, context)` — loads kubeconfig from flag → `KUBECONFIG` → `~/.kube/config`
- **Port-forward**: start/stop/status with PID tracking in `~/.config/hf/pf-*.pid`
  - Services: API (pf-api-port), Postgres (pf-pg-port), Maestro HTTP/gRPC
- **Log streaming**: label or name filter, multi-pod fan-out, replaces `stern`
- **Pod exec**: in-cluster curl via exec into debug pod
- **Debug pod**: ephemeral container or standalone debug pod

### Commands

| Command | Behavior |
|---|---|
| `hf kube port-forward start` | Start all port-forwards in background |
| `hf kube port-forward stop` | Kill all forwarded PIDs |
| `hf kube port-forward status` | Show running forwards with ports |
| `hf kube context` | Show/set current kubeconfig context |
| `hf kube curl [opts] <url>` | Exec curl inside a pod |
| `hf kube debug <deployment> [ns]` | Attach debug container |
| `hf logs <pattern> [flags]` | Stream logs matching pod name pattern |
| `hf logs adapter <pattern> [flags]` | Stream logs from adapter pods |

**Test (requires kubeconfig):** `hf kube context`, `hf kube port-forward start`, `hf logs hyperfleet`.

---

## Phase 12 — Repos

**Goal:** GitHub repository status overview.

### `internal/repos`

- `google/go-github/v60` client, auth from `GITHUB_TOKEN` env or config token
- `REGISTRY` config property controls the GitHub user/org to query

### Commands

| Command | Behavior |
|---|---|
| `hf repos` | List repos for registry owner with status (open PRs, CI status, branch) |

**Test:** `hf repos` → table of repos for configured registry user.

---

## Phase 13 — Interactive Commands

**Goal:** `hf workflow run` and `hf workflow api-only` — end-to-end scripted sequences.

### Commands

- `hf workflow run` — full workflow: config → port-forward → create cluster → nodepools → post statuses → verify ready
- `hf workflow api-only` — same but skips k8s/port-forward steps

These call the internal packages directly (not subprocess calls to other `hf` subcommands).

**Test:** `hf workflow api-only` against real cluster → creates resources and verifies convergence.

---

## Phase 14 — Non-Functional / Polish

1. **Shell completions** — `hf completion bash|zsh|fish|powershell` (Cobra built-in)
2. **Plugin architecture** — `internal/plugin`: scans `$PATH` for `hf-*` executables; delegates unknown commands
3. **GoReleaser** — `.goreleaser.yaml` with linux/mac/windows × amd64/arm64; SHA256 checksums
4. **Error & Usage polish** — ensure RFC 7807 errors render per `--output` format; `[ERROR]`/`[WARN]`/`[INFO]` on stderr
5. **Config doctor** — scan commands' required-config declarations; show readiness per command
6. **Config bootstrap** — interactive `hf config bootstrap` with prompts
7. **Unit tests** — `internal/output`, `internal/config`, `internal/api` error parsing
8. **Integration tests** — tagged `//go:build integration`, run against real cluster

---

## Iteration Order Summary

| Phase | Deliverable | Can Test Without Cluster? |
|---|---|---|
| 0 | Binary boots, `hf version` | ✅ Yes |
| 1 | `hf config` fully working | ✅ Yes |
| 2 | Shared API client + types | ✅ Yes (unit tests) |
| 3 | `hf cluster` CRUD | 🔶 Needs API |
| 4 | `hf nodepool` CRUD | 🔶 Needs API |
| 5 | `hf adapter post-status` | 🔶 Needs API |
| 6 | `hf table` dynamic dot-table | 🔶 Needs API |
| 7 | Watch mode (`-w`) | 🔶 Needs API |
| 8 | `hf db` PostgreSQL | 🔶 Needs DB |
| 9 | `hf maestro` | 🔶 Needs Maestro |
| 10 | `hf pubsub` / `hf rabbitmq` | 🔶 Needs GCP/RabbitMQ |
| 11 | `hf kube` / `hf logs` | 🔶 Needs kubeconfig |
| 12 | `hf repos` | 🔶 Needs GitHub token |
| 13 | `hf workflow` | 🔶 Needs full stack |
| 14 | Completions, plugins, GoReleaser | ✅ Yes |

---

## Key Design Decisions

1. **Config storage** — File-per-property at `~/.config/hf/<key>` (matches shell scripts exactly; enables `env activate` to copy files)
2. **No default "show usage" on create** — `hf cluster create` with no args creates with defaults (matches current script behavior)
3. **cluster-id/nodepool-id side effects** — `create` and `search` write the ID to config automatically
4. **Exit codes** — API errors exit 0 (backward compat with scripts); CLI-level errors (missing config, network down) exit non-zero
5. **Output default** — list commands default to `table`; get/create/patch default to `json`
6. **Maestro TUI** — only `hf maestro tui` shells out; all other maestro commands use HTTP API directly
7. **Port-forward PIDs** — tracked in `~/.config/hf/pf-<service>.pid` for start/stop/status management
