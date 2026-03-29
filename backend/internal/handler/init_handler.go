package handler

import (
	"context"
	"fmt"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"
	mq "gapi-platform/internal/mq"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type InitHandler struct {
	DB       *gorm.DB
	Redis    *repository.RedisClient
	MQ       *mq.Client
	cfgAdmin []config.AdminAccount
}

func NewInitHandler(db *gorm.DB, redisClient *repository.RedisClient, mqClient *mq.Client, adminCfg []config.AdminAccount) *InitHandler {
	return &InitHandler{DB: db, Redis: redisClient, MQ: mqClient, cfgAdmin: adminCfg}
}

// GetStatus returns the initialization status of the system
// Checks admin existence, DB/Redis/RabbitMQ connectivity
func (h *InitHandler) GetStatus(c *gin.Context) {
	// Admin existence
	var count int64
	adminExists := false
	if h.DB != nil {
		// Admin table: admin_users
		if err := h.DB.Model(&model.AdminUser{}).Count(&count).Error; err == nil {
			if count > 0 {
				adminExists = true
			}
		}
	}

	// DB connectivity
	dbConnected := false
	if h.DB != nil {
		if sqlDB, err := h.DB.DB(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err == nil {
				dbConnected = true
			}
		}
	}

	// Redis connectivity
	redisConnected := false
	if h.Redis != nil && h.Redis.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.Redis.Client.Ping(ctx).Err(); err == nil {
			redisConnected = true
		}
	}

	// RabbitMQ connectivity
	rabbitmqConnected := false
	if h.MQ != nil {
		rabbitmqConnected = h.MQ.IsConnected()
	}

	// needs_init is true if no admin exists
	needsInit := !adminExists

	response.Success(c, gin.H{
		"needs_init":         needsInit,
		"db_connected":       dbConnected,
		"redis_connected":    redisConnected,
		"rabbitmq_connected": rabbitmqConnected,
		"admin_exists":       adminExists,
	})
}

// TestDatabase tests the database connectivity and returns a simple result
func (h *InitHandler) TestDatabase(c *gin.Context) {
	if h.DB == nil {
		response.Fail(c, "DB_UNINITIALIZED", "database not configured")
		return
	}
	if sqlDB, err := h.DB.DB(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
			return
		}
		response.SuccessWithMessage(c, nil, "database connected")
		return
	}
	response.InternalError(c, "failed to access database connection")
}

// TestRedis tests the Redis connectivity (helper used by Init status)
func (h *InitHandler) TestRedis(c *gin.Context) {
	if h.Redis == nil || h.Redis.Client == nil {
		response.Fail(c, "REDIS_UNINITIALIZED", "redis not configured")
		return
	}
	ctx := context.Background()
	if err := h.Redis.Client.Ping(ctx).Err(); err != nil {
		response.Fail(c, "REDIS_CONNECTION_FAILED", err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "redis connected")
}

// CreateAdmin creates the first admin user. Password is stored as bcrypt hash.
func (h *InitHandler) CreateAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	// Ensure DB is ready
	if h.DB == nil {
		response.Fail(c, "DB_UNINITIALIZED", "database not configured")
		return
	}

	// Check if admin already exists
	var existing model.AdminUser
	if err := h.DB.Model(&model.AdminUser{}).Where("username = ? OR email = ?", req.Username, req.Email).First(&existing).Error; err == nil {
		response.Fail(c, "ADMIN_EXISTS", "admin user already exists")
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	admin := model.AdminUser{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
		Role:         "admin",
		Permissions:  "[]", // Empty JSON array for permissions
		Status:       "active",
	}

	if err := h.DB.Create(&admin).Error; err != nil {
		response.InternalError(c, "failed to create admin user")
		return
	}

	// Mark system as initialized
	var sc model.SystemConfig
	if err := h.DB.Where("config_key = ?", "system_initialized").First(&sc).Error; err != nil {
		// Not found, create new
		sc = model.SystemConfig{
			ConfigKey:   "system_initialized",
			ConfigValue: "true",
			ValueType:   "boolean",
			ConfigGroup: "general",
			Description: "System initialization completed",
		}
		if err := h.DB.Create(&sc).Error; err != nil {
			response.InternalError(c, "admin created but failed to mark initialization")
			return
		}
	} else {
		sc.ConfigValue = "true"
		if err := h.DB.Save(&sc).Error; err != nil {
			response.InternalError(c, "admin created but failed to update initialization flag")
			return
		}
	}

	response.SuccessWithMessage(c, gin.H{
		"admin_id": admin.ID,
		"username": admin.Username,
	}, "admin created and system initialized")
}

type DBConfig struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password"`
	DBName   string `json:"dbname" binding:"required"`
}

func (h *InitHandler) TestDatabaseWithConfig(c *gin.Context) {
	var req DBConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		req.Host, req.Port, req.User, req.Password, req.DBName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
		return
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
		return
	}

	sqlDB.Close()
	response.SuccessWithMessage(c, nil, "database connected")
}

func (h *InitHandler) InitializeDatabase(c *gin.Context) {
	var req DBConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		req.Host, req.Port, req.User, req.Password, req.DBName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
		return
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		response.Fail(c, "DB_CONNECTION_FAILED", err.Error())
		return
	}

	err = db.AutoMigrate(
		&model.AdminUser{},
		&model.User{},
		&model.Channel{},
		&model.Ability{},
		&model.Token{},
		&model.VIPPackage{},
		&model.RechargePackage{},
		&model.Order{},
		&model.Payment{},
		&model.RedemptionCode{},
		&model.QuotaTransaction{},
		&model.AuditLog{},
		&model.LoginLog{},
		&model.UsageLog{},
		&model.APIAccessLog{},
		&model.ChannelTestHistory{},
		&model.SystemConfig{},
		&model.SignupConfig{},
	)
	if err != nil {
		response.Fail(c, "DB_MIGRATION_FAILED", err.Error())
		return
	}

	response.SuccessWithMessage(c, nil, "database initialized")
}
