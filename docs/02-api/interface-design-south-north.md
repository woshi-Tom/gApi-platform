# API Proxy Platform - 接口细分设计文档 v1.0

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 接口划分概述

### 1.1 三大接口体系

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          接口体系架构图                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│                              ┌─────────────────┐                          │
│                              │   互联网用户     │                          │
│                              └────────┬────────┘                          │
│                                         │                                   │
│                                         ▼                                   │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                      北向接口 (Northbound API)                          │ │
│  │                    ────────────────────────────                         │ │
│  │  · 用户注册/登录                                                        │ │
│  │  · Token管理                                                           │ │
│  │  · 商品浏览/购买                                                       │ │
│  │  · 账户充值/VIP                                                        │ │
│  │  · 用量查询                                                            │ │
│  │  · 调用OpenAI兼容API                                                  │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                         │                                   │
│                                         ▼                                   │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                      南向接口 (Southbound API)                         │ │
│  │                    ────────────────────────────                         │ │
│  │  · 渠道管理 (CRUD)                                                    │ │
│  │  · 渠道健康检查                                                        │ │
│  │  · 渠道测试                                                            │ │
│  │  · 上游API调用                                                         │ │
│  │  · 负载均衡                                                            │ │
│  │  · 模型映射                                                            │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                         │                                   │
│                                         ▼                                   │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                      管理后台接口 (Admin API)                          │ │
│  │                    ────────────────────────────                         │ │
│  │  · 超级管理员登录 (内网)                                              │ │
│  │  · 商品管理                                                            │ │
│  │  · 用户管理                                                            │ │
│  │  · 订单管理                                                            │ │
│  │  · 审计日志                                                            │ │
│  │  · 系统配置                                                            │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 接口对比表

| 维度 | 北向接口 (用户) | 南向接口 (内部) | 管理后台 |
|------|---------------|----------------|---------|
| **访问者** | 外部用户、开发者 | 系统内部 | 管理员 (内网) |
| **认证方式** | Token/API Key | 内部Token | JWT + 内网IP |
| **主要功能** | 消费API服务 | 管理上游渠道 | 平台运营管理 |
| **流量方向** | 出 (用户→平台) | 入 (平台→上游) | 管 (管理平台) |
| **频率** | 高并发 | 中等 | 低 |
| **安全等级** | 中 | 高 | 极高 |

---

## 2. 北向接口 (Northbound API) - 面向用户

### 2.1 接口定位

**北向接口**是平台提供给外部用户和开发者使用的API，主要功能是：
- 让用户获取API调用凭证 (Token)
- 让用户调用AI服务 (OpenAI兼容格式)
- 让用户管理自己的账户和配额
- 让用户购买商品和VIP

### 2.2 接口清单

#### 2.2.1 用户认证模块

```
┌─────────────────────────────────────────────────────────────────┐
│                    用户认证模块 (/api/v1/auth)                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  POST   /register          用户注册 (赠送配额)                  │
│  POST   /login             用户登录                             │
│  POST   /logout            用户登出                             │
│  POST   /refresh           刷新Token                           │
│  GET    /verify-email      邮箱验证                             │
│  POST   /forgot-password   忘记密码                             │
│  PUT    /reset-password    重置密码                             │
│  PUT    /password          修改密码                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**注册接口 (赠送配额)**

```
POST /api/v1/auth/register
Content-Type: application/json

请求:
{
    "username": "newuser",
    "email": "user@example.com", 
    "password": "SecurePass123!"
}

响应 (成功):
{
    "code": 0,
    "data": {
        "user_id": 1001,
        "username": "newuser",
        "email": "user@example.com",
        "quota": 100000,              // 获赠配额
        "quota_type": "permanent",
        "is_trial_vip": false,
        "need_verify_email": true
    }
}
```

#### 2.2.2 用户Token管理模块

```
┌─────────────────────────────────────────────────────────────────┐
│                  Token管理模块 (/api/v1/tokens)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  GET    /                获取Token列表                          │
│  POST   /                创建Token                              │
│  GET    /:id             获取Token详情                          │
│  PUT    /:id             更新Token                              │
│  DELETE /:id             删除Token                              │
│                                                                 │
│  GET    /:id/usage       获取Token用量明细                      │
│  POST   /:id/reset       重置Token配额                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**创建Token**

