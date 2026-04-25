# Design: Config UX Simplification

## Context

The phase-01 config system implemented a precedence chain (defaults → config.yaml → env
profile → HF_* env vars) and annotated each value with its source. This was accurate but
surfaced implementation complexity to the user. The real mental model users need is simpler:
pick an environment, work in it. HF_* overrides undermined that model by creating a hidden
configuration channel that was not visible in `hf config show` without source annotations.

## Goals / Non-Goals

**Goals**
- Users can create a named environment in one command
- `hf config show` answers "what is my current config?" without noise
- Any command that needs config fails immediately when none is configured
- Config has exactly one source: the active environment file

**Non-Goals**
- Changing the on-disk storage format
- Adding config editing beyond the existing `hf config set`

## Decisions

### 1. `hf config env new` replaces `hf config bootstrap`

`bootstrap` was confusing — users are creating environments, not bootstrapping the system.
The new command lives under `hf config env` alongside `list`, `activate`, `deactivate`,
`show`, making the namespace coherent.

If `name` is not passed as a positional arg, the command prompts for it before any other
prompt. This keeps the command scriptable (`hf config env new prod`) while staying
interactive when invoked bare.

### 2. HF_* config value env vars removed

`HF_API_URL`, `HF_API_VERSION`, `HF_TOKEN`, `HF_CONTEXT`, `HF_NAMESPACE` are removed.
They created a hidden override channel that bypassed the active environment file, making
`hf config show` untrustworthy unless you also checked your shell environment. With a single
source of truth (the active env file), there is nothing to override.

`HF_CONFIG_DIR` is retained — it controls where the config directory lives, not a config
value, so it belongs at the infrastructure layer.

Concretely: `applyEnvVars()` in `store.go` and the `envVarFor` map in `resolve.go` are
deleted. `OverrideCfg` remains because it backs the `--api-url` / `--api-token` CLI flags,
which are explicit, visible, and intentional overrides for scripted use.

### 3. Source annotations removed from `hf config show`

The source column (`[default]`, `[config]`, `[env:kind]`) was designed to help debug the
precedence chain. With HF_* overrides gone and the active-environment guard in place, the
resolved value always comes from the active env file. Showing provenance adds noise without
answering a user question. The active environment name shown at the top provides sufficient
context.

### 4. Active environment guard as a shared helper

A single `requireActiveEnv(s *config.Store) error` function returns the guidance error.
Each gated command calls it at the top of its `RunE`. This keeps the guard logic in one
place. Commands that manage environments themselves (`env list`, `env new`, `env activate`,
`env deactivate`) and `clear state` are explicitly excluded.

### 5. Prompt defaults updated to match real deployments

The bootstrap defaults were inherited from the old bash scripts and assumed localhost
services. Real HyperFleet deployments use in-cluster service names and a standard test
credential set. Updating the defaults reduces copy-paste errors during setup.

| Field | Old default | New default |
|---|---|---|
| database.user | (empty) | hyperfleet |
| database.name | (empty) | hyperfleet |
| database.password | (empty) | foobar-bizz-buzz |
| rabbitmq.host | localhost | rabbitmq |
| rabbitmq.user | guest | guest (unchanged) |
| rabbitmq.password | (empty) | guest |

## Risks / Trade-offs

- Removing HF_* overrides breaks any workflow that set `HF_TOKEN` or `HF_API_URL` in CI to
  point at a specific cluster. Mitigation: create a named env file for each target and
  activate it; `--api-url` / `--api-token` CLI flags remain available for one-off scripted
  overrides.
- The active-env guard will break any script that called `hf config show` without setting
  an environment. Mitigation: the error message is actionable and the fix is one command.

## Migration Plan

No data migration. Users with existing `environments/*.yaml` files are unaffected.

- `hf config bootstrap` → `hf config env new` (same prompts, new name)
- `HF_TOKEN=x hf cluster list` → `hf --api-token x cluster list` or create a dedicated env
  file and activate it

## Open Questions

None.
