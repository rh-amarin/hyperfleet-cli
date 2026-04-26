# Proposal: Phase 11 — Repos

## Intent

Implement `hf repos` — a GitHub repository status overview for the configured registry owner,
replacing `hf.repos.sh`. Lists repos with open PR count, default branch, and latest CI status.

## Scope In

- `internal/repos` package: GitHub API client wrapping `go-github/v60`
- `cmd/repos.go`: `hf repos` command with table output and watch mode
- `registry.token` config key added to `internal/config` (additive)

## Scope Out

- No write operations to GitHub
- No GitHub authentication flows (token must be pre-configured)
- Private repo listing is supported if the token has the right scopes

## Testing Scope

| Package | Test Cases |
|---|---|
| `internal/repos` | `TestListRepos_ReturnsRepoInfo`: mock org repos endpoint, PR list, CI runs; assert correct `RepoInfo` struct |
| `internal/repos` | `TestListRepos_CIStatusMapping`: assert `success`, `failure`, `in_progress`, and empty conclusion all map correctly |
| `internal/repos` | `TestListRepos_UserFallback`: ListByOrg 404 → falls back to user listing |
| `cmd` | `TestReposCmd_RendersTable`: full command run against mock GitHub server; assert table columns present |
| `cmd` | `TestReposCmd_RegistryFlag`: `--registry` overrides config value |

## Live Verification

- Requires `GITHUB_TOKEN` env var with read access to `rh-amarin` org
- `hf repos --registry rh-amarin` → table of repos
- `hf repos --registry rh-amarin --watch --interval 5s` → refreshes 2 cycles then Ctrl-C
