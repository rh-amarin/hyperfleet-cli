package cmd

import (
	"context"
	"fmt"
	"sort"
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
				genMap := make(map[string]int32)
				for _, cond := range cl.Status.Conditions {
					condMap[cond.Type] = cond.Status
					genMap[cond.Type] = cond.ObservedGeneration
				}
				row := []string{
					cl.Name,
					fmt.Sprintf("%d", cl.Generation),
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

// renderCombinedTable fetches all clusters and their nodepools, then prints a combined table.
// Fixed columns: ID, NAME, GEN. Dynamic condition columns follow, then adapter columns.
// Nodepool rows are indented with two spaces on ID and NAME to express hierarchy.
func renderCombinedTable(c *api.Client, p *out.Printer) error {
	ctx := context.Background()

	clusterList, err := api.Get[resource.ListResponse[resource.Cluster]](c, ctx, "clusters")
	if err != nil {
		return err
	}

	type npData struct {
		np       resource.NodePool
		statuses []resource.AdapterStatus
	}
	type clData struct {
		cl        resource.Cluster
		statuses  []resource.AdapterStatus
		nodepools []npData
	}

	var entries []clData
	var allConditions [][]out.Condition
	adapterNameSeen := map[string]bool{}
	var adapterNames []string

	for _, cl := range clusterList.Items {
		var conds []out.Condition
		for _, cond := range cl.Status.Conditions {
			conds = append(conds, out.Condition{Type: cond.Type})
		}
		allConditions = append(allConditions, conds)

		clStatuses, _ := api.Get[resource.ListResponse[resource.AdapterStatus]](c, ctx, "clusters/"+cl.ID+"/statuses")
		for _, s := range clStatuses.Items {
			if !adapterNameSeen[s.Adapter] {
				adapterNameSeen[s.Adapter] = true
				adapterNames = append(adapterNames, s.Adapter)
			}
		}

		var nps []npData
		npList, err := api.Get[resource.ListResponse[resource.NodePool]](c, ctx, "clusters/"+cl.ID+"/nodepools")
		if err == nil {
			for _, np := range npList.Items {
				var npConds []out.Condition
				for _, cond := range np.Status.Conditions {
					npConds = append(npConds, out.Condition{Type: cond.Type})
				}
				allConditions = append(allConditions, npConds)

				npStatuses, _ := api.Get[resource.ListResponse[resource.AdapterStatus]](c, ctx, "clusters/"+cl.ID+"/nodepools/"+np.ID+"/statuses")
				for _, s := range npStatuses.Items {
					if !adapterNameSeen[s.Adapter] {
						adapterNameSeen[s.Adapter] = true
						adapterNames = append(adapterNames, s.Adapter)
					}
				}
				nps = append(nps, npData{np: np, statuses: npStatuses.Items})
			}
		}

		entries = append(entries, clData{cl: cl, statuses: clStatuses.Items, nodepools: nps})
	}

	sort.Strings(adapterNames)
	dynCols := out.DynamicColumns(allConditions)

	headers := append([]string{"ID", "NAME", "GEN"}, dynCols...)
	headers = append(headers, adapterNames...)

	var rows [][]string

	for _, entry := range entries {
		cl := entry.cl
		deleted := cl.DeletedTime != ""

		condMap := make(map[string]string)
		genMap := make(map[string]int32)
		for _, cond := range cl.Status.Conditions {
			condMap[cond.Type] = cond.Status
			genMap[cond.Type] = cond.ObservedGeneration
		}

		row := []string{cl.ID, cl.Name, p.GenCell(cl.Generation, deleted)}
		for _, col := range dynCols {
			row = append(row, p.DotWithGen(condMap[col], genMap[col]))
		}
		adpMap := adapterCells(entry.statuses, deleted)
		for _, name := range adapterNames {
			cell := adpMap[name]
			row = append(row, p.DotWithGen(cell.status, cell.gen))
		}
		rows = append(rows, row)

		for _, npd := range entry.nodepools {
			np := npd.np
			npDeleted := np.DeletedTime != ""

			npCondMap := make(map[string]string)
			npGenMap := make(map[string]int32)
			for _, cond := range np.Status.Conditions {
				npCondMap[cond.Type] = cond.Status
				npGenMap[cond.Type] = cond.ObservedGeneration
			}

			npRow := []string{"  " + np.ID, "  " + np.Name, p.GenCell(np.Generation, npDeleted)}
			for _, col := range dynCols {
				npRow = append(npRow, p.DotWithGen(npCondMap[col], npGenMap[col]))
			}
			npAdpMap := adapterCells(npd.statuses, npDeleted)
			for _, name := range adapterNames {
				cell := npAdpMap[name]
				npRow = append(npRow, p.DotWithGen(cell.status, cell.gen))
			}
			rows = append(rows, npRow)
		}
	}

	return p.PrintTable(headers, rows)
}

type adpCell struct {
	status string
	gen    int32
}

// adapterCells returns a map from adapter name to the chosen condition status and generation.
// For deleted resources (deleted_time set) it uses the Finalized condition; otherwise Available.
func adapterCells(statuses []resource.AdapterStatus, deleted bool) map[string]adpCell {
	condType := "Available"
	if deleted {
		condType = "Finalized"
	}
	m := make(map[string]adpCell, len(statuses))
	for _, s := range statuses {
		for _, cond := range s.Conditions {
			if cond.Type == condType {
				m[s.Adapter] = adpCell{status: cond.Status, gen: s.ObservedGeneration}
				break
			}
		}
	}
	return m
}
