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

#### Scenario: Query returns no rows

- WHEN the query returns 0 rows
- THEN the CLI MUST print an info message and exit 0

#### Scenario: List database tables

- GIVEN the database is accessible
- WHEN the user runs `hf db query "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"`
- THEN the output MUST list the tables: `migrations`, `adapter_statuses`, `clusters`, `node_pools`

### Requirement: Delete Table Rows

The CLI SHALL delete rows from a specific table with a confirmation prompt.

#### Scenario: Delete from table with WHERE clause

- GIVEN database connection is configured
- WHEN the user runs `hf db delete <table> <where>`
- THEN the CLI MUST show a count of records matching the WHERE clause
- AND prompt the user for confirmation (requiring `y` or `yes`)
- AND run `DELETE FROM <table> WHERE <where>` only after confirmation
- AND print the count of deleted rows

#### Scenario: Confirmation denied

- WHEN the user does not confirm
- THEN the CLI MUST print "Aborted" and exit 0 without deleting anything

### Requirement: Delete All Records from a Table

The CLI SHALL delete ALL records from a named table with confirmation.

#### Scenario: Delete all rows

- GIVEN database connection is configured
- WHEN the user runs `hf db delete-all <table>`
- THEN the CLI MUST show the total row count for that table
- AND prompt the user for confirmation (requiring `y` or `yes`)
- AND run `DELETE FROM <table>` only after confirmation

### Requirement: Database Configuration Display

The CLI SHALL display the resolved database connection parameters.

#### Scenario: Show DB config

- GIVEN the CLI is running
- WHEN the user runs `hf db config`
- THEN the CLI MUST print host, port, name, user as plain values
- AND mask the password as `<set>` or `<not set>`
- AND require no database connection to run
