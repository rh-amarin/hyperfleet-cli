package cmd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	ps "github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pubsubCmd)
}

var pubsubCmd = &cobra.Command{
	Use:   "pubsub",
	Short: "Manage GCP Pub/Sub topics and subscriptions",
}

func init() {
	pubsubCmd.AddCommand(pubsubListCmd)
	pubsubCmd.AddCommand(pubsubPublishCmd)
	pubsubPublishCmd.AddCommand(pubsubPublishClusterCmd)
	pubsubPublishCmd.AddCommand(pubsubPublishNodepoolCmd)
}

// gcpFactory creates a GCPPublisher. Overridable in tests.
var gcpFactory = func(ctx context.Context, projectID, token string) (ps.GCPPublisher, error) {
	return ps.NewGCP(ctx, projectID, token)
}

// ── list ──────────────────────────────────────────────────────────────────────

var pubsubListCmd = &cobra.Command{
	Use:          "list [filter]",
	Short:        "List GCP Pub/Sub subscriptions",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		cfg := cfgStore.Cfg()
		ctx := context.Background()

		client, err := gcpFactory(ctx, cfg.Hyperfleet.GCPProject, cfg.Hyperfleet.Token)
		if err != nil {
			return fmt.Errorf("create GCP client: %w", err)
		}
		defer client.Close()

		subs, err := client.ListSubscriptions(ctx, filter)
		if err != nil {
			return fmt.Errorf("list subscriptions: %w", err)
		}

		p := printer()
		headers := []string{"SUBSCRIPTION"}
		rows := make([][]string, len(subs))
		for i, s := range subs {
			rows[i] = []string{s}
		}
		return p.PrintTable(headers, rows)
	},
}

// ── publish ───────────────────────────────────────────────────────────────────

var pubsubPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a CloudEvent to a GCP Pub/Sub topic",
}

// cloudEvent wraps data in a CloudEvents 1.0 envelope.
func cloudEvent(eventType string, data any) ([]byte, error) {
	id, err := randomID()
	if err != nil {
		return nil, err
	}
	env := map[string]any{
		"specversion": "1.0",
		"type":        eventType,
		"source":      "/hyperfleet/cli",
		"id":          id,
		"data":        data,
	}
	return json.Marshal(env)
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

var pubsubPublishClusterCmd = &cobra.Command{
	Use:          "cluster <topic>",
	Short:        "Publish current cluster state to a GCP Pub/Sub topic",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		cfg := cfgStore.Cfg()
		ctx := context.Background()

		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}

		c := newClient()
		cluster, err := api.Get[resource.Cluster](c, ctx, "clusters/"+clusterID)
		if err != nil {
			return err
		}

		payload, err := cloudEvent("com.hyperfleet.cluster.changed", cluster)
		if err != nil {
			return err
		}

		gcpClient, err := gcpFactory(ctx, cfg.Hyperfleet.GCPProject, cfg.Hyperfleet.Token)
		if err != nil {
			return fmt.Errorf("create GCP client: %w", err)
		}
		defer gcpClient.Close()

		serverID, err := gcpClient.Publish(ctx, topic, payload)
		if err != nil {
			return fmt.Errorf("publish to GCP: %w", err)
		}

		out.Info(fmt.Sprintf("Published cluster %s to topic %s (msg-id: %s)", clusterID, topic, serverID))
		return nil
	},
}

var pubsubPublishNodepoolCmd = &cobra.Command{
	Use:          "nodepool <topic>",
	Short:        "Publish current nodepool state to a GCP Pub/Sub topic",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		cfg := cfgStore.Cfg()
		ctx := context.Background()

		clusterID, err := config.ClusterID(cfgStore, "")
		if err != nil {
			return err
		}
		nodepoolID, err := config.NodePoolID(cfgStore, "")
		if err != nil {
			return err
		}

		c := newClient()
		np, err := api.Get[resource.NodePool](c, ctx, "clusters/"+clusterID+"/nodepools/"+nodepoolID)
		if err != nil {
			return err
		}

		payload, err := cloudEvent("com.hyperfleet.nodepool.changed", np)
		if err != nil {
			return err
		}

		gcpClient, err := gcpFactory(ctx, cfg.Hyperfleet.GCPProject, cfg.Hyperfleet.Token)
		if err != nil {
			return fmt.Errorf("create GCP client: %w", err)
		}
		defer gcpClient.Close()

		serverID, err := gcpClient.Publish(ctx, topic, payload)
		if err != nil {
			return fmt.Errorf("publish to GCP: %w", err)
		}

		out.Info(fmt.Sprintf("Published nodepool %s to topic %s (msg-id: %s)", nodepoolID, topic, serverID))
		return nil
	},
}
