# Phase 13 Proposal — Non-Functional Requirements

## Why

All functional phases (00–12) deliver the core HyperFleet commands. Phase 13 closes the remaining engineering quality gaps before the CLI is production-ready: shell completion for discoverability, a plugin system for extensibility, a release pipeline for distribution, and polished error output for consistency.

## Scope In

- `hf completion bash|zsh|fish|powershell` — Cobra-generated shell completions
- `internal/plugin` — PATH-based plugin discovery (`hf-<name>` executables)
- Plugin dispatch wired into the root command (unknown subcommands delegate to plugins)
- `.github/workflows/ci.yml` — build/vet/test on every PR and push to main
- `.github/workflows/release.yml` — GoReleaser on tag push
- `.goreleaser.yaml` — finalized cross-platform build config with changelog and checksums
- `cmd/integration_test.go` — integration test skeleton with `//go:build integration` tag
- Error/usage polish: verify `SilenceUsage: true` on all commands; API errors render in `--output` format

## Scope Out

- Homebrew tap (not in plan)
- deb/rpm packaging (not in plan)
- `cmd/workflow.go` (phase-12 responsibility)

## Testing Scope

| Package | Test cases |
|---|---|
| `cmd` (completion) | bash/zsh/fish/powershell output non-empty, exits 0; unknown shell errors |
| `internal/plugin` | Find returns true for PATH-visible binary; Find returns false when absent; Exec forwards args |
| `cmd` (plugin_dispatch) | Unknown subcommand with plugin on PATH delegates; unknown without plugin returns error |

## Verification Steps

- (a) `go build ./...` — no cluster needed
- (b) `go vet ./...` — no cluster needed
- (c) `go test ./...` — no cluster needed
- (d) `hf completion bash` — no cluster needed
- (e) Plugin scan with dummy `hf-hello` script — no cluster needed
