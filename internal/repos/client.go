package repos

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/v68/github"
)

// Client wraps the GitHub API client.
type Client struct {
	gh *github.Client
}

// New creates a GitHub client. When token is non-empty, requests are authenticated.
func New(token string) *Client {
	gh := github.NewClient(nil)
	if token != "" {
		gh = gh.WithAuthToken(token)
	}
	return &Client{gh: gh}
}

// SetBaseURL overrides the GitHub API base URL (used in tests to point at httptest.Server).
// rawURL must end with "/".
func (c *Client) SetBaseURL(rawURL string) error {
	if !strings.HasSuffix(rawURL, "/") {
		rawURL += "/"
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid github api url %q: %w", rawURL, err)
	}
	c.gh.BaseURL = u
	return nil
}

// RepoInfo holds summary information for a single repository.
type RepoInfo struct {
	Name          string
	OpenPRs       int
	DefaultBranch string
	CIStatus      string
}

// ListRepos lists repositories for owner (user or org) and enriches each with
// open PR count and latest CI workflow run status.
func (c *Client) ListRepos(ctx context.Context, owner string) ([]RepoInfo, error) {
	ghRepos, err := c.fetchAllRepos(ctx, owner)
	if err != nil {
		return nil, err
	}

	result := make([]RepoInfo, 0, len(ghRepos))
	for _, repo := range ghRepos {
		info := RepoInfo{
			Name:          repo.GetName(),
			DefaultBranch: repo.GetDefaultBranch(),
		}
		if info.DefaultBranch == "" {
			info.DefaultBranch = "main"
		}

		info.OpenPRs = c.countOpenPRs(ctx, owner, info.Name)
		info.CIStatus = c.latestCIStatus(ctx, owner, info.Name, info.DefaultBranch)

		result = append(result, info)
	}
	return result, nil
}

// fetchAllRepos tries ListByOrg first, falls back to ListByUser.
func (c *Client) fetchAllRepos(ctx context.Context, owner string) ([]*github.Repository, error) {
	var repos []*github.Repository

	// Try as GitHub org first.
	orgOpts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		page, resp, err := c.gh.Repositories.ListByOrg(ctx, owner, orgOpts)
		if err != nil {
			break // fall through to user listing
		}
		repos = append(repos, page...)
		if resp.NextPage == 0 {
			return repos, nil
		}
		orgOpts.Page = resp.NextPage
	}

	// Fall back to user repos.
	repos = nil
	userOpts := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		page, resp, err := c.gh.Repositories.ListByUser(ctx, owner, userOpts)
		if err != nil {
			return nil, fmt.Errorf("listing repos for %q: %w", owner, err)
		}
		repos = append(repos, page...)
		if resp.NextPage == 0 {
			return repos, nil
		}
		userOpts.Page = resp.NextPage
	}
}

// countOpenPRs returns the number of open pull requests for a repository.
func (c *Client) countOpenPRs(ctx context.Context, owner, repo string) int {
	var count int
	opts := &github.PullRequestListOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		prs, resp, err := c.gh.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return count
		}
		count += len(prs)
		if resp.NextPage == 0 {
			return count
		}
		opts.Page = resp.NextPage
	}
}

// latestCIStatus returns the conclusion (or status) of the most recent workflow run
// on the given branch. Returns "-" when no runs exist.
func (c *Client) latestCIStatus(ctx context.Context, owner, repo, branch string) string {
	opts := &github.ListWorkflowRunsOptions{
		Branch:      branch,
		ListOptions: github.ListOptions{PerPage: 1},
	}
	runs, _, err := c.gh.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	if err != nil || runs == nil || len(runs.WorkflowRuns) == 0 {
		return "-"
	}
	run := runs.WorkflowRuns[0]
	if c := run.GetConclusion(); c != "" {
		return c
	}
	if s := run.GetStatus(); s != "" {
		return s
	}
	return "-"
}
