# Delta spec: technical-architecture

## MODIFIED Requirements

### Requirement: Go Module Structure

The CLI SHALL be organized as a single Go module. The `config.go` comment MUST NOT reference the removed `bootstrap` subcommand.

#### Scenario: Top-level module layout

- GIVEN the CLI is built as a Go project
- WHEN the repository is initialized
- THEN the module MUST follow this package structure:
  ```
  hf/
  ├── cmd/                    # Cobra command definitions (one file per command group)
  │   ├── root.go             # Root command, global flags, plugin loading
  │   ├── cluster.go          # hf cluster [create|get|list|search|patch|delete|conditions|statuses|table]
  │   ├── nodepool.go         # hf nodepool [create|get|list|search|patch|delete|conditions|statuses|table]
  │   ├── adapter.go          # hf cluster adapter post-status, hf nodepool adapter post-status
  │   ├── config.go           # hf config [show|set|clear|doctor|env]
  │   ├── db.go               # hf db [query|delete|delete-all|statuses|config]
  │   ├── maestro.go          # hf maestro [list|get|delete|bundles|consumers|tui]
  │   ├── pubsub.go           # hf pubsub [list|publish]
  │   ├── rabbitmq.go         # hf rabbitmq [publish]
  │   ├── kube.go             # hf kube [port-forward|context|curl|debug]
  │   ├── logs.go             # hf logs [<pattern>|adapter]
  │   ├── repos.go            # hf repos
  │   └── table.go            # hf table (combined overview)
  ├── internal/
  │   ├── api/                # HyperFleet API client
  │   ├── config/             # Configuration management (split YAML model)
  │   ├── output/             # Output formatting (JSON, table, YAML, colored dots)
  │   ├── resource/           # Shared resource types and data structures
  │   ├── watch/              # Watch mode (periodic refresh with diff)
  │   ├── kube/               # Kubernetes operations (client-go wrapper)
  │   ├── maestro/            # Maestro client (HTTP API + maestro-cli fallback)
  │   ├── pubsub/             # Pub/Sub and RabbitMQ event publishing
  │   ├── db/                 # PostgreSQL database operations
  │   ├── plugin/             # Plugin discovery and loading
  │   └── version/            # Build version info
  ├── plugins/                # Example plugin implementations
  ├── main.go                 # Entry point
  ├── go.mod
  └── go.sum
  ```
- AND each `cmd/` file MUST register its commands with the Cobra root command
- AND all business logic MUST reside in `internal/` packages, not in `cmd/` files

---

### Requirement: Cobra Command Tree

The CLI SHALL use Cobra for command routing. The command tree MUST NOT include `config bootstrap`; `hf config env new` supersedes it (see phase-02-config-ux).

#### Scenario: Command hierarchy

- GIVEN Cobra is the CLI framework
- WHEN commands are registered
- THEN the `config` subtree in the command tree MUST be:
  ```
  ├── config
  │   ├── show      [env-name]
  │   ├── set       <key> <value>
  │   ├── clear     <key|all>
  │   ├── doctor
  │   └── env
  │       ├── new      [name]
  │       ├── list
  │       ├── show     <name>
  │       └── activate <name>
  ```
