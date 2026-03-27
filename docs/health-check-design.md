# 渠道健康检查 & 令牌批量管理设计方案

> **设计时间**：2026-03-20  
> **设计角色**：产品经理 + 数据库架构师  
> **状态**：待确认

---

## 一、渠道健康检查机制

### 1.1 设计方案对比

| 方案 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| **A: status 字段区分** | 简单，无需新表 | 无法记录检查历史 | ✅ 推荐 |
| **B: 单独 disabled_channels 表** | 清晰分离 | 需要数据迁移，复杂度高 | ❌ 不推荐 |
| **C: status + health_logs 表** | 兼顾简单和可追溯 | 需要额外存储 | ⭐ 备选 |

### 1.2 推荐方案：status 字段 + health_logs 表

#### channels 表新增字段

```sql
ALTER TABLE channels ADD COLUMN health_status VARCHAR(20) DEFAULT 'unknown';
-- 值: unknown / healthy / unhealthy / checking

ALTER TABLE channels ADD COLUMN last_check_at TIMESTAMP NULL;
ALTER TABLE channels ADD COLUMN last_check_result TEXT NULL;
ALTER TABLE channels ADD COLUMN consecutive_failures INT DEFAULT 0;
ALTER TABLE channels ADD COLUMN consecutive_successes INT DEFAULT 0;

-- 索引
CREATE INDEX idx_channel_health ON channels(health_status, status);
```

#### channel_health_logs 表（检查历史）

```sql
CREATE TABLE channel_health_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    channel_id INT NOT NULL,
    check_type VARCHAR(20) NOT NULL,        -- models / chat / embedding
    status VARCHAR(20) NOT NULL,            -- success / failure / timeout
    response_time_ms INT,                   -- 响应时间（毫秒）
    error_message TEXT,                     -- 错误信息
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    INDEX idx_channel_checked (channel_id, checked_at DESC)
);
```

### 1.3 健康检查策略

| 检查类型 | 频率 | 检测方法 | 用途 |
|----------|------|----------|------|
| **快速检查** | 每 5 分钟 | GET /v1/models | 验证 Key 有效性 |
| **深度检查** | 每小时 | POST /v1/chat/completions (测试请求) | 验证实际可用性 |
| **余额检查** | 每天 | 查询余额接口 | 监控余额变化 |

### 1.4 状态转换逻辑

```
                    ┌─────────────┐
                    │   unknown   │ ← 初始状态
                    └──────┬──────┘
                           │ 首次检查
                           ▼
         ┌─────────────────┴─────────────────┐
         │                                   │
         ▼                                   ▼
┌─────────────────┐                 ┌─────────────────┐
│    healthy      │                 │    unhealthy    │
│  (连续成功≥3)   │                 │  (连续失败≥3)   │
└────────┬────────┘                 └────────┬────────┘
         │                                   │
         │ 连续失败≥3                         │ 连续成功≥3
         ▼                                   ▼
┌─────────────────┐                 ┌─────────────────┐
│    unhealthy    │                 │    healthy      │
└─────────────────┘                 └─────────────────┘
```

### 1.5 自动禁用/启用规则

```python
class HealthCheckService:
    FAILURE_THRESHOLD = 3      # 连续失败 3 次 → 禁用
    SUCCESS_THRESHOLD = 3      # 连续成功 3 次 → 启用
    CHECK_INTERVAL = 300       # 5 分钟检查一次
    
    def update_channel_health(self, channel_id: int, is_success: bool, error_msg: str = None):
        channel = Channel.query.get(channel_id)
        
        if is_success:
            channel.consecutive_failures = 0
            channel.consecutive_successes += 1
            channel.health_status = 'healthy'
            
            # 自动启用
            if channel.status == 3 and channel.consecutive_successes >= self.SUCCESS_THRESHOLD:
                channel.status = 1  # 启用
                logger.info(f"Channel {channel_id} auto-enabled after recovery")
        else:
            channel.consecutive_successes = 0
            channel.consecutive_failures += 1
            channel.health_status = 'unhealthy'
            channel.last_check_result = error_msg
            
            # 自动禁用
            if channel.status == 1 and channel.consecutive_failures >= self.FAILURE_THRESHOLD:
                channel.status = 3  # 自动禁用
                logger.warning(f"Channel {channel_id} auto-disabled after {self.FAILURE_THRESHOLD} failures")
        
        channel.last_check_at = datetime.utcnow()
        db.session.commit()
```

