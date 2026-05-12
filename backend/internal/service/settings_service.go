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
	ConfigKeyAllowedDomains     = "allowed_domains"
	ConfigKeyMaxAccountsPerIP   = "max_accounts_per_ip"
	ConfigKeyMinPasswordLength  = "min_password_length"
	ConfigKeySignupRewardType   = "signup_reward_type"
	ConfigKeySignupRewardAmount = "signup_reward_amount"
	ConfigGroupRegister         = "register"
)

// ConfigKey constants for Alipay settings
const (
	ConfigKeyAlipayEnabled    = "alipay_enabled"
	ConfigKeyAlipayAppID      = "alipay_app_id"
	ConfigKeyAlipayPrivateKey = "alipay_private_key"
	ConfigKeyAlipayPublicKey  = "alipay_public_key"
	ConfigKeyAlipayEncryptKey = "alipay_encrypt_key"
	ConfigKeyAlipaySandbox    = "alipay_sandbox"
	ConfigGroupPayment        = "payment"
)

// ConfigKey constants for General settings
const (
	ConfigKeySiteName        = "site_name"
	ConfigKeySiteLogo        = "site_logo"
	ConfigKeySiteDescription = "site_description"
	ConfigGroupGeneral       = "general"
)

// ConfigKey constants for Rate Limit settings
const (
	ConfigKeyFreeRPM     = "free_rpm"
	ConfigKeyFreeTPM     = "free_tpm"
	ConfigKeyVIPRPM      = "vip_rpm"
	ConfigKeyVIPTPM      = "vip_tpm"
	ConfigGroupRateLimit = "rate_limit"
)

// ConfigKey constants for Security settings
const (
	ConfigKeyJWTSecret          = "jwt_secret"
	ConfigKeyJWTExpireHours     = "jwt_expire_hours"
	ConfigKeyPasswordMinLength  = "password_min_length"
	ConfigKeyPasswordExpireDays = "password_expire_days"
	ConfigGroupSecurity         = "security"
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
	db             *gorm.DB
	smtpCache      *SMTPConfig
	alipayCache    *AlipayConfig
	generalCache   *GeneralSettings
	rateLimitCache *RateLimitSettings
	securityCache  *SecuritySettings
	cacheMutex     sync.RWMutex
	cacheTime      time.Time
	cacheTTL       time.Duration
}

// NewSettingsService creates a new settings service
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{
		db:       db,
		cacheTTL: 5 * time.Minute,
	}
}

