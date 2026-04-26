# Phase 13 Design — Non-Functional Requirements

## Shell Completions (`cmd/completion.go`)

Cobra provides completion generation out of the box. The `completion` command accepts exactly one arg from `[bash, zsh, fish, powershell]` and writes the corresponding script to stdout.

```
hf completion bash       # sources into ~/.bashrc
hf completion zsh        # eval "$(hf completion zsh)"
hf completion fish       # hf completion fish | source
hf completion powershell # hf completion powershell | Out-String | Invoke-Expression
```

`SilenceUsage: true` is set so that a bad argument doesn't dump the usage block alongside the error.

## Plugin Architecture (`internal/plugin/`)

### Find

`plugin.Find(name string) (string, bool)` searches `$PATH` for an executable named `hf-<lowercase(name)>` using `exec.LookPath`. Returns the resolved path and `true` if found.

### Exec

`plugin.Exec(path string, args []string) error` replaces the current process with the plugin binary via `syscall.Exec`. The full argv is `[path, args...]`. The plugin inherits `os.Environ()`. On platforms that don't support `syscall.Exec` (Windows), falls back to `exec.Command` with `os.Stdin/Stdout/Stderr` wired, then `os.Exit(cmd.ProcessState.ExitCode())`.

### Dispatch wiring (`cmd/plugin_dispatch.go`)

The root command currently has no `RunE`. Setting one from `plugin_dispatch.go`'s `init()` is safe and does not conflict with subcommand registration (each subcommand's own `init()` runs independently).

```go
func init() {
    rootCmd.Args = cobra.ArbitraryArgs
    rootCmd.RunE = func(cmd *cobra.Command, args []string) error { ... }
}
```

When `hf <name> [args...]` is invoked and `<name>` is not a registered subcommand, Cobra passes all remaining args to the root `RunE`. The dispatch function:
1. Checks `len(args) == 0` → print help
2. Calls `plugin.Find(args[0])` → if found, `plugin.Exec(path, args[1:])`
3. Otherwise, returns the same "unknown command" error Cobra would have shown

### Testability

`plugin.Find` uses `exec.LookPath` which respects the `$PATH` env var. Tests create a temp executable named `hf-testplugin`, prepend its directory to `$PATH` via `t.Setenv("PATH", ...)`, and verify `Find("testplugin")` returns true.

`plugin.Exec` is harder to test without replacing the process; tests verify the happy path by checking the found path is correct and that `Exec` returns an error only when the binary doesn't exist.

The dispatch integration test in `cmd/` sets up a temp `hf-hello` executable, overrides `PATH`, and runs the root command with `hello world` args.

## GoReleaser (`.goreleaser.yaml`)

Already has:
- Builds for linux/darwin/windows × amd64/arm64
- SHA256 checksums
- `tar.gz` archives with zip for windows
- Changelog with asc sort and exclusion filters

Additions:
- `release:` section to configure GitHub Releases auto-created on tag
- `changelog.use: git` for cleaner release notes
- Archive `files:` to include LICENSE and README if present

## GitHub Actions

### `ci.yml`

Trigger: `push` to main, `pull_request` to main.
Steps: checkout → setup-go (go-version-file: go.mod) → build → vet → test.
Unit tests only (no `-tags integration`).

### `release.yml`

Trigger: `push` to tags matching `v*`.
Steps: checkout (fetch-depth: 0 for full tag history) → setup-go → goreleaser-action@v6.

## Error & Usage Polish

All subcommands already set `SilenceUsage: true` (verified by audit). The root command has `SilenceUsage: true`.

API error rendering: the existing pattern in `clusterGetCmd` (return `printer().Print(apiErr)` for API errors) is the established pattern. Commands that propagate API errors directly via `return err` will cause the error text to be printed by Cobra's error handler. This is acceptable — the structured printing path (`IsAPIError` check) is already in `clusterGetCmd` as the reference pattern.

`Warn` and `Info` already write to stderr with correct prefixes (verified in `internal/output/printer.go`).

## Integration Tests (`cmd/integration_test.go`)

Build tag `//go:build integration` ensures the file is excluded from `go test ./...` without `-tags integration`. The file contains a `TestIntegration_ClusterList` skeleton that skips unless `HF_API_URL` and `HF_API_TOKEN` are set. This satisfies the CI requirement that unit tests run cleanly without a cluster.
