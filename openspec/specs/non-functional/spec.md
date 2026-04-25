# Non-Functional Requirements Specification

## Purpose

Define non-functional requirements for the HyperFleet CLI covering shell completions, multi-format output, cross-platform builds, plugin architecture, testing strategy, and distribution.

## Requirements

### Requirement: Shell Completions

The CLI SHALL generate shell completion scripts for all major shells.

#### Scenario: Generate completion scripts

- GIVEN Cobra provides built-in completion support
- WHEN the user runs `hf completion <shell>`
- THEN the CLI MUST generate completion scripts for:
  - `bash`: Bash completion v2
  - `zsh`: Zsh completion with descriptions
  - `fish`: Fish completion
  - `powershell`: PowerShell completion

#### Scenario: Dynamic completions

- GIVEN the user is tab-completing arguments
- WHEN a command accepts known values (e.g., `spec|labels` for patch, `start|stop|status` for port-forward)
- THEN the CLI MUST provide those values as completion suggestions
- AND for resource IDs (cluster-id, nodepool-id), the CLI SHOULD offer live completions by querying the API

#### Scenario: Completion installation instructions

- GIVEN the user runs `hf completion bash`
- WHEN the script is generated
- THEN the CLI MUST print instructions for installing the completion (e.g., `source <(hf completion bash)` or adding to `.bashrc`)

### Requirement: Multi-Format Output

The CLI SHALL support a global `--output` flag for controlling output format on every command that produces data.

#### Scenario: JSON output

- GIVEN `--output json` is specified (or `-o json`)
- WHEN any command produces output
- THEN the output MUST be well-formed, pretty-printed JSON with 2-space indentation
- AND the output MUST be suitable for piping to `jq` or other JSON tools

#### Scenario: Table output

- GIVEN `--output table` is specified or is the default for list/table commands
- WHEN the command produces output
- THEN the output MUST be a human-readable formatted table with:
  - Column headers in uppercase
  - A separator line (`---`) under headers
  - Aligned columns using tab-width or padding
  - Colored status indicators (respecting `--no-color`)

#### Scenario: YAML output

- GIVEN `--output yaml` is specified
- WHEN any command produces output
- THEN the output MUST be valid YAML

#### Scenario: Default output format per command type

- GIVEN no `--output` flag is specified
- WHEN a command produces output
- THEN the default MUST be:
  - `table` for list views: `cluster list`, `cluster table`, `nodepool list`, `nodepool table`, `table`, `repos`, `conditions-table`
  - `json` for single-resource views: `cluster get`, `nodepool get`, `cluster conditions`, `nodepool conditions`, `cluster statuses`, `nodepool statuses`
  - `text` for config commands, port-forward status, and log output

#### Scenario: Quiet mode for scripting

- GIVEN `-q` or `--quiet` is specified
- WHEN a command produces output
- THEN only the essential data MUST be output (no headers, no color, no decorative elements)
- AND for create/search commands, only the resource ID MUST be output

### Requirement: Cross-Platform Build and Distribution

The CLI SHALL be built and distributed for multiple platforms using GoReleaser.

#### Scenario: Supported platforms

- GIVEN GoReleaser is configured
- WHEN a release is built
- THEN binaries MUST be produced for:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- AND each binary MUST be a statically linked, self-contained executable

#### Scenario: Release artifacts

- GIVEN a new version is tagged
- WHEN GoReleaser runs
- THEN it MUST produce:
  - Compressed archives (`.tar.gz` for linux/mac, `.zip` for windows)
  - SHA256 checksums file
  - GitHub Release with changelog from conventional commits
- AND the binary name MUST be `hf`

#### Scenario: Version information

- GIVEN the binary is built with ldflags
- WHEN the user runs `hf version`
- THEN the CLI MUST display:
  - Version (semantic version from git tag)
  - Git commit SHA
  - Build date
  - Go version
  - OS/Arch

#### Scenario: Homebrew and package managers

- GIVEN the CLI is released
- WHEN distribution channels are configured
- THEN GoReleaser SHOULD generate a Homebrew formula
- AND the CLI SHOULD be installable via `brew install hf`

### Requirement: Plugin Architecture

The CLI SHALL support extending functionality through plugins without modifying the core binary.

#### Scenario: Plugin discovery

- GIVEN the CLI starts
- WHEN the plugin system initializes
- THEN the CLI MUST scan for plugins in:
  1. `~/.config/hf/plugins/`
  2. Directories listed in `HF_PLUGIN_PATH` environment variable
- AND plugins MUST be executable files named `hf-<command>` (e.g., `hf-custom-report`)

#### Scenario: Plugin execution

- GIVEN a plugin `hf-custom-report` exists in the plugin path
- WHEN the user runs `hf custom-report [args...]`
- THEN Cobra MUST delegate to the plugin executable
- AND pass all remaining arguments to the plugin
- AND set environment variables for the plugin:
  - `HF_CONFIG_DIR`: path to config directory
  - `HF_API_URL`: resolved API URL
  - `HF_API_TOKEN`: resolved API token
  - `HF_CLUSTER_ID`: current cluster ID from state
  - `HF_NODEPOOL_ID`: current nodepool ID from state
  - `HF_OUTPUT_FORMAT`: current output format

