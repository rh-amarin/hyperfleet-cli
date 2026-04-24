# Configuration Registry Specification

## Purpose

Define the complete configuration property registry, storage model, environment variable overrides, and environment profile system that underpins all CLI commands.

## Requirements

### Requirement: Configuration Storage

The CLI SHALL store configuration as individual files in a dedicated config directory.

#### Scenario: File-per-property storage

- GIVEN the CLI is installed
- WHEN any configuration property is set
- THEN the value MUST be stored as a single file at `~/.config/hf/<key>`
- AND the file content MUST be the raw value (no formatting or wrappers)
- AND reading the property MUST return the file content

### Requirement: Environment Variable Overrides

The CLI SHALL support environment variable overrides for specific configuration properties.

#### Scenario: Environment variable takes precedence

- GIVEN a config property has both a file value and an environment variable set
- WHEN the CLI reads the property
- THEN the environment variable MUST take precedence over the file value

#### Scenario: Supported environment variables

- GIVEN the following environment variable mappings exist:
  - `HF_API_URL` overrides `api-url`
  - `HF_API_VERSION` overrides `api-version`
  - `HF_TOKEN` overrides `token`
  - `HF_CONTEXT` overrides `context`
  - `HF_NAMESPACE` overrides `namespace`
- WHEN any of these environment variables are set
- THEN the corresponding config property MUST use the environment variable value

### Requirement: Configuration Property Registry

The CLI SHALL support the following configuration properties organized by section.

#### Scenario: HyperFleet section properties

- GIVEN the CLI is initialized
- THEN the following properties MUST be supported with their defaults:
  - `api-url`: default `http://localhost:8000`
  - `api-version`: default `v1`
  - `token`: default empty, flag=secret
  - `context`: default null
  - `namespace`: default null
  - `gcp-project`: default `hcm-hyperfleet`
  - `cluster-id`: default null
  - `cluster-name`: default null
  - `nodepool-id`: default null

#### Scenario: Maestro section properties

- GIVEN the CLI is initialized
- THEN the following properties MUST be supported:
  - `maestro-consumer`: default `cluster1`
  - `maestro-http-endpoint`: default `http://localhost:8100`
  - `maestro-grpc-endpoint`: default `localhost:8090`
  - `maestro-namespace`: default `maestro`

#### Scenario: Port forward section properties

- GIVEN the CLI is initialized
- THEN the following properties MUST be supported:
  - `pf-api-port`: default `8000`
  - `pf-pg-port`: default `5432`
  - `pf-maestro-http-port`: default `8100`
  - `pf-maestro-http-remote-port`: default `8000`
  - `pf-maestro-grpc-port`: default `8090`

#### Scenario: Database section properties

- GIVEN the CLI is initialized
- THEN the following properties MUST be supported:
  - `db-host`: default `localhost`
  - `db-port`: default `5432`
  - `db-name`: default `hyperfleet`
  - `db-user`: default `hyperfleet`
  - `db-password`: default `foobar-bizz-buzz`, flag=secret

#### Scenario: RabbitMQ section properties

- GIVEN the CLI is initialized
- THEN the following properties MUST be supported:
  - `rabbitmq-host`: default `localhost`
  - `rabbitmq-mgmt-port`: default `15672`
  - `rabbitmq-user`: default `guest`
  - `rabbitmq-password`: default `guest`, flag=secret
  - `rabbitmq-vhost`: default `/`

#### Scenario: Registry section properties

- GIVEN the CLI is initialized
- THEN the following properties MUST be supported:
  - `registry`: default `$USER` (expands to current username)

### Requirement: Secret Property Handling

The CLI SHALL mask secret property values in display output.

#### Scenario: Display secret properties

- GIVEN a property is flagged as secret (token, db-password, rabbitmq-password)
- WHEN the CLI displays configuration
- THEN the property value MUST be shown as `<set>` if a value exists
- AND shown as `<not set>` if no value is configured
- AND the actual value MUST NOT be displayed

### Requirement: Environment Profiles

The CLI SHALL support named environment profiles for managing multiple deployment configurations.

#### Scenario: Environment file naming

- GIVEN an environment named `kind` has overrides for `context` and `namespace`
- WHEN environment files are stored
- THEN they MUST be stored as `~/.config/hf/kind.context` and `~/.config/hf/kind.namespace`

#### Scenario: Activate environment

- GIVEN an environment profile exists
- WHEN the user runs `hf config env activate <name>`
- THEN the environment-specific files MUST be applied to the base config
- AND the active environment name MUST be tracked

#### Scenario: Multiple environments

- GIVEN multiple environment profiles exist (e.g., `local`, `gke-amarin`, `kind`, `compose`)
- WHEN the user lists environments
- THEN each environment MUST show its name and property count
- AND only one environment MAY be active at a time

### Requirement: Per-Script Config Requirements

Each CLI command SHALL declare its required configuration properties.

#### Scenario: Config requirement checking

- GIVEN a script requires specific config properties (e.g., `api-url`, `api-version`, `cluster-id`)
- WHEN the script is invoked and a required property is not set
- THEN the CLI MUST display an error indicating which property is missing
- AND exit with a non-zero exit code

#### Scenario: Doctor checks all scripts

- GIVEN the `hf config doctor` command is run
- WHEN it scans all scripts for required config
- THEN it MUST report each script's readiness based on whether all its required properties are set
