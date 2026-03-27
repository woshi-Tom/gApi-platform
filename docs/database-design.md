# Claw AI - 数据库设计方案（含健康检查 + One API 令牌）

> **设计时间**：2026-03-20  
> **设计角色**：产品经理 + 数据库架构师  
> **设计目标**：支持 API 健康检查、One API 令牌管理、审计日志溯源

---

## 📊 数据库架构总览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Claw AI 数据库架构                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐               │
│  │    users     │────▶│    tokens    │────▶│ usage_logs   │               │
│  │   (用户表)    │     │  (令牌表)    │     │ (使用日志)   │               │
│  └──────────────┘     └──────────────┘     └──────────────┘               │
│         │                    │                                              │
│         │                    │                                              │
│         ▼                    ▼                                              │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐               │
│  │oneapi_tokens │     │  abilities   │────▶│  channels    │               │
│  │(One API令牌) │     │  (能力表)    │     │  (渠道表)    │               │
│  └──────────────┘     └──────────────┘     └──────────────┘               │
│                                                    │                        │
│                                                    ▼                        │
│                                             ┌──────────────┐               │
│                                             │channel_health│               │
│                                             │(健康记录)     │               │
│                                             └──────────────┘               │
│                                                                             │
│  ┌──────────────┐     ┌──────────────┐                                    │
│  │  audit_logs  │     │  event_logs  │                                    │
│  │  (审计日志)   │     │  (事件日志)   │                                    │
│  └──────────────┘     └──────────────┘                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 一、核心表设计

### 1.1 渠道表 (channels) - 存储上游 API 配置

```sql
CREATE TABLE channels (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,           -- openai/azure/claude/nvidia_nim/deepseek
    base_url VARCHAR(500),
    api_key_encrypted TEXT NOT NULL,      -- 加密存储
    status INT NOT NULL DEFAULT 1,       -- 1=启用, 2=禁用, 3=自动禁用
    weight INT NOT NULL DEFAULT 1,
    priority INT NOT NULL DEFAULT 0,
    models TEXT,                          -- JSON
    model_mapping TEXT,                   -- JSON
    balance DECIMAL(10,4) NOT NULL DEFAULT 0,
    used_quota BIGINT NOT NULL DEFAULT 0,
    group_name VARCHAR(50) NOT NULL DEFAULT 'default',
    config TEXT,                          -- JSON
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_channel_status_priority (status, priority),
    INDEX idx_channel_group (group_name),
    INDEX idx_channel_type (type)
);
```

**产品经理问题**：真实 API 是否需要存入数据库？

**答案**：✅ **必须存入**

**原因**：
1. **持久化配置** - 渠道配置（URL、Key、模型列表）需要持久保存
2. **状态追踪** - 记录渠道的启用/禁用状态
3. **负载均衡** - 支持权重、优先级配置
4. **健康检查** - 基于数据库记录进行定时检测
5. **审计溯源** - 所有变更都有审计日志

---

### 1.2 渠道健康记录表 (channel_health)

```sql
CREATE TABLE channel_health (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    channel_id INT NOT NULL,
    check_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_healthy BOOLEAN NOT NULL,
    response_time_ms INT,
    status_code INT,
    error_message TEXT,
    consecutive_failures INT NOT NULL DEFAULT 0,
    
    INDEX idx_health_channel_time (channel_id, check_time),
    INDEX idx_health_unhealthy (is_healthy, check_time),
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
);
```

**产品经理问题**：不可用的 API 是否要移入其他表？

**答案**：❌ **不需要移入其他表**

**推荐方案**：使用 `status` 字段软删除

```
channels.status = 1  → 启用（可用）
channels.status = 2  → 手动禁用
channels.status = 3  → 自动禁用（健康检查失败）
```

**原因**：
1. **简化查询** - 无需跨表查询
2. **保留历史** - 故障渠道配置保留，便于恢复
3. **自动恢复** - 健康检查成功后可自动恢复为 status=1
4. **审计完整** - 所有状态变更都有日志记录

