# Phase 13 Tasks — Non-Functional Requirements

## 1. Shell Completions

- [x] 1.1 Create `cmd/completion.go` with bash/zsh/fish/powershell support
- [x] 1.2 Create `cmd/completion_test.go` — test all four shells + unknown shell error

## 2. Plugin Architecture

- [x] 2.1 Create `internal/plugin/plugin.go` — `Find` and `Exec` functions
- [x] 2.2 Create `internal/plugin/plugin_test.go` — PATH override tests, temp binary
- [x] 2.3 Create `cmd/plugin_dispatch.go` — wire plugin dispatch into root command
- [x] 2.4 Create `cmd/plugin_dispatch_test.go` — verify unknown cmd delegates to plugin

## 3. GoReleaser

- [x] 3.1 Add `release:` section to `.goreleaser.yaml`
- [x] 3.2 Add `changelog.use: git` for cleaner release notes

## 4. GitHub Actions

- [x] 4.1 Create `.github/workflows/ci.yml`
- [x] 4.2 Create `.github/workflows/release.yml`

## 5. Error & Usage Polish

- [x] 5.1 Audit all commands for `SilenceUsage: true`
- [x] 5.2 Verify `Warn`/`Info`/`Errorf` go to stderr with correct prefixes (already correct)

## 6. Integration Tests

- [x] 6.1 Create `cmd/integration_test.go` with `//go:build integration` tag

## Verify

- [x] (a) `go build ./...` succeeds
- [x] (b) `go vet ./...` no issues
- [x] (c) `go test ./...` passes → `verification_proof/tests.txt`
- [x] (d) `hf completion bash` outputs valid bash completion → `verification_proof/completion-bash.txt`
- [x] (e) Plugin scan: create a dummy `hf-hello` script, run `hf hello` → `verification_proof/plugin-hello.txt`
