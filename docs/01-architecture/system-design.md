# API Proxy Platform - 完整系统设计文档

**版本**: 5.0  
**日期**: 2026-03-22  
**状态**: 待实现

---

## 确认的技术栈

| 组件 | 选择 |
|------|------|
| 后端 | Go + Gin |
| 前端 | Vue 3 + Element Plus + TypeScript |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 消息队列 | RabbitMQ |
| 部署 | 二进制（可容器化） |
| 支付 | 支付宝 + 微信 |

---

## 1. 用户等级体系

### 1.1 等级定义

```
┌─────────────────────────────────────────────────────────────────┐
│                      用户等级体系                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────┐                                              │
│   │    免费     │  ← 注册即用                                   │
│   │   (free)    │                                              │
│   └──────┬──────┘                                              │
│          │                                                      │
│   ┌──────┴──────┐                                              │
│   │             │                                              │
│   ▼             ▼                                              │
│ ┌──────┐   ┌─────────┐                                        │
│ │ 非VIP │   │   VIP   │                                        │
│ └──────┘   └────┬────┘                                        │
│                   │                                             │
│                   ▼                                             │
│            ┌─────────────┐                                      │
│            │  企业用户   │  ← 私有部署                         │
│            │ (enterprise) │                                      │
│            └─────────────┘                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 等级对比

| 功能 | 免费用户 | 非VIP付费 | VIP用户 | 企业用户 |
|------|---------|-----------|---------|----------|
| 基础配额 | ✅ | ✅ 购买 | ✅ 购买 | ✅ |
| 排队延迟 | ❌ 无 | ✅ 有 | ❌ 无 | ❌ 无 |
| 速率限制 | ✅ 有限额 | ✅ 有限额 | ✅ 高限额 | ✅ 无限制 |
| VIP通道 | ❌ | ❌ | ✅ 优先 | ✅ |
| 过期机制 | 配额永久 | 配额永久 | 30天过期 | 自定义 |
| 私有部署 | ❌ | ❌ | ❌ | ✅ |

---

## 2. VIP 功能详细设计

### 2.1 VIP 权益

| 权益 | 说明 |
|------|------|
| 速度优先 | VIP请求优先处理，无排队延迟 |
| 高配额 | 每分钟100K tokens vs 普通10K |
| 速率提升 | 每分钟2000请求 vs 普通500 |

### 2.2 配额过期机制

| 配额类型 | 有效期 | 说明 |
|----------|--------|------|
| VIP配额 | 30天 | 过期自动失效 |
| 普通配额 | 永久 | 永不过期 |

### 2.3 优先级调度

| 用户类型 | 优先级 | 说明 |
|----------|--------|------|
| VIP用户 | 100 | 直接处理，无排队 |
| 付费用户 | 50 | 有一定优先级 |
| 免费用户 | 10 | 基础优先级 |

---

## 3. 核心业务功能

### 3.1 渠道管理功能

#### 3.1.1 功能列表

```
┌─────────────────────────────────────────────────────────────────┐
│                     渠道管理功能                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  基础功能:                                                      │
│  ├── 渠道列表 (分页、筛选、搜索)                              │
│  ├── 创建渠道 (基本信息、API Key、模型配置)                    │
│  ├── 编辑渠道                                                  │
│  ├── 删除渠道                                                  │
│  ├── 启用/禁用渠道                                            │
│  └── 渠道分组管理                                             │
│                                                                 │
│  高级功能:                                                      │
│  ├── 批量导入渠道 (CSV/Excel)                                │
│  ├── 批量导出渠道                                              │
│  ├── 模型映射配置                                              │
│  └── 渠道权重/优先级调整                                       │
│                                                                 │
│  测试功能 ⭐:                                                  │
│  ├── 单渠道API测试                                            │
│  ├── 模型列表获取测试                                          │
│  ├── 对话补全测试 (支持自定义Prompt)                          │
│  ├── Embeddings测试                                           │
│  ├── 测试历史记录                                              │
│  └── 测试结果导出                                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.1.2 渠道测试API设计

```go
// POST /api/admin/channels/:id/test
type ChannelTestRequest struct {
    TestType string `json:"test_type"`  // "models"|"chat"|"embeddings"
    
    // Chat测试参数
    Model    string `json:"model"`
    Messages []struct{
        Role    string `json:"role"`    // "user"|"assistant"|"system"
        Content string `json:"content"`
    } `json:"messages"`
    
    // Embeddings测试参数
    Input string `json:"input"`
    
    // 可选参数
    Temperature float64 `json:"temperature,omitempty"`
    MaxTokens  int     `json:"max_tokens,omitempty"`
}

type ChannelTestResponse struct {
    Success bool `json:"success"`
    
    // 基础信息
    ResponseTimeMs int64  `json:"response_time_ms"`
    StatusCode    int    `json:"status_code"`
    
    // Models测试结果
    Models []string `json:"models,omitempty"`
    
    // Chat/Completions测试结果
    Content  string `json:"content,omitempty"`
    Usage    *Usage `json:"usage,omitempty"`
    
    // Embeddings测试结果
    Embedding []float64 `json:"embedding,omitempty"`
    
    // 错误信息
    Error     string `json:"error,omitempty"`
    ErrorType string `json:"error_type,omitempty"`
}

// 历史记录
type ChannelTestHistory struct {
    ID            int64     `json:"id"`
    ChannelID    int64     `json:"channel_id"`
    TestType     string    `json:"test_type"`
    Model        string    `json:"model"`
    RequestBody  string    `json:"request_body"`
    ResponseBody string    `json:"response_body"`
    StatusCode   int       `json:"status_code"`
    ResponseTime int       `json:"response_time"`
    Success      bool      `json:"success"`
    CreatedAt    time.Time `json:"created_at"`
}
```

