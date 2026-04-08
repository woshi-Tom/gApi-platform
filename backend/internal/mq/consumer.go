package mq

import (
	"encoding/json"
	"fmt"
	"sync"

	"gapi-platform/internal/logger"
)

type Consumer struct {
	client   *Client
	handlers map[string]MessageHandler
	wg       sync.WaitGroup
}

type MessageHandler func([]byte) error

func NewConsumer(client *Client) *Consumer {
	return &Consumer{
		client:   client,
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
		logger.Info("Consumer started", "queue", queue)
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

	logger.Info("Usage logged",
		"user_id", msg.UserID,
		"token_id", msg.TokenID,
		"channel_id", msg.ChannelID,
		"model", msg.Model,
		"input_tokens", msg.InputTokens,
		"output_tokens", msg.OutputTokens,
		"cost", msg.Cost)

	return nil
}

func DefaultEmailHandler(data []byte) error {
	msg, err := ParseEmailMessage(data)
	if err != nil {
		return err
	}

	logger.Info("Email queued",
		"to", logger.RedactEmail(msg.To),
		"subject", msg.Subject,
		"type", msg.Type)

	return nil
}

func DefaultOrderPaymentHandler(data []byte) error {
	msg, err := ParseOrderPaymentMessage(data)
	if err != nil {
		return err
	}

	logger.Info("Order payment processed",
		"order_id", msg.OrderID,
		"order_no", msg.OrderNo,
		"user_id", msg.UserID,
		"amount", msg.Amount)

	return nil
}

func DefaultVIPExpireHandler(data []byte) error {
	msg, err := ParseVIPExpireMessage(data)
	if err != nil {
		return err
	}

	logger.Info("VIP expired",
		"user_id", msg.UserID,
		"expired_at", msg.ExpiredAt.Format("2006-01-02 15:04:05"))

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
