package model

import (
	"encoding/json"
	"time"
)

// ModelGroup represents a group of models for access control
type ModelGroup struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	TenantID    uint       `json:"tenant_id" gorm:"index"`
Name        string    `json:"name" gorm:"size:50;not null;index"`     // default, pro, vip
	DisplayName string     `json:"display_name" gorm:"size:100;not null"`    // 显示名称
	Description string     `json:"description" gorm:"type:text"`             // 分组描述
	SortOrder   int        `json:"sort_order" gorm:"default:0"`
	Status      string     `json:"status" gorm:"size:20;default:'active'"` // active|disabled
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedBy   *uint      `json:"created_by"`
	DeletedAt   *time.Time `json:"deleted_at" gorm:"index"`
}

func (ModelGroup) TableName() string {
	return "model_groups"
}

// ChannelGroupRelation represents the many-to-many relation between channels and model groups
type ChannelGroupRelation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TenantID  uint      `json:"tenant_id" gorm:"index"`
	ChannelID uint      `json:"channel_id" gorm:"not null;index"`
	GroupID   uint      `json:"group_id" gorm:"not null;index"`
	IsEnabled bool      `json:"is_enabled" gorm:"default:true"`
	Priority  int       `json:"priority" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ChannelGroupRelation) TableName() string {
	return "channel_group_relations"
}

// UserGroupRelation represents the relation between users and model groups
type UserGroupRelation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TenantID  uint      `json:"tenant_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	GroupID   uint      `json:"group_id" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *uint     `json:"created_by"`
}

func (UserGroupRelation) TableName() string {
	return "user_group_relations"
}

// ModelPricing represents pricing and metadata for a model
type ModelPricing struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	TenantID      uint       `json:"tenant_id" gorm:"index"`
	Model         string     `json:"model" gorm:"size:100;not null"`
	Provider      string     `json:"provider" gorm:"size:50"`
	DisplayName   string     `json:"display_name" gorm:"size:100"`
	PriceInput    float64    `json:"price_input" gorm:"default:0.03"`  // per 1K input tokens
	PriceOutput   float64    `json:"price_output" gorm:"default:0.06"` // per 1K output tokens
	AbilityTypes  string     `json:"ability_types" gorm:"type:jsonb"`  // ["chat","completion",...]
	ContextLength int        `json:"context_length"`                   // max context tokens
	MaxOutput     int        `json:"max_output"`                       // max output tokens
	GroupID       *uint      `json:"group_id" gorm:"index"`
	IsEnabled     bool       `json:"is_enabled" gorm:"default:true"`
	IsFeatured    bool       `json:"is_featured" gorm:"default:false"`
	SortOrder     int        `json:"sort_order" gorm:"default:0"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedBy     *uint      `json:"created_by"`
	DeletedAt     *time.Time `json:"deleted_at" gorm:"index"`
}

func (ModelPricing) TableName() string {
	return "model_pricing"
}

// GetAbilityTypes returns ability types as a slice
func (m *ModelPricing) GetAbilityTypes() []string {
	if m.AbilityTypes == "" || m.AbilityTypes == "null" {
		return []string{"chat"}
	}
	var types []string
	json.Unmarshal([]byte(m.AbilityTypes), &types)
	if len(types) == 0 {
		return []string{"chat"}
	}
	return types
}

// SetAbilityTypes sets ability types from a slice
func (m *ModelPricing) SetAbilityTypes(types []string) {
	b, _ := json.Marshal(types)
	m.AbilityTypes = string(b)
}
