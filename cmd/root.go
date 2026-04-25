package cmd

import (
	"fmt"
	"os"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgDir   string
	output   string
	noColor  bool
	verbose  bool
	apiURL   string
	apiToken string

	// store is initialised in PersistentPreRunE and shared across commands.
	store *config.Store
)

var rootCmd = &cobra.Command{
	Use:   "hf",
	Short: "HyperFleet CLI — manage clusters, nodepools, and fleet resources",
	Long: `hf is the HyperFleet CLI tool for managing clusters, nodepools,
adapter statuses, databases, Maestro resources, and Kubernetes operations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		store, err = config.NewStore(cfgDir)
		if err != nil {
			return fmt.Errorf("failed to initialise config store: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "config directory (default ~/.config/hf)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output format: json, table, yaml")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "override API URL")
	rootCmd.PersistentFlags().StringVar(&apiToken, "api-token", "", "override API token")
}
