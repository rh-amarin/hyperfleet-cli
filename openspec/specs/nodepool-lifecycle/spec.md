# NodePool Lifecycle Specification

## Purpose

Provide CLI commands for full CRUD lifecycle management of HyperFleet nodepools. Nodepools are always scoped to a parent cluster, requiring a `cluster-id` to be set in config. All nodepool operations interact with the HyperFleet API at `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`.

## Prerequisites

**cluster-id required**: All nodepool commands require `cluster-id` to be set in state. If it is not set, the CLI MUST display:
```
[ERROR] No cluster-id set in state. Run 'hf cluster create' or 'hf cluster search <name>' first.
```
AND exit with code 1 before making any API call.

**nodepool-id required for single-resource commands**: `hf nodepool patch`, `hf nodepool delete`, `hf nodepool conditions`, and `hf nodepool statuses` additionally require `nodepool-id` to be set in state (unless an explicit ID argument is provided). If cluster-id is set but nodepool-id is not, the CLI MUST display:
```
[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.
```
AND exit with code 1.

## Requirements

### Requirement: Create NodePool

The CLI SHALL create one or more nodepools in the current cluster with configurable name, count, and instance type.

#### Scenario: Create multiple nodepools

- GIVEN a cluster-id is set in config and the API is reachable
- WHEN the user runs `hf nodepool create <name> [count] [instance-type]`
- THEN the CLI MUST create `count` nodepools by sending POST requests to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools`
- AND each nodepool MUST be named `<name>-N` where N starts at 1 and increments for each additional nodepool (e.g., `my-nodepool-1`, `my-nodepool-2`, `my-nodepool-3` for count=3)
- AND each request payload MUST include:
  - `name`: `<name>-N`
  - `labels`: `{"counter": "N"}`
  - `spec`: `{"counter": "N", "platform": {"type": "<instance-type>"}, "replicas": 1}`
- AND the CLI MUST persist the LAST created nodepool's ID to active state via `config.SetNodePoolID`
- AND the CLI MUST print `[INFO] NodePool context set to '<id>'` on stderr after persisting
- AND the response MUST include `owner_references` pointing to the parent cluster

#### Scenario: Create nodepool with default arguments

- GIVEN no arguments are provided
- WHEN the user runs `hf nodepool create`
- THEN the CLI MUST use defaults: name=`my-nodepool`, count=`1`, instance_type=`m4`
- AND the CLI MUST NOT show a usage message — it MUST proceed with creation using defaults

#### Scenario: Invalid count argument

- GIVEN a count value less than 1 or not a valid integer is provided
- WHEN the user runs `hf nodepool create <name> <invalid-count>`
- THEN the CLI MUST display `[ERROR] count must be a positive integer (minimum 1)`
- AND display the usage message
- AND exit with code 1

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
- WHEN the user runs `hf nodepool list`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools` using the cluster-id from state
- AND output the response as JSON with shape `{"kind": "NodePoolList", "items": [...], "page": N, "size": N, "total": N}`

### Requirement: Search NodePool

The CLI SHALL search for a nodepool by name within the current cluster and set it as the current context.

#### Scenario: Search with no arguments

- GIVEN a nodepool-id is set in config
- WHEN the user runs `hf nodepool search` with no arguments
- THEN the CLI MUST behave identically to `hf nodepool get` — fetching and returning the current nodepool from state

#### Scenario: Search with no arguments and no nodepool in state

- GIVEN no nodepool-id is set in state
- WHEN the user runs `hf nodepool search` with no arguments
- THEN the CLI MUST display `[ERROR] No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.`
- AND exit with code 1

#### Scenario: Search for existing nodepool

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST filter nodepools by name within the cluster
- AND output the matching nodepools as a JSON array of full NodePool objects
- AND persist the found nodepool's ID to active state via `config.SetNodePoolID`
- AND print `[INFO] NodePool context set to '<id>'` on stderr after persisting

