package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

var validStatuses = map[string]bool{
	"True":    true,
	"False":   true,
	"Unknown": true,
}

func buildAdapterPostPayload(adapter, status string, generation int32) resource.AdapterStatusCreateRequest {
	now := time.Now().UTC().Format(time.RFC3339)
	msg := "Status posted via hf adapter post-status"
	return resource.AdapterStatusCreateRequest{
		Adapter:            adapter,
		ObservedGeneration: generation,
		ObservedTime:       now,
		Conditions: []resource.ConditionRequest{
			{Type: "Available", Status: status, Reason: "ManualStatusPost", Message: msg, LastTransitionTime: now},
			{Type: "Applied", Status: status, Reason: "ManualStatusPost", Message: msg, LastTransitionTime: now},
			{Type: "Health", Status: status, Reason: "ManualStatusPost", Message: msg, LastTransitionTime: now},
		},
	}
}

func init() {
	clusterCmd.AddCommand(clusterAdapterCmd)
	clusterAdapterCmd.AddCommand(clusterAdapterPostStatusCmd)

	nodepoolCmd.AddCommand(nodepoolAdapterCmd)
	nodepoolAdapterCmd.AddCommand(nodepoolAdapterPostStatusCmd)
}

// ── cluster adapter ────────────────────────────────────────────────────────────

var clusterAdapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Adapter operations for a cluster",
}

var clusterAdapterPostStatusCmd = &cobra.Command{
	Use:          "post-status <adapter> <True|False|Unknown> [generation]",
	Short:        "Post adapter status conditions for the current cluster",
	Args:         cobra.RangeArgs(2, 3),
	SilenceUsage: true,
	RunE:         runClusterAdapterPostStatus,
}

func runClusterAdapterPostStatus(_ *cobra.Command, args []string) error {
	adapter := args[0]
	status := args[1]
	generation := int32(1)

	if !validStatuses[status] {
		return fmt.Errorf("status must be one of: True, False, Unknown (got %q)", status)
	}
	if len(args) == 3 {
		n, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("generation must be an integer: %w", err)
		}
		generation = int32(n)
	}

	clusterID, err := config.ClusterID(cfgStore, "")
	if err != nil {
		return err
	}

	out.Info(fmt.Sprintf("Posting adapter status: %s = %s (gen: %d) for cluster: %s", adapter, status, generation, clusterID))

	payload := buildAdapterPostPayload(adapter, status, generation)
	c := newClient()
	resp, err := api.Post[resource.AdapterStatus](c, context.Background(), "clusters/"+clusterID+"/statuses", payload)
	if err != nil {
		return err
	}
	return printer().Print(resp)
}

// ── nodepool adapter ───────────────────────────────────────────────────────────

var nodepoolAdapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Adapter operations for a nodepool",
}

var nodepoolAdapterPostStatusCmd = &cobra.Command{
	Use:          "post-status <adapter> <True|False|Unknown> [generation] [nodepool_id]",
	Short:        "Post adapter status conditions for the current nodepool",
	Args:         cobra.RangeArgs(2, 4),
	SilenceUsage: true,
	RunE:         runNodePoolAdapterPostStatus,
}

func runNodePoolAdapterPostStatus(_ *cobra.Command, args []string) error {
	adapter := args[0]
	status := args[1]
	generation := int32(1)
	nodepoolArg := ""

	if !validStatuses[status] {
		return fmt.Errorf("status must be one of: True, False, Unknown (got %q)", status)
	}
	if len(args) >= 3 {
		n, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("generation must be an integer: %w", err)
		}
		generation = int32(n)
	}
	if len(args) == 4 {
		nodepoolArg = args[3]
	}

	clusterID, err := config.ClusterID(cfgStore, "")
	if err != nil {
		return err
	}
	nodepoolID, err := config.NodePoolID(cfgStore, nodepoolArg)
	if err != nil {
		return err
	}

	out.Info(fmt.Sprintf("Posting adapter status: %s = %s (gen: %d) for nodepool: %s", adapter, status, generation, nodepoolID))

	payload := buildAdapterPostPayload(adapter, status, generation)
	c := newClient()
	path := "clusters/" + clusterID + "/nodepools/" + nodepoolID + "/statuses"
	resp, err := api.Post[resource.AdapterStatus](c, context.Background(), path, payload)
	if err != nil {
		return err
	}
	return printer().Print(resp)
}
