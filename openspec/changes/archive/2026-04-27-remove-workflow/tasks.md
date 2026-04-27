# Tasks: Remove hf workflow Command

## 1. Delete Workflow Shell Scripts

- [ ] 1.1 Delete `scripts/hf.workflow.sh`
- [ ] 1.2 Delete `scripts/hf.workflow.api-only.sh`

## 2. Update Spec Index

- [ ] 2.1 `openspec/specs/index.md` — row 12 "Scripts Covered": remove workflow scripts, keep `watch/interactive modes`
- [ ] 2.2 `openspec/specs/index.md` — row 12 counts: req 8→7, scenarios 16→14
- [ ] 2.3 `openspec/specs/index.md` — Summary totals: req 77→75, scenarios 143→141, total req 104→102, total scenarios 211→209

## 3. Update Interactive Commands Spec

- [ ] 3.1 `openspec/specs/interactive-commands/spec.md` — remove entire "Requirement: Long-Running Workflows" section (both scenarios)

## 4. Update Technical Architecture Spec

- [ ] 4.1 `openspec/specs/technical-architecture/spec.md` — remove `workflow.go` line from module layout
- [ ] 4.2 `openspec/specs/technical-architecture/spec.md` — remove `workflow / run / api-only` subtree from Cobra command tree

## 5. Update Script Inventory Files

- [ ] 5.1 `scripts/README.md` — remove two workflow script rows from the table
- [ ] 5.2 `scripts/output/00-index.json` — remove `"hf.workflow.sh"` and `"hf.workflow.api-only.sh"` from `"utility"` array
- [ ] 5.3 `scripts/output/12-interactive-commands.json` — remove workflow script entries

## Verify

- [ ] (a) `go build ./...` succeeds
- [ ] (b) `go vet ./...` no issues
- [ ] (c) `go test ./...` passes
- [ ] (d) `grep -r "hf workflow" openspec/ scripts/output/` returns no matches
