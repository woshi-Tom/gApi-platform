# 模型分组、定价与权限系统设计

> 版本: v1.0  
> 日期: 2026-05-06  
> 状态: 📝 设计中  
> 参考: NewAPI/OneAPI 模型分组与权限体系

---

## 1. 设计目标

借鉴 NewAPI 的模型分组体系，实现：

1. **模型分组管理** - 将模型按能力/等级分组（如 default、pro、vip）
2. **用户分组授权** - 不同用户组可访问不同模型组
3. **统一模型定价** - 集中管理各模型的 input/output token 价格
4. **按权限过滤** - `/v1/models` 根据用户/Token 权限返回可用模型
5. **渠道-分组多对多** - 一个渠道可为多个分组提供服务

---

## 2. 数据库设计

### 2.1 模型分组表 (model_groups)

```sql
-- ============================================================
-- 模型分组表
-- ============================================================
CREATE TABLE model_groups (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),

    -- 分组信息
    name            VARCHAR(50) NOT NULL UNIQUE,            -- 分组名称: default, pro, vip
    display_name    VARCHAR(100) NOT NULL,                  -- 显示名称: 默认组, 专业组, VIP组
    description     TEXT,                                   -- 分组描述

    -- 排序与状态
    sort_order      INTEGER DEFAULT 0,
    status          VARCHAR(20) DEFAULT 'active',           -- active|disabled

    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_model_groups_tenant ON model_groups(tenant_id);
CREATE INDEX idx_model_groups_status ON model_groups(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE model_groups IS '模型分组表 - 定义模型访问分组';
COMMENT ON COLUMN model_groups.name IS '分组标识: default(默认)|pro(专业)|vip(VIP专属)';
```

### 2.2 渠道-分组关联表 (channel_group_relations)

```sql
-- ============================================================
-- 渠道-分组关联表 (多对多)
-- ============================================================
CREATE TABLE channel_group_relations (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    channel_id      BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    group_id        BIGINT NOT NULL REFERENCES model_groups(id) ON DELETE CASCADE,

    -- 渠道在此分组中的状态
    is_enabled      BOOLEAN DEFAULT TRUE,                   -- 是否在此分组中启用
    priority        INTEGER DEFAULT 0,                      -- 在此分组中的优先级

    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uk_channel_group UNIQUE (channel_id, group_id)
);

CREATE INDEX idx_cgr_channel ON channel_group_relations(channel_id);
CREATE INDEX idx_cgr_group ON channel_group_relations(group_id);
CREATE INDEX idx_cgr_enabled ON channel_group_relations(group_id, is_enabled) WHERE is_enabled = true;

COMMENT ON TABLE channel_group_relations IS '渠道-分组多对多关联';
COMMENT ON COLUMN channel_group_relations.is_enabled IS '渠道在此分组中是否启用';
```

### 2.3 用户-分组关联表 (user_group_relations)

```sql
-- ============================================================
-- 用户-分组关联表
-- ============================================================
CREATE TABLE user_group_relations (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id        BIGINT NOT NULL REFERENCES model_groups(id) ON DELETE CASCADE,

    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,

    CONSTRAINT uk_user_group UNIQUE (user_id, group_id)
);

CREATE INDEX idx_ugr_user ON user_group_relations(user_id);
CREATE INDEX idx_ugr_group ON user_group_relations(group_id);

COMMENT ON TABLE user_group_relations IS '用户-分组关联表 - 决定用户可访问哪些模型组';
```

### 2.4 模型定价表 (model_pricing)

```sql
-- ============================================================
-- 模型定价表
-- ============================================================
CREATE TABLE model_pricing (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),

    -- 模型信息
    model           VARCHAR(100) NOT NULL,                  -- 模型标识: gpt-4, gpt-3.5-turbo
    provider        VARCHAR(50),                            -- 提供方: openai, anthropic, deepseek
    display_name    VARCHAR(100),                           -- 显示名称: GPT-4

    -- 定价 (每 1K tokens)
    price_input     DECIMAL(10,6) NOT NULL DEFAULT 0.03,    -- 输入 token 价格
    price_output    DECIMAL(10,6) NOT NULL DEFAULT 0.06,    -- 输出 token 价格

    -- 模型能力
    ability_types   JSONB DEFAULT '["chat"]',               -- ["chat","completion","embedding","vision"]
    context_length  INTEGER,                                -- 最大上下文长度 (tokens)
    max_output      INTEGER,                                -- 最大输出长度

    -- 分组关联
    group_id        BIGINT REFERENCES model_groups(id),     -- 默认所属分组

    -- 状态
    is_enabled      BOOLEAN DEFAULT TRUE,
    is_featured     BOOLEAN DEFAULT FALSE,                  -- 是否推荐

    -- 排序与审计
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uk_model_pricing_model UNIQUE (model, deleted_at)
);

CREATE INDEX idx_model_pricing_tenant ON model_pricing(tenant_id);
CREATE INDEX idx_model_pricing_enabled ON model_pricing(is_enabled) WHERE deleted_at IS NULL AND is_enabled = true;
CREATE INDEX idx_model_pricing_group ON model_pricing(group_id);

COMMENT ON TABLE model_pricing IS '模型定价表 - 统一管理模型价格和元数据';
COMMENT ON COLUMN model_pricing.price_input IS '每 1K 输入 token 的价格 (元)';
COMMENT ON COLUMN model_pricing.price_output IS '每 1K 输出 token 的价格 (元)';
```

