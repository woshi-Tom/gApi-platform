# gAPI Platform - 文档索引

> 版本: v5.0
> 日期: 2026-05-12
> 更新说明: 按生命周期分层（规范/设计/评审/计划/待办/归档）

---

## 目录结构

```
docs/
├── development/        # ⚠️ 开发规范（低频变更，按需修订）
├── architecture/       # 架构设计
├── api/                # API 接口
├── features/           # 功能设计文档（含实现状态 ✅⚠️❌）
├── reviews/            # 🔄 评审专区（审查→整改→闭环，按版本组织）
├── plans/              # 📋 计划追踪（当前在做什么）
├── backlog/            # 📥 待办池（已知但未排期）
└── issues/             # ✅ 已归档问题
```

### 生命周期说明

| 目录 | 变更频率 | 用途 | 智能体怎么用 |
|------|---------|------|-------------|
| `development/` | 低频 | 开发规范（环境变量、git 流程） | 按需修订，非静态 |
| `architecture/` `/api/` | 按需 | 设计参考 | 需要时查阅 |
| `features/` | 随功能迭代 | 功能设计 + 实现状态 | 查功能是否已完成 |
| `reviews/` | 每次发布一次 | 审查报告 + 整改跟踪 | 查看当前版本审查结论 |
| `plans/` | 持续更新 | **当前活跃计划** | ⭐ 进门第一站：读 `current-sprint.md` |
| `backlog/` | 持续更新 | 已知问题 + 待决策 | 看还有哪些事没排期 |
| `issues/` | 偶尔追加 | 已关闭问题归档 | 排查历史 |

---

## 开发指南

> 开发规范文档，按需修订（如 git-workflow 会随流程优化更新）。

| 文档 | 说明 |
|------|------|
| [development-notes.md](development/development-notes.md) | ⚠️ **开发前必读** - 环境变量、接口清单、检查项 |
| [git-workflow.md](development/git-workflow.md) | Git 协作规范 + Release 发布流程 |

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
| 审计日志 | 优化审计日志存储和查询 | ✅ |

---

## 评审专区

> 当前活跃评审：**v1.2.0** 🟡 整改中

| 版本 | 结论 | 状态 | 报告 | 整改 |
|------|------|------|------|------|
| v1.2.0 | ❌ 不通过（5 P0） | 🟡 整改中 | [review-report.md](reviews/v1.2.0/review-report.md) | [remediation.md](reviews/v1.2.0/remediation.md) |

---

## 计划追踪

> 当前活跃 sprint 和路线图。

| 文档 | 说明 | 状态 |
|------|------|------|
| [current-sprint.md](plans/current-sprint.md) | v1.2.0 迭代工作项（5 个 P0） | ✅ 已完成 |
| [api-key-lifecycle-plan.md](plans/api-key-lifecycle-plan.md) | API Key 全生命周期与调用链路完善（v1.3.0） | 🟡 进行中（Phase 1+2 ✅） |
| [fix-plan-2026.md](plans/fix-plan-2026.md) | 2026 修复计划与技术债追踪 | ⬜ 待审核 |
| [roadmap.md](plans/roadmap.md) | 产品路线图与里程碑 | ⬜ 待完善 |

---

## 待办池

> 已知但未排期的问题和待决策事项。

| 文档 | 说明 |
|------|------|
| [known-issues.md](backlog/known-issues.md) | 已确认未修复的问题（10+ 项） |
| [pending-decisions.md](backlog/pending-decisions.md) | 待决策事项（含已决策记录） |

---

## 部署运维

| 文档 | 说明 |
|------|------|
| [deployment.md](deployment/deployment.md) | Docker 部署文档 |
| [security-deployment.md](deployment/security-deployment.md) | 安全与部署指南 |
| [business-detail.md](deployment/business-detail.md) | 业务详细设计 |
| [upgrade-v1.2.0.md](deployment/upgrade-v1.2.0.md) | v1.2.0 升级指南 |

---

## 问题追踪

> 仅包含已关闭的问题。进行中的问题见 `backlog/known-issues.md`。

| 编号 | 问题 | 状态 |
|------|------|------|
| 001 | [支付成功后 VIP 激活失败](issues/001-payment-vip-activation-failure.md) | ✅ 已修复 |
| 002 | [管理后台 Dashboard 502 错误](issues/002-admin-dashboard-502.md) | ✅ 已修复 |
| 003 | [操作日志显示无数据/数据膨胀](issues/003-operation-logs-empty.md) | ✅ 已修复 |
| 004 | [Code Review 全面修复](issues/004-code-review-fixes.md) | ✅ 已修复 |
| 005 | [批量健康检查全异常 — Adapter 代理支持](issues/005-batch-healthcheck-proxy-fix.md) | ✅ 已修复 |
| 006 | [前端全面修复 16 个 Bug](issues/006-frontend-bug-fixes.md) | ✅ 待编译测试 |
| 007 | [P0 API 调用链路修复](issues/007-p0-api-chain-fixes.md) | ✅ 待编译测试 |
| 008 | [Phase 2 限速与计费完善](issues/008-phase2-compile-test.md) | ✅ 待编译测试 |

---

## 相关链接

- [OneAPI 参考](https://github.com/songquanpeng/one-api)
- [OpenAI API](https://platform.openai.com/docs)
- [Element Plus](https://element-plus.org/)