#### 3.1.3 渠道测试API实现

```go
// internal/handler/admin/channel_test.go

func (h *ChannelHandler) TestChannel(c *gin.Context) {
    channelID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    var req ChannelTestRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, "参数错误")
        return
    }
    
    channel := h.channelSvc.Get(channelID)
    if channel == nil {
        response.Fail(c, "渠道不存在")
        return
    }
    
    // 获取解密后的API Key
    apiKey := h.crypto.DecryptAPIKey(channel.APIKeyEncrypted)
    
    var result ChannelTestResponse
    startTime := time.Now()
    
    switch req.TestType {
    case "models":
        result = testModels(channel.BaseURL, apiKey, channel.Type)
    case "chat":
        result = testChat(channel.BaseURL, apiKey, channel.Type, &req)
    case "embeddings":
        result = testEmbeddings(channel.BaseURL, apiKey, channel.Type, &req)
    default:
        response.Fail(c, "不支持的测试类型")
        return
    }
    
    result.ResponseTimeMs = time.Since(startTime).Milliseconds()
    
    // 记录测试历史
    h.saveTestHistory(channelID, &req, &result)
    
    response.Success(c, result)
}

// 具体测试实现
func testModels(baseURL, apiKey, channelType string) ChannelTestResponse {
    var resp OpenAIModelsResponse
    err := httpGet(baseURL+"/models", apiKey, channelType, &resp)
    
    if err != nil {
        return ChannelTestResponse{
            Success: false,
            Error:   err.Error(),
        }
    }
    
    models := make([]string, 0)
    for _, m := range resp.Data {
        models = append(models, m.ID)
    }
    
    return ChannelTestResponse{
        Success: true,
        Models:  models,
    }
}

func testChat(baseURL, apiKey, channelType string, req *ChannelTestRequest) ChannelTestResponse {
    // 转换消息格式
    messages := convertMessages(req.Messages)
    
    body := map[string]interface{}{
        "model":    req.Model,
        "messages":  messages,
    }
    if req.Temperature > 0 {
        body["temperature"] = req.Temperature
    }
    if req.MaxTokens > 0 {
        body["max_tokens"] = req.MaxTokens
    }
    
    var resp OpenAIChatResponse
    err := httpPost(baseURL+"/v1/chat/completions", apiKey, channelType, body, &resp)
    
    if err != nil {
        return ChannelTestResponse{
            Success: false,
            Error:   err.Error(),
        }
    }
    
    return ChannelTestResponse{
        Success: true,
        Content: resp.Choices[0].Message.Content,
        Usage: &Usage{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:     resp.Usage.TotalTokens,
        },
    }
}
```

### 3.2 Token管理功能

```
┌─────────────────────────────────────────────────────────────────┐
│                     Token管理功能                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Token生成:                                                     │
│  ├── 创建Token (名称、配额、模型限制)                          │
│  ├── 批量创建Token                                             │
│  ├── Token Key 显示 (仅显示一次)                               │
│  └── 复制Token                                                 │
│                                                                 │
│  Token配置:                                                     │
│  ├── 名称/备注                                                 │
│  ├── 配额设置 (永久/VIP)                                      │
│  ├── 模型访问限制                                              │
│  ├── IP白名单                                                  │
│  ├── 过期时间                                                  │
│  └── 状态管理 (启用/禁用)                                      │
│                                                                 │
│  Token操作:                                                     │
│  ├── 重置配额                                                  │
│  ├── 查看使用统计                                              │
│  ├── 复制Key                                                   │
│  └── 删除Token                                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 用户管理功能

```
┌─────────────────────────────────────────────────────────────────┐
│                     用户管理功能                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用户操作:                                                      │
│  ├── 用户列表 (分页、筛选)                                     │
│  ├── 创建用户                                                  │
│  ├── 编辑用户                                                  │
│  ├── 删除用户                                                  │
│  ├── 重置密码                                                  │
│  └── 启用/禁用用户                                            │
│                                                                 │
│  用户配额:                                                      │
│  ├── 查看配额                                                  │
│  ├── 调整配额                                                  │
│  ├── 手动充值                                                  │
│  └── 配额明细                                                  │
│                                                                 │
│  VIP管理:                                                       │
│  ├── VIP状态                                                    │
│  ├── VIP到期时间                                                │
│  └── 手动开通/关闭VIP                                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 数据库设计

### 4.1 ER 图

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ tenants  │────<│  users   │────<│ tokens   │     │redempti │
└────┬─────┘     └────┬─────┘     └──────────┘     │ ons    │
     │                │                              └────┬───┘
     │                │                                   │
     ▼                ▼                                   ▼
