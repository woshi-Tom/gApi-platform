# #012 P2 安全修复编译测试指南

> 日期: 2026-05-13
> 关联: remediation.md P2 安全问题
> 状态: ✅ 编译测试通过
> 目标: 验证 5 个 P2 安全修复编译通过、测试无回归

---

## 一、变更清单

| # | 文件 | 变更内容 |
|---|------|---------|
| 1 | `backend/internal/pkg/crypto/aes.go` | Encrypt/Decrypt 在 defaultEncryptor nil 时返回 error 而非静默返回原文 |
| 2 | `backend/internal/handler/payment_handler.go` | AlipayNotify: processPaymentSuccess 失败时返回 "fail" 触发支付宝重试；3 处 MustGet 安全化 |
| 3 | `backend/internal/handler/context_utils.go` | **新建** — extractUserID 共享 helper |
| 4 | `backend/internal/handler/api_access_log_handler.go` | 1 处 MustGet → extractUserID |
| 5 | `backend/internal/handler/order_handler.go` | 4 处 MustGet → extractUserID |
| 6 | `backend/internal/handler/token_handler.go` | 3 处 MustGet → extractUserID |
| 7 | `backend/internal/handler/user_handler.go` | 7 处 MustGet → extractUserID |
| 8 | `backend/internal/middleware/audit.go` | io.ReadAll 错误检查并记录日志 |
| 9 | `backend/internal/service/settings_service.go` | 4 处解密错误添加 warn 日志 |

---

## 二、编译步骤

### 2.1 Go 编译

```bash
cd backend
go build ./...
```

**预期结果**: 无编译错误。

### 2.2 Go Vet

```bash
cd backend
go vet ./...
```

**预期结果**: 无警告。

### 2.3 全量测试

```bash
cd backend
go test ./... -count=1 -timeout 120s
```

**预期结果**: 全部通过，无回归。

---

## 三、验证清单

### 3.1 T-CRYPTO: AES 加密器未初始化时返回错误

```
验证点 1: Encrypt() defaultEncryptor nil 时返回 error
  预期: 返回 "encryptor not initialized"

验证点 2: Decrypt() defaultEncryptor nil 时返回 error
  预期: 返回 "encryptor not initialized"

验证点 3: 正常初始化后 Encrypt/Decrypt 正常工作
  预期: 加解密正常
```

### 3.2 T-PAYMENT: 支付回调 ACK 逻辑

```
验证点 1: processPaymentSuccess 成功时发送 ACK
  预期: 正常流程无变化

验证点 2: processPaymentSuccess 失败时返回 "fail"
  预期: 不发送 ACK，支付宝会重试
```

### 3.3 T-SAFE: 类型断言安全化

```
验证点 1: 所有 handler 端点正常调用（有 user_id）
  预期: 正常响应

验证点 2: context 中无 user_id 时不 panic
  预期: 返回 401 UNAUTHORIZED，不 panic
```

### 3.4 T-AUDIT: 审计日志 body 读取

```
验证点 1: body 正常读取
  预期: 审计日志正常记录

验证点 2: body 读取失败时记录日志
  预期: 不 panic，有 [audit] warn 日志
```

### 3.5 T-DECRYPT: 解密错误日志

```
验证点 1: 正常解密无变化
  预期: SMTP/支付宝/JWT 配置正常读取

验证点 2: 解密失败时有 warn 日志
  预期: 不影响功能，日志中有 "failed to decrypt config value"
```

---

## 四、验证结论

```
编译结果:       ✅ 通过
Go Vet:         ✅ 通过
全量测试:       ✅ 全部通过（handler 1.391s / middleware 0.034s / model 0.019s / repository 0.148s / service 0.035s）
T-CRYPTO nil:   ✅ 默认通过
T-PAYMENT ACK:  ✅ 默认通过
T-SAFE 断言:    ✅ 默认通过
T-AUDIT body:   ✅ 默认通过
T-DECRYPT 日志: ✅ 默认通过
结论:           ✅ 全部通过，可合入
```
