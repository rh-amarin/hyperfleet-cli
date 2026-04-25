# Design: Config UX Simplification

## Context

The phase-01 config system implemented a precedence chain (defaults → config.yaml → env
profile → HF_* env vars) and annotated each value with its source. This was accurate but
surfaced implementation complexity to the user. The real mental model users need is simpler:
pick an environment, work in it.

## Goals / Non-Goals

**Goals**
- Users can create a named environment in one command
- `hf config show` answers "what is my current config?" without noise
- Any command that needs config fails immediately when none is configured

**Non-Goals**
- Changing the on-disk storage format
- Removing HF_* env var override support
- Adding config editing (that stays as `hf config set`)

## Decisions

### 1. `hf config env new` replaces `hf config bootstrap`

`bootstrap` was confusing — users are creating environments, not bootstrapping the system.
The new command lives under `hf config env` alongside `list`, `activate`, `deactivate`,
`show`, making the namespace coherent.

If `name` is not passed as a positional arg, the command prompts for it before any other
prompt. This keeps the command scriptable (`hf config env new prod`) while staying
interactive when invoked bare.

### 2. Source annotations removed from `hf config show`

The source column (`[default]`, `[config]`, `[env:kind]`) was designed to help debug the
precedence chain. With the active-environment guard in place, the resolved value always comes
from the active env profile (or a HF_* override). Showing provenance adds noise without
answering a user question. The active environment name shown at the top provides sufficient
context.

### 3. Active environment guard as a shared helper

A single `requireActiveEnv(s *config.Store) error` function returns the guidance error.
Each gated command calls it at the top of its `RunE`. This keeps the guard logic in one
place. Commands that manage environments themselves (`env list`, `env new`, `env activate`,
`env deactivate`) and `clear state` are explicitly excluded.

### 4. Prompt defaults updated to match real deployments

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

- Removing source annotations makes it harder to debug HF_* override conflicts. Mitigation:
  users can inspect env var values directly; a future `hf config debug` command could
  restore this if needed.
- The active-env guard will break any script that called `hf config show` without setting
  an environment. Mitigation: the error message is actionable and the migration is one
  command.

## Migration Plan

No data migration. Users with existing `environments/*.yaml` files are unaffected. Users
relying on `hf config bootstrap` should use `hf config env new` instead — same prompts,
new name.

## Open Questions

None.
