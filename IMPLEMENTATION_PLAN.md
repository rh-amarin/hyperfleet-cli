# HyperFleet CLI — Go Implementation Plan (OpenSpec Method)

## Overview

This plan governs the Go reimplementation of the HyperFleet bash script suite as a self-contained `hf` binary. It follows the **OpenSpec** spec-driven development methodology: every chunk of implementation work lives in its own change folder under `openspec/changes/`, with a proposal, design, tasks, and delta specs. Changes are implemented one at a time, verified, and archived into `openspec/specs/` when complete.

**Target repo:** `rh-amarin/hyperfleet-cli`  
**Language:** Go 1.22+  
**Framework:** Cobra (commands), pgx (postgres), client-go (k8s), Google Cloud Go SDK (pubsub)

---

## OpenSpec Directory Layout

The repo already has `openspec/specs/` as the source of truth (the 13 functional + 3 technical spec domains). All implementation work flows through `openspec/changes/`.

```
openspec/
├── specs/                         ← source of truth (already exists)
│   ├── technical-architecture/
│   ├── config-model/
│   ├── config/
│   ├── cluster-lifecycle/
│   ├── nodepool-lifecycle/
│   ├── adapter-status/
│   ├── tables-and-lists/
│   ├── database/
│   ├── maestro/
│   ├── pubsub/
│   ├── kubernetes/
│   ├── repos/
│   ├── errors-and-usage/
│   ├── interactive-commands/
│   ├── config-registry/
│   └── non-functional/
└── changes/                       ← in-progress implementation work
    ├── IMPLEMENTATION_ORDER.md    ← dependency narrative (this file)
    ├── phase-00-bootstrap/
    ├── phase-01-config/
    ├── phase-02-api-client/
    ├── phase-03-cluster/
    ├── phase-04-nodepool/
    ├── phase-05-adapter-status/
    ├── phase-06-tables-watch/
    ├── phase-07-database/
    ├── phase-08-maestro/
    ├── phase-09-pubsub/
    ├── phase-10-kubernetes/
    ├── phase-11-repos/
    ├── phase-12-workflow/
    └── phase-13-nfr/
```

---

## How to Work with This Plan (OpenSpec OPSX Workflow)

Each implementation session follows this loop:

```
/opsx:propose <change-name>    ← create or resume a change folder
/opsx:apply                    ← implement tasks, checking them off
/opsx:verify                   ← validate implementation against specs
/opsx:archive                  ← merge delta specs → openspec/specs/, move to archive/
```

**Rules:**
- Never implement code outside a change folder's scope.
- Check tasks off in `tasks.md` as you go — do not batch.
- If you discover the design needs to change, update `design.md` first, then code.
- Only archive when all tasks are checked and `/opsx:verify` passes.
- Archive one change before starting the next in the same dependency chain.

---

## Change Dependency Graph

Dependencies flow top to bottom. A change marked `dependsOn` another must not be archived first.

```
phase-00-bootstrap
        │
        ▼
phase-01-config ──────────────────────────────────────────────┐
        │                                                      │
        ▼                                                      │
phase-02-api-client                                            │
        │                                                      │
        ├──────────────────┐                                   │
        ▼                  ▼                                   │
phase-03-cluster    phase-07-database (needs phase-01 too)     │
        │                                                      │
        ▼                                                      │
phase-04-nodepool                                              │
        │                                                      │
        ├────────────────────────────────┐                     │
        ▼                               ▼                      │
phase-05-adapter-status         phase-08-maestro               │
        │                                                      │
        ▼                                                      │
phase-06-tables-watch                                          │
        │                                                      │
        │       ┌──────────────────────────────────────────────┘
        │       ▼
        │   phase-10-kubernetes (needs phase-01 too)
        │
        ├── phase-09-pubsub      (needs phase-03, phase-04)
        ├── phase-11-repos       (needs phase-00)
        ├── phase-12-workflow    (needs phase-03 through phase-06)
        └── phase-13-nfr        (needs all others)
```

---

## Change Folders

Each change folder follows the OpenSpec artifact structure:

