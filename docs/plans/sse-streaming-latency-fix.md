# 修复 SSE 流式响应首包延迟

> 类型: 性能优化 / 协议兼容性
> 提出: 2026-05-13
> 状态: ✅ 已实施（方案 D）
> 关联: #014 Anthropic Messages API

---

## 一、问题描述

### 现象

Claude Code 通过 `/api/v1/messages` 和 `/chat/completions` 调用时，发送消息后**无任何"运行中"状态指示**，用户无法判断请求是否被处理。

### 根因

两个流式 handler 均在 `ChatStream()`（阻塞等待上游 HTTP 响应，1-5 秒）**之后**才设置 SSE headers：

| 函数 | 行号（修复前） | 格式 |
|------|---------------|------|
| `handleMessagesStream` | L320 → L327 | Anthropic Messages |
| `handleStreamWithFailover` | L584 → L591 | OpenAI `/chat/completions` |

上游建连期间 HTTP 连接无任何数据返回，客户端无法显示状态。

---

## 二、影响范围

- **用户侧**：所有通过平台使用流式 API 的用户
- **非流式请求**：不受影响

---

## 三、修复方案（已实施：方案 D）

### 核心思路

将 SSE headers + 初始事件**提升到 channel 重试循环之前**发送，循环内部 channel 失败时静默重试，仅在所有 channel 耗尽时发送 SSE error 事件。

### Anthropic 路径 (`handleMessagesStream`)

```
Phase 1: 立即发送 SSE headers + message_start + content_block_start + Flush
Phase 2: 重试循环（失败静默 continue）
Phase 3: c.Stream() 流式传输
Phase 4: message_stop + billing
Phase 5（兜底）: 所有 channel 失败 → SSE error 事件
```

### OpenAI 路径 (`handleStreamWithFailover`)

```
Phase 1: 立即发送 SSE headers + SSE comment 触发 flush
Phase 2: 重试循环（失败静默 continue）
Phase 3: c.Stream() 流式传输 + billing
Phase 4（兜底）: 所有 channel 失败 → SSE error 事件
```

### 改动文件

| 文件 | 变更 | 行数 |
|------|------|------|
| `backend/internal/handler/api_handler.go` | 两个函数执行顺序重排 | ~50 行 |

### 关键取舍

**Channel failover 保留**：重试循环内失败的 channel 静默 `continue`，客户端无感知。仅在全部 channel 耗尽时通过 SSE error 事件通知。

**`Transfer-Encoding: chunked`**：已删除手动设置，Go `net/http` 自动处理。

---

## 四、验证清单

| # | 场景 | 操作 | 预期 |
|---|------|------|------|
| T1 | 正常流式 | POST stream=true | 立即收到 HTTP 200 + SSE headers + message_start |
| T2 | 上游失败 | 配置无效渠道 key | HTTP 200 + 初始事件 → SSE error 事件 |
| T3 | 上游超时 | 渠道指向不可达 URL | 同上 |
| T4 | 第 1 个 channel 失败，第 2 个正常 | 混合配置 | 客户端无感知 failover，正常收到内容 |
| T5 | 3 个 channel 全失败 | 所有渠道不可用 | 收到初始事件 + SSE error |
| T6 | 短对话 | 正常 Claude Code 对话 | 用户能立即看到状态指示 |
| T7 | 长响应 | 复杂任务请求 | 流式持续输出，message_stop 正确 |
| T8 | OpenAI 格式流式 | POST /chat/completions stream=true | 立即收到 HTTP 200 + SSE headers |

---

## 五、决策记录

| 决策项 | 结论 |
|--------|------|
| 是否修复 | ✅ 是 |
| 采用方案 | 方案 D（兼顾即时反馈 + channel failover） |
| 涉及函数 | `handleMessagesStream` + `handleStreamWithFailover` |
| 责任人 | TOM |
| 实施日期 | 2026-05-13 |
