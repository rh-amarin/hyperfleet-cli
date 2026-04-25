package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clusterCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage HyperFleet clusters",
}

func init() {
	// create
	clusterCmd.AddCommand(clusterCreateCmd)

	// get
	clusterCmd.AddCommand(clusterGetCmd)

	// list
	clusterCmd.AddCommand(clusterListCmd)

	// search
	clusterCmd.AddCommand(clusterSearchCmd)

	// patch
	clusterCmd.AddCommand(clusterPatchCmd)

	// delete
	clusterCmd.AddCommand(clusterDeleteCmd)

	// id
	clusterCmd.AddCommand(clusterIDCmd)

	// conditions + conditions table
	clusterCmd.AddCommand(clusterConditionsCmd)
	clusterConditionsCmd.AddCommand(clusterConditionsTableCmd)
	clusterConditionsCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh every 2s")

	// statuses
	clusterCmd.AddCommand(clusterStatusesCmd)
	clusterStatusesCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh every 2s")
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newClient() *api.Client {
	cfg := cfgStore.Cfg()
	return api.NewClient(cfg.Hyperfleet.APIURL, cfg.Hyperfleet.APIVersion, cfg.Hyperfleet.Token, verbose)
}

func printer() *out.Printer {
	return out.NewPrinter(output, noColor)
}

// watchLoop clears the terminal and calls fn every 2s until SIGINT.
func watchLoop(fn func() error) error {
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
		case <-time.After(2 * time.Second):
		}
	}
}

// ── create ────────────────────────────────────────────────────────────────────

var clusterCreateCmd = &cobra.Command{
	Use:          "create [name] [region] [version]",
	Short:        "Create a new cluster",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "my-cluster"
		region := "us-east-1"
		version := "4.15.0"
		if len(args) > 0 {
			name = args[0]
		}
		if len(args) > 1 {
			region = args[1]
		}
		if len(args) > 2 {
			version = args[2]
		}

		c := newClient()
		ctx := context.Background()

		// Duplicate guard
		existing, err := api.FindClusterByName(c, ctx, name)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			out.Warn(fmt.Sprintf("Cluster '%s' already exists, skipping creation", name))
			return nil
		}

		payload := map[string]any{
			"kind": "Cluster",
			"name": name,
			"labels": map[string]string{
				"counter":     "1",
				"environment": "development",
				"shard":       "1",
				"team":        "core",
			},
			"spec": map[string]string{
				"counter": "1",
				"region":  region,
				"version": version,
			},
		}

		cluster, err := api.Post[resource.Cluster](c, ctx, "clusters", payload)
		if err != nil {
			return err
		}

		// Persist cluster-id via search (get the canonical ID from the API)
		matches, err := api.FindClusterByName(c, ctx, name)
		if err == nil && len(matches) > 0 {
			if setErr := config.SetClusterID(cfgStore, matches[0].ID, matches[0].Name); setErr != nil {
				out.Warn(fmt.Sprintf("could not persist cluster-id: %v", setErr))
			}
		}

		return printer().Print(cluster)
	},
}

// ── get ───────────────────────────────────────────────────────────────────────

var clusterGetCmd = &cobra.Command{
	Use:   "get [cluster_id]",
	Short: "Get a cluster by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := ""
		if len(args) > 0 {
			arg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, arg)
		if err != nil {
			return err
		}

		c := newClient()
		cluster, err := api.Get[resource.Cluster](c, context.Background(), "clusters/"+clusterID)
		if err != nil {
			if apiErr, ok := api.IsAPIError(err); ok {
				return printer().Print(apiErr)
			}
			return err
		}
		return printer().Print(cluster)
	},
}

// ── list ──────────────────────────────────────────────────────────────────────

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		list, err := api.Get[resource.ListResponse[resource.Cluster]](c, context.Background(), "clusters")
		if err != nil {
			return err
		}
		return printer().Print(list)
	},
}

// ── search ────────────────────────────────────────────────────────────────────

var clusterSearchCmd = &cobra.Command{
	Use:   "search <name>",
	Short: "Search clusters by name and set as current",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		c := newClient()
		ctx := context.Background()

		matches, err := api.FindClusterByName(c, ctx, name)
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			out.Warn(fmt.Sprintf("No clusters found matching '%s'", name))
			return printer().Print([]resource.Cluster{})
		}

		if len(matches) > 1 {
			out.Warn(fmt.Sprintf("Multiple clusters found matching '%s', using first match", name))
		}

		if err := config.SetClusterID(cfgStore, matches[0].ID, matches[0].Name); err != nil {
			out.Warn(fmt.Sprintf("could not persist cluster-id: %v", err))
		}

		return printer().Print(matches)
	},
}

// ── patch ─────────────────────────────────────────────────────────────────────

