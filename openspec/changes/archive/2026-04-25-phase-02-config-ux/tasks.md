# Tasks: Phase 02 — Config UX Simplification

## 1. Remove HF_* env var overrides from store
- [x] 1.1 Delete `applyEnvVars()` function from store.go
- [x] 1.2 Remove `applyEnvVars(&s.cfg)` call from NewStore
- [x] 1.3 Delete `envVarFor` map from resolve.go
- [x] 1.4 Remove `[ENV]` source annotation branch from Resolve()

## 2. Update defaults in types.go
- [x] 2.1 Set database.user default to "hyperfleet"
- [x] 2.2 Set database.name default to "hyperfleet"
- [x] 2.3 Set database.password default to "foobar-bizz-buzz"
- [x] 2.4 Set rabbitmq.host default to "rabbitmq"
- [x] 2.5 Set rabbitmq.password default to "guest"

## 3. Add requireActiveEnv helper in cmd/config.go
- [x] 3.1 Add `requireActiveEnv(s *config.Store) error`
- [x] 3.2 Error message includes 'hf config env new' and 'hf config env activate' guidance

## 4. Remove hf config bootstrap, add hf config env new
- [x] 4.1 Delete configBootstrapCmd and remove from init()
- [x] 4.2 Add configEnvNewCmd under configEnvCmd with optional [name] arg
- [x] 4.3 Prompt for name interactively if not given as argument
- [x] 4.4 Use updated prompt defaults (db user/name/password, rabbitmq host/password)
- [x] 4.5 Success message tells user to run 'hf config env activate <name>'

## 5. Simplify hf config show
- [x] 5.1 Remove source annotation column from output (read directly from Cfg())
- [x] 5.2 Print active environment name at the top of output
- [x] 5.3 Call requireActiveEnv at start of configShowCmd.RunE and configCmd.RunE

## 6. Gate commands on active environment
- [x] 6.1 configSetCmd.RunE calls requireActiveEnv
- [x] 6.2 configClearCmd.RunE calls requireActiveEnv for section.key path (not 'state')
- [x] 6.3 configDoctorCmd.RunE calls requireActiveEnv

## 7. Verify
- [x] 7.1 `HF_TOKEN=x hf config show` — env var has no effect (no-active-env error, not a token error)
- [x] 7.2 `hf config show` with no active env exits non-zero with guidance message
- [x] 7.3 `hf config env new` (no arg) prompts for name then config values
- [x] 7.4 `hf config env new dev` skips name prompt, creates environments/dev.yaml
- [x] 7.5 `hf config show` with active env shows env name at top, no source column, correct port values
- [x] 7.6 `hf config set hyperfleet.api-url http://x.com` with no active env exits non-zero
- [x] 7.7 `hf config doctor` with no active env exits non-zero
- [x] 7.8 `hf config env list` with no active env succeeds
- [x] 7.9 Pressing enter on all prompts in `hf config env new` produces db user=hyperfleet, name=hyperfleet, password=foobar-bizz-buzz, rabbitmq host=rabbitmq, password=guest
