package pubsub

import "context"

// TopicGroup holds a topic name and its subscription names (short IDs).
type TopicGroup struct {
	Name          string
	Subscriptions []string
}

// GCPPublisher lists topics+subscriptions and publishes messages to GCP Pub/Sub topics.
type GCPPublisher interface {
	ListTopics(ctx context.Context, filter string) ([]TopicGroup, error)
	Publish(ctx context.Context, topicID string, data []byte) (string, error)
	Close() error
}

// RabbitPublisher publishes messages to a RabbitMQ exchange via HTTP Management API.
type RabbitPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Close() error
}
