# Proposal: Config UX Simplification

## Intent

The phase-01 config system exposes implementation internals in its UX. Source annotations
(`[default]`, `[config]`, `[env:kind]`) confuse users by revealing the precedence chain
instead of answering "what is my config right now?". The `bootstrap` command name is opaque.
Commands silently succeed when no environment is configured, making it impossible to tell
whether the CLI is pointed at anything real.

## Scope

Changes to `hf config` command group only. No changes to storage format, state.yaml,
HF_* env var behaviour, or any other command group.

## Approach

- Replace `hf config bootstrap` with `hf config env new [name]`, prompting for the name
  interactively if not supplied
- Remove source annotations from `hf config show`; display active environment name at top
- Gate all operational config commands on an active environment; fail fast with guidance
- Update prompt defaults in `hf config env new` to match real HyperFleet deployment values

## Capabilities

| Capability | Status |
|---|---|
| config/hf-config-show | MODIFIED |
| config/hf-config-env-new | ADDED |
| config/hf-config-bootstrap | REMOVED |
| config/active-env-guard | ADDED |
