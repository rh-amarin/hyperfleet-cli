# Database Operations Specification

## Purpose

Provide CLI commands for direct PostgreSQL database operations against the HyperFleet
database, using native Go (`pgxpool`) instead of `psql` subprocesses or `kubectl exec`.

## Config Keys

| Key | Default |
|---|---|
| `database.host` | `localhost` |
| `database.port` | `5432` |
| `database.name` | `hyperfleet` |
| `database.user` | `hyperfleet` |
| `database.password` | `foobar-bizz-buzz` |

DSN format: `postgres://<user>:<password>@<host>:<port>/<name>`

## Requirements

### Requirement: Execute SQL Query

The CLI SHALL execute arbitrary SQL queries against the HyperFleet PostgreSQL database.

#### Scenario: Query with inline SQL

- GIVEN database connection config is set (database.host/port/name/user/password)
- WHEN the user runs `hf db query "<SQL>"`
- THEN the CLI MUST connect via pgxpool using the resolved DSN
- AND execute the query natively (no subprocess)
- AND output results as a formatted table

#### Scenario: Query from file

- GIVEN database connection config is set
- WHEN the user runs `hf db query -f <filepath>`
- THEN the CLI MUST read the SQL from the specified file
- AND execute it natively against the database
- AND output results as a formatted table
- AND exit with code 1 and an `[ERROR]` message if the file cannot be read

#### Scenario: Query returns no rows

- WHEN the query returns 0 rows
- THEN the CLI MUST print an info message and exit 0

#### Scenario: List database tables

- GIVEN the database is accessible
- WHEN the user runs `hf db query "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"`
- THEN the output MUST list the tables: `migrations`, `adapter_statuses`, `clusters`, `node_pools`

### Requirement: Delete Records

The CLI SHALL delete all records from a specified table, or from all tables, with a confirmation prompt.

The `<target>` argument is required and MUST be one of: `clusters`, `nodepools`, `adapter_statuses`, or `ALL`.
The argument values MUST be offered as shell completions.

#### Scenario: Delete all records from a single table

- GIVEN database connection is configured
- WHEN the user runs `hf db delete clusters`
- THEN the CLI MUST show the total row count for that table
- AND prompt the user for confirmation (requiring `yes`)
- AND run `DELETE FROM clusters` only after confirmation
- AND print the count of deleted rows

- WHEN the user runs `hf db delete nodepools`
- THEN the same behavior applies for the `node_pools` table

- WHEN the user runs `hf db delete adapter_statuses`
- THEN the same behavior applies for the `adapter_statuses` table

#### Scenario: Delete all records from all tables

- GIVEN database connection is configured
- WHEN the user runs `hf db delete ALL`
- THEN the CLI MUST show the row count for each table
- AND prompt the user for confirmation (requiring `yes`)
- AND delete in dependency order: `adapter_statuses` first, then `node_pools`, then `clusters`
- AND print the count of deleted rows per table

#### Scenario: Confirmation denied

- WHEN the user does not confirm (any input other than `yes`)
- THEN the CLI MUST print "Aborted" and exit 0 without deleting anything

### Requirement: Database Configuration Display

The CLI SHALL display the resolved database connection parameters.

#### Scenario: Show DB config

- GIVEN the CLI is running
- WHEN the user runs `hf db config`
- THEN the CLI MUST print host, port, name, user as plain values
- AND mask the password as `<set>` or `<not set>`
- AND require no database connection to run