```
POST /api/v1/tokens
Authorization: Bearer <user_token>

请求:
{
    "name": "My API Key",
    "allowed_models": ["gpt-3.5-turbo", "gpt-4"],
    "allowed_ips": ["192.168.1.100"],
    "expires_at": "2027-12-31T23:59:59Z"
}

响应 (仅显示一次):
{
    "code": 0,
    "data": {
        "id": 1,
        "name": "My API Key",
        "token_key": "sk-ap-abc123...xyz",     // ⚠️ 仅此次显示
        "remain_quota": 100000,
        "allowed_models": ["gpt-3.5-turbo", "gpt-4"],
        "allowed_ips": ["192.168.1.100"],
        "expires_at": "2027-12-31T23:59:59Z",
        "created_at": "2026-03-23T12:00:00Z"
    }
}
```

#### 2.2.3 用户账户与配额模块

```
┌─────────────────────────────────────────────────────────────────┐
│                  用户账户模块 (/api/v1/account)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  GET    /profile           获取账户信息                         │
│  PUT    /profile           更新账户信息                         │
│  GET    /quota             获取配额信息                         │
│  GET    /usage             获取用量明细                         │
│  GET    /usage/daily       获取每日用量                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**获取配额信息**

```
GET /api/v1/account/quota
Authorization: Bearer <user_token>

响应:
{
    "code": 0,
    "data": {
        "user_id": 1001,
        "username": "newuser",
        "level": "vip",
        "is_vip": true,
        "vip_expired_at": "2026-04-22T00:00:00Z",
        
        // 永久配额
        "permanent_quota": 500000,
        "permanent_used": 100000,
        "permanent_remain": 400000,
        
        // VIP配额 (30天有效)
        "vip_quota": 1000000,
        "vip_used": 50000,
        "vip_remain": 950000,
        
        // 今日用量
        "used_today": 5000,
        "limit_today": 100000,
        
        // 可用渠道
        "channel_count": 5
    }
}
```

#### 2.2.4 商品与订单模块

```
┌─────────────────────────────────────────────────────────────────┐
│                  商品订单模块 (/api/v1/products, /orders)       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  商品:                                                           │
│  GET    /products           获取商品列表 (仅已上架)             │
│  GET    /products/:id       获取商品详情                        │
│                                                                 │
│  订单:                                                           │
│  GET    /orders              获取订单列表                        │
│  GET    /orders/:id         获取订单详情                        │
│  POST   /orders              创建订单                           │
│                                                                 │
│  支付:                                                           │
│  GET    /orders/:id/pay     获取支付信息                        │
│  POST   /orders/:id/pay     发起支付                            │
│  GET    /pay/:id/status     查询支付状态                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 2.2.5 VIP与充值模块

```
┌─────────────────────────────────────────────────────────────────┐
│                  VIP与充值模块 (/api/v1/vip, /recharge)          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  VIP:                                                           │
│  GET    /vip/packages       获取VIP套餐列表                      │
│  GET    /vip/status         获取VIP状态                         │
│  GET    /vip/orders         获取VIP订单列表                      │
│  POST   /vip/orders         购买VIP                              │
│                                                                 │
│  充值:                                                           │
│  GET    /recharge/packages  获取充值套餐列表                    │
│  GET    /recharge/orders    获取充值订单列表                    │
│  POST   /recharge/orders    创建充值订单                        │
│                                                                 │
│  兑换:                                                           │
│  POST   /redemption         使用兑换码                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 核心：OpenAI兼容API (北向)

这是北向接口的核心，用户通过此接口调用AI服务。

```
┌─────────────────────────────────────────────────────────────────┐
│               OpenAI兼容API (/v1) - 核心业务                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  聊天补全 (最常用)                                              │
│  POST   /chat/completions                                       │
│                                                                 │
│  文本补全                                                       │
│  POST   /completions                                            │
│                                                                 │
│  Embeddings                                                     │
│  POST   /embeddings                                             │
│                                                                 │
│  模型列表                                                       │
│  GET    /models                                                 │
│                                                                 │
│  文件管理 (GPT-4V)                                              │
│  POST   /files                                                  │
│  GET    /files                                                  │
│  DELETE /files/:id                                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**聊天补全接口流程**

