# Design: Phase 00 — Bootstrap

## Module

```
module github.com/rh-amarin/hyperfleet-cli
go 1.22
```

## Directory Layout

```
hf/
├── cmd/
│   ├── root.go       # Root command, persistent flags, Execute()
│   └── version.go    # hf version
├── internal/
│   └── version/
│       └── version.go  # Version, Commit, Date vars
├── main.go
├── Makefile
├── .goreleaser.yaml
├── go.mod
└── go.sum
```

## Root Command (`cmd/root.go`)

Persistent flags available on every subcommand:
- `--config <path>` — override config directory location
- `--output <format>` — json | table | yaml
- `--no-color` — disable ANSI colors
- `--verbose / -v` — debug logging to stderr
- `--api-url <url>` — override API URL for this invocation
- `--api-token <token>` — override API token for this invocation

## Version Command (`cmd/version.go`)

Prints: `hf version <Version> (commit: <Commit>, built: <Date>)`

## Version Package (`internal/version/version.go`)

```go
var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)
```

Injected at build time via:
```
-ldflags "-X github.com/rh-amarin/hyperfleet-cli/internal/version.Version=..."
```

## Makefile

```
build      go build -ldflags "..." -o bin/hf .
install    go install -ldflags "..." .
lint       golangci-lint run
test       go test ./...
```

## GoReleaser

Targets: linux/mac/windows × amd64/arm64. SHA256 checksums. Binary named `hf`.
