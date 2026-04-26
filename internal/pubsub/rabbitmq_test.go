package pubsub

import "testing"

// Verify RabbitClient satisfies RabbitPublisher at compile time.
var _ RabbitPublisher = (*RabbitClient)(nil)

func TestRabbitClient_InterfaceCompliance(t *testing.T) {
	t.Log("RabbitClient implements RabbitPublisher")
}
