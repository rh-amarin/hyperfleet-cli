package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/watch"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nodepoolCmd)
}

var nodepoolCmd = &cobra.Command{
	Use:   "nodepool",
	Short: "Manage HyperFleet nodepools",
}

func init() {
	nodepoolCmd.AddCommand(nodepoolCreateCmd)
	nodepoolCmd.AddCommand(nodepoolListCmd)
	nodepoolListCmd.Flags().Bool("table", false, "Display as formatted table")
	nodepoolCmd.AddCommand(nodepoolSearchCmd)
	nodepoolCmd.AddCommand(nodepoolGetCmd)
	nodepoolCmd.AddCommand(nodepoolPatchCmd)
	nodepoolCmd.AddCommand(nodepoolDeleteCmd)
	nodepoolCmd.AddCommand(nodepoolIDCmd)

	nodepoolCmd.AddCommand(nodepoolConditionsCmd)
	nodepoolConditionsCmd.AddCommand(nodepoolConditionsTableCmd)
	nodepoolConditionsCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh on interval")
	nodepoolConditionsCmd.Flags().DurationP("interval", "i", 2*time.Second, "refresh interval for watch mode")
	nodepoolConditionsCmd.Flags().Bool("table", false, "Display as formatted table")

	nodepoolCmd.AddCommand(nodepoolStatusesCmd)
	nodepoolStatusesCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh on interval")
	nodepoolStatusesCmd.Flags().DurationP("interval", "i", 2*time.Second, "refresh interval for watch mode")
	nodepoolStatusesCmd.Flags().Bool("table", false, "Display as formatted table")

	nodepoolCmd.AddCommand(nodepoolTableCmd)
	nodepoolTableCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh on interval")
	nodepoolTableCmd.Flags().DurationP("interval", "i", 2*time.Second, "refresh interval for watch mode")
}

// ── create ────────────────────────────────────────────────────────────────────

var nodepoolCreateCmd = &cobra.Command{
	Use:          "create [name] [count] [instance-type]",
	Short:        "Create one or more nodepools in the current cluster",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "my-nodepool"
		count := 1
		instanceType := "m4"
		if len(args) > 0 {
			name = args[0]
		}
		if len(args) > 1 {
			if n, err := strconv.Atoi(args[1]); err == nil {
				count = n
			}
		}
		if len(args) > 2 {
			instanceType = args[2]
		}

		c := newClient()
		ctx := context.Background()

		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}

		var last *resource.NodePool
		for i := 1; i <= count; i++ {
			npName := fmt.Sprintf("%s-%d", name, i)
			payload := map[string]any{
				"kind": "NodePool",
				"name": npName,
				"labels": map[string]string{
					"counter": fmt.Sprintf("%d", i),
				},
				"spec": map[string]any{
					"counter":  fmt.Sprintf("%d", i),
					"platform": map[string]string{"type": instanceType},
					"replicas": 1,
				},
			}
			np, err := api.Post[resource.NodePool](c, ctx, "clusters/"+clusterID+"/nodepools", payload)
			if err != nil {
				return err
			}
			last = np
		}

		if last != nil {
			if err := config.SetNodePoolID(cfgStore, last.ID); err != nil {
				out.Warn(fmt.Sprintf("could not persist nodepool-id: %v", err))
			}
			return printer().Print(last)
		}
		return nil
	},
}

// ── list ──────────────────────────────────────────────────────────────────────

