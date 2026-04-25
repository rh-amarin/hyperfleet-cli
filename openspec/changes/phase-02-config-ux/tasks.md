# Tasks: Phase 02 — Config UX Simplification

## 1. Remove HF_* env var overrides from store
- [ ] 1.1 Delete `applyEnvVars()` function from store.go
- [ ] 1.2 Remove `applyEnvVars(&s.cfg)` call from NewStore
- [ ] 1.3 Delete `envVarFor` map from resolve.go
- [ ] 1.4 Remove `[ENV]` source annotation branch from Resolve()

## 2. Update defaults in types.go
- [ ] 2.1 Set database.user default to "hyperfleet"
- [ ] 2.2 Set database.name default to "hyperfleet"
- [ ] 2.3 Set database.password default to "foobar-bizz-buzz"
- [ ] 2.4 Set rabbitmq.host default to "rabbitmq"
- [ ] 2.5 Set rabbitmq.password default to "guest"

## 3. Add requireActiveEnv helper in cmd/config.go
- [ ] 3.1 Add `requireActiveEnv(s *config.Store) error`
- [ ] 3.2 Error message includes 'hf config env new' and 'hf config env activate' guidance

## 4. Remove hf config bootstrap, add hf config env new
- [ ] 4.1 Delete configBootstrapCmd and remove from init()
- [ ] 4.2 Add configEnvNewCmd under configEnvCmd with optional [name] arg
- [ ] 4.3 Prompt for name interactively if not given as argument
- [ ] 4.4 Use updated prompt defaults (db user/name/password, rabbitmq host/password)
- [ ] 4.5 Success message tells user to run 'hf config env activate <name>'

## 5. Simplify hf config show
- [ ] 5.1 Remove source annotation column from printConfigShow output
- [ ] 5.2 Print active environment name at the top of output
- [ ] 5.3 Call requireActiveEnv at start of configShowCmd.RunE and configCmd.RunE

## 6. Gate commands on active environment
- [ ] 6.1 configSetCmd.RunE calls requireActiveEnv
- [ ] 6.2 configClearCmd.RunE calls requireActiveEnv for section.key path (not 'state')
- [ ] 6.3 configDoctorCmd.RunE calls requireActiveEnv

## 7. Verify
- [ ] 7.1 `HF_TOKEN=x hf config show` — env var has no effect (not applied to config)
- [ ] 7.2 `hf config show` with no active env exits non-zero with guidance message
- [ ] 7.3 `hf config env new` (no arg) prompts for name then config values
- [ ] 7.4 `hf config env new dev` skips name prompt, creates environments/dev.yaml
- [ ] 7.5 `hf config show` with active env shows env name at top, no source column
- [ ] 7.6 `hf config set hyperfleet.api-url http://x.com` with no active env exits non-zero
- [ ] 7.7 `hf config doctor` with no active env exits non-zero
- [ ] 7.8 `hf config env list` with no active env succeeds
- [ ] 7.9 Pressing enter on all prompts in `hf config env new` produces db user=hyperfleet, name=hyperfleet, password=foobar-bizz-buzz, rabbitmq host=rabbitmq, password=guest
