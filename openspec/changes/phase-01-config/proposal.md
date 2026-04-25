# Proposal: Phase 01 — Configuration System

## Intent

Implement `hf config` and all subcommands using file-per-property storage at
`~/.config/hf/<key>`, exactly mirroring the behaviour of `hf.config.sh` and
`hf.lib.sh`. No API calls — this phase runs standalone.

## Scope

**In scope:**
- `internal/config` package: registry of all 30 properties, file-based store,
  env profile management, ID helpers (cluster-id, nodepool-id)
- `cmd/config.go`: all subcommands (show, set, clear, doctor, bootstrap, env)
- Config precedence: CLI flag > HF_* env var > env profile file > base file > default
- Secret masking: token, db-password, rabbitmq-password shown as `<set>`/`<not set>`
- Env profile display: active env marked `●` (green), inactive `○`

**Out of scope:**
- API calls, Kubernetes, database
- Output package (json/table/yaml) — that is phase-02
