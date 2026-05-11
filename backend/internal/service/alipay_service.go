package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"

	"github.com/smartwalle/alipay/v3"
)

type AlipayService struct {
	settingsSvc *SettingsService
	orderRepo   *repository.OrderRepository
	paymentRepo *repository.PaymentRepository
	userRepo    *repository.UserRepository
	client      *alipay.Client
	clientMu    sync.RWMutex
	notifyURL   string
	serverMode  string
}

func NewAlipayService(
	settingsSvc *SettingsService,
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	userRepo *repository.UserRepository,
	serverMode string,
	notifyURL string,
) *AlipayService {
	return &AlipayService{
		settingsSvc: settingsSvc,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		serverMode:  serverMode,
		notifyURL:   notifyURL,
	}
}

func (s *AlipayService) getClient() (*alipay.Client, error) {
	s.clientMu.RLock()
	if s.client != nil {
		client := s.client
		s.clientMu.RUnlock()
		return client, nil
	}
	s.clientMu.RUnlock()

	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	cfg, err := s.settingsSvc.GetAlipayConfig()
	if err != nil || !cfg.Enabled || cfg.AppID == "" || cfg.PrivateKey == "" {
		return nil, errors.New("alipay not configured")
	}

	isProduction := s.serverMode == "production"

	var client *alipay.Client
	if cfg.Sandbox {
		client, err = alipay.New(cfg.AppID, cfg.PrivateKey, false, alipay.WithNewSandboxGateway())
	} else {
		client, err = alipay.New(cfg.AppID, cfg.PrivateKey, isProduction)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay client: %w", err)
	}

	if cfg.PublicKey != "" {
		err = client.LoadAliPayPublicKey(cfg.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load alipay public key: %w", err)
		}
	}

	if cfg.EncryptKey != "" {
		client.SetEncryptKey(cfg.EncryptKey)
	}

	s.client = client
	return client, nil
}

func (s *AlipayService) reloadClient() {
	s.clientMu.Lock()
	s.client = nil
	s.clientMu.Unlock()
}

func (s *AlipayService) IsEnabled() bool {
	cfg, err := s.settingsSvc.GetAlipayConfig()
	if err != nil {
		return false
	}
	return cfg.Enabled && cfg.AppID != "" && cfg.PrivateKey != ""
}

func (s *AlipayService) CreatePayment(orderNo string, amount float64, subject string) (string, string, error) {
	client, err := s.getClient()
	if err != nil {
		return "", "", errors.New("alipay not enabled or not configured")
	}

	timeout := 15 * time.Minute
	expireTime := time.Now().Add(timeout).Format("2006-01-02 15:04:05")

	var p = alipay.TradePreCreate{}
	p.OutTradeNo = orderNo
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	p.TimeoutExpress = "15m"

	resp, err := client.TradePreCreate(context.Background(), p)
	if err != nil {
		return "", "", fmt.Errorf("failed to create trade: %w", err)
	}

	if resp.IsFailure() {
		return "", "", fmt.Errorf("alipay error: %s - %s", resp.Code, resp.Msg)
	}

	if resp.QRCode == "" {
		return "", "", fmt.Errorf("alipay returned empty QR code (code: %s, msg: %s)", resp.Code, resp.Msg)
	}

	return resp.QRCode, expireTime, nil
}

func (s *AlipayService) QueryOrder(outTradeNo string) (*AlipayQueryResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, errors.New("alipay not enabled")
	}

	var p = alipay.TradeQuery{}
	p.OutTradeNo = outTradeNo

	resp, err := client.TradeQuery(context.Background(), p)
	if err != nil {
		return nil, fmt.Errorf("failed to query trade: %w", err)
	}

	if resp.IsFailure() {
		return nil, fmt.Errorf("alipay error: %s - %s", resp.Code, resp.Msg)
	}

	return &AlipayQueryResult{
		TradeNo:     resp.TradeNo,
		OutTradeNo:  resp.OutTradeNo,
		TradeStatus: string(resp.TradeStatus),
		TotalAmount: resp.TotalAmount,
	}, nil
}

func (s *AlipayService) CancelOrder(outTradeNo string) error {
	client, err := s.getClient()
	if err != nil {
		return errors.New("alipay not enabled")
	}

	var p = alipay.TradeClose{}
	p.OutTradeNo = outTradeNo

	resp, err := client.TradeClose(context.Background(), p)
	if err != nil {
		return fmt.Errorf("failed to close trade: %w", err)
	}

	if resp.IsFailure() {
		return fmt.Errorf("alipay error: %s - %s", resp.Code, resp.Msg)
	}

	return nil
}

func (s *AlipayService) HandleNotify(params map[string]string) (*NotifyResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, errors.New("alipay not enabled")
	}

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	notifyData, err := client.DecodeNotification(context.Background(), values)
	if err != nil {
		return nil, fmt.Errorf("failed to decode notification: %w", err)
	}

	result := &NotifyResult{
		OutTradeNo:  notifyData.OutTradeNo,
		TradeNo:     notifyData.TradeNo,
		TradeStatus: string(notifyData.TradeStatus),
		TotalAmount: notifyData.TotalAmount,
	}

	if notifyData.TradeStatus == alipay.TradeStatusSuccess || notifyData.TradeStatus == alipay.TradeStatusFinished {
		result.Success = true
	}

	return result, nil
}

func (s *AlipayService) ACKNotification(w http.ResponseWriter) {
	client, err := s.getClient()
	if err != nil {
		return
	}
	client.ACKNotification(w)
}

type AlipayQueryResult struct {
	TradeNo     string
	OutTradeNo  string
	TradeStatus string
	TotalAmount string
}

type NotifyResult struct {
	OutTradeNo  string
	TradeNo     string
	TradeStatus string
	TotalAmount string
	Success     bool
}

func (s *AlipayService) GetNotifyURL() string {
	return s.notifyURL
}

func (s *AlipayService) ReloadClient() {
	s.reloadClient()
}
