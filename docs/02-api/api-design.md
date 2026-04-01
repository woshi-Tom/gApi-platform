# API Proxy Platform - 接口设计文档 v1.0

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 规范总览

### 1.1 API 设计原则

| 原则 | 说明 |
|------|------|
| RESTful | 遵循 REST 设计规范 |
| JSON | 请求/响应使用 JSON 格式 |
| 版本化 | URL 包含版本号 `/api/v1/` |
| 认证 | Bearer Token 认证 |
| 幂等性 | POST 请求需要幂等标识 |

### 1.2 Base URL

```
开发环境: http://localhost:8080/api/v1
生产环境: https://api.example.com/api/v1
```

### 1.3 通用请求头

```http
Content-Type: application/json
Authorization: Bearer <token>
X-Request-ID: <uuid>
X-Trace-ID: <trace-id>
```

### 1.4 通用响应格式

```json
// 成功响应
{
    "code": 0,
    "message": "success",
    "data": { ... },
    "request_id": "uuid"
}

// 错误响应
{
    "code": 40001,
    "message": "Invalid parameters",
    "error": "field 'email' is required",
    "request_id": "uuid"
}
```

### 1.5 错误码定义

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40002 | 参数验证失败 |
| 40101 | 未授权 |
| 40102 | Token 过期 |
| 40103 | Token 无效 |
| 40301 | 权限不足 |
| 40302 | 资源不存在 |
| 40401 | 渠道不存在 |
| 40402 | Token 不存在 |
| 40403 | 用户不存在 |
| 40901 | 资源已存在 |
| 42201 | 配额不足 |
| 42202 | VIP 已过期 |
| 42901 | 请求过于频繁 |
| 50001 | 服务器内部错误 |
| 50002 | 第三方服务错误 |

---

## 2. OpenAI 兼容 API

### 2.1 聊天补全

```
POST /v1/chat/completions
```

**请求体:**

```json
{
    "model": "gpt-3.5-turbo",
    "messages": [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello!"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000,
    "stream": false,
    "top_p": 1.0,
    "frequency_penalty": 0.0,
    "presence_penalty": 0.0,
    "user": "user_id_optional"
}
```

**响应体:**

```json
{
    "id": "chatcmpl-123",
    "object": "chat.completion",
    "created": 1677652288,
    "model": "gpt-3.5-turbo-0613",
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": "Hello! How can I help you today?"
            },
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 20,
        "completion_tokens": 15,
        "total_tokens": 35
    }
}
```

### 2.2 文本补全

```
POST /v1/completions
```

**请求体:**

```json
{
    "model": "text-davinci-003",
    "prompt": "Say this is a test",
    "max_tokens": 1000,
    "temperature": 0.7,
    "stream": false
}
```

### 2.3 Embeddings

```
POST /v1/embeddings
```

**请求体:**

```json
{
    "model": "text-embedding-ada-002",
    "input": "The food was delicious and the waiter..."
}
```

**响应体:**

```json
{
    "object": "list",
    "data": [
        {
            "object": "embedding",
            "embedding": [0.0023064255, -0.009327292, ...],
            "index": 0
        }
    ],
    "model": "text-embedding-ada-002",
    "usage": {
        "prompt_tokens": 8,
        "total_tokens": 8
    }
}
```

### 2.4 模型列表

```
GET /v1/models
```

**响应体:**

```json
{
    "object": "list",
    "data": [
        {
            "id": "gpt-4",
            "object": "model",
            "created": 1687882411,
            "owned_by": "openai"
        }
    ]
}
```

---

## 3. 用户端 API

### 3.1 认证

#### 用户注册

```
POST /api/v1/user/auth/register
```

**请求体:**

```json
{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
}
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "user_id": 1,
        "username": "testuser",
        "email": "test@example.com"
    }
}
```

#### 用户登录

```
POST /api/v1/user/auth/login
```

**请求体:**

```json
{
    "email": "test@example.com",
    "password": "SecurePass123!"
}
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_at": "2026-03-24T00:00:00Z",
        "user": {
            "id": 1,
            "username": "testuser",
            "email": "test@example.com",
            "level": "free",
            "remain_quota": 0,
            "vip_quota": 0
        }
    }
}
```

#### 用户登出

```
POST /api/v1/user/auth/logout
```

#### 修改密码

```
PUT /api/v1/user/auth/password
```

**请求体:**

```json
{
    "old_password": "OldPass123!",
    "new_password": "NewPass123!",
    "confirm_password": "NewPass123!"
}
```

