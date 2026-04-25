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

---

# Delta spec: config

## REMOVED Requirements

### Requirement: hf-config-bootstrap

`hf config bootstrap [env-name]` is removed.

Migration: use `hf config env new [name]` instead.

---

### Requirement: hf-env-var-overrides

HF_* environment variable config overrides are removed: `HF_API_URL`, `HF_API_VERSION`,
`HF_TOKEN`, `HF_CONTEXT`, `HF_NAMESPACE`.

`HF_CONFIG_DIR` is retained (controls config directory location, not a config value).

Migration: use a named environment file and `hf config env activate`, or pass `--api-url`
/ `--api-token` CLI flags for one-off scripted overrides.

---

## MODIFIED Requirements

### Requirement: hf-config-show

`hf config show` (and `hf config` with no args) displays the resolved configuration.

#### Scenario: Active environment set

GIVEN an active environment is configured in state.yaml
WHEN the user runs `hf config show`
THEN the active environment name is printed at the top
AND config values are shown grouped by section with no source annotations
AND secrets (token, database.password, rabbitmq.password) are shown as `<set>` or `<not set>`

#### Scenario: No active environment

GIVEN no active environment is configured
WHEN the user runs `hf config show`
THEN the command exits non-zero
AND prints:
```
error: no active environment
  → run 'hf config env new' to create one
  → run 'hf config env activate <name>' to activate an existing one
```

---

## ADDED Requirements

### Requirement: hf-config-env-new

`hf config env new [name]` creates a new named environment profile.

#### Scenario: Name provided as argument

GIVEN the user runs `hf config env new dev`
WHEN the command starts
THEN no name prompt is shown
AND the user is prompted for configuration values with defaults
AND a sparse YAML file is saved to `environments/dev.yaml`
AND the success message instructs the user to run `hf config env activate dev`

#### Scenario: Name not provided

GIVEN the user runs `hf config env new` with no argument
WHEN the command starts
THEN the user is prompted to enter an environment name
AND after entering the name the config value prompts follow
AND behaviour is otherwise identical to the name-provided scenario

#### Scenario: Prompt defaults

GIVEN the user presses enter on every prompt
THEN the saved environment file contains:
- database.user = hyperfleet
- database.name = hyperfleet
- database.password = foobar-bizz-buzz
- rabbitmq.host = rabbitmq
- rabbitmq.user = guest
- rabbitmq.password = guest

### Requirement: active-env-guard

Commands that require a configured target must fail when no active environment is set.

#### Scenario: Command requires active env, none set

GIVEN no active environment is configured
WHEN the user runs any of: `hf config show`, `hf config set`, `hf config clear <key>`, `hf config doctor`
THEN the command exits non-zero
AND prints the no-active-environment error with guidance

#### Scenario: Always-available commands

GIVEN no active environment is configured
WHEN the user runs any of: `hf config env list`, `hf config env new`, `hf config env activate`, `hf config env deactivate`, `hf config clear state`
THEN the command succeeds normally
