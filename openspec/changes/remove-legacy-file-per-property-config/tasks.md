# Tasks: Remove Legacy File-Per-Property Config Detection

## 1. Remove warnLegacy from internal/config/store.go

- [x] 1.1 Delete the `// Warn about legacy file-per-property layout.` comment and the
       `s.warnLegacy()` call inside `NewStore`
- [x] 1.2 Delete the `warnLegacy()` method at the bottom of the file
- [x] 1.3 `"fmt"` kept — still used by `fmt.Errorf` calls throughout `store.go` (design note was incorrect)

## 2. Update openspec/specs/config-model/spec.md

- [x] 2.1 Trim the Purpose paragraph: remove "while maintaining backwards compatibility
       during migration"
- [x] 2.2 Delete the entire "### Requirement: Migration from File-Per-Property" section
       (including both scenarios: Auto-detect legacy config and No legacy config)

## 3. Update openspec/specs/config-registry/spec.md

- [x] 3.1 Delete the "(Previously: file-per-property storage at `~/.config/hf/<key>`.
       Superseded by config-model/spec.md.)" line

## 4. Verify

- [x] 4.1 `go build ./...` succeeds
- [x] 4.2 `go vet ./...` reports no issues
- [x] 4.3 `go test ./...` passes — output saved to `verification_proof/tests.txt`
