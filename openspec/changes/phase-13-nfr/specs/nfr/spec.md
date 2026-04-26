# Non-Functional Requirements Spec

## Shell Completions

`hf completion <shell>` writes the shell completion script to stdout.
Supported shells: `bash`, `zsh`, `fish`, `powershell`.
Unknown shell argument returns a non-zero exit code and an error message.

Install instructions:
- bash: `hf completion bash > /etc/bash_completion.d/hf`
- zsh: `hf completion zsh > "${fpath[1]}/_hf"`
- fish: `hf completion fish | source`
- powershell: `hf completion powershell | Out-String | Invoke-Expression`

## Plugin System

### Discovery

The CLI searches `$PATH` for executables named `hf-<name>` when `<name>` is not a registered subcommand. Plugin names are normalised to lowercase before the PATH lookup.

### Execution

When a plugin is found, the CLI replaces its own process with the plugin process via `syscall.Exec` (POSIX) so that the plugin inherits the terminal, signals, and exit code directly. Arguments after the plugin name are forwarded verbatim.

### Error on Missing Plugin

If neither a registered subcommand nor a `hf-<name>` executable is found, the CLI prints:
```
Error: unknown command "<name>" for "hf"
Run 'hf --help' for usage.
```
and exits non-zero.

## Release Pipeline

### CI (`.github/workflows/ci.yml`)

- Triggers: push to `main`, pull_request to `main`
- Steps: `go build ./...`, `go vet ./...`, `go test ./...`
- Integration tests are excluded (no `-tags integration` flag)

### Release (`.github/workflows/release.yml`)

- Trigger: push of a tag matching `v*`
- Uses `goreleaser/goreleaser-action@v6` with `--clean` flag
- Produces: binaries for linux/darwin/windows × amd64/arm64, `tar.gz`/`zip` archives, SHA256 checksums file, GitHub Release with auto-generated changelog

## Error Handling Conventions

| Situation | Output | Exit code |
|---|---|---|
| API error (RFC 9457) | Rendered via `printer().Print(apiErr)` to stdout in `--output` format | 0 |
| CLI usage error | Error message to stderr, usage block suppressed (`SilenceUsage: true`) | 1 |
| `[WARN]` message | `os.Stderr`, `[WARN] <msg>` prefix | 0 |
| `[INFO]` message | `os.Stderr`, `[INFO] <msg>` prefix | 0 |
| `[ERROR]` message | `os.Stderr`, `[ERROR] <msg>` prefix | varies |

`SilenceUsage: true` is set on the root command, ensuring that runtime errors (API failures, missing config) do not cause the help block to be printed alongside the error.
