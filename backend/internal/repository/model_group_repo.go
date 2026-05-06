package repository

import (
	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

type ModelGroupRepository struct {
	db *gorm.DB
}

func NewModelGroupRepository(db *gorm.DB) *ModelGroupRepository {
	return &ModelGroupRepository{db: db}
}

func (r *ModelGroupRepository) Create(g *model.ModelGroup) error {
	return r.db.Create(g).Error
}

func (r *ModelGroupRepository) GetByID(id uint) (*model.ModelGroup, error) {
	var g model.ModelGroup
	err := r.db.First(&g, id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *ModelGroupRepository) GetByName(name string) (*model.ModelGroup, error) {
	var g model.ModelGroup
	err := r.db.Where("name = ? AND deleted_at IS NULL", name).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *ModelGroupRepository) Update(g *model.ModelGroup) error {
	return r.db.Save(g).Error
}

func (r *ModelGroupRepository) Delete(id uint) error {
	return r.db.Delete(&model.ModelGroup{}, id).Error
}

func (r *ModelGroupRepository) List(page, pageSize int, status string) ([]model.ModelGroup, int64, error) {
	var groups []model.ModelGroup
	var total int64

	query := r.db.Model(&model.ModelGroup{}).Where("deleted_at IS NULL")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("sort_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&groups).Error
	return groups, total, err
}

func (r *ModelGroupRepository) ListAll() ([]model.ModelGroup, error) {
	var groups []model.ModelGroup
	err := r.db.Where("deleted_at IS NULL").Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}

func (r *ModelGroupRepository) GetOrCreateDefault() (*model.ModelGroup, error) {
	var g model.ModelGroup
	err := r.db.Where("name = ? AND deleted_at IS NULL", "default").First(&g).Error
	if err == nil {
		return &g, nil
	}
	g = model.ModelGroup{
		Name:        "default",
		DisplayName: "默认组",
		Description: "所有用户默认可访问的基础模型",
		SortOrder:   0,
		Status:      "active",
	}
	if err := r.db.Create(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

type ChannelGroupRelationRepository struct {
	db *gorm.DB
}

func NewChannelGroupRelationRepository(db *gorm.DB) *ChannelGroupRelationRepository {
	return &ChannelGroupRelationRepository{db: db}
}

func (r *ChannelGroupRelationRepository) Create(rel *model.ChannelGroupRelation) error {
	return r.db.Create(rel).Error
}

func (r *ChannelGroupRelationRepository) DeleteByChannelAndGroup(channelID, groupID uint) error {
	return r.db.Where("channel_id = ? AND group_id = ?", channelID, groupID).Delete(&model.ChannelGroupRelation{}).Error
}

func (r *ChannelGroupRelationRepository) DeleteByGroup(groupID uint) error {
	return r.db.Where("group_id = ?", groupID).Delete(&model.ChannelGroupRelation{}).Error
}

func (r *ChannelGroupRelationRepository) DeleteByChannel(channelID uint) error {
	return r.db.Where("channel_id = ?", channelID).Delete(&model.ChannelGroupRelation{}).Error
}

func (r *ChannelGroupRelationRepository) ListByGroupID(groupID uint) ([]model.ChannelGroupRelation, error) {
	var rels []model.ChannelGroupRelation
	err := r.db.Where("group_id = ?", groupID).Order("priority DESC").Find(&rels).Error
	return rels, err
}

func (r *ChannelGroupRelationRepository) ListByChannelID(channelID uint) ([]model.ChannelGroupRelation, error) {
	var rels []model.ChannelGroupRelation
	err := r.db.Where("channel_id = ?", channelID).Find(&rels).Error
	return rels, err
}

func (r *ChannelGroupRelationRepository) GetChannelIDsByGroupIDs(groupIDs []uint) ([]uint, error) {
	var rels []model.ChannelGroupRelation
	err := r.db.Select("channel_id").Where("group_id IN ? AND is_enabled = ?", groupIDs, true).Find(&rels).Error
	if err != nil {
		return nil, err
	}
	channelIDs := make([]uint, 0, len(rels))
	seen := make(map[uint]bool)
	for _, r := range rels {
		if !seen[r.ChannelID] {
			channelIDs = append(channelIDs, r.ChannelID)
			seen[r.ChannelID] = true
		}
	}
	return channelIDs, nil
}

type UserGroupRelationRepository struct {
	db *gorm.DB
}

func NewUserGroupRelationRepository(db *gorm.DB) *UserGroupRelationRepository {
	return &UserGroupRelationRepository{db: db}
}

func (r *UserGroupRelationRepository) Create(rel *model.UserGroupRelation) error {
	return r.db.Create(rel).Error
}

func (r *UserGroupRelationRepository) DeleteByUserAndGroup(userID, groupID uint) error {
	return r.db.Where("user_id = ? AND group_id = ?", userID, groupID).Delete(&model.UserGroupRelation{}).Error
}

func (r *UserGroupRelationRepository) DeleteByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserGroupRelation{}).Error
}

func (r *UserGroupRelationRepository) ListByUserID(userID uint) ([]model.UserGroupRelation, error) {
	var rels []model.UserGroupRelation
	err := r.db.Where("user_id = ?", userID).Find(&rels).Error
	return rels, err
}

func (r *UserGroupRelationRepository) GetGroupIDsByUserID(userID uint) ([]uint, error) {
	rels, err := r.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	groupIDs := make([]uint, 0, len(rels))
	for _, r := range rels {
		groupIDs = append(groupIDs, r.GroupID)
	}
	return groupIDs, nil
}

func (r *UserGroupRelationRepository) EnsureDefaultGroup(userID uint) error {
	existing, err := r.GetGroupIDsByUserID(userID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	defaultGroup, err := NewModelGroupRepository(r.db).GetOrCreateDefault()
	if err != nil {
		return err
	}
	return r.db.Create(&model.UserGroupRelation{
		UserID:  userID,
		GroupID: defaultGroup.ID,
	}).Error
}

type ModelPricingRepository struct {
	db *gorm.DB
}

func NewModelPricingRepository(db *gorm.DB) *ModelPricingRepository {
	return &ModelPricingRepository{db: db}
}

func (r *ModelPricingRepository) Create(p *model.ModelPricing) error {
	return r.db.Create(p).Error
}

func (r *ModelPricingRepository) GetByID(id uint) (*model.ModelPricing, error) {
	var p model.ModelPricing
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ModelPricingRepository) GetByModel(modelName string) (*model.ModelPricing, error) {
	var p model.ModelPricing
	err := r.db.Where("model = ? AND deleted_at IS NULL", modelName).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ModelPricingRepository) Update(p *model.ModelPricing) error {
	return r.db.Save(p).Error
}

func (r *ModelPricingRepository) Delete(id uint) error {
	return r.db.Delete(&model.ModelPricing{}, id).Error
}

func (r *ModelPricingRepository) List(page, pageSize int, provider string, isEnabled *bool) ([]model.ModelPricing, int64, error) {
	var pricings []model.ModelPricing
	var total int64

	query := r.db.Model(&model.ModelPricing{}).Where("deleted_at IS NULL")
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if isEnabled != nil {
		query = query.Where("is_enabled = ?", *isEnabled)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("sort_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pricings).Error
	return pricings, total, err
}

func (r *ModelPricingRepository) ListAll() ([]model.ModelPricing, error) {
	var pricings []model.ModelPricing
	err := r.db.Where("deleted_at IS NULL AND is_enabled = true").Order("sort_order ASC, id ASC").Find(&pricings).Error
	return pricings, err
}
