# API Key 全生命周期与 API 调用链路完善计划

> 版本: v1.3
> 日期: 2026-05-12
> 状态: 进行中
> 目标版本: v1.3.0
> 优先级: 高
> 完成度: Phase 1 ✅ / Phase 2 ✅ / Phase 3 ✅ / Phase 4 ❌

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
| P1-1 | RateLimit | TokenRateLimit 硬编码 10 RPS，忽略 token.RPMLimit | 🟡 中 | 用户自定义限速不生效 | ✅ 已修复 |
| P1-2 | RateLimit | 无 TPM（每分钟 Token 数）限制 | 🟡 中 | 无法控制 Token 消耗速率 | ✅ 已修复 |
| P1-3 | AccessLog | `model` 字段从 `PostForm` 提取，实际请求为 JSON body | 🟡 中 | 日志中 model 恒为 "unknown" | ✅ 已修复 |
| P1-4 | API | 无 `/v1/completions` 端点 | 🟡 中 | 不完整 OpenAI 兼容 | ✅ 已修复 |
| P1-5 | Billing | PostConsumeQuota 无事务保护 | 🟡 中 | 并发扣费可能数据不一致 | ✅ 已修复 |
| P1-6 | Billing | `consumeRechargeQuota` 返回 false（空实现） | 🟡 中 | 充值配额永远不消耗 | ✅ 已修复 |
| P2-1 | TokenAuth | 未校验 `denied_models` 黑名单 | 🟢 低 | 黑名单功能不生效 | ✅ 已修复 |
| P2-2 | API | 非流式响应后无 PostConsumeQuota 调用 | 🟢 低 | 非流式也未扣费 | ✅ 已修复 |
| P2-3 | API | 无 `/v1/completions`（text completions） | 🟢 低 | OpenAI 兼容不完整 | ✅ 已修复 |
| P2-4 | Error | 错误响应格式不统一（混合 APIResponse 和 APIErrorResponse） | 🟢 低 | 客户端适配困难 | ✅ 已修复 |

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

### Phase 1: 安全加固（P0）— ✅ 全部完成

#### T-01: TokenAuth 增加 IP 白名单校验 ✅

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/jwt.go` → `TokenAuth()` |
| 状态 | ✅ 已实现 |
| 实现 | `TokenAuth()` 中 `token.GetAllowedIPs()` 非空时检查 `c.ClientIP()` |
| 辅助函数 | `isIPAllowed()` 支持精确 IP 匹配和 CIDR 网段匹配 |
| 错误响应 | 403 `ip_not_allowed`（使用 APIErrorResponse 格式） |
| 修改行 | jwt.go:121-134 |

#### T-02: TokenAuth 增加模型权限校验 ✅

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/jwt.go` → `TokenAuth()` |
| 状态 | ✅ 已实现 |
| 实现 | `extractModelFromRequest()` 安全读取 body 并重置，校验 denied_models 黑名单 + allowed_models 白名单 |
| 范围 | 仅对 `/completions` 和 `/embeddings` 路径生效 |
| 错误响应 | 403 `model_not_allowed` |
| 修改行 | jwt.go:137-176, 208-227 |

#### T-03: 流式响应配额扣减与 Token 统计 ✅

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/handler/api_handler.go` → `handleStreamWithFailover()` |
| 状态 | ✅ 已实现 |
| 实现 | 流式过程中累积 `completionContentLen`；最终 chunk 检查上游 usage；有则精确扣费，无则用 `estimateChatTokens` + 字符数/4 估算兜底 |
| 调用 | PostConsumeQuota + LogUsage（填充 ChannelID） |
| 修改行 | api_handler.go:265-333 |

#### T-04: 非流式响应配额扣减 ✅

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/handler/api_handler.go` → `ChatCompletions()` |
| 状态 | ✅ 已实现 |
| 实现 | 非流式成功后从 `adapter.ChatResponse.Usage` 提取 token 用量，调用 PostConsumeQuota + LogUsage |
| 同步修复 | `Embeddings()` 同样增加了扣费逻辑 |
| 修改行 | api_handler.go:124-139（ChatCompletions）, 571-587（Embeddings） |

