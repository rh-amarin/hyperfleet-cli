# Spec: Repos

## Overview

`hf repos` lists GitHub repositories for the configured registry owner (user or org),
showing open PR count, default branch, and latest CI status for each repository.

## Command

```
hf repos [--registry <owner>] [--watch/-w] [--interval <duration>]
```

## Configuration

| Key | Source | Description |
|---|---|---|
| `registry.name` | `~/.config/hf/config.yaml` | GitHub user or org name (owner) |
| `registry.token` | `~/.config/hf/config.yaml` | GitHub personal access token (secret) |

## Token Resolution

1. `GITHUB_TOKEN` environment variable
2. `registry.token` config key

## Output

Table with columns: `NAME`, `OPEN PRS`, `DEFAULT BRANCH`, `CI STATUS`.

```
NAME                   OPEN PRS  DEFAULT BRANCH  CI STATUS
hyperfleet-cli         3         main            success
hyperfleet-api         0         main            failure
hyperfleet-operator    1         main            -
```

## CI Status Values

| Value | Meaning |
|---|---|
| `success` | Latest workflow run completed successfully |
| `failure` | Latest workflow run failed |
| `cancelled` | Latest workflow run was cancelled |
| `skipped` | Latest workflow run was skipped |
| `in_progress` | Workflow run is still running |
| `queued` | Workflow run is queued |
| `-` | No workflow runs found |

## Watch Mode

`--watch/-w` enables periodic refresh. `--interval` sets the refresh interval (default: 5s).
Terminal is cleared before each refresh. Footer shows "Last updated: HH:MM:SS  (Ctrl+C to stop)".

## GitHub API Endpoints Used

- `GET /orgs/{owner}/repos` or `GET /users/{owner}/repos` — list repositories
- `GET /repos/{owner}/{repo}/pulls?state=open` — count open PRs
- `GET /repos/{owner}/{repo}/actions/runs?branch={default_branch}&per_page=1` — latest CI run
