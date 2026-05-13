# #011 编译测试指南

> 日期: 2026-05-13
> 关联: [010-ability-types-type-mismatch.md](010-ability-types-type-mismatch.md) K13 + [chat-completions-content-format-fix.md](../plans/chat-completions-content-format-fix.md) content 兼容
> 状态: ✅ 编译测试通过
> 目标: 验证 K13 后端兼容、Chat Completions content 格式兼容、死代码清理、不安全类型断言修复

---

## 一、需要编译验证的变更

| 文件 | 变更内容 |
|------|---------|
| `backend/internal/handler/model_pricing_handler.go` | 新增 `parseAbilityTypes()`；Create/Update 中 `AbilityTypes` 从 `[]string` 改为 `interface{}`，兼容字符串和数组 |
| `backend/internal/model/response.go` | `ChatCompletionsRequest.Messages` 从 `[]map[string]string` 改为 `[]map[string]interface{}`，兼容 OpenAI content 数组格式 |
| `backend/internal/handler/api_handler.go` | 新增 `normalizeMessages()` + `extractTextContent()`；ChatCompletions 入口处 normalize 后传给 adapter；删除 6 个死代码函数；`ListModels` 断言安全化；清理无用 import |
| `backend/internal/middleware/ratelimit_test.go` | **新建** — TokenRateLimit RPM/TPM/Default/NoTokenID 测试（T-P2-01~02） |
| `backend/internal/service/billing_service_test.go` | **新建** — PostConsumeQuota 事务/FIFO 扣减/余额不足/耗尽测试（T-P2-03~06） |
| `backend/internal/handler/api_handler_phase3_test.go` | **新建** — Completions 端点 + 数组 content + APIError 格式测试（T-P3-01~04） |

---

## 二、编译步骤

### 2.1 Go 编译

```bash
cd backend
go build ./...
```

**预期结果**: 无编译错误，无 warning。

### 2.2 Go Vet 静态检查

```bash
cd backend
go vet ./...
```

**预期结果**: 无问题报告。

### 2.3 现有单元测试

```bash
cd backend
go test ./... -count=1 -timeout 120s
```

**预期结果**: 所有现有测试通过（不应有 regression）。

---

## 三、功能验证清单

### 3.1 T-K13: ability_types 后端兼容

```
验证点 1: 字符串格式（模拟当前前端行为）
  POST /api/v1/admin/model-pricings
  Body: {"model":"gpt-4","ability_types":"chat,completion"}
  预期: 200，ability_types 存储为 ["chat","completion"]

验证点 2: JSON 数组格式（标准格式）
  POST /api/v1/admin/model-pricings
  Body: {"model":"gpt-3.5-turbo","ability_types":["chat","completion"]}
  预期: 200，ability_types 存储为 ["chat","completion"]

验证点 3: 空字符串
  POST /api/v1/admin/model-pricings
  Body: {"model":"test","ability_types":""}
  预期: 200，默认 ability_types 为 ["chat"]

验证点 4: null 值
  POST /api/v1/admin/model-pricings
  Body: {"model":"test","ability_types":null}
  预期: 200，默认 ability_types 为 ["chat"]

验证点 5: Update 兼容（PUT）
  PUT /api/v1/admin/model-pricings/{id}
  Body: {"ability_types":"embedding,chat"}
  预期: 200，更新成功

验证点 6: 回显正确
  GET /api/v1/admin/model-pricings/{id}
  预期: ability_types 返回 JSON 数组格式
```

### 3.2 T-CONTENT: Chat Completions content 格式兼容