```
openspec/changes/<phase-name>/
├── .openspec.yaml    ← schema, created date, dependsOn
├── proposal.md       ← why + scope
├── design.md         ← technical approach, package layout, key decisions
├── tasks.md          ← numbered checklist, checked off during implementation
└── specs/            ← delta specs (ADDED/MODIFIED/REMOVED against openspec/specs/)
    └── <domain>/
        └── spec.md
```

---

## Phase 0 — Bootstrap

**Change folder:** `openspec/changes/phase-00-bootstrap/`  
**Testable without cluster:** Yes  
**dependsOn:** none

### proposal.md

**Intent:** Stand up the Go module, Cobra root command, and build tooling so the project compiles and `hf version` works.  
**Scope in:** go module init, main.go, root command, version command, Makefile, GoReleaser config.  
**Scope out:** any actual commands beyond `version` and `help`.

### design.md

**Module path:** `github.com/rh-amarin/hyperfleet-cli`  
**Entry point:** `main.go` → `cmd.Execute()`  
**Root command** (`cmd/root.go`): persistent flags `--config`, `--output`, `--no-color`, `--verbose/-v`, `--api-url`, `--api-token`  
**Version** (`internal/version/version.go`): `Version`, `Commit`, `Date` injected via `-ldflags` at build time  
**Makefile targets:** `build`, `install`, `lint`, `test`, `release-dry-run`  
**GoReleaser** (`.goreleaser.yaml`): linux/mac/windows × amd64/arm64, SHA256 checksums

### tasks.md

```
## 1. Module Setup
- [ ] 1.1 go mod init github.com/rh-amarin/hyperfleet-cli
- [ ] 1.2 Add cobra, yaml, pgx, client-go, pubsub, go-github to go.mod
- [ ] 1.3 Create main.go

## 2. Root Command
- [ ] 2.1 cmd/root.go with all persistent flags
- [ ] 2.2 hf version subcommand (reads internal/version)

## 3. Build Tooling
- [ ] 3.1 Makefile with build/install/lint/test targets
- [ ] 3.2 .goreleaser.yaml for cross-compilation

## 4. Verify
- [ ] 4.1 go build ./... succeeds
- [ ] 4.2 hf version prints version string
- [ ] 4.3 hf --help shows command tree stub
```

---

## Phase 1 — Configuration System

**Change folder:** `openspec/changes/phase-01-config/`  
**Testable without cluster:** Yes  
**dependsOn:** `phase-00-bootstrap`

### proposal.md

**Intent:** Implement `hf config` and all subcommands using file-per-property storage at `~/.config/hf/<key>`, exactly mirroring the bash `hf.config.sh` behavior.  
**Scope in:** `internal/config` package, `cmd/config.go`, all 30 config properties, env profiles, config registry.  
**Scope out:** API calls, Kubernetes, database.

### design.md

**Storage model:** One file per property at `~/.config/hf/<key>`. Env profiles stored as `~/.config/hf/<env>.<key>`. Active env pointer stored in `~/.config/hf/.active-env`.

**Precedence chain:** CLI flag → `HF_*` env var → env profile override → base config file → default value

**Config registry** — static slice of `Entry{Section, Key, Default, IsSecret}` covering all 30 properties (hyperfleet, maestro, portforward, database, rabbitmq, registry sections).

**Key types:**
```go
// internal/config
type Entry struct { Section, Key, Default string; IsSecret bool }
type Store struct { dir string; activeEnv string }
func (s *Store) Get(key string) string
func (s *Store) Set(key, value string) error
func (s *Store) Clear(key string) error
func (s *Store) ActiveEnv() string
func (s *Store) EnvList() []EnvProfile
func (s *Store) EnvActivate(name string) error
```

**Output:** Sections formatted as aligned two-column key/value text. Secrets rendered as `<set>` or `<not set>`. Active env marked with `●` (green), inactive with `○`.

### tasks.md

