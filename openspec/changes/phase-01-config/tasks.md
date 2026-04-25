# Tasks: Phase 01 — Configuration System

## 1. internal/config — Registry
- [x] 1.1 Define Entry struct (Section, Key, Default, EnvVar, IsSecret)
- [x] 1.2 Define Registry slice with all 30 properties

## 2. internal/config — Store
- [x] 2.1 NewStore(dir string): create dir if missing, load active env
- [x] 2.2 Get(key): precedence chain (flag skipped here — handled at cmd layer)
- [x] 2.3 Set(key, value): write ~/.config/hf/<key>
- [x] 2.4 Clear(key): delete file; Clear all: delete all key files
- [x] 2.5 EnvList(): scan dir for <env>.<key> files, return unique envs with counts
- [x] 2.6 EnvActivate(name): copy ~/.config/hf/<env>.<key> → ~/.config/hf/<key>
- [x] 2.7 EnvShow(name): return all entries annotated with whether env overrides

## 3. internal/config — ID helpers
- [x] 3.1 ClusterID(store, argVal string) (string, error): arg > file > error
- [x] 3.2 SetClusterID(store, id, name string) error
- [x] 3.3 NodePoolID(store, argVal string) (string, error)
- [x] 3.4 SetNodePoolID(store, id string) error

## 4. cmd/config.go
- [x] 4.1 hf config (no args): usage + env list + full config sections
- [x] 4.2 hf config show [env]: grouped sections, cyan [env] annotation
- [x] 4.3 hf config set <key> <value>
- [x] 4.4 hf config clear <key> and clear all
- [x] 4.5 hf config doctor: ● / ○ per command based on required keys
- [x] 4.6 hf config bootstrap [env]: interactive prompts per section
- [x] 4.7 hf config env list
- [x] 4.8 hf config env show <name>
- [x] 4.9 hf config env activate <name>

## 5. Wire store into root command
- [x] 5.1 Initialise Store in PersistentPreRunE on root, inject via context or package var
- [x] 5.2 Honour --config flag as config dir override

## 6. Verify
- [x] 6.1 hf config show renders all 6 sections with correct defaults
- [x] 6.2 hf config set api-url http://example.com && hf config show reflects change
- [x] 6.3 hf config clear api-url resets to default
- [x] 6.4 hf config env list shows any existing env profiles (or empty list)
- [x] 6.5 hf config doctor runs without panic
- [x] 6.6 Secret keys (token, db-password, rabbitmq-password) show <set>/<not set>
