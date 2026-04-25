# Tasks: Phase 01 — Configuration System

## 1. internal/config — types and defaults
- [x] 1.1 Define section structs (HyperfleetConfig, KubernetesConfig, MaestroConfig, PortForwardConfig, DatabaseConfig, RabbitMQConfig, RegistryConfig)
- [x] 1.2 Define Config and State structs with YAML tags
- [x] 1.3 Define defaults() returning Config populated with all default values
- [x] 1.4 Define secret field list (hyperfleet.token, database.password, rabbitmq.password)

## 2. internal/config — store (load/save)
- [x] 2.1 NewStore(dir): create dir + environments/ subdir, init config.yaml with defaults if missing, init empty state.yaml if missing
- [x] 2.2 Load config.yaml → Config (merge defaults ← file values)
- [x] 2.3 Load state.yaml → State
- [x] 2.4 If active-environment set, deep-merge environments/<name>.yaml on top of Config
- [x] 2.5 Save(): write config.yaml atomically (mode 0600)
- [x] 2.6 SetState(key, val): update State field and write state.yaml atomically
- [x] 2.7 ClearState(): zero all state fields, write state.yaml
- [x] 2.8 Detect legacy file-per-property layout on startup; print [WARN] if found

## 3. internal/config — value resolution with source annotation
- [x] 3.1 Resolve(path string) (value, source string): walk precedence chain (flag > ENV > env-profile > config.yaml > default)
- [x] 3.2 Expose resolved flat key-value map for display (e.g. "hyperfleet.api-url" → value)
- [x] 3.3 Track source per key: [default], [config], [env:<name>], [ENV], [flag]

## 4. internal/config — env profile management
- [x] 4.1 EnvList(): scan environments/ dir, return name + overridden-key count + isActive
- [x] 4.2 EnvActivate(name): verify file exists, write active-environment to state.yaml, reload
- [x] 4.3 EnvDeactivate(): clear active-environment from state.yaml, reload
- [x] 4.4 EnvShow(name): return resolved Config annotated with which keys come from that env

## 5. internal/config — ID helpers
- [x] 5.1 ClusterID(s *Store, arg string) (string, error): arg > state.yaml cluster-id > error
- [x] 5.2 SetClusterID(s *Store, id, name string) error: write both to state.yaml
- [x] 5.3 NodePoolID(s *Store, arg string) (string, error)
- [x] 5.4 SetNodePoolID(s *Store, id string) error

## 6. cmd/config.go
- [x] 6.1 hf config (no args): usage + env list + full config show
- [x] 6.2 hf config show [env]: sections with values and source annotations; env overrides in cyan
- [x] 6.3 hf config set <section.key> <value>: update config.yaml via Store.Save
- [x] 6.4 hf config clear <section.key>: reset field to default in config.yaml
- [x] 6.5 hf config clear state: call Store.ClearState()
- [x] 6.6 hf config doctor: ●/○ readiness per command group
- [x] 6.7 hf config bootstrap [env-name]: interactive prompts, save to config.yaml or environments/<env>.yaml
- [x] 6.8 hf config env list / show <name> / activate <name> / deactivate

## 7. Wire store into root command
- [x] 7.1 Initialise Store in PersistentPreRunE, honour --config flag and HF_CONFIG_DIR env var
- [x] 7.2 Apply --api-url and --api-token flag overrides into resolved config after load

## 8. Verify
- [x] 8.1 hf config show: renders all 7 sections with correct defaults and [default] annotations
- [x] 8.2 hf config set hyperfleet.api-url http://example.com → show reflects [config] annotation
- [x] 8.3 hf config clear hyperfleet.api-url → resets to default with [default] annotation
- [x] 8.4 hf config env list: shows profiles in environments/ with property counts
- [x] 8.5 hf config env activate kind → active-environment=kind in state.yaml, show marks active
- [x] 8.6 hf config env deactivate → active-environment cleared
- [x] 8.7 hf config doctor: runs without panic, shows correct ●/○ per command
- [x] 8.8 hf config bootstrap dev: prompts answered via stdin pipe, creates environments/dev.yaml
- [x] 8.9 Secrets (token, database.password, rabbitmq.password) show <set>/<not set>
- [x] 8.10 hf config clear state: empties cluster-id, nodepool-id, active-environment
