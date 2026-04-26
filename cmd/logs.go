package cmd

import (
	"context"
	"os"
	"os/exec"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
	"github.com/spf13/cobra"
)

var (
	logsKubeconfig string
	logsNamespace  string
	logsClusterID  string
)

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.PersistentFlags().StringVar(&logsKubeconfig, "kubeconfig", "", "path to kubeconfig (default: KUBECONFIG env, then ~/.kube/config)")
	logsCmd.PersistentFlags().StringVarP(&logsNamespace, "namespace", "n", "amarin-ns1", "Kubernetes namespace")
	logsCmd.AddCommand(logsAdapterCmd)
	logsAdapterCmd.Flags().StringVar(&logsClusterID, "cluster-id", "", "filter adapter logs by cluster ID (default: active cluster from config)")
}

var logsCmd = &cobra.Command{
	Use:          "logs [pattern]",
	Short:        "Stream logs from pods matching pattern (uses stern if available)",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}

		// Use stern if available — it handles multi-pod tailing natively.
		if sternPath, err := exec.LookPath("stern"); err == nil {
			sternArgs := []string{pattern, "-n", logsNamespace}
			if logsKubeconfig != "" {
				sternArgs = append(sternArgs, "--kubeconfig", logsKubeconfig)
			}
			c := exec.CommandContext(context.Background(), sternPath, sternArgs...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin
			return c.Run()
		}

		// Fall back to concurrent goroutine fan-out.
		cs, err := kube.NewClientset(logsKubeconfig)
		if err != nil {
			return err
		}
		return kube.StreamLogs(context.Background(), cs, logsNamespace, pattern, os.Stdout)
	},
}

var logsAdapterCmd = &cobra.Command{
	Use:          "adapter [pattern]",
	Short:        "Stream adapter pod logs filtered by cluster-id, showing time and message fields",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}

		// Prefix pattern with "adapter" to narrow pod search.
		searchPattern := "adapter"
		if pattern != "" {
			searchPattern = "adapter-" + pattern
		}

		// Resolve cluster-id: flag → active config state.
		clusterID := logsClusterID
		if clusterID == "" && cfgStore != nil {
			clusterID = cfgStore.State().ClusterID
		}

		cs, err := kube.NewClientset(logsKubeconfig)
		if err != nil {
			return err
		}
		return kube.StreamLogsFiltered(context.Background(), cs, logsNamespace, searchPattern, clusterID, os.Stdout)
	},
}
