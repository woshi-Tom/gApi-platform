# API Proxy Platform - 数据库设计文档 v2.0

**版本**: 2.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 设计原则

### 1.1 核心原则

| 原则 | 说明 |
|------|------|
| 多租户隔离 | 所有业务表包含 `tenant_id` 字段 |
| 数据分区 | `usage_logs` 和 `api_logs` 按月分区 |
| 审计追溯 | 所有变更记录 `old_value` / `new_value` |
| 软删除优先 | 关键数据使用软删除 (`deleted_at`) |
| 冗余设计 | 审计日志冗余用户名，避免关联查询 |

### 1.2 命名规范

```sql
-- 表名: 小写下划线命名
CREATE TABLE users (...);
CREATE TABLE channel_test_history (...);

-- 字段名: 小写下划线命名
user_id, created_at, api_key_encrypted

-- 索引名: idx_{表名}_{字段名}
CREATE INDEX idx_users_email ON users(email);

-- 外键名: fk_{表名}_{字段名}
```

---

## 2. 完整 DDL 脚本

### 2.1 初始化配置

```sql
-- ============================================================
-- PostgreSQL 数据库初始化
-- ============================================================

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 设置时区
SET timezone = 'Asia/Shanghai';

-- 创建专用schema (可选,用于隔离)
CREATE SCHEMA IF NOT EXISTS api_proxy;
ALTER DATABASE claw_db SET search_path TO public, api_proxy;
```

### 2.2 租户表 (tenants)

```sql
-- ============================================================
-- 租户表
-- ============================================================
CREATE TABLE tenants (
    id              BIGSERIAL PRIMARY KEY,
    
    -- 基础信息
    name            VARCHAR(100) NOT NULL,                 -- 租户名称
    code            VARCHAR(50) UNIQUE NOT NULL,          -- 租户代码 (唯一)
    description     TEXT,                                  -- 描述
    
    -- 配额配置
    max_users       INTEGER DEFAULT 100,                  -- 最大用户数
    max_channels    INTEGER DEFAULT 50,                   -- 最大渠道数
    max_tokens      INTEGER DEFAULT 100,                  -- 最大Token数
    
    -- 功能开关
    features        JSONB DEFAULT '{}',                    -- 功能开关配置
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',          -- active|suspended|deleted
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ                           -- 软删除
);

-- 索引
CREATE INDEX idx_tenants_code ON tenants(code);
CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_created ON tenants(created_at);

COMMENT ON TABLE tenants IS '租户表 - 支持多租户隔离';
COMMENT ON COLUMN tenants.features IS '功能开关: {vip_enabled: true, payment_enabled: false}';
```

### 2.3 管理员表 (admin_users)

```sql
-- ============================================================
-- 管理员表
-- ============================================================
CREATE TABLE admin_users (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),         -- 所属租户 (NULL表示超级管理员)
    
    -- 账号信息
    username        VARCHAR(50) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    email           VARCHAR(100) UNIQUE NOT NULL,
    phone           VARCHAR(20),
    avatar          VARCHAR(500),
    
    -- 角色权限
    role            VARCHAR(20) NOT NULL DEFAULT 'admin', -- super_admin|admin|operator|viewer
    permissions     JSONB DEFAULT '[]',                   -- 细粒度权限
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',          -- active|disabled
    last_login_at   TIMESTAMPTZ,
    last_login_ip   VARCHAR(50),
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_admin_tenant ON admin_users(tenant_id);
CREATE INDEX idx_admin_username ON admin_users(username);
CREATE INDEX idx_admin_email ON admin_users(email);
CREATE INDEX idx_admin_role ON admin_users(role);

COMMENT ON TABLE admin_users IS '管理员表 - 支持多租户管理员和超级管理员';
COMMENT ON COLUMN admin_users.role IS '角色: super_admin(超级管理员)|admin(管理员)|operator(运维)|viewer(只读)';
COMMENT ON COLUMN admin_users.permissions IS '细粒度权限: [channel.create, channel.delete, user.manage, ...]';
```

### 2.4 用户表 (users)

