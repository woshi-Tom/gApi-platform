# API Key 全生命周期与 API 调用链路完善计划

> 版本: v1.0
> 日期: 2026-05-12
> 状态: 待审批
> 目标版本: v1.3.0
> 优先级: 高

---

## 一、目标

用户通过平台调用 AI 服务的完整链路必须做到：**安全可控、配额精确、错误规范、可观测**。本文档覆盖从用户创建 API Key 到完成一次 API 调用的全链路，对齐 OpenAI/大厂 API 平台标准。

---

## 二、现状分析

### 2.1 已实现链路

```
用户请求 POST /v1/chat/completions
  │
  ├─ TokenAuth 中间件 ──→ 验证 sk-ap-xxx → 查库 → 设置 token_id/user_id
  │
  ├─ TokenRateLimit 中间件 ──→ 固定 10 RPS/Token（硬编码，未读取 token.RPMLimit）
  │
  ├─ APIAccessLog 中间件 ──→ 记录调用日志（model 字段提取有 bug）
  │
  ├─ apiHandler.ChatCompletions()
  │   ├─ 解析请求体
  │   ├─ PreConsumeQuota ──→ 估算 token 消耗 → 检查配额
  │   ├─ chatWithFailover ──→ SelectChannel → 适配器转发（最多 3 次重试）
  │   └─ 返回响应
  │
  └─ ❌ 缺失: 流式响应后无 PostConsumeQuota 调用
```

### 2.2 问题清单

| # | 模块 | 问题 | 严重度 | 影响 | 状态 |
|---|------|------|--------|------|------|
| P0-1 | TokenAuth | 未校验 Token 的 `allowed_ips` 白名单 | 🔴 高 | 任意 IP 可用他人 Token | ✅ 已修复 |
| P0-2 | TokenAuth | 未校验 Token 的 `allowed_models` 模型权限 | 🔴 高 | Token 限模型不生效 | ✅ 已修复 |
| P0-3 | Billing | 流式响应无 PostConsumeQuota，配额不扣减 | 🔴 高 | 流式调用免费 | ✅ 已修复 |
| P0-4 | Billing | 无流式 token 统计，usage 无法记录 | 🔴 高 | 流式调用无日志 | ✅ 已修复 |
| P1-1 | RateLimit | TokenRateLimit 硬编码 10 RPS，忽略 token.RPMLimit | 🟡 中 | 用户自定义限速不生效 |
| P1-2 | RateLimit | 无 TPM（每分钟 Token 数）限制 | 🟡 中 | 无法控制 Token 消耗速率 |
| P1-3 | AccessLog | `model` 字段从 `PostForm` 提取，实际请求为 JSON body | 🟡 中 | 日志中 model 恒为 "unknown" |
| P1-4 | API | 无 `/v1/completions` 端点 | 🟡 中 | 不完整 OpenAI 兼容 |
| P1-5 | Billing | PostConsumeQuota 无事务保护 | 🟡 中 | 并发扣费可能数据不一致 |
| P1-6 | Billing | `consumeRechargeQuota` 返回 false（空实现） | 🟡 中 | 充值配额永远不消耗 |
| P2-1 | TokenAuth | 未校验 `denied_models` 黑名单 | 🟢 低 | 黑名单功能不生效 |
| P2-2 | API | 非流式响应后无 PostConsumeQuota 调用 | 🟢 低 | 非流式也未扣费 |
| P2-3 | API | 无 `/v1/completions`（text completions） | 🟢 低 | OpenAI 兼容不完整 |
| P2-4 | Error | 错误响应格式不统一（混合 APIResponse 和 APIErrorResponse） | 🟢 低 | 客户端适配困难 |

---

## 三、API 调用全链路设计

### 3.1 标准调用流程（对标 OpenAI）

