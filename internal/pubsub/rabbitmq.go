package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RabbitClient publishes to RabbitMQ via the HTTP Management API.
// This mirrors the bash hf.rabbitmq.publish.cluster.change.sh implementation
// which uses POST /api/exchanges/{vhost}/{exchange}/publish.
type RabbitClient struct {
	baseURL    string // http://<host>:<mgmt-port>/api/exchanges/<vhost-enc>
	user       string
	password   string
	httpClient *http.Client
}

// NewRabbit creates a RabbitMQ HTTP Management API client.
// vhost "/" is URL-encoded to "%2F" as required by the RabbitMQ management API.
func NewRabbit(host string, mgmtPort int, user, password, vhost string) (*RabbitClient, error) {
	vhostEnc := url.PathEscape(vhost)
	if vhost == "/" {
		vhostEnc = "%2F"
	}
	baseURL := fmt.Sprintf("http://%s:%d/api/exchanges/%s", host, mgmtPort, vhostEnc)
	return &RabbitClient{
		baseURL:    baseURL,
		user:       user,
		password:   password,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type rabbitPublishRequest struct {
	Properties      map[string]any `json:"properties"`
	RoutingKey      string         `json:"routing_key"`
	Payload         string         `json:"payload"`
	PayloadEncoding string         `json:"payload_encoding"`
}

// Publish sends body (a JSON string) to the named exchange via the HTTP Management API.
// body is passed as-is in the "payload" field (string encoding).
func (c *RabbitClient) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	reqBody := rabbitPublishRequest{
		Properties:      map[string]any{},
		RoutingKey:      routingKey,
		Payload:         string(body),
		PayloadEncoding: "string",
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	endpoint := c.baseURL + "/" + exchange + "/publish"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RabbitMQ management API returned %d: %s", resp.StatusCode, respBody)
	}

	return nil
}

// Close is a no-op; HTTP Management API uses stateless requests.
func (c *RabbitClient) Close() error { return nil }
