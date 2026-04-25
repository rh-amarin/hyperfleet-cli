package api

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// FindNodePoolByName queries /clusters/{clusterID}/nodepools filtering by name and returns
// only exact, non-deleted matches. Used by both create (duplicate guard) and search.
func FindNodePoolByName(c *Client, ctx context.Context, clusterID, name string) ([]resource.NodePool, error) {
	path := "clusters/" + clusterID + "/nodepools?search=" + url.QueryEscape(fmt.Sprintf("name='%s'", name))
	list, err := Get[resource.ListResponse[resource.NodePool]](c, ctx, path)
	if err != nil {
		return nil, err
	}
	var matches []resource.NodePool
	for _, np := range list.Items {
		if np.Name == name && np.DeletedTime == "" {
			matches = append(matches, np)
		}
	}
	return matches, nil
}
