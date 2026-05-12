# 修复 Chat Completions content 字段兼容性

> 类型: 功能完善 / API 兼容性
> 提出: 2026-05-13
> 状态: ⬜ 待评审
> 关联: cc switch 协议转换报错 `cannot unmarshal array into Go struct field ChatCompletionsRequest.messages of type string`

---

## 一、问题描述

### 现象

通过 cc switch 代理调用平台 API 时，返回 400 错误：

```
json: cannot unmarshal array into Go struct field ChatCompletionsRequest.messages of type string
```

### 根因

平台 `ChatCompletionsRequest.Messages` 定义：

```go
Messages []map[string]string  // 只接受 content 为字符串
```

OpenAI 官方规范中 `messages[].content` 支持两种格式：

| 格式 | 示例 | 平台支持 |
|------|------|---------|
| **字符串** | `{"role":"user","content":"你好"}` | ✅ |
| **内容块数组** | `{"role":"user","content":[{"type":"text","text":"你好"}]}` | ❌ |

cc switch 做 Anthropic → OpenAI 协议转换时，使用内容块数组格式（这是完全合规的）。平台拒绝了解析。

### 影响范围

- **cc switch 用户**：无法通过 cc switch 使用平台
- **多模态客户端**：将来 OpenAI 多模态请求（图片/音频）也使用内容块数组格式，同样会被拒绝
- **API 合规性**：与 OpenAI 官方规范不完全一致

---

## 二、修复方案

### 方案 A（推荐）— Messages 改为 interface{} + content 标准化

**改动点：**

1. **`model/response.go`** — `Messages` 类型变更：

   ```go
   // 修改前
   Messages []map[string]string
   
   // 修改后
   Messages []map[string]interface{}
   ```

2. **`handler/api_handler.go`** — 新增 `normalizeMessages()`，在 `ChatCompletions` 入口处调用：

   ```go
   func normalizeMessages(msgs []map[string]interface{}) []map[string]string {
       result := make([]map[string]string, 0, len(msgs))
       for _, m := range msgs {
           normalized := make(map[string]string)
           for k, v := range m {
               switch val := v.(type) {
               case string:
                   normalized[k] = val
               case []interface{}:
                   // content 数组 → 提取 text 拼接
                   text := ""
                   for _, part := range val {
                       if block, ok := part.(map[string]interface{}); ok {
                           if block["type"] == "text" {
                               text += block["text"].(string)
                           }
                       }
                   }
                   normalized[k] = text
               default:
                   normalized[k] = fmt.Sprintf("%v", v)
               }
           }
           result = append(result, normalized)
       }
       return result
   }
   ```

3. **`handler/api_handler.go`** — `estimateChatTokens()` 签名改为接收 `[]map[string]interface{}`，或在调用前先 normalize。

**涉及文件：** 2 个

| 文件 | 变更类型 | 行数 |
|------|---------|------|
| `model/response.go` | 修改 1 行（类型定义） | +0/-0 |
| `handler/api_handler.go` | 新增 normalizeMessages + 调用 | +35 行 |

**优点：**
- 完全遵循 OpenAI 规范
- 为将来多模态请求铺路
- 改动量小，2 个文件
- 无性能影响（仅 JSON 解析后多一次遍历）

**缺点：**
- 需要重新编译部署

---

### 方案 B — 仅修改 cc switch 配置

在 cc switch 侧配置 content 展开为纯字符串。

**优点：** 不改平台代码

**缺点：**
- 每个使用 cc switch 的用户都需要单独配置
- 其他客户端（如直接调用 API 的脚本）仍可能遇到同样问题
- 平台 API 仍然不符合 OpenAI 规范
- 多模态请求未来仍会失败

---

### 方案 C — 新增 /v1/messages Anthropic 兼容端点

新增一个路由 `/v1/messages`，接收 Anthropic Messages API 格式，内部转为 OpenAI 格式处理。

**优点：** Claude Code 等 Anthropic 原生客户端可直接对接

**缺点：**
- 改动量大（新路由 + 新 handler + 响应格式转换）
- 需要同时处理流式响应格式
- 需要维护两套 API 格式的转换逻辑

---

## 三、各角色评估意见

| 角色 | 观点 | 倾向方案 |
|------|------|---------|
| 📋 **产品经理** | API 兼容性是平台的核心价值。OpenAI 生态是最大的开发者生态，不应因为规范理解偏差限制用户。cc switch 是合理的中转工具。 | 方案 A |
| 📊 **项目经理** | 方案 A 改动 2 个文件 +35 行，测试验证简单。方案 C 涉及新端点，测试周期长。建议分阶段：先 A（修兼容性），后 C（加 Anthropic 端点）。 | 方案 A（先）→ 方案 C（后） |
| 💻 **开发** | 方案 A 实现简单，无技术风险。content 数组提取逻辑清晰。需要关注流式（streaming）场景下 content 格式是否也会出现数组。 | 方案 A |
| 🧪 **QA/测试** | 验证点：字符串 content ✅、数组 content ✅、空 content ❓、混合格式 ✅。回归测试覆盖现有的 curl 正常调用不受影响。 | 方案 A |
| 🎨 **UI/前端** | 不涉及前端变更，无影响。 | 方案 A |
| 🔒 **安全** | content 数组可能包含非 text 类型（如 image_url），标准化时应仅提取 text 部分，忽略其他类型。无注入风险。 | 方案 A（注意仅提取 text） |

---

## 四、建议执行方案

**第一阶段（本次）：方案 A**
- 2 个文件，约 +35 行代码
- 工时：开发 0.5h + 测试 0.5h
- 验证清单：见第五章节

**第二阶段（后续 Sprint）：方案 C**
- 新增 `/v1/messages` 端点
- 使 Claude Code 可直接对接，无需 cc switch 中转
- 排期到后续 Sprint

---

## 五、验证清单

### 编译验证

```bash
cd backend
go build ./...
go vet ./...
```

**预期**: 无编译错误，无 vet 警告。

### 功能验证

| 编号 | 场景 | 请求 | 预期 |
|------|------|------|------|
| T1 | 字符串 content（回归） | `{"messages":[{"role":"user","content":"hi"}]}` | 200 |
| T2 | 数组 content（本次修复） | `{"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}` | 200 |
| T3 | 混合 messages | `[{"role":"user","content":"hi"},{"role":"assistant","content":[{"type":"text","text":"hello"}]}]` | 200 |
| T4 | 空 content | `{"messages":[{"role":"user","content":""}]}` | 200 或合理错误 |
| T5 | 多模态格式（忽略非 text） | `{"messages":[{"role":"user","content":[{"type":"text","text":"hi"},{"type":"image_url","url":"..."}]}]}` | 200，仅提取 text |
| T6 | cc switch 全链路 | 通过 cc switch 发送 "你好" | 200，正常回复 |
| T7 | 流式（stream=true） | 数组 content + stream=true | 200，SSE 正常 |

---

## 六、决策

| 决策项 | 结论 |
|--------|------|
| 是否修复 | ⬜ 待定 |
| 采用方案 | ⬜ 方案 A / 方案 B / 方案 C |
| 是否分阶段 | ⬜ 仅 A / A→C 分阶段 |
| 责任人 | ⬜ |
| 排期 | ⬜ 当前 Sprint / 后续 Sprint |