var nodepoolListCmd = &cobra.Command{
	Use:   "list [cluster_id]",
	Short: "List all nodepools in the current cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterArg := ""
		if len(args) > 0 {
			clusterArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, clusterArg)
		if err != nil {
			return err
		}

		tableMode, _ := cmd.Flags().GetBool("table")
		c := newClient()
		ctx := context.Background()

		list, err := api.Get[resource.ListResponse[resource.NodePool]](c, ctx, "clusters/"+clusterID+"/nodepools")
		if err != nil {
			return err
		}

		if !tableMode {
			return printer().Print(list)
		}

		p := printer()

		// Fetch per-nodepool adapter statuses.
		allStatuses := make([][]resource.AdapterStatus, len(list.Items))
		for i, np := range list.Items {
			sl, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
				c, ctx, "clusters/"+clusterID+"/nodepools/"+np.ID+"/statuses",
			)
			if err != nil {
				if apiErr, ok := api.IsAPIError(err); ok && apiErr.Status == 404 {
					allStatuses[i] = nil
					continue
				}
				return err
			}
			allStatuses[i] = sl.Items
		}

		// Build dynamic condition columns.
		allConds := make([][]out.Condition, len(list.Items))
		for i, np := range list.Items {
			var conds []out.Condition
			for _, cond := range np.Status.Conditions {
				conds = append(conds, out.Condition{Type: cond.Type})
			}
			allConds[i] = conds
		}
		condCols := out.DynamicColumns(allConds)

		// Build dynamic adapter columns.
		adapterCols := out.AdapterNames(allStatuses)

		headers := append(append([]string{"ID", "NAME", "REPLICAS", "TYPE", "GEN"}, condCols...), adapterCols...)

		rows := make([][]string, 0, len(list.Items))
		for i, np := range list.Items {
			deleted := np.DeletedTime != ""

			replicas := fmt.Sprintf("%v", np.Spec["replicas"])
			instanceType := "-"
			if platform, ok := np.Spec["platform"].(map[string]any); ok {
				if t, ok := platform["type"].(string); ok {
					instanceType = t
				}
			}

			condMap := make(map[string]string)
			genMap := make(map[string]int32)
			for _, cond := range np.Status.Conditions {
				condMap[cond.Type] = cond.Status
				genMap[cond.Type] = cond.ObservedGeneration
			}

			row := []string{
				np.ID,
				np.Name,
				replicas,
				instanceType,
				p.GenCell(np.Generation, np.DeletedTime != ""),
			}
			for _, col := range condCols {
				row = append(row, p.DotWithGen(condMap[col], genMap[col]))
			}
			for _, adapter := range adapterCols {
				row = append(row, p.AdapterDot(allStatuses[i], adapter, deleted))
			}
			rows = append(rows, row)
		}

		return p.PrintTable(headers, rows)
	},
}

// ── search ────────────────────────────────────────────────────────────────────

var nodepoolSearchCmd = &cobra.Command{
	Use:          "search [name]",
	Short:        "Search nodepools by name and set as current",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			npID := cfgStore.State().NodePoolID
			if npID == "" {
				out.Errorf("No nodepool-id set in state. Run 'hf nodepool create' or 'hf nodepool search <name>' first.")
				return fmt.Errorf("no nodepool-id set in state")
			}
			clusterID, err := config.ClusterID(cfgStore, "")
			if err != nil {
				return err
			}
			c := newClient()
			np, err := api.Get[resource.NodePool](c, context.Background(), "clusters/"+clusterID+"/nodepools/"+npID)
			if err != nil {
				if apiErr, ok := api.IsAPIError(err); ok {
					return printer().Print(apiErr)
				}
				return err
			}
			return printer().Print(np)
		}

		name := args[0]
		c := newClient()
		ctx := context.Background()

		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}

		matches, err := api.FindNodePoolByName(c, ctx, clusterID, name)
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			out.Warn(fmt.Sprintf("No nodepools found matching '%s'", name))
			return printer().Print([]resource.NodePool{})
		}

		if len(matches) > 1 {
			out.Warn(fmt.Sprintf("Multiple nodepools found matching '%s', using first result", name))
		}

		if err := config.SetNodePoolID(cfgStore, matches[0].ID); err != nil {
			out.Warn(fmt.Sprintf("could not persist nodepool-id: %v", err))
		}

		return printer().Print(matches)
	},
}

// ── get ───────────────────────────────────────────────────────────────────────

var nodepoolGetCmd = &cobra.Command{
	Use:   "get [nodepool_id]",
	Short: "Get a nodepool by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		npArg := ""
		if len(args) > 0 {
			npArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, npArg)
		if err != nil {
			return err
		}

		c := newClient()
		np, err := api.Get[resource.NodePool](c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID)
		if err != nil {
			if apiErr, ok := api.IsAPIError(err); ok {
				return printer().Print(apiErr)
			}
			return err
		}
		return printer().Print(np)
	},
}

