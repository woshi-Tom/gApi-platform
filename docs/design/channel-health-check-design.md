# 渠道健康检测功能设计方案

> 版本: v2.0
> 日期: 2026-04-01
> 状态: ✅ 已实现（部分）
> 审核意见: 核心功能已完成，需完善前端UI和配置界面

---

## 1. 功能概述

渠道健康检测功能通过定时任务自动检测API渠道的可用性，自动标记不健康渠道并支持自动恢复。确保平台稳定性和高可用性。

## 2. 现有实现

### 2.1 已实现文件

| 文件 | 说明 | 状态 |
|------|------|------|
| `backend/internal/service/health_check.go` | 健康检测服务核心逻辑 | ✅ 完成 |
| `backend/internal/model/channel.go` | 渠道模型（含健康状态字段） | ✅ 完成 |
| `backend/internal/pkg/adapter/factory.go` | 适配器工厂 | ✅ 完成 |
| `backend/internal/repository/channel_repository.go` | 渠道数据访问层 | ✅ 完成 |

### 2.2 现有配置常量

```go
const (
    FailureThreshold     = 3      // 连续失败3次后禁用
    CheckIntervalMinutes = 5      // 每5分钟检测一次
    DeadRetryHours      = 1       // 禁用渠道重试间隔
    RequestTimeout       = 30     // 请求超时30秒
)
```

### 2.3 现有数据库字段

```sql
ALTER TABLE channels ADD COLUMN IF NOT EXISTS is_healthy BOOLEAN DEFAULT true;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS failure_count INTEGER DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS last_success_at TIMESTAMP;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS last_check_at TIMESTAMP;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS response_time_avg INTEGER DEFAULT 0; -- ms
```

### 2.4 现有适配器类型

| 类型 | 官方API | 默认BaseURL |
|------|---------|-------------|
| openai | https://api.openai.com | ✅ 自动填充 |
| nvidia | NVIDIA NIM | ✅ 自动填充 |
| azure | Azure OpenAI | ❌ 必须填写 |
| claude | https://api.anthropic.com | ✅ 自动填充 |
| gemini | https://generativelanguage.googleapis.com | ✅ 自动填充 |
| deepseek | https://api.deepseek.com | ✅ 自动填充 |
| zhipu | 智谱AI | ✅ 自动填充 |
| baidu | 百度千帆 | ✅ 自动填充 |
| yi | 零一万物 | ✅ 自动填充 |
| ollama | 本地部署 | ✅ 自动填充 |
| groq | https://api.groq.com | ✅ 自动填充 |
| custom | 自定义端点 | ❌ 必须填写 |

### 2.5 BaseURL处理逻辑

**创建/更新渠道时**：
- 如果用户填写了自定义BaseURL → 使用用户填写的地址
- 如果用户不填（留空）→ 自动使用该渠道类型的官方API地址

**健康检查时**：
- 始终使用渠道保存的BaseURL进行检查
- 如果BaseURL为空，则使用适配器的默认地址

**代码位置**：
- 适配器默认地址：`backend/internal/pkg/adapter/*.go`
- 渠道创建/更新：`backend/internal/handler/channel_handler.go`, `backend/internal/handler/admin_handler.go`

---

## 3. 核心逻辑

### 3.1 服务启动

```go
func (s *HealthCheckService) Start() {
    if s.isRunning {
        return
    }
    s.isRunning = true
    go s.run()
    log.Println("Health check service started")
}
```

- 使用 goroutine 启动后台任务
- 防止重复启动

### 3.2 定时检测

```go
func (s *HealthCheckService) run() {
    ticker := time.NewTicker(time.Duration(CheckIntervalMinutes) * time.Minute)
    defer ticker.Stop()

    s.checkAllChannels()

    for {
        select {
        case <-s.stopChan:
            return
        case <-ticker.C:
            s.checkAllChannels()
        }
    }
}
```

- 每5分钟执行一次全量检测
- 支持优雅停止

### 3.3 检测流程