```sql
-- ============================================================
-- 用户表
-- ============================================================
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 账号信息
    username        VARCHAR(50) NOT NULL,
    email           VARCHAR(100) NOT NULL,
    phone           VARCHAR(20),
    password_hash   VARCHAR(255) NOT NULL,
    
    -- 认证信息
    email_verified  BOOLEAN DEFAULT FALSE,
    verify_token    VARCHAR(100),
    verify_expired  TIMESTAMPTZ,
    
    -- 用户等级
    level           VARCHAR(20) DEFAULT 'free',           -- free|premium|vip|enterprise
    vip_expired_at  TIMESTAMPTZ,                          -- VIP到期时间
    vip_package_id  BIGINT,                               -- 当前VIP套餐
    
    -- 配额信息
    remain_quota    BIGINT DEFAULT 0,                     -- 永久配额 (tokens)
    vip_quota       BIGINT DEFAULT 0,                     -- VIP配额 (tokens)
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',          -- active|disabled|suspended
    disabled_reason VARCHAR(200),
    last_login_at   TIMESTAMPTZ,
    last_login_ip   VARCHAR(50),
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    
    -- 唯一约束
    CONSTRAINT uk_users_email_tenant UNIQUE (tenant_id, email),
    CONSTRAINT uk_users_username_tenant UNIQUE (tenant_id, username)
);

-- 索引
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_level ON users(level);
CREATE INDEX idx_users_vip_expire ON users(vip_expired_at) WHERE vip_expired_at IS NOT NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created ON users(created_at DESC);

COMMENT ON TABLE users IS '用户表 - 存储最终用户信息';
COMMENT ON COLUMN users.level IS '用户等级: free(免费)|premium(付费)|vip(企业)';
COMMENT ON COLUMN users.remain_quota IS '永久配额 - 永不过期';
COMMENT ON COLUMN users.vip_quota IS 'VIP配额 - 30天过期';
```

### 2.5 渠道表 (channels)

```sql
-- ============================================================
-- 渠道表
-- ============================================================
CREATE TABLE channels (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 基础信息
    name            VARCHAR(100) NOT NULL,
    type            VARCHAR(50) NOT NULL,                 -- openai|azure|claude|gemini|anthropic|custom
    base_url        VARCHAR(500) NOT NULL,
    
    -- 认证
    api_key_encrypted   VARCHAR(500) NOT NULL,            -- AES-256-GCM 加密
    key_version     INTEGER DEFAULT 1,                    -- 密钥版本 (用于密钥轮换)
    
    -- 模型配置
    models          JSONB DEFAULT '[]',                    -- ["gpt-4", "gpt-3.5-turbo"]
    model_mapping   JSONB DEFAULT '{}',                   -- {"gpt-4": "gpt-4-0613"}
    
    -- 负载均衡
    weight          INTEGER DEFAULT 100,                   -- 权重 (1-1000)
    priority        INTEGER DEFAULT 0,                     -- 优先级 (数值越大越优先)
    
    -- 速率限制
    rpm_limit       INTEGER DEFAULT 1000,                  -- 每分钟请求限制
    tpm_limit       INTEGER DEFAULT 100000,               -- 每分钟Token限制
    
    -- 成本配置
    cost_factor     DECIMAL(10,4) DEFAULT 1.0,            -- 成本系数
    price_per_1k_input  DECIMAL(10,4) DEFAULT 0.01,     -- 每1K输入Token价格
    price_per_1k_output DECIMAL(10,4) DEFAULT 0.03,      -- 每1K输出Token价格
    
    -- 分组
    group_name      VARCHAR(50) DEFAULT 'default',        -- 渠道分组
    
    -- 状态
    status          INTEGER DEFAULT 1,                     -- 0:禁用 1:启用 2:维护中
    is_healthy      BOOLEAN DEFAULT TRUE,
    failure_count   INTEGER DEFAULT 0,
    last_success_at TIMESTAMPTZ,
    last_check_at   TIMESTAMPTZ,
    last_error      TEXT,
    response_time_avg INTEGER DEFAULT 0,                  -- 平均响应时间 (ms)
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_channels_tenant ON channels(tenant_id);
CREATE INDEX idx_channels_type ON channels(type);
CREATE INDEX idx_channels_status ON channels(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_channels_group ON channels(group_name) WHERE deleted_at IS NULL;
CREATE INDEX idx_channels_healthy ON channels(is_healthy, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_channels_created ON channels(created_at DESC);

COMMENT ON TABLE channels IS '渠道表 - 存储API渠道配置';
COMMENT ON COLUMN channels.type IS '渠道类型: openai|azure|claude|gemini|anthropic|ollama|localai|custom';
COMMENT ON COLUMN channels.api_key_encrypted IS 'API Key加密存储(AES-256-GCM)';
COMMENT ON COLUMN channels.status IS '状态: 0(禁用)|1(启用)|2(维护中)';
```

### 2.6 能力表 (abilities)