### 2.5 Token-分组关联表 (token_group_relations) [可选]

如果 Token 需要独立于用户设置分组权限：

```sql
-- ============================================================
-- Token-分组关联表 (可选)
-- ============================================================
CREATE TABLE token_group_relations (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    token_id        BIGINT NOT NULL REFERENCES tokens(id) ON DELETE CASCADE,
    group_id        BIGINT NOT NULL REFERENCES model_groups(id) ON DELETE CASCADE,

    CONSTRAINT uk_token_group UNIQUE (token_id, group_id)
);

CREATE INDEX idx_tgr_token ON token_group_relations(token_id);
CREATE INDEX idx_tgr_group ON token_group_relations(group_id);

COMMENT ON TABLE token_group_relations IS 'Token-分组关联表 - Token 级别覆盖用户分组';
```

---

## 3. 现有表变更

### 3.1 channels 表 (保持现有，新增逻辑)

```sql
-- channels.group_name 保留为向后兼容，但新的分组逻辑使用 channel_group_relations
-- 可选：添加迁移脚本将现有 group_name 迁移到 channel_group_relations
```

### 3.2 users 表 (无需变更)

- 用户分组通过 `user_group_relations` 表管理
- 保留 `users.level` 作为兼容，可同时存在

### 3.3 usage_logs 表 (新增关联)

```sql
-- 在用量记录中记录模型分组信息，方便按组统计
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS group_id BIGINT REFERENCES model_groups(id);
```

---

## 4. 数据流与权限模型

### 4.1 用户请求模型列表的权限流程

```
用户请求 /v1/models
    ↓
1. Token 认证 → 获取 user_id
    ↓
2. 查询 user_group_relations → 获取用户所属分组列表 [group_ids]
    ↓
3. 如果没有分组 → 默认使用 'default' 分组
    ↓
4. 查询 channel_group_relations WHERE group_id IN (group_ids) AND is_enabled=true
    → 获取可用渠道列表 [channel_ids]
    ↓
5. 从 channels WHERE id IN (channel_ids) AND status=1 AND is_healthy=true
    → 获取活跃渠道
    ↓
6. 聚合各渠道的 models 字段，去重后返回
```

### 4.2 用户请求 /v1/chat/completions 的权限流程

```
用户请求 /v1/chat/completions with model=X
    ↓
1. Token 认证 → 获取 user_id, token
    ↓
2. Token 级别检查:
   - 检查 token.allowed_models (如果不为空，X 必须在其中)
   - 检查 token.denied_models (X 不能在其中)
    ↓
3. 用户级别检查:
   - 查询用户分组 → 获取分组关联的渠道
   - 检查模型 X 是否在任一可用渠道的 models 列表中
    ↓
4. 模型定价检查:
   - 查询 model_pricing WHERE model=X → 获取定价
    ↓
5. 渠道选择:
   - 从可用渠道中选择支持模型 X 的渠道
   - 按权重/优先级选择
    ↓
6. 转发请求到渠道
    ↓
7. 记录用量 + 计算成本 (使用 model_pricing 定价)
```

### 4.3 默认分组策略

| 场景 | 行为 |
|------|------|
| 新用户注册 | 自动加入 `default` 分组 |
| VIP 升级 | 自动加入对应 VIP 分组 |
| 新渠道创建 | 默认关联 `default` 分组 |
| 新模型添加 | 自动创建定价记录（使用默认价格） |

---

## 5. API 设计

### 5.1 管理后台 API

#### 模型分组

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/admin/model-groups` | 获取分组列表 |
| `POST` | `/api/v1/admin/model-groups` | 创建分组 |
| `PUT` | `/api/v1/admin/model-groups/:id` | 更新分组 |
| `DELETE` | `/api/v1/admin/model-groups/:id` | 删除分组 |
| `POST` | `/api/v1/admin/model-groups/:id/channels` | 关联渠道 |
| `DELETE` | `/api/v1/admin/model-groups/:id/channels/:cid` | 取消关联渠道 |

#### 模型定价

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/admin/model-pricing` | 获取定价列表 |
| `POST` | `/api/v1/admin/model-pricing` | 创建/更新定价 |
| `PUT` | `/api/v1/admin/model-pricing/:id` | 更新定价 |
| `DELETE` | `/api/v1/admin/model-pricing/:id` | 删除定价 |
| `GET` | `/api/v1/admin/model-pricing/summary` | 获取所有模型汇总 |

