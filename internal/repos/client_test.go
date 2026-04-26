package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// repoJSON returns minimal GitHub repo JSON.
func repoJSON(name, defaultBranch string) map[string]any {
	return map[string]any{
		"name":           name,
		"default_branch": defaultBranch,
		"full_name":      "owner/" + name,
	}
}

// workflowRunJSON returns a minimal workflow run JSON.
func workflowRunJSON(conclusion, status string) map[string]any {
	return map[string]any{
		"id":         1,
		"conclusion": conclusion,
		"status":     status,
	}
}

// newTestClient creates a repos.Client pointed at a local httptest.Server.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c := New("")
	if err := c.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatalf("SetBaseURL: %v", err)
	}
	return c
}

// TestListRepos_ReturnsRepoInfo tests that ListRepos returns the correct RepoInfo
// for a mocked GitHub API response with one repo, two open PRs, and one CI run.
func TestListRepos_ReturnsRepoInfo(t *testing.T) {
	const owner = "testorg"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos"):
			json.NewEncoder(w).Encode([]any{repoJSON("myrepo", "main")})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/"+owner+"/myrepo/pulls"):
			json.NewEncoder(w).Encode([]any{
				map[string]any{"number": 1, "title": "PR one"},
				map[string]any{"number": 2, "title": "PR two"},
			})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/"+owner+"/myrepo/actions/runs"):
			json.NewEncoder(w).Encode(map[string]any{
				"total_count":   1,
				"workflow_runs": []any{workflowRunJSON("success", "completed")},
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	repos, err := c.ListRepos(context.Background(), owner)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(repos))
	}
	got := repos[0]
	if got.Name != "myrepo" {
		t.Errorf("Name = %q, want myrepo", got.Name)
	}
	if got.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want main", got.DefaultBranch)
	}
	if got.OpenPRs != 2 {
		t.Errorf("OpenPRs = %d, want 2", got.OpenPRs)
	}
	if got.CIStatus != "success" {
		t.Errorf("CIStatus = %q, want success", got.CIStatus)
	}
}

// TestListRepos_NoPRsNoCIRuns tests repos with empty PR list and no workflow runs.
func TestListRepos_NoPRsNoCIRuns(t *testing.T) {
	const owner = "testorg"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos"):
			json.NewEncoder(w).Encode([]any{repoJSON("emptyrepo", "main")})
		case strings.Contains(r.URL.Path, "/pulls"):
			json.NewEncoder(w).Encode([]any{})
		case strings.Contains(r.URL.Path, "/actions/runs"):
			json.NewEncoder(w).Encode(map[string]any{
				"total_count":   0,
				"workflow_runs": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	repos, err := c.ListRepos(context.Background(), owner)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(repos))
	}
	got := repos[0]
	if got.OpenPRs != 0 {
		t.Errorf("OpenPRs = %d, want 0", got.OpenPRs)
	}
	if got.CIStatus != "-" {
		t.Errorf("CIStatus = %q, want -", got.CIStatus)
	}
}

// TestListRepos_UserFallback tests that when the org endpoint returns 404,
// the client falls back to the user endpoint.
func TestListRepos_UserFallback(t *testing.T) {
	const owner = "testuser"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"Not Found"}`)
		case strings.Contains(r.URL.Path, "/users/"+owner+"/repos"):
			json.NewEncoder(w).Encode([]any{repoJSON("userrepo", "develop")})
		case strings.Contains(r.URL.Path, "/pulls"):
			json.NewEncoder(w).Encode([]any{})
		case strings.Contains(r.URL.Path, "/actions/runs"):
			json.NewEncoder(w).Encode(map[string]any{
				"total_count":   0,
				"workflow_runs": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	repos, err := c.ListRepos(context.Background(), owner)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(repos))
	}
	if repos[0].Name != "userrepo" {
		t.Errorf("Name = %q, want userrepo", repos[0].Name)
	}
	if repos[0].DefaultBranch != "develop" {
		t.Errorf("DefaultBranch = %q, want develop", repos[0].DefaultBranch)
	}
}

// TestListRepos_CIStatusInProgress tests that an in-progress run without a conclusion
// returns the status field instead.
func TestListRepos_CIStatusInProgress(t *testing.T) {
	const owner = "testorg"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos"):
			json.NewEncoder(w).Encode([]any{repoJSON("runningrepo", "main")})
		case strings.Contains(r.URL.Path, "/pulls"):
			json.NewEncoder(w).Encode([]any{})
		case strings.Contains(r.URL.Path, "/actions/runs"):
			// conclusion is empty (run not yet complete)
			json.NewEncoder(w).Encode(map[string]any{
				"total_count":   1,
				"workflow_runs": []any{workflowRunJSON("", "in_progress")},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	repos, err := c.ListRepos(context.Background(), owner)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(repos))
	}
	if repos[0].CIStatus != "in_progress" {
		t.Errorf("CIStatus = %q, want in_progress", repos[0].CIStatus)
	}
}

// TestListRepos_MultipleRepos tests pagination and multiple repos in one response.
func TestListRepos_MultipleRepos(t *testing.T) {
	const owner = "testorg"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos"):
			json.NewEncoder(w).Encode([]any{
				repoJSON("alpha", "main"),
				repoJSON("beta", "main"),
				repoJSON("gamma", "trunk"),
			})
		case strings.Contains(r.URL.Path, "/pulls"):
			json.NewEncoder(w).Encode([]any{})
		case strings.Contains(r.URL.Path, "/actions/runs"):
			json.NewEncoder(w).Encode(map[string]any{
				"total_count":   1,
				"workflow_runs": []any{workflowRunJSON("failure", "completed")},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	repos, err := c.ListRepos(context.Background(), owner)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("got %d repos, want 3", len(repos))
	}
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name
	}
	for _, want := range []string{"alpha", "beta", "gamma"} {
		found := false
		for _, n := range names {
			if n == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("repo %q not found in results %v", want, names)
		}
	}
	// gamma has default branch "trunk"
	for _, r := range repos {
		if r.Name == "gamma" && r.DefaultBranch != "trunk" {
			t.Errorf("gamma DefaultBranch = %q, want trunk", r.DefaultBranch)
		}
	}
}
