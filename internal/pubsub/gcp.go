package pubsub

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCPClient wraps a GCP Pub/Sub client.
type GCPClient struct {
	client    *pubsub.Client
	projectID string
}

// NewGCP creates a new GCP Pub/Sub client authenticated with the given OAuth2 access token.
func NewGCP(ctx context.Context, projectID, token string) (*GCPClient, error) {
	var opts []option.ClientOption
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		opts = append(opts, option.WithTokenSource(ts))
	}
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %w", err)
	}
	return &GCPClient{client: client, projectID: projectID}, nil
}

// ListSubscriptions returns full subscription resource names for the project,
// optionally filtered by substring match on the subscription ID.
func (c *GCPClient) ListSubscriptions(ctx context.Context, filter string) ([]string, error) {
	var result []string
	it := c.client.Subscriptions(ctx)
	for {
		sub, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing subscriptions: %w", err)
		}
		id := fmt.Sprintf("projects/%s/subscriptions/%s", c.projectID, sub.ID())
		if filter == "" || strings.Contains(sub.ID(), filter) {
			result = append(result, id)
		}
	}
	return result, nil
}

// Publish sends data to the named topic and returns the server-assigned message ID.
func (c *GCPClient) Publish(ctx context.Context, topicID string, data []byte) (string, error) {
	t := c.client.Topic(topicID)
	defer t.Stop()
	result := t.Publish(ctx, &pubsub.Message{Data: data})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("publish to topic %q: %w", topicID, err)
	}
	return id, nil
}

// Close releases the GCP Pub/Sub connection.
func (c *GCPClient) Close() error {
	return c.client.Close()
}
