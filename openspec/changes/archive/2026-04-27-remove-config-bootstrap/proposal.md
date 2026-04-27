# Proposal: Remove hf config bootstrap from Technical Architecture Spec

## Why

`hf config bootstrap` was superseded by `hf config env new` in phase-02-config-ux and was never
implemented. The config spec already documents the removal, but the technical architecture spec
still lists `bootstrap` in both the module layout comment and the Cobra command tree, making it
inconsistent with the authoritative config spec.

## What Changes

- Remove `bootstrap|` from the `config.go` comment in the module layout
- Remove `bootstrap [env-name]` from the `config` subtree in the Cobra command tree
- Add `new [name]` under `env` in the Cobra command tree (the replacement, currently missing)

## Scope Out

- Archived change directories under `openspec/changes/archive/` — historical record, not updated
- `openspec/specs/config/spec.md` — already correct; no changes needed

## Testing Scope

No code changes. Pure spec correction; no tests required.