```sql
-- ============================================================
-- 能力表 - 定义渠道支持的API能力
-- ============================================================
CREATE TABLE abilities (
    id              BIGSERIAL PRIMARY KEY,
    channel_id      BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    
    -- 能力类型
    ability_type    VARCHAR(50) NOT NULL,                  -- chat|completion|embedding|moderation|audio|vision
    
    -- 模型支持
    model           VARCHAR(100) NOT NULL,                 -- 具体模型名称
    model_alias     VARCHAR(100),                          -- 模型别名/映射
    
    -- 能力配置
    config          JSONB DEFAULT '{}',                    -- 能力特定配置
    is_enabled      BOOLEAN DEFAULT TRUE,
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_abilities_channel_model UNIQUE (channel_id, model)
);

-- 索引
CREATE INDEX idx_abilities_channel ON abilities(channel_id);
CREATE INDEX idx_abilities_type ON abilities(ability_type);

COMMENT ON TABLE abilities IS '能力表 - 定义渠道支持的模型和能力';
COMMENT ON COLUMN abilities.ability_type IS '能力类型: chat(对话补全)|completion(文本补全)|embedding(向量)|moderation(审核)|audio(语音)|vision(视觉)';
```

### 2.7 Token表 (tokens)

```sql
-- ============================================================
-- Token表 - 用户API Key
-- ============================================================
CREATE TABLE tokens (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    
    -- Token信息
    name            VARCHAR(100) NOT NULL,
    token_key       VARCHAR(64) NOT NULL UNIQUE,          -- sk-xxx格式的Key
    token_hash      VARCHAR(64) NOT NULL,                  -- SHA-256哈希值
    key_prefix      VARCHAR(10) DEFAULT 'sk-ap-',        -- Key前缀
    
    -- 配额限制
    remain_quota    BIGINT DEFAULT 0,                      -- 剩余配额
    is_vip_quota    BOOLEAN DEFAULT FALSE,                 -- 是否使用VIP配额
    
    -- 访问控制
    allowed_models  JSONB DEFAULT '[]',                     -- 允许的模型列表 []表示全部
    denied_models   JSONB DEFAULT '[]',                    -- 拒绝的模型列表
    allowed_ips     JSONB DEFAULT '[]',                     -- IP白名单 []表示不限制
    
    -- 速率限制 (覆盖全局)
    rpm_limit       INTEGER,                               -- 每分钟请求限制 (NULL继承全局)
    tpm_limit       INTEGER,                               -- 每分钟Token限制 (NULL继承全局)
    
    -- 使用限制
    max_usage_per_day BIGINT,                              -- 每日最大用量 (NULL不限制)
    expires_at      TIMESTAMPTZ,                           -- Token过期时间 (NULL永不过期)
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',            -- active|disabled|expired
    used_quota      BIGINT DEFAULT 0,                      -- 已使用配额
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ,
    last_used_ip    VARCHAR(50),
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_tokens_tenant ON tokens(tenant_id);
CREATE INDEX idx_tokens_user ON tokens(user_id);
CREATE INDEX idx_tokens_key ON tokens(token_key);
CREATE INDEX idx_tokens_status ON tokens(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tokens_expire ON tokens(expires_at) WHERE expires_at IS NOT NULL;

COMMENT ON TABLE tokens IS 'Token表 - 用户API Key管理';
COMMENT ON COLUMN tokens.token_key IS '原始Key,仅创建时显示一次,之后不可查看';
COMMENT ON COLUMN tokens.token_hash IS 'SHA-256哈希值,用于验证';
```

### 2.8 VIP套餐表 (vip_packages)

```sql
-- ============================================================
-- VIP套餐表
-- ============================================================
CREATE TABLE vip_packages (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 套餐信息
    name            VARCHAR(100) NOT NULL,                 -- 套餐名称
    description     TEXT,                                  -- 套餐描述
    price           DECIMAL(10,2) NOT NULL,                -- 价格 (元)
    original_price  DECIMAL(10,2),                          -- 原价
    
    -- 有效期
    duration_days   INTEGER NOT NULL DEFAULT 30,            -- 有效期(天)
    
    -- 权益配置
    quota           BIGINT NOT NULL DEFAULT 1000000,        -- 配额 (tokens)
    rpm_limit       INTEGER DEFAULT 2000,                   -- 每分钟请求
    tpm_limit       INTEGER DEFAULT 100000,                 -- 每分钟Tokens
    concurrent_limit INTEGER DEFAULT 10,                    -- 并发限制
    
    -- 功能开关
    features        JSONB DEFAULT '{}',                     -- 额外功能
    
    -- 显示配置
    sort_order      INTEGER DEFAULT 0,                     -- 排序
    is_recommended  BOOLEAN DEFAULT FALSE,                 -- 是否推荐
    is_popular      BOOLEAN DEFAULT FALSE,                 -- 是否热门
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',            -- active|disabled|deleted
    is_visible      BOOLEAN DEFAULT TRUE,                   -- 是否显示
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_vip_tenant ON vip_packages(tenant_id);
CREATE INDEX idx_vip_status ON vip_packages(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_vip_sort ON vip_packages(sort_order);

COMMENT ON TABLE vip_packages IS 'VIP套餐表';
COMMENT ON COLUMN vip_packages.duration_days IS '有效期天数,默认30天';
```

