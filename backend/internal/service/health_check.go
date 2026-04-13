package service

import (
	"context"
	"sync"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/adapter"
	"gapi-platform/internal/pkg/crypto"
	"gapi-platform/internal/pkg/logger"
	"gapi-platform/internal/repository"
)

const (
	FailureThreshold     = 3
	CheckIntervalMinutes = 5
	DeadRetryHours       = 1
	RequestTimeout       = 30
)

type HealthCheckService struct {
	channelRepo  *repository.ChannelRepository
	channelCache map[uint]*CachedChannel
	mu           sync.RWMutex
	stopChan     chan struct{}
	isRunning    bool
}

type CachedChannel struct {
	Channel   *model.Channel
	LastCheck time.Time
	IsHealthy bool
	Failures  int
}

func NewHealthCheckService(channelRepo *repository.ChannelRepository) *HealthCheckService {
	return &HealthCheckService{
		channelRepo:  channelRepo,
		channelCache: make(map[uint]*CachedChannel),
		stopChan:     make(chan struct{}),
	}
}

func (s *HealthCheckService) Start() {
	if s.isRunning {
		return
	}
	s.isRunning = true
	go s.run()
	logger.Info("Health check service started")
}

func (s *HealthCheckService) Stop() {
	if !s.isRunning {
		return
	}
	close(s.stopChan)
	s.isRunning = false
	logger.Info("Health check service stopped")
}

func (s *HealthCheckService) run() {
	ticker := time.NewTicker(time.Duration(CheckIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	s.checkAllChannels()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAllChannels()
		}
	}
}

func (s *HealthCheckService) checkAllChannels() {
	channels, err := s.channelRepo.GetActiveChannels()
	if err != nil {
		logger.Errorf("Failed to get channels for health check: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func(channel model.Channel) {
			defer wg.Done()
			s.checkChannel(channel.ID)
		}(ch)
	}
	wg.Wait()
}

func (s *HealthCheckService) checkChannel(channelID uint) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		logger.Errorf("Failed to get channel %d: %v", channelID, err)
		return
	}

	apiKey, err := crypto.Decrypt(channel.APIKeyEncrypted)
	if err != nil {
		apiKey = channel.APIKeyEncrypted
	}

	chatAdapter, err := adapter.GetAdapter(channel.Type)
	if err != nil {
		logger.Errorf("No adapter for channel %d type %s: %v", channelID, channel.Type, err)
		s.markUnhealthy(channel, "unsupported channel type: "+channel.Type)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(RequestTimeout)*time.Second)
	defer cancel()

	result := s.testChannel(ctx, chatAdapter, channel.BaseURL, apiKey)

	if result.Success {
		s.markHealthy(channel, result.ResponseTimeMs)
	} else {
		s.markFailed(channel, result.Error)
	}
}

type TestResult struct {
	Success        bool
	ResponseTimeMs int64
	Error          string
}

func (s *HealthCheckService) testChannel(ctx context.Context, chatAdapter adapter.Adapter, baseURL, apiKey string) *TestResult {
	start := time.Now()

	channel := &adapter.Channel{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}

	modelsResp, err := chatAdapter.ListModels(ctx, channel)
	if err != nil {
		return &TestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	if modelsResp == nil {
		return &TestResult{
			Success: false,
			Error:   "empty response",
		}
	}

	return &TestResult{
		Success:        true,
		ResponseTimeMs: time.Since(start).Milliseconds(),
	}
}

func (s *HealthCheckService) markHealthy(channel *model.Channel, responseTimeMs int64) {
	err := s.channelRepo.ResetFailureCount(channel.ID)
	if err != nil {
		logger.Errorf("Failed to reset failure count for channel %d: %v", channel.ID, err)
		return
	}

	if responseTimeMs > 0 {
		err = s.channelRepo.UpdateResponseTime(channel.ID, int(responseTimeMs))
		if err != nil {
			logger.Errorf("Failed to update response time for channel %d: %v", channel.ID, err)
		}
	}

	s.mu.Lock()
	s.channelCache[channel.ID] = &CachedChannel{
		Channel:   channel,
		LastCheck: time.Now(),
		IsHealthy: true,
		Failures:  0,
	}
	s.mu.Unlock()

	logger.Info("Channel marked healthy",
		"channel_id", channel.ID,
		"name", channel.Name,
		"response_time_ms", responseTimeMs)
}

func (s *HealthCheckService) markFailed(channel *model.Channel, errorMsg string) {
	err := s.channelRepo.IncrementFailureCount(channel.ID)
	if err != nil {
		logger.Errorf("Failed to increment failure count for channel %d: %v", channel.ID, err)
		return
	}

	s.mu.Lock()
	cached, exists := s.channelCache[channel.ID]
	if exists {
		cached.Failures++
	} else {
		s.channelCache[channel.ID] = &CachedChannel{
			Channel:   channel,
			LastCheck: time.Now(),
			IsHealthy: false,
			Failures:  1,
		}
		cached = s.channelCache[channel.ID]
	}
	s.mu.Unlock()

	if cached.Failures >= FailureThreshold {
		s.markUnhealthy(channel, errorMsg)
	} else {
		logger.Warnf("Channel %d (%s) failed check (attempt %d/%d): %s",
			channel.ID, channel.Name, cached.Failures, FailureThreshold, errorMsg)
	}
}

func (s *HealthCheckService) markUnhealthy(channel *model.Channel, reason string) {
	err := s.channelRepo.UpdateHealthStatus(channel.ID, false, FailureThreshold, reason)
	if err != nil {
		logger.Errorf("Failed to mark channel %d unhealthy: %v", channel.ID, err)
		return
	}

	s.mu.Lock()
	s.channelCache[channel.ID] = &CachedChannel{
		Channel:   channel,
		LastCheck: time.Now(),
		IsHealthy: false,
		Failures:  FailureThreshold,
	}
	s.mu.Unlock()

	logger.Warnf("Channel %d (%s) marked unhealthy: %s", channel.ID, channel.Name, reason)
}

func (s *HealthCheckService) CheckChannelManually(channelID uint) *TestResult {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return &TestResult{Success: false, Error: err.Error()}
	}

	apiKey, err := crypto.Decrypt(channel.APIKeyEncrypted)
	if err != nil {
		apiKey = channel.APIKeyEncrypted
	}

	chatAdapter, err := adapter.GetAdapter(channel.Type)
	if err != nil {
		return &TestResult{Success: false, Error: err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(RequestTimeout)*time.Second)
	defer cancel()

	return s.testChannel(ctx, chatAdapter, channel.BaseURL, apiKey)
}

func (s *HealthCheckService) GetChannelStatus(channelID uint) (bool, int, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cached, exists := s.channelCache[channelID]
	if !exists {
		return true, 0, time.Time{}
	}

	return cached.IsHealthy, cached.Failures, cached.LastCheck
}

func (s *HealthCheckService) GetStats() map[string]interface{} {
	channels, err := s.channelRepo.GetActiveChannels()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	healthy := 0
	unhealthy := 0
	unknown := 0

	s.mu.RLock()
	for _, ch := range channels {
		if cached, exists := s.channelCache[ch.ID]; exists {
			if cached.IsHealthy {
				healthy++
			} else {
				unhealthy++
			}
		} else {
			unknown++
		}
	}
	s.mu.RUnlock()

	return map[string]interface{}{
		"total":             len(channels),
		"healthy":           healthy,
		"unhealthy":         unhealthy,
		"unknown":           unknown,
		"failure_threshold": FailureThreshold,
		"check_interval":    CheckIntervalMinutes,
	}
}
