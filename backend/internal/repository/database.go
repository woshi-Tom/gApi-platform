package repository

import (
	"context"
	"fmt"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps both pgx and gorm
type Database struct {
	DB   *gorm.DB
	Pool *pgxpool.Pool
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	// First connect with pgx to create database if needed
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Now setup GORM
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DSN(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Database{
		DB:   db,
		Pool: pool,
	}, nil
}

// AutoMigrate runs auto migration
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&model.Tenant{},
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
		&model.ChannelTestHistory{},
		&model.AuditLog{},
		&model.LoginLog{},
		&model.UsageLog{},
		&model.SystemConfig{},
		&model.SignupConfig{},
	)
}

// Close closes the database connection
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
	if sqlDB, err := d.DB.DB(); err == nil {
		sqlDB.Close()
	}
}

// GetDB returns the GORM database
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}