// GetSMTPConfig retrieves SMTP configuration
func (s *SettingsService) GetSMTPConfig() (*SMTPConfig, error) {
	s.cacheMutex.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.smtpCache != nil {
		cfg := s.smtpCache
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
	s.smtpCache = cfg
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
		ConfigKeySMTPPassword:  {Value: cfg.Password, IsSensitive: true},
		ConfigKeySMTPFromName:  {Value: cfg.FromName, IsSensitive: false},
		ConfigKeySMTPFromEmail: {Value: cfg.FromEmail, IsSensitive: false},
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, data := range updates {
			value := data.Value

			if data.IsSensitive && value != "" {
				encrypted, err := crypto.Encrypt(value)
				if err != nil {
					return fmt.Errorf("failed to encrypt %s: %w", key, err)
				}
				value = encrypted
			}

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

	config.ConfigValue = value
	if isSensitive {
		config.IsSensitive = true
	}
	return tx.Save(&config).Error
}

func (s *SettingsService) InvalidateCache() {
	s.cacheMutex.Lock()
	s.smtpCache = nil
	s.alipayCache = nil
	s.generalCache = nil
	s.rateLimitCache = nil
	s.securityCache = nil
	s.cacheTime = time.Time{}
	s.cacheMutex.Unlock()
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getConfigDescription(key string) string {
	descriptions := map[string]string{
		ConfigKeySMTPEnabled:        "启用邮箱服务",
		ConfigKeySMTPHost:           "SMTP 服务器地址",
		ConfigKeySMTPPort:           "SMTP 端口",
		ConfigKeySMTPUseTLS:         "使用 TLS 加密",
		ConfigKeySMTPUsername:       "SMTP 用户名",
		ConfigKeySMTPPassword:       "SMTP 密码 (加密存储)",
		ConfigKeySMTPFromName:       "发件人名称",
		ConfigKeySMTPFromEmail:      "发件人邮箱",
		ConfigKeyAlipayEnabled:      "启用支付宝支付",
		ConfigKeyAlipayAppID:        "支付宝应用 APP ID",
		ConfigKeyAlipayPrivateKey:   "支付宝应用私钥 (加密存储)",
		ConfigKeyAlipayPublicKey:    "支付宝公钥",
		ConfigKeyAlipayEncryptKey:   "支付宝加密密钥 (加密存储)",
		ConfigKeyAlipaySandbox:      "启用沙箱模式",
		ConfigKeySiteName:           "网站名称",
		ConfigKeySiteLogo:           "网站 Logo",
		ConfigKeySiteDescription:    "网站描述",
		ConfigKeyFreeRPM:            "免费用户每分钟请求数",
		ConfigKeyFreeTPM:            "免费用户每分钟 Token 数",
		ConfigKeyVIPRPM:             "VIP 用户每分钟请求数",
		ConfigKeyVIPTPM:             "VIP 用户每分钟 Token 数",
		ConfigKeyJWTSecret:          "JWT 签名密钥 (加密存储)",
		ConfigKeyJWTExpireHours:     "JWT Token 过期时间 (小时)",
		ConfigKeyPasswordMinLength:  "密码最小长度",
		ConfigKeyPasswordExpireDays: "密码过期天数",
	}
	if desc, ok := descriptions[key]; ok {
		return desc
	}
	return ""
}

type RegisterSettings struct {
	AllowRegister            bool   `json:"allow_register"`
	RequireEmailVerify       *bool `json:"require_email_verify,omitempty"`
	EnableCaptcha            bool  `json:"enable_captcha"`
	NewUserQuota            int   `json:"new_user_quota"`
	TrialVIPDays             int   `json:"trial_vip_days"`
	AllowedDomains           string `json:"allowed_domains"`
	MaxAccountsPerIP         int    `json:"max_accounts_per_ip"`
	MinPasswordLength        int    `json:"min_password_length"`
	SignupRewardType         string `json:"signup_reward_type"`
	SignupRewardAmount       int64  `json:"signup_reward_amount"`
}

type AlipayConfig struct {
	Enabled    bool   `json:"enabled"`
	AppID      string `json:"app_id"`
	PrivateKey string `json:"-"`
	PublicKey  string `json:"public_key"`
	EncryptKey string `json:"-"`
	Sandbox    bool   `json:"sandbox"`
}

type GeneralSettings struct {
	SiteName        string `json:"site_name"`
	SiteLogo        string `json:"site_logo"`
	SiteDescription string `json:"site_description"`
}

type RateLimitSettings struct {
	FreeRPM int `json:"free_rpm"`
	FreeTPM int `json:"free_tpm"`
	VIPRPM  int `json:"vip_rpm"`
	VIPTPM  int `json:"vip_tpm"`
}

type SecuritySettings struct {
	JWTSecret          string `json:"jwt_secret"`
	JWTExpireHours     int    `json:"jwt_expire_hours"`
	PasswordMinLength  int    `json:"password_min_length"`
	PasswordExpireDays int    `json:"password_expire_days"`
}

func (s *SettingsService) GetRegisterSettings() (*RegisterSettings, error) {
	settings := &RegisterSettings{
		AllowRegister:        true,
		RequireEmailVerify:  boolPtr(true),
		EnableCaptcha:        true,
		NewUserQuota:         100000,
		TrialVIPDays:         0,
		AllowedDomains:       "",
		MaxAccountsPerIP:     5,
		MinPasswordLength:    8,
		SignupRewardType:    "quota",
		SignupRewardAmount:  100000,
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
		case ConfigKeyAllowedDomains:
			settings.AllowedDomains = c.ConfigValue
		case ConfigKeyMaxAccountsPerIP:
			fmt.Sscanf(c.ConfigValue, "%d", &settings.MaxAccountsPerIP)
		case ConfigKeyMinPasswordLength:
			fmt.Sscanf(c.ConfigValue, "%d", &settings.MinPasswordLength)
		case ConfigKeySignupRewardType:
			settings.SignupRewardType = c.ConfigValue
		case ConfigKeySignupRewardAmount:
			fmt.Sscanf(c.ConfigValue, "%d", &settings.SignupRewardAmount)
		}
	}

	return settings, nil
}

func (s *SettingsService) UpdateRegisterSettings(settings *RegisterSettings) error {
	type configEntry struct {
		value     string
		valueType string
	}
	configs := map[string]configEntry{
		ConfigKeyAllowRegister:       {boolToString(settings.AllowRegister), "boolean"},
		ConfigKeyEnableCaptcha:       {boolToString(settings.EnableCaptcha), "boolean"},
		ConfigKeyNewUserQuota:        {fmt.Sprintf("%d", settings.NewUserQuota), "number"},
		ConfigKeyTrialVIPDays:        {fmt.Sprintf("%d", settings.TrialVIPDays), "number"},
		ConfigKeyAllowedDomains:      {settings.AllowedDomains, "string"},
		ConfigKeyMaxAccountsPerIP:    {fmt.Sprintf("%d", settings.MaxAccountsPerIP), "number"},
		ConfigKeyMinPasswordLength:   {fmt.Sprintf("%d", settings.MinPasswordLength), "number"},
		ConfigKeySignupRewardType:    {settings.SignupRewardType, "string"},
		ConfigKeySignupRewardAmount:  {fmt.Sprintf("%d", settings.SignupRewardAmount), "number"},
	}

	if settings.RequireEmailVerify != nil {
		configs[ConfigKeyRequireEmailVerify] = configEntry{boolToString(*settings.RequireEmailVerify), "boolean"}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for key, entry := range configs {
			if err := s.upsertConfigWithGroup(tx, key, entry.value, entry.valueType, false, ConfigGroupRegister); err != nil {
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

func (s *SettingsService) GetAlipayConfig() (*AlipayConfig, error) {
	s.cacheMutex.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.alipayCache != nil {
		cfg := s.alipayCache
		s.cacheMutex.RUnlock()
		return cfg, nil
	}
	s.cacheMutex.RUnlock()

	cfg := &AlipayConfig{
		Enabled: false,
		Sandbox: true,
	}

	configs, err := s.getConfigsByGroup(ConfigGroupPayment)
	if err != nil {
		return cfg, err
	}

	for _, c := range configs {
		switch c.ConfigKey {
		case ConfigKeyAlipayEnabled:
			cfg.Enabled = c.ConfigValue == "true" || c.ConfigValue == "1"
		case ConfigKeyAlipayAppID:
			cfg.AppID = c.ConfigValue
		case ConfigKeyAlipayPrivateKey:
			if c.ConfigValue != "" {
				decrypted, err := crypto.Decrypt(c.ConfigValue)
				if err == nil {
					cfg.PrivateKey = decrypted
				}
			}
		case ConfigKeyAlipayPublicKey:
			cfg.PublicKey = c.ConfigValue
		case ConfigKeyAlipayEncryptKey:
			if c.ConfigValue != "" {
				decrypted, err := crypto.Decrypt(c.ConfigValue)
				if err == nil {
					cfg.EncryptKey = decrypted
				}
			}
		case ConfigKeyAlipaySandbox:
			cfg.Sandbox = c.ConfigValue == "true" || c.ConfigValue == "1"
		}
	}

	s.cacheMutex.Lock()
	s.alipayCache = cfg
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	return cfg, nil
}

func (s *SettingsService) UpdateAlipayConfig(cfg *AlipayConfig) error {
	updates := map[string]struct {
		Value       string
		IsSensitive bool
	}{
		ConfigKeyAlipayEnabled:    {Value: boolToString(cfg.Enabled), IsSensitive: false},
		ConfigKeyAlipayAppID:      {Value: cfg.AppID, IsSensitive: false},
		ConfigKeyAlipayPrivateKey: {Value: cfg.PrivateKey, IsSensitive: true},
		ConfigKeyAlipayPublicKey:  {Value: cfg.PublicKey, IsSensitive: false},
		ConfigKeyAlipayEncryptKey: {Value: cfg.EncryptKey, IsSensitive: true},
		ConfigKeyAlipaySandbox:    {Value: boolToString(cfg.Sandbox), IsSensitive: false},
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for key, data := range updates {
			value := data.Value

			if data.IsSensitive && value != "" {
				encrypted, err := crypto.Encrypt(value)
				if err != nil {
					return fmt.Errorf("failed to encrypt %s: %w", key, err)
				}
				value = encrypted
			}

			if key == ConfigKeyAlipayPrivateKey && value == "" {
				continue
			}
			if key == ConfigKeyAlipayEncryptKey && value == "" {
				continue
			}

			if err := s.upsertConfigWithGroup(tx, key, value, "string", data.IsSensitive, ConfigGroupPayment); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		s.InvalidateCache()
	}
	return err
}

func (s *SettingsService) upsertConfigWithGroup(tx *gorm.DB, key, value, valueType string, isSensitive bool, group string) error {
	var config model.SystemConfig
	err := tx.Where("config_key = ?", key).First(&config).Error

	description := getConfigDescription(key)

	if err == gorm.ErrRecordNotFound {
		config = model.SystemConfig{
			ConfigKey:   key,
			ConfigValue: value,
			ValueType:   valueType,
			ConfigGroup: group,
			IsSensitive: isSensitive,
			Description: description,
		}
		return tx.Create(&config).Error
	} else if err != nil {
		return err
	}

	config.ConfigValue = value
	if isSensitive {
		config.IsSensitive = true
	}
	return tx.Save(&config).Error
}

func (s *SettingsService) IsAlipayEnabled() bool {
	cfg, err := s.GetAlipayConfig()
	if err != nil {
		return false
	}
	return cfg.Enabled && cfg.AppID != "" && cfg.PrivateKey != ""
}

func (s *SettingsService) GetGeneralSettings() (*GeneralSettings, error) {
	s.cacheMutex.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.generalCache != nil {
		cfg := s.generalCache
		s.cacheMutex.RUnlock()
		return cfg, nil
	}
	s.cacheMutex.RUnlock()

	cfg := &GeneralSettings{
		SiteName:        "API Proxy Platform",
		SiteLogo:        "/static/logo.png",
		SiteDescription: "OpenAI API 代理平台",
	}

	configs, err := s.getConfigsByGroup(ConfigGroupGeneral)
	if err != nil {
		return cfg, err
	}

	for _, c := range configs {
		switch c.ConfigKey {
		case ConfigKeySiteName:
			cfg.SiteName = c.ConfigValue
		case ConfigKeySiteLogo:
			cfg.SiteLogo = c.ConfigValue
		case ConfigKeySiteDescription:
			cfg.SiteDescription = c.ConfigValue
		}
	}

	s.cacheMutex.Lock()
	s.generalCache = cfg
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	return cfg, nil
}

func (s *SettingsService) UpdateGeneralSettings(cfg *GeneralSettings) error {
	configs := map[string]struct {
		Value     string
		ValueType string
	}{
		ConfigKeySiteName:        {cfg.SiteName, "string"},
		ConfigKeySiteLogo:        {cfg.SiteLogo, "string"},
		ConfigKeySiteDescription: {cfg.SiteDescription, "string"},
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for key, data := range configs {
			if err := s.upsertConfigWithGroup(tx, key, data.Value, data.ValueType, false, ConfigGroupGeneral); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		s.InvalidateCache()
	}
	return err
}

func (s *SettingsService) GetRateLimitSettings() (*RateLimitSettings, error) {
	s.cacheMutex.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.rateLimitCache != nil {
		cfg := s.rateLimitCache
		s.cacheMutex.RUnlock()
		return cfg, nil
	}
	s.cacheMutex.RUnlock()

	cfg := &RateLimitSettings{
		FreeRPM: 60,
		FreeTPM: 10000,
		VIPRPM:  2000,
		VIPTPM:  500000,
	}

	configs, err := s.getConfigsByGroup(ConfigGroupRateLimit)
	if err != nil {
		return cfg, err
	}

	for _, c := range configs {
		switch c.ConfigKey {
		case ConfigKeyFreeRPM:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.FreeRPM)
		case ConfigKeyFreeTPM:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.FreeTPM)
		case ConfigKeyVIPRPM:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.VIPRPM)
		case ConfigKeyVIPTPM:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.VIPTPM)
		}
	}

	s.cacheMutex.Lock()
	s.rateLimitCache = cfg
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	return cfg, nil
}

func (s *SettingsService) UpdateRateLimitSettings(cfg *RateLimitSettings) error {
	configs := map[string]struct {
		Value     string
		ValueType string
	}{
		ConfigKeyFreeRPM: {fmt.Sprintf("%d", cfg.FreeRPM), "number"},
		ConfigKeyFreeTPM: {fmt.Sprintf("%d", cfg.FreeTPM), "number"},
		ConfigKeyVIPRPM:  {fmt.Sprintf("%d", cfg.VIPRPM), "number"},
		ConfigKeyVIPTPM:  {fmt.Sprintf("%d", cfg.VIPTPM), "number"},
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for key, data := range configs {
			if err := s.upsertConfigWithGroup(tx, key, data.Value, data.ValueType, false, ConfigGroupRateLimit); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		s.InvalidateCache()
	}
	return err
}

func (s *SettingsService) GetSecuritySettings() (*SecuritySettings, error) {
	s.cacheMutex.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && s.securityCache != nil {
		cfg := s.securityCache
		s.cacheMutex.RUnlock()
		return cfg, nil
	}
	s.cacheMutex.RUnlock()

	cfg := &SecuritySettings{
		JWTSecret:          "gapi-jwt-secret-key-change-in-production",
		JWTExpireHours:     168,
		PasswordMinLength:  8,
		PasswordExpireDays: 90,
	}

	configs, err := s.getConfigsByGroup(ConfigGroupSecurity)
	if err != nil {
		return cfg, err
	}

	for _, c := range configs {
		switch c.ConfigKey {
		case ConfigKeyJWTSecret:
			if c.ConfigValue != "" {
				decrypted, err := crypto.Decrypt(c.ConfigValue)
				if err == nil {
					cfg.JWTSecret = decrypted
				}
			}
		case ConfigKeyJWTExpireHours:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.JWTExpireHours)
		case ConfigKeyPasswordMinLength:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.PasswordMinLength)
		case ConfigKeyPasswordExpireDays:
			fmt.Sscanf(c.ConfigValue, "%d", &cfg.PasswordExpireDays)
		}
	}

	s.cacheMutex.Lock()
	s.securityCache = cfg
	s.cacheTime = time.Now()
	s.cacheMutex.Unlock()

	return cfg, nil
}

func (s *SettingsService) UpdateSecuritySettings(cfg *SecuritySettings) error {
	updates := map[string]struct {
		Value       string
		IsSensitive bool
	}{
		ConfigKeyJWTSecret:          {Value: cfg.JWTSecret, IsSensitive: true},
		ConfigKeyJWTExpireHours:     {Value: fmt.Sprintf("%d", cfg.JWTExpireHours), IsSensitive: false},
		ConfigKeyPasswordMinLength:  {Value: fmt.Sprintf("%d", cfg.PasswordMinLength), IsSensitive: false},
		ConfigKeyPasswordExpireDays: {Value: fmt.Sprintf("%d", cfg.PasswordExpireDays), IsSensitive: false},
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for key, data := range updates {
			value := data.Value

			if data.IsSensitive && value != "" {
				encrypted, err := crypto.Encrypt(value)
				if err != nil {
					return fmt.Errorf("failed to encrypt %s: %w", key, err)
				}
				value = encrypted
			}

			if key == ConfigKeyJWTSecret && value == "" {
				continue
			}

			if err := s.upsertConfigWithGroup(tx, key, value, "string", data.IsSensitive, ConfigGroupSecurity); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		s.InvalidateCache()
	}
	return err
}

func boolPtr(b bool) *bool {
	return &b
}
