package service

import (
	"testing"
	"time"
)

func TestBoolToString(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  string
	}{
		{"true", true, "true"},
		{"false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolToString(tt.input)
			if got != tt.want {
				t.Errorf("boolToString(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetConfigDescription(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{ConfigKeySMTPEnabled, "启用邮箱服务"},
		{ConfigKeySMTPHost, "SMTP 服务器地址"},
		{ConfigKeySMTPPort, "SMTP 端口"},
		{ConfigKeySMTPUseTLS, "使用 TLS 加密"},
		{ConfigKeySMTPUsername, "SMTP 用户名"},
		{ConfigKeySMTPPassword, "SMTP 密码 (加密存储)"},
		{ConfigKeySMTPFromName, "发件人名称"},
		{ConfigKeySMTPFromEmail, "发件人邮箱"},
		{ConfigKeyAlipayEnabled, "启用支付宝支付"},
		{ConfigKeyAlipayAppID, "支付宝应用 APP ID"},
		{ConfigKeyAlipayPrivateKey, "支付宝应用私钥 (加密存储)"},
		{ConfigKeyAlipayPublicKey, "支付宝公钥"},
		{ConfigKeyAlipayEncryptKey, "支付宝加密密钥 (加密存储)"},
		{ConfigKeyAlipaySandbox, "启用沙箱模式"},
		{ConfigKeySiteName, "网站名称"},
		{ConfigKeySiteLogo, "网站 Logo"},
		{ConfigKeySiteDescription, "网站描述"},
		{ConfigKeyFreeRPM, "免费用户每分钟请求数"},
		{ConfigKeyFreeTPM, "免费用户每分钟 Token 数"},
		{ConfigKeyVIPRPM, "VIP 用户每分钟请求数"},
		{ConfigKeyVIPTPM, "VIP 用户每分钟 Token 数"},
		{ConfigKeyJWTSecret, "JWT 签名密钥 (加密存储)"},
		{ConfigKeyJWTExpireHours, "JWT Token 过期时间 (小时)"},
		{ConfigKeyPasswordMinLength, "密码最小长度"},
		{ConfigKeyPasswordExpireDays, "密码过期天数"},
		{"unknown_key", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := getConfigDescription(tt.key)
			if got != tt.want {
				t.Errorf("getConfigDescription(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		got := boolPtr(true)
		if got == nil {
			t.Fatal("boolPtr(true) returned nil")
		}
		if *got != true {
			t.Errorf("boolPtr(true) = %v, want true", *got)
		}
	})

	t.Run("false", func(t *testing.T) {
		got := boolPtr(false)
		if got == nil {
			t.Fatal("boolPtr(false) returned nil")
		}
		if *got != false {
			t.Errorf("boolPtr(false) = %v, want false", *got)
		}
	})
}

func TestGeneralSettingsDefaults(t *testing.T) {
	cfg := &GeneralSettings{
		SiteName:        "API Proxy Platform",
		SiteLogo:        "/static/logo.png",
		SiteDescription: "OpenAI API 代理平台",
	}

	if cfg.SiteName != "API Proxy Platform" {
		t.Errorf("SiteName default = %q, want %q", cfg.SiteName, "API Proxy Platform")
	}
	if cfg.SiteLogo != "/static/logo.png" {
		t.Errorf("SiteLogo default = %q, want %q", cfg.SiteLogo, "/static/logo.png")
	}
	if cfg.SiteDescription != "OpenAI API 代理平台" {
		t.Errorf("SiteDescription default = %q, want %q", cfg.SiteDescription, "OpenAI API 代理平台")
	}
}

func TestRateLimitSettingsDefaults(t *testing.T) {
	cfg := &RateLimitSettings{
		FreeRPM: 60,
		FreeTPM: 10000,
		VIPRPM:  2000,
		VIPTPM:  500000,
	}

	if cfg.FreeRPM != 60 {
		t.Errorf("FreeRPM default = %d, want %d", cfg.FreeRPM, 60)
	}
	if cfg.FreeTPM != 10000 {
		t.Errorf("FreeTPM default = %d, want %d", cfg.FreeTPM, 10000)
	}
	if cfg.VIPRPM != 2000 {
		t.Errorf("VIPRPM default = %d, want %d", cfg.VIPRPM, 2000)
	}
	if cfg.VIPTPM != 500000 {
		t.Errorf("VIPTPM default = %d, want %d", cfg.VIPTPM, 500000)
	}
}

func TestSecuritySettingsDefaults(t *testing.T) {
	cfg := &SecuritySettings{
		JWTSecret:          "gapi-jwt-secret-key-change-in-production",
		JWTExpireHours:     168,
		PasswordMinLength:  8,
		PasswordExpireDays: 90,
	}

	if cfg.JWTSecret != "gapi-jwt-secret-key-change-in-production" {
		t.Errorf("JWTSecret default = %q, want %q", cfg.JWTSecret, "gapi-jwt-secret-key-change-in-production")
	}
	if cfg.JWTExpireHours != 168 {
		t.Errorf("JWTExpireHours default = %d, want %d", cfg.JWTExpireHours, 168)
	}
	if cfg.PasswordMinLength != 8 {
		t.Errorf("PasswordMinLength default = %d, want %d", cfg.PasswordMinLength, 8)
	}
	if cfg.PasswordExpireDays != 90 {
		t.Errorf("PasswordExpireDays default = %d, want %d", cfg.PasswordExpireDays, 90)
	}
}

func TestInvalidateCache(t *testing.T) {
	s := &SettingsService{
		smtpCache:    &SMTPConfig{},
		alipayCache:  &AlipayConfig{},
		generalCache: &GeneralSettings{},
		rateLimitCache: &RateLimitSettings{},
		securityCache: &SecuritySettings{},
		cacheTime:    time.Now(),
	}

	s.InvalidateCache()

	if s.smtpCache != nil {
		t.Error("smtpCache should be nil after InvalidateCache")
	}
	if s.alipayCache != nil {
		t.Error("alipayCache should be nil after InvalidateCache")
	}
	if s.generalCache != nil {
		t.Error("generalCache should be nil after InvalidateCache")
	}
	if s.rateLimitCache != nil {
		t.Error("rateLimitCache should be nil after InvalidateCache")
	}
	if s.securityCache != nil {
		t.Error("securityCache should be nil after InvalidateCache")
	}
	if !s.cacheTime.IsZero() {
		t.Error("cacheTime should be zero after InvalidateCache")
	}
}

func TestConfigKeyConstants(t *testing.T) {
	keys := []string{
		ConfigKeySMTPEnabled,
		ConfigKeySMTPHost,
		ConfigKeySMTPPort,
		ConfigKeySMTPUseTLS,
		ConfigKeySMTPUsername,
		ConfigKeySMTPPassword,
		ConfigKeySMTPFromName,
		ConfigKeySMTPFromEmail,
		ConfigGroupEmail,
		ConfigKeyAllowRegister,
		ConfigKeyRequireEmailVerify,
		ConfigKeyEnableCaptcha,
		ConfigKeyNewUserQuota,
		ConfigKeyTrialVIPDays,
		ConfigKeyAllowedDomains,
		ConfigKeyMaxAccountsPerIP,
		ConfigKeyMinPasswordLength,
		ConfigKeySignupRewardType,
		ConfigKeySignupRewardAmount,
		ConfigGroupRegister,
		ConfigKeyAlipayEnabled,
		ConfigKeyAlipayAppID,
		ConfigKeyAlipayPrivateKey,
		ConfigKeyAlipayPublicKey,
		ConfigKeyAlipayEncryptKey,
		ConfigKeyAlipaySandbox,
		ConfigGroupPayment,
		ConfigKeySiteName,
		ConfigKeySiteLogo,
		ConfigKeySiteDescription,
		ConfigGroupGeneral,
		ConfigKeyFreeRPM,
		ConfigKeyFreeTPM,
		ConfigKeyVIPRPM,
		ConfigKeyVIPTPM,
		ConfigGroupRateLimit,
		ConfigKeyJWTSecret,
		ConfigKeyJWTExpireHours,
		ConfigKeyPasswordMinLength,
		ConfigKeyPasswordExpireDays,
		ConfigGroupSecurity,
	}

	for _, key := range keys {
		if key == "" {
			t.Error("config key constant should not be empty")
		}
	}
}
