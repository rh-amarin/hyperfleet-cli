package config

import "fmt"

// ClusterID resolves the cluster ID: explicit arg > state.yaml > error.
func ClusterID(s *Store, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	if id := s.State().ClusterID; id != "" {
		return id, nil
	}
	return "", fmt.Errorf("no cluster-id set — run 'hf cluster search <name>' or pass a cluster ID")
}

// SetClusterID writes cluster-id and cluster-name to state.yaml.
func SetClusterID(s *Store, id, name string) error {
	if err := s.SetState("cluster-id", id); err != nil {
		return err
	}
	return s.SetState("cluster-name", name)
}

// NodePoolID resolves the nodepool ID: explicit arg > state.yaml > error.
func NodePoolID(s *Store, arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	if id := s.State().NodePoolID; id != "" {
		return id, nil
	}
	return "", fmt.Errorf("no nodepool-id set — run 'hf nodepool search <name>' or pass a nodepool ID")
}

// SetNodePoolID writes nodepool-id to state.yaml.
func SetNodePoolID(s *Store, id string) error {
	return s.SetState("nodepool-id", id)
}
