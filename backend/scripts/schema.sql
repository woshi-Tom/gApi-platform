-- gAPI Platform Database Schema
-- Generated from GORM models
-- PostgreSQL 14+
-- Run: psql -U gapi -d gapi -f schema.sql

BEGIN;

-- ============================================================
-- Core Tables
-- ============================================================

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(50) UNIQUE NOT NULL,
    description     TEXT,
    max_users       INTEGER DEFAULT 100,
    max_channels    INTEGER DEFAULT 50,
    max_tokens      INTEGER DEFAULT 100,
    features        JSONB DEFAULT '{}',
    status          VARCHAR(20) DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tenants_code ON tenants(code);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;

-- Admin Users table
CREATE TABLE IF NOT EXISTS admin_users (
    id                   BIGSERIAL PRIMARY KEY,
    tenant_id            BIGINT,
    username             VARCHAR(50) NOT NULL UNIQUE,
    password_hash        VARCHAR(255) NOT NULL,
    email               VARCHAR(100) NOT NULL UNIQUE,
    phone               VARCHAR(20),
    avatar              VARCHAR(500),
    role                VARCHAR(20) DEFAULT 'admin',
    permissions         JSONB,
    status              VARCHAR(20) DEFAULT 'active',
    last_login_at       TIMESTAMPTZ,
    last_login_ip       VARCHAR(50),
    password_changed_at  TIMESTAMPTZ DEFAULT NOW(),
    password_expire_days INTEGER DEFAULT 90,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          BIGINT,
    deleted_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_admin_users_tenant ON admin_users(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_admin_users_deleted ON admin_users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    username         VARCHAR(50) NOT NULL,
    email            VARCHAR(100) NOT NULL,
    phone            VARCHAR(20),
    password_hash    VARCHAR(255) NOT NULL,
    email_verified  BOOLEAN DEFAULT FALSE,
    verify_token    VARCHAR(100),
    verify_expired  TIMESTAMPTZ,
    level           VARCHAR(20) DEFAULT 'free',  -- free|vip_bronze|vip_silver|vip_gold
    vip_expired_at  TIMESTAMPTZ,
    vip_package_id   BIGINT,
    -- New quota fields (all quotas expire)
    free_quota      BIGINT DEFAULT 0,           -- Free quota (7 days)
    free_expired_at TIMESTAMPTZ,                -- Free quota expiry
    vip_quota       BIGINT DEFAULT 0,           -- VIP quota (30 days)
    -- Legacy field kept for backward compatibility
    remain_quota    BIGINT DEFAULT 0,           -- Deprecated, use free_quota
    status          VARCHAR(20) DEFAULT 'active',
    disabled_reason VARCHAR(200),
    last_login_at   TIMESTAMPTZ,
    last_login_ip   VARCHAR(50),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Channels table
CREATE TABLE IF NOT EXISTS channels (
    id                    BIGSERIAL PRIMARY KEY,
    tenant_id             BIGINT NOT NULL DEFAULT 1,
    name                  VARCHAR(100) NOT NULL,
    type                  VARCHAR(50) NOT NULL,
    base_url              VARCHAR(500) NOT NULL,
    api_key_encrypted     VARCHAR(500) NOT NULL,
    key_version           INTEGER DEFAULT 1,
    models                JSONB,
    model_mapping         JSONB,
    weight                INTEGER DEFAULT 100,
    priority              INTEGER DEFAULT 0,
    rpm_limit             INTEGER DEFAULT 1000,
    tpm_limit             INTEGER DEFAULT 100000,
    cost_factor           DECIMAL(5,2) DEFAULT 1.0,
    price_per_1k_input    DECIMAL(10,4) DEFAULT 0.01,
    price_per_1k_output   DECIMAL(10,4) DEFAULT 0.03,
    group_name            VARCHAR(50) DEFAULT 'default',
    status                INTEGER DEFAULT 1,
    is_healthy            BOOLEAN DEFAULT TRUE,
    failure_count         INTEGER DEFAULT 0,
    last_success_at       TIMESTAMPTZ,
    last_check_at         TIMESTAMPTZ,
    last_error            TEXT,
    response_time_avg     INTEGER DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by            BIGINT,
    deleted_at            TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_channels_tenant ON channels(tenant_id);
CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status);
CREATE INDEX IF NOT EXISTS idx_channels_priority ON channels(priority);
CREATE INDEX IF NOT EXISTS idx_channels_deleted ON channels(deleted_at) WHERE deleted_at IS NOT NULL;

-- Abilities table (channel capabilities)
CREATE TABLE IF NOT EXISTS abilities (
    id            BIGSERIAL PRIMARY KEY,
    channel_id   BIGINT NOT NULL,
    ability_type VARCHAR(50) NOT NULL,
    model         VARCHAR(100) NOT NULL,
    model_alias   VARCHAR(100),
    config        JSONB,
    is_enabled    BOOLEAN DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_abilities_channel ON abilities(channel_id);

-- Tokens table
CREATE TABLE IF NOT EXISTS tokens (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    user_id          BIGINT NOT NULL,
    name             VARCHAR(100) NOT NULL,
    token_key        VARCHAR(64) NOT NULL UNIQUE,
    token_hash       VARCHAR(64) NOT NULL,
    key_prefix       VARCHAR(10) DEFAULT 'sk-ap-',
    remain_quota     BIGINT DEFAULT 0,
    is_vip_quota     BOOLEAN DEFAULT FALSE,
    allowed_models   JSONB,
    denied_models    JSONB,
    allowed_ips      JSONB,
    rpm_limit        INTEGER,
    tpm_limit        INTEGER,
    max_usage_per_day BIGINT,
    expires_at       TIMESTAMPTZ,
    unlimited_quota  BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) DEFAULT 'active',
    used_quota       BIGINT DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at     TIMESTAMPTZ,
    last_used_ip     VARCHAR(50),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tokens_tenant ON tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_deleted ON tokens(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- VIP & Recharge Packages
-- ============================================================

-- VIP Packages
CREATE TABLE IF NOT EXISTS vip_packages (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    name             VARCHAR(100) NOT NULL,
    description      TEXT,
    price            DECIMAL(10,2) NOT NULL,
    original_price   DECIMAL(10,2),
    duration_days    INTEGER DEFAULT 30,
    quota            BIGINT DEFAULT 1000000,
    rpm_limit        INTEGER DEFAULT 2000,
    tpm_limit        INTEGER DEFAULT 100000,
    concurrent_limit INTEGER DEFAULT 10,
    features         TEXT,
    sort_order       INTEGER DEFAULT 0,
    is_recommended   BOOLEAN DEFAULT FALSE,
    is_popular       BOOLEAN DEFAULT FALSE,
    level            VARCHAR(20) DEFAULT 'vip_bronze',
    status           VARCHAR(20) DEFAULT 'active',
    is_visible       BOOLEAN DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_vip_packages_tenant ON vip_packages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vip_packages_status ON vip_packages(status);
CREATE INDEX IF NOT EXISTS idx_vip_packages_deleted ON vip_packages(deleted_at) WHERE deleted_at IS NOT NULL;

-- Recharge Packages
CREATE TABLE IF NOT EXISTS recharge_packages (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    name             VARCHAR(100) NOT NULL,
    description      TEXT,
    price            DECIMAL(10,2) NOT NULL,
    original_price   DECIMAL(10,2),
    quota            BIGINT NOT NULL,
    bonus_quota      BIGINT DEFAULT 0,
    valid_days       INTEGER DEFAULT 7,          -- Validity period in days
    rpm_limit        INTEGER DEFAULT 60,
    tpm_limit        INTEGER DEFAULT 6000,
    sort_order       INTEGER DEFAULT 0,
    is_recommended   BOOLEAN DEFAULT FALSE,
    is_popular       BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) DEFAULT 'active',
    is_visible       BOOLEAN DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_recharge_packages_tenant ON recharge_packages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recharge_packages_status ON recharge_packages(status);
CREATE INDEX IF NOT EXISTS idx_recharge_packages_deleted ON recharge_packages(deleted_at) WHERE deleted_at IS NOT NULL;

-- User Recharge Records (for FIFO consumption tracking)
CREATE TABLE IF NOT EXISTS user_recharge_records (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    package_id  BIGINT NOT NULL,
    order_id    BIGINT NOT NULL,
    quota       BIGINT NOT NULL,
    remaining   BIGINT NOT NULL,
    expired_at  TIMESTAMPTZ NOT NULL,
    status      VARCHAR(20) DEFAULT 'active',  -- active|used|expired
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_recharge_user ON user_recharge_records(user_id);
CREATE INDEX IF NOT EXISTS idx_user_recharge_order ON user_recharge_records(order_id);
CREATE INDEX IF NOT EXISTS idx_user_recharge_status ON user_recharge_records(status);

-- ============================================================
-- Orders & Payments
-- ============================================================

-- Orders
CREATE TABLE IF NOT EXISTS orders (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL DEFAULT 1,
    user_id         BIGINT NOT NULL,
    order_no        VARCHAR(50) NOT NULL UNIQUE,
    order_type      VARCHAR(20) NOT NULL,
    package_id      BIGINT,
    package_name    VARCHAR(100),
    total_amount    DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    pay_amount      DECIMAL(10,2) NOT NULL,
    status          VARCHAR(20) DEFAULT 'pending',
    paid_at         TIMESTAMPTZ,
    cancel_reason   VARCHAR(200),
    refund_reason   TEXT,
    refund_amount   DECIMAL(10,2),
    expire_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_tenant ON orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_package ON orders(package_id);

-- Payments
CREATE TABLE IF NOT EXISTS payments (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    user_id          BIGINT NOT NULL,
    order_id         BIGINT NOT NULL,
    payment_no       VARCHAR(100) NOT NULL UNIQUE,
    payment_method   VARCHAR(20) NOT NULL,
    amount           DECIMAL(10,2) NOT NULL,
    status           VARCHAR(20) DEFAULT 'pending',
    paid_at          TIMESTAMPTZ,
    channel_order_no VARCHAR(100),
    channel_trade_no VARCHAR(100),
    payment_url      TEXT,
    qr_code          TEXT,
    callback_url     VARCHAR(500),
    callback_body    TEXT,
    callback_at      TIMESTAMPTZ,
    error_code       VARCHAR(50),
    error_message    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payments_tenant ON payments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(order_id);

-- Redemption Codes
CREATE TABLE IF NOT EXISTS redemption_codes (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT NOT NULL DEFAULT 1,
    code         VARCHAR(50) NOT NULL UNIQUE,
    code_type    VARCHAR(20) NOT NULL,
    quota        BIGINT DEFAULT 0,
    quota_type   VARCHAR(10) DEFAULT 'permanent',
    vip_days     INTEGER DEFAULT 0,
    is_permanent BOOLEAN DEFAULT FALSE,
    max_uses     INTEGER DEFAULT 1,
    used_count   INTEGER DEFAULT 0,
    valid_from   TIMESTAMPTZ,
    valid_until  TIMESTAMPTZ,
    bound_user_id BIGINT,
    bound_at     TIMESTAMPTZ,
    status       VARCHAR(20) DEFAULT 'active',
    created_by   BIGINT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at      TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_redemption_codes_tenant ON redemption_codes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_redemption_codes_bound ON redemption_codes(bound_user_id) WHERE bound_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_redemption_codes_deleted ON redemption_codes(deleted_at) WHERE deleted_at IS NOT NULL;

-- Quota Transactions
CREATE TABLE IF NOT EXISTS quota_transactions (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT NOT NULL DEFAULT 1,
    user_id         BIGINT NOT NULL,
    token_id        BIGINT,
    type            VARCHAR(20) NOT NULL,
    quota_type      VARCHAR(10) NOT NULL,
    change_amount   BIGINT NOT NULL,
    balance_before  BIGINT NOT NULL,
    balance_after   BIGINT NOT NULL,
    order_id        BIGINT,
    package_id      BIGINT,
    channel_id      BIGINT,
    model           VARCHAR(100),
    description     VARCHAR(500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_quota_trans_tenant ON quota_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_quota_trans_user ON quota_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_quota_trans_token ON quota_transactions(token_id) WHERE token_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_quota_trans_order ON quota_transactions(order_id) WHERE order_id IS NOT NULL;

-- ============================================================
-- Logs & Audit
-- ============================================================

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT,
    user_id         BIGINT,
    username        VARCHAR(100),
    action          VARCHAR(100) NOT NULL,
    action_group    VARCHAR(50) NOT NULL,
    resource_type   VARCHAR(50),
    resource_id     BIGINT,
    request_method  VARCHAR(10),
    request_path    VARCHAR(500),
    request_body    TEXT,
    request_ip      VARCHAR(50),
    request_ua     VARCHAR(500),
    status_code     INTEGER,
    response_body   TEXT,
    success         BOOLEAN DEFAULT TRUE,
    error_message   TEXT,
    old_value       TEXT,
    new_value       TEXT,
    user_agent      VARCHAR(500),
    session_id      VARCHAR(100),
    trace_id        VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit_logs(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);

-- Login Logs
CREATE TABLE IF NOT EXISTS login_logs (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT,
    user_id      BIGINT,
    username     VARCHAR(100),
    login_type   VARCHAR(20) NOT NULL,
    ip           VARCHAR(50),
    ip_location  VARCHAR(200),
    user_agent   VARCHAR(500),
    device_type  VARCHAR(50),
    success      BOOLEAN DEFAULT FALSE,
    fail_reason  VARCHAR(100),
    token        VARCHAR(500),
    token_expired_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_login_tenant ON login_logs(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_login_user ON login_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_login_created ON login_logs(created_at);

-- Usage Logs (for billing)
CREATE TABLE IF NOT EXISTS usage_logs (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    user_id          BIGINT NOT NULL,
    token_id         BIGINT,
    channel_id       BIGINT,
    request_id       VARCHAR(64),
    model            VARCHAR(100) NOT NULL,
    prompt_tokens    INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens     INTEGER DEFAULT 0,
    cost             DECIMAL(10,4) DEFAULT 0,
    status_code      INTEGER,
    response_time_ms INTEGER DEFAULT 0,
    error_message    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_usage_tenant ON usage_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_user ON usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_token ON usage_logs(token_id) WHERE token_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_usage_channel ON usage_logs(channel_id) WHERE channel_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_usage_created ON usage_logs(created_at);

-- API Access Logs (for user dashboard)
CREATE TABLE IF NOT EXISTS api_access_logs (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL,
    tenant_id         BIGINT NOT NULL DEFAULT 1,
    endpoint          VARCHAR(100) NOT NULL,
    method            VARCHAR(10) NOT NULL,
    model             VARCHAR(50),
    token_id          BIGINT,
    prompt_tokens     INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens      INTEGER DEFAULT 0,
    status_code       INTEGER DEFAULT 0,
    response_time     INTEGER DEFAULT 0,
    error_message     TEXT,
    request_ip        VARCHAR(50),
    user_agent        VARCHAR(200),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_api_access_user ON api_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_api_access_tenant ON api_access_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_access_token ON api_access_logs(token_id) WHERE token_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_access_created ON api_access_logs(created_at);

-- Channel Test History
CREATE TABLE IF NOT EXISTS channel_test_history (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT NOT NULL DEFAULT 1,
    channel_id       BIGINT NOT NULL,
    user_id          BIGINT,
    test_type        VARCHAR(20) NOT NULL,
    model            VARCHAR(100),
    request_body     TEXT,
    status_code      INTEGER,
    response_body    TEXT,
    response_time_ms INTEGER DEFAULT 0,
    success          BOOLEAN,
    error_message    TEXT,
    error_type       VARCHAR(50),
    request_ip       VARCHAR(50),
    user_agent       VARCHAR(500),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_channel_test_tenant ON channel_test_history(tenant_id);
CREATE INDEX IF NOT EXISTS idx_channel_test_channel ON channel_test_history(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_test_user ON channel_test_history(user_id) WHERE user_id IS NOT NULL;

-- ============================================================
-- System Configuration
-- ============================================================

-- System Configs
CREATE TABLE IF NOT EXISTS system_configs (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT,
    config_key   VARCHAR(100) NOT NULL,
    config_value TEXT,
    value_type   VARCHAR(20) DEFAULT 'string',
    config_group VARCHAR(50) DEFAULT 'general',
    description  VARCHAR(200),
    is_public    BOOLEAN DEFAULT FALSE,
    is_sensitive BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by   BIGINT,
    updated_by   BIGINT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(config_key) WHERE tenant_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_system_configs_tenant ON system_configs(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_system_configs_group ON system_configs(config_group);

-- Signup Configs
CREATE TABLE IF NOT EXISTS signup_configs (
    id                         BIGSERIAL PRIMARY KEY,
    tenant_id                  BIGINT NOT NULL DEFAULT 1,
    enabled                    BOOLEAN DEFAULT TRUE,
    quota_amount              BIGINT DEFAULT 100000,
    quota_type                VARCHAR(10) DEFAULT 'permanent',
    trial_vip_days            INTEGER DEFAULT 0,
    trial_quota               BIGINT DEFAULT 0,
    per_ip_limit              INTEGER DEFAULT 3,
    per_email_verification    BOOLEAN DEFAULT TRUE,
    valid_from                TIMESTAMPTZ,
    valid_until               TIMESTAMPTZ,
    description               VARCHAR(200),
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by                BIGINT
);
CREATE INDEX IF NOT EXISTS idx_signup_configs_tenant ON signup_configs(tenant_id);

-- ============================================================
-- Default Data
-- ============================================================

-- Insert default signup config
INSERT INTO signup_configs (tenant_id, quota_amount, quota_type, trial_quota) 
VALUES (1, 50000, 'free', 50000)
ON CONFLICT DO NOTHING;

-- Insert default system configs
INSERT INTO system_configs (config_key, config_value, value_type, config_group, description)
VALUES 
    ('system_initialized', 'false', 'boolean', 'general', 'System initialization completed'),
    ('allow_register', 'true', 'boolean', 'general', 'Allow new user registration'),
    ('require_email_verify', 'false', 'boolean', 'general', 'Require email verification for signup'),
    ('default_vip_quota', '100000', 'number', 'quota', 'Default VIP quota for new VIP users'),
    ('default_free_quota', '50000', 'number', 'quota', 'Default free quota for new users (50K tokens)'),
    ('free_quota_days', '7', 'number', 'quota', 'Free quota validity period in days'),
    ('vip_discount', '90', 'number', 'vip', 'VIP user recharge discount percentage (90 = 9折)'),
    ('vip_recharge_discount', 'true', 'boolean', 'vip', 'Enable VIP user recharge discount')
ON CONFLICT DO NOTHING;

COMMIT;
