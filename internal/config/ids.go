package config

import "fmt"

// ClusterID resolves the cluster ID: explicit arg > config file > error.
func ClusterID(s *Store, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	if id := s.Get("cluster-id"); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("no cluster-id set — run 'hf cluster search <name>' or pass a cluster ID")
}

// SetClusterID writes cluster-id and cluster-name to the store.
func SetClusterID(s *Store, id, name string) error {
	if err := s.Set("cluster-id", id); err != nil {
		return err
	}
	return s.Set("cluster-name", name)
}

// NodePoolID resolves the nodepool ID: explicit arg > config file > error.
func NodePoolID(s *Store, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	if id := s.Get("nodepool-id"); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("no nodepool-id set — run 'hf nodepool search <name>' or pass a nodepool ID")
}

// SetNodePoolID writes nodepool-id to the store.
func SetNodePoolID(s *Store, id string) error {
	return s.Set("nodepool-id", id)
}
