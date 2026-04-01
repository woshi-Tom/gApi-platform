# 功能设计方案

> 待实现功能的设计文档

## 文档列表

| 编号 | 功能模块 | 文档 | 后端状态 | 前端状态 |
|------|----------|------|----------|----------|
| 01 | 渠道管理 | [channel-management-design.md](./channel-management-design.md) | ✅ | ⚠️ 列表页面 |
| 02 | 渠道健康检测 | [channel-health-check-design.md](./channel-health-check-design.md) | ✅ | ❌ |
| 03 | 注册配置 | [signup-config-design.md](./signup-config-design.md) | ⚠️ 配置存储 | ⚠️ 部分 |
| 04 | 兑换码 | [redemption-code-design.md](./redemption-code-design.md) | ❌ | ❌ |
| 05 | 审计日志优化 | [audit-log-optimization-design.md](./audit-log-optimization-design.md) | ✅ 已实现 | ✅ 已实现 |

---

## 状态说明

| 状态 | 含义 |
|------|------|
| ✅ | 已完成实现 |
| ⚠️ | 部分实现，待完善 |
| ❌ | 未开始实现 |

---

## 实现优先级

### 🔴 高优先级

1. **渠道管理前端** - 新增/编辑渠道表单
2. **渠道健康状态显示** - 状态列、手动检测按钮

### 🟡 中优先级

3. **注册配置流程** - 让auth_service读取并执行signup_config

### 🟢 低优先级

4. **兑换码** - 完整实现

---

## 更新记录

| 日期 | 版本 | 更新内容 |
|------|------|----------|
| 2026-04-01 | v1.0 | 初始版本，标注实现状态 |
