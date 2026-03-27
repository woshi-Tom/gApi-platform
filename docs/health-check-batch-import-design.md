# API 健康检查 & 批量令牌管理 - 设计方案

> **设计时间**：2026-03-20  
> **设计角色**：产品经理 + 架构师  
> **状态**：待确认

---

## 一、需求分析

### 1.1 API Key 健康检查

**用户痛点**：
- 不知道哪些渠道可用
- 渠道故障后需要手动禁用
- 恢复后需要手动启用

**期望**：
- 自动检测渠道可用性
- 可用的打上"可用"标签
- 不可用的自动禁用
- 恢复后自动启用

### 1.2 用户令牌批量管理

**用户痛点**：
- 通过网页逐个创建令牌效率低
- 批量导入/导出不便
- 脚本化操作困难

**期望**：
- 支持批量导入令牌
- 支持 API 接口调用
- 支持 CSV/Excel 导入

---

## 二、设计方案

### 2.1 API Key 健康检查

#### 2.1.1 数据库设计变更

**channels 表新增字段**：

```sql
ALTER TABLE channels ADD COLUMN last_check_at TIMESTAMP NULL;
ALTER TABLE channels ADD COLUMN last_check_status VARCHAR(20) DEFAULT 'unknown';
ALTER TABLE channels ADD COLUMN error_count INT DEFAULT 0;
ALTER TABLE channels ADD COLUMN last_error_message TEXT NULL;
ALTER TABLE channels ADD COLUMN check_interval INT DEFAULT 300; -- 秒，默认5分钟
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| last_check_at | TIMESTAMP | 最后检查时间 |
| last_check_status | VARCHAR(20) | 检查结果：healthy/unhealthy/unknown |
| error_count | INT | 连续失败次数 |
| last_error_message | TEXT | 最后错误信息 |
| check_interval | INT | 检查间隔（秒） |

#### 2.1.2 定时任务设计

**技术选型**：APScheduler（轻量级，适合单体应用）

```python
# backend/app/services/health_check_service.py
from apscheduler.schedulers.background import BackgroundScheduler

class HealthCheckService:
    def __init__(self):
        self.scheduler = BackgroundScheduler()
    
    def start(self):
        """启动定时检查"""
        self.scheduler.add_job(
            self.check_all_channels,
            'interval',
            seconds=60,  # 每分钟扫描一次
            id='health_check'
        )
        self.scheduler.start()
    
    def check_all_channels(self):
        """检查所有活跃渠道"""
        channels = Channel.query.filter_by(status=1).all()
        
        for channel in channels:
            # 判断是否需要检查
            if self.should_check(channel):
                self.check_channel(channel)
    
    def should_check(self, channel: Channel) -> bool:
        """判断是否需要检查"""
        if not channel.last_check_at:
            return True
        
        elapsed = (datetime.now() - channel.last_check_at).total_seconds()
        return elapsed >= channel.check_interval
    
    def check_channel(self, channel: Channel):
        """检查单个渠道"""
        try:
            # 使用 /v1/models 端点检查，不消耗额度
            response = requests.get(
                f"{channel.base_url}/v1/models",
                headers={"Authorization": f"Bearer {channel.api_key}"},
                timeout=10
            )
            
            if response.status_code == 200:
                channel.last_check_status = 'healthy'
                channel.error_count = 0
                channel.last_error_message = None
                
                # 如果之前是自动禁用，恢复启用
                if channel.status == 3:
                    channel.status = 1
            else:
                self.mark_unhealthy(channel, f"HTTP {response.status_code}")
                
        except requests.Timeout:
            self.mark_unhealthy(channel, "Request timeout")
        except requests.RequestException as e:
            self.mark_unhealthy(channel, str(e))
        
        channel.last_check_at = datetime.now()
        db.session.commit()
    
    def mark_unhealthy(self, channel: Channel, error: str):
        """标记为不健康"""
        channel.last_check_status = 'unhealthy'
        channel.error_count += 1
        channel.last_error_message = error
        
        # 连续失败 3 次自动禁用
        if channel.error_count >= 3:
            channel.status = 3  # 自动禁用
            logger.warning(f"Channel {channel.id} auto-disabled: {error}")
