# Workflow Specification

## Purpose

Provide end-to-end lifecycle workflow scripts that exercise the full cluster and nodepool create/patch/delete cycle. Two variants exist: one that drives real adapters (waiting for them to converge), and one that simulates adapter convergence by posting statuses manually via the CLI.

## Requirements

### Requirement: Real-Adapter Workflow

The CLI SHALL provide a workflow command that exercises the complete lifecycle against real adapters.

#### Scenario: Run full workflow with real adapters

- GIVEN api-url and api-version are configured and real adapters are running
- WHEN the user runs `hf workflow`
- THEN the CLI MUST execute the following ordered steps using a randomly generated name prefix (e.g., `test-<hex8>`):

  **Cluster lifecycle:**
  1. Create a cluster named `<prefix>`
  2. Display cluster table
  3. Wait for the cluster to reach `status.conditions[type=Ready].status == True` (timeout 120s, poll every 3s)
  4. Patch cluster labels; display cluster table
  5. Wait for `Ready=True` again (adapters reconverge after generation bump)
  6. Patch cluster spec; display cluster table
  7. Wait for `Ready=True` again

  **NodePool lifecycle:**
  8. Create a nodepool named `<prefix>` in the current cluster
  9. Display nodepool table
  10. Wait for the nodepool to reach `Ready=True`
  11. Patch nodepool labels; display nodepool table
  12. Wait for nodepool `Ready=True` again
  13. Patch nodepool spec; display nodepool table
  14. Wait for nodepool `Ready=True` again

  **Cluster deletion:**
  15. Delete the cluster; display cluster table
  16. Wait for cluster `Ready=True` (deletion acknowledged by adapters)

  **NodePool post-deletion:**
  17. Display nodepool table
  18. Wait for nodepool `Ready=True` (deletion acknowledged by adapters)

#### Scenario: Wait timeout

- GIVEN an adapter does not converge within the timeout period
- WHEN `hf workflow` is waiting for `Ready=True` or `Ready=False`
- THEN the CLI MUST print `[ERROR] Timed out after <N>s waiting for <resource> Ready=<expected> (current: <status>)`
- AND exit with a non-zero code

### Requirement: API-Only Workflow

The CLI SHALL provide a workflow command that simulates adapter convergence by posting statuses manually, without requiring real adapters.

#### Scenario: Run full workflow with simulated adapters

- GIVEN api-url and api-version are configured
- WHEN the user runs `hf workflow --api-only`
- THEN the CLI MUST execute the same lifecycle steps as the real-adapter workflow, **but** instead of waiting for adapters:
  - After each create or patch step the CLI MUST explicitly post adapter statuses via `hf cluster adapter post-status` / `hf nodepool adapter post-status` for each required adapter at the appropriate generation
  - For clusters: post status for `adapter1` and `adapter2`
  - For nodepools: post status for `adapter3`
  - Each step MUST print an expectation message (e.g., `>>> Expected: Ready=True`) after the table display

#### Scenario: Deterministic generation tracking

- GIVEN the workflow is in the api-only variant
- WHEN posting adapter statuses after a patch
- THEN the generation argument to `adapter post-status` MUST match the resource's current `generation` value after the patch
- AND the CLI MUST NOT re-use a previously posted generation — each patch increments the generation and adapters MUST report at the new generation

## Workflow Step Reference

| Step | Command used |
|------|-------------|
| Create cluster | `hf cluster create <name>` |
| Display cluster table | `hf cluster list --table` |
| Patch cluster labels | `hf cluster patch labels` |
| Patch cluster spec | `hf cluster patch spec` |
| Delete cluster | `hf cluster delete` |
| Post cluster adapter status | `hf cluster adapter post-status <adapter> <status> <gen>` |
| Create nodepool | `hf nodepool create <name>` |
| Display nodepool table | `hf nodepool list --table` |
| Patch nodepool labels | `hf nodepool patch labels` |
| Patch nodepool spec | `hf nodepool patch spec` |
| Post nodepool adapter status | `hf nodepool adapter post-status <adapter> <status> <gen>` |
