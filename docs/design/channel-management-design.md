# 渠道管理功能设计方案

> 版本: v2.0
> 日期: 2026-04-13
> 状态: ✅ 已实现

---

## 1. 功能概述

渠道管理允许管理员配置多个API渠道（如OpenAI、Claude、DeepSeek等），支持负载均衡、故障转移和动态权重调整。

## 2. 业务模型

### 2.1 渠道类型

| 类型 | 说明 | 示例 |
|------|------|------|
| openai | OpenAI兼容API | chat.openai.com, Azure OpenAI |
| claude | Anthropic Claude | api.anthropic.com |
| deepseek | DeepSeek | api.deepseek.com |
| custom | 自定义 | 其他兼容API |

### 2.2 渠道状态

| 状态 | 说明 | 影响 |
|------|------|------|
| enabled | 启用 | 正常路由流量 |
| disabled | 禁用 | 不路由流量，可手动启用 |
| unhealthy | 不健康 | 自动禁用，需排查后手动启用 |

### 2.3 渠道分组

用于区分不同用途的渠道：
- `default` - 默认组
- `premium` - 高级渠道（VIP专用）
- `backup` - 备用渠道

## 3. 数据模型

### 3.1 channels 表扩展字段

```sql
ALTER TABLE channels ADD COLUMN IF NOT EXISTS channel_type VARCHAR(20) DEFAULT 'openai';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS group_name VARCHAR(50) DEFAULT 'default';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS is_healthy BOOLEAN DEFAULT true;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS health_check_url VARCHAR(500);
ALTER TABLE channels ADD COLUMN IF NOT EXISTS health_check_interval INTEGER DEFAULT 60;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS consecutive_failures INTEGER DEFAULT 0;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS auto_disable_threshold INTEGER DEFAULT 5;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS proxy_enabled BOOLEAN DEFAULT false;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS proxy_type VARCHAR(20) DEFAULT 'none';
ALTER TABLE channels ADD COLUMN IF NOT EXISTS proxy_url VARCHAR(500);
```

### 3.2 渠道模型

```go
type Channel struct {
    ID              uint
    Name            string           // 渠道名称，如 "OpenAI API"
    ChannelType     string           // openai/claude/deepseek/custom
    BaseURL         string           // API基础URL
    APIKeyEncrypted string           // 加密的API Key
    Models          string           // 支持的模型列表JSON
    GroupName       string           // 渠道分组
    Weight          int              // 权重（用于负载均衡）
    Priority        int              // 优先级（数字越小优先级越高）
    Status          string           // enabled/disabled/unhealthy
    IsHealthy       bool             // 健康状态
    HealthCheckURL  string          // 健康检查URL
    RPMLimit        int              // 每分钟请求限制
    TPMLimit        int              // 每分钟Token限制
    CostFactor      float64          // 成本系数
    
    // 代理支持 (v2.0)
    ProxyEnabled    bool             // 是否启用代理
    ProxyType       string           // socks5/http/none
    ProxyURL        string           // 代理地址
    
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

## 4. 功能模块

### 4.1 渠道列表

**功能**：
- 显示所有渠道卡片/列表
- 支持分组筛选
- 支持状态筛选（启用/禁用/不健康）
- 支持搜索（名称、类型）
- 显示健康状态指示灯

**UI布局**：
```
┌─────────────────────────────────────────────────────────────┐
│ 渠道管理                                    [+ 添加渠道]    │
├─────────────────────────────────────────────────────────────┤
│ [全部] [默认组] [高级组] [备用组]   [启用] [禁用] [全部]  │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐│
│ │ ● OpenAI API   │  │ ● Claude API   │  │ ● DeepSeek API ││
│ │   状态: 启用   │  │   状态: 启用   │  │   状态: 禁用   ││
│ │   模型: 3个   │  │   模型: 2个   │  │   模型: 1个   ││
│ │   权重: 10   │  │   权重: 5     │  │   权重: 3     ││
│ │ [编辑] [禁用] │  │ [编辑] [禁用] │  │ [编辑] [启用] ││
│ └─────────────────┘  └─────────────────┘  └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### 4.2 添加/编辑渠道

