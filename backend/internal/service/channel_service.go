package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
)

type ChannelService struct {
	repo         *repository.ChannelRepository
	mu           sync.RWMutex
	channelCache map[uint]*model.Channel
}

func NewChannelService(repo *repository.ChannelRepository) *ChannelService {
	return &ChannelService{
		repo:         repo,
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

func (s *ChannelService) GetActiveChannelsByIDs(ids []uint) ([]model.Channel, error) {
	return s.repo.GetActiveChannelsByIDs(ids)
}

func (s *ChannelService) GetByModel(modelName string) ([]model.Channel, error) {
	return s.repo.GetByModel(modelName)
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
