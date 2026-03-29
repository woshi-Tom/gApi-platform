package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/crypto"
	"gorm.io/gorm"
)

// ConfigKey constants for email settings
const (
	ConfigKeySMTPEnabled        = "smtp_enabled"
	ConfigKeySMTPHost           = "smtp_host"
	ConfigKeySMTPPort           = "smtp_port"
	ConfigKeySMTPUseTLS         = "smtp_use_tls"
	ConfigKeySMTPUsername       = "smtp_username"
	ConfigKeySMTPPassword       = "smtp_password"
	ConfigKeySMTPFromName       = "smtp_from_name"
	ConfigKeySMTPFromEmail      = "smtp_from_email"
	ConfigGroupEmail            = "email"
	ConfigKeyAllowRegister      = "allow_register"
	ConfigKeyRequireEmailVerify = "require_email_verify"
	ConfigKeyEnableCaptcha      = "enable_captcha"
	ConfigKeyNewUserQuota       = "new_user_quota"
	ConfigKeyTrialVIPDays       = "trial_vip_days"
	ConfigGroupRegister         = "register"
)

// SMTPConfig represents SMTP settings
type SMTPConfig struct {
	Enabled   bool   `json:"enabled"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	UseTLS    bool   `json:"use_tls"`
	Username  string `json:"username"`
	Password  string `json:"-"` // Never expose password in JSON
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
}

// SettingsService handles system configuration
type SettingsService struct {
	db         *gorm.DB
	cache      map[string]*SMTPConfig
	cacheMutex sync.RWMutex
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// NewSettingsService creates a new settings service
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{
		db:       db,
		cache:    make(map[string]*SMTPConfig),
		cacheTTL: 5 * time.Minute,
	}
}

// GetSMTPConfig retrieves SMTP configuration
func (s *SettingsService) GetSMTPConfig() (*SMTPConfig, error) {
	s.cacheMutex.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.cache["smtp"] != nil {
		cfg := s.cache["smtp"]
		s.cacheMutex.RUnlock()
		return cfg, nil
	}
	s.cacheMutex.RUnlock()

	cfg := &SMTPConfig{
		Enabled:   false,
		Host:      "",
		Port:      587,
		UseTLS:    true,
		Username:  "",
		Password:  "",
		FromName:  "gAPI Platform",
		FromEmail: "noreply@gapi.com",
	}

	configs, err := s.getConfigsByGroup(ConfigGroupEmail)
	if err != nil {
		return cfg, err
	}

	for _, c := range configs {
		switch c.ConfigKey {
		case ConfigKeySMTPEnabled:
			cfg.Enabled = c.ConfigValue == "true" || c.ConfigValue == "1"
		case ConfigKeySMTPHost:
			cfg.Host = c.ConfigValue
		case ConfigKeySMTPPort:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.Port)
		case ConfigKeySMTPUseTLS:
			cfg.UseTLS = c.ConfigValue == "true" || c.ConfigValue == "1"
		case ConfigKeySMTPUsername:
			cfg.Username = c.ConfigValue
		case ConfigKeySMTPPassword:
			// Decrypt password
			if c.ConfigValue != "" {
				decrypted, err := crypto.Decrypt(c.ConfigValue)
				if err == nil {
					cfg.Password = decrypted
				}
			}
		case ConfigKeySMTPFromName:
			cfg.FromName = c.ConfigValue
		case ConfigKeySMTPFromEmail:
			cfg.FromEmail = c.ConfigValue
		}
	}

	s.cacheMutex.Lock()
	s.cache["smtp"] = cfg
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	return cfg, nil
}

// UpdateSMTPConfig updates SMTP configuration
func (s *SettingsService) UpdateSMTPConfig(cfg *SMTPConfig) error {
	updates := map[string]struct {
		Value       string
		IsSensitive bool
	}{
		ConfigKeySMTPEnabled:   {Value: boolToString(cfg.Enabled), IsSensitive: false},
		ConfigKeySMTPHost:      {Value: cfg.Host, IsSensitive: false},
		ConfigKeySMTPPort:      {Value: fmt.Sprintf("%d", cfg.Port), IsSensitive: false},
		ConfigKeySMTPUseTLS:    {Value: boolToString(cfg.UseTLS), IsSensitive: false},
		ConfigKeySMTPUsername:  {Value: cfg.Username, IsSensitive: false},
		ConfigKeySMTPPassword:  {Value: cfg.Password, IsSensitive: true}, // Will be encrypted
		ConfigKeySMTPFromName:  {Value: cfg.FromName, IsSensitive: false},
		ConfigKeySMTPFromEmail: {Value: cfg.FromEmail, IsSensitive: false},
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, data := range updates {
			value := data.Value

			// Encrypt password if sensitive and not empty
			if data.IsSensitive && value != "" {
				encrypted, err := crypto.Encrypt(value)
				if err != nil {
					return fmt.Errorf("failed to encrypt %s: %w", key, err)
				}
				value = encrypted
			}

			// Skip empty password updates (don't overwrite existing)
			if key == ConfigKeySMTPPassword && value == "" {
				continue
			}

			if err := s.upsertConfig(tx, key, value, "boolean", data.IsSensitive); err != nil {
				return err
			}
		}
		return nil
	})
}

// TestSMTPConnection tests SMTP connection by sending a test email
func (s *SettingsService) TestSMTPConnection(testEmail string) error {
	cfg, err := s.GetSMTPConfig()
	if err != nil {
		return fmt.Errorf("failed to get SMTP config: %w", err)
	}

	if !cfg.Enabled {
		return errors.New("SMTP is not enabled")
	}

	if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		return errors.New("SMTP configuration is incomplete")
	}

	// Create a temporary mailer with test config
	mailer := &testMailer{
		host:      cfg.Host,
		port:      cfg.Port,
		username:  cfg.Username,
		password:  cfg.Password,
		useTLS:    cfg.UseTLS,
		fromName:  cfg.FromName,
		fromEmail: cfg.FromEmail,
	}

	return mailer.sendTestEmail(testEmail)
}

func (s *SettingsService) getConfigsByGroup(group string) ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := s.db.Where("config_group = ?", group).Find(&configs).Error
	return configs, err
}

func (s *SettingsService) upsertConfig(tx *gorm.DB, key, value, valueType string, isSensitive bool) error {
	var config model.SystemConfig
	err := tx.Where("config_key = ?", key).First(&config).Error

	description := getConfigDescription(key)

	if err == gorm.ErrRecordNotFound {
		// Insert new
		config = model.SystemConfig{
			ConfigKey:   key,
			ConfigValue: value,
			ValueType:   valueType,
			ConfigGroup: ConfigGroupEmail,
			IsSensitive: isSensitive,
			Description: description,
		}
		return tx.Create(&config).Error
	} else if err != nil {
		return err
	}

	// Update existing
	config.ConfigValue = value
	if isSensitive {
		config.IsSensitive = true
	}
	return tx.Save(&config).Error
}

func (s *SettingsService) InvalidateCache() {
	s.cacheMutex.Lock()
	s.cache = make(map[string]*SMTPConfig)
	s.cacheTime = time.Time{}
	s.cacheMutex.Unlock()
}

// Helper functions
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getConfigDescription(key string) string {
	descriptions := map[string]string{
		ConfigKeySMTPEnabled:   "启用邮箱服务",
		ConfigKeySMTPHost:      "SMTP 服务器地址",
		ConfigKeySMTPPort:      "SMTP 端口",
		ConfigKeySMTPUseTLS:    "使用 TLS 加密",
		ConfigKeySMTPUsername:  "SMTP 用户名",
		ConfigKeySMTPPassword:  "SMTP 密码 (加密存储)",
		ConfigKeySMTPFromName:  "发件人名称",
		ConfigKeySMTPFromEmail: "发件人邮箱",
	}
	if desc, ok := descriptions[key]; ok {
		return desc
	}
	return ""
}

type RegisterSettings struct {
	AllowRegister      bool  `json:"allow_register"`
	RequireEmailVerify *bool `json:"require_email_verify,omitempty"`
	EnableCaptcha      bool  `json:"enable_captcha"`
	NewUserQuota       int   `json:"new_user_quota"`
	TrialVIPDays       int   `json:"trial_vip_days"`
}

func (s *SettingsService) GetRegisterSettings() (*RegisterSettings, error) {
	settings := &RegisterSettings{
		AllowRegister:      true,
		RequireEmailVerify: boolPtr(true),
		EnableCaptcha:      true,
		NewUserQuota:       100000,
		TrialVIPDays:       0,
	}

	configs, err := s.getConfigsByGroup(ConfigGroupRegister)
	if err != nil {
		return settings, err
	}

	for _, c := range configs {
		switch c.ConfigKey {
		case ConfigKeyAllowRegister:
			settings.AllowRegister = c.ConfigValue == "true" || c.ConfigValue == "1"
		case ConfigKeyRequireEmailVerify:
			val := c.ConfigValue == "true" || c.ConfigValue == "1"
			settings.RequireEmailVerify = &val
		case ConfigKeyEnableCaptcha:
			settings.EnableCaptcha = c.ConfigValue == "true" || c.ConfigValue == "1"
		case ConfigKeyNewUserQuota:
			fmt.Sscanf(c.ConfigValue, "%d", &settings.NewUserQuota)
		case ConfigKeyTrialVIPDays:
			fmt.Sscanf(c.ConfigValue, "%d", &settings.TrialVIPDays)
		}
	}

	return settings, nil
}

func (s *SettingsService) UpdateRegisterSettings(settings *RegisterSettings) error {
	configs := map[string]string{
		ConfigKeyAllowRegister: boolToString(settings.AllowRegister),
		ConfigKeyEnableCaptcha: boolToString(settings.EnableCaptcha),
		ConfigKeyNewUserQuota:  fmt.Sprintf("%d", settings.NewUserQuota),
		ConfigKeyTrialVIPDays:  fmt.Sprintf("%d", settings.TrialVIPDays),
	}

	if settings.RequireEmailVerify != nil {
		configs[ConfigKeyRequireEmailVerify] = boolToString(*settings.RequireEmailVerify)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			if err := s.upsertConfig(tx, key, value, "boolean", false); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SettingsService) IsSMTPEnabled() bool {
	cfg, err := s.GetSMTPConfig()
	if err != nil {
		return false
	}
	return cfg.Enabled && cfg.Host != "" && cfg.Username != "" && cfg.Password != ""
}

func boolPtr(b bool) *bool {
	return &b
}
