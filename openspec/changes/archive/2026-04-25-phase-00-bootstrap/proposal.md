# Proposal: Phase 00 — Bootstrap

## Intent

Stand up the Go module, Cobra root command, and build tooling so the project compiles and `hf version` works. This is the foundation every subsequent phase builds on.

## Scope

**In scope:**
- Go module initialization (`github.com/rh-amarin/hyperfleet-cli`)
- `main.go` entry point
- Root Cobra command with all global persistent flags
- `hf version` subcommand
- `internal/version` package with ldflags injection
- `Makefile` with build/install/lint/test targets
- `.goreleaser.yaml` for cross-compilation (linux/mac/windows × amd64/arm64)

**Out of scope:**
- Any functional commands beyond `version` and `help`
- Configuration loading
- API calls
