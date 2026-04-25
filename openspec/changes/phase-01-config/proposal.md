# Proposal: Phase 01 — Configuration System

## Intent

Implement `hf config` and all subcommands using the **split YAML model** defined
in `openspec/specs/config-model/spec.md`. Configuration lives in two files:
`config.yaml` (static settings) and `state.yaml` (active runtime state). Named
environment profiles are stored under `environments/<name>.yaml`.

This phase also resolves a spec contradiction: `config-registry/spec.md` and
`config/spec.md` both describe file-per-property storage (carried over from the
bash scripts), which conflicts with the authoritative `config-model/spec.md`.
Delta specs in this change update those two specs to align with the YAML model.

No API calls — this phase runs standalone.

## Scope

**In scope:**
- `internal/config` package: YAML loader/writer for config.yaml + state.yaml,
  env profiles under `environments/`, full precedence chain, secret masking,
  source annotations, atomic state writes, legacy-migration detection
- `cmd/config.go`: all subcommands (show, set, clear, doctor, bootstrap, env)
- Delta specs correcting `config/spec.md` and `config-registry/spec.md`

**Out of scope:**
- API calls, Kubernetes, database
- Output package (json/table/yaml) — phase-02