```
┌─────────────────────────────────────────────────────────────────┐
│                   聊天API调用流程                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用户请求                                                        │
│  POST /v1/chat/completions                                      │
│  Authorization: Bearer sk-ap-xxx                                │
│  {                                                              │
│      "model": "gpt-3.5-turbo",                                 │
│      "messages": [{"role": "user", "content": "Hello!"}]       │
│  }                                                              │
│         │                                                       │
│         ▼                                                       │
│  1. 验证Token                                                   │
│     ├── Token存在且有效                                         │
│     ├── Token未过期                                             │
│     ├── Token配额充足                                           │
│     └── IP在白名单中(如果配置)                                   │
│         │                                                       │
│         ▼                                                       │
│  2. 检查用户配额                                                 │
│     ├── VIP配额优先使用                                          │
│     ├── 配额充足则继续                                          │
│     └── 配额不足返回429                                         │
│         │                                                       │
│         ▼                                                       │
│  3. 选择渠道 (负载均衡)                                          │
│     ├── 按权重分配                                              │
│     ├── 优先健康渠道                                            │
│     └── 考虑用户VIP等级                                         │
│         │                                                       │
│         ▼                                                       │
│  4. 调用上游API                                                 │
│     ├── OpenAI格式转换                                          │
│     └── 转发到上游渠道                                          │
│         │                                                       │
│         ▼                                                       │
│  5. 记录用量                                                    │
│     ├── 扣减配额                                                │
│     ├── 记录usage_log                                           │
│     └── 记录quota_transaction                                    │
│         │                                                       │
│         ▼                                                       │
│  6. 返回响应                                                    │
│     {                                                          │
│         "id": "chatcmpl-xxx",                                   │
│         "choices": [{"message": {"content": "Hello!"}}]        │
│     }                                                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**流式响应支持**

```
POST /v1/chat/completions
Content-Type: application/json

请求:
{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "讲个故事"}],
    "stream": true
}

响应 (SSE):
data: {"id":"chatcmpl-abc","choices":[{"delta":{"role":"assistant","content":"从前"},"index":0}]}

data: {"id":"chatcmpl-abc","choices":[{"delta":{"content":"有"},"index":0}]}

