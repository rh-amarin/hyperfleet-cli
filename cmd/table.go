package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/rh-amarin/hyperfleet-cli/internal/watch"
	"github.com/spf13/cobra"
)

func init() {
	clusterCmd.AddCommand(clusterTableCmd)
	clusterTableCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh on interval")
	clusterTableCmd.Flags().DurationP("interval", "i", 2*time.Second, "refresh interval for watch mode")

	rootCmd.AddCommand(tableCmd)
	tableCmd.Flags().BoolP("watch", "w", false, "watch mode: refresh on interval")
	tableCmd.Flags().DurationP("interval", "i", 2*time.Second, "refresh interval for watch mode")
}

// ── hf cluster table ──────────────────────────────────────────────────────────

var clusterTableCmd = &cobra.Command{
	Use:   "table",
	Short: "List clusters as a formatted table with dynamic condition columns",
	RunE: func(cmd *cobra.Command, args []string) error {
		watchMode, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")
		c := newClient()
		p := printer()

		fetch := func() error {
			list, err := api.Get[resource.ListResponse[resource.Cluster]](c, context.Background(), "clusters")
			if err != nil {
				return err
			}

			allConditions := make([][]out.Condition, 0, len(list.Items))
			for _, cl := range list.Items {
				var conds []out.Condition
				for _, cond := range cl.Status.Conditions {
					conds = append(conds, out.Condition{Type: cond.Type})
				}
				allConditions = append(allConditions, conds)
			}
			dynCols := out.DynamicColumns(allConditions)

			headers := append([]string{"NAME", "GEN"}, dynCols...)

			rows := make([][]string, 0, len(list.Items))
			for _, cl := range list.Items {
				condMap := make(map[string]string)
				for _, cond := range cl.Status.Conditions {
					condMap[cond.Type] = cond.Status
				}
				row := []string{
					cl.Name,
					fmt.Sprintf("%d", cl.Generation),
				}
				for _, col := range dynCols {
					row = append(row, p.Dot(condMap[col]))
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

// ── hf table (combined clusters + nodepools) ──────────────────────────────────

var tableCmd = &cobra.Command{
	Use:   "table",
	Short: "Combined table of all clusters and their nodepools with dynamic condition columns",
	RunE: func(cmd *cobra.Command, args []string) error {
		watchMode, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")
		c := newClient()
		p := printer()

		fetch := func() error {
			return renderCombinedTable(c, p)
		}

		if watchMode {
			return watch.Watch(interval, fetch)
		}
		return fetch()
	},
}

// renderCombinedTable fetches all clusters and their nodepools, then prints a flat table.
func renderCombinedTable(c *api.Client, p *out.Printer) error {
	ctx := context.Background()

	clusterList, err := api.Get[resource.ListResponse[resource.Cluster]](c, ctx, "clusters")
	if err != nil {
		return err
	}

	// Resolve each cluster's nodepools and collect all condition types for column ordering.
	type npEntry struct {
		np          resource.NodePool
		clusterName string
	}
	var allNPs []npEntry
	var allConditions [][]out.Condition

	for _, cl := range clusterList.Items {
		var conds []out.Condition
		for _, cond := range cl.Status.Conditions {
			conds = append(conds, out.Condition{Type: cond.Type})
		}
		allConditions = append(allConditions, conds)

		npList, err := api.Get[resource.ListResponse[resource.NodePool]](c, ctx, "clusters/"+cl.ID+"/nodepools")
		if err == nil {
			for _, np := range npList.Items {
				allNPs = append(allNPs, npEntry{np: np, clusterName: cl.Name})
				var npConds []out.Condition
				for _, cond := range np.Status.Conditions {
					npConds = append(npConds, out.Condition{Type: cond.Type})
				}
				allConditions = append(allConditions, npConds)
			}
		}
	}

	dynCols := out.DynamicColumns(allConditions)
	headers := append([]string{"NAME", "KIND", "CLUSTER", "GEN"}, dynCols...)

	var rows [][]string

	for _, cl := range clusterList.Items {
		condMap := make(map[string]string)
		for _, cond := range cl.Status.Conditions {
			condMap[cond.Type] = cond.Status
		}
		row := []string{
			cl.Name,
			"Cluster",
			"",
			fmt.Sprintf("%d", cl.Generation),
		}
		for _, col := range dynCols {
			row = append(row, p.Dot(condMap[col]))
		}
		rows = append(rows, row)

		// Append nodepools for this cluster immediately after.
		for _, entry := range allNPs {
			if entry.clusterName != cl.Name {
				continue
			}
			np := entry.np
			npCondMap := make(map[string]string)
			for _, cond := range np.Status.Conditions {
				npCondMap[cond.Type] = cond.Status
			}
			npRow := []string{
				np.Name,
				"NodePool",
				cl.Name,
				fmt.Sprintf("%d", np.Generation),
			}
			for _, col := range dynCols {
				npRow = append(npRow, p.Dot(npCondMap[col]))
			}
			rows = append(rows, npRow)
		}
	}

	return p.PrintTable(headers, rows)
}
