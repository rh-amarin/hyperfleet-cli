# Tasks: Phase 00 — Bootstrap

## 1. Module Setup
- [x] 1.1 go mod init github.com/rh-amarin/hyperfleet-cli
- [x] 1.2 Add cobra dependency to go.mod
- [x] 1.3 Create main.go

## 2. Root Command
- [x] 2.1 cmd/root.go with all persistent flags
- [x] 2.2 internal/version/version.go with Version/Commit/Date vars
- [x] 2.3 cmd/version.go hf version subcommand

## 3. Build Tooling
- [x] 3.1 Makefile with build/install/lint/test targets
- [x] 3.2 .goreleaser.yaml for cross-compilation

## 4. Verify
- [x] 4.1 go build ./... succeeds with no errors
- [x] 4.2 hf version prints version string
- [x] 4.3 hf --help shows usage and flag list