**健康检查流程**：
```
定时任务 (每5分钟)
  ↓
检查所有 status=1 的渠道
  ↓
┌─────────────────────────────────────────┐
│  调用 GET /v1/models 测试               │
│  记录响应时间、状态码、错误信息           │
└─────────────────────────────────────────┘
  ↓
检查成功 → 更新 channel_health 记录
检查失败 → 连续失败次数 + 1
  ↓
连续失败 ≥ 3 次 → 自动禁用 (status=3)
  ↓
记录审计日志 + 触发告警事件
```

---

### 1.3 One API 令牌表 (oneapi_tokens)

```sql
CREATE TABLE oneapi_tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    oneflow_token_key VARCHAR(200) NOT NULL,
    name VARCHAR(100),
    source VARCHAR(50) NOT NULL DEFAULT 'manual',  -- manual/import/api
    status INT NOT NULL DEFAULT 1,                  -- 1=可用, 2=禁用, 3=耗尽, 4=过期
    remain_quota BIGINT NOT NULL DEFAULT 0,
    used_quota BIGINT NOT NULL DEFAULT 0,
    unlimited_quota BOOLEAN NOT NULL DEFAULT FALSE,
    models TEXT,                                    -- JSON
    expired_time BIGINT NOT NULL DEFAULT -1,
    last_used_at TIMESTAMP,
    imported_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    
    INDEX idx_oneapi_token_user (user_id),
    INDEX idx_oneapi_token_key (oneflow_token_key),
    INDEX idx_oneapi_token_status (status),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

**产品经理问题**：One API 用户令牌是否需要专用表？

**答案**：✅ **需要专用表**

**原因**：
1. **批量导入** - 支持批量导入 One API 令牌
2. **快速检索** - 专用索引支持快速查找可用令牌
3. **状态追踪** - 独立追踪每个令牌的使用状态
4. **来源区分** - 区分手动添加、批量导入、API 获取
5. **效率提升** - 不需要通过网页点点点，直接操作数据库

**批量导入接口设计**：
```python
POST /admin/oneapi-tokens/import
Content-Type: application/json

{
    "tokens": [
        {"key": "sk-xxx1", "name": "令牌1", "quota": 1000000},
        {"key": "sk-xxx2", "name": "令牌2", "quota": 2000000}
    ],
    "source": "import",
    "default_status": 1
}
```

**快速分配令牌**：
```python
def allocate_oneapi_token(user_id: int, model: str) -> Optional[str]:
    """为用户分配一个可用的 One API 令牌"""
    with get_session() as session:
        token = session.query(OneApiToken).filter(
            OneApiToken.status == 1,
            OneApiToken.remain_quota > 0,
            or_(
                OneApiToken.models.is_(None),
                OneApiToken.models.contains(model)
            )
        ).order_by(OneApiToken.remain_quota.desc()).first()
        
        if token:
            token.last_used_at = datetime.utcnow()
            session.commit()
            return token.oneflow_token_key
        
        return None
```

---

## 二、扩展表设计

### 2.1 扩展 users 表

```sql
ALTER TABLE users 
    ADD COLUMN quota BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN used_quota BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN api_group VARCHAR(50) NOT NULL DEFAULT 'default';