### 2.9 充值套餐表 (recharge_packages)

```sql
-- ============================================================
-- 充值套餐表 - 永久配额充值
-- ============================================================
CREATE TABLE recharge_packages (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 套餐信息
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    price           DECIMAL(10,2) NOT NULL,
    original_price  DECIMAL(10,2),
    
    -- 配额
    quota           BIGINT NOT NULL,                       -- 配额 (tokens)
    bonus_quota     BIGINT DEFAULT 0,                      -- 赠送配额
    
    -- 显示配置
    sort_order      INTEGER DEFAULT 0,
    is_recommended  BOOLEAN DEFAULT FALSE,
    is_popular      BOOLEAN DEFAULT FALSE,
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',
    is_visible      BOOLEAN DEFAULT TRUE,
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_recharge_tenant ON recharge_packages(tenant_id);
CREATE INDEX idx_recharge_status ON recharge_packages(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE recharge_packages IS '充值套餐表 - 永久配额';
```

### 2.10 订单表 (orders)

```sql
-- ============================================================
-- 订单表
-- ============================================================
CREATE TABLE orders (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    
    -- 订单信息
    order_no        VARCHAR(50) UNIQUE NOT NULL,           -- 订单号
    order_type      VARCHAR(20) NOT NULL,                  -- recharge|vip|package
    
    -- 关联信息
    package_id      BIGINT,                               -- 套餐ID
    package_name    VARCHAR(100),                          -- 套餐名称冗余
    
    -- 金额
    total_amount    DECIMAL(10,2) NOT NULL,                -- 总金额
    discount_amount DECIMAL(10,2) DEFAULT 0,               -- 优惠金额
    pay_amount      DECIMAL(10,2) NOT NULL,               -- 实付金额
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'pending',         -- pending|paid|cancelled|refunded|expired
    paid_at         TIMESTAMPTZ,                           -- 支付时间
    cancel_reason   VARCHAR(200),
    refund_reason   TEXT,
    refund_amount   DECIMAL(10,2),
    
    -- 过期时间
    expire_at       TIMESTAMPTZ,                           -- 订单过期时间
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_orders_tenant ON orders(tenant_id);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_no ON orders(order_no);
CREATE INDEX idx_orders_type ON orders(order_type);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_expire ON orders(expire_at) WHERE status = 'pending';
CREATE INDEX idx_orders_created ON orders(created_at DESC);

COMMENT ON TABLE orders IS '订单表';
COMMENT ON COLUMN orders.order_type IS '订单类型: recharge(充值)|vip(VIP购买)|package(套餐)';
```

### 2.11 支付表 (payments)

```sql
-- ============================================================
-- 支付表
-- ============================================================
CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    order_id        BIGINT NOT NULL REFERENCES orders(id),
    
    -- 支付信息
    payment_no      VARCHAR(100) UNIQUE NOT NULL,          -- 支付流水号
    payment_method  VARCHAR(20) NOT NULL,                  -- alipay|wechat|bank
    
    -- 金额
    amount          DECIMAL(10,2) NOT NULL,                -- 支付金额
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'pending',         -- pending|success|failed|refunded
    paid_at         TIMESTAMPTZ,
    
    -- 第三方支付信息
    channel_order_no VARCHAR(100),                         -- 支付宝/微信订单号
    channel_trade_no VARCHAR(100),                         -- 第三方交易号
    payment_url     TEXT,                                  -- 支付链接/二维码
    qr_code         TEXT,                                  -- 二维码Base64
    
    -- 回调信息
    callback_url    VARCHAR(500),                          -- 回调URL
    callback_body   TEXT,                                  -- 回调原始数据
    callback_at     TIMESTAMPTZ,
    
    -- 错误信息
    error_code      VARCHAR(50),
    error_message   TEXT,
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_payments_tenant ON payments(tenant_id);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_payments_no ON payments(payment_no);
CREATE INDEX idx_payments_method ON payments(payment_method);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created ON payments(created_at DESC);

COMMENT ON TABLE payments IS '支付表 - 存储支付流水';
```