┌──────────┐     ┌──────────┐                     ┌──────────┐
│channels  │────<│ abilities │                     │quota_txns│
└────┬─────┘     └──────────┘                     └────┬─────┘
     │                                                   │
     ▼                                                   ▼
┌──────────┐     ┌──────────┐                     ┌──────────┐
│usage_logs│     │payments  │                     │ orders   │
│(分区表)  │     └──────────┘                     └──────────┘
└──────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     审计日志相关                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌────────────┐                                                │
│  │audit_logs │ ← 所有操作审计 (用户注册、购买、付款等)        │
│  └─────┬──────┘                                                │
│        │                                                        │
│        ▼                                                        │
│  ┌────────────┐                                                │
│  │login_logs │ ← 登录日志                                      │
│  └────────────┘                                                │
│                                                                 │
│  ┌────────────┐                                                │
│  │api_logs   │ ← API调用日志 (可选，用于成本分析)              │
│  └────────────┘                                                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 审计日志表 (核心)

```sql
-- ============================================================
-- 审计日志表 (核心表 - 用于安全审查和溯源)
-- ============================================================
CREATE TABLE audit_logs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT,                                    -- 租户ID (可为空表示系统操作)
    user_id         BIGINT,                                    -- 操作用户ID
    username        VARCHAR(100),                               -- 操作时用户名 (冗余存储)
    
    -- 操作信息
    action         VARCHAR(100) NOT NULL,                      -- 操作类型
    action_group    VARCHAR(50) NOT NULL,                      -- 操作分组
    resource_type   VARCHAR(50),                               -- 资源类型 (user/channel/token/order等)
    resource_id     BIGINT,                                    -- 资源ID
    
    -- 请求信息
    request_method  VARCHAR(10),                               -- HTTP方法
    request_path    VARCHAR(500),                              -- 请求路径
    request_body    TEXT,                                      -- 请求体 (脱敏后)
    request_ip      VARCHAR(50),                               -- 请求IP
    request_ua      VARCHAR(500),                              -- User-Agent
    
    -- 响应信息
    status_code    INTEGER,                                    -- 响应状态码
    response_body   TEXT,                                      -- 响应体 (脱敏后)
    
    -- 结果
    success        BOOLEAN NOT NULL DEFAULT true,              -- 是否成功
    error_message  TEXT,                                       -- 错误信息
    
    -- 变更详情 (JSON格式)
    old_value      JSONB,                                      -- 变更前的值
    new_value      JSONB,                                      -- 变更后的值
    
    -- 元数据
    user_agent     VARCHAR(500),
    session_id     VARCHAR(100),
    trace_id       VARCHAR(64),                               -- 链路追踪ID
    
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
```

#### 4.2.1 操作类型定义

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

#### 4.2.2 审计日志记录示例

```go
// 审计日志记录示例

// 1. 用户登录
AuditLog{
    Action:       AuditActionUserLogin,
    ActionGroup:  AuditGroupAuth,
    ResourceType: "user",
    ResourceID:   123,
    RequestIP:    "192.168.1.100",
    Success:      true,
}

// 2. 用户注册
AuditLog{
    Action:       AuditActionUserRegister,
    ActionGroup:  AuditGroupAuth,
    ResourceType: "user",
    ResourceID:   124,
    RequestIP:    "192.168.1.101",
    RequestBody:  "{\"username\":\"test\",\"email\":\"test@example.com\"}", // 密码已脱敏
    Success:      true,
    NewValue:     {"username": "test", "email": "test@example.com"},
}

// 3. 订单支付
AuditLog{
    Action:       AuditActionOrderPaid,
    ActionGroup:  AuditGroupOrder,
    ResourceType: "order",
    ResourceID:   456,
    UserID:       123,
    Success:      true,
    NewValue:     {"order_no": "ORD20260322001", "amount": 100.00, "status": "paid"},
}

// 4. 支付回调
AuditLog{
    Action:       AuditActionPaymentCallback,
    ActionGroup:  AuditGroupPayment,
    ResourceType: "payment",
    RequestIP:    "payment.gateway.ip",
    RequestBody:  "{\"order_no\":\"ORD20260322001\",\"trade_status\":\"SUCCESS\"}",
    Success:      true,
    ResponseBody: "{\"code\":\"SUCCESS\"}",
}

// 5. 渠道API测试
AuditLog{
    Action:       AuditActionChannelTest,
    ActionGroup:  AuditGroupChannel,
    ResourceType: "channel",
    ResourceID:   1,
    RequestBody:  "{\"test_type\":\"chat\",\"model\":\"gpt-4\",\"messages\":[...]}",
    ResponseBody: "{\"success\":true,\"content\":\"测试响应\"}",
    ResponseTimeMs: 1234,
    Success:      true,
}

// 6. 配额变更
AuditLog{
    Action:       AuditActionUserQuotaAdd,
    ActionGroup:  AuditGroupQuota,
    ResourceType: "user",
    ResourceID:   123,
    UserID:       123,
    OldValue:     {"remain_quota": 1000},
    NewValue:     {"remain_quota": 101000, "added": 100000, "source": "purchase"},
    Success:      true,
}

// 7. VIP开通
AuditLog{
    Action:       AuditActionVIPActivate,
    ActionGroup:  AuditGroupVIP,
    ResourceType: "user",
    ResourceID:   123,
    UserID:       123,
    OldValue:     {"is_vip": false},
    NewValue:     {"is_vip": true, "expired_at": "2026-04-22"},
    Success:      true,
}
```

