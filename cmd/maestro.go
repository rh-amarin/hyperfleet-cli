package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/rh-amarin/hyperfleet-cli/internal/maestro"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(maestroCmd)
}

var maestroCmd = &cobra.Command{
	Use:   "maestro",
	Short: "Manage Maestro resources",
}

func init() {
	maestroCmd.AddCommand(maestroListCmd)
	maestroCmd.AddCommand(maestroGetCmd)
	maestroCmd.AddCommand(maestroDeleteCmd)
	maestroCmd.AddCommand(maestroBundlesCmd)
	maestroCmd.AddCommand(maestroConsumersCmd)
	maestroCmd.AddCommand(maestroTUICmd)
	maestroDeleteCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
}

func newMaestroClient() *maestro.Client {
	cfg := cfgStore.Cfg()
	return maestro.New(cfg.Maestro.HTTPEndpoint, cfg.Maestro.Consumer, cfg.Hyperfleet.Token)
}

// ── list ──────────────────────────────────────────────────────────────────────

var maestroListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List Maestro resource bundles for the configured consumer",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newMaestroClient()
		resources, err := c.List(context.Background())
		if err != nil {
			return err
		}

		headers := []string{"NAME", "CONSUMER", "VERSION", "MANIFESTS"}
		rows := make([][]string, 0, len(resources))
		for _, r := range resources {
			rows = append(rows, []string{
				r.Metadata["name"],
				r.ConsumerName,
				fmt.Sprintf("%d", r.Version),
				fmt.Sprintf("%d", r.ManifestCount),
			})
		}
		return printer().PrintTable(headers, rows)
	},
}

// ── get ───────────────────────────────────────────────────────────────────────

var maestroGetCmd = &cobra.Command{
	Use:          "get <id>",
	Short:        "Get a Maestro resource bundle by ID",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newMaestroClient()
		resource, err := c.Get(context.Background(), args[0])
		if err != nil {
			return err
		}
		return printer().Print(resource)
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var maestroDeleteCmd = &cobra.Command{
	Use:          "delete <id>",
	Short:        "Delete a Maestro resource bundle by ID",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		yes, _ := cmd.Flags().GetBool("yes")

		if !yes {
			fmt.Fprintf(os.Stderr, "Delete Maestro resource bundle '%s'? [y/N]: ", id)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				out.Info("Deletion cancelled")
				return nil
			}
		}

		c := newMaestroClient()
		if err := c.Delete(context.Background(), id); err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Deleted Maestro resource bundle '%s'", id))
		return nil
	},
}

// ── bundles ───────────────────────────────────────────────────────────────────

var maestroBundlesCmd = &cobra.Command{
	Use:          "bundles",
	Short:        "List all Maestro resource bundles (unfiltered)",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newMaestroClient()
		bundles, err := c.ListBundles(context.Background())
		if err != nil {
			return err
		}
		return printer().Print(bundles)
	},
}

// ── consumers ─────────────────────────────────────────────────────────────────

var maestroConsumersCmd = &cobra.Command{
	Use:          "consumers",
	Short:        "List Maestro consumers",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newMaestroClient()
		consumers, err := c.ListConsumers(context.Background())
		if err != nil {
			return err
		}

		headers := []string{"ID", "NAME"}
		rows := make([][]string, 0, len(consumers))
		for _, consumer := range consumers {
			rows = append(rows, []string{consumer.ID, consumer.Name})
		}
		return printer().PrintTable(headers, rows)
	},
}

// ── tui ───────────────────────────────────────────────────────────────────────

var maestroTUICmd = &cobra.Command{
	Use:          "tui",
	Short:        "Launch the Maestro terminal UI (requires maestro-cli in PATH)",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		maestroCLI, err := exec.LookPath("maestro-cli")
		if err != nil {
			return fmt.Errorf(
				"maestro-cli not found in PATH\n"+
					"Install it from the Maestro release artifacts or run:\n"+
					"  oc apply -f <maestro-deployment-url>",
			)
		}
		cfg := cfgStore.Cfg().Maestro
		env := append(os.Environ(),
			"MAESTRO_HTTP_ENDPOINT="+cfg.HTTPEndpoint,
			"MAESTRO_GRPC_ENDPOINT="+cfg.GRPCEndpoint,
		)
		return syscall.Exec(maestroCLI, []string{"maestro-cli", "tui"}, env)
	},
}