```
checkAllChannels()
    │
    ├── 获取所有活跃渠道 (GetActiveChannels)
    │
    └── 并发检测每个渠道 (goroutine)
            │
            ├── 获取渠道信息 (GetByID)
            │
            ├── 解密API Key
            │
            ├── 获取适配器 (adapter.GetAdapter)
            │
            └── testChannel()
                    │
                    └── ListModels() 测试API连接
                            │
                            ├── 成功 → markHealthy()
                            │       ├── ResetFailureCount
                            │       └── UpdateResponseTime
                            │
                            └── 失败 → markFailed()
                                    ├── IncrementFailureCount
                                    └── if failures >= 3 → markUnhealthy()
```

### 3.4 失败处理

```go
func (s *HealthCheckService) markFailed(channel *model.Channel, errorMsg string) {
    err := s.channelRepo.IncrementFailureCount(channel.ID)
    // ...

    if cached.Failures >= FailureThreshold {
        s.markUnhealthy(channel, errorMsg)
    }
}
```

- 连续失败3次后自动标记为不健康
- 更新 `is_healthy = false`
- 设置 `failure_count = 3`

---

## 4. 待完成功能

### 4.1 前端UI

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 渠道列表健康状态显示 | 🔴 高 | 状态指示灯、最后检测时间、响应时间 |
| 手动检测按钮 | 🔴 高 | 触发单渠道立即检测 |
| 健康详情页面 | 🟡 中 | 统计图表、检测历史 |
| 健康检测配置 | 🟡 中 | 检测间隔、超时时间等 |

### 4.2 后端API

| 接口 | 方法 | 说明 | 状态 |
|------|------|------|------|
| `/api/v1/admin/channels/:id/health` | POST | 手动触发检测 | ✅ 已实现 |
| `/api/v1/admin/channels/:id/health` | GET | 获取健康状态 | ✅ 已实现 |
| `/api/v1/admin/channels/health/stats` | GET | 获取统计信息 | ✅ 已实现 |

### 4.3 高级功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 自动恢复 | ❌ 未实现 | 连续成功N次后自动启用 |
| 健康历史记录 | ❌ 未实现 | 记录每次检测结果 |
| 动态阈值配置 | ❌ 未实现 | 允许自定义失败阈值 |
| 告警通知 | ❌ 未实现 | 渠道禁用时通知管理员 |

---

## 5. 前端UI设计

### 5.1 渠道列表增强

```
┌─────────────────────────────────────────────────────────────┐
│ 渠道管理                                          [+ 新增渠道] │
├─────────────────────────────────────────────────────────────┤
│ [全部] [启用] [禁用] [不健康]                                 │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 🟢 OpenAI API     启用   模型: 3   权重: 100           │ │
│ │    最后检测: 2分钟前  响应: 230ms  成功率: 98.6%      │ │
│ │    [检测] [编辑] [禁用]                                │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ 🔴 DeepSeek API    禁用   模型: 1   权重: 50          │ │
│ │    最后检测: 5分钟前  错误: Connection timeout        │ │
│ │    [启用] [编辑] [删除]                                │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 状态指示灯

| 状态 | 条件 | 颜色 |
|------|------|------|
| 健康 | `is_healthy = true` 且 `failure_count = 0` | 绿色 🟢 |
| 降级 | `is_healthy = true` 且 `failure_count > 0` | 黄色 🟡 |
| 不健康 | `is_healthy = false` | 红色 🔴 |
| 未知 | 未检测过 | 灰色 ⚪ |

### 5.3 手动检测功能

```vue
<template>
  <el-button 
    type="primary" 
    size="small"
    :loading="checking"
    @click="handleCheck(channel.id)"
  >
    {{ checking ? '检测中...' : '检测' }}
  </el-button>
</template>

