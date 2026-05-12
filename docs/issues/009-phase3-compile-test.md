# #009 Phase 3 编译测试指南

> 日期: 2026-05-12
> 关联: [api-key-lifecycle-plan.md](../plans/api-key-lifecycle-plan.md) Phase 3（P1-P2）
> 状态: 待编译验证

---

## 一、需要编译验证的变更

| 文件 | 变更内容 |
|------|---------|
| `backend/internal/model/response.go` | APIError 增加 Type/Param 字段；新增 CompletionsRequest |
| `backend/internal/handler/api_handler.go` | 新增 Completions() handler；所有错误增加 Type；代码格式统一小写 |
| `backend/internal/handler/token_handler.go` | List() 返回 used_quota/remain_quota/usage_percent |
| `backend/internal/middleware/jwt.go` | TokenAuth 所有错误统一 APIErrorResponse + Type |
| `backend/internal/router/router.go` | 新增 `/v1/completions` 路由 |

---

## 二、编译步骤

```bash
cd backend
go build ./...
go vet ./...
go test ./... -count=1 -timeout 120s
```

---

## 三、功能验证清单

### T-09: /v1/completions

```
验证点 1: 简单 prompt 调用
  curl POST /v1/completions -d '{"model":"gpt-3.5-turbo","prompt":"Hello"}'
  预期: 200 {"object":"text_completion","choices":[...]}

验证点 2: 数组 prompt
  curl POST /v1/completions -d '{"model":"gpt-3.5-turbo","prompt":["Hello","World"]}'
  预期: 200

验证点 3: 缺少 model → 400
  预期: {"error":{"type":"invalid_request_error","code":"missing_model",...}}
```

### T-10: APIError 统一

```
验证点 1: 所有错误包含 type/code/message 三字段
  - 401 invalid_api_key
  - 401 expired_api_key
  - 403 model_not_allowed
  - 403 ip_not_allowed
  - 400 invalid_request
  - 402 quota_insufficient
  - 429 rate_limit_rpm / rate_limit_tpm
  - 502/503 upstream_error / no_available_channel

验证点 2: 无 model.APIResponse 残留（仅 APIErrorResponse）
```

### T-11: Token 配额

```
验证点: GET /api/v1/tokens 返回 per-token 配额
  预期: 每个 token 包含 used_quota, remain_quota, usage_percent, is_unlimited
```

---

## 四、验证结论模板

```
编译结果: ✅ / ❌
Go Vet:   ✅ / ❌
测试:     ✅ / ❌
T-09:     ✅ / ❌ / ⏭️
T-10:     ✅ / ❌ / ⏭️
T-11:     ✅ / ❌ / ⏭️
结论:     可合入 / 需修复
```
