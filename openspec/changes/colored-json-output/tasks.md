# Tasks: Colored JSON Output by Default

## 1. Add ANSI constants to internal/output/dots.go

- [x] 1.1 Add `colorCyan = "\033[36m"` and `colorDim = "\033[2m"` to the const block

## 2. Create internal/output/json.go

- [x] 2.1 Create `internal/output/json.go` with `isTerminal`, `colorizeJSON`, `writeValue`,
       `writeObject`, and `writeArray` as specified in design.md

## 3. Modify internal/output/printer.go

- [x] 3.1 Update `printJSON` to call `colorizeJSON` when `!p.noColor`, `NO_COLOR == ""`,
       and `isTerminal(p.writer)` are all true

## 4. Unit Tests

- [x] 4.1 Add `TestColorizeJSON_KeysStringsNumbersBoolsNull` — assert that ANSI codes for
       cyan, green, yellow, red, and dim appear in the colorized output for each token type
- [x] 4.2 Add `TestColorizeJSON_Nested` — assert that nested objects and arrays round-trip
       correctly (strip ANSI codes and compare to original indented JSON)
- [x] 4.3 Add `TestPrinter_JSONColor_NonTTY` — create a Printer with a `bytes.Buffer` writer
       (non-TTY), assert output contains no ANSI escape codes
- [x] 4.4 Add `TestPrinter_JSONColor_NoColorFlag` — create a Printer with `noColor=true`,
       assert output contains no ANSI escape codes
- [x] 4.5 Add `TestPrinter_JSONColor_NoColorEnv` — set `NO_COLOR=1`, assert output contains
       no ANSI escape codes (writer can be anything)

## 5. Update openspec/specs/output-formatting/spec.md

- [x] 5.1 Add the "Colored JSON Output" requirement with all four scenarios as specified
       in design.md

## 6. Verify

- [x] 6.1 `go build ./...` succeeds
- [x] 6.2 `go vet ./...` reports no issues
- [x] 6.3 `go test ./...` passes — output saved to `verification_proof/tests.txt`
