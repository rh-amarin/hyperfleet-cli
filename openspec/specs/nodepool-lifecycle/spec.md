# NodePool Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet nodepools. Nodepools are always scoped to a parent cluster, requiring a `cluster-id` to be set in config. All nodepool operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`.

## Requirements

### Requirement: Create NodePool

The CLI SHALL create one or more nodepools in the current cluster with configurable name, count, and instance type.

#### Scenario: Create multiple nodepools

- GIVEN a cluster-id is set in config and the API is reachable
- WHEN the user runs `hf nodepool create <name> [count] [instance-type]`
- THEN the CLI MUST create `count` nodepools by sending POST requests to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`
- AND each nodepool MUST be named `<name>-N` where N is a sequential number (1, 2, ...)
- AND each request payload MUST include:
  - `name`: `<name>-N`
  - `labels`: `{"counter": "N"}`
  - `spec`: `{"counter": "N", "platform": {"type": "<instance-type>"}, "replicas": 1}`
- AND the CLI MUST persist the LAST created nodepool's ID to active state via the shared `config.SetNodePoolID` function
- AND the response MUST include `owner_references` pointing to the parent cluster

#### Scenario: Create nodepool with default arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST use defaults: name=`my-nodepool`, count=`1`, instance_type=`m4`
- AND the CLI MUST NOT show a usage message — it MUST proceed with creation using defaults

#### Scenario: Initial nodepool status conditions

- GIVEN a nodepool was just created
- WHEN the API responds with the created nodepool
- THEN the nodepool MUST have initial conditions:
  - `Ready: False` with reason `MissingRequiredAdapters`
  - `Available: False` with reason `AdaptersNotAtSameGeneration`

### Requirement: List NodePools

The CLI SHALL list all nodepools in the current cluster.

#### Scenario: List nodepools

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf nodepool list [cluster_id]`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`
- AND output the response as JSON with shape `{"kind": "NodePoolList", "items": [...], "page": N, "size": N, "total": N}`

### Requirement: Search NodePool

The CLI SHALL search for a nodepool by name within the current cluster and set it as the current context.

#### Scenario: Search for existing nodepool

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST filter nodepools by name within the cluster
- AND output the matching nodepools as a JSON array
- AND persist the found nodepool's ID to active state via the shared `config.SetNodePoolID` function

### Requirement: Get NodePool

The CLI SHALL retrieve and display full details of a specific nodepool.

#### Scenario: Get current nodepool

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool get`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
- AND output the full nodepool JSON

#### Scenario: Get nodepool by explicit ID

- GIVEN a valid nodepool ID is provided
- WHEN the user runs `hf nodepool get <nodepool_id>`
- THEN the CLI MUST use the provided ID instead of the configured nodepool-id

### Requirement: Patch NodePool

The CLI SHALL increment a counter field in the nodepool's spec or labels section, triggering a generation bump.

#### Scenario: Patch spec counter

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool patch spec`
- THEN the CLI MUST fetch the current nodepool
- AND increment `spec.counter`
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
- AND the nodepool's generation MUST increment

#### Scenario: Patch labels counter

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool patch labels`
- THEN the CLI MUST follow the same pattern as spec but for `labels.counter`

### Requirement: Delete NodePool

The CLI SHALL delete a nodepool by ID.

#### Scenario: Delete nodepool

- GIVEN a nodepool exists
- WHEN the user runs `hf nodepool delete [nodepool_id]`
- THEN the CLI MUST send DELETE to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
- AND the response MUST include the full nodepool object with `deleted_by`, `deleted_time`, and incremented `generation`

#### Scenario: Delete current nodepool

- GIVEN nodepool-id is set in config and no explicit ID is provided
- WHEN the user runs `hf nodepool delete`
- THEN the CLI MUST use the configured nodepool-id

### Requirement: Get NodePool Conditions

The CLI SHALL display the generation and status conditions of a nodepool.

#### Scenario: Get conditions

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool conditions`
- THEN the CLI MUST fetch the nodepool and extract `generation` and `status.conditions` as JSON

#### Scenario: Watch conditions

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool conditions -w`
- THEN the CLI MUST display conditions with live updates using a watch mechanism

### Requirement: Get NodePool Conditions Table

The CLI SHALL display nodepool conditions in a formatted table.

#### Scenario: Display conditions table before adapters report

- GIVEN a nodepool exists with no adapter statuses
- WHEN the user runs `hf nodepool conditions table`
- THEN the CLI MUST output a table with columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE
- AND Ready and Available MUST show `False`

#### Scenario: Display conditions table after all adapters report

- GIVEN all required adapters have reported `Available=True` at the current generation
- WHEN the user runs `hf nodepool conditions table`
- THEN Ready and Available MUST show `True` (green)
- AND per-adapter conditions (e.g., `NpConfigmapSuccessful`) MUST appear as additional rows

### Requirement: Get NodePool Adapter Statuses

The CLI SHALL display adapter statuses for a nodepool.

#### Scenario: Get statuses

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool statuses`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}/adapter-statuses`
- AND output the `AdapterStatusList` response

#### Scenario: Watch statuses

- GIVEN cluster-id and nodepool-id are set
- WHEN the user runs `hf nodepool statuses -w`
- THEN the CLI MUST display statuses with live updates

### Requirement: Display NodePool Table

The CLI SHALL display nodepools in the current cluster as a formatted table with dynamic condition columns.

#### Scenario: Display nodepool table

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool table`
- THEN the CLI MUST output a table with columns: ID, NAME, REPLICAS, TYPE, GEN, Available, [dynamic condition columns], Ready
- AND status values MUST be displayed as colored dots: green=True, red=False, yellow=Unknown, `-`=not present
- AND dynamic columns MUST appear based on which conditions exist across all nodepools
