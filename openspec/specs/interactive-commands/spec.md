# Interactive Commands Specification

## Purpose

Define the behavior of CLI commands that require interactive input, produce streaming output, or run as long-running processes. These commands cannot be fully automated in non-interactive environments but must be supported by the CLI tool.

## Requirements

### Requirement: Interactive Menu Selection

The CLI SHALL support interactive menu-based selection for certain commands.

#### Scenario: Kubectl context selection

- GIVEN kubectl is available
- WHEN the user runs `hf kube context`
- THEN the CLI MUST present an interactive menu of kubectl contexts
- AND after selection, present an interactive menu of namespaces
- AND save selections to config

#### Scenario: Maestro get without name

- GIVEN maestro is configured
- WHEN the user runs `hf maestro get` with no name argument
- THEN the CLI MUST list available resources and present an interactive selection menu
- AND display the selected resource's details

#### Scenario: Maestro delete without name

- GIVEN maestro is configured
- WHEN the user runs `hf maestro delete` with no name argument
- THEN the CLI MUST list available resources and present an interactive selection menu

### Requirement: Confirmation Prompts

The CLI SHALL require confirmation before destructive operations.

#### Scenario: Database delete confirmation

- GIVEN the user runs `hf db delete <table>`
- WHEN the command shows a preview of records to be deleted
- THEN the CLI MUST prompt the user to type `yes` to confirm
- AND only proceed with deletion after receiving `yes`
- AND abort if any other input is received

#### Scenario: Database delete-all confirmation

- GIVEN the user runs `hf db delete-all`
- WHEN the command shows record counts for all tables
- THEN the CLI MUST prompt the user to type `yes` to confirm
- AND delete in order: adapter_statuses, node_pools, clusters

### Requirement: Watch Mode

The CLI SHALL support live-updating watch mode for status and condition views.

#### Scenario: Watch cluster conditions

- GIVEN a cluster-id is set
- WHEN the user runs `hf cluster conditions -w`
- THEN the CLI MUST display conditions with automatic periodic refresh
- AND highlight changes between refreshes

#### Scenario: Watch cluster statuses

- GIVEN a cluster-id is set
- WHEN the user runs `hf cluster statuses -w`
- THEN the CLI MUST display adapter statuses with automatic periodic refresh

#### Scenario: Watch nodepool conditions

- GIVEN cluster-id and nodepool-id are set
- WHEN the user runs `hf nodepool conditions -w`
- THEN the CLI MUST display conditions with automatic periodic refresh

#### Scenario: Watch nodepool statuses

- GIVEN cluster-id and nodepool-id are set
- WHEN the user runs `hf nodepool statuses -w`
- THEN the CLI MUST display adapter statuses with automatic periodic refresh

### Requirement: Streaming Log Output

The CLI SHALL support streaming log output from Kubernetes pods.

#### Scenario: Stream pod logs

- GIVEN matching pods exist
- WHEN the user runs `hf logs <pattern> -f`
- THEN the CLI MUST continuously stream log output until interrupted

#### Scenario: Stream adapter logs

- GIVEN matching adapter pods exist
- WHEN the user runs `hf logs adapter <pattern> -f`
- THEN the CLI MUST continuously stream filtered log output

### Requirement: Interactive Shell Access

The CLI SHALL support creating debug pods with interactive shell access.

#### Scenario: Debug pod shell

- GIVEN a deployment exists
- WHEN the user runs `hf kube debug <deployment-name>`
- THEN the CLI MUST create a debug pod and exec into it with an interactive shell
- AND the shell session MUST persist until the user exits

### Requirement: Terminal UI

The CLI SHALL support launching terminal-based UIs.

#### Scenario: Maestro TUI

- GIVEN maestro is configured
- WHEN the user runs `hf maestro tui`
- THEN the CLI MUST launch the maestro-cli terminal UI
- AND the TUI MUST be fully interactive

### Requirement: Long-Running Workflows

The CLI SHALL support long-running workflow execution.

#### Scenario: Full adapter workflow

- GIVEN all required config is set
- WHEN the user runs `hf workflow`
- THEN the CLI MUST execute the full lifecycle: create cluster, wait for adapters, patch, wait, create nodepool, patch, wait, delete
- AND the workflow MUST run until completion

#### Scenario: API-only workflow

- GIVEN API config is set
- WHEN the user runs `hf workflow api-only`
- THEN the CLI MUST execute a simulated workflow with manual adapter status posts
- AND demonstrate the full API interaction flow

### Requirement: Interactive Database Configuration

The CLI SHALL support interactive database configuration.

#### Scenario: Configure database

- GIVEN the CLI is running
- WHEN the user runs `hf db config`
- THEN the CLI MUST prompt interactively for: host, port, database name, user, password
- AND save each value to the config directory