// ── patch ─────────────────────────────────────────────────────────────────────

var nodepoolPatchCmd = &cobra.Command{
	Use:   "patch <spec|labels> [nodepool_id]",
	Short: "Increment the counter field in nodepool spec or labels",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != "spec" && args[0] != "labels") {
			fmt.Fprintln(os.Stderr, "Usage: hf nodepool patch spec|labels [nodepool_id]")
			return fmt.Errorf("target must be 'spec' or 'labels'")
		}
		target := args[0]

		npArg := ""
		if len(args) > 1 {
			npArg = args[1]
		}
		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, npArg)
		if err != nil {
			return err
		}

		c := newClient()
		ctx := context.Background()

		np, err := api.Get[resource.NodePool](c, ctx, "clusters/"+clusterID+"/nodepools/"+nodepoolID)
		if err != nil {
			return err
		}

		var section map[string]any
		if target == "spec" {
			section = np.Spec
		} else {
			section = make(map[string]any, len(np.Labels))
			for k, v := range np.Labels {
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
			updated := make(map[string]any, len(np.Spec))
			for k, v := range np.Spec {
				updated[k] = v
			}
			updated["counter"] = nextStr
			payload = map[string]any{"spec": updated}
		} else {
			updatedLabels := make(map[string]string, len(np.Labels))
			for k, v := range np.Labels {
				updatedLabels[k] = v
			}
			updatedLabels["counter"] = nextStr
			payload = map[string]any{"labels": updatedLabels}
		}

		patched, err := api.Patch[resource.NodePool](c, ctx, "clusters/"+clusterID+"/nodepools/"+nodepoolID, payload)
		if err != nil {
			return err
		}
		return printer().Print(patched)
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var nodepoolDeleteCmd = &cobra.Command{
	Use:   "delete [nodepool_id]",
	Short: "Delete a nodepool",
	RunE: func(cmd *cobra.Command, args []string) error {
		npArg := ""
		if len(args) > 0 {
			npArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, npArg)
		if err != nil {
			return err
		}

		c := newClient()
		deleted, err := api.Delete[resource.NodePool](c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID)
		if err != nil {
			return err
		}
		return printer().Print(deleted)
	},
}

// ── id ────────────────────────────────────────────────────────────────────────

var nodepoolIDCmd = &cobra.Command{
	Use:   "id",
	Short: "Print the configured nodepool-id",
	RunE: func(cmd *cobra.Command, args []string) error {
		id := cfgStore.State().NodePoolID
		if id == "" {
			out.Warn("No nodepool-id set in state")
			return nil
		}
		fmt.Println(id)
		return nil
	},
}

// ── conditions ────────────────────────────────────────────────────────────────

var nodepoolConditionsCmd = &cobra.Command{
	Use:   "conditions [nodepool_id]",
	Short: "Show nodepool conditions",
	RunE: func(cmd *cobra.Command, args []string) error {
		npArg := ""
		if len(args) > 0 {
			npArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, npArg)
		if err != nil {
			return err
		}

		tableMode, _ := cmd.Flags().GetBool("table")
		watchMode, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")
		p := printer()
		c := newClient()

		if tableMode {
			np, err := api.Get[resource.NodePool](c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID)
			if err != nil {
				return err
			}
			return renderConditionsTable(p, np.Status.Conditions)
		}

		fetch := func() error {
			np, err := api.Get[resource.NodePool](c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID)
			if err != nil {
				return err
			}
			result := map[string]any{
				"generation": np.Generation,
				"status":     map[string]any{"conditions": np.Status.Conditions},
			}
			return p.Print(result)
		}

		if watchMode {
			return watch.Watch(interval, fetch)
		}
		return fetch()
	},
}

// ── conditions table ──────────────────────────────────────────────────────────

var nodepoolConditionsTableCmd = &cobra.Command{
	Use:   "table [nodepool_id]",
	Short: "Show nodepool conditions as a formatted table",
	RunE: func(cmd *cobra.Command, args []string) error {
		npArg := ""
		if len(args) > 0 {
			npArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, npArg)
		if err != nil {
			return err
		}

		c := newClient()
		p := printer()

		np, err := api.Get[resource.NodePool](c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID)
		if err != nil {
			return err
		}
		return renderConditionsTable(p, np.Status.Conditions)
	},
}

// ── statuses ──────────────────────────────────────────────────────────────────

var nodepoolStatusesCmd = &cobra.Command{
	Use:   "statuses [nodepool_id]",
	Short: "Show nodepool adapter statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		npArg := ""
		if len(args) > 0 {
			npArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, npArg)
		if err != nil {
			return err
		}

		tableMode, _ := cmd.Flags().GetBool("table")
		watchMode, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")
		p := printer()
		c := newClient()

		emptyStatuses := resource.ListResponse[resource.AdapterStatus]{
			Items: []resource.AdapterStatus{},
			Kind:  "AdapterStatusList",
			Page:  1,
			Size:  0,
			Total: 0,
		}

		if tableMode {
			list, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
				c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses",
			)
			if err != nil {
				if apiErr, ok := api.IsAPIError(err); ok && apiErr.Status == 404 {
					return renderStatusesTable(p, nil)
				}
				return err
			}
			return renderStatusesTable(p, list.Items)
		}

		fetch := func() error {
			list, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
				c, context.Background(), "clusters/"+clusterID+"/nodepools/"+nodepoolID+"/statuses",
			)
			if err != nil {
				if apiErr, ok := api.IsAPIError(err); ok && apiErr.Status == 404 {
					return p.Print(emptyStatuses)
				}
				return err
			}
			return p.Print(list)
		}

		if watchMode {
			return watch.Watch(interval, fetch)
		}
		return fetch()
	},
}