### 2.12 配额流水表 (quota_transactions)

```sql
-- ============================================================
-- 配额流水表
-- ============================================================
CREATE TABLE quota_transactions (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    token_id        BIGINT REFERENCES tokens(id),
    
    -- 交易类型
    type            VARCHAR(20) NOT NULL,                   -- recharge|purchase|vip_grant|usage|refund|adjust|expire
    quota_type      VARCHAR(10) NOT NULL,                  -- permanent|vip
    
    -- 变动金额
    change_amount   BIGINT NOT NULL,                       -- 变动数量 (正数增加,负数减少)
    balance_before  BIGINT NOT NULL,                       -- 变动前余额
    balance_after   BIGINT NOT NULL,                       -- 变动后余额
    
    -- 关联信息
    order_id        BIGINT,
    package_id      BIGINT,
    channel_id      BIGINT,                                -- 用量来源
    model           VARCHAR(100),                          -- 使用模型
    
    -- 描述
    description     VARCHAR(500),
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_quota_tenant ON quota_transactions(tenant_id);
CREATE INDEX idx_quota_user ON quota_transactions(user_id);
CREATE INDEX idx_quota_type ON quota_transactions(type);
CREATE INDEX idx_quota_created ON quota_transactions(created_at DESC);
CREATE INDEX idx_quota_user_time ON quota_transactions(user_id, created_at DESC);

COMMENT ON TABLE quota_transactions IS '配额流水表 - 记录所有配额变动';
```

### 2.13 兑换码表 (redemption_codes)

```sql
-- ============================================================
-- 兑换码表
-- ============================================================
CREATE TABLE redemption_codes (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 兑换码信息
    code            VARCHAR(50) UNIQUE NOT NULL,
    code_type       VARCHAR(20) NOT NULL,                  -- recharge|vip|quota
    
    -- 奖励内容
    quota           BIGINT DEFAULT 0,                       -- 配额奖励
    quota_type      VARCHAR(10) DEFAULT 'permanent',       -- permanent|vip
    vip_days        INTEGER DEFAULT 0,                     -- VIP天数
    is_permanent    BOOLEAN DEFAULT FALSE,                 -- 是否永久VIP
    
    -- 使用限制
    max_uses        INTEGER DEFAULT 1,                     -- 最大使用次数
    used_count      INTEGER DEFAULT 0,                     -- 已使用次数
    valid_from      TIMESTAMPTZ,                          -- 有效期开始
    valid_until     TIMESTAMPTZ,                          -- 有效期结束
    
    -- 绑定信息
    bound_user_id   BIGINT REFERENCES users(id),          -- 绑定的用户
    bound_at        TIMESTAMPTZ,
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',          -- active|disabled|expired|used
    
    -- 创建信息
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at         TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_redemption_tenant ON redemption_codes(tenant_id);
CREATE INDEX idx_redemption_code ON redemption_codes(code);
CREATE INDEX idx_redemption_status ON redemption_codes(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_redemption_valid ON redemption_codes(valid_from, valid_until) WHERE status = 'active';

COMMENT ON TABLE redemption_codes IS '兑换码表';
```

### 2.14 渠道测试历史表 (channel_test_history)

```sql
-- ============================================================
-- 渠道测试历史表
-- ============================================================
CREATE TABLE channel_test_history (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    channel_id      BIGINT NOT NULL REFERENCES channels(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    
    -- 测试信息
    test_type       VARCHAR(20) NOT NULL,                 -- models|chat|embeddings
    model           VARCHAR(100),
    
    -- 请求
    request_body    TEXT,                                  -- 请求体 (JSON)
    
    -- 响应
    status_code     INTEGER,
    response_body   TEXT,                                  -- 响应体 (JSON,截断)
    response_time_ms INTEGER NOT NULL DEFAULT 0,
    
    -- 结果
    success         BOOLEAN NOT NULL,
    error_message   TEXT,
    error_type      VARCHAR(50),
    
    -- 环境
    request_ip      VARCHAR(50),
    user_agent      VARCHAR(500),
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_test_tenant ON channel_test_history(tenant_id);
CREATE INDEX idx_test_channel ON channel_test_history(channel_id);
CREATE INDEX idx_test_user ON channel_test_history(user_id);
CREATE INDEX idx_test_type ON channel_test_history(test_type);
CREATE INDEX idx_test_created ON channel_test_history(created_at DESC);
CREATE INDEX idx_test_channel_time ON channel_test_history(channel_id, created_at DESC);

COMMENT ON TABLE channel_test_history IS '渠道测试历史表';
```

