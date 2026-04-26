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

// ListTopics returns all topics (with their subscriptions) for the project,
// applying an optional substring filter to topic and subscription short names.
// Filter logic mirrors the bash hf.pubsub.list.sh behaviour:
//   - If the topic name contains filter → include it with all its subscriptions.
//   - Otherwise → include only subscriptions whose name contains filter (if any).
func (c *GCPClient) ListTopics(ctx context.Context, filter string) ([]TopicGroup, error) {
	topicIt := c.client.Topics(ctx)
	var groups []TopicGroup

	for {
		topic, err := topicIt.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing topics: %w", err)
		}

		topicName := topic.ID()

		// Collect all subscription IDs for this topic.
		var allSubs []string
		subIt := topic.Subscriptions(ctx)
		for {
			sub, err := subIt.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("listing subscriptions for topic %q: %w", topicName, err)
			}
			allSubs = append(allSubs, sub.ID())
		}

		if filter == "" {
			groups = append(groups, TopicGroup{Name: topicName, Subscriptions: allSubs})
			continue
		}

		if strings.Contains(topicName, filter) {
			// Topic matches → show all its subscriptions.
			groups = append(groups, TopicGroup{Name: topicName, Subscriptions: allSubs})
		} else {
			// Check if any subscription matches.
			var matched []string
			for _, s := range allSubs {
				if strings.Contains(s, filter) {
					matched = append(matched, s)
				}
			}
			if len(matched) > 0 {
				groups = append(groups, TopicGroup{Name: topicName, Subscriptions: matched})
			}
		}
	}

	return groups, nil
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
