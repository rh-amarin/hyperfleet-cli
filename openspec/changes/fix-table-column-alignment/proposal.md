# Proposal: Fix Table Column Alignment

## Why

Table commands (`hf cluster table`, `hf nodepool table`, `hf table`) display status
columns that contain ANSI color-coded dot characters (e.g., `\x1b[32m●\x1b[0m`).
Go's `text/tabwriter` calculates column widths by counting runes — ANSI escape sequences
are invisible on the terminal but contribute extra runes per colored dot, so tabwriter
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

- **`internal/output/table.go`** — replace `tabwriter`-based rendering with a
  manual approach that strips ANSI codes before measuring column widths, then
  pads each cell to the correct visible width. No changes to the ANSI dot strings
  or any other files.
- **`openspec/specs/output-formatting/spec.md`** — add a requirement scenario specifying
  that colored dot cells MUST be aligned correctly when mixed with plain-text header cells.

## Scope Out

- No changes to `dots.go`, command files (`cmd/`), resource types, API client, or watch logic.
- No new third-party dependencies.

## Testing Scope

### `internal/output` (existing package — `output_test.go`)

- `TestPrintTable_AlignedWithColoredDots`: create a `Printer` with `noColor=false`;
  call `PrintTable` with a header row and a data row containing colored dots; strip
  ANSI codes from the data line and assert the second dot starts at the same rune
  offset as `"COL2"` in the (pure-ASCII) header, proving no phantom padding.

## Which Steps Require Live Cluster

Step (d) live verification requires the HyperFleet API at `34.175.27.173:8000`.
