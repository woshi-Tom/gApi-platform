package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/pkg/logger"
)

const turnstileSiteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

var (
	ErrTurnstileTokenRequired      = errors.New("turnstile token required")
	ErrTurnstileNotConfigured     = errors.New("turnstile not configured")
	ErrTurnstileVerificationFailed = errors.New("turnstile verification failed")
)

type TurnstileService struct {
	cfg        config.TurnstileConfig
	httpClient *http.Client
	timeout    time.Duration
}

type turnstileSiteverifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func NewTurnstileService(cfg config.TurnstileConfig) *TurnstileService {
	timeout := 5 * time.Second
	return &TurnstileService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

func (s *TurnstileService) Enabled() bool {
	return s != nil && s.cfg.Enabled
}

func (s *TurnstileService) Verify(ctx context.Context, token, remoteIP string) error {
	if !s.Enabled() {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTurnstileTokenRequired
	}

	secret := strings.TrimSpace(s.cfg.SecretKey)
	if secret == "" {
		return ErrTurnstileNotConfigured
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, turnstileSiteverifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%w: create request: %v", ErrTurnstileVerificationFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Warnf("Turnstile siteverify request failed: %v", err)
		return fmt.Errorf("%w: request failed", ErrTurnstileVerificationFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logger.Warnf("Turnstile siteverify returned status %d", resp.StatusCode)
		return fmt.Errorf("%w: upstream status", ErrTurnstileVerificationFailed)
	}

	var result turnstileSiteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Warnf("Turnstile siteverify response decode failed: %v", err)
		return fmt.Errorf("%w: decode response", ErrTurnstileVerificationFailed)
	}

	if !result.Success {
		logger.Warnf("Turnstile verification failed, error codes: %v", result.ErrorCodes)
		return ErrTurnstileVerificationFailed
	}

	return nil
}
