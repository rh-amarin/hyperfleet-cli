# Design: Fix Table Column Alignment

## Root Cause

`text/tabwriter` aligns columns by counting the rune width of each cell's content.
ANSI terminal color codes (`\x1b[32m`, `\x1b[0m`, etc.) are invisible on screen but
are counted as real runes:

| String | Rune count | Visual width |
|---|---|---|
| `\x1b[32m●\x1b[0m` | 10 | 1 |
| `AVAILABLE` | 9 | 9 |

When a data cell (colored dot, 10 runes) is wider than the header (9 runes), tabwriter
pads subsequent columns with 9 extra spaces of phantom width, pushing all following
columns to the right on every row that contains a dot.

## Fix: tabwriter Escape Mechanism

Go's `text/tabwriter` provides a built-in zero-width escape mechanism:

> Content between two `tabwriter.Escape` bytes (`\xff`) is passed through to output
> unchanged but contributes **zero width** to column alignment.

With the `tabwriter.StripEscape` flag, the `\xff` delimiters are removed from the
final output (they are only used internally for width accounting).

### `internal/output/dots.go`

Change the colored-dot return values to wrap every ANSI sequence in `\xff` pairs:

```go
import "text/tabwriter"

const esc = string(tabwriter.Escape) // "\xff"

// Before:
return colorGreen + dotChar + colorReset

// After:
return esc + colorGreen + esc + dotChar + esc + colorReset + esc
```

The result is that:
- `tabwriter` sees: `<esc><5 bytes><esc>●<esc><4 bytes><esc>`
- Width counted: 1 rune (just `●`, which is a single rune in UTF-8)
- Output written: `\x1b[32m●\x1b[0m` (identical to before; `\xff` stripped)

The `dot` function receives `noColor bool`; the escaping is only needed for the
colored branch. No-color returns plain ASCII text and needs no change.

### `internal/output/table.go`

Add `tabwriter.StripEscape` to the writer flags:

```go
// Before:
tw := tabwriter.NewWriter(p.writer, 0, 0, 2, ' ', 0)

// After:
tw := tabwriter.NewWriter(p.writer, 0, 0, 2, ' ', tabwriter.StripEscape)
```

Without this flag, the `\xff` bytes would appear in terminal output as garbage
characters. With `StripEscape` they are silently removed after alignment is computed.

## Spec Update: `openspec/specs/output-formatting/spec.md`

Add a new scenario under "Requirement: Multi-Format Output Dispatch" to explicitly
capture the alignment invariant:

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
| Use `tabwriter.Escape` | Yes | Stdlib mechanism; no new dependencies; transparent to callers |
| Escape whole ANSI sequences or just boundaries | Each ANSI code individually | Matches tabwriter contract: only the escape payload is zero-width, visible content (`●`) must still be counted |
| Change `Dot()` package-level function | Yes, same as `dot()` | Both functions delegate to `dot()`; fix is in `dot()` only |
| Update spec | Yes, add scenario | Makes the alignment invariant explicit and testable |

## Unit Test Plan

### `internal/output/output_test.go` (existing file)

| Test | Setup | Assertion |
|---|---|---|
| `TestPrintTable_AlignedWithColoredDots` | `Printer{noColor: false}`; headers `["COL1", "COL2"]`; row `[dot("True"), dot("True")]` | Split output into lines; verify column 2 starts at the same byte offset in both the header line and the data line |
| `TestDot_TabwriterEscaped` | Call `dot("True", false)`, `dot("False", false)`, `dot("Unknown", false)` | Each returned string begins with `\xff` and ends with `\xff` |

The existing `TestDot_*` and `TestPrintTable_*` tests continue to pass unchanged.
