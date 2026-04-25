# Configuration Model Specification

## Purpose

Define the split YAML configuration model for the HyperFleet CLI. Configuration is divided into two files: a static configuration file for connection properties and settings that rarely change, and an active state file for transient runtime state like the current cluster or nodepool selection. This replaces the file-per-property model from the shell scripts while maintaining backwards compatibility during migration.

## Requirements

### Requirement: Split Configuration Files

The CLI SHALL use two YAML files stored in `~/.config/hf/`.

#### Scenario: Static configuration file

- GIVEN the CLI is initialized
- WHEN configuration is loaded
- THEN the static config MUST be stored at `~/.config/hf/config.yaml`
- AND it MUST contain the following sections with their properties:
  ```yaml
  hyperfleet:
    api-url: "http://localhost:8000"
    api-version: "v1"
    token: ""
    gcp-project: "hcm-hyperfleet"

  kubernetes:
    context: ""
    namespace: ""

  maestro:
    consumer: "cluster1"
    http-endpoint: "http://localhost:8100"
    grpc-endpoint: "localhost:8090"
    namespace: "maestro"

  port-forward:
    api-port: 8000
    pg-port: 5432
    maestro-http-port: 8100
    maestro-http-remote-port: 8000
    maestro-grpc-port: 8090

  database:
    host: "localhost"
    port: 5432
    name: "hyperfleet"
    user: "hyperfleet"
    password: "foobar-bizz-buzz"

  rabbitmq:
    host: "localhost"
    mgmt-port: 15672
    user: "guest"
    password: "guest"
    vhost: "/"

  registry:
    name: ""  # defaults to $USER at runtime
  ```

#### Scenario: Active state file

- GIVEN the CLI is managing runtime state
- WHEN state changes occur (e.g., cluster selection)
- THEN the state MUST be stored at `~/.config/hf/state.yaml`
- AND properties MUST be top-level (flat, not nested):
  ```yaml
  active-environment: "kind"
  cluster-id: "019dbf43-65c5-7562-9077-e0a2331a1070"
  cluster-name: "test-1e317d46"
  nodepool-id: "019dbf43-7199-7ea6-b786-d617fc793c28"
  ```
- AND the file MUST be updated atomically (write to temp, then rename)

#### Scenario: Config directory creation

- GIVEN the config directory does not exist
- WHEN the CLI is first invoked
- THEN the CLI MUST create `~/.config/hf/` if it does not exist
- AND create `config.yaml` with default values if it does not exist
- AND create `state.yaml` as an empty file if it does not exist

### Requirement: Configuration Precedence

The CLI SHALL resolve configuration values with a defined precedence order.

#### Scenario: Precedence chain

- GIVEN multiple sources may define the same property
- WHEN the CLI resolves a configuration value
- THEN the precedence order MUST be (highest to lowest):
  1. CLI flags (`--api-url`, `--api-token`)
  2. Environment variables (`HF_API_URL`, `HF_API_VERSION`, `HF_TOKEN`, `HF_CONTEXT`, `HF_NAMESPACE`)
  3. Active environment overrides (from `environments/<name>.yaml`)
  4. Static config file (`config.yaml`)
  5. Built-in defaults

### Requirement: Environment Profiles

The CLI SHALL support named environment profiles that override static configuration.

#### Scenario: Environment file storage

- GIVEN environment profiles exist
- WHEN environments are stored
- THEN each environment MUST be stored as `~/.config/hf/environments/<name>.yaml`
- AND the file MUST contain only the properties that differ from the base config
- AND the file MUST use the same YAML structure as `config.yaml` (nested sections)
  ```yaml
  # ~/.config/hf/environments/gke-production.yaml
  hyperfleet:
    api-url: "https://hyperfleet.prod.internal:8000"
    token: "prod-token-xxx"
  kubernetes:
    context: "gke_project_region_cluster"
    namespace: "hyperfleet-prod"
  database:
    host: "db.prod.internal"
    password: "prod-password"
  ```

#### Scenario: List environments

- GIVEN environment profiles exist in `~/.config/hf/environments/`
- WHEN the user runs `hf config env list`
- THEN the CLI MUST list each environment by filename (without `.yaml`)
- AND show the count of overridden properties
- AND mark the active environment (from `state.yaml` `active-environment`) with a distinct indicator

#### Scenario: Activate environment

