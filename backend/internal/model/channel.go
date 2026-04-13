package model

import (
	"encoding/json"
	"time"
)

// Channel represents an API channel
type Channel struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	TenantID uint   `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:100;not null"`
	Type     string `json:"type" gorm:"size:50;not null"` // openai|azure|claude|gemini|anthropic|custom
	BaseURL  string `json:"base_url" gorm:"size:500;not null"`

	// Auth
	APIKeyEncrypted string `json:"-" gorm:"size:500;not null"` // AES-256-GCM encrypted
	KeyVersion      int    `json:"key_version" gorm:"default:1"`

	// Models
	Models       string `json:"models" gorm:"type:jsonb"`        // JSON array ["gpt-4", "gpt-3.5-turbo"]
	ModelMapping string `json:"model_mapping" gorm:"type:jsonb"` // JSON object {"gpt-4": "gpt-4-0613"}

	// Load balancing
	Weight   int `json:"weight" gorm:"default:100"` // 1-1000
	Priority int `json:"priority" gorm:"default:0"` // Higher = more priority

	// Rate limits
	RPMLimit int `json:"rpm_limit" gorm:"default:1000"`
	TPMLimit int `json:"tpm_limit" gorm:"default:100000"`

	// Cost
	CostFactor       float64 `json:"cost_factor" gorm:"default:1.0"`
	PricePer1KInput  float64 `json:"price_per_1k_input" gorm:"default:0.01"`
	PricePer1KOutput float64 `json:"price_per_1k_output" gorm:"default:0.03"`

	// Group
	GroupName string `json:"group_name" gorm:"size:50;default:'default'`

	ProxyEnabled bool   `json:"proxy_enabled" gorm:"default:false"`
	ProxyType    string `json:"proxy_type" gorm:"size:20;default:'none'"`
	ProxyURL     string `json:"proxy_url" gorm:"size:500"`

	// Status
	Status          int        `json:"status" gorm:"default:1"` // 0:disabled, 1:enabled, 2:maintenance
	IsHealthy       bool       `json:"is_healthy" gorm:"default:true"`
	FailureCount    int        `json:"failure_count" gorm:"default:0"`
	LastSuccessAt   *time.Time `json:"last_success_at"`
	LastCheckAt     *time.Time `json:"last_check_at"`
	LastError       string     `json:"last_error" gorm:"type:text"`
	ResponseTimeAvg int        `json:"response_time_avg" gorm:"default:0"` // ms

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uint      `json:"created_by"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Relations
	Abilities []Ability `json:"abilities" gorm:"foreignKey:ChannelID"`
}

func (Channel) TableName() string {
	return "channels"
}

// GetModels returns the models as a slice
func (c *Channel) GetModels() []string {
	var models []string
	json.Unmarshal([]byte(c.Models), &models)
	return models
}

// SetModels sets the models from a slice
func (c *Channel) SetModels(models []string) {
	b, _ := json.Marshal(models)
	c.Models = string(b)
}

// GetModelMapping returns the model mapping as a map
func (c *Channel) GetModelMapping() map[string]string {
	var mapping map[string]string
	json.Unmarshal([]byte(c.ModelMapping), &mapping)
	return mapping
}

// SetModelMapping sets the model mapping from a map
func (c *Channel) SetModelMapping(mapping map[string]string) {
	b, _ := json.Marshal(mapping)
	c.ModelMapping = string(b)
}

// Ability represents a channel's capability
type Ability struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ChannelID   uint      `json:"channel_id" gorm:"not null;index"`
	AbilityType string    `json:"ability_type" gorm:"size:50;not null"` // chat|completion|embedding|moderation|audio|vision
	Model       string    `json:"model" gorm:"size:100;not null"`
	ModelAlias  string    `json:"model_alias" gorm:"size:100"`
	Config      string    `json:"config" gorm:"type:jsonb"` // Ability-specific config
	IsEnabled   bool      `json:"is_enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Ability) TableName() string {
	return "abilities"
}

// ChannelTestHistory represents channel test results
type ChannelTestHistory struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	TenantID  uint   `json:"tenant_id" gorm:"index"`
	ChannelID uint   `json:"channel_id" gorm:"not null;index"`
	UserID    uint   `json:"user_id" gorm:"index"`
	TestType  string `json:"test_type" gorm:"size:20;not null"` // models|chat|embeddings
	Model     string `json:"model" gorm:"size:100"`

	// Request
	RequestBody string `json:"request_body" gorm:"type:text"`

	// Response
	StatusCode     int    `json:"status_code"`
	ResponseBody   string `json:"response_body" gorm:"type:text"`
	ResponseTimeMs int    `json:"response_time_ms" gorm:"default:0"`

	// Result
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	ErrorType    string `json:"error_type" gorm:"size:50"`

	// Environment
	RequestIP string `json:"request_ip" gorm:"size:50"`
	UserAgent string `json:"user_agent" gorm:"size:500"`

	// Timestamp
	CreatedAt time.Time `json:"created_at"`
}

func (ChannelTestHistory) TableName() string {
	return "channel_test_history"
}
