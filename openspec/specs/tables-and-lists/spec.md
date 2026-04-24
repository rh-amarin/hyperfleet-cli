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

The CLI SHALL display clusters in a formatted table with dynamic condition columns.

#### Scenario: Empty table

- GIVEN no clusters exist
- WHEN the user runs `hf cluster table`
- THEN the CLI MUST output table headers only: `ID  NAME  GEN` with a separator line

#### Scenario: Populated table with conditions

- GIVEN clusters exist with various adapter conditions
- WHEN the user runs `hf cluster table`
- THEN the CLI MUST output a table with fixed columns: ID, NAME, GEN
- AND dynamic columns for each unique condition type found across all clusters (e.g., `Available`, `ClDeploymentSuccessful`, `ClJobSuccessful`, `ClNamespaceSuccessful`, `Ready`)
- AND the `Available` column MUST appear first among dynamic columns
- AND the `Ready` column MUST appear last
- AND status values MUST be displayed as colored dots:
  - Green dot = `True`
  - Red dot = `False`
  - Yellow dot = `Unknown`
  - `-` = condition not present for that cluster

### Requirement: NodePool Table View

The CLI SHALL display nodepools in the current cluster as a formatted table.

#### Scenario: Populated nodepool table

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool table`
- THEN the CLI MUST output a table with fixed columns: ID, NAME, REPLICAS, TYPE, GEN
- AND dynamic condition columns following the same pattern as cluster table
- AND REPLICAS MUST show the `spec.replicas` value
- AND TYPE MUST show the `spec.platform.type` value

### Requirement: Combined Overview Table

The CLI SHALL display a combined table of all clusters and their nested nodepools.

#### Scenario: Display combined table

- GIVEN clusters and nodepools exist
- WHEN the user runs `hf table`
- THEN the CLI MUST output a table with columns: ID, NAME, GEN, READY, ADAPTERS
- AND clusters MUST appear at the top level
- AND each cluster's nodepools MUST appear indented beneath it
- AND the ADAPTERS column MUST show the count of adapters that have reported status
- AND READY MUST show a colored dot based on the Ready condition status

#### Scenario: macOS compatibility

- GIVEN the CLI is running on macOS
- WHEN the combined table uses awk for sorting
- THEN the CLI SHOULD handle the absence of `asorti()` (GNU awk extension) gracefully
- AND still produce output even if sorting is degraded