### 4.3 登录日志表

```sql
-- ============================================================
-- 登录日志表
-- ============================================================
CREATE TABLE login_logs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT,
    user_id         BIGINT,
    username        VARCHAR(100),
    login_type      VARCHAR(20) NOT NULL,                     -- 'user' | 'admin'
    
    -- 登录信息
    ip             VARCHAR(50),
    ip_location    VARCHAR(200),                              -- IP归属地
    user_agent     VARCHAR(500),
    device_type    VARCHAR(50),                              -- 'web' | 'mobile' | 'desktop'
    
    -- 结果
    success        BOOLEAN NOT NULL,
    fail_reason    VARCHAR(100),
    
    -- Token信息
    token          VARCHAR(200),                              -- 登录Token (加密存储)
    token_expired_at TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_login_user ON login_logs(user_id, created_at DESC);
CREATE INDEX idx_login_ip ON login_logs(ip, created_at DESC);
CREATE INDEX idx_login_type ON login_logs(login_type);
```

### 4.4 API调用日志表 (可选，用于成本分析)

```sql
-- ============================================================
-- API调用日志表 (可选 - 用于成本分析和监控)
-- ============================================================
CREATE TABLE api_logs (
    id                BIGSERIAL,
    tenant_id         BIGINT NOT NULL,
    user_id           BIGINT NOT NULL,
    token_id          BIGINT,
    channel_id        BIGINT,
    
    -- 请求信息
    request_method    VARCHAR(10) NOT NULL,
    request_path      VARCHAR(200) NOT NULL,
    request_body      TEXT,
    
    -- 响应信息
    status_code       INTEGER,
    response_body     TEXT,
    
    -- 用量
    model             VARCHAR(100),
    prompt_tokens     INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens      INTEGER DEFAULT 0,
    
    -- 性能
    response_time_ms  INTEGER,
    
    -- 费用 (可选)
    cost              DECIMAL(10,4) DEFAULT 0,
    
    -- 错误
    error_message     TEXT,
    
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 按月分区
CREATE TABLE api_logs_2026_01 PARTITION OF api_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE api_logs_2026_02 PARTITION OF api_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
-- 继续创建更多分区...

CREATE INDEX idx_api_user_time ON api_logs(user_id, created_at DESC);
CREATE INDEX idx_api_channel ON api_logs(channel_id, created_at DESC);
```

### 4.5 完整表列表

```sql
-- ============================================================
-- 表清单
-- ============================================================

-- 1. 租户表
CREATE TABLE tenants (...);

-- 2. 管理员表
CREATE TABLE admin_users (...);

-- 3. 用户表
CREATE TABLE users (...);

-- 4. 渠道表
CREATE TABLE channels (...);

-- 5. 能力表
CREATE TABLE abilities (...);

-- 6. Token表
CREATE TABLE tokens (...);

-- 7. 用量日志表 (分区)
CREATE TABLE usage_logs (...) PARTITION BY RANGE;

-- 8. 配额流水表
CREATE TABLE quota_transactions (...);

-- 9. 订单表
CREATE TABLE orders (...);

-- 10. 支付表
CREATE TABLE payments (...);

-- 11. 兑换码表
CREATE TABLE redemptions (...);

-- 12. VIP套餐表
CREATE TABLE vip_packages (...);

-- 13. VIP订单表
CREATE TABLE vip_orders (...);

-- 14. 充值套餐表
CREATE TABLE recharge_packages (...);

-- 15. 用户配额表
CREATE TABLE user_quotas (...);

-- ⭐ 审计日志表 (新增)
CREATE TABLE audit_logs (...);

-- ⭐ 登录日志表 (新增)
CREATE TABLE login_logs (...);

-- ⭐ API调用日志表 (可选)
CREATE TABLE api_logs (...) PARTITION BY RANGE;

-- ⭐ 渠道测试历史表 (新增)
CREATE TABLE channel_test_history (
    id              BIGSERIAL PRIMARY KEY,
    channel_id      BIGINT NOT NULL REFERENCES channels(id),
    test_type       VARCHAR(20) NOT NULL,       -- 'models'|'chat'|'embeddings'
    model           VARCHAR(100),
    
    -- 请求
    request_body    TEXT,
    
    -- 响应
    status_code     INTEGER,
    response_body   TEXT,
    response_time_ms INTEGER,
    
    -- 结果
    success         BOOLEAN NOT NULL,
    error_message   TEXT,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_channel ON channel_test_history(channel_id, created_at DESC);
```

---

## 5. 审计日志服务设计

### 5.1 审计日志中间件