```
## 1. internal/config package
- [ ] 1.1 Define Registry (all 30 properties with section/key/default/isSecret)
- [ ] 1.2 Implement Store.Get (flag > env var > env profile > file > default)
- [ ] 1.3 Implement Store.Set (write ~/.config/hf/<key>)
- [ ] 1.4 Implement Store.Clear (delete file or all files)
- [ ] 1.5 Implement env profile management (list, show, activate)
- [ ] 1.6 Implement hf_cluster_id / hf_nodepool_id helpers (read/write ID files)

## 2. cmd/config.go
- [ ] 2.1 hf config (no args) → help + env list + full config
- [ ] 2.2 hf config show [env] → grouped sections, env overrides in cyan
- [ ] 2.3 hf config set <key> <value>
- [ ] 2.4 hf config clear <key> / clear all
- [ ] 2.5 hf config doctor (check required config per command)
- [ ] 2.6 hf config bootstrap [env] (interactive prompts)
- [ ] 2.7 hf config env list / show <name> / activate <name>

## 3. Verify
- [ ] 3.1 hf config show renders all 6 sections
- [ ] 3.2 hf config set api-url http://x && hf config show reflects change
- [ ] 3.3 hf config env activate kind → [kind] annotation on overridden keys
- [ ] 3.4 hf config doctor runs without panic
- [ ] 3.5 Secret keys show <set>/<not set>, never raw value
```

### specs/ (delta)

`specs/config-registry/spec.md` — ADDED: Go config registry replaces bash HF_CONFIG_REGISTRY array; file-per-property storage; env profile activation by file copy.

---

## Phase 2 — Shared API Client & Resource Types

**Change folder:** `openspec/changes/phase-02-api-client/`  
**Testable without cluster:** Yes (unit tests)  
**dependsOn:** `phase-00-bootstrap`

### proposal.md

**Intent:** Build the shared `internal/api` HTTP client and `internal/resource` Go types that all subsequent command phases depend on.  
**Scope in:** `internal/api`, `internal/resource`, `internal/output`.  
**Scope out:** any command implementations.

### design.md

**`internal/api.Client`**
- `BaseURL` = `{api-url}/api/hyperfleet/{api-version}/`
- Auth: `Authorization: Bearer <token>` when token is set
- Generic methods: `Get[T]`, `Post[T]`, `Patch[T]`, `Delete`
- RFC 7807 error parsing → `APIError{Code, Detail, Status, Title, TraceID, Type, Timestamp}`
- Verbose mode logs method, URL, status code, duration to stderr
- Default timeout: 30s; context-cancellable

**`internal/resource`** — `Cluster`, `NodePool`, `AdapterStatus`, `Condition`, `CloudEvent`, `ListResponse[T]` with JSON tags matching API field names. `Spec` and `Labels` are `map[string]any`.

**`internal/output.Printer`**
- Dispatches on `--output` flag: `json` (indented), `table` (tabwriter), `yaml`
- Colored dot renderer: `True` → green `●`, `False` → red `●`, `Unknown` → yellow `●`, absent → `-`
- Respects `--no-color`
- Dynamic column ordering: fixed cols → `Available` → alphabetical adapter conditions → `Ready` last

### tasks.md

```
## 1. internal/api
- [ ] 1.1 Client struct with NewClient(cfg *config.Store)
- [ ] 1.2 Generic Get[T], Post[T], Patch[T], Delete methods
- [ ] 1.3 RFC 7807 APIError type and parser
- [ ] 1.4 Verbose request/response logging to stderr
- [ ] 1.5 Unit tests for error parser against RFC 7807 JSON fixtures

## 2. internal/resource
- [ ] 2.1 Cluster, ClusterStatus structs
- [ ] 2.2 NodePool, NodePoolStatus structs
- [ ] 2.3 Condition struct
- [ ] 2.4 AdapterStatus struct
- [ ] 2.5 CloudEvent struct
- [ ] 2.6 ListResponse[T] generic wrapper
- [ ] 2.7 Unit tests for JSON round-trip on each type

## 3. internal/output
- [ ] 3.1 Printer with json/table/yaml dispatch
- [ ] 3.2 Colored dot renderer (True/False/Unknown/absent)
- [ ] 3.3 Dynamic column builder for condition tables
- [ ] 3.4 --no-color stripping
- [ ] 3.5 Unit tests for column ordering algorithm
```