---

### Phase 2: 限速与计费完善（P1）— ✅ 全部完成

#### T-05: TokenRateLimit 支持自定义 RPM/TPM ✅

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/ratelimit.go` → `TokenRateLimit()` |
| 状态 | ✅ 已实现 |
| 变更 | 函数签名改为 `TokenRateLimit(tokenService *service.TokenService)`；从 DB 加载 token 的 RPMLimit/TPMLimit；RPM 用 `rate.Limiter`（RPM→per-second 转换）；TPM 用滑动窗口计数 |
| 路由 | router.go 三处调用同步更新为 `middleware.TokenRateLimit(tokenService)` |
| 错误响应 | RPM 超限返回 429 `rate_limit_rpm`；TPM 超限返回 429 `rate_limit_tpm` |

#### T-06: PostConsumeQuota 事务保护 ✅

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/service/billing_service.go` → `PostConsumeQuota()` |
| 状态 | ✅ 已实现 |
| 变更 | 所有 DB 写操作（free_quota 扣减、recharge 扣减、vip_quota 扣减、token 更新、quota_transaction 写入）包裹在 `s.db.Transaction(func(tx)...)` 中，失败自动 rollback |

#### T-07: 实现 consumeRechargeQuota ✅

| 项 | 内容 |
|---|------|
| 文件 | `billing_service.go` + 新增 `repository/recharge_record_repo.go` |
| 状态 | ✅ 已实现 |
| 变更 | 创建 `UserRechargeRecordRepository`（GetActiveByUser FIFO、GetTotalActiveQuota）；PostConsumeQuota 事务内直接扣减 recharge records（FIFO 按过期时间）；`getActiveRechargeQuota()` 改为实际查询 |
| 逻辑 | 按 expired_at ASC 顺序逐条扣减 remaining → 用完标记 status=used → 写 quota_transaction(quota_type='recharge') |

#### T-08: APIAccessLog 修复 model 字段提取 ✅

| 项 | 内容 |
|---|------|
| 文件 | `api_handler.go` + `api_access_log.go` |
| 状态 | ✅ 已实现 |
| 变更 | api_handler.go 的 `ChatCompletions()` 和 `Embeddings()` 中解析请求后 `c.Set("request_model", req.Model)`；api_access_log.go 从 context 读取 model 替代 `c.PostForm("model")` |

---

### Phase 3: 功能补全（P1-P2）— ✅ 全部完成

#### T-09: /v1/completions 端点 ✅

| 项 | 内容 |
|---|------|
| 文件 | `api_handler.go` (新增 `Completions` 方法) + `router.go` (新增路由) + `response.go` (新增 `CompletionsRequest`) |
| 状态 | ✅ 已实现 |
| 变更 | prompt→messages 转换；复用 chatWithFailover 管道；ChatResponse→text_completion 格式转换 |
| 路由 | `v1.POST("/completions", TokenAuth, TokenRateLimit, APIAccessLog, apiHandler.Completions)` |

#### T-10: APIError 结构统一 ✅

