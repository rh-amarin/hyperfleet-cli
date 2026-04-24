# Configuration Management Specification

## Purpose

Provide CLI commands to manage HyperFleet CLI configuration, including viewing, setting, clearing config properties, environment profile management, and diagnostic checks. Configuration is stored as individual files in `~/.config/hf/` with one file per property.

## Requirements

### Requirement: Show Configuration

The CLI SHALL display the current configuration values grouped by section when invoked with `hf config show` or `hf config` (no subcommand).

#### Scenario: Show full configuration with no arguments

- GIVEN the CLI is installed and config directory exists
- WHEN the user runs `hf config` with no arguments
- THEN the CLI MUST display help/usage text, environment list, and full configuration
- AND configuration MUST be grouped by sections: hyperfleet, maestro, portforward, database, rabbitmq, registry
- AND each property MUST show its key and current value
- AND secret properties (token, db-password, rabbitmq-password) MUST display `<set>` or `<not set>` instead of actual values

#### Scenario: Show configuration with show subcommand

- GIVEN the CLI is installed
- WHEN the user runs `hf config show`
- THEN the CLI MUST display the same configuration sections and values as `hf config`
- AND properties overridden by an active environment MUST be annotated with `[env-name]` suffix

#### Scenario: Show configuration for a specific environment

- GIVEN named environments exist in the config directory
- WHEN the user runs `hf config env show <name>`
- THEN the CLI MUST display configuration with environment-specific overrides highlighted
- AND overridden properties MUST show `[env-name]` suffix in a distinct color (cyan)

### Requirement: Set Configuration Value

The CLI SHALL allow setting individual configuration properties.

#### Scenario: Set a config property

- GIVEN the CLI is installed
- WHEN the user runs `hf config set <key> <value>`
- THEN the value MUST be written to `~/.config/hf/<key>`
- AND subsequent reads of that key MUST return the new value

### Requirement: Clear Configuration Value

The CLI SHALL allow clearing individual or all configuration properties.

#### Scenario: Clear a single property

- GIVEN a config property exists
- WHEN the user runs `hf config clear <key>`
- THEN the property file MUST be removed from the config directory
- AND subsequent reads MUST return the default value or null

#### Scenario: Clear all configuration

- GIVEN config properties exist
- WHEN the user runs `hf config clear all`
- THEN all property files MUST be removed from the config directory

### Requirement: Environment Profile Management

The CLI SHALL support named environment profiles that override base configuration.

#### Scenario: List environments

- GIVEN environment profiles exist as `<env-name>.<property>` files in the config directory
- WHEN the user runs `hf config env list`
- THEN the CLI MUST list all environments with property counts
- AND the active environment MUST be marked with a filled indicator (green)
- AND inactive environments MUST be marked with an empty indicator

#### Scenario: Activate an environment

- GIVEN a named environment exists
- WHEN the user runs `hf config env activate <name>`
- THEN the environment-specific property files MUST be copied over the base config
- AND subsequent config reads MUST reflect the activated environment values

### Requirement: Configuration Doctor

The CLI SHALL provide a diagnostic check of script readiness.

#### Scenario: Run doctor check

- GIVEN scripts exist with config requirements
- WHEN the user runs `hf config doctor`
- THEN the CLI MUST scan all scripts for required configuration
- AND display each script with a status indicator (green = all config present, red = missing config)
- AND report a summary count of ready vs. not-ready scripts

### Requirement: Interactive Bootstrap

The CLI SHALL provide an interactive setup wizard for initial configuration.

#### Scenario: Run bootstrap

- GIVEN the CLI is installed
- WHEN the user runs `hf config bootstrap [env-name]`
- THEN the CLI MUST prompt the user for each configuration property interactively
- AND save entered values to the config directory

### Requirement: Display Current Cluster ID

The CLI SHALL display the currently configured cluster ID.

#### Scenario: Show cluster ID

- GIVEN a cluster-id is set in config
- WHEN the user runs `hf cluster id`
- THEN the CLI MUST output the UUID string of the current cluster
- AND exit with code 0

#### Scenario: No cluster ID configured

- GIVEN no cluster-id is set in config
- WHEN the user runs `hf cluster id`
- THEN the CLI MUST exit with a non-zero exit code
- AND display an error message indicating cluster-id is not set

### Requirement: Display Current Nodepool ID

The CLI SHALL display the currently configured nodepool ID.

#### Scenario: Show nodepool ID

- GIVEN a nodepool-id is set in config
- WHEN the user runs `hf nodepool id`
- THEN the CLI MUST output the UUID string of the current nodepool
- AND exit with code 0
