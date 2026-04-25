# Tasks: Phase 02 — Config UX Simplification

## 1. Update defaults in types.go
- [ ] 1.1 Set database.user default to "hyperfleet"
- [ ] 1.2 Set database.name default to "hyperfleet"
- [ ] 1.3 Set database.password default to "foobar-bizz-buzz"
- [ ] 1.4 Set rabbitmq.host default to "rabbitmq"
- [ ] 1.5 Set rabbitmq.password default to "guest"

## 2. Add requireActiveEnv helper in cmd/config.go
- [ ] 2.1 Add `requireActiveEnv(s *config.Store) error`
- [ ] 2.2 Error message includes 'hf config env new' and 'hf config env activate' guidance

## 3. Remove hf config bootstrap, add hf config env new
- [ ] 3.1 Delete configBootstrapCmd and remove from init()
- [ ] 3.2 Add configEnvNewCmd under configEnvCmd with optional [name] arg
- [ ] 3.3 Prompt for name interactively if not given as argument
- [ ] 3.4 Use updated prompt defaults (db user/name/password, rabbitmq host/password)
- [ ] 3.5 Success message tells user to run 'hf config env activate <name>'

## 4. Simplify hf config show
- [ ] 4.1 Remove source annotation column from printConfigShow output
- [ ] 4.2 Print active environment name at the top of output
- [ ] 4.3 Call requireActiveEnv at start of configShowCmd.RunE and configCmd.RunE

## 5. Gate commands on active environment
- [ ] 5.1 configSetCmd.RunE calls requireActiveEnv
- [ ] 5.2 configClearCmd.RunE calls requireActiveEnv for section.key path (not 'state')
- [ ] 5.3 configDoctorCmd.RunE calls requireActiveEnv

## 6. Verify
- [ ] 6.1 `hf config show` with no active env exits non-zero with guidance message
- [ ] 6.2 `hf config env new` (no arg) prompts for name then config values
- [ ] 6.3 `hf config env new dev` skips name prompt, creates environments/dev.yaml
- [ ] 6.4 `hf config show` with active env shows env name at top, no source column
- [ ] 6.5 `hf config set hyperfleet.api-url http://x.com` with no active env exits non-zero
- [ ] 6.6 `hf config doctor` with no active env exits non-zero
- [ ] 6.7 `hf config env list` with no active env succeeds
- [ ] 6.8 Pressing enter on all prompts in `hf config env new` produces db user=hyperfleet, name=hyperfleet, password=foobar-bizz-buzz, rabbitmq host=rabbitmq, password=guest