var clusterPatchCmd = &cobra.Command{
	Use:   "patch <spec|labels> [cluster_id]",
	Short: "Increment the counter field in cluster spec or labels",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != "spec" && args[0] != "labels") {
			fmt.Fprintln(os.Stderr, "Usage: hf cluster patch spec|labels [cluster_id]")
			return fmt.Errorf("target must be 'spec' or 'labels'")
		}
		target := args[0]

		clusterArg := ""
		if len(args) > 1 {
			clusterArg = args[1]
		}
		clusterID, err := config.ClusterID(cfgStore, clusterArg)
		if err != nil {
			return err
		}

		c := newClient()
		ctx := context.Background()

		cluster, err := api.Get[resource.Cluster](c, ctx, "clusters/"+clusterID)
		if err != nil {
			return err
		}

		// Read current counter from the target section
		var section map[string]any
		if target == "spec" {
			section = cluster.Spec
		} else {
			// Labels is map[string]string — convert for uniform handling
			section = make(map[string]any, len(cluster.Labels))
			for k, v := range cluster.Labels {
				section[k] = v
			}
		}

		currentStr, _ := section["counter"].(string)
		current := 0
		if currentStr != "" {
			fmt.Sscanf(currentStr, "%d", &current)
		}
		next := current + 1
		nextStr := fmt.Sprintf("%d", next)

		out.Info(fmt.Sprintf("Incrementing %s.counter: %s -> %s", target, currentStr, nextStr))

		var payload map[string]any
		if target == "spec" {
			updated := make(map[string]any, len(cluster.Spec))
			for k, v := range cluster.Spec {
				updated[k] = v
			}
			updated["counter"] = nextStr
			payload = map[string]any{"spec": updated}
		} else {
			updatedLabels := make(map[string]string, len(cluster.Labels))
			for k, v := range cluster.Labels {
				updatedLabels[k] = v
			}
			updatedLabels["counter"] = nextStr
			payload = map[string]any{"labels": updatedLabels}
		}

		patched, err := api.Patch[resource.Cluster](c, ctx, "clusters/"+clusterID, payload)
		if err != nil {
			return err
		}
		return printer().Print(patched)
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var clusterDeleteCmd = &cobra.Command{
	Use:   "delete [cluster_id]",
	Short: "Delete a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := ""
		if len(args) > 0 {
			arg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, arg)
		if err != nil {
			return err
		}

		c := newClient()
		deleted, err := api.Delete[resource.Cluster](c, context.Background(), "clusters/"+clusterID)
		if err != nil {
			return err
		}
		return printer().Print(deleted)
	},
}

// ── id ────────────────────────────────────────────────────────────────────────

var clusterIDCmd = &cobra.Command{
	Use:   "id",
	Short: "Print the configured cluster-id",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

// ── conditions ────────────────────────────────────────────────────────────────

var clusterConditionsCmd = &cobra.Command{
	Use:   "conditions [cluster_id]",
	Short: "Show cluster conditions",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := ""
		if len(args) > 0 {
			arg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, arg)
		if err != nil {
			return err
		}

		watch, _ := cmd.Flags().GetBool("watch")
		p := printer()
		c := newClient()

		fetch := func() error {
			cluster, err := api.Get[resource.Cluster](c, context.Background(), "clusters/"+clusterID)
			if err != nil {
				return err
			}
			result := map[string]any{
				"generation": cluster.Generation,
				"status":     map[string]any{"conditions": cluster.Status.Conditions},
			}
			return p.Print(result)
		}

		if watch {
			return watchLoop(fetch)
		}
		return fetch()
	},
}

// ── conditions table ──────────────────────────────────────────────────────────

var clusterConditionsTableCmd = &cobra.Command{
	Use:   "table [cluster_id]",
	Short: "Show cluster conditions as a formatted table",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := ""
		if len(args) > 0 {
			arg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, arg)
		if err != nil {
			return err
		}

		c := newClient()
		p := printer()

		cluster, err := api.Get[resource.Cluster](c, context.Background(), "clusters/"+clusterID)
		if err != nil {
			return err
		}

		headers := []string{"TYPE", "STATUS", "LAST TRANSITION", "REASON", "MESSAGE"}
		rows := make([][]string, 0, len(cluster.Status.Conditions))
		for _, cond := range cluster.Status.Conditions {
			rows = append(rows, []string{
				cond.Type,
				p.Dot(cond.Status),
				cond.LastTransitionTime,
				cond.Reason,
				cond.Message,
			})
		}
		return p.PrintTable(headers, rows)
	},
}

// ── statuses ──────────────────────────────────────────────────────────────────

var clusterStatusesCmd = &cobra.Command{
	Use:   "statuses [cluster_id]",
	Short: "Show cluster adapter statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := ""
		if len(args) > 0 {
			arg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, arg)
		if err != nil {
			return err
		}

		watch, _ := cmd.Flags().GetBool("watch")
		p := printer()
		c := newClient()

		emptyStatuses := resource.ListResponse[resource.AdapterStatus]{
			Items: []resource.AdapterStatus{},
			Kind:  "AdapterStatusList",
			Page:  1,
			Size:  0,
			Total: 0,
		}

		fetch := func() error {
			list, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
				c, context.Background(), "clusters/"+clusterID+"/adapter-statuses",
			)
			if err != nil {
				if apiErr, ok := api.IsAPIError(err); ok && apiErr.Status == 404 {
					return p.Print(emptyStatuses)
				}
				return err
			}
			return p.Print(list)
		}

		if watch {
			return watchLoop(fetch)
		}
		return fetch()
	},
}
