# Proposal: Colored JSON Output by Default

## Why

Every `hf` command prints JSON by default, but the output is plain monochrome text. Colorized
JSON is the standard UX expectation in modern CLIs — it dramatically improves readability at
a glance: keys stand out from values, types are immediately distinguishable, and nested
structure is easier to scan. The `--no-color` flag and `NO_COLOR` env var are already wired
across the CLI; color can be disabled without any friction for scripting or CI use.

## What Changes

- Add `internal/output/json.go` — a token-aware JSON colorizer using `encoding/json.Decoder`,
  with no new external dependencies. Color scheme:
  - Object keys: cyan
  - String values: green
  - Numbers: yellow
  - `true`: green, `false`: red
  - `null`: dim
  - Delimiters and punctuation: uncolored
- Add `colorCyan` and `colorDim` ANSI constants to `internal/output/dots.go` (reused by
  both dots and JSON colorizer)
- Add `isTerminal(io.Writer) bool` helper — colors are applied only when stdout is a TTY, so
  piping and file redirection automatically get plain output
- Modify `Printer.printJSON` to call the colorizer when color is enabled (not `--no-color`,
  not `NO_COLOR` env var, and writer is a TTY)
- Update `openspec/specs/output-formatting/spec.md` to document the colored JSON requirement
  and the conditions under which color is suppressed

## Scope Out

- Table and YAML output formats — not colorized in this change
- `fmt.Printf` calls in `version.go` — does not use `Printer`, out of scope
- Shell completion output — not user-facing data output

## Testing Scope

### internal/output (extended)

- `TestColorizeJSON_KeysStringsNumbersBoolsNull` — verify ANSI codes appear for each token type
- `TestColorizeJSON_NoColorFlag` — verify plain output when `noColor = true`
- `TestColorizeJSON_NoColorEnvVar` — verify plain output when `NO_COLOR` is set
- `TestColorizeJSON_NonTTYWriter` — verify plain output when writer is a `bytes.Buffer`
  (not a TTY)
- `TestColorizeJSON_Nested` — verify nested objects and arrays colorize correctly

No live cluster verification required — pure output logic.
