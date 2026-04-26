package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	ps "github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rabbitmqCmd)
}

var rabbitmqCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "Manage RabbitMQ messaging",
}

func init() {
	rabbitmqCmd.AddCommand(rabbitmqPublishCmd)
	rabbitmqPublishCmd.AddCommand(rabbitmqPublishClusterCmd)
}

// rabbitFactory creates a RabbitPublisher. Overridable in tests.
var rabbitFactory = func(amqpURL string) (ps.RabbitPublisher, error) {
	return ps.NewRabbit(amqpURL)
}

// buildRabbitURL constructs an AMQP URL from the resolved config.
func buildRabbitURL(cfg config.Config) string {
	vhost := cfg.RabbitMQ.VHost
	if vhost == "/" {
		vhost = ""
	}
	u := &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(cfg.RabbitMQ.User, cfg.RabbitMQ.Password),
		Host:   cfg.RabbitMQ.Host,
		Path:   "/" + vhost,
	}
	return u.String()
}

// ── publish ───────────────────────────────────────────────────────────────────

var rabbitmqPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish messages to RabbitMQ",
}

var rabbitmqPublishClusterCmd = &cobra.Command{
	Use:          "cluster <exchange> [routing-key]",
	Short:        "Publish current cluster state to a RabbitMQ exchange",
	Args:         cobra.RangeArgs(1, 2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		exchange := args[0]
		routingKey := ""
		if len(args) > 1 {
			routingKey = args[1]
		}

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

		rabbitClient, err := rabbitFactory(buildRabbitURL(cfg))
		if err != nil {
			return fmt.Errorf("create RabbitMQ client: %w", err)
		}
		defer rabbitClient.Close()

		if err := rabbitClient.Publish(ctx, exchange, routingKey, payload); err != nil {
			return fmt.Errorf("publish to RabbitMQ: %w", err)
		}

		out.Info(fmt.Sprintf("Published cluster %s to exchange %s (routing-key: %q)", clusterID, exchange, routingKey))
		return nil
	},
}