### 2.15 审计日志表 (audit_logs)

```sql
-- ============================================================
-- 审计日志表 (核心表 - 用于安全审查和溯源)
-- ============================================================
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT,                                -- 租户ID (可为空表示系统操作)
    user_id         BIGINT,                                -- 操作用户ID
    username        VARCHAR(100),                          -- 操作时用户名 (冗余存储)
    
    -- 操作信息
    action          VARCHAR(100) NOT NULL,                  -- 操作类型
    action_group    VARCHAR(50) NOT NULL,                  -- 操作分组
    resource_type   VARCHAR(50),                           -- 资源类型
    resource_id     BIGINT,                                -- 资源ID
    
    -- 请求信息
    request_method  VARCHAR(10),
    request_path    VARCHAR(500),
    request_body    TEXT,
    request_ip      VARCHAR(50),
    request_ua      VARCHAR(500),
    
    -- 响应信息
    status_code     INTEGER,
    response_body   TEXT,
    
    -- 结果
    success         BOOLEAN NOT NULL DEFAULT true,
    error_message   TEXT,
    
    -- 变更详情 (JSON)
    old_value       JSONB,
    new_value       JSONB,
    
    -- 元数据
    user_agent      VARCHAR(500),
    session_id      VARCHAR(100),
    trace_id        VARCHAR(64),
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_audit_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_action_group ON audit_logs(action_group);
CREATE INDEX idx_audit_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_ip ON audit_logs(request_ip) WHERE request_ip IS NOT NULL;
CREATE INDEX idx_audit_success ON audit_logs(success) WHERE success = false;

COMMENT ON TABLE audit_logs IS '审计日志表 - 记录所有操作行为';
```

### 2.16 登录日志表 (login_logs)

```sql
-- ============================================================
-- 登录日志表
-- ============================================================
CREATE TABLE login_logs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT,
    user_id         BIGINT,
    username        VARCHAR(100),
    login_type      VARCHAR(20) NOT NULL,                 -- user|admin
    
    -- 登录信息
    ip              VARCHAR(50),
    ip_location     VARCHAR(200),                          -- IP归属地
    user_agent      VARCHAR(500),
    device_type     VARCHAR(50),                          -- web|mobile|desktop
    
    -- 结果
    success         BOOLEAN NOT NULL,
    fail_reason     VARCHAR(100),
    
    -- Token信息
    token           VARCHAR(200),
    token_expired_at TIMESTAMPTZ,
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_login_tenant ON login_logs(tenant_id);
CREATE INDEX idx_login_user ON login_logs(user_id, created_at DESC);
CREATE INDEX idx_login_ip ON login_logs(ip, created_at DESC);
CREATE INDEX idx_login_type ON login_logs(login_type);
CREATE INDEX idx_login_created ON login_logs(created_at DESC);

COMMENT ON TABLE login_logs IS '登录日志表';
```

### 2.17 用量日志表 (usage_logs) - 分区表

```sql
-- ============================================================
-- 用量日志表 (按月分区)
-- ============================================================
CREATE TABLE usage_logs (
    id              BIGSERIAL,
    tenant_id       BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    token_id        BIGINT,
    channel_id      BIGINT,
    
    -- 请求信息
    request_id      VARCHAR(64),                            -- 请求ID (幂等)
    model           VARCHAR(100) NOT NULL,
    
    -- Token用量
    prompt_tokens   INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens    INTEGER DEFAULT 0,
    
    -- 成本
    cost            DECIMAL(10,4) DEFAULT 0,
    
    -- 响应
    status_code     INTEGER,
    response_time_ms INTEGER DEFAULT 0,
    
    -- 错误
    error_message   TEXT,
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 创建月度分区
CREATE TABLE usage_logs_2026_01 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE usage_logs_2026_02 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE usage_logs_2026_03 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE usage_logs_2026_04 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE usage_logs_2026_05 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE usage_logs_2026_06 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE usage_logs_2026_07 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE usage_logs_2026_08 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE usage_logs_2026_09 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE usage_logs_2026_10 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE usage_logs_2026_11 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE usage_logs_2026_12 PARTITION OF usage_logs
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- 分区索引
CREATE INDEX idx_usage_tenant ON usage_logs(tenant_id);
CREATE INDEX idx_usage_user ON usage_logs(user_id);
CREATE INDEX idx_usage_token ON usage_logs(token_id);
CREATE INDEX idx_usage_channel ON usage_logs(channel_id);
CREATE INDEX idx_usage_model ON usage_logs(model);
CREATE INDEX idx_usage_user_time ON usage_logs(user_id, created_at DESC);
CREATE INDEX idx_usage_created ON usage_logs(created_at DESC);

COMMENT ON TABLE usage_logs IS '用量日志表 - 按月分区,保留12个月';
```

