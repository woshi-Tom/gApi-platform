# gAPI Platform - 文档索引

> 版本: v4.0
> 日期: 2026-05-12
> 更新说明: 重构目录结构，去掉编号前缀，合并 features/ 统一管理

---

## 目录结构

```
docs/
├── development/        # ⚠️ 开发必读 / 审查报告
├── architecture/       # 架构设计
├── api/                # API 接口
├── features/           # 功能文档（含实现状态）
├── deployment/         # 部署运维
└── issues/             # 问题追踪
```

---

## 开发指南

| 文档 | 说明 |
|------|------|
| [development-notes.md](development/development-notes.md) | ⚠️ **开发前必读** - 环境变量、接口清单、检查项 |
| [git-workflow.md](development/git-workflow.md) | Git 协作规范 + Release 发布流程 |
| [fix-plan-2026.md](development/fix-plan-2026.md) | 2026 修复计划与技术债追踪 |
| [release-review-v1.2.0.md](development/release-review-v1.2.0.md) | v1.2.0 发布审查报告与整改措施 |

---

## 架构设计

| 文档 | 说明 |
|------|------|
| [system-design.md](architecture/system-design.md) | 系统整体架构设计 |
| [database-design-v2.md](architecture/database-design-v2.md) | 数据库完整 DDL |
| [project-structure.md](architecture/project-structure.md) | 项目目录结构 |

---

## API 接口

| 文档 | 说明 |
|------|------|
| [interface-design-south-north.md](api/interface-design-south-north.md) | 北向/南向/管理后台接口概览 |
| [api-design.md](api/api-design.md) | API 详细设计 |

---

## 功能文档

> 所有功能（已实现/待完善/未实现）统一在 [features/README.md](features/README.md) 中管理。

| 模块 | 说明 | 状态 |
|------|------|------|
| 支付 | 支付宝支付、商品套餐 | ✅ |
| 用户 | 注册验证、滑块验证码、邮箱/SMTP | ✅ |
| 渠道 | 渠道管理 CRUD、健康检测 | ✅ + ⚠️ |
| 模型管理 | 模型分组、定价、用户组权限 | ✅ |
| 兑换码 | 生成、兑换、禁用、批次管理 | ✅ |
| 审计日志 | 优化审计日志存储和查询 | ⚠️ |

---

## 部署运维

| 文档 | 说明 |
|------|------|
| [deployment.md](deployment/deployment.md) | Docker 部署文档 |
| [security-deployment.md](deployment/security-deployment.md) | 安全与部署指南 |
| [business-detail.md](deployment/business-detail.md) | 业务详细设计 |

---

## 问题追踪

| 编号 | 问题 | 状态 |
|------|------|------|
| 001 | [支付成功后 VIP 激活失败](issues/001-payment-vip-activation-failure.md) | ✅ 已修复 |
| 002 | [管理后台 Dashboard 502 错误](issues/002-admin-dashboard-502.md) | ✅ 已修复 |
| 003 | [操作日志显示无数据/数据膨胀](issues/003-operation-logs-empty.md) | ✅ 已修复 |

---

## 相关链接

- [OneAPI 参考](https://github.com/songquanpeng/one-api)
- [OpenAI API](https://platform.openai.com/docs)
- [Element Plus](https://element-plus.org/)