#### Scenario: Plugin listing

- GIVEN plugins are installed
- WHEN the user runs `hf plugin list`
- THEN the CLI MUST list all discovered plugins with:
  - Name (derived from filename)
  - Path
  - Whether it is executable

#### Scenario: Plugin help integration

- GIVEN a plugin `hf-custom-report` exists
- WHEN the user runs `hf --help`
- THEN the plugin MUST appear in the help output under an "Available Plugins" section
- AND running `hf custom-report --help` MUST delegate to the plugin

#### Scenario: Plugin conflict resolution

- GIVEN a plugin name conflicts with a built-in command
- WHEN the CLI resolves the command
- THEN the built-in command MUST take precedence
- AND a warning MUST be displayed: `[WARN] Plugin 'hf-<name>' is shadowed by built-in command '<name>'`

### Requirement: Testing Strategy

The CLI SHALL maintain a comprehensive test suite.

#### Scenario: Unit tests

- GIVEN all `internal/` packages contain business logic
- WHEN tests are run
- THEN each package MUST have unit tests covering:
  - Config loading, merging, and precedence
  - API client request construction and error parsing
  - Output formatting (JSON, table, YAML rendering)
  - Resource type marshaling/unmarshaling
  - Watch mode tick behavior

#### Scenario: Integration tests

- GIVEN the CLI interacts with external services (API, database, Kubernetes)
- WHEN integration tests are run
- THEN tests MUST use:
  - `httptest.Server` for API client tests against a mock HyperFleet API
  - `testcontainers-go` or similar for PostgreSQL integration tests
  - `envtest` from controller-runtime for Kubernetes client tests
- AND integration tests MUST be tagged with `//go:build integration`

#### Scenario: Command tests

- GIVEN commands are defined in `cmd/`
- WHEN command tests are run
- THEN each command MUST have at least one test verifying:
  - Correct flag parsing
  - Expected API calls (using HTTP mock)
  - Expected output format
  - Exit code behavior

#### Scenario: End-to-end tests

- GIVEN a full test environment is available
- WHEN E2E tests are run
- THEN the CLI MUST be tested against a real HyperFleet API instance
- AND the test suite MUST exercise the full cluster lifecycle: create, get, list, patch, adapter-status, conditions, delete

### Requirement: Graceful Degradation

The CLI SHALL degrade gracefully when optional dependencies are unavailable.

#### Scenario: Missing GCP credentials

- GIVEN GCP credentials are not configured
- WHEN the user runs `hf pubsub list` or `hf pubsub publish`
- THEN the CLI MUST display a clear error: `[ERROR] GCP credentials not found. Run 'gcloud auth application-default login' or set GOOGLE_APPLICATION_CREDENTIALS`
- AND other commands MUST continue to work

#### Scenario: Missing maestro-cli for TUI

- GIVEN `maestro-cli` is not installed
- WHEN the user runs `hf maestro tui`
- THEN the CLI MUST display: `[ERROR] maestro-cli not found. Install from <url> or use 'hf maestro list/get/delete' for API-based access`
- AND `hf maestro list`, `hf maestro get`, `hf maestro delete`, `hf maestro bundles`, `hf maestro consumers` MUST work without maestro-cli

#### Scenario: Unreachable API

- GIVEN the HyperFleet API is unreachable
- WHEN any API command is invoked
- THEN the CLI MUST display a clear error with the attempted URL
- AND suggest checking `hf config show` and `hf kube port-forward status`

#### Scenario: Database unreachable

- GIVEN the PostgreSQL database is unreachable
- WHEN a database command is invoked
- THEN the CLI MUST display the connection error with host:port
- AND suggest checking `hf config show` for database settings

### Requirement: Performance

The CLI SHALL respond promptly for common operations.

#### Scenario: Command startup time

- GIVEN the CLI binary is loaded
- WHEN any command is executed
- THEN the startup time (before first network call) MUST be under 100ms

#### Scenario: Parallel data fetching

- GIVEN `hf repos` queries 7 repositories
- WHEN the data is fetched
- THEN requests MUST be made concurrently using goroutines
- AND the total execution time MUST be bounded by the slowest single request, not the sum of all requests

#### Scenario: Large list pagination

- GIVEN the API returns paginated results
- WHEN a list command encounters pagination
- THEN the CLI MUST handle pagination transparently, fetching all pages
- AND support `--limit` to cap the total number of results returned

### Requirement: Security

The CLI SHALL follow security best practices.

#### Scenario: Token storage

- GIVEN API tokens are stored in config
- WHEN the config file is created or updated
- THEN the file permissions MUST be set to `0600` (owner read/write only)

#### Scenario: No token in logs

- GIVEN verbose mode is enabled
- WHEN HTTP requests are logged
- THEN the `Authorization` header value MUST be masked (e.g., `Bearer <redacted>`)
- AND the `password` fields in config show MUST remain masked

#### Scenario: TLS verification

- GIVEN the API URL uses HTTPS
- WHEN the CLI connects
- THEN TLS certificate verification MUST be enabled by default
- AND `--insecure` flag MAY be provided to skip verification with a warning
