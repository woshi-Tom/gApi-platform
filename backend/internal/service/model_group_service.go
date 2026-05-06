package service

import (
	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
)

type ModelGroupService struct {
	groupRepo   *repository.ModelGroupRepository
	channelRel  *repository.ChannelGroupRelationRepository
	userRel     *repository.UserGroupRelationRepository
	pricingRepo *repository.ModelPricingRepository
}

func NewModelGroupService(
	groupRepo *repository.ModelGroupRepository,
	channelRel *repository.ChannelGroupRelationRepository,
	userRel *repository.UserGroupRelationRepository,
	pricingRepo *repository.ModelPricingRepository,
) *ModelGroupService {
	return &ModelGroupService{
		groupRepo:   groupRepo,
		channelRel:  channelRel,
		userRel:     userRel,
		pricingRepo: pricingRepo,
	}
}

func (s *ModelGroupService) CreateGroup(g *model.ModelGroup) error {
	return s.groupRepo.Create(g)
}

func (s *ModelGroupService) UpdateGroup(g *model.ModelGroup) error {
	return s.groupRepo.Update(g)
}

func (s *ModelGroupService) DeleteGroup(id uint) error {
	return s.groupRepo.Delete(id)
}

func (s *ModelGroupService) GetGroup(id uint) (*model.ModelGroup, error) {
	return s.groupRepo.GetByID(id)
}

func (s *ModelGroupService) ListGroups(page, pageSize int, status string) ([]model.ModelGroup, int64, error) {
	return s.groupRepo.List(page, pageSize, status)
}

func (s *ModelGroupService) ListAllGroups() ([]model.ModelGroup, error) {
	return s.groupRepo.ListAll()
}

func (s *ModelGroupService) AddChannelToGroup(channelID, groupID uint, priority int) error {
	return s.channelRel.Create(&model.ChannelGroupRelation{
		ChannelID: channelID,
		GroupID:   groupID,
		IsEnabled: true,
		Priority:  priority,
	})
}

func (s *ModelGroupService) RemoveChannelFromGroup(channelID, groupID uint) error {
	return s.channelRel.DeleteByChannelAndGroup(channelID, groupID)
}

func (s *ModelGroupService) GetChannelsByGroup(groupID uint) ([]model.ChannelGroupRelation, error) {
	return s.channelRel.ListByGroupID(groupID)
}

func (s *ModelGroupService) SetUserGroups(userID uint, groupIDs []uint) error {
	if err := s.userRel.DeleteByUser(userID); err != nil {
		return err
	}
	for _, gid := range groupIDs {
		if err := s.userRel.Create(&model.UserGroupRelation{UserID: userID, GroupID: gid}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ModelGroupService) GetUserGroups(userID uint) ([]model.UserGroupRelation, error) {
	return s.userRel.ListByUserID(userID)
}

func (s *ModelGroupService) GetUserGroupIDs(userID uint) ([]uint, error) {
	return s.userRel.GetGroupIDsByUserID(userID)
}

func (s *ModelGroupService) CreatePricing(p *model.ModelPricing) error {
	return s.pricingRepo.Create(p)
}

func (s *ModelGroupService) UpdatePricing(p *model.ModelPricing) error {
	return s.pricingRepo.Update(p)
}

func (s *ModelGroupService) DeletePricing(id uint) error {
	return s.pricingRepo.Delete(id)
}

func (s *ModelGroupService) GetPricing(id uint) (*model.ModelPricing, error) {
	return s.pricingRepo.GetByID(id)
}

func (s *ModelGroupService) GetPricingByModel(modelName string) (*model.ModelPricing, error) {
	return s.pricingRepo.GetByModel(modelName)
}

func (s *ModelGroupService) ListPricing(page, pageSize int, provider string, isEnabled *bool) ([]model.ModelPricing, int64, error) {
	return s.pricingRepo.List(page, pageSize, provider, isEnabled)
}

func (s *ModelGroupService) ListAllPricing() ([]model.ModelPricing, error) {
	return s.pricingRepo.ListAll()
}

func (s *ModelGroupService) GetAvailableChannelIDsForUser(userID uint) ([]uint, error) {
	groupIDs, err := s.userRel.GetGroupIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(groupIDs) == 0 {
		defaultGroup, err := s.groupRepo.GetOrCreateDefault()
		if err != nil {
			return nil, err
		}
		groupIDs = []uint{defaultGroup.ID}
	}
	return s.channelRel.GetChannelIDsByGroupIDs(groupIDs)
}
