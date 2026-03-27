package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

type Consumer struct {
	client  *Client
	handlers map[string]MessageHandler
	wg      sync.WaitGroup
}

type MessageHandler func([]byte) error

func NewConsumer(client *Client) *Consumer {
	return &Consumer{
		client:  client,
		handlers: make(map[string]MessageHandler),
	}
}

func (c *Consumer) RegisterHandler(queue string, handler MessageHandler) {
	c.handlers[queue] = handler
}

func (c *Consumer) Start() error {
	for queue, handler := range c.handlers {
		if err := c.client.Consume(queue, handler); err != nil {
			return fmt.Errorf("start consumer for %s: %w", queue, err)
		}
		log.Printf("Consumer started for queue: %s", queue)
	}
	return nil
}

func (c *Consumer) Stop() {
	c.wg.Wait()
}

type OrderPaymentHandler struct {
	orderRepo interface {
		GetByID(id uint) (interface{}, error)
		Update(interface{}) error
	}
	userRepo interface {
		GetByID(id uint) (interface{}, error)
		Update(interface{}) error
	}
}

func ParseOrderPaymentMessage(data []byte) (*OrderPaymentMessage, error) {
	var msg OrderPaymentMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

type EmailHandler struct{}

func ParseEmailMessage(data []byte) (*EmailMessage, error) {
	var msg EmailMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

type UsageLogHandler struct{}

func ParseUsageLogMessage(data []byte) (*UsageLogMessage, error) {
	var msg UsageLogMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

type VIPExpireHandler struct{}

func ParseVIPExpireMessage(data []byte) (*VIPExpireMessage, error) {
	var msg VIPExpireMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func DefaultUsageLogHandler(data []byte) error {
	msg, err := ParseUsageLogMessage(data)
	if err != nil {
		return err
	}

	log.Printf("Usage log: user=%d token=%d channel=%d model=%s input=%d output=%d cost=%.4f",
		msg.UserID, msg.TokenID, msg.ChannelID, msg.Model,
		msg.InputTokens, msg.OutputTokens, msg.Cost)

	return nil
}

func DefaultEmailHandler(data []byte) error {
	msg, err := ParseEmailMessage(data)
	if err != nil {
		return err
	}

	log.Printf("Email: to=%s subject=%s type=%s", msg.To, msg.Subject, msg.Type)

	return nil
}

func DefaultOrderPaymentHandler(data []byte) error {
	msg, err := ParseOrderPaymentMessage(data)
	if err != nil {
		return err
	}

	log.Printf("Order payment: order_id=%d order_no=%s user_id=%d amount=%.2f",
		msg.OrderID, msg.OrderNo, msg.UserID, msg.Amount)

	return nil
}

func DefaultVIPExpireHandler(data []byte) error {
	msg, err := ParseVIPExpireMessage(data)
	if err != nil {
		return err
	}

	log.Printf("VIP expire: user_id=%d expired_at=%s", msg.UserID, msg.ExpiredAt.Format("2006-01-02 15:04:05"))

	return nil
}

func SetupDefaultConsumer(client *Client) (*Consumer, error) {
	consumer := NewConsumer(client)

	consumer.RegisterHandler(QueueUsageLog, DefaultUsageLogHandler)
	consumer.RegisterHandler(QueueEmailSend, DefaultEmailHandler)
	consumer.RegisterHandler(QueueOrderPayment, DefaultOrderPaymentHandler)
	consumer.RegisterHandler(QueueVIPExpire, DefaultVIPExpireHandler)

	if err := consumer.Start(); err != nil {
		return nil, err
	}

	return consumer, nil
}