#### 刷新 Token

```
POST /api/v1/user/auth/refresh
```

### 3.2 Token 管理

#### 创建 Token

```
POST /api/v1/user/tokens
```

**请求体:**

```json
{
    "name": "My API Key",
    "allowed_models": ["gpt-3.5-turbo", "gpt-4"],
    "allowed_ips": ["192.168.1.1", "10.0.0.0/8"],
    "expires_at": "2027-01-01T00:00:00Z"
}
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "name": "My API Key",
        "token_key": "sk-ap-abc123xyz...",
        "token_key_full": "sk-ap-abc123xyz...fullkey",
        "allowed_models": ["gpt-3.5-turbo", "gpt-4"],
        "allowed_ips": ["192.168.1.1", "10.0.0.0/8"],
        "expires_at": "2027-01-01T00:00:00Z",
        "created_at": "2026-03-23T12:00:00Z"
    }
}
```

**注意:** `token_key_full` 仅在创建时返回一次，之后不再显示。

#### 获取 Token 列表

```
GET /api/v1/user/tokens
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |

#### 更新 Token

```
PUT /api/v1/user/tokens/:id
```

#### 删除 Token

```
DELETE /api/v1/user/tokens/:id
```

### 3.3 配额与用量

#### 获取配额信息

```
GET /api/v1/user/quota
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "remain_quota": 500000,
        "vip_quota": 1000000,
        "vip_expired_at": "2026-04-22T00:00:00Z",
        "is_vip": true,
        "level": "vip",
        "used_quota_today": 5000,
        "used_quota_month": 50000
    }
}
```

#### 获取用量明细

```
GET /api/v1/user/usage
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| start_date | string | 开始日期 |
| end_date | string | 结束日期 |
| model | string | 模型筛选 |
| page | int | 页码 |
| page_size | int | 每页数量 |

### 3.4 充值与 VIP

#### 获取充值套餐

```
GET /api/v1/user/recharge/packages
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": 1,
            "name": "Starter",
            "description": "入门套餐",
            "price": 1.0,
            "quota": 100000,
            "bonus_quota": 0,
            "is_popular": false
        }
    ]
}
```

#### 获取 VIP 套餐

```
GET /api/v1/user/vip/packages
```

#### 创建充值订单

```
POST /api/v1/user/recharge/orders
```

**请求体:**

```json
{
    "package_id": 1,
    "payment_method": "alipay"
}
```

#### 创建 VIP 订单

```
POST /api/v1/user/vip/orders
```

**请求体:**

```json
{
    "package_id": 2,
    "payment_method": "wechat"
}
```

#### 获取订单列表

```
GET /api/v1/user/orders
```

#### 获取订单详情

```
GET /api/v1/user/orders/:id
```

#### 兑换码兑换

```
POST /api/v1/user/redemption
```

**请求体:**

```json
{
    "code": "PROMO2026"
}
```

### 3.5 用户信息

#### 获取用户信息

```
GET /api/v1/user/profile
```

#### 更新用户信息

```
PUT /api/v1/user/profile
```

---

## 4. 管理后台 API

### 4.1 管理员认证

#### 管理员登录

```
POST /api/v1/admin/auth/login
```

**请求体:**

```json
{
    "username": "admin",
    "password": "AdminPass123!"
}
```

### 4.2 仪表盘

#### 获取统计数据

```
GET /api/v1/admin/dashboard/stats
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_users": 1000,
        "active_users_today": 150,
        "total_channels": 20,
        "healthy_channels": 18,
        "total_orders_today": 50,
        "total_revenue_today": 500.00,
        "total_quota_used_today": 5000000,
        "vip_users_count": 100
    }
}
```

#### 获取用量趋势

