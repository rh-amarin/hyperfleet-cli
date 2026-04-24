# Maestro Operations Specification

## Purpose

Provide CLI commands for managing Maestro resources, which are used by HyperFleet adapters to deploy Kubernetes manifests to managed clusters. Maestro provides both a CLI tool (`maestro-cli`) and an HTTP API for resource management.

## Requirements

### Requirement: List Maestro Resources

The CLI SHALL list maestro resources via the maestro-cli tool.

#### Scenario: List resources

- GIVEN maestro-consumer and maestro-http-endpoint are configured
- WHEN the user runs `hf maestro list`
- THEN the CLI MUST invoke maestro-cli to list resources
- AND output a JSON array where each item contains:
  - `id`: UUID
  - `name`: resource name (e.g., `mw-<cluster-uuid>`)
  - `consumerName`: the consumer (e.g., `cluster1`)
  - `version`: integer version number
  - `manifestCount`: number of Kubernetes manifests
  - `manifests`: array of `{kind, name, namespace}` summaries
  - `conditions`: array of `{type, status, reason}` (Applied, Available)

### Requirement: List Maestro Resource Bundles

The CLI SHALL list maestro resource bundles via the HTTP API.

#### Scenario: List bundles

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro bundles`
- THEN the CLI MUST send GET to `/api/maestro/v1/resource-bundles`
- AND output the JSON response containing:
  - `kind`: `ResourceBundleList`
  - `items`: array of resource bundles with full Kubernetes manifests, manifest_configs, metadata (labels/annotations), and per-resource feedback status

### Requirement: List Maestro Consumers

The CLI SHALL list maestro consumers via the HTTP API.

#### Scenario: List consumers

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro consumers`
- THEN the CLI MUST send GET to `/api/maestro/v1/consumers`
- AND output the JSON response with shape `{"items": [{id, kind: "Consumer", name}], "kind": "ConsumerList", "total": N}`

### Requirement: Get Maestro Resource

The CLI SHALL retrieve a specific maestro resource by name.

#### Scenario: Get by name

- GIVEN maestro-consumer, maestro-http-endpoint, and maestro-grpc-endpoint are configured
- WHEN the user runs `hf maestro get <name>`
- THEN the CLI MUST retrieve the specified resource via maestro-cli

#### Scenario: Get with interactive selection

- GIVEN no name argument is provided
- WHEN the user runs `hf maestro get`
- THEN the CLI MUST list available resources and present an interactive selection menu

### Requirement: Delete Maestro Resource

The CLI SHALL delete a maestro resource by name.

#### Scenario: Delete by name

- GIVEN maestro configuration is set
- WHEN the user runs `hf maestro delete <name>`
- THEN the CLI MUST delete the specified resource via maestro-cli

#### Scenario: Delete with interactive selection

- GIVEN no name argument is provided
- WHEN the user runs `hf maestro delete`
- THEN the CLI MUST list available resources and present an interactive selection menu

### Requirement: Maestro Terminal UI

The CLI SHALL launch the maestro-cli terminal UI for interactive resource management.

#### Scenario: Launch TUI

- GIVEN maestro-http-endpoint is configured
- WHEN the user runs `hf maestro tui`
- THEN the CLI MUST launch `maestro-cli tui --api-server={endpoint}`
- AND present a full terminal UI for resource management