### specs/ (delta)

`specs/technical-architecture/spec.md` — ADDED: Go type definitions for all resource structs; RFC 7807 error client; output package with colored dot renderer.

---

## Phase 3 — Cluster Lifecycle

**Change folder:** `openspec/changes/phase-03-cluster/`  
**Testable without cluster:** No (needs API)  
**dependsOn:** `phase-01-config`, `phase-02-api-client`

### proposal.md

**Intent:** Implement all `hf cluster` subcommands — the core CRUD loop for clusters.  
**Scope in:** `cmd/cluster.go`, cluster-scoped API endpoints.  
**Scope out:** nodepools, adapter status, tables.

### design.md

**API base path:** `GET|POST|PATCH|DELETE /clusters[/{id}]`

**ID resolution:** `hf cluster <cmd> [cluster_id]` — explicit arg > `~/.config/hf/cluster-id` > error.

**Side effects:**
- `create` and `search` write found cluster ID to `~/.config/hf/cluster-id` and name to `~/.config/hf/cluster-name`
- `create` with no args uses defaults (`name=my-cluster`, `region=us-east-1`, `version=4.15.0`) — **does not show usage**

**Create payload:**
```json
{"name":"<name>","labels":{"counter":"1","environment":"development","shard":"1","team":"core"},"spec":{"counter":"1","region":"<region>","version":"<version>"}}
```

**Duplicate guard:** `create` searches by name first; warns and skips if found.

**Output defaults:** `get/create/patch` → JSON; `list/search` → table.

### tasks.md

```
## 1. cmd/cluster.go subcommands
- [ ] 1.1 hf cluster create <name> [region] [version] (defaults, no-args creates)
- [ ] 1.2 hf cluster get [cluster_id]
- [ ] 1.3 hf cluster list (table default)
- [ ] 1.4 hf cluster search <name> (sets cluster-id)
- [ ] 1.5 hf cluster patch spec|labels [cluster_id] (JSON from stdin/flag)
- [ ] 1.6 hf cluster delete [cluster_id]
- [ ] 1.7 hf cluster id (reads config file, prints)
- [ ] 1.8 hf cluster conditions [-w] [cluster_id]
- [ ] 1.9 hf cluster conditions-table [cluster_id]
- [ ] 1.10 hf cluster statuses [-w] [cluster_id]

## 2. Verify against real cluster
- [ ] 2.1 Create cluster, confirm cluster-id written to config
- [ ] 2.2 hf cluster search <name> → sets cluster-id
- [ ] 2.3 hf cluster get → JSON output matches API response shape
- [ ] 2.4 hf cluster patch labels → generation increments
- [ ] 2.5 hf cluster conditions → Ready/Available conditions visible
- [ ] 2.6 hf cluster delete → cluster removed
- [ ] 2.7 Duplicate create → [WARN] and skip
```

---

## Phase 4 — NodePool Lifecycle

**Change folder:** `openspec/changes/phase-04-nodepool/`  
**Testable without cluster:** No  
**dependsOn:** `phase-03-cluster`

### proposal.md

**Intent:** Implement all `hf nodepool` subcommands, always scoped to the current cluster-id.  
**Scope in:** `cmd/nodepool.go`, nodepool-scoped API endpoints.

### design.md

**API base path:** `GET|POST|PATCH|DELETE /clusters/{cluster_id}/nodepools[/{id}]`

**Multi-create:** `create <name> [count] [instance-type]` creates `count` nodepools named `<name>-1`, `<name>-2`, …, each with `labels.counter=N`. Sets nodepool-id to the last created. Default: `name=my-nodepool`, count=1, instance-type=`m4`.

**Payload per pool:**
```json
{"name":"<name>-<N>","labels":{"counter":"<N>"},"spec":{"counter":"<N>","platform":{"type":"<type>"},"replicas":1}}
```

### tasks.md

