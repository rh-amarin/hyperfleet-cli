package api

import (
	"context"
	"encoding/json"
	"net/http"
)

func Get[T any](c *Client, ctx context.Context, path string) (*T, error) {
	resp, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decode[T](resp)
}

func Post[T any](c *Client, ctx context.Context, path string, body any) (*T, error) {
	resp, err := c.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decode[T](resp)
}

func Patch[T any](c *Client, ctx context.Context, path string, body any) (*T, error) {
	resp, err := c.Do(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decode[T](resp)
}

func Delete[T any](c *Client, ctx context.Context, path string) (*T, error) {
	resp, err := c.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decode[T](resp)
}

func decode[T any](resp *http.Response) (*T, error) {
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}
