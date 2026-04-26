# Database Spec — Phase 07 Delta

## Changes from openspec/specs/database/spec.md

### MODIFIED: Execute SQL Query

Original spec required `psql` subprocess with env vars. This implementation uses
native Go `pgxpool` instead:

- GIVEN database connection config is set
- WHEN the user runs `hf db query "<SQL>"`
- THEN the CLI MUST connect via pgxpool DSN `postgres://<user>:<password>@<host>:<port>/<name>`
- AND execute the query natively (no subprocess)
- AND output results as a formatted table using `printer().PrintTable`

### MODIFIED: Delete Table Rows

- `hf db delete <table> <where>` — deletes rows matching `WHERE <where>` clause
- Shows COUNT of affected rows before prompting
- Requires `y` or `yes` confirmation

### MODIFIED: Delete All Records

- `hf db delete-all <table>` — deletes all rows from a single named table
- Shows COUNT before prompting
- Requires `y` or `yes` confirmation

### MODIFIED: Adapter Statuses (statuses / statuses-delete)

- `hf db statuses` — connects via pgxpool, runs `SELECT * FROM adapter_statuses`, prints as table
- `hf db statuses-delete` — shows count, prompts, then `DELETE FROM adapter_statuses`
- No longer requires `kubectl exec`; uses direct DB connection

### ADDED: Database Config Display

- `hf db config` — prints resolved DB connection parameters (host, port, name, user)
- Password is masked as `<set>` or `<not set>`
- No DB connection required; reads from in-memory config only

## Config Keys

| Key | Default |
|---|---|
| `database.host` | `localhost` |
| `database.port` | `5432` |
| `database.name` | `hyperfleet` |
| `database.user` | `hyperfleet` |
| `database.password` | `foobar-bizz-buzz` |
