# #008 Phase 2 编译测试指南

> 日期: 2026-05-12
> 关联: [api-key-lifecycle-plan.md](../plans/api-key-lifecycle-plan.md) Phase 2（P1）
> 状态: 待编译验证
> 目标: 验证 Phase 2 五项修复的编译通过和基础功能正确性

---

## 一、需要编译验证的变更

| 文件 | 变更内容 |
|------|---------|
| `backend/internal/middleware/ratelimit.go` | `TokenRateLimit()` 接收 tokenService；从 DB 加载 RPM/TPM；RPM→per-second 转换 + TPM 滑动窗口 |
| `backend/internal/service/billing_service.go` | `PostConsumeQuota()` 包裹在 GORM 事务中；内联 recharge FIFO 扣减逻辑；`getActiveRechargeQuota()` 实际查询 |
| `backend/internal/repository/recharge_record_repo.go` | **新增文件** — `UserRechargeRecordRepository`（GetActiveByUser、GetTotalActiveQuota） |
| `backend/internal/router/router.go` | 三处 `TokenRateLimit()` 调用更新为传入 tokenService；创建 rechargeRecordRepo 传入 BillingService |
| `backend/internal/handler/api_handler.go` | `ChatCompletions()` 和 `Embeddings()` 设置 `c.Set("request_model", req.Model)` |
| `backend/internal/middleware/api_access_log.go` | model 字段从 context 读取（替代 `c.PostForm`） |

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

### 3.1 T-05: TokenRateLimit RPM/TPM

```
验证点 1: Token 无自定义限速 → 默认 RPM 生效
  - 创建 Token（不设 rpm_limit / tpm_limit）
  - 快速连续调用 /v1/chat/completions
  - 预期: 默认限速生效（约 10 RPM）

验证点 2: Token 设置自定义 RPM → 按配置限速
  - 创建 Token（rpm_limit=30）
  - 快速连续调用
  - 预期: 约 30 RPM 生效

验证点 3: RPM 超限 → 429 rate_limit_rpm
  - 用低 RPM Token 超速调用
  - 预期: 429 {"error":{"type":"rate_limit_error","code":"rate_limit_rpm",...}}

验证点 4: TPM 超限 → 429 rate_limit_tpm
  - 创建 Token（tpm_limit=100）
  - 发送大量 token 的请求超过限额
  - 预期: 429 rate_limit_tpm
```

### 3.2 T-06: PostConsumeQuota 事务保护

```
验证点 1: 正常扣费 → 配额变化原子生效
  - 调用前记录 user.free_quota 和 token.used_quota
  - 发起非流式调用
  - 预期: user.free_quota 减少 + token.used_quota 增加 + quota_transactions 有记录
  - 三者变化量一致（原子操作）

验证点 2: 并发扣费正确性（如环境允许）
  - 10 并发各扣 100 quota
  - 预期: 最终余额 = 初始 - 总消耗（无竞态）
```

### 3.3 T-07: 充值配额 FIFO 扣减

```
验证点 1: 有充值记录时配额扣减
  - 确保 user_recharge_records 表有 active 记录
  - 调用 API 消耗配额
  - 预期: recharge records 的 remaining 减少

验证点 2: FIFO 顺序
  - 设置多条 recharge records（不同 expired_at）
  - 调用 API
  - 预期: 先过期的记录先被扣减

验证点 3: 充值配额用完标记 used
  - 将一条 recharge record 扣至 0
  - 预期: 该记录 status 变为 'used'

验证点 4: getActiveRechargeQuota 返回正确总额
  - 查询用户可用配额
  - 预期: 充值配额计入总额（不再返回 0）
```

### 3.4 T-08: APIAccessLog model 字段

```
验证点 1: 非流式调用后日志 model 正确
  - POST /v1/chat/completions {"model":"gpt-3.5-turbo",...}
  - 查看 api_access_logs 表
  - 预期: model 字段 = "gpt-3.5-turbo"（非 "unknown"）

验证点 2: 流式调用后日志 model 正确
  - POST /v1/chat/completions {"model":"gpt-4",...,"stream":true}
  - 预期: model = "gpt-4"

验证点 3: Embeddings 日志 model 正确
  - POST /v1/embeddings {"model":"text-embedding-ada-002",...}
  - 预期: model = "text-embedding-ada-002"
```

---

## 四、已知限制（Phase 3 将修复）

| 问题 | 说明 |
|------|------|
| 无 /v1/completions 端点 | OpenAI text completions 不兼容 |
| 错误格式不完全统一 | APIError 缺 Type/Param 字段 |
| Token 列表缺配额详情 | 只返回用户总配额，非 token 级 used/remain |
| 无压力测试 | Phase 4 范围 |

以上不影响 Phase 2 核心功能验证。

---

## 五、验证结论模板

```
编译结果: ✅ 通过 / ❌ 失败（错误信息: ...）
Go Vet:   ✅ 通过 / ❌ 失败
现有测试: ✅ 全部通过 / ❌ 有失败（用例: ...）
T-05 RPM/TPM: ✅ / ❌ / ⏭️
T-06 事务:    ✅ / ❌ / ⏭️
T-07 充值:    ✅ / ❌ / ⏭️
T-08 日志:    ✅ / ❌ / ⏭️
结论:     可合入 / 需修复后重新验证
```
