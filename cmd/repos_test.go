package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// reposGitHubServer creates an httptest.Server that mocks the GitHub API
// for a given owner's org repos, their pulls, and their CI runs.
func reposGitHubServer(t *testing.T, owner string, repoNames []string, openPRs int, ciStatus string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos"):
			repos := make([]map[string]any, len(repoNames))
			for i, name := range repoNames {
				repos[i] = map[string]any{
					"name":           name,
					"default_branch": "main",
					"full_name":      owner + "/" + name,
				}
			}
			json.NewEncoder(w).Encode(repos)

		case strings.Contains(r.URL.Path, "/pulls"):
			prs := make([]map[string]any, openPRs)
			for i := range prs {
				prs[i] = map[string]any{"number": i + 1}
			}
			json.NewEncoder(w).Encode(prs)

		case strings.Contains(r.URL.Path, "/actions/runs"):
			if ciStatus == "-" {
				json.NewEncoder(w).Encode(map[string]any{
					"total_count":   0,
					"workflow_runs": []any{},
				})
			} else {
				json.NewEncoder(w).Encode(map[string]any{
					"total_count": 1,
					"workflow_runs": []any{map[string]any{
						"id":         1,
						"conclusion": ciStatus,
						"status":     "completed",
					}},
				})
			}

		default:
			http.NotFound(w, r)
		}
	}))
}

// runReposCmdRaw runs the repos command against a mocked GitHub API server.
func runReposCmdRaw(t *testing.T, ghSrv *httptest.Server, args ...string) (string, string, error) {
	t.Helper()
	cfgDir := t.TempDir()
	fullArgs := append(
		[]string{"--config", cfgDir, "--github-api-url", ghSrv.URL + "/"},
		args...,
	)
	return runCmdRaw(t, fullArgs)
}

// TestReposCmd_RendersTable verifies the repos command produces a correctly
// structured table with the right headers and repo data.
func TestReposCmd_RendersTable(t *testing.T) {
	const owner = "myorg"
	repoNames := []string{"repo-alpha", "repo-beta"}

	ghSrv := reposGitHubServer(t, owner, repoNames, 3, "success")
	defer ghSrv.Close()

	stdout, _, err := runReposCmdRaw(t, ghSrv, "repos", "--registry", owner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Headers must be present.
	for _, h := range []string{"NAME", "OPEN PRS", "DEFAULT BRANCH", "CI STATUS"} {
		if !strings.Contains(stdout, h) {
			t.Errorf("expected header %q in output:\n%s", h, stdout)
		}
	}

	// Repo names must appear.
	for _, name := range repoNames {
		if !strings.Contains(stdout, name) {
			t.Errorf("expected repo %q in output:\n%s", name, stdout)
		}
	}

	// PR count must appear.
	if !strings.Contains(stdout, "3") {
		t.Errorf("expected PR count 3 in output:\n%s", stdout)
	}

	// CI status must appear.
	if !strings.Contains(stdout, "success") {
		t.Errorf("expected CI status 'success' in output:\n%s", stdout)
	}
}

// TestReposCmd_RegistryFlag verifies that --registry overrides the config value.
func TestReposCmd_RegistryFlag(t *testing.T) {
	const owner = "override-owner"

	ghSrv := reposGitHubServer(t, owner, []string{"my-repo"}, 0, "-")
	defer ghSrv.Close()

	// Deliberately write a different registry.name in config to confirm override.
	cfgDir := t.TempDir()
	cfgContent := "registry:\n  name: wrong-owner\n"
	os.WriteFile(cfgDir+"/config.yaml", []byte(cfgContent), 0600)

	fullArgs := []string{
		"--config", cfgDir,
		"--github-api-url", ghSrv.URL + "/",
		"repos",
		"--registry", owner,
	}
	stdout, _, err := runCmdRaw(t, fullArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "my-repo") {
		t.Errorf("expected my-repo in output:\n%s", stdout)
	}
}

// TestReposCmd_NoRegistryAndNoConfig verifies an error when no owner is configured.
// It explicitly passes --registry "" to reset any flag state from prior tests.
func TestReposCmd_NoRegistryAndNoConfig(t *testing.T) {
	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ghSrv.Close()

	cfgDir := t.TempDir()
	// Write config with empty registry section to ensure no name is set.
	os.WriteFile(cfgDir+"/config.yaml", []byte("registry:\n  name: \"\"\n"), 0600)

	// Passing --registry "" explicitly resets any flag value left from prior test runs
	// (Cobra does not reset flag values between Execute() calls on a shared command).
	fullArgs := []string{
		"--config", cfgDir,
		"--github-api-url", ghSrv.URL + "/",
		"repos",
		"--registry", "",
	}
	_, _, err := runCmdRaw(t, fullArgs)
	if err == nil {
		t.Fatal("expected error when registry owner is not configured")
	}
	if !strings.Contains(err.Error(), "registry owner not set") {
		t.Errorf("expected 'registry owner not set' error, got: %v", err)
	}
}

// TestReposCmd_ZeroRepos verifies that the command succeeds with an empty repo list.
func TestReposCmd_ZeroRepos(t *testing.T) {
	const owner = "emptyorg"

	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/orgs/"+owner+"/repos") {
			fmt.Fprint(w, "[]")
			return
		}
		http.NotFound(w, r)
	}))
	defer ghSrv.Close()

	stdout, _, err := runReposCmdRaw(t, ghSrv, "repos", "--registry", owner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only headers, no rows.
	if !strings.Contains(stdout, "NAME") {
		t.Errorf("expected header NAME in output:\n%s", stdout)
	}
}
