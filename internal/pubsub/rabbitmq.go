package pubsub

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitClient wraps an AMQP connection.
type RabbitClient struct {
	conn *amqp.Connection
}

// NewRabbit dials the given AMQP URL and returns a RabbitClient.
// url format: amqp://<user>:<pass>@<host>/<vhost>
func NewRabbit(url string) (*RabbitClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("amqp.Dial: %w", err)
	}
	return &RabbitClient{conn: conn}, nil
}

// Publish sends body to the named exchange with the given routing key.
// A new channel is opened per call; channels are cheap and this avoids concurrency issues.
func (c *RabbitClient) Publish(_ context.Context, exchange, routingKey string, body []byte) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()
	return ch.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// Close releases the AMQP connection.
func (c *RabbitClient) Close() error {
	return c.conn.Close()
}
