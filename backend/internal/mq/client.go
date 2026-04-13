package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueOrderPayment = "order.payment"
	QueueOrderNotify  = "order.notify"
	QueueEmailSend    = "email.send"
	QueueUsageLog     = "usage.log"
	QueueVIPExpire    = "vip.expire"
	QueueHealthCheck  = "health.check"
)

var (
	defaultClient *Client
	once          sync.Once
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.RabbitMQConfig
	mu      sync.RWMutex
}

func NewClient(cfg *config.RabbitMQConfig) (*Client, error) {
	client := &Client{config: cfg}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func DefaultClient() *Client {
	return defaultClient
}

func SetDefaultClient(client *Client) {
	defaultClient = client
}

func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	url := c.config.Addr()
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	c.conn = conn
	c.channel = ch

	return nil
}

func (c *Client) Reconnect() error {
	if c.conn != nil && !c.conn.IsClosed() {
		return nil
	}
	return c.connect()
}

func (c *Client) Channel() (*amqp.Channel, error) {
	if err := c.Reconnect(); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channel, nil
}

func (c *Client) DeclareQueue(name string, durable bool) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		name,
		durable,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", name, err)
	}

	return nil
}

func (c *Client) DeclareAllQueues() error {
	queues := []string{
		QueueOrderPayment,
		QueueOrderNotify,
		QueueEmailSend,
		QueueUsageLog,
		QueueVIPExpire,
		QueueHealthCheck,
	}

	for _, q := range queues {
		if err := c.DeclareQueue(q, true); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Publish(queue string, msg interface{}) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = ch.PublishWithContext(
		context.Background(),
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", queue, err)
	}

	return nil
}

func (c *Client) Consume(queue string, handler func([]byte) error) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume from %s: %w", queue, err)
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				logger.Errorf("Handler error for %s: %v", queue, err)
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && !c.conn.IsClosed()
}