- GIVEN a named environment exists
- WHEN the user runs `hf config env activate <name>`
- THEN the CLI MUST set `active-environment: <name>` in `state.yaml`
- AND subsequent config reads MUST merge the environment overrides on top of the base config
- AND the CLI MUST NOT modify `config.yaml` — overrides are applied at runtime

#### Scenario: Deactivate environment

- GIVEN an environment is active
- WHEN the user runs `hf config env activate ""` or `hf config env deactivate`
- THEN the CLI MUST clear `active-environment` from `state.yaml`
- AND subsequent config reads MUST use only the base config and defaults

### Requirement: Environment Variable Overrides

The CLI SHALL support environment variable overrides for key configuration properties.

#### Scenario: Supported environment variables

- GIVEN the following mappings exist:
  | Environment Variable | Config Path |
  |---------------------|-------------|
  | `HF_API_URL` | `hyperfleet.api-url` |
  | `HF_API_VERSION` | `hyperfleet.api-version` |
  | `HF_TOKEN` | `hyperfleet.token` |
  | `HF_CONTEXT` | `kubernetes.context` |
  | `HF_NAMESPACE` | `kubernetes.namespace` |
- WHEN any of these environment variables are set
- THEN the corresponding config value MUST use the environment variable
- AND the environment variable MUST take precedence over file-based config and environment profiles

### Requirement: State Management

The CLI SHALL manage active state separately from configuration.

#### Scenario: Set cluster context

- GIVEN a cluster is created or found via search
- WHEN the CLI updates the active cluster
- THEN `cluster-id` and `cluster-name` MUST be updated in `state.yaml`
- AND the static `config.yaml` MUST NOT be modified

#### Scenario: Set nodepool context

- GIVEN a nodepool is created or found via search
- WHEN the CLI updates the active nodepool
- THEN `nodepool-id` MUST be updated in `state.yaml`

#### Scenario: Clear state

- GIVEN the user wants to reset state
- WHEN the user runs `hf config clear state`
- THEN all properties in `state.yaml` MUST be cleared
- AND `config.yaml` MUST NOT be affected

### Requirement: Secret Handling

The CLI SHALL protect sensitive configuration values.

#### Scenario: Display secrets

- GIVEN a property is a secret (token, database.password, rabbitmq.password)
- WHEN `hf config show` displays the property
- THEN the value MUST be shown as `<set>` if non-empty or `<not set>` if empty
- AND the actual value MUST NOT be displayed in config show output

#### Scenario: Secrets in config file

- GIVEN secrets are stored in `config.yaml`
- WHEN the file is written
- THEN the CLI SHOULD warn the user about file permissions
- AND the config file SHOULD be created with mode `0600` (owner read/write only)

### Requirement: Config Show with Source Annotation

The CLI SHALL annotate displayed config values with their source.

#### Scenario: Show config with sources

- GIVEN config values come from multiple sources
- WHEN the user runs `hf config show`
- THEN each value MUST be annotated with its source:
  - `[default]` for built-in defaults
  - `[config]` for values from `config.yaml`
  - `[env:<name>]` for environment profile overrides
  - `[ENV]` for environment variable overrides
  - `[flag]` for CLI flag overrides (when applicable)
- AND the source annotation MUST appear after the value in a distinct color

### Requirement: Migration from File-Per-Property

The CLI SHALL support one-time migration from the legacy file-per-property config format.

#### Scenario: Auto-detect legacy config

- GIVEN individual property files exist in `~/.config/hf/` (e.g., `api-url`, `cluster-id`)
- WHEN the CLI starts and no `config.yaml` exists
- THEN the CLI MUST detect the legacy format
- AND offer to migrate by reading all property files and writing `config.yaml` and `state.yaml`
- AND upon successful migration, rename the legacy files to `~/.config/hf/legacy/` for backup

#### Scenario: No legacy config

- GIVEN no legacy files exist and no `config.yaml` exists
- WHEN the CLI starts
- THEN the CLI MUST create `config.yaml` with defaults
- AND create an empty `state.yaml`

### Requirement: Config File Path Override

The CLI SHALL support overriding the config directory location.

#### Scenario: Custom config path

- GIVEN the `--config` flag or `HF_CONFIG_DIR` environment variable is set
- WHEN the CLI loads configuration
- THEN it MUST look for `config.yaml`, `state.yaml`, and `environments/` in the specified directory instead of `~/.config/hf/`