### 2.18 系统配置表 (system_configs)

```sql
-- ============================================================
-- 系统配置表
-- ============================================================
CREATE TABLE system_configs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),         -- NULL表示全局配置
    
    -- 配置信息
    config_key      VARCHAR(100) NOT NULL,
    config_value    TEXT,
    value_type      VARCHAR(20) DEFAULT 'string',         -- string|number|boolean|json
    
    -- 配置分组
    config_group    VARCHAR(50) DEFAULT 'general',       -- general|payment|email|sms|oauth
    
    -- 描述
    description     VARCHAR(200),
    
    -- 状态
    is_public       BOOLEAN DEFAULT FALSE,                -- 是否公开(可被用户查看)
    is_sensitive    BOOLEAN DEFAULT FALSE,               -- 是否敏感(需要权限)
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    updated_by      BIGINT,
    
    CONSTRAINT uk_config_key_tenant UNIQUE (tenant_id, config_key)
);

-- 索引
CREATE INDEX idx_config_tenant ON system_configs(tenant_id);
CREATE INDEX idx_config_group ON system_configs(config_group);
CREATE INDEX idx_config_key ON system_configs(config_key);

COMMENT ON TABLE system_configs IS '系统配置表 - Key-Value配置存储';
```

---

## 3. 操作类型定义 (Go代码)

```go
// internal/model/audit.go

// ActionGroup: 操作分组
const (
    AuditGroupAuth       = "auth"        // 认证相关
    AuditGroupUser       = "user"        // 用户管理
    AuditGroupChannel    = "channel"      // 渠道管理
    AuditGroupToken      = "token"       // Token管理
    AuditGroupOrder      = "order"       // 订单相关
    AuditGroupPayment    = "payment"      // 支付相关
    AuditGroupQuota      = "quota"       // 配额相关
    AuditGroupVIP        = "vip"         // VIP相关
    AuditGroupSystem     = "system"      // 系统操作
)

// Action: 详细操作
const (
    // 认证 (auth)
    AuditActionUserLogin          = "user.login"           // 用户登录
    AuditActionUserLogout         = "user.logout"         // 用户登出
    AuditActionUserRegister        = "user.register"       // 用户注册
    AuditActionPasswordChange      = "user.password_change" // 密码修改
    AuditActionAdminLogin          = "admin.login"         // 管理员登录
    
    // 用户 (user)
    AuditActionUserCreate          = "user.create"         // 创建用户
    AuditActionUserUpdate         = "user.update"         // 更新用户
    AuditActionUserDelete         = "user.delete"         // 删除用户
    AuditActionUserEnable         = "user.enable"         // 启用用户
    AuditActionUserDisable        = "user.disable"        // 禁用用户
    AuditActionUserQuotaAdd      = "user.quota_add"      // 用户充值
    AuditActionUserQuotaDeduct    = "user.quota_deduct"   // 配额扣除
    
    // 渠道 (channel)
    AuditActionChannelCreate       = "channel.create"     // 创建渠道
    AuditActionChannelUpdate       = "channel.update"      // 更新渠道
    AuditActionChannelDelete       = "channel.delete"      // 删除渠道
    AuditActionChannelEnable       = "channel.enable"      // 启用渠道
    AuditActionChannelDisable      = "channel.disable"      // 禁用渠道
    AuditActionChannelTest         = "channel.test"        // 渠道测试
    
    // Token
    AuditActionTokenCreate         = "token.create"       // 创建Token
    AuditActionTokenUpdate         = "token.update"       // 更新Token
    AuditActionTokenDelete         = "token.delete"       // 删除Token
    AuditActionTokenResetQuota     = "token.reset_quota"  // 重置配额
    
    // 订单 (order)
    AuditActionOrderCreate         = "order.create"       // 创建订单
    AuditActionOrderPaid           = "order.paid"         // 订单支付
    AuditActionOrderCancelled      = "order.cancelled"    // 订单取消
    AuditActionOrderRefunded       = "order.refunded"     // 订单退款
    
    // 支付 (payment)
    AuditActionPaymentInit         = "payment.init"        // 支付发起
    AuditActionPaymentSuccess      = "payment.success"     // 支付成功
    AuditActionPaymentFailed       = "payment.failed"     // 支付失败
    AuditActionPaymentCallback     = "payment.callback"   // 支付回调
    
    // VIP
    AuditActionVIPActivate         = "vip.activate"       // VIP开通
    AuditActionVIPExpired          = "vip.expired"        // VIP过期
    AuditActionVIPCancelled       = "vip.cancelled"      // VIP取消
    
    // 兑换码
    AuditActionRedemptionCreate    = "redemption.create"  // 创建兑换码
    AuditActionRedemptionUse      = "redemption.use"      // 使用兑换码
)
```

