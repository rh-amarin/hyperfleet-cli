# Design: Remove hf workflow Command

## Files Deleted

- `scripts/hf.workflow.sh`
- `scripts/hf.workflow.api-only.sh`

## openspec/specs/index.md

Row 12 "Scripts Covered" column: remove `hf.workflow.sh, hf.workflow.api-only.sh`; replace with
`watch/interactive modes` only.

Update the row counts for domain 12 (Interactive Commands):
- Requirements: 8 → 7 (remove "Long-Running Workflows")
- Scenarios: 16 → 14 (remove 2 workflow scenarios)

Update the Summary table totals:
- Functional (01–13) Requirements: 77 → 75
- Functional (01–13) Scenarios: 143 → 141
- Total Requirements: 104 → 102
- Total Scenarios: 211 → 209

## openspec/specs/interactive-commands/spec.md

Remove the entire "Requirement: Long-Running Workflows" section:

```
### Requirement: Long-Running Workflows
...
#### Scenario: API-only workflow
...
```

(Two scenarios, one requirement block — lines 120–136 in the original file.)

## openspec/specs/technical-architecture/spec.md

**Module layout** — remove this line from the `cmd/` listing:

```
│   └── workflow.go         # hf workflow [run|api-only]
```

Replace `└── workflow.go` (the last entry) by promoting `table.go` to the last position, or
keeping `repos.go` as the new last entry under `cmd/`.

**Cobra command tree** — remove the three-line `workflow` subtree:

```
├── workflow
│   ├── run
│   └── api-only
```

## scripts/README.md

Remove two rows from the script inventory table:

```
| `hf.workflow.sh`          | End-to-end test workflow: ...      |
| `hf.workflow.api-only.sh` | API-only workflow: ...             |
```

## scripts/output/00-index.json

Remove `"hf.workflow.sh"` and `"hf.workflow.api-only.sh"` from the `"utility"` array.

## scripts/output/12-interactive-commands.json

Remove the two script entries for `hf.workflow.sh` (type `"long-running"`) and
`hf.workflow.api-only.sh` (type `"template"`).