```
┌─────────────────────────────────────────────────────────────────────┐
│                    API 调用全链路（12 步）                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ①  请求到达                                                       │
│      POST /v1/chat/completions                                     │
│      Authorization: Bearer sk-ap-xxxxx                             │
│      {"model":"gpt-4","messages":[...],"stream":true}              │
│       │                                                             │
│       ▼                                                             │
│  ②  Token 鉴权 (TokenAuth)                                         │
│      ├─ 提取 Bearer token                                          │
│      ├─ 查库验证 token 存在且 status=active                         │
│      ├─ 检查过期时间                                                │
│      └─ 设置 context: token_id, user_id                            │
│       │                                                             │
│       ▼                                                             │
│  ③  IP 白名单校验                                                  │
│      ├─ token.allowed_ips 为空 → 跳过                              │
│      └─ 不为空 → 检查 client IP 是否在白名单内                      │
│       │                                                             │
│       ▼                                                             │
│  ④  模型权限校验                                                   │
│      ├─ 解析请求中的 model 字段                                     │
│      ├─ token.allowed_models 不为空 → model 必须在列表中            │
│      ├─ token.denied_models 不为空 → model 不能在列表中             │
│      └─ 不匹配 → 返回 403 model_not_allowed                        │
│       │                                                             │
│       ▼                                                             │
│  ⑤  限速检查 (TokenRateLimit)                                      │
│      ├─ RPM: 按 token.RPMLimit 或默认值                            │
│      ├─ TPM: 按 token.TPMLimit 或默认值（滑动窗口）                 │
│      └─ 超限 → 返回 429 rate_limit_exceeded                        │
│       │                                                             │
│       ▼                                                             │
│  ⑥  配额预检 (PreConsumeQuota)                                     │
│      ├─ 估算本次消耗 token 数                                       │
│      ├─ 检查用户总可用配额（free + recharge + vip）                 │
│      ├─ 检查 Token 级配额（token.remain_quota）                    │
│      └─ 不足 → 返回 402 quota_insufficient                         │
│       │                                                             │
│       ▼                                                             │
│  ⑦  渠道选择 (SelectChannel)                                       │
│      ├─ 按 model 查找支持的渠道                                     │
│      ├─ 权重随机选择                                                │
│      ├─ 跳过不健康渠道                                              │
│      └─ 失败 → 返回 503 no_available_channel                       │
│       │                                                             │
│       ▼                                                             │
│  ⑧  上游转发                                                       │
│      ├─ 解密渠道 API Key                                           │
│      ├─ 模型映射（如有）                                            │
│      ├─ 调用适配器 (OpenAI/Claude/Gemini/...)                      │
│      └─ 失败 → 自动重试下一个渠道（最多 3 次）                      │
│       │                                                             │
│       ▼                                                             │
│  ⑨  流式 / 非流式分流                                              │
│      ├─ 非流式: 直接返回完整响应                                    │
│      └─ 流式: SSE 逐 chunk 推送                                    │
│       │                                                             │
│       ▼                                                             │
│  ⑩  实际消耗统计                                                   │
│      ├─ 非流式: 从响应 JSON 中提取 usage.prompt_tokens/completion_tokens │
│      └─ 流式: 在最后一个 chunk 中提取 usage（或客户端上报）         │
│       │                                                             │
│       ▼                                                             │
│  ⑪  配额扣减 (PostConsumeQuota)                                    │
│      ├─ 按优先级扣减: free → recharge(FIFO) → vip                  │
│      ├─ 更新 token.used_quota / token.remain_quota                 │
│      ├─ 写入 quota_transactions 记录                               │
│      └─ 事务保护，失败回滚                                          │
│       │                                                             │
│       ▼                                                             │
│  ⑫  使用日志记录 (LogUsage)                                        │
│      ├─ 写入 usage_logs                                            │
│      └─ 写入 api_access_logs                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 错误响应规范（OpenAI 兼容）

所有 `/v1/*` 接口统一使用以下格式：

```json
{
    "error": {
        "message": "Incorrect API key provided: sk-ap-xxxx. You can find your API key at https://platform.example.com.",
        "type": "invalid_request_error",
        "param": null,
        "code": "invalid_api_key"
    }
}
```

| HTTP Status | error.type | error.code | 场景 |
|-------------|-----------|------------|------|
| 401 | `invalid_request_error` | `invalid_api_key` | Token 无效/不存在 |
| 401 | `invalid_request_error` | `expired_api_key` | Token 已过期 |
| 403 | `invalid_request_error` | `model_not_allowed` | 模型不在允许列表 |
| 403 | `invalid_request_error` | `ip_not_allowed` | IP 不在白名单 |
| 400 | `invalid_request_error` | `invalid_request` | 请求体格式错误 |
| 400 | `invalid_request_error` | `missing_model` | 缺少 model 字段 |
| 402 | `invalid_request_error` | `quota_insufficient` | 配额不足 |
| 429 | `rate_limit_error` | `rate_limit_rpm` | RPM 超限 |
| 429 | `rate_limit_error` | `rate_limit_tpm` | TPM 超限 |
| 503 | `server_error` | `no_available_channel` | 无可用渠道 |
| 502 | `server_error` | `upstream_error` | 上游服务错误 |
| 500 | `server_error` | `internal_error` | 内部错误 |

---

## 四、实现任务清单

### Phase 1: 安全加固（P0）

#### T-01: TokenAuth 增加 IP 白名单校验

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/jwt.go` → `TokenAuth()` |
| 修改点 | 在 token 验证通过后、设置 context 前，增加 IP 白名单检查 |
| 逻辑 | `token.GetAllowedIPs()` 非空时，`c.ClientIP()` 必须在列表中 |
| 错误响应 | `403 {"error":{"type":"invalid_request_error","code":"ip_not_allowed"}}` |
| 测试 | Token 设置 allowed_ips=["10.0.0.1"]，用其他 IP 请求应被拒绝 |

```go
// 在 TokenAuth 中，token 校验通过后添加:
allowedIPs := token.GetAllowedIPs()
if len(allowedIPs) > 0 {
    clientIP := c.ClientIP()
    if !isIPAllowed(clientIP, allowedIPs) {
        c.JSON(http.StatusForbidden, model.APIErrorResponse{
            Error: &model.APIError{
                Type:    "invalid_request_error",
                Code:    "ip_not_allowed",
                Message: fmt.Sprintf("IP %s is not in the token's allowed list", clientIP),
            },
        })
        c.Abort()
        return
    }
}
```

#### T-02: TokenAuth 增加模型权限校验

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/jwt.go` → `TokenAuth()` |
| 修改点 | 在 IP 白名单校验之后，解析请求 model 字段并校验权限 |
| 逻辑 | 1. 从请求体中提取 model（需读取 body 并重置）<br>2. `allowed_models` 非空 → model 必须在列表中<br>3. `denied_models` 非空 → model 不能在列表中 |
| 错误响应 | `403 {"error":{"type":"invalid_request_error","code":"model_not_allowed"}}` |
| 注意 | 需要读取 body，对 `/v1/models` 等无 model 字段的请求跳过 |

```go
// 辅助函数：从请求体中安全提取 model 字段
func extractModelFromRequest(c *gin.Context) string {
    // 仅对 chat/completions 和 embeddings 生效
    path := c.Request.URL.Path
    if !strings.Contains(path, "/completions") && !strings.Contains(path, "/embeddings") {
        return ""
    }

    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        return ""
    }
    // 重置 body 供后续 handler 读取
    c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

    var partial struct {
        Model string `json:"model"`
    }
    if err := json.Unmarshal(body, &partial); err != nil {
        return ""
    }
    return partial.Model
}
```

#### T-03: 流式响应配额扣减与 Token 统计

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/handler/api_handler.go` → `handleStreamWithFailover()` |
| 修改点 | 流式结束后调用 PostConsumeQuota 和 LogUsage |
| 关键问题 | 流式模式下，上游返回的是 SSE stream，不会在单个 response 中返回 usage |
| 方案 | 在最后一个 chunk（`[DONE]` 之前）或 `[DONE]` 之后，从累计统计中提取 token 用量 |

**方案 A：客户端报告（推荐，对标 OpenAI 官方）**

OpenAI 官方流式响应中，最后一个 chunk 包含 `usage` 字段（需传 `stream_options.include_usage = true`）。在平台侧：
- 如果上游返回了 usage chunk → 直接使用
- 如果上游未返回 → 用 `estimateChatTokens` 估算（兜底）

**方案 B：平台侧计数**

在 adapter 层统计返回的 content token 数，但精度不高。

```go
// 在 c.Stream 回调中，DONE 时触发扣费
if chunk.Done {
    // 从 chunk 中提取 usage（如果上游返回了）
    if chunk.Usage != nil {
        go h.postConsumeAfterStream(c, chatReq.Model, selectedChannel.ID,
            chunk.Usage.PromptTokens, chunk.Usage.CompletionTokens)
    } else {
        // 兜底：用估算值
        estimated := h.estimateChatTokens(chatReq.Messages, chatReq.MaxTokens)
        go h.postConsumeAfterStream(c, chatReq.Model, selectedChannel.ID,
            estimated/2, estimated/2) // 粗略分配 prompt/completion
    }
    return false
}
```

#### T-04: 非流式响应配额扣减

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/handler/api_handler.go` → `ChatCompletions()` |
| 修改点 | `chatWithFailover` 成功后，从响应中提取 usage 并调用 PostConsumeQuota |
| 逻辑 | 1. adapter.Chat() 返回的响应包含 usage 字段<br>2. 调用 billingService.PostConsumeQuota<br>3. 调用 billingService.LogUsage |

```go
// chatWithFailover 成功后:
if h.billingService != nil {
    userID := getUserID(c)
    tokenID := getTokenID(c)
    if userID > 0 && tokenID > 0 {
        if chatResp, ok := resp.(*adapter.ChatResponse); ok && chatResp.Usage != nil {
            h.billingService.PostConsumeQuota(userID, tokenID, req.Model,
                chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens)
        }
    }
}
```

---

### Phase 2: 限速与计费完善（P1）

#### T-05: TokenRateLimit 支持自定义 RPM/TPM

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/ratelimit.go` → `TokenRateLimit()` |
| 修改点 | 当前硬编码 `rate.NewLimiter(10, 10)`，改为从 token 记录读取 |
| 方案 | TokenRateLimit 接收 tokenService 参数，从 context 中的 token_id 查库获取 RPMLimit/TPMLimit |
| 缓存 | 使用 sync.Map 缓存 token 限速配置，避免每次请求查库 |

```go
func TokenRateLimit(tokenService *service.TokenService) gin.HandlerFunc {
    // 缓存: tokenID → {rpm, tpm}
    configs := sync.Map{}

    return func(c *gin.Context) {
        tokenID, _ := c.Get("token_id")
        tid := tokenID.(uint)

        // 从缓存或数据库获取限速配置
        cfg, _ := configs.Load(tid)
        if cfg == nil {
            token, err := tokenService.GetByID(tid)
            if err == nil && token != nil {
                rpm := 10 // 默认
                if token.RPMLimit != nil && *token.RPMLimit > 0 {
                    rpm = *token.RPMLimit
                }
                tpm := 0
                if token.TPMLimit != nil {
                    tpm = *token.TPMLimit
                }
                cfg = &tokenRateConfig{rpm: rpm, tpm: tpm}
                configs.Store(tid, cfg)
            }
        }
        // ... 限速检查逻辑
    }
}
```

#### T-06: PostConsumeQuota 事务保护

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/service/billing_service.go` → `PostConsumeQuota()` |
| 修改点 | 当前 user/token 更新无事务，需包裹在 DB 事务中 |
| 逻辑 | 1. `s.userRepo.BeginTx()`<br>2. 扣减 user 配额 + 写 quota_transaction<br>3. 更新 token.used_quota/remain_quota<br>4. 失败则 `tx.Rollback()` |

#### T-07: 实现 consumeRechargeQuota

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/service/billing_service.go` → `consumeRechargeQuota()` |
| 现状 | 返回 `false`（空实现），充值配额永远不消耗 |
| 逻辑 | 1. 查询用户有效的 `user_recharge_records`（按过期时间 FIFO 排序）<br>2. 逐条扣减 `remaining`<br>3. 用完的记录标记 status=used<br>4. 写入 quota_transaction |

#### T-08: APIAccessLog 修复 model 字段提取

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/api_access_log.go` → `APIAccessLog()` |
| 问题 | `c.PostForm("model")` 返回空值（请求体为 JSON，非 form） |
| 方案 | 从 gin.Context 中读取已解析的 model（handler 设置到 context 中），或从 URL path 判断 |

```go
// 方案: 在 handler 中解析完请求后，将 model 设置到 context
// api_handler.go ChatCompletions() 开头:
c.Set("request_model", req.Model)

// api_access_log.go 中:
modelName := "unknown"
if m, exists := c.Get("request_model"); exists {
    modelName = m.(string)
}
```

---

### Phase 3: 功能补全（P1-P2）

#### T-09: 新增 /v1/completions 端点

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/handler/api_handler.go`（新增方法）|
| 路由 | `backend/internal/router/router.go` |
| 路径 | `v1.POST("/completions", middleware.TokenAuth(...), ..., apiHandler.Completions)` |
| 逻辑 | 类似 ChatCompletions，但请求体为 `{model, prompt, max_tokens, stream}` |
| 适配器 | 复用 adapter 层，将 prompt 转换为 messages 格式 |

#### T-10: APIError 结构统一

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/model/response.go` |
| 修改 | APIError 增加 `Type` 字段，统一所有 `/v1/*` 接口的错误格式 |
| 格式 | `{"error":{"type":"...","code":"...","message":"...","param":null}}` |

```go
type APIError struct {
    Type    string      `json:"type"`
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Param   interface{} `json:"param"`
}
```

#### T-11: Token 列表返回配额使用情况

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/handler/token_handler.go` → `List()` |
| 修改 | 每个 token 返回 `used_quota`、`remain_quota`、`usage_percent` |

---

### Phase 4: 压力测试与可观测（P2）

#### T-12: API 压力测试脚本

| 项 | 内容 |
|---|------|
| 工具 | `k6`（推荐）或扩展 `test_api.py` |
| 场景 | 见第六节测试计划 |
| 输出 | TPS、P50/P95/P99 延迟、错误率 |

#### T-13: API 调用链路监控增强

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/api_access_log.go` |
| 新增字段 | `pre_check_duration_ms`（配额预检耗时）、`channel_select_duration_ms`（渠道选择耗时）、`upstream_duration_ms`（上游调用耗时）|
| 目的 | 定位链路中的性能瓶颈 |

---

## 五、配额扣减详细设计

### 5.1 扣减优先级

```
用户调用 API 消耗 1000 quota
  │
  ├─ 第一优先: free_quota（注册赠送/永久）
  │   剩余 500 → 扣 500，剩 0
  │
  ├─ 第二优先: recharge_quota（充值购买，FIFO 按过期时间）
  │   record_1: 剩余 200（先过期）→ 扣 200，剩 0
  │   record_2: 剩余 800 → 扣 300，剩 500
  │
  └─ 第三优先: vip_quota（VIP 套餐内含）
      （上层已扣完，不触发）
```

### 5.2 事务边界

```go
func (s *BillingService) PostConsumeQuota(...) error {
    return s.userRepo.Transaction(func(tx *gorm.DB) error {
        // 1. 扣减 free_quota
        // 2. 扣减 recharge_records (FIFO)
        // 3. 扣减 vip_quota
        // 4. 更新 token.used_quota / remain_quota
        // 5. 写入 quota_transaction 记录
        // 6. 写入 usage_log
        // 任一步失败 → 整个事务回滚
    })
}
```

---

## 六、测试计划

### 6.1 单元测试

| 编号 | 测试用例 | 文件 |
|------|---------|------|
| UT-01 | Token IP 白名单：允许 IP 通过 | middleware/jwt_test.go |
| UT-02 | Token IP 白名单：拒绝非白名单 IP | middleware/jwt_test.go |
| UT-03 | Token IP 白名单：空列表不校验 | middleware/jwt_test.go |
| UT-04 | Token 模型白名单：允许模型通过 | middleware/jwt_test.go |
| UT-05 | Token 模型白名单：拒绝非允许模型 | middleware/jwt_test.go |
| UT-06 | Token 模型黑名单：拒绝黑名单模型 | middleware/jwt_test.go |
| UT-07 | PostConsumeQuota 事务：free→recharge→vip 扣减链路 | service/billing_test.go |
| UT-08 | PostConsumeQuota 事务：配额不足时回滚 | service/billing_test.go |
| UT-09 | consumeRechargeQuota FIFO 扣减 | service/billing_test.go |
| UT-10 | TokenRateLimit 自定义 RPM 生效 | middleware/ratelimit_test.go |

### 6.2 集成测试

| 编号 | 测试场景 | 验证点 |
|------|---------|--------|
| IT-01 | 完整非流式调用链路 | Token 校验 → 配额预检 → 渠道选择 → 转发 → 配额扣减 → 日志 |
| IT-02 | 完整流式调用链路 | 同上 + 流式 chunk 推送 + 流式扣费 |
| IT-03 | Token 过期拒绝 | expired token → 401 |
| IT-04 | 配额耗尽拒绝 | 配额不足 → 402 |
| IT-05 | 渠道故障自动切换 | 第一个渠道超时 → 自动切换到第二个 |
| IT-06 | 所有渠道失败 | 3 次重试均失败 → 502 |
| IT-07 | /v1/models 模型列表 | 返回用户有权访问的模型列表 |
| IT-08 | 并发配额扣减正确性 | 10 并发各扣 100，最终配额正确 |

### 6.3 压力测试

| 场景 | 并发数 | 持续时间 | 关注指标 |
|------|--------|---------|---------|
| 非流式 Chat API | 50 | 5 min | TPS, P95 延迟, 错误率 |
| 流式 Chat API | 30 | 5 min | TPS, P95 延迟, 连接稳定性 |
| /v1/models 列表 | 100 | 2 min | TPS, P99 延迟 |
| 混合负载 | 50 chat + 20 embed + 30 models | 10 min | 整体稳定性 |
| 配额扣减正确性 | 20 并发 × 100 请求 | - | 最终余额 = 初始 - 消耗 |

---

## 七、实现顺序

```
Phase 1: 安全加固（P0）── 优先级最高
├── T-01: TokenAuth IP 白名单校验          ~2h
├── T-02: TokenAuth 模型权限校验            ~3h
├── T-03: 流式配额扣减                      ~4h
├── T-04: 非流式配额扣减                    ~2h
└── Phase 1 集成测试                        ~2h

Phase 2: 限速与计费（P1）
├── T-05: TokenRateLimit 自定义 RPM/TPM     ~3h
├── T-06: PostConsumeQuota 事务保护          ~2h
├── T-07: consumeRechargeQuota 实现          ~3h
├── T-08: APIAccessLog model 字段修复        ~1h
└── Phase 2 单元测试 + 集成测试              ~3h

Phase 3: 功能补全（P1-P2）
├── T-09: /v1/completions 端点              ~3h
├── T-10: APIError 结构统一                  ~1h
├── T-11: Token 列表返回配额使用情况          ~1h
└── Phase 3 测试                            ~2h

Phase 4: 压力测试（P2）
├── T-12: 压力测试脚本 + 执行               ~4h
├── T-13: 监控增强                          ~2h
└── Phase 4 调优 + 回归测试                  ~4h

预估总工时: ~42h（约 5-6 个工作日）
```

---

## 八、完成标准

- [ ] 任意 Token 的 IP 白名单限制生效，非白名单 IP 请求返回 403
- [ ] 任意 Token 的模型白名单/黑名单限制生效
- [ ] 非流式和流式 API 调用均正确扣减配额
- [ ] 流式调用结束后配额扣减有兜底机制
- [ ] Token 级 RPM/TPM 限速按用户配置生效
- [ ] 充值配额按 FIFO 正确扣减
- [ ] PostConsumeQuota 在 DB 事务中执行
- [ ] /v1/completions 端点可用
- [ ] 所有 /v1/* 接口错误格式统一为 OpenAI 标准
- [ ] APIAccessLog 中 model 字段正确记录
- [ ] 单元测试 ≥ 90% 覆盖新增代码
- [ ] 集成测试全链路通过
- [ ] 压力测试 50 并发 P95 < 3s，错误率 < 1%
- [ ] 配额并发扣减正确性验证通过

---

## 九、风险与回退

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 流式扣费无精确 usage | 配额扣减不准 | 兜底使用估算值，标记为 estimated=true |
| TokenAuth 读取 body 影响性能 | P99 延迟增加 | 仅对 completions/embeddings 路径读取 body |
| 事务锁影响并发性能 | TPS 下降 | 事务内只做写操作，查询放事务外 |
| 压测发现性能瓶颈 | 上线延迟 | Phase 4 可推迟，先保证功能正确性 |

---

**文档版本**: v1.0
**创建人**: TOM
**下一步**: 审批通过后按 Phase 1 → Phase 4 顺序实施