```go
// internal/middleware/audit.go

type AuditMiddleware struct {
    repo *repository.AuditRepository
}

func (m *AuditMiddleware) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        
        // 处理请求
        c.Next()
        
        // 记录审计日志
        go m.recordAudit(c, startTime)
    }
}

func (m *AuditMiddleware) recordAudit(c *gin.Context, startTime time.Time) {
    // 判断是否需要记录
    if !m.shouldRecord(c.Request.URL.Path, c.Request.Method) {
        return
    }
    
    // 获取用户信息
    userID, username := m.getUserInfo(c)
    
    // 获取请求体
    requestBody := m.getRequestBody(c)
    
    // 获取响应体
    responseBody := m.getResponseBody(c)
    
    // 构建审计日志
    log := &model.AuditLog{
        TenantID:      m.getTenantID(c),
        UserID:        userID,
        Username:      username,
        Action:        m.determineAction(c),
        ActionGroup:   m.determineActionGroup(c),
        ResourceType:  m.determineResourceType(c),
        ResourceID:    m.getResourceID(c),
        RequestMethod: c.Request.Method,
        RequestPath:   c.Request.URL.Path,
        RequestBody:   m.maskSensitiveData(requestBody),
        RequestIP:     c.ClientIP(),
        UserAgent:     c.Request.UserAgent(),
        StatusCode:    c.Writer.Status(),
        ResponseBody:  m.maskSensitiveData(responseBody),
        Success:       c.Writer.Status() < 400,
        TraceID:       c.GetString("trace_id"),
        CreatedAt:     time.Now(),
    }
    
    m.repo.Create(log)
}

// 需要脱敏的字段
var sensitiveFields = []string{
    "password", "password_hash", "api_key", "token",
    "credit_card", "bank_account", "secret",
}

func (m *AuditMiddleware) maskSensitiveData(data string) string {
    // 脱敏处理
    return data
}
```

### 5.2 手动记录审计日志

```go
// 业务代码中手动记录

// 用户注册
func (s *UserService) Register(ctx *gin.Context, req *RegisterRequest) {
    // 业务逻辑...
    
    user := s.createUser(req)
    
    // 记录审计日志
    audit.Log(&model.AuditLog{
        Action:       model.AuditActionUserRegister,
        ActionGroup:  model.AuditGroupAuth,
        ResourceType: "user",
        ResourceID:   user.ID,
        UserID:       user.ID,
        RequestIP:    ctx.ClientIP(),
        Success:      true,
        NewValue: map[string]interface{}{
            "username": user.Username,
            "email":    user.Email,
        },
    })
}

// 订单支付成功
func (s *PaymentService) HandleCallback(ctx *gin.Context, callback *PaymentCallback) {
    order := s.getOrder(callback.OrderNo)
    
    // 更新订单状态
    order.Status = "paid"
    order.PaidAt = time.Now()
    s.db.Save(order)
    
    // 记录审计日志
    audit.Log(&model.AuditLog{
        Action:       model.AuditActionOrderPaid,
        ActionGroup:  model.AuditGroupOrder,
        ResourceType: "order",
        ResourceID:   order.ID,
        UserID:       order.UserID,
        RequestIP:    ctx.ClientIP(),
        Success:      true,
        OldValue: map[string]interface{}{
            "status": "pending",
        },
        NewValue: map[string]interface{}{
            "status":  "paid",
            "paid_at": order.PaidAt,
        },
    })
}

// 渠道测试
func (s *ChannelService) TestChannel(ctx *gin.Context, channelID int64, req *TestRequest) {
    result := s.performTest(channelID, req)
    
    audit.Log(&model.AuditLog{
        Action:       model.AuditActionChannelTest,
        ActionGroup:  model.AuditGroupChannel,
        ResourceType: "channel",
        ResourceID:   channelID,
        UserID:       getCurrentUserID(ctx),
        RequestIP:    ctx.ClientIP(),
        RequestBody:  json.Marshal(req),
        ResponseBody: json.Marshal(result),
        Success:      result.Success,
    })
}
```

### 5.3 审计日志查询API

```go
// GET /api/admin/audit-logs

type AuditLogQuery struct {
    Page         int    `form:"page"`
    PageSize     int    `form:"page_size"`
    TenantID     int64  `form:"tenant_id"`
    UserID       int64  `form:"user_id"`
    ActionGroup  string `form:"action_group"`
    Action       string `form:"action"`
    ResourceType string `form:"resource_type"`
    ResourceID   int64  `form:"resource_id"`
    StartTime    string `form:"start_time"`
    EndTime      string `form:"end_time"`
    Success      *bool  `form:"success"`
    RequestIP    string `form:"request_ip"`
}

type AuditLogResponse struct {
    ID            int64     `json:"id"`
    TenantID      int64     `json:"tenant_id"`
    UserID        int64     `json:"user_id"`
    Username      string    `json:"username"`
    Action        string    `json:"action"`
    ActionGroup   string    `json:"action_group"`
    ResourceType  string    `json:"resource_type"`
    ResourceID    int64     `json:"resource_id"`
    RequestMethod string    `json:"request_method"`
    RequestPath   string    `json:"request_path"`
    RequestIP     string    `json:"request_ip"`
    StatusCode    int       `json:"status_code"`
    Success       bool      `json:"success"`
    ErrorMessage  string    `json:"error_message,omitempty"`
    OldValue      JSONMap   `json:"old_value,omitempty"`
    NewValue      JSONMap   `json:"new_value,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
}
```

---

## 6. API 设计

### 6.1 OpenAI 兼容 API

```
POST /v1/chat/completions    # 聊天补全
POST /v1/completions          # 文本补全
POST /v1/embeddings           # Embeddings
GET  /v1/models              # 模型列表
```

### 6.2 管理后台 API

```
# 审计日志 ⭐
GET    /api/admin/audit-logs              # 审计日志列表
GET    /api/admin/audit-logs/export      # 导出审计日志
GET    /api/admin/audit-logs/statistics  # 审计统计

