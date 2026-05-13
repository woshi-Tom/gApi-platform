# #013 billing 事务修复编译测试指南

> 日期: 2026-05-13
> 关联: #011 P3 — billing_service.go:151
> 状态: ⬜ 待验证
> 目标: 验证 billing 事务修复编译通过、测试无回归

---

## 一、变更清单

| # | 文件 | 变更内容 |
|---|------|---------|
| 1 | `backend/internal/repository/recharge_record_repo.go` | 新增 `WithTx(tx)` 方法，返回使用事务连接的 repo 副本 |
| 2 | `backend/internal/service/billing_service.go` | PostConsumeQuota 中 `GetActiveByUser` 改为 `WithTx(tx).GetActiveByUser` |

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

### 3.1 T-TX: 事务内读取一致性

```
验证点 1: PostConsumeQuota 正常扣减
  预期: free_quota + recharge_quota 扣减正常

验证点 2: WithTx 不影响现有 GetActiveByUser 调用
  预期: 非事务场景（测试中直接调用）行为不变

验证点 3: FIFO 顺序正确
  预期: 按 expired_at ASC 顺序扣减
```

---

## 四、验证结论

```
编译结果:       ⬜ 待验证
Go Vet:         ⬜ 待验证
全量测试:       ⬜ 待验证
T-TX 事务读取:  ⬜ 待验证
结论:           ⬜ 待远程智能体验证
```