```

#### 2.1.3 健康检查策略

| 策略 | 说明 |
|------|------|
| **检查端点** | `GET /v1/models`（不消耗额度） |
| **检查频率** | 默认 5 分钟，可配置 |
| **失败阈值** | 连续 3 次失败自动禁用 |
| **恢复策略** | 检测到恢复后自动启用 |
| **超时设置** | 10 秒 |
| **并发控制** | 单线程顺序检查 |

#### 2.1.4 API 接口

```python
# 手动触发检查
POST /admin/channels/{channel_id}/check

# 批量检查
POST /admin/channels/check-all

# 获取检查历史
GET /admin/channels/{channel_id}/check-history
```

---

### 2.2 用户令牌批量管理

#### 2.2.1 批量导入接口

**接口设计**：

```python
POST /admin/tokens/batch-import
Content-Type: application/json

{
  "tokens": [
    {
      "name": "用户A",
      "user_id": 1001,
      "quota": 1000000,
      "unlimited_quota": false,
      "models": ["gpt-3.5-turbo", "gpt-4"]
    },
    {
      "name": "用户B",
      "user_id": 1002,
      "quota": 2000000,
      "unlimited_quota": true
    }
  ],
  "batch_name": "2026-03-20 批量导入"
}
```

**响应**：

```json
{
  "success": true,
  "imported_count": 2,
  "failed_count": 0,
  "results": [
    {
      "name": "用户A",
      "key": "sk-xxxxxxxxxxxx",
      "status": "success"
    },
    {
      "name": "用户B",
      "key": "sk-yyyyyyyyyyyy",
      "status": "success"
    }
  ]
}
```

#### 2.2.2 CSV 导入接口

```python
POST /admin/tokens/import-csv
Content-Type: multipart/form-data

# CSV 格式
name,user_id,quota,unlimited_quota,models
用户A,1001,1000000,false,"gpt-3.5-turbo,gpt-4"
用户B,1002,0,true,
```

#### 2.2.3 批量导出接口

```python
GET /admin/tokens/export?format=csv