```
## 1. cmd/nodepool.go subcommands
- [ ] 1.1 hf nodepool create <name> [count] [instance-type]
- [ ] 1.2 hf nodepool get [nodepool_id]
- [ ] 1.3 hf nodepool list [cluster_id]
- [ ] 1.4 hf nodepool search <name> (sets nodepool-id)
- [ ] 1.5 hf nodepool patch spec|labels [nodepool_id]
- [ ] 1.6 hf nodepool delete [nodepool_id]
- [ ] 1.7 hf nodepool id
- [ ] 1.8 hf nodepool conditions [-w] [nodepool_id]
- [ ] 1.9 hf nodepool conditions-table [nodepool_id]
- [ ] 1.10 hf nodepool statuses [-w] [nodepool_id]

## 2. Verify against real cluster
- [ ] 2.1 Create 2 nodepools → names get -1/-2 suffix
- [ ] 2.2 nodepool-id written to config for last created
- [ ] 2.3 hf nodepool list → all pools for current cluster
- [ ] 2.4 hf nodepool patch spec → generation increments
- [ ] 2.5 hf nodepool delete → pool removed
```

---

## Phase 5 — Adapter Status

**Change folder:** `openspec/changes/phase-05-adapter-status/`  
**Testable without cluster:** No  
**dependsOn:** `phase-04-nodepool`

### proposal.md

**Intent:** Implement `hf cluster adapter post-status` and `hf nodepool adapter post-status` — the commands that simulate adapters reporting back to the API.

### design.md

**Payload:** Always posts 3 conditions (`Available`, `Applied`, `Health`) all set to the given status value (True/False/Unknown).

```json
{
  "adapter": "<adapter>",
  "conditions": [
    {"type":"Available","status":"<status>","reason":"ManualStatusPost","message":"Status posted via hf adapter post-status","last_transition_time":"<now>"},
    {"type":"Applied",  "status":"<status>","reason":"ManualStatusPost","message":"Status posted via hf adapter post-status","last_transition_time":"<now>"},
    {"type":"Health",   "status":"<status>","reason":"ManualStatusPost","message":"Status posted via hf adapter post-status","last_transition_time":"<now>"}
  ],
  "observed_generation": <generation>,
  "data": {},
  "created_time": "<now>",
  "last_report_time": "<now>"
}
```

**API endpoints:**
- Cluster: `POST /clusters/{cid}/adapter-statuses`
- NodePool: `POST /clusters/{cid}/nodepools/{nid}/adapter-statuses`

### tasks.md

```
## 1. cmd/adapter.go
- [ ] 1.1 hf cluster adapter post-status <adapter> <True|False|Unknown> [generation]
- [ ] 1.2 hf nodepool adapter post-status <adapter> <True|False|Unknown> [generation] [nodepool_id]
- [ ] 1.3 Build 3-condition payload with current timestamp
- [ ] 1.4 Usage error when adapter or status args missing

## 2. Verify
- [ ] 2.1 Post True for all required cluster adapters → Ready flips to True
- [ ] 2.2 Post False → conditions-table shows red dots
- [ ] 2.3 Post Unknown → conditions-table shows yellow dots
- [ ] 2.4 Nodepool adapter post → nodepool Ready converges
```

---

## Phase 6 — Tables & Watch Mode

**Change folder:** `openspec/changes/phase-06-tables-watch/`  
**Testable without cluster:** No  
**dependsOn:** `phase-05-adapter-status`

### proposal.md

**Intent:** Implement the dynamic dot-table combined view (`hf table`, `hf cluster table`, `hf nodepool table`) and watch mode (`-w` flag) for conditions/statuses commands.

### design.md

**Dynamic column algorithm:**
1. Fixed: `NAME`, `GENERATION`
2. `AVAILABLE`
3. All unique adapter condition types across all items, sorted alphabetically
4. `READY` last

**Watch mode** (`internal/watch`):
- `Watch(interval time.Duration, fn func() error)` — clears terminal (`\033[H\033[2J`), calls `fn`, prints "Last updated: <timestamp>" footer
- Default interval: 2s; stops on SIGINT (Ctrl+C)

### tasks.md

