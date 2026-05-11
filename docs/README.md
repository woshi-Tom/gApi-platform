# gAPI Platform - Documentation Index

> 版本: v3.2
> 日期: 2026-05-12
> 更新说明: 补充planned/目录、fix-plan、CI/CD release流程文档

---

## 📂 文档目录结构

```
docs/
├── 01-architecture/          # 架构设计
├── 02-api/                   # API设计
├── 03-features/              # 已实现功能
├── 04-deployment/           # 部署运维
├── 05-development/           # 开发指南
├── design/                   # 功能设计方案（待实现）
├── planned/                  # 开发计划
└── issues/                   # 问题追踪
```

---

## 1️⃣ 架构设计 (`01-architecture/`)

| 文档 | 说明 | 状态 |
|------|------|------|
| [system-design.md](./01-architecture/system-design.md) | 系统整体架构设计 | ✅ |
| [database-design-v2.md](./01-architecture/database-design-v2.md) | 数据库完整DDL | ✅ |
| [project-structure.md](./01-architecture/project-structure.md) | 项目目录结构 | ✅ |

---

## 2️⃣ API设计 (`02-api/`)

| 文档 | 说明 | 状态 |
|------|------|------|
| [interface-design-south-north.md](./02-api/interface-design-south-north.md) | 北向/南向/管理后台接口 | ✅ |
| [api-design.md](./02-api/api-design.md) | API详细设计 | ⚠️ 部分 |

---

## 3️⃣ 已实现功能 (`03-features/`)

### 支付模块 ✅

| 文档 | 说明 | 代码位置 |
|------|------|----------|
| [payment-module-design.md](./03-features/payment-module-design.md) | 支付模块设计 | ✅ |
| [alipay-payment-design.md](./03-features/alipay-payment-design.md) | 支付宝支付集成 | ✅ `alipay_service.go` |
| [payment-module-fix-log.md](./03-features/payment-module-fix-log.md) | 修复日志 | ✅ |
| [payment-module-issues.md](./03-features/payment-module-issues.md) | 问题记录 | ✅ |

> ⚠️ **注意**: 微信支付仅配置存在，代码中 `wechat_enabled: false`，未实际集成

### 用户模块
| 文档 | 说明 | 状态 |
|------|------|------|
| [email-verification-design.md](./03-features/email-verification-design.md) | 邮箱验证 | ✅ |
| [smtp-config-design.md](./03-features/smtp-config-design.md) | SMTP配置 | ✅ |
| [user-api-monitor-design.md](./03-features/user-api-monitor-design.md) | API监控 | ✅ |

### 商品/套餐
| 文档 | 说明 | 状态 |
|------|------|------|
| [business-package-spec.md](./03-features/business-package-spec.md) | 商品/套餐规格 | ✅ |

---

## 4️⃣ 部署运维 (`04-deployment/`)

| 文档 | 说明 | 状态 |
|------|------|------|
| [deployment.md](./04-deployment/deployment.md) | Docker部署文档 | ✅ |
| [security-deployment.md](./04-deployment/security-deployment.md) | 安全与部署指南 | ✅ |
| [business-detail.md](./04-deployment/business-detail.md) | 业务详细设计 | ⚠️ 部分 |

---

## 5️⃣ 开发指南 (`05-development/`)

| 文档 | 说明 | 状态 |
|------|------|------|
| [development-notes.md](./05-development/development-notes.md) | ⚠️ **开发前必读** | ✅ |
| [git-workflow.md](./05-development/git-workflow.md) | ⚠️ **Git 协作规范 + Release 发布流程** | ✅ |
| [fix-plan-2026.md](./05-development/fix-plan-2026.md) | 2026修复计划与技术债追踪 | ✅ |

---

## 🎯 功能设计方案 (`design/`) - 待实现

