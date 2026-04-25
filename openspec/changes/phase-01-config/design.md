# Design: Phase 01 — Configuration System

## Storage Model

One file per property at `~/.config/hf/<key>`.
Env profiles stored as `~/.config/hf/<env>.<key>` files.
Active env pointer stored in `~/.config/hf/.active-env`.

## Precedence Chain (highest to lowest)

1. CLI flag (`--api-url`, `--api-token`)
2. `HF_*` environment variable (e.g. `HF_API_URL`)
3. Active env profile file (`~/.config/hf/<env>.<key>`)
4. Base config file (`~/.config/hf/<key>`)
5. Default value from registry

## Config Registry (all 30 properties)

| Section    | Key                         | Default              | Secret |
|------------|-----------------------------|----------------------|--------|
| hyperfleet | api-url                     | http://localhost:8000 |        |
| hyperfleet | api-version                 | v1                   |        |
| hyperfleet | token                       |                      | ✓      |
| hyperfleet | context                     |                      |        |
| hyperfleet | namespace                   |                      |        |
| hyperfleet | gcp-project                 | hcm-hyperfleet       |        |
| hyperfleet | cluster-id                  |                      |        |
| hyperfleet | cluster-name                |                      |        |
| hyperfleet | nodepool-id                 |                      |        |
| maestro    | maestro-consumer            | cluster1             |        |
| maestro    | maestro-http-endpoint       | http://localhost:8100 |       |
| maestro    | maestro-grpc-endpoint       | localhost:8090       |        |
| maestro    | maestro-namespace           | maestro              |        |
| portforward| pf-api-port                 | 8000                 |        |
| portforward| pf-pg-port                  | 5432                 |        |
| portforward| pf-maestro-http-port        | 8100                 |        |
| portforward| pf-maestro-http-remote-port | 8000                 |        |
| portforward| pf-maestro-grpc-port        | 8090                 |        |
| database   | db-host                     | localhost            |        |
| database   | db-port                     | 5432                 |        |
| database   | db-name                     |                      |        |
| database   | db-user                     |                      |        |
| database   | db-password                 |                      | ✓      |
| rabbitmq   | rabbitmq-host               | localhost            |        |
| rabbitmq   | rabbitmq-mgmt-port          | 15672                |        |
| rabbitmq   | rabbitmq-user               | guest                |        |
| rabbitmq   | rabbitmq-password           |                      | ✓      |
| rabbitmq   | rabbitmq-vhost              | /                    |        |
| registry   | registry                    | (current $USER)      |        |

## Package Layout

```
internal/config/
├── registry.go   — Entry struct, Registry slice (all 30 entries)
├── store.go      — Store struct, Get/Set/Clear, env helpers
└── ids.go        — ClusterID/NodePoolID read+write helpers
cmd/
└── config.go     — all hf config subcommands
```

## Key Types

```go
type Entry struct {
    Section  string
    Key      string
    Default  string
    EnvVar   string // e.g. "HF_API_URL"
    IsSecret bool
}

type Store struct {
    dir       string // ~/.config/hf
    activeEnv string // read from .active-env file, "" if none
}

func NewStore(dir string) (*Store, error)
func (s *Store) Get(key string) string          // full precedence chain
func (s *Store) Set(key, value string) error    // write ~/.config/hf/<key>
func (s *Store) Clear(key string) error         // delete file
func (s *Store) ClearAll() error                // delete all key files
func (s *Store) ActiveEnv() string
func (s *Store) EnvList() ([]EnvProfile, error)
func (s *Store) EnvActivate(name string) error  // copy <env>.<key> → <key>
func (s *Store) EnvShow(name string) []EnvEntry // all keys with override flag
```

## Command Behaviours

| Command | Behaviour |
|---|---|
| `hf config` (no args) | usage text + env list + full config (same as show) |
| `hf config show [env]` | grouped sections; env overrides annotated `[env]` in cyan |
| `hf config set <key> <value>` | writes file, prints confirmation |
| `hf config clear <key>` | deletes file |
| `hf config clear all` | deletes all key files |
| `hf config doctor` | lists all scripts/commands; ● ready, ○ missing config |
| `hf config bootstrap [env]` | interactive prompts per section |
| `hf config env list` | lists env profiles with property counts |
| `hf config env show <name>` | full config with env overrides highlighted |
| `hf config env activate <name>` | copies env files over base files |
