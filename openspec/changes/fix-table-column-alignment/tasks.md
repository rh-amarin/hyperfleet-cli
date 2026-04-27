# Tasks: Fix Table Column Alignment

## 1. Fix ANSI Zero-Width Escaping in dots.go

- [ ] 1.1 In `internal/output/dots.go`, import `"text/tabwriter"` and define a local
  constant `esc = string(tabwriter.Escape)`.
- [ ] 1.2 Wrap every ANSI color sequence in `esc` delimiter pairs inside the `dot()`
  function (colored branch only):
  - `"True"` → `esc + colorGreen + esc + dotChar + esc + colorReset + esc`
  - `"False"` → `esc + colorRed + esc + dotChar + esc + colorReset + esc`
  - `"Unknown"` → `esc + colorYellow + esc + dotChar + esc + colorReset + esc`
  - Default / absent → `"-"` (unchanged)
  - No-color branch (`noColor=true`) → unchanged (returns plain ASCII text)

## 2. Enable StripEscape in table.go

- [ ] 2.1 In `internal/output/table.go`, change the `tabwriter.NewWriter` call to pass
  the `tabwriter.StripEscape` flag so the `\xff` delimiters are removed from final output:
  ```go
  tw := tabwriter.NewWriter(p.writer, 0, 0, 2, ' ', tabwriter.StripEscape)
  ```

## 3. Update the Output Formatting Spec

- [ ] 3.1 In `openspec/specs/output-formatting/spec.md`, add a new scenario under
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

## 4. Update Unit Tests

- [ ] 4.1 In `internal/output/output_test.go`, update `TestDotRendering_Color` to
  expect `\xff`-wrapped output now that `dot()` emits escape delimiters. Change the
  `want` values to match the new format:
  ```go
  esc := string(tabwriter.Escape)
  {"True",    esc + colorGreen  + esc + dotChar + esc + colorReset + esc},
  {"False",   esc + colorRed    + esc + dotChar + esc + colorReset + esc},
  {"Unknown", esc + colorYellow + esc + dotChar + esc + colorReset + esc},
  ```
- [ ] 4.2 Add `TestPrintTable_AlignedWithColoredDots` to `internal/output/output_test.go`:
  - Create a `Printer` with `noColor=false` writing to a `bytes.Buffer`.
  - Call `PrintTable(["COL1", "COL2"], [[dot("True", false), dot("True", false)]])`.
  - Split output into lines; strip all ANSI escape sequences from both lines (pattern
    `\x1b\[[0-9;]*m` and `\xff`).
  - Assert the byte offset of "COL2" in the stripped header matches the byte offset of
    the second column in the stripped data line, confirming no phantom padding.
- [ ] 4.3 Add `TestDot_TabwriterEscaped` to `internal/output/output_test.go`:
  - For each of `dot("True", false)`, `dot("False", false)`, `dot("Unknown", false)`:
    assert the returned string starts with `\xff` (tabwriter.Escape byte) and ends
    with `\xff`.

## Verify

- [ ] (a) `go build ./...` succeeds
- [ ] (b) `go vet ./...` reports no issues
- [ ] (c) `go test ./...` passes — capture full output and save to
  `verification_proof/tests.txt`
- [ ] (d) Live verification against the real cluster — capture output per command:
  - [ ] d.1 `hf cluster table` renders correctly aligned cluster table — save to
    `verification_proof/d1-cluster-table.txt`
  - [ ] d.2 `hf nodepool table` renders correctly aligned nodepool table — save to
    `verification_proof/d2-nodepool-table.txt`
  - [ ] d.3 `hf table` renders correctly aligned combined table — save to
    `verification_proof/d3-table.txt`