```
GET /api/v1/admin/dashboard/usage-trend
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| days | int | 天数 (默认7) |

### 4.3 渠道管理

#### 获取渠道列表

```
GET /api/v1/admin/channels
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| type | string | 渠道类型 |
| status | int | 状态 |
| group | string | 分组名 |
| keyword | string | 搜索关键词 |

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 20,
        "list": [
            {
                "id": 1,
                "name": "OpenAI Primary",
                "type": "openai",
                "base_url": "https://api.openai.com",
                "status": 1,
                "is_healthy": true,
                "models": ["gpt-4", "gpt-3.5-turbo"],
                "weight": 100,
                "priority": 0,
                "rpm_limit": 1000,
                "tpm_limit": 100000,
                "failure_count": 0,
                "last_success_at": "2026-03-23T12:00:00Z",
                "response_time_avg": 500,
                "created_at": "2026-03-01T00:00:00Z"
            }
        ]
    }
}
```

#### 创建渠道

```
POST /api/v1/admin/channels
```

**请求体:**

```json
{
    "name": "OpenAI Primary",
    "type": "openai",
    "base_url": "https://api.openai.com",
    "api_key": "sk-xxx...",
    "models": ["gpt-4", "gpt-3.5-turbo"],
    "model_mapping": {},
    "weight": 100,
    "priority": 0,
    "rpm_limit": 1000,
    "tpm_limit": 100000,
    "group_name": "openai",
    "cost_factor": 1.0,
    "price_per_1k_input": 0.01,
    "price_per_1k_output": 0.03
}
```

#### 获取渠道详情

```
GET /api/v1/admin/channels/:id
```

#### 更新渠道

```
PUT /api/v1/admin/channels/:id
```

#### 删除渠道

```
DELETE /api/v1/admin/channels/:id
```

#### 启用/禁用渠道

```
POST /api/v1/admin/channels/:id/enable
POST /api/v1/admin/channels/:id/disable
```

#### 测试渠道 ⭐

```
POST /api/v1/admin/channels/:id/test
```

**请求体:**

```json
{
    "test_type": "chat",
    "model": "gpt-3.5-turbo",
    "messages": [
        {"role": "user", "content": "Hello!"}
    ],
    "temperature": 0.7,
    "max_tokens": 100
}
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "success": true,
        "response_time_ms": 1234,
        "status_code": 200,
        "models": ["gpt-4", "gpt-3.5-turbo"],
        "content": "Hello! How can I help you today?",
        "usage": {
            "prompt_tokens": 20,
            "completion_tokens": 15,
            "total_tokens": 35
        },
        "embedding": null,
        "error": null
    }
}
```

| test_type | 说明 |
|-----------|------|
| models | 获取模型列表 |
| chat | 对话补全测试 |
| embeddings | Embeddings 测试 |

#### 获取测试历史

```
GET /api/v1/admin/channels/:id/test-history
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |

### 4.4 用户管理

#### 获取用户列表

```
GET /api/v1/admin/users
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| level | string | 用户等级 |
| status | string | 状态 |
| keyword | string | 搜索关键词 |

#### 创建用户

```
POST /api/v1/admin/users
```

#### 获取用户详情

```
GET /api/v1/admin/users/:id
```

#### 更新用户

```
PUT /api/v1/admin/users/:id
```

#### 删除用户

```
DELETE /api/v1/admin/users/:id
```

#### 调整用户配额

```
POST /api/v1/admin/users/:id/adjust-quota
```

**请求体:**

```json
{
    "quota_type": "permanent",
    "amount": 100000,
    "reason": "Manual adjustment for promotion"
}
```

#### 手动开通 VIP

```
POST /api/v1/admin/users/:id/activate-vip
```

**请求体:**

```json
{
    "package_id": 2,
    "days": 30
}
```

#### 关闭 VIP

```
POST /api/v1/admin/users/:id/deactivate-vip
```

### 4.5 Token 管理

#### 获取 Token 列表

```
GET /api/v1/admin/tokens
```

#### 重置 Token 配额

```
POST /api/v1/admin/tokens/:id/reset-quota
```

### 4.6 VIP 管理

#### 获取 VIP 套餐列表

```
GET /api/v1/admin/vip-packages
```

#### 创建 VIP 套餐

```
POST /api/v1/admin/vip-packages
```

#### 更新 VIP 套餐

```
PUT /api/v1/admin/vip-packages/:id
```

#### 删除 VIP 套餐

```
DELETE /api/v1/admin/vip-packages/:id
```

### 4.7 充值套餐管理

#### 获取充值套餐列表

```
GET /api/v1/admin/recharge-packages
```

#### 创建充值套餐

```
POST /api/v1/admin/recharge-packages
```

#### 更新充值套餐

```
PUT /api/v1/admin/recharge-packages/:id
```

### 4.8 订单管理

#### 获取订单列表

```
GET /api/v1/admin/orders
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| order_type | string | 订单类型 |
| status | string | 状态 |
| start_date | string | 开始日期 |
| end_date | string | 结束日期 |

#### 获取订单详情

```
GET /api/v1/admin/orders/:id
```

#### 取消订单

```
POST /api/v1/admin/orders/:id/cancel
```

### 4.9 审计日志 ⭐

#### 获取审计日志列表