<script setup>
const handleCheck = async (channelId) => {
  checking.value = true
  try {
    const res = await fetch(`/api/v1/admin/channels/${channelId}/health`, {
      method: 'POST'
    })
    const data = await res.json()
    if (data.success) {
      ElMessage.success('检测成功')
      refreshList()
    } else {
      ElMessage.error(data.message || '检测失败')
    }
  } finally {
    checking.value = false
  }
}
</script>
```

---

## 6. 建议的数据库扩展

如需支持更多配置选项：

```sql
-- 可选：健康检测配置表
CREATE TABLE IF NOT EXISTS health_check_configs (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER REFERENCES channels(id) UNIQUE,
    check_interval_minutes INTEGER DEFAULT 5,
    timeout_seconds INTEGER DEFAULT 30,
    failure_threshold INTEGER DEFAULT 3,
    auto_disable BOOLEAN DEFAULT true,
    auto_enable BOOLEAN DEFAULT false,
    auto_enable_threshold INTEGER DEFAULT 3,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 可选：健康检测历史
CREATE TABLE IF NOT EXISTS channel_health_logs (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER REFERENCES channels(id),
    success BOOLEAN NOT NULL,
    response_time_ms INTEGER,
    error_message TEXT,
    checked_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_health_logs_channel ON channel_health_logs(channel_id);
CREATE INDEX idx_health_logs_time ON channel_health_logs(checked_at);
```

---

## 7. 后续扩展计划

| 功能 | 优先级 | 预计工时 |
|------|--------|----------|
| 前端健康状态显示 | 🔴 高 | 2h |
| 手动检测按钮 | 🔴 高 | 1h |
| 健康详情统计页面 | 🟡 中 | 4h |
| 检测历史记录 | 🟡 中 | 3h |
| 自动恢复功能 | 🟢 低 | 2h |
| 告警通知 | 🟢 低 | 4h |

---

## 8. 代码参考

### 核心文件位置

```
backend/internal/
├── service/
│   └── health_check.go      # 健康检测服务 (304行)
├── model/
│   └── channel.go           # 渠道模型 (138行)
├── repository/
│   └── channel_repository.go
└── pkg/adapter/
    ├── factory.go           # 适配器工厂
    ├── openai.go
    ├── azure.go
    ├── claude.go
    ├── gemini.go
    ├── deepseek.go
    └── ... (其他适配器)
```

### 关键方法

| 方法 | 文件 | 说明 |
|------|------|------|
| `Start()` | health_check.go | 启动健康检测服务 |
| `Stop()` | health_check.go | 停止服务 |
| `CheckChannelManually()` | health_check.go | 手动触发检测 |
| `GetChannelStatus()` | health_check.go | 获取渠道状态 |
| `GetStats()` | health_check.go | 获取统计信息 |
| `GetActiveChannels()` | channel_repository.go | 获取活跃渠道 |
| `UpdateHealthStatus()` | channel_repository.go | 更新健康状态 |
| `IncrementFailureCount()` | channel_repository.go | 增加失败计数 |
| `ResetFailureCount()` | channel_repository.go | 重置失败计数 |
| `UpdateResponseTime()` | channel_repository.go | 更新响应时间 |

---

## 审核意见

### v1.0 待审核 → v2.0 审核通过 (2026-04-01)

1. ✅ 核心健康检测逻辑已实现
2. ✅ 支持所有适配器类型
3. ✅ 自动禁用失败渠道
4. 🔴 前端需添加健康状态显示
5. 🔴 前端需添加手动检测按钮
6. 🟡 建议添加自动恢复功能
7. 🟢 建议添加检测历史记录

### 下一步行动

1. **前端开发**: 渠道列表页面添加健康状态列
2. **API确认**: 验证手动检测API是否正常
3. **测试**: 模拟渠道失败场景，验证自动禁用逻辑

---

## 附录：适配器官方API地址

| 渠道类型 | 官方API地址 |
|----------|-------------|
| OpenAI | https://api.openai.com/v1 |
| Claude | https://api.anthropic.com |
| Gemini | https://generativelanguage.googleapis.com/v1beta |
| DeepSeek | https://api.deepseek.com/v1 |
| Azure | 由Azure门户提供 |
| NVIDIA | https://integrate.api.nvidia.com/v1 |
| Groq | https://api.groq.com/openai/v1 |
| Ollama | http://localhost:11434 (本地) |
