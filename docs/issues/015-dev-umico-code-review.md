# 015 — dev-umico PR #3/#4 代码审查问题

**日期**: 2026-05-15
**审查范围**: `17b7833`, `9d5be7a` (dev-umico → main, 2 PR)
**审查人**: TOM
**严重度**: 中等

## 变更概述

dev-umico 提交了 2 个 PR 合入 main：
- **PR #3** (`17b7833`): Redis 密码修复（docker-compose.yml + redis.conf）
- **PR #4** (`9d5be7a`): 部署配置修复 + 初始化流程改进 + admin 登录支持 DB 查询

## 发现的问题

### 1. AutoMigrate 模型列表不完整（中等）

`init_handler.go` 的 `InitializeDatabaseDefault()` 和 `InitializeDatabase()` 内手动列出 AutoMigrate 模型，与 `repository/database.go` 的权威列表不一致，**缺少 8 个模型**：

| 缺失模型 | 所在文件 | 用途 |
|----------|----------|------|
| `Tenant` | user.go:90 | 多租户 |
| `UserRechargeRecord` | token.go:197 | 充值记录 |
| `EmailVerification` | email_verification.go:7 | 邮箱验证 |
| `PasswordReset` | email_verification.go:27 | 密码重置 |
| `ModelGroup` | model_group.go:9 | 模型分组 |
| `ChannelGroupRelation` | model_group.go:28 | 通道-分组关联 |
| `UserGroupRelation` | model_group.go:44 | 用户-分组关联 |
| `ModelPricing` | model_group.go:58 | 模型定价 |

**影响**: 通过 InitWizard 初始化的系统，后续访问 Model Groups / Pricing / 邮箱验证 等功能会因表缺失而报错。

### 2. Admin Login 无失败日志记录（中等）

`AdminHandler.Login()` 有 `loginLogRepo` 字段但完全未使用。用户登录流程（`UserHandler.Login`）有完整的登录日志记录（成功/失败/IP/UserAgent），admin 登录缺少同等保护。

**影响**: 无法检测针对 admin 登录的暴力破解攻击。

### 3. Config 与 DB 管理员同名双密码隐式行为（低）

`AdminHandler.Login()` 先遍历 config admin users，用户名匹配但密码错误时 `continue`，然后落到 DB 查询。如果 config 和 DB 存在同名管理员，该用户名将有两个独立密码同时有效。

**影响**: 运维困惑，非预期行为。需要注释说明双通道设计意图。

### 4. InitializeDatabaseDefault 无幂等保护（低）

系统未初始化期间可反复调用 `/init/init-db-default`，每次都执行 AutoMigrate。虽然 GORM AutoMigrate 通常可重入，但某些约束操作可能失败 —— `database.go` 对 AutoMigrate 错误只 warn，而 `init_handler.go` 返回失败给用户。

**影响**: 重复调用可能显示虚假错误。

### 5. InitWizard onMounted Loading 状态延迟（低）

`InitWizard.vue` 的 `onMounted` 检测到 DB 已连接后直接调用 `tryAutoInit()`，但 `autoInitMode = true` 只有在 `tryAutoInit` 函数内部才设置。在异步调用发起前有短暂间隙，UI 短暂显示空白表单。

**影响**: 短暂 UI 闪动。

## 修复方案

| # | 问题 | 修复方案 |
|---|------|----------|
| 1 | AutoMigrate 列表不完整 | `repository/` 新增 `AutoMigrateModels()` 函数返回完整模型列表，`database.go` 和 `init_handler.go` 共同引用 |
| 2 | admin 登录无日志 | `AdminHandler.Login()` 失败/成功时调用 `loginLogRepo.Record()` |
| 3 | config admin 安全后门 | 彻底删除 config admin_users 登录路径，DB 为唯一认证源；`config.AdminUsers` 标记 Deprecated |
| 4 | 无幂等保护 | `InitializeDatabaseDefault` 先检查 `SystemConfig` 中 `system_initialized` 标记 |
| 5 | autoInitMode UI 延迟 | `onMounted` 中提前设置 `autoInitMode = true` 再调用 `tryAutoInit` |

## 修复状态

- [x] #1: AutoMigrate 统一 — `repository.AutoMigrateModels()` 供 database.go 和 init_handler.go 共用
- [x] #2: admin 登录日志 — Login 在 user-not-found / wrong-password / success 三个分支记录
- [x] #3: 删除 config admin 路径 — 移除 `AdminHandler.adminUsers`、`InitHandler.cfgAdmin`，DB 为唯一认证源
- [x] #4: init-db-default 幂等 — 检查 `system_initialized` 标记，已初始化则跳过
- [x] #5: InitWizard UI 优化 — `onMounted` 提前设置 `autoInitMode = true`

**修复分支**: dev-tom