```

### 2.2 令牌表 (tokens) - 增加限速字段

```sql
CREATE TABLE tokens (
    -- ... 现有字段 ...
    rate_limit INT NOT NULL DEFAULT 0,  -- 每分钟请求限制 (0=不限)
    
    INDEX idx_token_key (`key`),
    INDEX idx_token_user (user_id),
    INDEX idx_token_status (status)
);
```

### 2.3 使用日志表 (usage_logs)

```sql
CREATE TABLE usage_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    token_id INT,
    channel_id INT,
    model VARCHAR(100) NOT NULL,
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    quota BIGINT NOT NULL DEFAULT 0,
    request_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    response_time_ms INT,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    
    INDEX idx_usage_user_time (user_id, request_at),
    INDEX idx_usage_token (token_id),
    INDEX idx_usage_channel (channel_id),
    INDEX idx_usage_model (model)
);
```

---

## 三、审计日志增强

### 3.1 审计日志表索引

```sql
CREATE INDEX idx_audit_actor ON audit_logs(actor_type, actor_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_object ON audit_logs(object_type, object_id);
CREATE INDEX idx_audit_time ON audit_logs(created_at);
```

### 3.2 事件日志表索引

```sql
CREATE INDEX idx_event_type ON event_logs(event_type);
CREATE INDEX idx_event_related ON event_logs(related_type, related_id);
CREATE INDEX idx_event_time ON event_logs(created_at);
```

---

## 四、数据流设计

### 4.1 渠道健康检查数据流

```
定时任务 (APScheduler)
  ↓
HealthCheckService.check_all_channels()
  ↓
┌─────────────────────────────────────────────────┐
│  SELECT * FROM channels WHERE status IN (1, 3)  │
└─────────────────────────────────────────────────┘
  ↓
并发测试每个渠道
  ↓
┌─────────────────────────────────────────────────┐
│  INSERT INTO channel_health (...)               │
│  UPDATE channels SET status = 3 WHERE ...       │
│  INSERT INTO audit_logs (...)                   │
│  INSERT INTO event_logs (...)                   │
└─────────────────────────────────────────────────┘
```

### 4.2 One API 令牌分配数据流

```
用户请求 API
  ↓
检查本地令牌池
  ↓
SELECT * FROM tokens WHERE user_id = ? AND status = 1 AND remain_quota > 0
  ↓
如果没有可用令牌 → 分配 One API 令牌
  ↓
SELECT * FROM oneapi_tokens WHERE status = 1 AND remain_quota > 0 LIMIT 1
  ↓
返回 oneflow_token_key 给用户
  ↓
UPDATE oneapi_tokens SET last_used_at = NOW()
```

---

## 五、索引设计总结

| 表名 | 索引名 | 字段 | 用途 |
|------|--------|------|------|
| channels | idx_channel_status_priority | status, priority | 可用渠道查询 |
| channels | idx_channel_group | group_name | 分组查询 |
| channels | idx_channel_type | type | 类型筛选 |
| abilities | idx_abilities_group_model | group_name, model, enabled | 模型→渠道查询 |
| channel_health | idx_health_channel_time | channel_id, check_time | 健康历史查询 |
| channel_health | idx_health_unhealthy | is_healthy, check_time | 故障渠道查询 |
| tokens | idx_token_key | key | 令牌验证 |
| tokens | idx_token_user | user_id | 用户令牌查询 |
| tokens | idx_token_status | status | 状态筛选 |
| usage_logs | idx_usage_user_time | user_id, request_at | 用户用量统计 |
| usage_logs | idx_usage_token | token_id | 令牌使用统计 |
| usage_logs | idx_usage_channel | channel_id | 渠道使用统计 |
| oneapi_tokens | idx_oneapi_token_user | user_id | 用户令牌查询 |
| oneapi_tokens | idx_oneapi_token_key | oneflow_token_key | 令牌查找 |
| oneapi_tokens | idx_oneapi_token_status | status | 可用令牌查询 |
| audit_logs | idx_audit_actor | actor_type, actor_id | 操作者查询 |
| audit_logs | idx_audit_action | action | 操作类型查询 |
| audit_logs | idx_audit_object | object_type, object_id | 对象查询 |
| audit_logs | idx_audit_time | created_at | 时间范围查询 |

---

## 六、文件清单

| 文件 | 路径 | 用途 |
|------|------|------|
| 数据库模型 | `source_code/app/models/pool_models.py` | SQLAlchemy 模型定义 |
| 审计服务 | `source_code/app/services/audit_service.py` | 审计日志写入 |
| 健康检查服务 | `source_code/app/services/health_check_service.py` | 渠道健康检查 |
| 迁移脚本 | `scripts/migrate_pool_tables.sql` | 数据库迁移 SQL |
| 设计文档 | `.sisyphus/plans/database-design.md` | 本文档 |

---

## ✅ 设计决策总结

| 问题 | 决策 | 理由 |
|------|------|------|
| 真实 API 是否存入数据库？ | ✅ 是 | 持久化配置、状态追踪、健康检查 |
| 不可用 API 是否移入其他表？ | ❌ 否 | 使用 status 字段软删除，简化查询 |
| One API 令牌是否专用表？ | ✅ 是 | 批量导入、快速检索、状态追踪 |
| 健康检查如何设计？ | 独立表 | 记录历史、支持趋势分析 |
| API Key 如何存储？ | 加密存储 | 安全考虑 |