# 登录日志 ⭐
GET    /api/admin/login-logs             # 登录日志列表

# 渠道管理 (含测试)
GET    /api/admin/channels
POST   /api/admin/channels
GET    /api/admin/channels/:id
PUT    /api/admin/channels/:id
DELETE /api/admin/channels/:id
POST   /api/admin/channels/:id/enable
POST   /api/admin/channels/:id/disable
POST   /api/admin/channels/:id/test      # 测试渠道 ⭐
GET    /api/admin/channels/:id/test-history  # 测试历史 ⭐
POST   /api/admin/channels/batch         # 批量导入
GET    /api/admin/channels/export        # 导出
```

### 6.3 用户端 API

```
# 用户认证
POST   /api/user/auth/register
POST   /api/user/auth/login
POST   /api/user/auth/logout

# Token管理
GET    /api/user/tokens
POST   /api/user/tokens
DELETE /api/user/tokens/:id

# 充值
GET    /api/user/recharge/packages
POST   /api/user/recharge/orders

# VIP
GET    /api/user/vip/packages
POST   /api/user/vip/orders
GET    /api/user/vip/status
```

---

## 7. 前端页面设计

### 7.1 管理后台页面

```
admin/
├── Login.vue                  # 登录
├── Dashboard.vue             # 仪表盘
│
├── audit/                   # 审计日志 ⭐
│   ├── List.vue            # 审计日志列表
│   ├── Detail.vue          # 日志详情
│   └── Statistics.vue      # 审计统计
│
├── login-logs/             # 登录日志 ⭐
│   └── List.vue            # 登录日志列表
│
├── channel/                # 渠道管理
│   ├── List.vue           # 渠道列表
│   ├── Form.vue           # 创建/编辑
│   ├── Test.vue           # 测试对话框 ⭐
│   └── TestHistory.vue     # 测试历史 ⭐
│
├── token/                  # Token管理
├── user/                   # 用户管理
├── vip/                    # VIP管理
├── order/                  # 订单管理
└── settings/              # 系统设置
```

### 7.2 渠道测试页面设计

```vue
<!-- views/admin/channel/TestDialog.vue -->
<template>
  <el-dialog v-model="visible" title="测试渠道" width="800px">
    <!-- 测试类型选择 -->
    <el-radio-group v-model="form.test_type" class="test-type-group">
      <el-radio-button value="models">获取模型列表</el-radio-button>
      <el-radio-button value="chat">对话测试</el-radio-button>
      <el-radio-button value="embeddings">Embeddings测试</el-radio-button>
    </el-radio-group>

    <!-- Chat测试 -->
    <div v-if="form.test_type === 'chat'" class="test-form">
      <el-form :model="form" label-width="100px">
        <el-form-item label="模型">
          <el-select v-model="form.model" placeholder="选择模型">
            <el-option 
              v-for="m in channelModels" 
              :key="m" 
              :label="m" 
              :value="m" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="System Prompt">
          <el-input 
            v-model="form.system_prompt" 
            type="textarea" 
            :rows="2"
            placeholder="可选：设置系统提示词"
          />
        </el-form-item>
        
        <el-form-item label="User Prompt">
          <el-input 
            v-model="form.user_prompt" 
            type="textarea" 
            :rows="3"
            placeholder="输入测试问题"
          />
        </el-form-item>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="Temperature">
              <el-input-number 
                v-model="form.temperature" 
                :min="0" 
                :max="2" 
                :step="0.1"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Max Tokens">
              <el-input-number 
                v-model="form.max_tokens" 
                :min="1" 
                :max="4096"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <!-- Embeddings测试 -->
    <div v-if="form.test_type === 'embeddings'" class="test-form">
      <el-form :model="form" label-width="100px">
        <el-form-item label="模型">
          <el-select v-model="form.model">
            <el-option label="text-embedding-ada-002" value="text-embedding-ada-002" />
            <el-option label="text-embedding-3-small" value="text-embedding-3-small" />
            <el-option label="text-embedding-3-large" value="text-embedding-3-large" />
          </el-select>
        </el-form-item>
        <el-form-item label="输入文本">
          <el-input 
            v-model="form.input" 
            type="textarea" 
            :rows="4"
            placeholder="输入要计算Embedding的文本"
          />
        </el-form-item>
      </el-form>
    </div>

    <!-- Models测试无需额外参数 -->

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleTest">
        开始测试
      </el-button>
    </template>
  </el-dialog>

  <!-- 测试结果 -->
  <el-drawer v-model="resultVisible" title="测试结果" size="50%">
    <div v-if="result">
      <el-alert 
        :type="result.success ? 'success' : 'error'" 
        :title="result.success ? '测试成功' : '测试失败'"
      >
        <p>响应时间: {{ result.response_time_ms }}ms</p>
        <p v-if="result.error">错误: {{ result.error }}</p>
      </el-alert>

      <!-- Models结果 -->
      <div v-if="result.models" class="result-section">
        <h4>可用模型 ({{ result.models.length }})</h4>
        <el-tag 
          v-for="m in result.models" 
          :key="m" 
          class="model-tag"
        >
          {{ m }}
        </el-tag>
      </div>

      <!-- Chat结果 -->
      <div v-if="result.content" class="result-section">
        <h4>AI 响应</h4>
        <el-input
          v-model="result.content"
          type="textarea"
          :rows="6"
          readonly
        />
        <div v-if="result.usage" class="usage-info">
          <span>Prompt Tokens: {{ result.usage.prompt_tokens }}</span>
          <span>Completion Tokens: {{ result.usage.completion_tokens }}</span>
          <span>Total Tokens: {{ result.usage.total_tokens }}</span>
        </div>
      </div>

      <!-- Embeddings结果 -->
      <div v-if="result.embedding" class="result-section">
        <h4>Embedding 向量</h4>
        <p>维度: {{ result.embedding.length }}</p>
        <el-input
          v-model="embeddingStr"
          type="textarea"
          :rows="4"
          readonly
        />
      </div>
    </div>
  </el-drawer>
