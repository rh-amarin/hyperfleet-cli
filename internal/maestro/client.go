package maestro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a Maestro HTTP API client. It is separate from internal/api
// because Maestro has its own base URL independent of the HyperFleet API.
type Client struct {
	httpClient *http.Client
	baseURL    string // "<http-endpoint>/api/maestro/v1/"
	consumer   string
	token      string
}

// New creates a Maestro client for the given endpoint and consumer.
func New(httpEndpoint, consumer, token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    httpEndpoint + "/api/maestro/v1/",
		consumer:   consumer,
		token:      token,
	}
}

// Resource represents a Maestro resource.
type Resource struct {
	ID            string      `json:"id"`
	Kind          string      `json:"kind"`
	Name          string      `json:"name"`
	ConsumerName  string      `json:"consumer_name"`
	Version       int         `json:"version"`
	ManifestCount int         `json:"manifest_count"`
	Manifests     []Manifest  `json:"manifests"`
	Conditions    []Condition `json:"conditions"`
}

// Manifest is a summary of a Kubernetes manifest within a resource.
type Manifest struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Condition is a status condition on a Maestro resource.
type Condition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// Bundle represents a Maestro resource bundle.
type Bundle struct {
	ID              string            `json:"id"`
	Kind            string            `json:"kind"`
	Name            string            `json:"name"`
	Labels          map[string]string `json:"labels"`
	Manifests       []any             `json:"manifests"`
	ManifestConfigs []any             `json:"manifest_configs"`
}

// Consumer represents a Maestro consumer.
type Consumer struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type listResponse[T any] struct {
	Items []T    `json:"items"`
	Kind  string `json:"kind"`
	Total int    `json:"total"`
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errBody)
	}
	return resp, nil
}

// List returns all resources for the configured consumer.
func (c *Client) List(ctx context.Context) ([]Resource, error) {
	path := "resources"
	if c.consumer != "" {
		path += "?consumer_name=" + c.consumer
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result listResponse[Resource]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// Get retrieves a single resource by name.
func (c *Client) Get(ctx context.Context, name string) (*Resource, error) {
	resp, err := c.do(ctx, http.MethodGet, "resources/"+name, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result Resource
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a resource by name.
func (c *Client) Delete(ctx context.Context, name string) error {
	resp, err := c.do(ctx, http.MethodDelete, "resources/"+name, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ListBundles returns all resource bundles.
func (c *Client) ListBundles(ctx context.Context) ([]Bundle, error) {
	resp, err := c.do(ctx, http.MethodGet, "resource-bundles", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result listResponse[Bundle]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// ListConsumers returns all consumers.
func (c *Client) ListConsumers(ctx context.Context) ([]Consumer, error) {
	resp, err := c.do(ctx, http.MethodGet, "consumers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result listResponse[Consumer]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}