data: {"id":"chatcmpl-abc","choices":[{"delta":{"content":"一","index":0}]}

data: {"id":"chatcmpl-abc","choices":[{"delta":{},"finish_reason":"stop"}]}
```

---

## 3. 南向接口 (Southbound API) - 面向内部/渠道管理

### 3.1 接口定位

**南向接口**是平台内部用于管理和调用下游服务提供商的API，主要功能是：
- 渠道配置和管理 (连接OpenAI/Azure/Claude等)
- 渠道健康检查和状态监控
- 渠道测试验证
- 负载均衡和路由策略
- 模型映射和转换

### 3.2 接口清单

#### 3.2.1 渠道管理模块

```
┌─────────────────────────────────────────────────────────────────┐
│                渠道管理模块 (/api/v1/internal/channels)         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  渠道CRUD:                                                      │
│  GET    /channels           获取渠道列表 (内部)                  │
│  POST   /channels          创建渠道                            │
│  GET    /channels/:id      获取渠道详情                         │
│  PUT    /channels/:id      更新渠道                            │
│  DELETE /channels/:id      删除渠道                            │
│                                                                 │
│  渠道控制:                                                      │
│  POST   /channels/:id/enable    启用渠道                        │
│  POST   /channels/:id/disable  禁用渠道                        │
│  POST   /channels/:id/test     测试渠道 ⭐                      │
│  GET    /channels/:id/health   健康状态                        │
│                                                                 │
│  渠道分组:                                                      │
│  GET    /channel-groups     获取分组列表                        │
│  POST   /channel-groups     创建分组                           │
│  DELETE /channel-groups/:id 删除分组                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.2.2 渠道测试接口 ⭐ (核心南向功能)

这是南向接口最重要的功能，用于验证渠道是否可用。

```
┌─────────────────────────────────────────────────────────────────┐
│               渠道测试模块 (/api/v1/internal/channels/:id/test) │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  测试类型:                                                      │
│  1. models    - 获取模型列表                                     │
│  2. chat     - 对话补全测试                                      │
│  3. embeddings - 向量计算测试                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**测试请求/响应**

```
POST /api/v1/internal/channels/1/test
Authorization: Bearer <internal_token>
Content-Type: application/json

// 测试1: 获取模型列表
{
    "test_type": "models"
}

响应:
{
    "code": 0,
    "data": {
        "success": true,
        "response_time_ms": 234,
        "status_code": 200,
        "models": [
            "gpt-4",
            "gpt-4-0613", 
            "gpt-3.5-turbo",
            "gpt-3.5-turbo-0613"
        ]
    }
}

// 测试2: 对话补全
{
    "test_type": "chat",
    "model": "gpt-3.5-turbo",
    "messages": [
        {"role": "user", "content": "Hello!"}
    ],
    "temperature": 0.7,
    "max_tokens": 100
}

响应:
{
    "code": 0,
    "data": {
        "success": true,
        "response_time_ms": 1234,
        "status_code": 200,
        "content": "Hello! How can I help you today?",
        "usage": {
            "prompt_tokens": 20,
            "completion_tokens": 15,
            "total_tokens": 35
        }
    }
}

// 测试3: Embeddings
{
    "test_type": "embeddings",
    "model": "text-embedding-ada-002",
    "input": "The food was delicious"
}

响应:
{
    "code": 0,
    "data": {
        "success": true,
        "response_time_ms": 456,
        "status_code": 200,
        "embedding": [0.0023, -0.0098, 0.0211, ...],  // 1536维
        "embedding_dim": 1536
    }
}
```

#### 3.2.3 渠道健康检查模块

```
┌─────────────────────────────────────────────────────────────────┐
│              渠道健康检查 (/api/v1/internal/health)              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  健康状态:                                                      │
│  GET    /health                    服务健康状态                 │
│  GET    /health/channels           所有渠道健康状态             │
│  GET    /health/channels/:id       特定渠道健康状态             │
│  GET    /health/channels/:id/history 健康状态历史               │
│                                                                 │
│  自动检查:                                                      │
│  POST   /health/check              手动触发健康检查             │
│  GET    /health/check/stats        检查统计                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**健康检查响应**

```
GET /api/v1/internal/health/channels

响应:
{
    "code": 0,
    "data": [
        {
            "channel_id": 1,
            "channel_name": "OpenAI Primary",
            "status": "healthy",
            "is_healthy": true,
            "failure_count": 0,
            "last_success_at": "2026-03-23T12:00:00Z",
            "last_check_at": "2026-03-23T12:05:00Z",
            "response_time_avg": 450,
            "success_rate_1h": 0.998,
            "success_rate_24h": 0.995
        },
        {
            "channel_id": 2,
            "channel_name": "Azure Secondary", 
            "status": "unhealthy",
            "is_healthy": false,
            "failure_count": 5,
            "last_success_at": "2026-03-22T18:00:00Z",
            "last_check_at": "2026-03-23T12:05:00Z",
            "response_time_avg": 0,
            "success_rate_1h": 0.85,
            "last_error": "401 Unauthorized"
        }
    ]
}
```

#### 3.2.4 负载均衡与路由模块

```
┌─────────────────────────────────────────────────────────────────┐
│               负载均衡模块 (/api/v1/internal/loadbalance)      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  路由策略:                                                      │
│  GET    /routing/strategies     获取路由策略列表                │
│  POST   /routing/strategies     创建路由策略                    │
│  DELETE /routing/strategies/:id 删除路由策略                    │
│                                                                 │
│  权重管理:                                                      │
│  GET    /channels/:id/weight    获取渠道权重                    │
│  PUT    /channels/:id/weight    修改渠道权重                   │
│  POST   /channels/batch-weight  批量修改权重                   │
│                                                                 │
│  优先级管理:                                                    │
│  GET    /channels/priorities    获取优先级                     │
│  PUT    /channels/priorities    修改优先级                     │
│                                                                 │
│  智能路由:                                                      │
│  GET    /routing/recommend?model=gpt-4  推荐最优渠道            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.2.5 模型映射模块

```
┌─────────────────────────────────────────────────────────────────┐
│               模型映射模块 (/api/v1/internal/model-mapping)      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  映射规则:                                                      │
│  GET    /model-mappings           获取映射规则                  │
│  POST   /model-mappings           创建映射规则                  │
│  PUT    /model-mappings/:id       更新映射规则                  │
│  DELETE /model-mappings/:id       删除映射规则                  │
│                                                                 │
│  应用场景:                                                      │
│  · 用户请求 "gpt-4" -> 映射到 "gpt-4-0613"                     │
│  · 用户请求 "gpt-3.5" -> 映射到 "gpt-3.5-turbo"               │
│  · 不同渠道模型名称不同，通过映射统一                           │
│                                                                 │
│  示例:                                                          │
│  {                                                             │
│      "source_model": "gpt-4",                                   │
│      "target_channel_id": 1,                                   │
│      "target_model": "gpt-4-0613",                              │
│      "priority": 100                                           │
│  }                                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.2.6 上游API调用模块 (内部)

```
┌─────────────────────────────────────────────────────────────────┐
│               上游调用模块 (/api/v1/internal/upstream)          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  调用转发:                                                      │
│  POST   /upstream/chat          转发聊天请求                    │
│  POST   /upstream/completions   转发补全请求                    │
│  POST   /upstream/embeddings    转发Embeddings请求              │
│  GET    /upstream/models        获取上游模型列表                │
│                                                                 │
│  响应处理:                                                      │
│  · 格式转换 (OpenAI -> 平台格式)                                │
│  · 错误标准化                                                   │
│  · 超时处理                                                     │
│  · 重试机制                                                     │
│                                                                 │
│  日志记录:                                                      │
│  GET    /upstream/logs         调用日志                         │
│  GET    /upstream/logs/:id     详细日志                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.2.7 渠道统计模块

```
┌─────────────────────────────────────────────────────────────────┐
│               渠道统计模块 (/api/v1/internal/stats)             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用量统计:                                                      │
│  GET    /stats/channels              渠道用量统计               │
│  GET    /stats/channels/:id          单渠道用量                  │
│  GET    /stats/models                模型用量统计               │
│  GET    /stats/usage trend           用量趋势                   │
│                                                                 │
│  成本分析:                                                      │
│  GET    /stats/cost                  成本统计                   │
│  GET    /stats/cost/channels         渠道成本                   │
│  GET    /stats/cost/daily            每日成本                   │
│                                                                 │
│  性能分析:                                                      │
│  GET    /stats/performance          性能统计                   │
│  GET    /stats/performance/response  响应时间                  │
│  GET    /stats/performance/errors   错误率                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 管理后台接口 (Admin API) - 面向管理员

### 4.1 接口定位

**管理后台接口**是平台运营管理使用的API，只对内网IP开放，主要功能是：
- 超级管理员登录 (内网专属)
- 商品上下架管理
- 用户管理
- 订单管理
- 财务对账
- 审计日志

### 4.2 接口清单

#### 4.2.1 管理员认证 (内网专属)

```
┌─────────────────────────────────────────────────────────────────┐
│              管理员认证 (/api/v1/admin/auth) - 内网             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  限制:                                                          │
│  · 仅内网IP可访问 (10.x.x.x / 172.16.x.x / 192.168.x.x)        │
│  · 强制要求强密码                                                │
│  · 登录验证码                                                    │
│  · 双因素认证 (可选)                                             │
│                                                                 │
│  接口:                                                          │
│  POST   /login              管理员登录                         │
│  POST   /logout             管理员登出                          │
│  GET    /me                 当前管理员信息                      │
│  PUT    /password           修改密码                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.2.2 仪表盘

```
┌─────────────────────────────────────────────────────────────────┐
│               仪表盘 (/api/v1/admin/dashboard)                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  概览:                                                          │
│  GET    /stats              核心统计数据                         │
│  GET    /stats/realtime    实时数据                             │
│  GET    /stats/users        用户统计                            │
│  GET    /stats/revenue      营收统计                             │
│  GET    /stats/usage        用量统计                            │
│                                                                 │
│  图表:                                                          │
│  GET    /chart/trend        趋势图表                             │
│  GET    /chart/distribution 分布图表                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**核心统计**

```
GET /api/v1/admin/dashboard/stats

响应:
{
    "code": 0,
    "data": {
        "users": {
            "total": 10000,
            "today_new": 50,
            "active_7d": 3000,
            "vip": 500
        },
        "channels": {
            "total": 20,
            "healthy": 18,
            "unhealthy": 2
        },
        "orders": {
            "today_count": 100,
            "today_amount": 5000.00,
            "month_amount": 150000.00
        },
        "usage": {
            "tokens_today": 50000000,
            "tokens_month": 1500000000
        }
    }
}
```

#### 4.2.3 商品管理 ⭐

```
┌─────────────────────────────────────────────────────────────────┐
│               商品管理 (/api/v1/admin/products)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  商品:                                                          │
│  GET    /products           获取商品列表                        │
│  POST   /products           创建商品                            │
│  GET    /products/:id       获取商品详情                        │
│  PUT    /products/:id       更新商品                            │
│  DELETE /products/:id       删除商品                            │
│                                                                 │
│  上架/下架 ⭐:                                                  │
│  POST   /products/:id/publish     上架商品                      │
│  POST   /products/:id/unpublish   下架商品                      │
│                                                                 │
│  分类:                                                          │
│  GET    /categories        获取分类列表                        │
│  POST   /categories        创建分类                            │
│  PUT    /categories/:id    更新分类                            │
│  DELETE /categories/:id    删除分类                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.2.4 用户管理

```
┌─────────────────────────────────────────────────────────────────┐
│               用户管理 (/api/v1/admin/users)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  GET    /users              获取用户列表                        │
│  POST   /users              创建用户                            │
│  GET    /users/:id          获取用户详情                        │
│  PUT    /users/:id          更新用户                            │
│  DELETE /users/:id          删除用户                            │
│                                                                 │
│  状态控制:                                                      │
│  POST   /users/:id/enable   启用用户                            │
│  POST   /users/:id/disable  禁用用户                            │
│                                                                 │
│  配额管理:                                                      │
│  POST   /users/:id/quota/add    增加配额                        │
│  POST   /users/:id/quota/deduct 扣减配额                        │
│                                                                 │
│  VIP管理:                                                       │
│  POST   /users/:id/vip/activate    开通VIP                      │
│  POST   /users/:id/vip/deactivate  关闭VIP                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.2.5 订单管理

```
┌─────────────────────────────────────────────────────────────────┐
│               订单管理 (/api/v1/admin/orders)                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  订单:                                                          │
│  GET    /orders              获取订单列表                        │
│  GET    /orders/:id          获取订单详情                        │
│  POST   /orders/:id/cancel  取消订单                            │
│  POST   /orders/:id/refund  退款                                │
│                                                                 │
│  支付:                                                          │
│  GET    /payments           支付记录列表                        │
│  GET    /payments/:id       支付详情                            │
│                                                                 │
│  对账:                                                          │
│  GET    /reconciliation     财务对账                            │
│  GET    /reconciliation/daily  每日对账                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.2.6 审计日志 ⭐

```
┌─────────────────────────────────────────────────────────────────┐
│               审计日志 (/api/v1/admin/audit-logs)               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  日志查询:                                                      │
│  GET    /audit-logs         获取审计日志列表                    │
│  GET    /audit-logs/:id     获取日志详情                        │
│                                                                 │
│  统计分析:                                                      │
│  GET    /audit-logs/stats   获取统计信息                        │
│  GET    /audit-logs/trend   获取趋势                            │
│                                                                 │
│  导出:                                                          │
│  GET    /audit-logs/export  导出日志                           │
│                                                                 │
│  登录日志:                                                      │
│  GET    /login-logs        获取登录日志                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 4.2.7 系统配置

```
┌─────────────────────────────────────────────────────────────────┐
│               系统配置 (/api/v1/admin/configs)                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  配置管理:                                                      │
│  GET    /configs              获取配置列表                      │
│  GET    /configs/:key         获取单个配置                      │
│  PUT    /configs/:key         更新配置                          │
│                                                                 │
│  注册配置:                                                      │
│  GET    /signup-config        获取注册配置                      │
│  PUT    /signup-config        更新注册配置                      │
│                                                                 │
│  支付配置:                                                      │
│  GET    /payment-config       获取支付配置                      │
│  PUT    /payment-config       更新支付配置                      │
│                                                                 │
│  通道配置:                                                      │
│  GET    /channel-templates   获取可用渠道模板                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. 接口认证方式对比

### 5.1 认证矩阵

| 接口类型 | 认证方式 | Token类型 | 有效期 | IP限制 |
|---------|---------|----------|-------|--------|
| **北向-用户** | Bearer Token | API Key (sk-*) | 长期 | 可选白名单 |
| **北向-API** | Bearer Token | 用户Token | 短期 | 可选白名单 |
| **南向-内部** | Bearer Token | 内部服务Token | 短期 | **固定IP** |
| **管理后台** | JWT + Session | 管理员Token | 短期 | **内网** |

### 5.2 内部Token示例

```go
// 南向接口内部Token
type InternalToken struct {
    ServiceID   string    `json:"service_id"`    // 服务ID
    ServiceName string    `json:"service_name"`  // 服务名称
    IssuedAt    time.Time `json:"issued_at"`
    ExpiresAt   time.Time `json:"expires_at"`
    Permissions []string  `json:"permissions"`     // channel:read, channel:write...
}

// 内部Token示例
// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
// payload: {
//     "service_id": "channel-manager",
//     "service_name": "Channel Management Service", 
//     "permissions": ["channel:read", "channel:write", "channel:test"],
//     "issued_at": "2026-03-23T12:00:00Z",
//     "expires_at": "2026-03-23T13:00:00Z"
// }
```

---

## 6. 完整接口清单汇总

### 6.1 北向接口汇总

| 模块 | 路径 | 方法 | 说明 |
|-----|------|-----|------|
| 认证 | `/api/v1/auth/register` | POST | 注册(赠配额) |
| 认证 | `/api/v1/auth/login` | POST | 登录 |
| 认证 | `/api/v1/auth/logout` | POST | 登出 |
| Token | `/api/v1/tokens` | GET/POST | Token列表/创建 |
| Token | `/api/v1/tokens/:id` | GET/PUT/DELETE | Token操作 |
| 账户 | `/api/v1/account/profile` | GET/PUT | 账户信息 |
| 账户 | `/api/v1/account/quota` | GET | 配额信息 |
| 账户 | `/api/v1/account/usage` | GET | 用量明细 |
| 商品 | `/api/v1/products` | GET | 商品列表 |
| 商品 | `/api/v1/products/:id` | GET | 商品详情 |
| 订单 | `/api/v1/orders` | GET/POST | 订单列表/创建 |
| 订单 | `/api/v1/orders/:id` | GET | 订单详情 |
| VIP | `/api/v1/vip/packages` | GET | VIP套餐 |
| VIP | `/api/v1/vip/status` | GET | VIP状态 |
| **核心** | `/api/v1/chat/completions` | POST | 聊天API |
| **核心** | `/api/v1/completions` | POST | 补全API |
| **核心** | `/api/v1/embeddings` | POST | 向量API |
| **核心** | `/api/v1/models` | GET | 模型列表 |

### 6.2 南向接口汇总

| 模块 | 路径 | 方法 | 说明 |
|-----|------|-----|------|
| 渠道管理 | `/api/v1/internal/channels` | GET/POST | 渠道CRUD |
| 渠道管理 | `/api/v1/internal/channels/:id` | GET/PUT/DELETE | 渠道操作 |
| 渠道测试 | `/api/v1/internal/channels/:id/test` | POST | **测试API** |
| 渠道健康 | `/api/v1/internal/health/channels` | GET | 健康状态 |
| 渠道健康 | `/api/v1/internal/channels/:id/health` | GET | 单渠道健康 |
| 负载均衡 | `/api/v1/internal/channels/:id/weight` | GET/PUT | 权重管理 |
| 模型映射 | `/api/v1/internal/model-mappings` | CRUD | 映射规则 |
| 上游调用 | `/api/v1/internal/upstream/chat` | POST | 转发聊天 |
| 统计 | `/api/v1/internal/stats/channels` | GET | 渠道统计 |
| 统计 | `/api/v1/internal/stats/cost` | GET | 成本统计 |

### 6.3 管理后台接口汇总

| 模块 | 路径 | 方法 | 说明 |
|-----|------|-----|------|
| 管理员 | `/api/v1/admin/auth/login` | POST | 登录(内网) |
| 仪表盘 | `/api/v1/admin/dashboard/stats` | GET | 核心统计 |
| 商品管理 | `/api/v1/admin/products` | CRUD | 商品CRUD |
| 商品管理 | `/api/v1/admin/products/:id/publish` | POST | **上架** |
| 商品管理 | `/api/v1/admin/products/:id/unpublish` | POST | **下架** |
| 用户管理 | `/api/v1/admin/users` | CRUD | 用户CRUD |
| 用户管理 | `/api/v1/admin/users/:id/quota` | POST | 配额调整 |
| 用户管理 | `/api/v1/admin/users/:id/vip` | POST | VIP管理 |
| 订单管理 | `/api/v1/admin/orders` | CRUD | 订单CRUD |
| 订单管理 | `/api/v1/admin/orders/:id/refund` | POST | 退款 |
| 审计日志 | `/api/v1/admin/audit-logs` | GET | 审计日志 |
| 系统配置 | `/api/v1/admin/configs` | CRUD | 配置管理 |

---

**文档版本**: 1.0  
**下一步**: 开始实现