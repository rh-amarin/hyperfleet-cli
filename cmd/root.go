package cmd

import (
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
)

// cfgStore is initialised in PersistentPreRunE and used by all subcommands.
var cfgStore *config.Store

var rootCmd = &cobra.Command{
	Use:   "hf",
	Short: "HyperFleet CLI — manage clusters, nodepools, and fleet resources",
	Long: `hf is the HyperFleet CLI tool for managing clusters, nodepools,
adapter statuses, databases, Maestro resources, and Kubernetes operations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// HF_CONFIG_DIR env var as fallback for --config flag
		dir := cfgDir
		if dir == "" {
			dir = os.Getenv("HF_CONFIG_DIR")
		}
		s, err := config.NewStore(dir)
		if err != nil {
			return err
		}
		// Apply flag overrides into the resolved config
		if apiURL != "" {
			if err := s.OverrideCfg("hyperfleet.api-url", apiURL); err != nil {
				return err
			}
		}
		if apiToken != "" {
			if err := s.OverrideCfg("hyperfleet.token", apiToken); err != nil {
				return err
			}
		}
		cfgStore = s
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
