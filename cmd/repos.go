package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/repos"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reposCmd)

	reposCmd.Flags().String("registry", "", "GitHub user or org name (default: registry.name config key)")
	reposCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh on interval")
	reposCmd.Flags().Duration("interval", 5*time.Second, "watch refresh interval")

	// Hidden flag used in tests to redirect GitHub API calls to an httptest.Server.
	reposCmd.Flags().String("github-api-url", "", "")
	reposCmd.Flags().MarkHidden("github-api-url") //nolint:errcheck
}

var reposCmd = &cobra.Command{
	Use:          "repos",
	Short:        "List GitHub repositories for the configured registry owner",
	SilenceUsage: true,
	RunE:         runRepos,
}

func runRepos(cmd *cobra.Command, _ []string) error {
	// Resolve owner: --registry flag > registry.name config.
	registry, _ := cmd.Flags().GetString("registry")
	if registry == "" {
		registry = cfgStore.Cfg().Registry.Name
	}
	if registry == "" {
		return fmt.Errorf("registry owner not set: use --registry or set registry.name in config")
	}

	// Token resolution: GITHUB_TOKEN env > registry.token config.
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = cfgStore.Cfg().Registry.Token
	}

	client := repos.New(token)

	// Override GitHub API base URL if provided (test hook).
	if ghURL, _ := cmd.Flags().GetString("github-api-url"); ghURL != "" {
		if err := client.SetBaseURL(ghURL); err != nil {
			return err
		}
	}

	watch, _ := cmd.Flags().GetBool("watch")
	interval, _ := cmd.Flags().GetDuration("interval")
	p := printer()

	fetch := func() error {
		list, err := client.ListRepos(context.Background(), registry)
		if err != nil {
			return fmt.Errorf("listing repos for %q: %w", registry, err)
		}

		headers := []string{"NAME", "OPEN PRS", "DEFAULT BRANCH", "CI STATUS"}
		rows := make([][]string, 0, len(list))
		for _, r := range list {
			rows = append(rows, []string{
				r.Name,
				fmt.Sprintf("%d", r.OpenPRs),
				r.DefaultBranch,
				r.CIStatus,
			})
		}
		return p.PrintTable(headers, rows)
	}

	if watch {
		return reposWatchLoop(interval, fetch)
	}
	return fetch()
}

// reposWatchLoop clears the terminal and calls fn every interval until SIGINT.
func reposWatchLoop(interval time.Duration, fn func() error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	for {
		fmt.Print("\033[H\033[2J")
		if err := fn(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "\nLast updated: %s  (Ctrl+C to stop)\n", time.Now().Format("15:04:05"))

		select {
		case <-quit:
			return nil
		case <-time.After(interval):
		}
	}
}
