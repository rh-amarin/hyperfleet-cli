# Proposal: Fix Table Column Alignment

## Why

Table commands (`hf cluster table`, `hf nodepool table`, `hf table`) display status
columns that contain ANSI color-coded dot characters (e.g., `\x1b[32m●\x1b[0m`).
Go's `text/tabwriter` calculates column widths by counting runes — ANSI escape sequences
are invisible on the terminal but contribute 9 extra runes per colored dot, so tabwriter
over-pads every column that follows a dot, producing misaligned output like:

```
NAME         GEN  AVAILABLE           READY
my-cluster   1    ●                   ●
```

instead of the correct:

```
NAME         GEN  AVAILABLE  READY
my-cluster   1    ●          ●
```

## What Changes

- **`internal/output/dots.go`** — wrap each ANSI escape sequence in `tabwriter.Escape`
  (`\xff`) delimiter pairs so that tabwriter treats those bytes as zero-width. No change
  to visible terminal output.
- **`internal/output/table.go`** — add `tabwriter.StripEscape` flag to the `tabwriter.Writer`
  so the `\xff` delimiters are removed from output (only alignment metadata, never rendered).
- **`openspec/specs/output-formatting/spec.md`** — add a requirement scenario specifying
  that colored dot cells MUST be aligned correctly when mixed with plain-text header cells.

## Scope Out

- No changes to command files (`cmd/`), resource types, API client, or watch logic.
- No changes to `--no-color` / `NO_COLOR` paths — those return plain text and are
  unaffected by tabwriter.
- No new third-party dependencies.

## Testing Scope

### `internal/output` (existing package — `output_test.go`)

- `TestPrintTable_AlignedWithColoredDots`: create a `Printer` with `noColor=false`;
  call `PrintTable` with a header row and one data row containing colored dots; assert
  that each column in the output starts at the same byte offset as the same column in
  the header row, confirming ANSI codes do not shift alignment.
- `TestDot_TabwriterEscaped`: assert that `Dot("True")`, `Dot("False")`, and
  `Dot("Unknown")` return strings that begin and end with `\xff` (tabwriter.Escape),
  sandwiching the ANSI codes, so callers outside of table rendering are not affected.

## Which Steps Require Live Cluster

Step (d) live verification requires the HyperFleet API at `34.175.27.173:8000`.
