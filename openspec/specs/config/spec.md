# Delta for config

## MODIFIED Requirements

### Requirement: Show Configuration
The CLI SHALL display configuration grouped by YAML section (hyperfleet, kubernetes,
maestro, port-forward, database, rabbitmq, registry), not by flat key list.
Each value MUST be annotated with its source: `[default]`, `[config]`, `[env:<name>]`, `[ENV]`, or `[flag]`.

(Previously: "individual files in `~/.config/hf/` with one file per property" in Purpose.
Updated to reflect YAML storage model from config-model/spec.md.)

#### Scenario: Show full configuration with source annotations
- GIVEN config.yaml and state.yaml exist
- WHEN the user runs `hf config show`
- THEN the CLI MUST display all sections with current values and source annotations
- AND secret properties (hyperfleet.token, database.password, rabbitmq.password) MUST show `<set>` or `<not set>`

### Requirement: Set Configuration Value
The CLI SHALL allow setting individual configuration properties using dotted section.key notation.

#### Scenario: Set a config property
- GIVEN the CLI is installed
- WHEN the user runs `hf config set <section.key> <value>`
- THEN the value MUST be written into the correct section of `config.yaml`
- AND subsequent reads MUST return the new value with `[config]` annotation

(Previously: write to `~/.config/hf/<key>` flat file. Updated to YAML model.)

### Requirement: Clear Configuration Value

#### Scenario: Clear a single property
- GIVEN a config property has been set in config.yaml
- WHEN the user runs `hf config clear <section.key>`
- THEN the field MUST be reset to its default value in config.yaml

#### Scenario: Clear runtime state
- GIVEN runtime state exists in state.yaml
- WHEN the user runs `hf config clear state`
- THEN all fields in state.yaml MUST be cleared (cluster-id, nodepool-id, active-environment)
- AND config.yaml MUST NOT be affected

### Requirement: Interactive Bootstrap

#### Scenario: Run bootstrap without env name
- GIVEN the CLI is installed
- WHEN the user runs `hf config bootstrap`
- THEN the CLI MUST prompt for each config.yaml property interactively
- AND save entered values to config.yaml

#### Scenario: Run bootstrap with env name
- GIVEN the CLI is installed
- WHEN the user runs `hf config bootstrap <env-name>`
- THEN the CLI MUST prompt for each property interactively
- AND save entered values to `~/.config/hf/environments/<env-name>.yaml`
- AND print confirmation that the env was created and how to activate it
