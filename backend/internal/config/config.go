package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration
type Config struct {
	Server     ServerConfig   `yaml:"server" json:"server"`
	Database   DatabaseConfig `yaml:"database" json:"database"`
	Redis      RedisConfig    `yaml:"redis" json:"redis"`
	RabbitMQ   RabbitMQConfig `yaml:"rabbitmq" json:"rabbitmq"`
	JWT        JWTConfig      `yaml:"jwt" json:"jwt"`
	Payment    PaymentConfig  `yaml:"payment" json:"payment"`
	SMTP       SMTPConfig     `yaml:"smtp" json:"smtp"`
	Log        LogConfig      `yaml:"log" json:"log"`
	AdminUsers []AdminAccount `yaml:"admin_users" json:"admin_users"`
}

// AdminAccount represents an admin user account
type AdminAccount struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Role     string `yaml:"role" json:"role"`
}

type ServerConfig struct {
	Port          string `yaml:"port" json:"port"`
	AdminPort     string `yaml:"admin_port" json:"admin_port"`
	AdminBind     string `yaml:"admin_bind" json:"admin_bind"`
	Mode          string `yaml:"mode" json:"mode"`
	Timeout       int    `yaml:"timeout" json:"timeout"`
	Frontend      string `yaml:"frontend" json:"frontend"`
	AdminFrontend string `yaml:"admin_frontend" json:"admin_frontend"`
	AdminSecret   string `yaml:"admin_secret" json:"admin_secret"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Database string `yaml:"database" json:"database"`
	MaxOpen  int    `yaml:"max_open" json:"max_open"`
	MaxIdle  int    `yaml:"max_idle" json:"max_idle"`
}

type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret" json:"secret"`
	ExpireHour int    `yaml:"expire_hour" json:"expire_hour"`
}

type PaymentConfig struct {
	Alipay AlipayConfig `yaml:"alipay" json:"alipay"`
	Wechat WechatConfig `yaml:"wechat" json:"wechat"`
}

type AlipayConfig struct {
	AppID           string `yaml:"app_id" json:"app_id"`
	PrivateKey      string `yaml:"private_key" json:"private_key"`
	AlipayPublicKey string `yaml:"alipay_public_key" json:"alipay_public_key"`
	Sandbox         bool   `yaml:"sandbox" json:"sandbox"`
	Enabled         bool   `yaml:"enabled" json:"enabled"`
}

type WechatConfig struct {
	AppID    string `yaml:"app_id" json:"app_id"`
	MchID    string `yaml:"mch_id" json:"mch_id"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	CertPath string `yaml:"cert_path" json:"cert_path"`
	Enabled  bool   `yaml:"enabled" json:"enabled"`
}

type SMTPConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	From     string `yaml:"from" json:"from"`
	FromName string `yaml:"from_name" json:"from_name"`
	TLS      bool   `yaml:"tls" json:"tls"`
}

type LogConfig struct {
	Level  string `yaml:"level" json:"level"`
	Path   string `yaml:"path" json:"path"`
	Format string `yaml:"format" json:"format"`
}

// Load loads configuration from YAML file and environment variables
func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Try to load from YAML file first
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			// If file doesn't exist, continue with env vars only
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("read config file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse yaml: %w", err)
			}
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	// Set defaults
	cfg.setDefaults()

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) loadFromEnv() {
	// Server
	if v := os.Getenv("GAPI_SERVER_PORT"); v != "" {
		c.Server.Port = v
	}
	if v := os.Getenv("GAPI_ADMIN_PORT"); v != "" {
		c.Server.AdminPort = v
	}
	if v := os.Getenv("GAPI_ADMIN_BIND"); v != "" {
		c.Server.AdminBind = v
	}
	if v := os.Getenv("GAPI_MODE"); v != "" {
		c.Server.Mode = v
	}
	if v := os.Getenv("GAPI_FRONTEND_URL"); v != "" {
		c.Server.Frontend = v
	}
	if v := os.Getenv("GAPI_ADMIN_SECRET"); v != "" {
		c.Server.AdminSecret = v
	}

	// Database
	if v := os.Getenv("GAPI_DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("GAPI_DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Database.Port = port
		}
	}
	if v := os.Getenv("GAPI_DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("GAPI_DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("GAPI_DB_NAME"); v != "" {
		c.Database.Database = v
	}

	// Redis
	if v := os.Getenv("GAPI_REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("GAPI_REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Redis.Port = port
		}
	}
	if v := os.Getenv("GAPI_REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}

	// RabbitMQ
	if v := os.Getenv("GAPI_RABBITMQ_HOST"); v != "" {
		c.RabbitMQ.Host = v
	}
	if v := os.Getenv("GAPI_RABBITMQ_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.RabbitMQ.Port = port
		}
	}
	if v := os.Getenv("GAPI_RABBITMQ_USER"); v != "" {
		c.RabbitMQ.User = v
	}
	if v := os.Getenv("GAPI_RABBITMQ_PASSWORD"); v != "" {
		c.RabbitMQ.Password = v
	}

	if v := os.Getenv("GAPI_SMTP_HOST"); v != "" {
		c.SMTP.Host = v
	}
	if v := os.Getenv("GAPI_SMTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.SMTP.Port = port
		}
	}
	if v := os.Getenv("GAPI_SMTP_USERNAME"); v != "" {
		c.SMTP.Username = v
	}
	if v := os.Getenv("GAPI_SMTP_PASSWORD"); v != "" {
		c.SMTP.Password = v
	}
	if v := os.Getenv("GAPI_SMTP_FROM"); v != "" {
		c.SMTP.From = v
	}
	if v := os.Getenv("GAPI_SMTP_FROM_NAME"); v != "" {
		c.SMTP.FromName = v
	}
	if v := os.Getenv("GAPI_SMTP_TLS"); v != "" {
		c.SMTP.TLS = v == "true"
	}

	if v := os.Getenv("GAPI_JWT_SECRET"); v != "" {
		c.JWT.Secret = v
	}

	// Log
	if v := os.Getenv("GAPI_LOG_PATH"); v != "" {
		c.Log.Path = v
	}
}

func (c *Config) setDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = "8080"
	}
	if c.Server.AdminPort == "" {
		c.Server.AdminPort = "9000"
	}
	if c.Server.AdminBind == "" {
		c.Server.AdminBind = "127.0.0.1"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.Server.Timeout == 0 {
		c.Server.Timeout = 60
	}
	if c.Database.MaxOpen == 0 {
		c.Database.MaxOpen = 100
	}
	if c.Database.MaxIdle == 0 {
		c.Database.MaxIdle = 10
	}
	if c.Redis.DB == 0 {
		c.Redis.DB = 0
	}
	if c.JWT.ExpireHour == 0 {
		c.JWT.ExpireHour = 24
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "console"
	}
}

func (c *Config) Validate() error {
	// Database validation
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// Redis validation
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	// JWT validation
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters")
	}

	return nil
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Database)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RabbitMQAddr returns the RabbitMQ address
func (c *RabbitMQConfig) Addr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.User, c.Password, c.Host, c.Port)
}

// JWTExpiry returns the JWT expiry duration
func (c *JWTConfig) Expiry() time.Duration {
	return time.Duration(c.ExpireHour) * time.Hour
}