#### Scenario: Search for non-existent nodepool

- GIVEN no nodepool matches the search name within the cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST display `[WARN] No nodepools found matching '<name>'`
- AND output an empty JSON array `[]`
- AND exit with code 0

#### Scenario: Multiple matches

- GIVEN multiple nodepools match the search name within the cluster
- WHEN the user runs `hf nodepool search <name>`
- THEN the CLI MUST display `[WARN] Multiple nodepools found matching '<name>', using first result`
- AND set nodepool-id to the first element in the returned `items` array
- AND persist that nodepool-id to active state

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

#### Scenario: Patch with no arguments

- GIVEN the user provides no arguments
- WHEN the user runs `hf nodepool patch`
- THEN the CLI MUST display usage: `Usage: hf nodepool patch {spec|labels} [nodepool_id]`
- AND exit with code 1

#### Scenario: Patch spec counter

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool patch spec`
- THEN the CLI MUST fetch the current nodepool
- AND read the current `spec.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}` with the incremented counter as a string
- AND display `[INFO] Incrementing spec.counter: <old> -> <new>` where `<old>` and `<new>` are integer strings (e.g., `1 -> 2`; first increment displays `0 -> 1`)
- AND the nodepool's generation MUST increment

#### Scenario: Patch labels counter

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool patch labels`
- THEN the CLI MUST fetch the current nodepool
- AND read the current `labels.counter` value as an integer (if absent, treat as `0`)
- AND increment it by 1
- AND send a PATCH to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}` with the incremented counter as a string
- AND display `[INFO] Incrementing labels.counter: <old> -> <new>`
- AND the nodepool's generation MUST increment

### Requirement: Delete NodePool

The CLI SHALL delete a nodepool by ID.

#### Scenario: Delete nodepool

- GIVEN a nodepool exists
- WHEN the user runs `hf nodepool delete [nodepool_id]`
- THEN the CLI MUST send DELETE to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
- AND the response MUST include the full nodepool object with `deleted_by`, `deleted_time`, and incremented `generation`
- AND the CLI MUST output the deleted nodepool object subject to the `--output` flag (default: JSON)

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


### Requirement: Get NodePool Conditions Table

The CLI SHALL display nodepool conditions in a formatted table via the `--table` flag.

#### Scenario: Display conditions table before adapters report

- GIVEN a nodepool exists with no adapter statuses
- WHEN the user runs `hf nodepool conditions --table`
- THEN the CLI MUST output a table with columns: TYPE, STATUS, LAST TRANSITION, REASON, MESSAGE
- AND Ready and Available MUST show `False`

#### Scenario: Display conditions table after all adapters report

- GIVEN all required adapters have reported `Available=True` at the current generation
- WHEN the user runs `hf nodepool conditions --table`
- THEN Ready and Available MUST show `True` (green)
- AND per-adapter conditions (e.g., `NpConfigmapSuccessful`) MUST appear as additional rows

### Requirement: Get NodePool Adapter Statuses

The CLI SHALL display adapter statuses for a nodepool.

#### Scenario: Get statuses

- GIVEN cluster-id and nodepool-id are set in config
- WHEN the user runs `hf nodepool statuses`
- THEN the CLI MUST send GET to `/api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}/adapter-statuses`
- AND output the `AdapterStatusList` response


### Requirement: Display NodePool Table

The CLI SHALL display nodepools in the current cluster as a formatted table when the `--table` flag is passed to `hf nodepool list`.

#### Scenario: Display nodepool table

- GIVEN nodepools exist in the current cluster
- WHEN the user runs `hf nodepool list --table`
- THEN the CLI MUST output a table with columns: ID, NAME, REPLICAS, TYPE, GEN, Available, [dynamic condition columns], Ready
- AND status values MUST be displayed as colored dots: green=True, red=False, yellow=Unknown, `-`=not present
- AND dynamic columns MUST appear based on which conditions exist across all nodepools
