package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	verbose    bool
}

func NewClient(apiURL, apiVersion, token string, verbose bool) *Client {
	return &Client{
		baseURL:    fmt.Sprintf("%s/api/hyperfleet/%s/", apiURL, apiVersion),
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		verbose:    verbose,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s → %d (%dms)\n", method, url, resp.StatusCode, time.Since(start).Milliseconds())
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, parseError(resp)
	}

	return resp, nil
}
