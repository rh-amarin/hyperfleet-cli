# Proposal: Remove hf workflow Command

## Intent

Remove the `hf workflow` command and all associated spec and script references. The `hf workflow`
lifecycle sequences (`run` and `api-only`) are superseded by direct composition of individual
`hf` commands; there is no plan to implement `cmd/workflow.go`.

## Scope In

- Delete `scripts/hf.workflow.sh` and `scripts/hf.workflow.api-only.sh`
- Remove "Long-Running Workflows" requirement and its two scenarios from
  `openspec/specs/interactive-commands/spec.md`
- Remove `workflow.go` from the module layout and `workflow [run|api-only]` from the Cobra
  command tree in `openspec/specs/technical-architecture/spec.md`
- Remove row 12 workflow scripts from `openspec/specs/index.md`; update requirement and scenario
  counts accordingly
- Remove workflow script entries from `scripts/output/00-index.json` and
  `scripts/output/12-interactive-commands.json`
- Remove workflow script rows from `scripts/README.md`

## Scope Out

- GitHub Actions CI/CD "workflow" terminology in `openspec/specs/non-functional/spec.md` — those
  refer to standard GitHub Actions pipelines, not the `hf workflow` command
- `internal/repos/client.go` `latestCIStatus()` and `cmd/repos_test.go` `workflowRunJSON()` —
  these use GitHub's API terminology for CI runs and power `hf repos`; they are unrelated to
  `hf workflow`
- Archived change directories under `openspec/changes/archive/` — historical record, not updated

## Testing Scope

No new tests required. This is a pure removal; existing tests continue to pass.
