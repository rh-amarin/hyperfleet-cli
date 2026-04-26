package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	ps "github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
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
	Short:        "List GCP Pub/Sub topics and subscriptions",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		cfg := cfgStore.Cfg()
		ctx := context.Background()

		out.Info(fmt.Sprintf("Listing Pub/Sub topics and subscriptions for project: %s", cfg.Hyperfleet.GCPProject))
		if filter != "" {
			out.Info(fmt.Sprintf("Filtering by: %s", filter))
		}

		client, err := gcpFactory(ctx, cfg.Hyperfleet.GCPProject, cfg.Hyperfleet.Token)
		if err != nil {
			return fmt.Errorf("create GCP client: %w", err)
		}
		defer client.Close()

		groups, err := client.ListTopics(ctx, filter)
		if err != nil {
			return fmt.Errorf("list topics: %w", err)
		}

		if len(groups) == 0 {
			fmt.Println("No topics or subscriptions found.")
			return nil
		}

		for _, g := range groups {
			fmt.Println(g.Name)
			for _, s := range g.Subscriptions {
				fmt.Printf("    %s\n", s)
			}
		}
		return nil
	},
}

// ── CloudEvent builders ───────────────────────────────────────────────────────

type cloudEventData struct {
	ID              string           `json:"id"`
	Kind            string           `json:"kind"`
	Href            string           `json:"href"`
	Generation      int              `json:"generation"`
	OwnerReferences *ownerReferences `json:"owner_references,omitempty"`
}

type ownerReferences struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Href       string `json:"href"`
	Generation int    `json:"generation"`
}

type cloudEventEnvelope struct {
	SpecVersion     string         `json:"specversion"`
	Type            string         `json:"type"`
	Source          string         `json:"source"`
	ID              string         `json:"id"`
	Time            string         `json:"time"`
	DataContentType string         `json:"datacontenttype"`
	Data            cloudEventData `json:"data"`
}

func buildClusterEvent(clusterID string) ([]byte, error) {
	env := cloudEventEnvelope{
		SpecVersion:     "1.0",
		Type:            "com.redhat.hyperfleet.cluster.reconcile.v1",
		Source:          "/hyperfleet/service/sentinel",
		ID:              clusterID,
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: cloudEventData{
			ID:         clusterID,
			Kind:       "Cluster",
			Href:       fmt.Sprintf("https://api.hyperfleet.com/v1/clusters/%s", clusterID),
			Generation: 1,
		},
	}
	return json.MarshalIndent(env, "", "  ")
}

func buildNodePoolEvent(clusterID, nodepoolID string) ([]byte, error) {
	env := cloudEventEnvelope{
		SpecVersion:     "1.0",
		Type:            "com.redhat.hyperfleet.nodepool.reconcile.v1",
		Source:          "/hyperfleet/service/sentinel",
		ID:              nodepoolID,
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: cloudEventData{
			ID:         nodepoolID,
			Kind:       "NodePool",
			Href:       fmt.Sprintf("http://localhost:8000/api/hyperfleet/v1/clusters/%s/node_pools/%s", clusterID, nodepoolID),
			Generation: 1,
			OwnerReferences: &ownerReferences{
				ID:         clusterID,
				Kind:       "NodePool",
				Href:       fmt.Sprintf("http://localhost:8000/api/hyperfleet/v1/clusters/%s", clusterID),
				Generation: 1,
			},
		},
	}
	return json.MarshalIndent(env, "", "  ")
}

// ── publish ───────────────────────────────────────────────────────────────────

var pubsubPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a CloudEvent to a GCP Pub/Sub topic",
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

		payload, err := buildClusterEvent(clusterID)
		if err != nil {
			return err
		}

		out.Info(fmt.Sprintf("Publishing change message to topic: %s", topic))
		fmt.Println(string(payload))

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

		payload, err := buildNodePoolEvent(clusterID, nodepoolID)
		if err != nil {
			return err
		}

		out.Info(fmt.Sprintf("Publishing change message to topic: %s", topic))
		fmt.Println(string(payload))

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
