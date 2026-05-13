# API Key 生命周期 Phase 2/3/4 测试补充方案

> 类型: 测试补充
> 提出: 2026-05-13
> 状态: ⬜ 待评审
> 关联: [api-key-lifecycle-plan.md](api-key-lifecycle-plan.md) Phase 2/3/4 "待编译验证后补充"
> 背景: Phase 1-4 的功能代码已全部完成并编译通过，但 Phase 2/3/4 的测试尚未编写

---

## 一、现状分析

### 1.1 已有测试覆盖（2708 行）

| 文件 | 覆盖模块 | 行数 |
|------|---------|------|
| `integration_test.go` | ChatCompletions 集成 | 491 |
| `redemption_handler_test.go` | 兑换码 CRUD | 292 |
| `settings_handler_test.go` | Settings JSON 绑定 | 246 |
| `audit_test.go` | 审计日志中间件 | 243 |
| `middleware_test.go` | JWT/TokenAuth | 192 |
| `order_test.go` | 订单模型 | 175 |
| `repository_test.go` | 数据仓储层 | 487 |
| `auth_service_test.go` | 认证服务 | 245 |
| `settings_service_test.go` | Settings 服务 | 238 |
| `redemption_cache_service_test.go` | 兑换码缓存 | 99 |

### 1.2 缺失测试（按 api-key-lifecycle-plan.md 标注）

| Phase | 功能 | 代码位置 | 测试状态 |
|-------|------|---------|---------|
| P2 | TokenRateLimit 自定义 RPM/TPM | `middleware/ratelimit.go` | ❌ 无 |
| P2 | PostConsumeQuota 事务保护 | `service/billing_service.go` | ❌ 无 |
| P2 | consumeRechargeQuota FIFO | `service/billing_service.go` + `repository/recharge_record_repo.go` | ❌ 无 |
| P3 | /v1/completions 端点 | `handler/api_handler.go` Completions() | ❌ 无 |
| P3 | APIError 统一格式 | 多文件 | ⚠️ 集成测试仅覆盖部分 |
| P4 | 压测回归 | `test_api.py` | ⚠️ 脚本存在但无自动化断言 |

---

## 二、团队评审

### 📋 产品经理

> 测试是产品质量的保障。Phase 2 涉及计费和限速——这是用户最敏感的部分，配额扣错会导致客诉。Phase 3 的 /v1/completions 是 OpenAI 兼容性的核心卖点。建议优先保障 Phase 2 计费测试。

**关注点**: 计费正确性 > 限速正确性 > 接口兼容性

### 📊 项目经理

> 当前测试缺口是 api-key-lifecycle-plan.md 的遗留项，计划本身已标记"全部完成"但测试未补。建议控制范围：本次只补单元测试，压测回归（Phase 4）推迟到后续 Sprint。

**建议范围**:
- ✅ Phase 2: 单元测试（计费 + 限速）
- ✅ Phase 3: 单元测试（completions + APIError）
- ⏭️ Phase 4: 压测回归推迟

### 💻 后端开发

> 补充测试的技术要点：
>
> **Phase 2 计费测试**:
> - `PostConsumeQuota` 需要 mock DB 事务，验证原子性
> - `consumeRechargeQuota` FIFO 需要准备多条 recharge records 测试数据
> - `TokenRateLimit` 需要 mock Redis + 时间，验证 RPM/TPM 限速逻辑
>
> **Phase 3 功能测试**:
> - `/v1/completions` 可以用现有 integration_test.go 框架扩展
> - APIError 格式可以用表驱动测试覆盖所有错误路径
>
> **测试难度评估**:
> | 模块 | 难度 | 原因 |
> |------|------|------|
> | TokenRateLimit | 中 | 需 mock Redis + 时间 |
> | PostConsumeQuota | 高 | 事务 mock 复杂，需验证并发 |
> | Recharge FIFO | 中 | 需准备测试数据 |
> | Completions | 低 | 可复用集成测试框架 |
> | APIError | 低 | 表驱动即可 |

