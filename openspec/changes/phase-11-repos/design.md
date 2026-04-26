# Design: Phase 11 — Repos

## Command Tree

```
hf repos [--registry <owner>] [--watch/-w] [--interval <duration>]
```

## New Files

- `internal/repos/client.go` — package `repos`
- `internal/repos/client_test.go`
- `cmd/repos.go`
- `cmd/repos_test.go`

## Config Changes (additive)

Add `Token string` to `RegistryConfig` in `internal/config/types.go`.
Add `registry.token` cases to `getField`, `setField`, and `AllPaths` in `internal/config/fields.go`.
Mark `registry.token` as a secret in `secretPaths`.

## internal/repos/client.go

```go
type Client struct {
    gh *github.Client
}

func New(token string) *Client

type RepoInfo struct {
    Name          string
    OpenPRs       int
    DefaultBranch string
    CIStatus      string
}

func (c *Client) ListRepos(ctx context.Context, owner string) ([]RepoInfo, error)
func (c *Client) SetBaseURL(rawURL string) error  // for testing
```

### New(token)

Uses `github.NewClient(nil).WithAuthToken(token)` when token is non-empty;
otherwise `github.NewClient(nil)` for unauthenticated access.

### ListRepos(ctx, owner)

1. Try `Repositories.ListByOrg(ctx, owner, opts)` with `PerPage=100`, paginating.
2. If that fails, fall back to `Repositories.List(ctx, owner, opts)` (user repos).
3. For each repo: count open PRs via `PullRequests.List(ctx, owner, name, {state:"open", PerPage:100})`.
4. For each repo: fetch latest CI run via `Actions.ListRepositoryWorkflowRuns(ctx, owner, name, {Branch: defaultBranch, PerPage:1})`.
   - CI status = `conclusion` if non-empty, else `status`, else `"-"`.

### SetBaseURL(rawURL)

Sets `c.gh.BaseURL` to the parsed URL (ensures trailing slash). Used by tests to redirect to httptest.Server.

## cmd/repos.go

```go
func init() { rootCmd.AddCommand(reposCmd) }

var reposCmd = &cobra.Command{
    Use:          "repos",
    Short:        "List GitHub repositories for the configured registry owner",
    SilenceUsage: true,
    RunE:         runRepos,
}
```

### Flags

| Flag | Type | Default | Source |
|---|---|---|---|
| `--registry` | string | `cfgStore.Cfg().Registry.Name` | config key `registry.name` |
| `--watch/-w` | bool | false | — |
| `--interval` | duration | 5s | — |
| `--github-api-url` | string | `""` | hidden, for tests only |

### Token Resolution

```go
token := os.Getenv("GITHUB_TOKEN")
if token == "" {
    token = cfgStore.Cfg().Registry.Token
}
```

### Output

Table columns: `NAME`, `OPEN PRS`, `DEFAULT BRANCH`, `CI STATUS`.
Uses `printer().PrintTable(headers, rows)`.

### Watch Mode

When `--watch` is set: clear terminal, print table, print timestamp footer, wait `--interval`, repeat until SIGINT.

## Unit Test Design

### internal/repos/client_test.go

All tests use `httptest.NewServer` and `client.SetBaseURL(srv.URL+"/")`.

The mock server handles:
- `GET /orgs/{owner}/repos` → returns JSON array of repo objects
- `GET /repos/{owner}/{repo}/pulls` → returns JSON array of PRs
- `GET /repos/{owner}/{repo}/actions/runs` → returns JSON workflow runs

Tests:
- `TestListRepos_ReturnsRepoInfo`: 1 repo, 2 open PRs, 1 CI run with conclusion=success
- `TestListRepos_NoPRsNoCIRuns`: repo with empty PR list and empty runs
- `TestListRepos_UserFallback`: `/orgs/…/repos` returns 404; `/users/…/repos` returns repos
- `TestListRepos_CIStatusInProgress`: run with status=in_progress, empty conclusion

### cmd/repos_test.go

Uses `runCmdRaw` with `--github-api-url` pointing to httptest.Server.

Tests:
- `TestReposCmd_RendersTable`: 2 repos; verify NAME/OPEN PRS/DEFAULT BRANCH/CI STATUS headers and row content
- `TestReposCmd_RegistryFlag`: `--registry override-owner` overrides config

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Dependency | `go-github/v60` only, no oauth2 | `WithAuthToken` in v53+ handles token auth without oauth2 |
| Org/user detection | Try org first, fall back to user | Works transparently for both |
| PR count | Paginate `PullRequests.List` | Most accurate; no GitHub Search API rate limits |
| CI status | Latest workflow run conclusion | Matches bash script behavior |
| Hidden `--github-api-url` | Hidden cobra flag | Allows cmd-level tests without subprocess |
| Watch interval default | 5s | GitHub API rate limits make 2s unsafe |