#### 用户分组

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/admin/users/:id/groups` | 查看用户分组 |
| `PUT` | `/api/v1/admin/users/:id/groups` | 设置用户分组 |

### 5.2 用户端 API (改造现有)

| 方法 | 路径 | 说明 | 变更 |
|------|------|------|------|
| `GET` | `/api/v1/models` | 获取可用模型 | **改造**: 按用户分组过滤 |
| `POST` | `/api/v1/chat/completions` | 聊天补全 | **改造**: 增加分组权限检查 |

### 5.3 /v1/models 响应格式 (OpenAI 兼容)

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1712345678,
      "owned_by": "openai",
      "permission": [{
        "id": "modelperm-xxx",
        "object": "model_permission",
        "created": 1712345678,
        "allow_create_engine": false,
        "allow_sampling": true,
        "allow_logprobs": true,
        "allow_search_indices": false,
        "allow_view": true,
        "allow_fine_tuning": false,
        "organization": "*",
        "group": null,
        "is_blocking": false
      }],
      "root": "gpt-4",
      "parent": null
    }
  ]
}
```

---

## 6. 初始化数据

```sql
-- 默认分组
INSERT INTO model_groups (name, display_name, description, sort_order) VALUES
('default', '默认组', '所有用户默认可访问的基础模型', 0),
('pro', '专业组', '需要升级后可访问的高级模型', 1),
('vip', 'VIP组', 'VIP 专属模型', 2);

-- 默认模型定价
INSERT INTO model_pricing (model, provider, display_name, price_input, price_output, context_length, max_output, group_id, is_enabled, is_featured) VALUES
-- GPT 系列 (default)
('gpt-3.5-turbo', 'openai', 'GPT-3.5 Turbo', 0.0015, 0.0020, 16385, 4096, 1, true, true),
('gpt-4o-mini', 'openai', 'GPT-4o Mini', 0.00015, 0.0006, 128000, 16384, 1, true, true),
('gpt-4o', 'openai', 'GPT-4o', 0.005, 0.015, 128000, 16384, 2, true, true),
('gpt-4-turbo', 'openai', 'GPT-4 Turbo', 0.01, 0.03, 128000, 4096, 2, true, false),
('gpt-4', 'openai', 'GPT-4', 0.03, 0.06, 8192, 8192, 3, true, false),

-- Claude 系列
('claude-3-haiku', 'anthropic', 'Claude 3 Haiku', 0.00025, 0.00125, 200000, 4096, 1, true, true),
('claude-3-sonnet', 'anthropic', 'Claude 3 Sonnet', 0.003, 0.015, 200000, 4096, 2, true, true),
('claude-3-opus', 'anthropic', 'Claude 3 Opus', 0.015, 0.075, 200000, 4096, 3, true, false),

-- DeepSeek
('deepseek-chat', 'deepseek', 'DeepSeek Chat', 0.0001, 0.0002, 128000, 8192, 1, true, true),
('deepseek-coder', 'deepseek', 'DeepSeek Coder', 0.0001, 0.0002, 128000, 8192, 1, true, false),

-- Gemini
('gemini-pro', 'google', 'Gemini Pro', 0.0005, 0.0015, 128000, 8192, 1, true, true),
('gemini-1.5-pro', 'google', 'Gemini 1.5 Pro', 0.0035, 0.0105, 2000000, 8192, 2, true, false);

-- 渠道默认关联 default 分组
-- (通过迁移脚本，将现有渠道的 group_name='default' 关联到 default 分组)
```

---

## 7. 迁移策略

### 7.1 向后兼容

1. `channels.group_name` 字段保留不变
2. 新用户/新渠道默认加入 `default` 分组
3. 现有用户自动加入 `default` 分组
4. 现有渠道自动关联 `default` 分组

### 7.2 迁移脚本

```sql
-- 迁移现有渠道的 group_name 到 channel_group_relations
INSERT INTO channel_group_relations (tenant_id, channel_id, group_id, is_enabled, priority)
SELECT 
    c.tenant_id, 
    c.id, 
    mg.id,
    true,
    0
FROM channels c
JOIN model_groups mg ON mg.name = c.group_name
WHERE NOT EXISTS (
    SELECT 1 FROM channel_group_relations cgr 
    WHERE cgr.channel_id = c.id
);

-- 迁移现有用户到 default 分组
INSERT INTO user_group_relations (tenant_id, user_id, group_id, created_at)
SELECT 
    u.tenant_id,
    u.id,
    (SELECT id FROM model_groups WHERE name = 'default'),
    NOW()
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM user_group_relations ugr 
    WHERE ugr.user_id = u.id
);
```

---

## 8. 后续扩展

| 功能 | 说明 |
|------|------|
| 模型别名 | 在 `model_pricing` 中增加 `aliases` JSONB 字段 |
| 动态定价 | 根据渠道的 `cost_factor` 动态调整用户端定价 |
| 模型排行榜 | 基于 `usage_logs` 统计模型使用量排名 |
| 模型状态监控 | 实时监控各模型的可用性/延迟 |
