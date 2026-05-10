# Issue #001: 支付成功后VIP激活失败

## 问题描述

用户完成支付宝支付后，订单状态卡在"待支付"，VIP未成功激活。

## 根本原因

### 数据库列名不一致

数据库表 `users` 中的列名使用下划线格式：
```sql
v_ip_quota
v_ip_expired_at
v_ip_package_id
```

但代码中 `payment_handler.go` 使用了错误的列名：
```go
// 错误写法
Update("vip_quota", newQuota)
Update("vip_expired_at", ...)
Update("vip_package_id", ...)
```

GORM 在使用 `tx.Model(&model.User{}).Update("column_name", value)` 时使用字面值，不会应用模型中定义的 `column:` 标签。

## 影响

1. VIP用户支付成功后，VIP等级/配额/到期时间不会更新
2. 订单状态停留在 `pending`，不会变为 `completed`
3. 前端显示"支付成功"但订单列表显示"待支付"

## 修复方案

将所有错误的列名改为正确的下划线格式：

| 错误写法 | 正确写法 |
|---------|---------|
| `vip_quota` | `v_ip_quota` |
| `vip_expired_at` | `v_ip_expired_at` |
| `vip_package_id` | `v_ip_package_id` |

### 修复文件

- `/backend/internal/handler/payment_handler.go`

## 预防措施

### 1. 代码规范
- 统一使用 GORM 模型定义的字段访问方式
- 优先使用 `tx.Save(&user)` 而不是 `tx.Update("column", value)`
- 或者使用 `tx.Model(&model.User{}).Update("v_ip_quota", value)` 明确指定列名

### 2. 审查清单
提交代码前检查：
- [ ] 数据库列名引用是否与 schema.sql 一致
- [ ] 是否使用了 GORM 模型的列名映射标签

### 3. 测试验证
支付流程测试清单：
- [ ] 创建VIP订单
- [ ] 完成支付
- [ ] 验证订单状态变为 `completed`
- [ ] 验证用户 level 更新为正确的 VIP 等级
- [ ] 验证用户 v_ip_quota 更新
- [ ] 验证用户 v_ip_expired_at 更新

## 相关文件

| 文件 | 说明 |
|------|------|
| `backend/internal/model/user.go` | 用户模型定义 |
| `backend/internal/handler/payment_handler.go` | 支付处理逻辑 |
| `backend/scripts/schema.sql` | 数据库表结构定义 |

## 发现日期

2026-04-01

## 状态

✅ 已修复
