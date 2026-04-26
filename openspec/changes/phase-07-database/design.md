# Design: Phase 07 — Database

## Package Layout

```
internal/db/
  client.go       ← DBClient interface + Client struct (pgxpool)
  client_test.go  ← mock implementation + contract tests
cmd/
  db.go           ← all hf db subcommands
  db_test.go      ← unit tests using mock DBClient
```

## DBClient Interface (internal/db/client.go)

```go
type DBClient interface {
    Query(ctx context.Context, sql string, args ...any) (headers []string, rows [][]string, err error)
    Exec(ctx context.Context, sql string, args ...any) error
    Close()
}
```

Defined in `internal/db` so `cmd` package can import it for mock injection.

## Client struct

Wraps `*pgxpool.Pool`. `New(ctx, dsn)` calls `pgxpool.New` then `pool.Ping`.
`Query` reads `FieldDescriptions()` for headers; formats all values with `fmt.Sprintf("%v", v)`.
`Close` calls `pool.Close()`.

## DSN Construction

```
postgres://<user>:<password>@<host>:<port>/<name>
```

Config keys from `cfgStore.Cfg().Database`:
- `host` (default: localhost)
- `port` (default: 5432)
- `name`
- `user`
- `password`

## Command Tree

```
hf db
  query <sql>              SELECT → table via printer().PrintTable
  delete <table> <where>   DELETE FROM <table> WHERE <where> with count + confirm
  delete-all <table>       DELETE FROM <table> with count + confirm
  statuses                 SELECT * FROM adapter_statuses → table
  statuses-delete          DELETE FROM adapter_statuses with count + confirm
  config                   Show resolved DB config (password masked)
```

## Testability Design

Two package-level variables in `cmd/db.go` are overrideable in tests:

```go
var dbClientFactory = func(ctx context.Context, dsn string) (db.DBClient, error) {
    return db.New(ctx, dsn)
}
var stdinReader io.Reader = os.Stdin
```

Tests in `cmd` package set these before calling `runCmdRaw`, then restore via `defer`.

## Confirmation Prompt

`confirmPrompt(msg)` writes `msg [y/N]: ` to stderr and reads one line from `stdinReader`.
Returns true only for `y` or `yes` (case-insensitive).

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Dependency | `pgxpool` | Connection pooling, context support, idiomatic pgx v5 |
| Interface location | `internal/db` | Allows cmd package to import for mocking |
| SQL building | `fmt.Sprintf` for table/where fragments | Table names cannot be parameterized; this is an admin tool |
| Output format | `PrintTable` for all DB results | DB results are always tabular; JSON output of raw rows is not useful |
| Password display | `<set>` / `<not set>` | Consistent with existing secret handling in config show |
