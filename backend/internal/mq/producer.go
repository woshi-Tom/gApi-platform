package mq

import (
	"encoding/json"
	"time"
)

type OrderPaymentMessage struct {
	OrderID   uint    `json:"order_id"`
	OrderNo   string  `json:"order_no"`
	UserID    uint    `json:"user_id"`
	Amount    float64 `json:"amount"`
	PackageID uint    `json:"package_id"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderNotifyMessage struct {
	OrderID   uint   `json:"order_id"`
	OrderNo   string `json:"order_no"`
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	NotifyType string `json:"notify_type"`
}

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Type    string `json:"type"`
}

type UsageLogMessage struct {
	UserID    uint   `json:"user_id"`
	TokenID   uint   `json:"token_id"`
	ChannelID uint  `json:"channel_id"`
	Model     string `json:"model"`
	InputTokens int  `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	Cost      float64 `json:"cost"`
	Timestamp time.Time `json:"timestamp"`
}

type VIPExpireMessage struct {
	UserID    uint   `json:"user_id"`
	ExpiredAt time.Time `json:"expired_at"`
}

type HealthCheckMessage struct {
	ChannelID uint      `json:"channel_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Producer struct {
	client *Client
}

func NewProducer(client *Client) *Producer {
	return &Producer{client: client}
}

func (p *Producer) PublishOrderPayment(msg *OrderPaymentMessage) error {
	return p.client.Publish(QueueOrderPayment, msg)
}

func (p *Producer) PublishOrderNotify(msg *OrderNotifyMessage) error {
	return p.client.Publish(QueueOrderNotify, msg)
}

func (p *Producer) PublishEmail(msg *EmailMessage) error {
	return p.client.Publish(QueueEmailSend, msg)
}

func (p *Producer) PublishUsageLog(msg *UsageLogMessage) error {
	return p.client.Publish(QueueUsageLog, msg)
}

func (p *Producer) PublishVIPExpire(msg *VIPExpireMessage) error {
	return p.client.Publish(QueueVIPExpire, msg)
}

func (p *Producer) PublishHealthCheck(msg *HealthCheckMessage) error {
	return p.client.Publish(QueueHealthCheck, msg)
}

func (p *Producer) SendWelcomeEmail(userID uint, email, username string) error {
	msg := &EmailMessage{
		To:      email,
		Subject: "Welcome to gAPI Platform",
		Body:    "Hello " + username + ",\n\nWelcome to gAPI Platform! Start exploring our AI services today.",
		Type:    "welcome",
	}
	return p.PublishEmail(msg)
}

func (p *Producer) SendOrderConfirmation(orderID uint, orderNo string, userID uint, email string) error {
	msg := &EmailMessage{
		To:      email,
		Subject: "Order Confirmed - " + orderNo,
		Body:    "Your order " + orderNo + " has been confirmed.",
		Type:    "order_confirm",
	}
	notifyMsg := &OrderNotifyMessage{
		OrderID:    orderID,
		OrderNo:    orderNo,
		UserID:     userID,
		Email:      email,
		NotifyType: "order_confirm",
	}
	if err := p.PublishEmail(msg); err != nil {
		return err
	}
	return p.PublishOrderNotify(notifyMsg)
}

func (p *Producer) LogAPIUsage(userID, tokenID, channelID uint, model string, inputTokens, outputTokens int, cost float64) error {
	msg := &UsageLogMessage{
		UserID:     userID,
		TokenID:    tokenID,
		ChannelID:  channelID,
		Model:      model,
		InputTokens: inputTokens,
		OutputTokens: outputTokens,
		Cost:       cost,
		Timestamp:  time.Now(),
	}
	return p.PublishUsageLog(msg)
}

func MarshalMessage(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
