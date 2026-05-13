# Issue #014: Anthropic Messages API 兼容 — 编译测试

> 分支: `dev-tom`
> 关联: [chat-completions-content-format-fix.md](../plans/chat-completions-content-format-fix.md) 方案 C
> 本次提交: Anthropic Messages API 入站端点 + x-api-key 认证 + Claude adapter 流式

---

## 目标

验证 Anthropic Messages API 兼容代码编译通过、vet 无警告、现有测试无回归。

---

## 变更文件清单

| 文件 | 类型 | 变更说明 |
|------|------|---------|
| `backend/internal/model/anthropic.go` | 新建 | Anthropic 请求/响应/流式事件结构体 |
| `backend/internal/handler/anthropic_convert.go` | 新建 | Anthropic↔OpenAI 格式转换函数 |
| `backend/internal/middleware/jwt.go` | 修改 | TokenAuth 支持 `x-api-key` header；`extractModelFromRequest` 支持 `/messages` 路径 |
| `backend/internal/router/router.go` | 修改 | 注册 `POST /api/v1/messages` 路由；CORS 增加 `X-Api-Key` |
| `backend/internal/handler/api_handler.go` | 修改 | 新增 `Messages` handler + `handleMessagesStream` 流式处理 |
| `backend/internal/pkg/adapter/claude.go` | 修改 | 修复 system 消息 bug（提取到顶层 `system` 字段）；实现 `ChatStream` 流式 |
| `backend/internal/handler/anthropic_convert_test.go` | 新建 | 20 个用例：请求/响应转换、stop_reason 映射、content 提取 |
| `backend/internal/pkg/adapter/claude_test.go` | 新建 | system 消息提取 + Anthropic SSE 流式解析 + 空 body |
| `backend/internal/middleware/middleware_test.go` | 修改 | x-api-key header 认证 + Bearer 回归测试（+4 用例） |

---

## 编译验证命令

```bash
cd backend
go build ./...
go vet ./...
```

**预期**: 无编译错误，无 vet 警告。

---

## 测试验证命令

```bash
cd backend
go test ./... -count=1 -timeout 120s
```

**预期**: 所有现有测试通过（无回归）。

---

## 重点检查项

### 1. 新文件编译
- `backend/internal/model/anthropic.go` — 结构体定义无语法错误
- `backend/internal/handler/anthropic_convert.go` — 引用 `model.Anthropic*` 和 `adapter.ChatResponse` 类型

### 2. jwt.go 修改
- `TokenAuth` 函数: `x-api-key` header 读取 + `Authorization: Bearer` 兜底逻辑
- `extractModelFromRequest`: 新增 `strings.Contains(path, "/messages")` 判断

### 3. api_handler.go 修改
- `Messages` 方法: 绑定 `model.AnthropicMessagesRequest`，调用 `anthropicToOpenAIChatRequest`，复用 `chatWithFailover`
- `handleMessagesStream` 方法: 流式 Anthropic SSE 输出，调用 `writeAnthropicSSEvent`
- 引用 `service.ErrQuotaInsufficient`、`service.ErrTokenDisabled`、`service.ErrTokenExpired`

### 4. claude.go 修改
- `Chat` 方法: system 消息提取逻辑（`strings.Join` 需要 `strings` 包已导入）
- `ChatStream` 方法: 完整流式实现，解析 Anthropic SSE event/data 格式
- `mapStopReasonReverseStream` 函数: stop_reason 映射

### 5. router.go 修改
- 路由注册: `v1.POST("/messages", ...)` 在 `/chat/completions` 之后
- CORS: `X-Api-Key` 已加入 `Access-Control-Allow-Headers`

---

## 已知风险点

1. **SSE 格式**: `writeAnthropicSSEvent` 使用 `fmt.Fprintf` 直接写入 `io.Writer`，不使用 Gin 的 `c.SSEvent()`（后者固定用 `event: message`）。需要确认 Gin 的 `c.Stream` 回调中的 `io.Writer` 可以直接写入。

2. **Claude adapter 流式**: `ChatStream` 中的匿名结构体与 `ChatStreamResponse.Choices` 类型必须完全匹配（含嵌套 struct）。如果编译报类型不匹配，需调整。

3. **import 依赖**: `api_handler.go` 新增了对 `model.AnthropicErrorResponse` 等类型的引用，需确认 `model` 包导入已存在。

---

## 验证完成后

- [x] `go build ./...` 通过
- [x] `go vet ./...` 无警告
- [x] `go test ./... -count=1 -timeout 120s` 全部通过
- [x] 将结果更新到本文档下方的 **测试报告** 章节

---

## 测试报告

> 测试时间: 2026-05-13
> 测试人: 远程编译测试 agent

### 第一轮（Anthropic 入站功能）

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `go build ./...` | ✅ 通过 | 无编译错误 |
| `go vet ./...` | ✅ 通过 | 无警告 |
| `go test ./...` | ✅ 全部通过 | handler 1.402s / middleware 0.023s / model 0.018s / repository 0.171s / service 0.034s |
| 新增 normalize 测试 | ✅ 4/4 通过 | `TestNormalizeMessages_*` |

### 第二轮（补充测试）

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `go build ./...` | ✅ 通过 | 无编译错误 |
| `go vet ./...` | ✅ 通过 | 无警告 |
| `go test ./...` | ✅ 全部通过 | handler 1.312s / middleware 0.039s / model 0.016s / adapter 0.007s / repository 0.143s / service 0.031s |
| anthropic_convert_test.go | ✅ 24/24 通过 | 请求转换、响应转换、stop_reason 映射、content 提取、normalize |
| claude_test.go | ✅ 6/6 通过 | system 提取、SSE 流式解析、空 body |
| middleware_test.go 新增 | ✅ 4/4 通过 | x-api-key 认证、Bearer 回归（已修复 `5ee1003`） |
