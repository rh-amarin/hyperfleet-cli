package api

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// FindClusterByName queries /clusters filtering by name and returns only exact,
// non-deleted matches. Used by both create (duplicate guard) and search.
func FindClusterByName(c *Client, ctx context.Context, name string) ([]resource.Cluster, error) {
	path := "clusters?search=" + url.QueryEscape(fmt.Sprintf("name='%s'", name))
	list, err := Get[resource.ListResponse[resource.Cluster]](c, ctx, path)
	if err != nil {
		return nil, err
	}
	var matches []resource.Cluster
	for _, cl := range list.Items {
		if cl.Name == name && cl.DeletedTime == "" {
			matches = append(matches, cl)
		}
	}
	return matches, nil
}