**表单字段**：
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 渠道名称 | 输入框 | 是 | 如 "OpenAI API" |
| 渠道类型 | 下拉框 | 是 | openai/claude/deepseek/custom |
| API类型 | 下拉框 | 是 | official/custom |
| Base URL | 输入框 | 是 | API基础地址 |
| API Key | 密码框 | 是 | 加密存储 |
| 支持模型 | 多选框 | 是 | 根据类型显示可选模型 |
| 渠道分组 | 下拉框 | 否 | default/premium/backup |
| 权重 | 数字输入 | 否 | 默认10，用于负载均衡 |
| 优先级 | 数字输入 | 否 | 默认0，数字越小越优先 |
| 请求限制-RPM | 数字输入 | 否 | 每分钟最大请求数 |
| 请求限制-TPM | 数字输入 | 否 | 每分钟最大Token数 |
| 成本系数 | 数字输入 | 否 | 默认1.0，用于成本计算 |
| 健康检查URL | 输入框 | 否 | 自定义健康检查地址 |

**代理设置 (v2.0)**：
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 启用代理 | 开关 | 否 | 是否通过代理访问 |
| 代理类型 | 下拉框 | 否 | socks5/http |
| 代理地址 | 输入框 | 否 | 如 `socks5://user:pass@host:port` |

**验证规则**：
- Base URL 必须以 `https://` 开头
- API Key 不能为空
- 至少选择一个模型
- 如果启用代理，代理地址不能为空

### 4.3 渠道测试

**功能**：
- 点击"测试"按钮，发送测试请求到渠道
- 显示响应时间、状态码、响应内容
- 成功/失败状态提示

**测试请求**：
```json
{
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hi"}],
  "max_tokens": 10
}
```

### 4.4 健康检测

**定时任务**：
- 每5分钟检测所有启用渠道
- 连续失败3次标记为 `unhealthy`
- 失败时自动增加失败计数
- 成功时重置失败计数

**检测逻辑**：
1. 调用渠道的 ListModels 接口
2. 检查响应状态码是否为 200
3. 验证响应时间和内容
4. 记录失败次数

## 5. API设计

### 5.1 渠道管理API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/channels` | GET | 获取渠道列表 |
| `/api/v1/admin/channels` | POST | 创建渠道 |
| `/api/v1/admin/channels/:id` | GET | 获取渠道详情 |
| `/api/v1/admin/channels/:id` | PUT | 更新渠道 |
| `/api/v1/admin/channels/:id` | DELETE | 删除渠道 |
| `/api/v1/admin/channels/:id/enable` | POST | 启用渠道 |
| `/api/v1/admin/channels/:id/disable` | POST | 禁用渠道 |
| `/api/v1/admin/channels/:id/test` | POST | 测试渠道 |

### 5.2 响应格式

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "OpenAI API",
    "channel_type": "openai",
    "base_url": "https://api.openai.com",
    "models": ["gpt-3.5-turbo", "gpt-4"],
    "group_name": "default",
    "weight": 10,
    "status": "enabled",
    "is_healthy": true
  }
}
```

## 6. 用户流程

### 6.1 添加渠道流程

```
1. 管理员进入"渠道管理"页面
2. 点击"+ 添加渠道"
3. 填写渠道信息
4. 点击"保存"
5. 系统验证并保存
6. 自动进行健康检查
7. 显示渠道列表
```

### 6.2 故障处理流程

```
1. 健康检测发现渠道故障
2. 记录失败次数
3. 连续失败3次 → 标记为unhealthy
4. 自动禁用渠道
5. 发送通知（可选）
6. 管理员收到通知
7. 排查问题并修复
8. 手动启用渠道
```

## 7. 安全考虑

- API Key 必须加密存储
- 日志中脱敏显示API Key
- 渠道删除需二次确认
- 敏感操作记录审计日志

## 8. 后续扩展

- [ ] 渠道自动扩容建议
- [ ] 成本分析和报表
- [ ] 渠道使用统计
- [ ] 多租户隔离

---

## 审核意见

（待填写）
