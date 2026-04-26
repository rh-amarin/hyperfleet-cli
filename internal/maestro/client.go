package maestro

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

// Resource represents a Maestro resource bundle (resource-bundles endpoint).
type Resource struct {
	ID            string            `json:"id"`
	Kind          string            `json:"kind"`
	Metadata      map[string]string `json:"metadata"`
	ConsumerName  string            `json:"consumer_name"`
	Version       int               `json:"version"`
	ManifestCount int               `json:"manifest_count"`
	Manifests     []Manifest        `json:"manifests"`
	Conditions    []Condition       `json:"conditions"`
}

// Manifest is a summary of a Kubernetes manifest within a resource bundle.
type Manifest struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Condition is a status condition on a Maestro resource bundle.
type Condition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
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

// List returns all resource bundles for the configured consumer.
// The consumer filter uses the Maestro SQL-like search syntax.
func (c *Client) List(ctx context.Context) ([]Resource, error) {
	path := "resource-bundles"
	if c.consumer != "" {
		path += "?search=" + url.QueryEscape(fmt.Sprintf("consumer_name = '%s'", c.consumer))
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

// Get retrieves a single resource bundle by ID.
func (c *Client) Get(ctx context.Context, id string) (*Resource, error) {
	resp, err := c.do(ctx, http.MethodGet, "resource-bundles/"+id, nil)
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

// Delete removes a resource bundle by ID.
func (c *Client) Delete(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, "resource-bundles/"+id, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ListBundles returns all resource bundles without consumer filtering.
func (c *Client) ListBundles(ctx context.Context) ([]Resource, error) {
	resp, err := c.do(ctx, http.MethodGet, "resource-bundles", nil)
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