### 1.6 不需要单独的 disabled_channels 表

**原因**：
1. `channels.status` 字段已经足够区分（1=启用, 2=手动禁用, 3=自动禁用）
2. 单独表会增加复杂度和数据同步问题
3. 通过 `channel_health_logs` 可以追溯检查历史

**查询可用渠道**：
```sql
SELECT * FROM channels 
WHERE status = 1 
  AND health_status IN ('healthy', 'unknown')
ORDER BY priority DESC, weight DESC;
```

---

## 二、用户令牌批量管理

### 2.1 tokens 表字段增强

```sql
ALTER TABLE tokens ADD COLUMN rate_limit_per_minute INT DEFAULT 60;
ALTER TABLE tokens ADD COLUMN rate_limit_per_hour INT DEFAULT 1000;
ALTER TABLE tokens ADD COLUMN ip_whitelist TEXT NULL;  -- JSON 数组
ALTER TABLE tokens ADD COLUMN last_used_at TIMESTAMP NULL;
ALTER TABLE tokens ADD COLUMN usage_count BIGINT DEFAULT 0;
ALTER TABLE tokens ADD COLUMN tags VARCHAR(255) NULL;  -- 标签，用于批量管理

-- 索引
CREATE INDEX idx_token_user ON tokens(user_id, status);
CREATE INDEX idx_token_tags ON tokens(tags);
```

### 2.2 批量管理 API 设计

#### 管理员 API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/admin/tokens/batch-create` | POST | 批量创建令牌 |
| `/admin/tokens/batch-update` | PUT | 批量更新令牌 |
| `/admin/tokens/batch-disable` | POST | 批量禁用令牌 |
| `/admin/tokens/export` | GET | 导出令牌列表 |
| `/admin/tokens/import` | POST | 导入令牌 |
| `/admin/tokens/stats` | GET | 令牌统计 |

#### 批量创建示例

```python
@admin.route('/tokens/batch-create', methods=['POST'])
def batch_create_tokens():
    """批量创建令牌"""
    data = request.get_json()
    user_id = data['userId']
    count = data.get('count', 1)
    quota_per_token = data.get('quota', 0)
    tags = data.get('tags', '')
    
    tokens = []
    for _ in range(count):
        token = Token(
            user_id=user_id,
            key=generate_token_key(),
            name=f"Auto-generated {datetime.now().strftime('%Y%m%d%H%M%S')}",
            remain_quota=quota_per_token,
            tags=tags
        )
        db.session.add(token)
        tokens.append(token)
    
    db.session.commit()
    
    # 记录审计日志
    AuditService.log(
        action='batch_create_tokens',
        object_type='token',
        object_id=user_id,
        actor_type='admin',
        actor_id=get_current_admin().id,
        after_data={'count': count, 'quota': quota_per_token}
    )
    
    return {'tokens': [t.to_dict() for t in tokens]}, 201
```

### 2.3 CLI 工具设计（可选，Phase 2）

```bash
# 创建令牌
python manage.py token create --user-id 1 --quota 10000 --count 10

# 列出令牌
python manage.py token list --user-id 1 --status active

# 批量禁用
python manage.py token disable --tags "batch-202403"

# 导出令牌
python manage.py token export --user-id 1 --output tokens.csv
```

---

## 三、数据库表结构最终设计

### 3.1 channels 表（更新后）

