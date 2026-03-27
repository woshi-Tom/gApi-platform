package service

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
)

type ChannelService struct {
	repo     *repository.ChannelRepository
	crypto   CryptoService
	mu       sync.RWMutex
	channelCache map[uint]*model.Channel
}

type CryptoService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type NoOpCrypto struct{}

func (c *NoOpCrypto) Encrypt(plaintext string) (string, error) {
	return plaintext, nil
}

func (c *NoOpCrypto) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}

func NewChannelService(repo *repository.ChannelRepository) *ChannelService {
	return &ChannelService{
		repo:     repo,
		crypto:   &NoOpCrypto{},
		channelCache: make(map[uint]*model.Channel),
	}
}

func NewChannelServiceWithCrypto(repo *repository.ChannelRepository, crypto CryptoService) *ChannelService {
	return &ChannelService{
		repo:     repo,
		crypto:   crypto,
		channelCache: make(map[uint]*model.Channel),
	}
}

func (s *ChannelService) Create(channel *model.Channel) error {
	if channel.Models == "" {
		channel.Models = "[]"
	}
	if channel.ModelMapping == "" {
		channel.ModelMapping = "{}"
	}
	return s.repo.Create(channel)
}

func (s *ChannelService) GetByID(id uint) (*model.Channel, error) {
	return s.repo.GetByID(id)
}

func (s *ChannelService) Update(channel *model.Channel) error {
	return s.repo.Update(channel)
}

func (s *ChannelService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ChannelService) List(page, pageSize int, channelType, status, group, keyword string) ([]model.Channel, int64, error) {
	return s.repo.List(page, pageSize, channelType, status, group, keyword)
}

func (s *ChannelService) GetActiveChannels() ([]model.Channel, error) {
	return s.repo.GetActiveChannels()
}

func (s *ChannelService) GetByModel(modelName string) ([]model.Channel, error) {
	return s.repo.GetByModel(modelName)
}

func (s *ChannelService) EncryptAPIKey(channel *model.Channel, apiKey string) error {
	encrypted, err := s.crypto.Encrypt(apiKey)
	if err != nil {
		return err
	}
	channel.APIKeyEncrypted = encrypted
	return nil
}

func (s *ChannelService) DecryptAPIKey(channel *model.Channel) (string, error) {
	return s.crypto.Decrypt(channel.APIKeyEncrypted)
}

func (s *ChannelService) SelectChannel(modelName string) (*model.Channel, error) {
	channels, err := s.GetByModel(modelName)
	if err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		channels, err = s.GetActiveChannels()
		if err != nil {
			return nil, err
		}
	}
	if len(channels) == 0 {
		return nil, errors.New("no available channel")
	}
	return s.weightedSelect(channels), nil
}

func (s *ChannelService) weightedSelect(channels []model.Channel) *model.Channel {
	if len(channels) == 0 {
		return nil
	}
	if len(channels) == 1 {
		return &channels[0]
	}

	totalWeight := 0
	for _, ch := range channels {
		totalWeight += ch.Weight
	}

	randNum, _ := rand.Int(rand.Reader, big.NewInt(int64(totalWeight)))
	runningWeight := 0
	for i, ch := range channels {
		runningWeight += ch.Weight
		if randNum.Int64() < int64(runningWeight) {
			return &channels[i]
		}
	}
	return &channels[0]
}

func (s *ChannelService) UpdateHealthStatus(id uint, isHealthy bool, failureCount int, lastError string) error {
	return s.repo.UpdateHealthStatus(id, isHealthy, failureCount, lastError)
}

func (s *ChannelService) IncrementFailureCount(id uint) error {
	return s.repo.IncrementFailureCount(id)
}

func (s *ChannelService) ResetFailureCount(id uint) error {
	return s.repo.ResetFailureCount(id)
}

func (s *ChannelService) UpdateResponseTime(id uint, responseTimeMs int) error {
	return s.repo.UpdateResponseTime(id, responseTimeMs)
}

func (s *ChannelService) GetStats() (map[string]interface{}, error) {
	counts, err := s.repo.CountByStatus()
	if err != nil {
		return nil, err
	}

	total, _, err := s.repo.List(1, 1, "", "", "", "")
	if err != nil {
		return nil, err
	}

	active, _ := s.repo.GetActiveChannels()

	return map[string]interface{}{
		"total_channels":   len(total),
		"healthy_channels": len(active),
		"status_counts":    counts,
	}, nil
}

type ChannelTestResult struct {
	Success       bool
	StatusCode   int
	ResponseTimeMs int64
	Models        []string
	Content       string
	Error         string
}

func (s *ChannelService) TestChannel(channel *model.Channel, testType, model string) (*ChannelTestResult, error) {
	result := &ChannelTestResult{}

	apiKey, err := s.DecryptAPIKey(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	switch testType {
	case "models":
		models, statusCode, err := s.testModels(channel.BaseURL, apiKey)
		result.Models = models
		result.StatusCode = statusCode
		result.Success = err == nil
		if err != nil {
			result.Error = err.Error()
		}
	case "chat":
		content, statusCode, err := s.testChat(channel.BaseURL, apiKey, model)
		result.Content = content
		result.StatusCode = statusCode
		result.Success = err == nil
		if err != nil {
			result.Error = err.Error()
		}
	default:
		return nil, fmt.Errorf("unsupported test type: %s", testType)
	}

	return result, nil
}

func (s *ChannelService) testModels(baseURL, apiKey string) ([]string, int, error) {
	client := &HTTPClient{Timeout: 10}
	resp, err := client.Get(baseURL+"/v1/models", map[string]string{
		"Authorization": "Bearer " + apiKey,
	})
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, err
	}

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, resp.StatusCode, nil
}

func (s *ChannelService) testChat(baseURL, apiKey, modelName string) (string, int, error) {
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}
	payload := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 10,
	}

	client := &HTTPClient{Timeout: 30}
	resp, err := client.Post(baseURL+"/v1/chat/completions", payload, map[string]string{
		"Authorization": "Bearer " + apiKey,
	})
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", resp.StatusCode, err
	}

	content := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
	}
	return content, resp.StatusCode, nil
}

type HTTPClient struct {
	Timeout int
}

func (c *HTTPClient) Get(url string, headers map[string]string) (*HTTPResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	if c.Timeout > 0 {
		client.Timeout = time.Duration(c.Timeout) * time.Second
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{Response: resp}, nil
}

func (c *HTTPClient) Post(url string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	if c.Timeout > 0 {
		client.Timeout = time.Duration(c.Timeout) * time.Second
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{Response: resp}, nil
}

type HTTPResponse struct {
	*http.Response
}