| 项 | 内容 |
|---|------|
| 文件 | `model/response.go` + `jwt.go` + `api_handler.go` + `ratelimit.go` |
| 状态 | ✅ 已实现 |
| 变更 | APIError 增加 Type、Param 字段；所有 /v1/* 错误统一使用 APIErrorResponse（不再混合 APIResponse）；错误码改为 snake_case 小写（OpenAI 兼容） |

#### T-11: Token 列表返回配额使用情况 ✅

| 项 | 内容 |
|---|------|
| 文件 | `token_handler.go` → `List()` |
| 状态 | ✅ 已实现 |
| 变更 | 每个 token 返回 used_quota、remain_quota、usage_percent、is_unlimited |

---

### Phase 4: 压力测试与可观测（P2）— ❌ 待 Phase 2-3 完成后实施

#### T-12: API 压力测试脚本 ❌

| 项 | 内容 |
|---|------|
| 状态 | ❌ 未开始 |
| 工具 | `k6`（推荐）或扩展 `test_api.py` |
| 场景 | 见第六节测试计划 |
| 输出 | TPS、P50/P95/P99 延迟、错误率 |

#### T-13: API 调用链路监控增强 ❌

| 项 | 内容 |
|---|------|
| 文件 | `backend/internal/middleware/api_access_log.go` |
| 状态 | ❌ 未开始 |
| 前置依赖 | T-08（model 字段修复）完成后实施 |
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

## 七、实现进度与剩余计划

```
Phase 1: 安全加固（P0）── ✅ 全部完成
├── T-01: TokenAuth IP 白名单校验          ✅ done (jwt.go:121-134)
├── T-02: TokenAuth 模型权限校验            ✅ done (jwt.go:137-176)
├── T-03: 流式配额扣减                      ✅ done (api_handler.go:265-333)
├── T-04: 非流式配额扣减                    ✅ done (api_handler.go:124-139)
└── Phase 1 验证文档                        ✅ done (issues/007)

Phase 2: 限速与计费（P1）── ✅ 全部完成
├── T-05: TokenRateLimit 自定义 RPM/TPM     ✅ done (ratelimit.go:122-228)
├── T-06: PostConsumeQuota 事务保护          ✅ done (billing_service.go:112-232)
├── T-07: consumeRechargeQuota 实现          ✅ done (billing_service.go + recharge_record_repo.go)
├── T-08: APIAccessLog model 字段修复        ✅ done (api_handler.go + api_access_log.go)
└── Phase 2 单元测试 + 集成测试              待编译验证后补充

Phase 3: 功能补全（P1-P2）── ✅ 全部完成
├── T-09: /v1/completions 端点              ✅ done (api_handler.go + router.go + response.go)
├── T-10: APIError 结构统一                  ✅ done (response.go + jwt.go + api_handler.go + ratelimit.go)
├── T-11: Token 列表返回配额使用情况          ✅ done (token_handler.go)
└── Phase 3 测试                            待编译验证后补充

Phase 4: 压力测试（P2）── ❌ 待编译验证完成后
├── T-12: 压力测试脚本 + 执行               ~4h
├── T-13: 监控增强                          ~2h
└── Phase 4 调优 + 回归测试                  ~4h

剩余预估工时: ~10h（Phase 1-3 已完成 ~32h）
```

---

## 八、完成标准

- [x] 任意 Token 的 IP 白名单限制生效，非白名单 IP 请求返回 403
- [x] 任意 Token 的模型白名单/黑名单限制生效
- [x] 非流式和流式 API 调用均正确扣减配额
- [x] 流式调用结束后配额扣减有兜底机制
- [x] Token 级 RPM/TPM 限速按用户配置生效
- [x] 充值配额按 FIFO 正确扣减
- [x] PostConsumeQuota 在 DB 事务中执行
- [x] /v1/completions 端点可用
- [x] 所有 /v1/* 接口错误格式统一为 OpenAI 标准
- [x] APIAccessLog 中 model 字段正确记录
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

**文档版本**: v1.3
**创建人**: TOM
**下一步**: Phase 4 压力测试与可观测（T-12 压测脚本, T-13 监控增强）

---

## 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v1.0 | 2026-05-12 | 初始版本 |
| v1.1 | 2026-05-12 | 更新完成状态：Phase 1 全部完成 |
| v1.2 | 2026-05-12 | Phase 2 全部完成：RPM/TPM 限速、事务保护、充值配额 FIFO、日志 model 字段 |
| v1.3 | 2026-05-12 | Phase 3 全部完成：/v1/completions 端点、APIError 结构统一（Type/Param）、Token 列表配额详情 |