---

## 4. 分区管理

### 4.1 自动创建分区存储过程

```sql
-- 创建未来月份的分区
CREATE OR REPLACE FUNCTION create_usage_partition()
RETURNS void AS $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- 创建下个月的分区
    partition_date := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    partition_name := 'usage_logs_' || TO_CHAR(partition_date, 'YYYY_MM');
    start_date := partition_date;
    end_date := partition_date + INTERVAL '1 month';
    
    -- 检查分区是否已存在
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = partition_name
    ) THEN
        EXECUTE FORMAT(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF usage_logs FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
        RAISE NOTICE 'Created partition: %', partition_name;
    END IF;
END;
$$ LANGUAGE plpgsql;
```

### 4.2 分区清理策略

```sql
-- 清理12个月前的分区 (保留最近12个月)
CREATE OR REPLACE FUNCTION cleanup_old_partitions()
RETURNS void AS $$
DECLARE
    partition_record RECORD;
    cutoff_date DATE;
BEGIN
    cutoff_date := DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months');
    
    FOR partition_record IN
        SELECT c.relname as partition_name
        FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        JOIN pg_class p ON p.oid = i.inhparent
        WHERE p.relname = 'usage_logs'
        AND c.relname < 'usage_logs_' || TO_CHAR(cutoff_date, 'YYYY_MM')
    LOOP
        RAISE NOTICE 'Detaching old partition: %', partition_record.partition_name;
        EXECUTE FORMAT('ALTER TABLE usage_logs DETACH PARTITION %I', partition_record.partition_name);
        EXECUTE FORMAT('DROP TABLE IF EXISTS %I', partition_record.partition_name);
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

---

## 5. 初始化数据

```sql
-- ============================================================
-- 初始化数据
-- ============================================================

-- 创建默认租户
INSERT INTO tenants (name, code, description) VALUES 
('Default Tenant', 'default', 'System default tenant');

-- 创建超级管理员
INSERT INTO admin_users (username, password_hash, email, role, status) VALUES 
('admin', '$2a$10$...', 'admin@example.com', 'super_admin', 'active');

-- 创建VIP套餐
INSERT INTO vip_packages (name, description, price, duration_days, quota, rpm_limit, tpm_limit, sort_order, is_recommended) VALUES 
('VIP Basic', '基础VIP套餐', 9.9, 30, 1000000, 1000, 50000, 1, false),
('VIP Pro', '专业VIP套餐', 29.9, 30, 5000000, 2000, 100000, 2, true),
('VIP Enterprise', '企业VIP套餐', 99.9, 30, 20000000, 5000, 500000, 3, false);

-- 创建充值套餐
INSERT INTO recharge_packages (name, description, price, quota, bonus_quota, sort_order) VALUES 
('Starter', '入门套餐', 1.0, 100000, 0, 1),
('Basic', '基础套餐', 10.0, 1000000, 100000, 2),
('Pro', '专业套餐', 50.0, 5000000, 1000000, 3),
('Enterprise', '企业套餐', 100.0, 10000000, 3000000, 4);

-- 系统配置
INSERT INTO system_configs (config_key, config_value, config_group, description, is_public) VALUES 
('site_name', 'API Proxy Platform', 'general', '网站名称', true),
('site_logo', '/static/logo.png', 'general', '网站Logo', true),
('alipay_enabled', 'false', 'payment', '启用支付宝', false),
('wechat_enabled', 'false', 'payment', '启用微信支付', false),
('rate_limit_free_rpm', '60', 'rate_limit', '免费用户RPM限制', false),
('rate_limit_free_tpm', '10000', 'rate_limit', '免费用户TPM限制', false);
```

---

**文档版本**: 2.0  
**下一步**: API接口设计、项目结构设计
