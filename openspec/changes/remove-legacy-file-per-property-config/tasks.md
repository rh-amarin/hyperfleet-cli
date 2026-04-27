# Tasks: Remove Legacy File-Per-Property Config Detection

## 1. Remove warnLegacy from internal/config/store.go

- [ ] 1.1 Delete the `// Warn about legacy file-per-property layout.` comment and the
       `s.warnLegacy()` call inside `NewStore`
- [ ] 1.2 Delete the `warnLegacy()` method at the bottom of the file
- [ ] 1.3 Remove `"fmt"` from the import block (no longer used after 1.2)

## 2. Update openspec/specs/config-model/spec.md

- [ ] 2.1 Trim the Purpose paragraph: remove "while maintaining backwards compatibility
       during migration"
- [ ] 2.2 Delete the entire "### Requirement: Migration from File-Per-Property" section
       (including both scenarios: Auto-detect legacy config and No legacy config)

## 3. Update openspec/specs/config-registry/spec.md

- [ ] 3.1 Delete the "(Previously: file-per-property storage at `~/.config/hf/<key>`.
       Superseded by config-model/spec.md.)" line

## 4. Verify

- [ ] 4.1 `go build ./...` succeeds
- [ ] 4.2 `go vet ./...` reports no issues
- [ ] 4.3 `go test ./...` passes — capture full output and save to
       `verification_proof/tests.txt`