```
## 1. internal/watch
- [ ] 1.1 Watch(interval, fn) with ANSI clear and timestamp footer
- [ ] 1.2 SIGINT handler for clean exit
- [ ] 1.3 Wire -w flag to cluster/nodepool conditions and statuses commands

## 2. cmd/table.go
- [ ] 2.1 hf table (clusters + nodepools combined)
- [ ] 2.2 hf cluster table
- [ ] 2.3 hf nodepool table

## 3. Verify
- [ ] 3.1 hf table shows clusters with correct colored dots after adapter status posts
- [ ] 3.2 hf cluster conditions -w refreshes every 2s; Ctrl+C exits cleanly
- [ ] 3.3 Column order: fixed → Available → alpha adapters → Ready
- [ ] 3.4 --no-color renders plain text dots
```

---

## Phase 7 — Database

**Change folder:** `openspec/changes/phase-07-database/`  
**Testable without cluster:** Needs DB only  
**dependsOn:** `phase-01-config`, `phase-02-api-client`

### proposal.md

**Intent:** Direct PostgreSQL access via `jackc/pgx/v5`, replacing `psql`.  
**Scope in:** `internal/db`, `cmd/db.go`.

### design.md

**`internal/db.Client`** — wraps `pgxpool.Pool`, connects from `db-host/port/name/user/password` config. `Query(sql) (headers []string, rows [][]string, err error)`. `Exec(sql) error`.

### tasks.md

```
## 1. internal/db
- [ ] 1.1 Client with NewClient(cfg) using pgxpool
- [ ] 1.2 Query → headers + rows
- [ ] 1.3 Exec for DML statements

## 2. cmd/db.go
- [ ] 2.1 hf db query <sql> or -f <file>
- [ ] 2.2 hf db delete <table> [id]
- [ ] 2.3 hf db delete-all
- [ ] 2.4 hf db statuses
- [ ] 2.5 hf db statuses-delete
- [ ] 2.6 hf db config (show connection info)

## 3. Verify
- [ ] 3.1 hf db query "SELECT 1" returns result row
- [ ] 3.2 hf db statuses shows adapter_statuses table
- [ ] 3.3 hf db config shows host/port/name without password
```

---

## Phase 8 — Maestro

**Change folder:** `openspec/changes/phase-08-maestro/`  
**Testable without cluster:** Needs Maestro endpoint  
**dependsOn:** `phase-02-api-client`

### proposal.md

**Intent:** Maestro resource management via HTTP API; `hf maestro tui` delegates to `maestro-cli` binary.

### design.md

**`internal/maestro.Client`** — `net/http` to `maestro-http-endpoint`, scoped by `maestro-consumer`. Endpoints under `/api/maestro/v1/`. TUI: `exec.LookPath("maestro-cli")`; if missing, print error with install instructions.

### tasks.md

```
## 1. internal/maestro
- [ ] 1.1 Client with NewClient(cfg)
- [ ] 1.2 List, Get, Delete resource bundles
- [ ] 1.3 ListBundles, ListConsumers

## 2. cmd/maestro.go
- [ ] 2.1 hf maestro list
- [ ] 2.2 hf maestro get [name]
- [ ] 2.3 hf maestro delete [name]
- [ ] 2.4 hf maestro bundles
- [ ] 2.5 hf maestro consumers
- [ ] 2.6 hf maestro tui (shell out; clear error if not found)

## 3. Verify
- [ ] 3.1 hf maestro list shows resource bundles
- [ ] 3.2 hf maestro consumers shows consumer list
- [ ] 3.3 hf maestro tui with no binary → actionable error message
```

---

## Phase 9 — Pub/Sub & RabbitMQ

**Change folder:** `openspec/changes/phase-09-pubsub/`  
**Testable without cluster:** Needs GCP credentials / RabbitMQ  
**dependsOn:** `phase-03-cluster`, `phase-04-nodepool`

### proposal.md

**Intent:** Publish CloudEvents to GCP Pub/Sub and RabbitMQ, replacing `gcloud` and custom curl calls.

### design.md