### 🧪 QA/测试

> 建议按风险优先级排序：
>
> 1. **P0**: PostConsumeQuota 事务 — 扣错钱是最高风险
> 2. **P0**: Recharge FIFO — 充值配额扣减顺序错误会导致用户余额异常
> 3. **P1**: TokenRateLimit — 限速失效导致资源耗尽
> 4. **P1**: /v1/completions — 核心功能回归
> 5. **P2**: APIError 格式 — 不影响功能，只影响客户端解析
>
> 验收标准：每个被测函数的核心分支覆盖 ≥ 80%

### 🔒 安全

> 关注点：
> - 限速绕过测试：验证 RPM/TPM 限制无法被绕过
> - 配额负值测试：验证配额不会扣到负数
> - 并发竞态测试：验证并发扣费不会出现超额消费

### 🎨 前端

> 无直接影响。/v1/completions 的响应格式变化可能影响前端调用，但当前前端不直调此端点。

---

## 三、建议方案

### 3.1 范围

本次补充 **Phase 2 + Phase 3 单元测试**，Phase 4 压测回归推迟。

### 3.2 测试清单

#### Phase 2: 计费与限速（~6 个测试函数）

| 编号 | 测试函数 | 验证点 |
|------|---------|--------|
| T-P2-01 | `TestTokenRateLimit_RPM` | 自定义 RPM 生效，超限返回 429 |
| T-P2-02 | `TestTokenRateLimit_TPM` | 自定义 TPM 生效，超限返回 429 |
| T-P2-03 | `TestPostConsumeQuota_Atomic` | 事务中 user quota + token quota + transaction 记录原子更新 |
| T-P2-04 | `TestPostConsumeQuota_Insufficient` | 配额不足时不扣减 |
| T-P2-05 | `TestConsumeRechargeQuota_FIFO` | 多条 recharge 按过期时间顺序扣减 |
| T-P2-06 | `TestConsumeRechargeQuota_Exhaust` | 单条 recharge 扣至 0 后标记 used |

#### Phase 3: 功能补全（~4 个测试函数）

| 编号 | 测试函数 | 验证点 |
|------|---------|--------|
| T-P3-01 | `TestCompletions_StringPrompt` | 字符串 prompt → 200 text_completion |
| T-P3-02 | `TestCompletions_ArrayPrompt` | 数组 prompt → 200 |
| T-P3-03 | `TestCompletions_MissingModel` | 缺少 model → 400 missing_model |
| T-P3-04 | `TestAPIError_Format` | 所有错误路径返回 type/code/message 三字段 |

### 3.3 涉及文件

| 文件 | 操作 |
|------|------|
| `backend/internal/middleware/ratelimit_test.go` | **新建** — T-P2-01, T-P2-02 |
| `backend/internal/service/billing_service_test.go` | **新建** — T-P2-03 ~ T-P2-06 |
| `backend/internal/handler/api_handler_test.go` | **新建** — T-P3-01 ~ T-P3-03 |
| `backend/internal/handler/api_handler_test.go` | T-P3-04（表驱动） |

### 3.4 预估工时

| 阶段 | 工时 |
|------|------|
| Phase 2 限速测试 | 1.5h |
| Phase 2 计费测试 | 2.5h |
| Phase 3 功能测试 | 1.5h |
| 调试 + 验证 | 0.5h |
| **合计** | **6h** |

---

## 四、决策

| 决策项 | 结论 |
|--------|------|
| 是否补充测试 | ⬜ 待定 |
| 范围 | ⬜ Phase 2+3 / 仅 Phase 2 / 全部 |
| 计费测试优先级 | ⬜ P0 先行 / 按顺序 |
| Phase 4 压测回归 | ⬜ 本次包含 / 推迟 |
| 责任人 | ⬜ |
| 排期 | ⬜ 当前 Sprint / 后续 Sprint |
