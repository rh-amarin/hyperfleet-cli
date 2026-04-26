package cmd

import (
	"context"
	"os"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
	"github.com/spf13/cobra"
)

var (
	logsKubeconfig    string
	logsNamespace     string
)

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.PersistentFlags().StringVar(&logsKubeconfig, "kubeconfig", "", "path to kubeconfig (default: KUBECONFIG env, then ~/.kube/config)")
	logsCmd.PersistentFlags().StringVarP(&logsNamespace, "namespace", "n", "amarin-ns1", "Kubernetes namespace")
	logsCmd.AddCommand(logsAdapterCmd)
}

var logsCmd = &cobra.Command{
	Use:          "logs <pattern>",
	Short:        "Stream logs from pods matching pattern",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}

		cs, err := kube.NewClientset(logsKubeconfig)
		if err != nil {
			return err
		}

		return kube.StreamLogs(context.Background(), cs, logsNamespace, pattern, os.Stdout)
	},
}

var logsAdapterCmd = &cobra.Command{
	Use:          "adapter <pattern>",
	Short:        "Stream logs from adapter pods matching pattern, filtered by cluster-id",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}

		// Prefix pattern with "adapter" when non-empty to narrow search.
		searchPattern := "adapter"
		if pattern != "" {
			searchPattern = "adapter-" + pattern
		}

		cs, err := kube.NewClientset(logsKubeconfig)
		if err != nil {
			return err
		}

		return kube.StreamLogs(context.Background(), cs, logsNamespace, searchPattern, os.Stdout)
	},
}