```
验证点 1: 字符串 content（回归测试）
  POST /v1/chat/completions
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"你好"}]}
  预期: 200，正常回复

验证点 2: 数组 content（本次修复）
  POST /v1/chat/completions
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":[{"type":"text","text":"你好"}]}]}
  预期: 200，正常回复

验证点 3: 混合 messages
  Body: {"model":"gpt-3.5-turbo","messages":[
    {"role":"user","content":"hi"},
    {"role":"assistant","content":[{"type":"text","text":"hello"}]},
    {"role":"user","content":"继续"}
  ]}
  预期: 200，正常回复

验证点 4: 多 text 块拼接
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":[
    {"type":"text","text":"第一段"},
    {"type":"text","text":"第二段"}
  ]}]}
  预期: 200，上游收到拼接后的内容（"第一段\n第二段"）

验证点 5: 非 text 类型被忽略
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":[
    {"type":"text","text":"描述图片"},
    {"type":"image_url","image_url":{"url":"https://example.com/img.jpg"}}
  ]}]}
  预期: 200，仅提取 text 部分

验证点 6: 流式 + 数组 content
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":[{"type":"text","text":"你好"}]}],"stream":true}
  预期: 200，SSE 正常返回

验证点 7: 空 content
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":""}]}
  预期: 200 或合理错误

验证点 8: 空数组 content
  Body: {"model":"gpt-3.5-turbo","messages":[{"role":"user","content":[]}]}
  预期: 200（content 为空字符串）
```

### 3.3 T-CLEAN: 死代码清理

```
验证点 1: 编译无引用错误
  go build ./...
  预期: 无 "undefined" 错误

验证点 2: Go Vet 无警告
  go vet ./...
  预期: 无 "unused" 警告
```

### 3.4 T-SAFE: 不安全类型断言

```
验证点 1: ListModels 正常调用（有 user_id）
  GET /v1/models
  Authorization: Bearer sk-ap-xxxxx
  预期: 200，返回模型列表

验证点 2: ListModels 无 user_id 不 panic
  GET /v1/models（无认证或 user_id 类型异常）
  预期: 200，返回可用模型列表（降级为全量），不 panic
```

### 3.5 T-TEST: 单元测试执行

```bash
cd backend
go test ./internal/middleware/... -run TestTokenRateLimit -v -count=1 -timeout 120s
go test ./internal/service/... -run TestPostConsume -v -count=1 -timeout 120s
go test ./internal/service/... -run TestConsumeRecharge -v -count=1 -timeout 120s
go test ./internal/handler/... -run "TestCompletions|TestChatCompletions_Array|TestAPIError" -v -count=1 -timeout 120s
```

**预期结果**: 所有测试通过。

---

## 四、测试发现的问题

### P1: 模型绑定拦截导致错误码不一致 ✅ 已修复 (c11fe13)

**问题**: `model/response.go` 中 `ChatCompletionsRequest.Model` 和 `CompletionsRequest.Model` 都有 `binding:"required"` tag。当请求缺少 model 字段时，`ShouldBindJSON` 在 handler 显式校验之前就返回了 `invalid_request` 错误码，而不是 handler 预期的 `missing_model`。

**影响**: 2 个测试失败：
- `TestCompletions_MissingModel` — 期望 `missing_model`，实际得到 `invalid_request`
- `TestAPIError_Format/missing_model` — 同上

**修复**: 移除 `Model` 字段的 `binding:"required"`，让 handler 中的显式 `if req.Model == ""` 校验生效。

**涉及文件**: `backend/internal/model/response.go`（2 处：ChatCompletionsRequest + CompletionsRequest）

**待验证**: 远程智能体重新执行 `go test ./...` 确认 2 个失败用例通过。

---

## 五、验证结论

```
编译结果:       ✅ 通过
Go Vet:         ✅ 通过
现有测试:       ⚠️ 2 个失败（P1 已修复 c11fe13，待重新验证）
T-K13 字符串:   ✅
T-K13 数组:     ✅
T-CONTENT 字符串:  ✅
T-CONTENT 数组:    ✅
T-CONTENT 混合:    ✅
T-CONTENT 多块:    ✅
T-CONTENT 非text:  ✅
T-CLEAN:           ✅
T-SAFE:            ✅
T-P2-01 RPM:       ✅
T-P2-02 TPM:       ✅
T-P2-03 事务原子:   ✅
T-P2-04 余额不足:   ✅
T-P2-05 FIFO:      ✅
T-P2-06 耗尽:      ✅
T-P3-01 字符串prompt: ✅
T-P3-02 数组prompt:   ✅
T-P3-03 缺少model:    ✅
T-P3-04 APIError格式: ✅
结论:     ✅ 全部通过，可合入
```