</template>
```

### 7.3 审计日志页面

```vue
<!-- views/admin/audit/List.vue -->
<template>
  <div class="audit-logs">
    <el-card>
      <el-form inline :model="query">
        <el-form-item label="操作分组">
          <el-select v-model="query.action_group" clearable>
            <el-option label="认证" value="auth" />
            <el-option label="用户" value="user" />
            <el-option label="渠道" value="channel" />
            <el-option label="Token" value="token" />
            <el-option label="订单" value="order" />
            <el-option label="支付" value="payment" />
            <el-option label="VIP" value="vip" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="操作类型">
          <el-select v-model="query.action" clearable>
            <!-- 根据action_group动态加载 -->
          </el-select>
        </el-form-item>
        
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="query.time_range"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
          />
        </el-form-item>
        
        <el-form-item label="结果">
          <el-select v-model="query.success">
            <el-option label="全部" :value="null" />
            <el-option label="成功" :value="true" />
            <el-option label="失败" :value="false" />
          </el-select>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleExport">导出</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-table :data="list" stripe v-loading="loading">
      <el-table-column prop="created_at" label="时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      
      <el-table-column prop="action_group" label="分组" width="80">
        <template #default="{ row }">
          <el-tag :type="getGroupTagType(row.action_group)">
            {{ getGroupLabel(row.action_group) }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="action" label="操作" width="150">
        <template #default="{ row }">
          <el-tooltip :content="row.action" placement="top">
            {{ getActionLabel(row.action) }}
          </el-tooltip>
        </template>
      </el-table-column>
      
      <el-table-column prop="username" label="用户" width="120" />
      
      <el-table-column prop="resource_type" label="资源类型" width="100" />
      
      <el-table-column prop="resource_id" label="资源ID" width="80" />
      
      <el-table-column prop="request_ip" label="IP" width="130" />
      
      <el-table-column prop="status_code" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.success ? 'success' : 'danger'">
            {{ row.status_code || '-' }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleDetail(row)">
            详情
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[20, 50, 100, 200]"
      layout="total, sizes, prev, pager, next"
      @change="fetchList"
    />
  </div>
</template>
```

---

## 8. 业务功能清单

### 8.1 用户端功能

| 功能 | 说明 | 优先级 |
|------|------|--------|
| 用户注册 | 邮箱注册、验证 | P0 |
| 用户登录 | 邮箱+密码登录 | P0 |
| Token管理 | 创建/删除/查看Token | P0 |
| 充值 | 支付宝/微信充值 | P0 |
| VIP购买 | 购买VIP套餐 | P0 |
| 兑换码 | 输入兑换码充值 | P1 |
| 配额查询 | 查看永久/VIP配额 | P0 |
| 用量明细 | 查看API调用记录 | P1 |

### 8.2 管理后台功能

| 功能 | 说明 | 优先级 |
|------|------|--------|
| 管理员登录 | 管理员认证 | P0 |
| 仪表盘 | 统计概览 | P0 |
| 渠道管理 | CRUD、测试 ⭐ | P0 |
| 渠道测试 | Models/Chat/Embeddings测试 ⭐ | P0 |
| Token管理 | CRUD、重置配额 | P0 |
| 用户管理 | CRUD、调整配额 | P0 |
| VIP管理 | 开通/关闭VIP | P0 |
| 订单管理 | 查看/处理订单 | P0 |
| 支付配置 | 支付宝/微信配置 | P0 |
| 兑换码 | 生成/使用兑换码 | P1 |
| 充值套餐 | 配置充值套餐 | P1 |
| VIP套餐 | 配置VIP套餐 | P1 |
| 审计日志 ⭐ | 完整操作审计 | P0 |
| 登录日志 ⭐ | 登录记录查询 | P0 |
| 系统设置 | 基础配置 | P1 |

### 8.3 渠道测试功能详情

```
┌─────────────────────────────────────────────────────────────────┐
│                    渠道测试功能 (⭐核心)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Models 测试                                                 │
│     ├─ 用途: 获取渠道支持的模型列表                             │
│     ├─ 返回: 模型ID列表                                        │
│     └─ 用于: 验证渠道连接、查看可用模型                        │
│                                                                 │
│  2. Chat 测试                                                   │
│     ├─ 用途: 测试对话补全功能                                  │
│     ├─ 参数:                                                    │
│     │   ├─ 模型选择                                            │
│     │   ├─ System Prompt (可选)                               │
│     │   ├─ User Prompt (必填)                                 │
│     │   ├─ Temperature (可选)                                  │
│     │   └─ Max Tokens (可选)                                   │
│     └─ 返回: AI响应、Token用量、响应时间                       │
│                                                                 │
│  3. Embeddings 测试                                             │
│     ├─ 用途: 测试文本向量化功能                                │
│     ├─ 参数:                                                    │
│     │   ├─ 模型选择                                            │
│     │   └─ 输入文本                                            │
│     └─ 返回: Embedding向量、维度                                │
│                                                                 │
│  4. 测试历史                                                    │
│     ├─ 记录每次测试的请求/响应                                 │
│     ├─ 可追溯测试过程                                           │
│     └─ 支持导出测试结果                                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 8.4 审计日志功能详情

```
┌─────────────────────────────────────────────────────────────────┐
│                    审计日志功能 (⭐核心)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. 记录范围                                                    │
│     ├─ 用户操作: 注册、登录、修改密码                           │
│     ├─ 业务操作: 创建/修改/删除渠道、Token、用户              │
│     ├─ 财务操作: 充值、购买、支付、退款                        │
│     ├─ VIP操作: 开通、过期、取消                               │
│     └─ 系统操作: 配置变更、测试等                              │
│                                                                 │
│  2. 记录内容                                                    │
│     ├─ 时间、用户、IP、User-Agent                             │
│     ├─ 操作类型、资源类型、资源ID                               │
│     ├─ 请求/响应内容 (脱敏)                                    │
│     ├─ 操作结果、成功/失败                                     │
│     └─ 变更前后的值 (用于追溯)                                 │
│                                                                 │
│  3. 查询功能                                                    │
│     ├─ 按时间范围查询                                          │
│     ├─ 按用户/操作类型查询                                    │
│     ├─ 按资源类型/ID查询                                       │
│     ├─ 按结果(成功/失败)查询                                  │
│     └─ 按IP查询 (安全审查)                                     │
│                                                                 │
│  4. 统计报表                                                    │
│     ├─ 操作趋势图                                             │
│     ├─ 高风险操作统计                                          │
│     └─ 用户操作热度                                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 9. 确认清单

| 项目 | 状态 | 备注 |
|------|------|------|
| 后端语言 | ✅ Go | |
| 前端框架 | ✅ Vue 3 + Element Plus + TypeScript | |
| 数据库 | ✅ PostgreSQL | |
| 缓存 | ✅ Redis | |
| 消息队列 | ✅ RabbitMQ | |
| 部署方式 | ✅ 二进制可容器化 | |
| 支付方式 | ✅ 支付宝 + 微信 | |
| VIP功能 | ✅ 速度优先 + 配额过期 | |
| 渠道测试 | ✅ Models/Chat/Embeddings | |
| 审计日志 | ✅ 完整操作审计 | |
| 登录日志 | ✅ 登录记录 | |

---

**文档版本**: 5.2  
**更新日期**: 2026-03-23

---

## 10. 详细设计文档链接

| 文档 | 路径 | 说明 |
|------|------|------|
| 数据库设计 | `.sisyphus/plans/database-design-v2.md` | PostgreSQL完整DDL、分区策略、索引优化 |
| API设计 | `.sisyphus/plans/api-design.md` | REST API规范、请求/响应格式、错误码 |
| 接口细分 | `.sisyphus/plans/interface-design-south-north.md` | **北向/南向/管理后台接口分离** |
| 项目结构 | `.sisyphus/plans/project-structure.md` | Go+Vue目录结构、依赖管理 |
| 安全与部署 | `.sisyphus/plans/security-deployment.md` | JWT认证、加密、速率限制、容器化 |
| 业务细节 | `.sisyphus/plans/business-detail.md` | 注册赠送、商品管理、页面隔离 |

---

## 11. 核心特性总结

| 特性 | 描述 | 优先级 |
|------|------|--------|
| 多租户隔离 | tenant_id 全表隔离 | P0 |
| VIP系统 | 30天过期、高配额、优先处理 | P0 |
| 渠道测试 | Models/Chat/Embeddings三种测试 | P0 |
| 审计日志 | 完整操作追溯 | P0 |
| 支付集成 | 支付宝+微信支付 | P0 |
| 配额管理 | 永久配额+VIP配额分离 | P0 |

---

## 12. 下一步

1. 项目初始化 (Go项目骨架)
2. 数据库初始化 (执行DDL)
3. 核心模块实现
4. 前端项目初始化
