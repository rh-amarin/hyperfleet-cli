# Tables and Lists Specification

## Purpose

Provide aggregated table and list views that display multiple resources in formatted output, including cluster lists, cluster tables with dynamic condition columns, nodepool tables, and a combined overview table.

## Requirements

### Requirement: List Clusters (JSON)

The CLI SHALL list all clusters as raw JSON.

#### Scenario: Empty cluster list

- GIVEN no clusters exist in the API
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST output `{"items": [], "kind": "ClusterList", "page": 1, "size": 0, "total": 0}`

#### Scenario: Populated cluster list

- GIVEN clusters exist
- WHEN the user runs `hf cluster list`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters`
- AND output the full JSON response with all cluster objects in the `items` array
- AND include pagination fields: `page`, `size`, `total`

### Requirement: Cluster Table View

The CLI SHALL display clusters in a formatted table with dynamic condition columns when the `--table` flag is passed to `hf cluster list`.

#### Scenario: Empty table

- GIVEN no clusters exist
- WHEN the user runs `hf cluster list --table`
- THEN the CLI MUST output table headers only: `NAME  GEN` with a separator line

#### Scenario: Populated table with conditions

- GIVEN clusters exist with various adapter conditions
- WHEN the user runs `hf cluster list --table`
- THEN the CLI MUST output a table with fixed columns: NAME, GEN
- AND dynamic columns for each unique condition type found across all clusters (e.g., `Available`, `ClDeploymentSuccessful`, `ClJobSuccessful`, `ClNamespaceSuccessful`, `Ready`)
- AND the `Available` column MUST appear first among dynamic columns
- AND the `Ready` column MUST appear last
- AND all other condition types MUST appear alphabetically between `Available` and `Ready`
- AND status values MUST be displayed as colored dots:
  - Green dot = `True`
  - Red dot = `False`
  - Yellow dot = `Unknown`
  - `-` = condition not present for that cluster

### Requirement: NodePool Table View

The CLI SHALL display nodepools in the current cluster as a formatted table when the `--table` flag is passed to `hf nodepool list`.

#### Scenario: Populated nodepool table

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool list --table`
- THEN the CLI MUST output a table with fixed columns: NAME, REPLICAS, TYPE, GEN
- AND dynamic condition columns following the same ordering pattern as the cluster table (`Available` first, `Ready` last, others alphabetical)
- AND REPLICAS MUST show the `spec.replicas` value
- AND TYPE MUST show the `spec.platform.type` value

### Requirement: Combined Overview Table

The CLI SHALL display a combined table of all clusters and their nested nodepools.

#### Scenario: Display combined table

- GIVEN clusters and nodepools exist
- WHEN the user runs `hf table`
- THEN the CLI MUST output a table with columns: NAME, KIND, CLUSTER, GEN, then dynamic condition columns
- AND cluster rows MUST have an empty CLUSTER field
- AND nodepool rows MUST show the parent cluster name in the CLUSTER field
- AND both clusters and nodepools MUST share the same dynamic column set (union of all condition types)
- AND status values MUST be displayed as colored dots

### Requirement: Watch Mode

The CLI SHALL support a watch mode that periodically refreshes output.

#### Scenario: Watch with default interval

- GIVEN the user wants to monitor resources continuously
- WHEN the user runs any watch-capable command with the `-w` flag (e.g., `hf cluster conditions -w`)
- THEN the CLI MUST refresh output every 2 seconds (default interval)
- AND MUST clear the terminal between refreshes using ANSI escape sequence `\033[H\033[2J`
- AND MUST print a footer to stderr with the format `Last updated: HH:MM:SS  (Ctrl+C to stop)`

#### Scenario: Watch with custom interval

- GIVEN the user wants to monitor with a specific refresh rate
- WHEN the user runs a watch-capable command with `-w -i <duration>` (e.g., `hf cluster conditions -w -i 5s`)
- THEN the CLI MUST refresh output at the specified interval

#### Scenario: Watch termination

- GIVEN the CLI is running in watch mode
- WHEN the process receives SIGTERM or the configured timeout expires
- THEN the CLI MUST exit cleanly with no error

### Requirement: Watch-Capable Commands

The following commands MUST support the `-w` / `--watch` and `-i` / `--interval` flags:

| Command | Description |
|---|---|
| `hf cluster list --table` | Cluster list as dot table with dynamic condition columns |
| `hf table` | Combined clusters + nodepools dot table |
| `hf cluster conditions -w -i <dur>` | Watch cluster conditions with configurable interval |
| `hf cluster statuses -w -i <dur>` | Watch cluster adapter statuses with configurable interval |
| `hf nodepool conditions -w -i <dur>` | Watch nodepool conditions with configurable interval |
| `hf nodepool statuses -w -i <dur>` | Watch nodepool adapter statuses with configurable interval |
| `hf nodepool list --table -w -i <dur>` | Nodepool table with watch support |
