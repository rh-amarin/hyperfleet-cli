package pubsub

import "context"

// GCPPublisher can list GCP Pub/Sub subscriptions and publish messages to topics.
type GCPPublisher interface {
	ListSubscriptions(ctx context.Context, filter string) ([]string, error)
	Publish(ctx context.Context, topicID string, data []byte) (string, error)
	Close() error
}

// RabbitPublisher publishes messages to a RabbitMQ exchange.
type RabbitPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Close() error
}
