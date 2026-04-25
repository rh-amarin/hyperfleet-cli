# Design: Phase 01 — Configuration System

## Spec Reconciliation

`config-model/spec.md` (T2 — technical spec) is authoritative over
`config/spec.md` and `config-registry/spec.md` (functional specs written to
describe the bash scripts). The YAML model wins. Delta specs in this change
update the two conflicting functional specs.

## File Layout

```
~/.config/hf/
├── config.yaml          ← static settings (sections: hyperfleet, kubernetes,
│                           maestro, port-forward, database, rabbitmq, registry)
├── state.yaml           ← flat runtime state (active-environment, cluster-id,
│                           cluster-name, nodepool-id)
└── environments/
    ├── kind.yaml        ← overrides: same nested structure as config.yaml,
    └── gke-prod.yaml       only the keys that differ
```

Both files are created with mode `0600` on first use.

## config.yaml structure

```yaml
hyperfleet:
  api-url: "http://localhost:8000"
  api-version: "v1"
  token: ""
  gcp-project: "hcm-hyperfleet"

kubernetes:
  context: ""
  namespace: ""

maestro:
  consumer: "cluster1"
  http-endpoint: "http://localhost:8100"
  grpc-endpoint: "localhost:8090"
  namespace: "maestro"

port-forward:
  api-port: 8000
  pg-port: 5432
  maestro-http-port: 8100
  maestro-http-remote-port: 8000
  maestro-grpc-port: 8090

database:
  host: "localhost"
  port: 5432
  name: ""
  user: ""
  password: ""

rabbitmq:
  host: "localhost"
  mgmt-port: 15672
  user: "guest"
  password: ""
  vhost: "/"

registry:
  name: ""   # defaults to $USER at runtime
```

## state.yaml structure (flat, top-level keys)

```yaml
active-environment: ""
cluster-id: ""
cluster-name: ""
nodepool-id: ""
```

Written atomically: write to temp file, then os.Rename.

## Environment profiles

```yaml
# environments/kind.yaml
kubernetes:
  context: "kind-kind"
  namespace: "amarin-ns1"
```

Activation: write `active-environment: kind` to `state.yaml`. No file copying.
On every config read, if `active-environment` is set, deep-merge the env YAML
on top of config.yaml (env wins on key conflict).

## Precedence chain (highest → lowest)

1. CLI flag (`--api-url`, `--api-token`)
2. `HF_*` environment variable
3. Active env profile (`environments/<name>.yaml`)
4. `config.yaml`
5. Built-in defaults (defined in registry)

## Source annotations

`hf config show` annotates each value with its source:
- `[default]` — built-in default
- `[config]`  — from config.yaml
- `[env:<name>]` — from active env profile
- `[ENV]`  — from HF_* environment variable
- `[flag]` — from CLI flag

## Key Go types

```go
// internal/config

type Config struct {
    Hyperfleet  HyperfleetConfig  `yaml:"hyperfleet"`
    Kubernetes  KubernetesConfig  `yaml:"kubernetes"`
    Maestro     MaestroConfig     `yaml:"maestro"`
    PortForward PortForwardConfig `yaml:"port-forward"`
    Database    DatabaseConfig    `yaml:"database"`
    RabbitMQ    RabbitMQConfig    `yaml:"rabbitmq"`
    Registry    RegistryConfig    `yaml:"registry"`
}

type State struct {
    ActiveEnvironment string `yaml:"active-environment"`
    ClusterID         string `yaml:"cluster-id"`
    ClusterName       string `yaml:"cluster-name"`
    NodePoolID        string `yaml:"nodepool-id"`
}

type Store struct {
    dir    string
    cfg    Config   // merged: defaults ← config.yaml ← env profile
    state  State
    source map[string]string // key → source annotation
}

func NewStore(dir string) (*Store, error)
func (s *Store) Save() error              // write config.yaml
func (s *Store) SetState(key, val string) error  // atomic write to state.yaml
func (s *Store) ClearState() error
func (s *Store) EnvList() ([]EnvProfile, error)
func (s *Store) EnvActivate(name string) error
func (s *Store) EnvDeactivate() error
```

## bootstrap behaviour

`hf config bootstrap [env-name]`:
- Prompts for each section's properties interactively (reads from stdin)
- If `env-name` given: saves to `environments/<env-name>.yaml`
- If no `env-name`: saves to `config.yaml`
- Skips blank input (keeps current value)

## Legacy migration

On startup, if `config.yaml` does not exist but flat property files do
(e.g., `~/.config/hf/api-url`), detect legacy format, print a warning, and
offer to migrate with `hf config migrate`.

## Package layout

```
internal/config/
├── types.go      — Config, State, section structs, defaults
├── store.go      — Store: NewStore, load, Save, SetState, ClearState
├── env.go        — EnvList, EnvActivate, EnvDeactivate, EnvShow
├── resolve.go    — value resolution with precedence + source annotation
└── ids.go        — ClusterID/NodePoolID helpers

cmd/config.go     — all hf config subcommands
```