| 编号 | 功能模块 | 文档路径 | 后端状态 | 前端状态 |
|------|----------|----------|----------|----------|
| 01 | 渠道管理 | [channel-management-design.md](./design/channel-management-design.md) | ✅ | ⚠️ 列表页面 |
| 02 | 渠道健康检测 | [channel-health-check-design.md](./design/channel-health-check-design.md) | ✅ | ❌ |
| 03 | 注册配置 | [signup-config-design.md](./design/signup-config-design.md) | ⚠️ 配置存储 | ⚠️ 部分 |
| 04 | 兑换码 | [redemption-code-design.md](./design/redemption-code-design.md) | ❌ | ❌ |
| 05 | 审计日志优化 | [audit-log-optimization-design.md](./design/audit-log-optimization-design.md) | ✅ 已实现 | ✅ 已实现 |
| 06 | 模型分组/定价/权限 | [model-group-pricing-permission-design.md](./design/model-group-pricing-permission-design.md) | ⚠️ 设计中 | ❌ |

> 详细状态见 [`design/README.md`](./design/README.md)

### 已实现详情

| 功能 | 后端 | 前端 |
|------|------|------|
| 渠道管理 | ✅ CRUD接口 | ⚠️ 列表页面 |
| 健康检测 | ✅ 定时任务 | ⚠️ 状态显示 |

### 待实现

| 功能 | 说明 |
|------|------|
| 兑换码 | 兑换码生成和使用 |
| 注册配置 | 注册开关、验证方式、奖励 |
| 渠道表单 | 新增/编辑渠道 |

---

## 🐛 问题追踪 (`issues/`)

| 编号 | 问题 | 状态 |
|------|------|------|
| 001 | [支付成功后VIP激活失败](./issues/001-payment-vip-activation-failure.md) | ✅ 已修复 |
| 002 | [管理后台Dashboard 502错误](./issues/002-admin-dashboard-502.md) | ✅ 已修复 |
| 003 | [操作日志显示无数据/数据膨胀](./issues/003-operation-logs-empty.md) | ✅ 已修复 |

---

## 📋 开发计划 (`planned/`)

| 文档 | 说明 | 状态 |
|------|------|------|
| [backend-health-check-api-plan.md](./planned/backend-health-check-api-plan.md) | 渠道手动健康检测API | ✅ 已完成 |

---

## 📊 项目实际实现状态

### ✅ 已完成模块

| 模块 | 功能 | 关键文件 |
|------|------|----------|
| 用户 | 注册/登录/Token管理 | `auth_service.go`, `token_service.go` |
| 用户 | 邮箱验证 | `email_verification_service.go` |
| 管理 | 用户/商品/订单管理 | `admin_handler.go`, `product_handler.go` |
| 管理 | 设置管理 | `settings_handler.go` |
| 支付 | 支付宝支付 + VIP激活 | `alipay_service.go`, `billing_service.go` |
| 渠道 | 渠道CRUD + 健康检测 | `channel_handler.go`, `health_check.go` |
| 监控 | API日志 / 操作日志 / 登录日志 | `api_access_log_handler.go` |
| CI/CD | GitHub Actions 自动构建 + Release | `.github/workflows/build.yml` |
| CI/CD | 跨平台发布 (Linux tgz + Windows zip) | `deploy/release/` |

### ⚠️ 待完善

| 功能 | 说明 |
|------|------|
| 渠道管理前端 | 新增/编辑渠道表单 |
| 健康状态UI | 渠道列表状态列 |
| 注册配置流程 | auth_service执行signup_config |

### ❌ 未实现

| 功能 | 说明 |
|------|------|
| 兑换码 | 完整功能 |
| 微信支付 | 未集成 |

---

## 🔗 相关链接

- [OneAPI 参考](https://github.com/songquanpeng/one-api)
- [OpenAI API](https://platform.openai.com/docs)
- [Element Plus](https://element-plus.org/)

---

## 📝 更新记录

| 日期 | 版本 | 更新内容 |
|------|------|----------|
| 2026-05-12 | v3.2 | 补充planned/目录、fix-plan、CI/CD release、模型定价设计文档 |
| 2026-04-01 | v3.1 | 更新文档状态匹配实际实现 |
| 2026-04-01 | v3.0 | 重新组织文档结构 |
| 2026-04-01 | v2.0 | 删除冗余文档 |
| 2026-04-01 | v1.0 | 初始版本 |
