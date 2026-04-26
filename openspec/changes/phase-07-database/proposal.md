# Proposal: Phase 07 — Database

## Intent

Implement direct PostgreSQL access for the HyperFleet CLI, replacing the previous
`psql`/`kubectl exec` approach with a native Go client backed by `jackc/pgx/v5`.

## Scope In

- `internal/db` package: `DBClient` interface, `Client` struct wrapping `pgxpool.Pool`
- `cmd/db.go`: all `hf db` subcommands
- `go.mod`/`go.sum`: add `github.com/jackc/pgx/v5`

## Scope Out

- Kubernetes port-forward setup (phase 10)
- Interactive `hf db config bootstrap` (out of scope for this phase)
- RabbitMQ/pubsub (phase 9)

## Why

The bash scripts used `psql` with environment variables and `kubectl exec` for in-cluster
access. The Go CLI replaces these with a native pgx connection pool, giving us typed
results, connection reuse, and proper error handling without shelling out.

## Testing Scope

| Package | Test Cases |
|---|---|
| `internal/db` | mockClient implements DBClient (compile-time), Query/Exec/Close contract |
| `cmd` | db config, db query (table output), db delete (confirm Y/N), db delete-all, db statuses, db statuses-delete |

Live cluster verification requires DB connectivity; unit tests use mocks only.
