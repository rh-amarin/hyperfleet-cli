//go:build integration

package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

func TestIntegration_ClusterList(t *testing.T) {
	apiURL := os.Getenv("HF_API_URL")
	apiToken := os.Getenv("HF_API_TOKEN")
	if apiURL == "" || apiToken == "" {
		t.Skip("HF_API_URL and HF_API_TOKEN not set — skipping integration test")
	}

	c := api.NewClient(apiURL, "v1", apiToken, false)
	list, err := api.Get[resource.ListResponse[resource.Cluster]](c, context.Background(), "clusters")
	if err != nil {
		t.Fatalf("cluster list: %v", err)
	}
	t.Logf("cluster list returned %d items", len(list.Items))
}

func TestIntegration_ClusterCreate_And_Delete(t *testing.T) {
	apiURL := os.Getenv("HF_API_URL")
	apiToken := os.Getenv("HF_API_TOKEN")
	if apiURL == "" || apiToken == "" {
		t.Skip("HF_API_URL and HF_API_TOKEN not set — skipping integration test")
	}

	c := api.NewClient(apiURL, "v1", apiToken, false)
	ctx := context.Background()

	payload := map[string]any{
		"kind": "Cluster",
		"name": "integration-test-cluster",
		"labels": map[string]string{
			"counter":     "1",
			"environment": "integration",
			"shard":       "1",
			"team":        "core",
		},
		"spec": map[string]string{
			"counter": "1",
			"region":  "us-east-1",
			"version": "4.15.0",
		},
	}

	cluster, err := api.Post[resource.Cluster](c, ctx, "clusters", payload)
	if err != nil {
		t.Fatalf("cluster create: %v", err)
	}
	t.Logf("created cluster id=%s", cluster.ID)

	_, err = api.Delete[resource.Cluster](c, ctx, "clusters/"+cluster.ID)
	if err != nil {
		t.Errorf("cluster delete: %v", err)
	}
}