# 响应：CSV 文件下载
name,key,quota,used_quota,status,created_at
用户A,sk-xxxx,1000000,50000,active,2026-03-20
用户B,sk-yyyy,unlimited,100000,active,2026-03-20
```

#### 2.2.4 数据库增强

**tokens 表新增字段**：

```sql
ALTER TABLE tokens ADD COLUMN batch_id VARCHAR(50) NULL;
ALTER TABLE tokens ADD COLUMN batch_name VARCHAR(100) NULL;
ALTER TABLE tokens ADD COLUMN import_source VARCHAR(20) DEFAULT 'manual';
```

**批次记录表**：

```sql
CREATE TABLE token_batches (
    batch_id VARCHAR(50) PRIMARY KEY,
    batch_name VARCHAR(100),
    total_count INT,
    success_count INT,
    failed_count INT,
    created_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 2.2.5 脚本化管理

**Python SDK 示例**：

```python
from claw_ai_sdk import ClawAI

client = ClawAI(
    base_url="https://your-domain.com/api",
    admin_token="admin-token-xxx"
)

# 批量创建令牌
tokens = client.tokens.batch_create([
    {"name": "用户A", "quota": 1000000},
    {"name": "用户B", "quota": 2000000}
])

# 查询所有令牌
all_tokens = client.tokens.list()

# 更新配额
client.tokens.update_quota(token_id=1, quota=5000000)

# 禁用令牌
client.tokens.disable(token_id=1)
```

---

### 2.3 渠道与令牌关联

#### 2.3.1 是否需要关联？

**当前设计**：渠道和令牌是独立的
- 渠道 = 上游 API 配置
- 令牌 = 用户 API Key
- 通过 abilities 表间接关联（模型-渠道映射）

**是否需要直接关联**？

| 场景 | 需要关联 | 不需要关联 |
|------|----------|------------|
| 专属渠道 | ✅ 某用户只能用某渠道 | ❌ |
| 负载均衡 | ❌ | ✅ 自动选择 |
| 备用渠道 | ❌ | ✅ 自动切换 |

**建议**：当前不需要直接关联，保持解耦设计。

#### 2.3.2 扩展能力

如果未来需要，可以添加：

```sql
-- 令牌-渠道绑定表（可选）
CREATE TABLE token_channel_bindings (
    token_id INT,
    channel_id INT,
    PRIMARY KEY (token_id, channel_id),
    FOREIGN KEY (token_id) REFERENCES tokens(id),
    FOREIGN KEY (channel_id) REFERENCES channels(id)
);
```

---

## 三、技术选型

### 3.1 定时任务

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| APScheduler | 轻量、易集成 | 不支持分布式 | ✅ 单体应用 |
| Celery | 功能强大 | 需要 Redis/RabbitMQ | ❌ 过重 |
| cron | 简单 | 不够灵活 | ❌ |

**推荐**：APScheduler（当前阶段）

### 3.2 批量导入

| 选项 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| JSON API | 灵活、易脚本化 | 需要编程 | ✅ |
| CSV 上传 | 用户友好 | 格式限制 | ✅ |
| Excel 上传 | 功能丰富 | 依赖库 | 🟡 可选 |

**推荐**：JSON API + CSV 上传

---

## 四、API 接口清单

### 4.1 健康检查相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /admin/channels/{id}/check | 手动检查单个渠道 |
| POST | /admin/channels/check-all | 批量检查所有渠道 |
| GET | /admin/channels/{id}/health | 获取渠道健康状态 |
| GET | /admin/channels/health-summary | 获取健康状态汇总 |

### 4.2 令牌批量管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /admin/tokens/batch-import | JSON 批量导入 |
| POST | /admin/tokens/import-csv | CSV 文件导入 |
| GET | /admin/tokens/export | 导出令牌列表 |
| GET | /admin/tokens/batches | 获取导入批次列表 |
| GET | /admin/tokens/batches/{id} | 获取批次详情 |

---

## 五、待办任务更新

### 补充任务（新增）

| 序号 | 任务 | 优先级 | 预计时间 |
|------|------|--------|----------|
| 5 | channels 表增加健康检查字段 | 🔴 | 0.5天 |
| 6 | 实现 HealthCheckService | 🔴 | 1天 |
| 7 | 实现批量导入接口 | 🟡 | 1天 |
| 8 | 实现 CSV 导入/导出 | 🟡 | 1天 |
| 9 | 实现 Python SDK（可选） | 🟢 | 2天 |

---

## ✅ 结论

### 建议方案

1. **API Key 健康检查**：✅ 推荐
   - channels 表增加健康检查字段
   - 使用 APScheduler 定时检查
   - 检查 `/v1/models` 端点（不消耗额度）
   - 连续失败 3 次自动禁用

2. **用户令牌批量管理**：✅ 推荐
   - JSON API 批量导入
   - CSV 文件导入/导出
   - 批次记录追踪
   - Python SDK（后期）

3. **渠道与令牌关联**：❌ 当前不需要
   - 保持解耦设计
   - 通过 abilities 表间接关联
   - 未来可扩展直接关联

### 数据库变更汇总

```sql
-- channels 表
ALTER TABLE channels ADD COLUMN last_check_at TIMESTAMP NULL;
ALTER TABLE channels ADD COLUMN last_check_status VARCHAR(20) DEFAULT 'unknown';
ALTER TABLE channels ADD COLUMN error_count INT DEFAULT 0;
ALTER TABLE channels ADD COLUMN last_error_message TEXT NULL;
ALTER TABLE channels ADD COLUMN check_interval INT DEFAULT 300;

-- tokens 表
ALTER TABLE tokens ADD COLUMN batch_id VARCHAR(50) NULL;
ALTER TABLE tokens ADD COLUMN batch_name VARCHAR(100) NULL;
ALTER TABLE tokens ADD COLUMN import_source VARCHAR(20) DEFAULT 'manual';

-- 新增批次表
CREATE TABLE token_batches (
    batch_id VARCHAR(50) PRIMARY KEY,
    batch_name VARCHAR(100),
    total_count INT,
    success_count INT,
    failed_count INT,
    created_by INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
