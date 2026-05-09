# Proposal: Fix search no-args state routing and id WARN exit-0

## What

Fix two command behaviour mismatches in `hf cluster` and `hf nodepool`:

1. **`search` with no arguments** — currently fails with a cobra arity error because
   `Args: cobra.ExactArgs(1)` is set. The spec requires state-aware routing:
   - If an ID is already stored in state → behave identically to `get`
   - If no ID in state → print `[ERROR] No <resource>-id set in state...` and exit 1

2. **`id` when no ID is set in state** — currently exits 1 with `Error: no <resource>-id set ...`
   from cobra's error printer. The spec requires a `[WARN]` to stderr and exit 0.

## Why

These are documented spec mismatches in:
- `openspec/specs/cluster-lifecycle/spec.md` — "Search with no arguments" scenarios
- `openspec/specs/nodepool-lifecycle/spec.md` — "Search with no arguments" and prerequisites

Fixing them makes the CLI consistent with the bash scripts it replaces, which always
checked state rather than failing on missing arguments.

## Scope

- `cmd/cluster.go` — `clusterSearchCmd`, `clusterIDCmd`
- `cmd/nodepool.go` — `nodepoolSearchCmd`, `nodepoolIDCmd`
- `cmd/cluster_test.go` — new tests for the two behaviours
- `cmd/nodepool_test.go` — new tests for the two behaviours
- No new packages, no API changes, no spec delta required