- GCP: `cloud.google.com/go/pubsub` — `PublishClusterChange`, `PublishNodePoolChange`
- RabbitMQ: `github.com/rabbitmq/amqp091-go` — `PublishClusterChange`
- CloudEvent payload: `{specversion, type, source, id, data}`

### tasks.md

```
## 1. internal/pubsub
- [ ] 1.1 GCP PubSub client with project from config
- [ ] 1.2 PublishClusterChange(topic, cluster)
- [ ] 1.3 PublishNodePoolChange(topic, nodepool)
- [ ] 1.4 RabbitMQ client (amqp091-go) with config credentials
- [ ] 1.5 RabbitMQ PublishClusterChange(exchange, routingKey, cluster)

## 2. cmd/pubsub.go + cmd/rabbitmq.go
- [ ] 2.1 hf pubsub list [filter]
- [ ] 2.2 hf pubsub publish cluster <topic>
- [ ] 2.3 hf pubsub publish nodepool <topic>
- [ ] 2.4 hf rabbitmq publish cluster <exchange> [routing-key]

## 3. Verify
- [ ] 3.1 hf pubsub list returns topic list for gcp-project
- [ ] 3.2 hf pubsub publish cluster → message visible in GCP console
- [ ] 3.3 hf rabbitmq publish cluster → message in exchange
```

---

## Phase 10 — Kubernetes

**Change folder:** `openspec/changes/phase-10-kubernetes/`  
**Testable without cluster:** Needs kubeconfig  
**dependsOn:** `phase-01-config`

### proposal.md

**Intent:** Port-forward lifecycle, pod log streaming, debug pod, and kube context — all via `client-go`. No `kubectl` required.

### design.md

**`internal/kube.Client`** — kubeconfig loaded from `--kubeconfig` flag → `KUBECONFIG` env → `~/.kube/config`.

**Port-forward:** PID files at `~/.config/hf/pf-<service>.pid`. Services: API (`pf-api-port`), Postgres (`pf-pg-port`), Maestro HTTP/gRPC.

**Log streaming:** label/name filter, multi-pod fan-out (replaces `stern`). Streams to stdout with pod-name prefix.

**Pod exec:** `hf kube curl` execs `curl` inside a running pod. `hf kube debug` attaches an ephemeral debug container.

### tasks.md

```
## 1. internal/kube
- [ ] 1.1 NewClient(ctx, kubeconfig, context)
- [ ] 1.2 StartPortForward(service, localPort, remotePort) → writes PID file
- [ ] 1.3 StopPortForward(service) → kills PID
- [ ] 1.4 PortForwardStatus() → list of running forwards
- [ ] 1.5 StreamLogs(pattern, namespace, follow) → fan-out across matching pods
- [ ] 1.6 ExecInPod(pod, namespace, cmd) → exec + stream output

## 2. cmd/kube.go + cmd/logs.go
- [ ] 2.1 hf kube port-forward start
- [ ] 2.2 hf kube port-forward stop
- [ ] 2.3 hf kube port-forward status
- [ ] 2.4 hf kube context
- [ ] 2.5 hf kube curl [opts] <url>
- [ ] 2.6 hf kube debug <deployment> [namespace]
- [ ] 2.7 hf logs <pattern> [flags]
- [ ] 2.8 hf logs adapter <pattern> [flags]

## 3. Verify (needs kubeconfig)
- [ ] 3.1 hf kube context prints current context
- [ ] 3.2 hf kube port-forward start → ports accessible on localhost
- [ ] 3.3 hf kube port-forward status → shows running forwards with ports
- [ ] 3.4 hf kube port-forward stop → ports no longer accessible
- [ ] 3.5 hf logs hyperfleet → streams pod logs with pod prefix
```

---

## Phase 11 — Repos

**Change folder:** `openspec/changes/phase-11-repos/`  
**Testable without cluster:** Needs GitHub token  
**dependsOn:** `phase-00-bootstrap`

### proposal.md

**Intent:** `hf repos` — GitHub repository status overview for the configured registry owner, replacing `hf.repos.sh`.

### design.md

**`google/go-github/v60`** client. Auth from `GITHUB_TOKEN` env or config token. `registry` config property is the GitHub user/org.

### tasks.md

