# Design: Fix Table Column Alignment

## Root Cause

`text/tabwriter` aligns columns by counting runes in each cell. ANSI terminal color
codes (e.g., `\x1b[32m`, `\x1b[0m`) are invisible on screen but are counted as
real runes:

| String | Rune count | Visual width |
|---|---|---|
| `\x1b[32m●\x1b[0m` | 10 | 1 |
| `AVAILABLE` | 9 | 9 |

When a data cell (colored dot, 10 runes) is wider than the header (9 runes), tabwriter
pads subsequent columns with phantom whitespace — the `Available` cell is treated as
10 chars wide but only 1 char wide on screen, so every following column shifts right.

> Note: Go's `tabwriter.Escape` (`\xff`) mechanism was considered but ruled out.
> `string(tabwriter.Escape)` produces the 2-byte UTF-8 sequence for U+00FF (`\xc3\xbf`),
> not the literal `\xff` byte tabwriter actually watches for — a subtle trap.
> Additionally, the escape-pair mechanism only prevents tab/newline interpretation
> inside the sequence; it does not make the enclosed bytes zero-width for column
> measurement, so it cannot fix the alignment bug.

## Fix: ANSI-Aware Column Width in `PrintTable`

Replace the `tabwriter`-based rendering with a manual approach in
`internal/output/table.go`:

1. **Compute visible widths**: strip ANSI codes (`\x1b[...m`) from every cell using
   a compiled regex, then count runes with `unicode/utf8`.
2. **Track per-column max**: iterate headers and all data rows to find the maximum
   visible width per column.
3. **Write with explicit padding**: for each row, write the cell content (ANSI codes
   intact), then append `(col_width - visible_width + 2)` space characters. Skip
   trailing padding on the last column.

```go
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleRuneCount(s string) int {
    return utf8.RuneCountInString(ansiEscape.ReplaceAllString(s, ""))
}

func (p *Printer) PrintTable(headers []string, rows [][]string) error {
    // ... compute widths from visibleRuneCount ...
    // ... write each row, padding by (width - visibleRuneCount(cell) + 2) ...
}
```

The dot strings in `dots.go` are unchanged; the fix lives entirely in `PrintTable`.

## Spec Update: `openspec/specs/output-formatting/spec.md`

Add a new scenario under "Requirement: Multi-Format Output Dispatch":

```
#### Scenario: Colored dot cells align with plain-text header cells

- GIVEN `--output table` is set and color output is enabled
- WHEN a command calls `Printer.PrintTable` with headers and rows that contain
  colored dot characters
- THEN each data column MUST start at the same horizontal position as the
  corresponding header column
- AND ANSI color escape sequences MUST NOT contribute to the computed column width
```

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Manual padding vs tabwriter | Manual | tabwriter has no reliable way to make ANSI codes zero-width; the escape mechanism is for tab/newline isolation, not width override |
| Regex compiled at package level | Yes | Avoids recompilation on every `PrintTable` call |
| Dots unchanged | Yes | The fix is entirely in the width-measurement layer — callers need not change |
| Trailing column padding | None | Matches tabwriter convention; avoids trailing spaces on last column |

## Unit Test Plan

### `internal/output/output_test.go` (existing file)

| Test | Setup | Assertion |
|---|---|---|
| `TestPrintTable_AlignedWithColoredDots` | `Printer{noColor: false}`; headers `["col1", "col2"]`; row `[dot("True"), dot("True")]` | Strip ANSI from data line; assert `[]rune(stripped)[col2HeaderIdx]` == `●`, i.e. second column starts at the same rune offset as `"COL2"` in the ASCII header |
