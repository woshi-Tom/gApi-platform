# #007 P0 API 调用链路修复

> 日期: 2026-05-12
> 关联计划: [api-key-lifecycle-plan.md](../plans/api-key-lifecycle-plan.md)
> 状态: 代码完成，待编译验证
> 编译测试: [007-compile-test.md](007-compile-test.md)

---

## 修复内容

### T-01: TokenAuth IP 白名单校验
- **文件**: `backend/internal/middleware/jwt.go`
- **变更**: TokenAuth 中间件在 token 校验通过后，检查 `token.allowed_ips` 是否包含客户端 IP
- **新增函数**: `isIPAllowed()` 支持精确匹配和 CIDR 网段匹配

### T-02: TokenAuth 模型权限校验
- **文件**: `backend/internal/middleware/jwt.go`
- **变更**: TokenAuth 中间件读取请求体中的 `model` 字段，校验 `allowed_models` 白名单和 `denied_models` 黑名单
- **新增函数**: `extractModelFromRequest()` 安全读取 body 并重置，仅对 `/completions` 和 `/embeddings` 路径生效

### T-04: 非流式响应配额扣减
- **文件**: `backend/internal/handler/api_handler.go`
- **变更**:
  - `ChatCompletions()`: 非流式成功后从 `adapter.ChatResponse.Usage` 提取 token 用量，调用 `PostConsumeQuota` + `LogUsage`
  - `Embeddings()`: 同理，从 `adapter.EmbeddingsResponse.Usage` 提取

### T-03: 流式响应配额扣减
- **文件**: `backend/internal/handler/api_handler.go`
- **变更**: `handleStreamWithFailover()` 中：
  - 流式过程中累积 `completionContentLen`（返回内容总字符数）
  - 最终 chunk（`chunk.Done == true`）检查是否包含 `Usage` 字段
  - 有上游 usage → 精确扣费；无上游 usage → 用 `estimateChatTokens` + 内容长度/4 估算兜底
  - 流程结束后调用 `PostConsumeQuota` + `LogUsage`

---

## 编译验证步骤

```bash
cd backend
go build ./...
```

预期：无编译错误。

---

## 需要补充的测试

### 1. 单元测试

#### middleware/jwt_test.go（新增或扩展）

| 用例 | 说明 |
|------|------|
| `TestIsIPAllowed_ExactMatch` | 精确 IP 匹配通过 |
| `TestIsIPAllowed_CIDRMatch` | CIDR 网段匹配通过 |
| `TestIsIPAllowed_Denied` | 非白名单 IP 返回 false |
| `TestIsIPAllowed_EmptyList` | 空列表不校验（任何 IP 通过）|
| `TestExtractModelFromRequest_ChatCompletions` | 从 `/v1/chat/completions` body 提取 model |
| `TestExtractModelFromRequest_Embeddings` | 从 `/v1/embeddings` body 提取 model |
| `TestExtractModelFromRequest_Models` | `/v1/models` 路径返回空字符串 |
| `TestExtractModelFromRequest_BodyReset` | 读取后 body 可被下游再次读取 |

#### handler/api_handler_test.go（扩展）

| 用例 | 说明 |
|------|------|
| `TestChatCompletions_PostConsumeQuota` | 非流式调用后 billingService 被调用 |
| `TestChatCompletions_Streaming_PostConsumeQuota` | 流式结束后 billingService 被调用 |
| `TestEmbeddings_PostConsumeQuota` | Embeddings 调用后 billingService 被调用 |

### 2. 集成测试

| 场景 | 请求 | 预期结果 |
|------|------|---------|
| Token 无 IP 限制 | POST /v1/chat/completions (任意 IP) | 200 正常通过 |
| Token 设置 IP 白名单，匹配 | POST /v1/chat/completions (白名单 IP) | 200 正常通过 |
| Token 设置 IP 白名单，不匹配 | POST /v1/chat/completions (非白名单 IP) | 403 `ip_not_allowed` |
| Token 设置 allowed_models，匹配 | POST /v1/chat/completions {"model":"gpt-4"} | 200 正常通过 |
| Token 设置 allowed_models，不匹配 | POST /v1/chat/completions {"model":"claude-3"} | 403 `model_not_allowed` |
| Token 设置 denied_models | POST /v1/chat/completions {"model":"gpt-4"} | 403 `model_not_allowed` |
| 非流式调用后配额扣减 | 调用前后查询 user 配额 | 配额减少 = 实际 usage |
| 流式调用后配额扣减 | 流式调用后查询 user 配额 | 配额减少（精确或估算）|
| Embeddings 调用后配额扣减 | 调用前后查询 user 配额 | 配额减少 |

### 3. 压力测试（后续阶段）

| 场景 | 并发 | 关注指标 |
|------|------|---------|
| TokenAuth + IP/模型检查开销 | 50 | P99 延迟增量（预期 < 1ms）|
| 流式配额扣减并发正确性 | 20 × 50 请求 | 最终余额 = 初始 - 总消耗 |

---

## 修改文件清单

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `backend/internal/middleware/jwt.go` | 修改 | 新增 `isIPAllowed()`, `extractModelFromRequest()`；`TokenAuth()` 增加 IP 和模型校验 |
| `backend/internal/handler/api_handler.go` | 修改 | `ChatCompletions()` 增加非流式扣费；`handleStreamWithFailover()` 增加流式扣费；`Embeddings()` 增加扣费 |

---

## 注意事项

1. `extractModelFromRequest` 会读取 request body 并重置，不影响下游 handler 正常解析
2. 流式扣费的 fallback 估算精度有限（prompt 估算 /4，completion 按字符数 /4），后续可优化
3. `PostConsumeQuota` 当前无事务保护（T-06 计划在 Phase 2 修复），并发场景下可能有竞态
4. `service` 包中 `UsageRecord` 结构体已包含 `ChannelID` 字段，流式扣费时已填充