```
GET /api/v1/admin/audit-logs
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| user_id | int | 用户ID |
| action_group | string | 操作分组 |
| action | string | 操作类型 |
| resource_type | string | 资源类型 |
| resource_id | int | 资源ID |
| success | bool | 是否成功 |
| request_ip | string | 请求IP |
| start_time | string | 开始时间 |
| end_time | string | 结束时间 |

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 1000,
        "list": [
            {
                "id": 1,
                "user_id": 1,
                "username": "testuser",
                "action": "user.login",
                "action_group": "auth",
                "resource_type": "user",
                "resource_id": 1,
                "request_method": "POST",
                "request_path": "/api/v1/user/auth/login",
                "request_ip": "192.168.1.1",
                "status_code": 200,
                "success": true,
                "created_at": "2026-03-23T12:00:00Z"
            }
        ]
    }
}
```

#### 获取审计日志详情

```
GET /api/v1/admin/audit-logs/:id
```

#### 导出审计日志

```
GET /api/v1/admin/audit-logs/export
```

**查询参数:** 同列表接口

**响应:** CSV 文件下载

#### 获取审计统计

```
GET /api/v1/admin/audit-logs/statistics
```

**响应体:**

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total_count": 10000,
        "success_count": 9500,
        "failed_count": 500,
        "top_actions": [
            {"action": "user.login", "count": 5000},
            {"action": "channel.test", "count": 2000}
        ],
        "trend": [
            {"date": "2026-03-01", "count": 500},
            {"date": "2026-03-02", "count": 600}
        ]
    }
}
```

### 4.10 登录日志

#### 获取登录日志列表

```
GET /api/v1/admin/login-logs
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| login_type | string | 登录类型 |
| success | bool | 是否成功 |
| user_id | int | 用户ID |
| start_time | string | 开始时间 |
| end_time | string | 结束时间 |

### 4.11 系统配置

#### 获取配置列表

```
GET /api/v1/admin/configs
```

#### 获取配置详情

```
GET /api/v1/admin/configs/:key
```

#### 更新配置

```
PUT /api/v1/admin/configs/:key
```

### 4.12 兑换码管理

#### 获取兑换码列表

```
GET /api/v1/admin/redemptions
```

#### 创建兑换码

```
POST /api/v1/admin/redemptions
```

**请求体:**

```json
{
    "code": "PROMO2026",
    "code_type": "quota",
    "quota": 100000,
    "quota_type": "permanent",
    "max_uses": 100,
    "valid_until": "2026-12-31T23:59:59Z"
}
```

#### 禁用兑换码

```
POST /api/v1/admin/redemptions/:id/disable
```

---

## 5. WebSocket API (可选)

### 5.1 流式聊天

```
WS /api/v1/ws/chat
```

**连接参数:**

```
ws://localhost:8080/api/v1/ws/chat?token=xxx
```

**发送消息:**

```json
{
    "type": "chat",
    "data": {
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": "Hello!"}
        ],
        "temperature": 0.7
    }
}
```

**接收消息:**

```json
{
    "type": "chunk",
    "data": {
        "content": "Hello",
        "index": 0
    }
}
```

```json
{
    "type": "done",
    "data": {
        "usage": {
            "prompt_tokens": 20,
            "completion_tokens": 15,
            "total_tokens": 35
        }
    }
}
```

---

## 6. 错误码详细说明

| 错误码 | HTTP状态码 | 说明 | 解决方案 |
|--------|------------|------|----------|
| 40001 | 400 | 参数错误 | 检查请求参数 |
| 40002 | 400 | 参数验证失败 | 检查字段格式 |
| 40101 | 401 | 未授权 | 登录后重试 |
| 40102 | 401 | Token 过期 | 刷新Token |
| 40103 | 401 | Token 无效 | 重新登录 |
| 40301 | 403 | 权限不足 | 申请权限 |
| 40302 | 403 | 资源不存在 | 检查资源ID |
| 40401 | 404 | 渠道不存在 | 检查渠道ID |
| 40402 | 404 | Token 不存在 | 检查Token |
| 40403 | 404 | 用户不存在 | 检查用户ID |
| 40901 | 409 | 资源已存在 | 更换名称 |
| 42201 | 422 | 配额不足 | 充值后重试 |
| 42202 | 422 | VIP已过期 | 续费VIP |
| 42901 | 429 | 请求过于频繁 | 降低请求频率 |
| 50001 | 500 | 服务器内部错误 | 联系技术支持 |
| 50002 | 502 | 第三方服务错误 | 检查配置 |

---

**文档版本**: 1.0  
**下一步**: 项目结构设计