```
## 1. internal/repos
- [ ] 1.1 Client with NewClient(cfg)
- [ ] 1.2 ListRepos(owner) → name, open PRs, default branch, CI status

## 2. cmd/repos.go
- [ ] 2.1 hf repos → table of repos for registry owner

## 3. Verify
- [ ] 3.1 hf repos → table renders with correct columns
```

---

## Phase 12 — Workflow & Interactive Commands

**Change folder:** `openspec/changes/phase-12-workflow/`  
**Testable without cluster:** No  
**dependsOn:** `phase-03-cluster` through `phase-06-tables-watch`

### proposal.md

**Intent:** `hf workflow run` and `hf workflow api-only` — end-to-end scripted lifecycle sequences that call internal packages directly (no subprocess calls to other `hf` subcommands).

### tasks.md

```
## 1. cmd/workflow.go
- [ ] 1.1 hf workflow api-only → config → create cluster → nodepools → post statuses → verify ready
- [ ] 1.2 hf workflow run → full stack (includes port-forward steps)
- [ ] 1.3 Progress output with [INFO] prefix at each step
- [ ] 1.4 Abort on first failure with [ERROR] + exit non-zero

## 2. Verify
- [ ] 2.1 hf workflow api-only runs to completion on real cluster
- [ ] 2.2 hf workflow run runs to completion with kubeconfig set
```

---

## Phase 13 — Non-Functional Requirements

**Change folder:** `openspec/changes/phase-13-nfr/`  
**Testable without cluster:** Mostly yes  
**dependsOn:** all prior phases

### proposal.md

**Intent:** Shell completions, plugin architecture, GoReleaser release pipeline, integration tests, and error-handling polish.

### tasks.md

```
## 1. Shell Completions
- [ ] 1.1 hf completion bash|zsh|fish|powershell (Cobra built-in)

## 2. Plugin Architecture
- [ ] 2.1 internal/plugin: scan PATH for hf-* executables on unknown command
- [ ] 2.2 Delegate to plugin with remaining args; forward exit code

## 3. GoReleaser
- [ ] 3.1 Finalize .goreleaser.yaml (checksums, homebrew tap, deb/rpm optional)
- [ ] 3.2 GitHub Actions CI workflow (lint, test, build)
- [ ] 3.3 GitHub Actions release workflow (on tag push)

## 4. Error & Usage Polish
- [ ] 4.1 API errors render in current --output format (json/table/yaml)
- [ ] 4.2 CLI-level errors → stderr [ERROR] prefix, non-zero exit
- [ ] 4.3 [WARN]/[INFO] to stderr, exit 0
- [ ] 4.4 All commands exit 0 on API errors (backward compat)

## 5. Testing
- [ ] 5.1 Unit tests: internal/config, internal/output, internal/api error parsing
- [ ] 5.2 Integration tests tagged //go:build integration; skip without cluster
- [ ] 5.3 CI runs unit tests on every PR
```

---

## Implementation Order Summary

| Phase | Change Folder | Needs Cluster | dependsOn |
|---|---|---|---|
| 0 | `phase-00-bootstrap` | No | — |
| 1 | `phase-01-config` | No | 0 |
| 2 | `phase-02-api-client` | No (unit tests) | 0 |
| 3 | `phase-03-cluster` | API | 1, 2 |
| 4 | `phase-04-nodepool` | API | 3 |
| 5 | `phase-05-adapter-status` | API | 4 |
| 6 | `phase-06-tables-watch` | API | 5 |
| 7 | `phase-07-database` | DB | 1, 2 |
| 8 | `phase-08-maestro` | Maestro | 2 |
| 9 | `phase-09-pubsub` | GCP/RabbitMQ | 3, 4 |
| 10 | `phase-10-kubernetes` | kubeconfig | 1 |
| 11 | `phase-11-repos` | GitHub token | 0 |
| 12 | `phase-12-workflow` | Full stack | 3–6 |
| 13 | `phase-13-nfr` | No | all |

Start phases 0 → 1 → 2 immediately (no cluster needed). Share the kubeconfig when ready to begin phase 3.