```sql
CREATE TABLE channels (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,              -- openai / nvidia_nim / claude / gemini
    base_url VARCHAR(500),
    api_key_encrypted TEXT NOT NULL,        -- 加密存储
    status INT DEFAULT 1,                   -- 1=启用, 2=手动禁用, 3=自动禁用
    weight INT DEFAULT 1,
    priority INT DEFAULT 0,
    models TEXT,                            -- JSON 数组
    model_mapping TEXT,                     -- JSON 对象
    balance DECIMAL(10,4) DEFAULT 0,
    used_quota BIGINT DEFAULT 0,
    group_name VARCHAR(50) DEFAULT 'default',
    config TEXT,                            -- JSON 扩展配置
    
    -- 新增：健康检查字段
    health_status VARCHAR(20) DEFAULT 'unknown',
    last_check_at TIMESTAMP NULL,
    last_check_result TEXT NULL,
    consecutive_failures INT DEFAULT 0,
    consecutive_successes INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_channel_status_priority (status, priority DESC),
    INDEX idx_channel_health (health_status, status),
    INDEX idx_channel_group (group_name)
);
```

### 3.2 channel_health_logs 表

```sql
CREATE TABLE channel_health_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    channel_id INT NOT NULL,
    check_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    response_time_ms INT,
    error_message TEXT,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    INDEX idx_channel_checked (channel_id, checked_at DESC)
);
```

### 3.3 abilities 表（不变）

```sql
CREATE TABLE abilities (
    group_name VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    channel_id INT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    
    PRIMARY KEY (group_name, model, channel_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    INDEX idx_abilities_group_model (group_name, model, enabled)
);
```

### 3.4 tokens 表（更新后）

```sql
CREATE TABLE tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    key VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100),
    status INT DEFAULT 1,                   -- 1=启用, 2=禁用, 3=过期, 4=额度耗尽
    remain_quota BIGINT DEFAULT 0,
    used_quota BIGINT DEFAULT 0,
    unlimited_quota BOOLEAN DEFAULT FALSE,
    models TEXT,                            -- JSON 数组，允许的模型
    subnet VARCHAR(100),                    -- IP 白名单
    expired_time BIGINT DEFAULT -1,         -- -1=永不过期
    
    -- 新增字段
    rate_limit_per_minute INT DEFAULT 60,
    rate_limit_per_hour INT DEFAULT 1000,
    ip_whitelist TEXT NULL,                 -- JSON 数组
    last_used_at TIMESTAMP NULL,
    usage_count BIGINT DEFAULT 0,
    tags VARCHAR(255) NULL,                -- 批量管理标签
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token_user (user_id, status),
    INDEX idx_token_key (key),
    INDEX idx_token_tags (tags)
);
```

### 3.5 usage_logs 表（不变）

```sql
CREATE TABLE usage_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    token_id INT,
    channel_id INT,
    model VARCHAR(100) NOT NULL,
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    quota BIGINT DEFAULT 0,
    request_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    INDEX idx_usage_user_time (user_id, request_at DESC),
    INDEX idx_usage_token (token_id),
    INDEX idx_usage_channel (channel_id)
);
```

---

## 四、实现优先级

| 功能 | 优先级 | 阶段 | 说明 |
|------|--------|------|------|
| channels 表新增健康字段 | 🔴 P0 | Phase 1 | 基础设施 |
| channel_health_logs 表 | 🔴 P0 | Phase 1 | 检查历史 |
| 健康检查定时任务 | 🔴 P0 | Phase 1 | 核心功能 |
| 自动禁用/启用逻辑 | 🟡 P1 | Phase 1 | 依赖健康检查 |
| tokens 表新增字段 | 🟡 P1 | Phase 1 | 批量管理基础 |
| 批量创建 API | 🟡 P1 | Phase 1 | 提高效率 |
| 批量导入/导出 | 🟢 P2 | Phase 2 | 可选功能 |
| CLI 工具 | 🟢 P2 | Phase 2 | 可选功能 |

---

## 五、风险提示

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 健康检查频率过高 | 上游限流 | 设置合理间隔（5分钟） |
| 健康检查请求成本 | 资金消耗 | 使用免费端点（/v1/models） |
| 自动禁用误判 | 可用渠道被禁 | 连续失败阈值设为 3 |
| 令牌批量创建 | 数据库压力 | 分批提交（每批 100 个） |
| 加密 Key 泄露 | 所有渠道泄露 | Key 单独管理，定期轮换 |
