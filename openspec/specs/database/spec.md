# Database Operations Specification

## Purpose

Provide CLI commands for direct PostgreSQL database operations against the HyperFleet database, including arbitrary SQL queries, selective deletion, bulk deletion, and in-cluster database access via kubectl.

## Requirements

### Requirement: Execute SQL Query

The CLI SHALL execute arbitrary SQL queries against the HyperFleet PostgreSQL database.

#### Scenario: Query with inline SQL

- GIVEN database connection config is set (db-host, db-port, db-name, db-user, db-password)
- WHEN the user runs `hf db query "<SQL>"`
- THEN the CLI MUST set PGHOST, PGPORT, PGDATABASE, PGUSER, PGPASSWORD environment variables from config
- AND execute the query via `psql`
- AND output the results in psql's default formatted table output

#### Scenario: Query from SQL file

- GIVEN database connection config is set
- WHEN the user runs `hf db query -f <file.sql>`
- THEN the CLI MUST execute the SQL from the specified file

#### Scenario: Query cluster data

- GIVEN clusters exist in the database
- WHEN the user runs `hf db query "SELECT id, name, generation FROM clusters LIMIT 5"`
- THEN the output MUST show a formatted table with the requested columns

#### Scenario: List database tables

- GIVEN the database is accessible
- WHEN the user runs `hf db query "SELECT table_name FROM information_schema.tables WHERE table_schema='public'"`
- THEN the output MUST list the tables: `migrations`, `adapter_statuses`, `clusters`, `node_pools`

### Requirement: Delete Table Rows

The CLI SHALL delete rows from a specific table with a confirmation preview.

#### Scenario: Delete from table

- GIVEN database connection is configured
- WHEN the user runs `hf db delete <table> [id]`
- THEN the CLI MUST show a count of records to be deleted
- AND prompt the user for confirmation (requiring `yes`)
- AND only proceed with deletion after confirmation
- AND support tables: `clusters`, `node_pools`, `adapter_statuses`

#### Scenario: Delete specific record by ID

- GIVEN a table and record ID are provided
- WHEN the user runs `hf db delete <table> <id>`
- THEN only the record matching the ID MUST be deleted

### Requirement: Delete All Records

The CLI SHALL delete ALL records from all HyperFleet tables with confirmation.

#### Scenario: Delete all records

- GIVEN database connection is configured
- WHEN the user runs `hf db delete-all`
- THEN the CLI MUST show record counts per table
- AND prompt the user for confirmation (requiring `yes`)
- AND delete in foreign-key-safe order: `adapter_statuses` first, then `node_pools`, then `clusters`

### Requirement: In-Cluster Database Status Query

The CLI SHALL query adapter statuses from a PostgreSQL instance running inside Kubernetes.

#### Scenario: Query in-cluster statuses

- GIVEN context, namespace, and cluster-id are configured
- WHEN the user runs `hf db statuses`
- THEN the CLI MUST find the postgres pod in the Kubernetes cluster
- AND execute the SQL query via `kubectl exec`
- AND display the adapter_statuses results

### Requirement: In-Cluster Status Deletion

The CLI SHALL delete adapter statuses from a PostgreSQL instance running inside Kubernetes.

#### Scenario: Delete in-cluster statuses

- GIVEN context, namespace, and cluster-id are configured
- WHEN the user runs `hf db statuses delete`
- THEN the CLI MUST connect through kubectl to the postgres pod
- AND delete adapter_statuses scoped to the current cluster-id

### Requirement: Database Configuration

The CLI SHALL provide interactive configuration of PostgreSQL connection parameters.

#### Scenario: Configure database interactively

- GIVEN the CLI is running
- WHEN the user runs `hf db config`
- THEN the CLI MUST prompt for: host, port, database name, user, password
- AND save each value to the config directory
