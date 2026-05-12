# #007 编译测试指南

> 日期: 2026-05-12
> 关联: [api-key-lifecycle-plan.md](../plans/api-key-lifecycle-plan.md) Phase 1（P0）
> 状态: 待编译验证
> 目标: 验证 Phase 1 四项修复的编译通过和基础功能正确性

---

## 一、需要编译验证的变更

| 文件 | 变更内容 |
|------|---------|
| `backend/internal/middleware/jwt.go` | 新增 `isIPAllowed()`, `extractModelFromRequest()`；`TokenAuth()` 增加 IP 白名单 + 模型权限校验 |
| `backend/internal/handler/api_handler.go` | `ChatCompletions()` 非流式扣费；`handleStreamWithFailover()` 流式扣费；`Embeddings()` 扣费 |

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

### 3.1 编译后冒烟测试（如环境允许）

启动后端服务后，用 `curl` 或 Postman 验证以下场景：

#### 场景 A: IP 白名单

```bash
# 1. 创建 Token 并设置 allowed_ips=["127.0.0.1"]
# 2. 用该 Token 从本机调用：
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Authorization: Bearer sk-ap-xxxxx" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}'
# 预期: 200 或上游错误（IP 在白名单内）
```

#### 场景 B: 模型白名单

```bash
# 1. 创建 Token 并设置 allowed_models=["gpt-4"]
# 2. 用 gpt-3.5-turbo 调用：
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Authorization: Bearer sk-ap-xxxxx" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"hi"}]}'
# 预期: 403 {"error":{"type":"invalid_request_error","code":"model_not_allowed",...}}
```

#### 场景 C: 模型黑名单

```bash
# 1. 创建 Token 并设置 denied_models=["gpt-4"]
# 2. 用 gpt-4 调用：
# 预期: 403 model_not_allowed
# 3. 用 gpt-3.5-turbo 调用：
# 预期: 正常通过
```

#### 场景 D: 非流式配额扣减

```bash
# 1. 记录调用前 user.free_quota 和 token.used_quota
# 2. 发起非流式调用
# 3. 查看调用后配额变化
# 预期: user.free_quota 减少，token.used_quota 增加
```

#### 场景 E: 流式配额扣减

```bash
# 1. 记录调用前配额
# 2. 发起流式调用（stream=true）
# 3. 等待流式结束
# 4. 查看配额变化
# 预期: 配额被扣减（精确值或估算值）
```

### 3.2 日志验证

查看后端日志，确认以下输出：

- TokenAuth 中间件不应有 panic 或异常错误
- API 调用后应有 usage 日志输出（包含 model、token 数）
- 流式调用结束后应有 PostConsumeQuota 相关日志

---

## 四、已知限制（Phase 1 范围外，Phase 2 将修复）

| 问题 | 说明 |
|------|------|
| TokenRateLimit 硬编码 | 当前所有 Token 限速 10 RPS，忽略用户配置 |
| PostConsumeQuota 无事务 | 并发扣费可能存在竞态 |
| consumeRechargeQuota 空实现 | 充值配额不参与扣减 |
| APIAccessLog model 字段 | 日志中 model 恒为 "unknown" |
| 错误格式不完全统一 | 部分错误用 APIResponse，部分用 APIErrorResponse |

以上问题不影响 Phase 1 核心功能验证，属于 Phase 2（P1）修复范围。

---

## 五、验证结论模板

完成验证后填写：

```
编译结果: ✅ 通过 / ❌ 失败（错误信息: ...）
Go Vet:   ✅ 通过 / ❌ 失败
现有测试: ✅ 全部通过 / ❌ 有失败（用例: ...）
冒烟测试: ✅ / ❌ / ⏭️ 跳过（无测试环境）
结论:     可合入 / 需修复后重新验证
```
