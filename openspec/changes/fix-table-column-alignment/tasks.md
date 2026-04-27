# Tasks: Fix Table Column Alignment

## 1. Fix PrintTable in table.go

- [x] 1.1 Add `regexp` and `unicode/utf8` imports to `internal/output/table.go`.
- [x] 1.2 Add package-level `ansiEscape` regex (`\x1b\[[0-9;]*m`) and `visibleRuneCount`
  helper function.
- [x] 1.3 Rewrite `PrintTable` to:
  - Compute per-column max visible widths from headers and all data rows using `visibleRuneCount`.
  - Write each row cell followed by `(col_width - visibleRuneCount(cell) + 2)` spaces,
    omitting trailing padding for the last column.
  - Remove the `text/tabwriter` import (no longer needed).

## 2. Update the Output Formatting Spec

- [x] 2.1 In `openspec/specs/output-formatting/spec.md`, add a new scenario under
  "Requirement: Multi-Format Output Dispatch":

  ```
  #### Scenario: Colored dot cells align with plain-text header cells

  - GIVEN `--output table` is set and color output is enabled
  - WHEN a command calls `Printer.PrintTable` with headers and rows that contain
    colored dot characters
  - THEN each data column MUST start at the same horizontal position as the
    corresponding header column
  - AND ANSI color escape sequences MUST NOT contribute to the computed column width
  ```

## 3. Update Unit Tests

- [x] 3.1 In `internal/output/output_test.go`, add `TestPrintTable_AlignedWithColoredDots`:
  - `Printer{noColor: false}` writing to a `bytes.Buffer`.
  - `PrintTable(["col1", "col2"], [[dot("True", false), dot("True", false)]])`.
  - Split output; strip ANSI from data line; assert second dot starts at rune offset
    matching `strings.Index(header, "COL2")`.

## Verify

- [x] (a) `go build ./...` succeeds
- [x] (b) `go vet ./...` reports no issues
- [x] (c) `go test ./...` passes — output saved to `verification_proof/tests.txt`
- [x] (d) Live verification against real cluster (`34.175.27.173:8000`):
  - [x] d.1 `hf cluster table` — saved to `verification_proof/d1-cluster-table.txt`
  - [x] d.2 `hf nodepool table` — saved to `verification_proof/d2-nodepool-table.txt`
  - [x] d.3 `hf table` — saved to `verification_proof/d3-table.txt`