// ── table ─────────────────────────────────────────────────────────────────────

var nodepoolTableCmd = &cobra.Command{
	Use:   "table [cluster_id]",
	Short: "List nodepools as a formatted table with dynamic condition columns",
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterArg := ""
		if len(args) > 0 {
			clusterArg = args[0]
		}
		clusterID, err := config.ClusterID(cfgStore, clusterArg)
		if err != nil {
			return err
		}

		watchMode, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")
		c := newClient()
		p := printer()

		fetch := func() error {
			list, err := api.Get[resource.ListResponse[resource.NodePool]](c, context.Background(), "clusters/"+clusterID+"/nodepools")
			if err != nil {
				return err
			}

			allConditions := make([][]out.Condition, 0, len(list.Items))
			for _, np := range list.Items {
				var conds []out.Condition
				for _, c := range np.Status.Conditions {
					conds = append(conds, out.Condition{Type: c.Type})
				}
				allConditions = append(allConditions, conds)
			}
			dynCols := out.DynamicColumns(allConditions)

			headers := append([]string{"ID", "NAME", "REPLICAS", "TYPE", "GEN"}, dynCols...)

			rows := make([][]string, 0, len(list.Items))
			for _, np := range list.Items {
				replicas := fmt.Sprintf("%v", np.Spec["replicas"])
				instanceType := ""
				if platform, ok := np.Spec["platform"].(map[string]any); ok {
					instanceType, _ = platform["type"].(string)
				}

				condMap := make(map[string]string)
				genMap := make(map[string]int32)
				for _, cond := range np.Status.Conditions {
					condMap[cond.Type] = cond.Status
					genMap[cond.Type] = cond.ObservedGeneration
				}

				row := []string{
					np.ID,
					np.Name,
					replicas,
					instanceType,
					fmt.Sprintf("%d", np.Generation),
				}
				for _, col := range dynCols {
					row = append(row, p.DotWithGen(condMap[col], genMap[col]))
				}
				rows = append(rows, row)
			}

			return p.PrintTable(headers, rows)
		}

		if watchMode {
			return watch.Watch(interval, fetch)
		}
		return fetch()
	},
}
