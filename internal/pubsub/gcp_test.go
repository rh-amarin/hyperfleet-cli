package pubsub

import "testing"

// Verify GCPClient satisfies GCPPublisher at compile time.
var _ GCPPublisher = (*GCPClient)(nil)

func TestGCPClient_InterfaceCompliance(t *testing.T) {
	t.Log("GCPClient implements GCPPublisher")
}
