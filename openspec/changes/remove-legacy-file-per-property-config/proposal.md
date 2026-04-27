# Proposal: Remove Legacy File-Per-Property Config Detection

## Why

The `warnLegacy()` method in `internal/config/store.go` detects pre-YAML flat files
(`api-url`, `api-version`, `token`, `cluster-id`) in the config directory and prints a
warning pointing users to `hf config migrate`. That command was never implemented and the
file-per-property shell-script era is long past. All users are on the YAML config model.
The warning is dead code that references a non-existent command, and `config-model/spec.md`
carries a migration requirement that will never be built.

## What Changes

- `internal/config/store.go` — remove the `warnLegacy()` method and its call site in `NewStore`
- `openspec/specs/config-model/spec.md` — remove the "Migration from File-Per-Property"
  requirement (two scenarios: auto-detect and no-legacy), and trim the backwards-compatibility
  clause from the Purpose paragraph
- `openspec/specs/config-registry/spec.md` — remove the "(Previously: file-per-property…)"
  historical note

## Scope Out

- Archived change directories under `openspec/changes/archive/` — historical record, not updated
- `openspec/specs/technical-architecture/spec.md` — contains no legacy/migrate references
- `openspec/specs/config/spec.md` — contains no legacy/migrate references
- No `hf config migrate` command exists in any `cmd/` file, so nothing to delete there

## Testing Scope

No new tests required. `warnLegacy()` has no unit tests and is pure removal. Existing
`go test ./...` must continue to pass.
