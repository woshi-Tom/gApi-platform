# 账号池功能开发 - 任务清单

> 更新日期：2026-03-21  
> 状态：规划完成，待执行

---

## 📋 任务概览

| 阶段 | 任务数 | 预计时间 | 优先级 |
|------|--------|----------|--------|
| 项目合并 | 3 | 1-2 天 | P1 |
| 数据库设计 | 4 | 0.5 天 | P0 |
| 渠道管理 | 5 | 2 天 | P1 |
| 健康检测 | 4 | 1 天 | P1 |
| Token 管理 | 4 | 1 天 | P1 |
| 审计日志 | 4 | 1 天 | P1 |
| API 代理 | 6 | 2 天 | P0 |
| 前端实现 | 5 | 2 天 | P1 |
| 测试部署 | 3 | 1 天 | P1 |
| **总计** | **38** | **11.5 天** | - |

---

## 🎯 P0 阻塞任务（必须优先完成）

### 1. 数据库迁移脚本

| 项目 | 内容 |
|------|------|
| 文件 | `backend/migrations/xxx_add_account_pool.sql` |
| 内容 | 创建 channels, abilities, tokens, usage_logs, audit_logs 表 |
| 依赖 | 无 |

```sql
-- 主要DDL（详见 account-pool-plan.md）
CREATE TABLE channels (...);
CREATE TABLE abilities (...);
CREATE TABLE tokens (...);
CREATE TABLE usage_logs (...);
CREATE TABLE audit_logs (...);
ALTER TABLE users ADD COLUMN quota, used_quota, api_group;
```

### 2. API 代理核心接口

| 项目 | 内容 |
|------|------|
| 文件 | `backend/app/blueprints/api_proxy.py` |
| 接口 | `POST /v1/chat/completions`, `GET /v1/models` |
| 依赖 | channels, abilities, tokens 表 |

---

## 🟠 P1 高优先级任务

### 3. 渠道管理服务

| ID | 任务 | 依赖 |
|----|------|------|
| T3.1 | Channel 模型 | DB |
| T3.2 | Channel Service | Model |
| T3.3 | Admin API (CRUD) | Service |
| T3.4 | API Key 加密 | 无 |
| T3.5 | 批量导入 (YAML) | Service |

### 4. 健康检测服务

| ID | 任务 | 依赖 |
|----|------|------|
| T4.1 | 健康检测 Service | Channel |
| T4.2 | APScheduler 定时任务 | Service |
| T4.3 | 手动检测 API | Service |
| T4.4 | 恢复死亡渠道 | Service |

### 5. Token 管理

| ID | 任务 | 依赖 |
|----|------|------|
| T5.1 | Token 模型 | DB |
| T5.2 | Token 生成器 | 无 |
| T5.3 | Token Service | Model |
| T5.4 | Admin API (CRUD) | Service |

### 6. 审计日志

| ID | 任务 | 依赖 |
|----|------|------|
| T6.1 | AuditLog 模型 | DB |
| T6.2 | Audit Service | Model |
| T6.3 | 登录日志 | Auth |
| T6.4 | 查询/导出 API | Service |

---

## 🟡 P2 中优先级任务

### 7. 分发与计费

| ID | 任务 | 依赖 |
|----|------|------|
| T7.1 | Distributor Service | Channel |
| T7.2 | Billing Service | Token |
| T7.3 | 失败重试逻辑 | Distributor |
| T7.4 | 配额预扣/调整 | Billing |

### 8. 适配器实现

| ID | 任务 | 依赖 |
|----|------|------|
| T8.1 | Base Adapter | 无 |
| T8.2 | OpenAI Adapter | Base |
| T8.3 | Claude Adapter | Base |
| T8.4 | NVIDIA NIM Adapter | Base |
| T8.5 | 其他适配器 | Base |

---

## 🟢 P3 低优先级任务（可后续迭代）

### 9. 高级功能

- [ ] 使用统计看板
- [ ] 配额预警通知
- [ ] Redis 缓存层
- [ ] 分布式部署支持
- [ ] 多租户隔离

---

## ✅ 完成标准

每个任务完成后必须满足：

1. **代码完成** - 功能实现完毕
2. **单元测试** - 核心逻辑有测试覆盖
3. **文档更新** - `docs/PROJECT_PROGRESS.md` 更新
4. **验收通过** - 人工验证功能正常

---

## 📊 进度追踪

| 里程碑 | 目标日期 | 状态 | 验收结果 |
|--------|----------|------|----------|
| M1: 项目合并 | Day 2 | ⬜ | - |
| M2: 渠道管理 | Day 5 | ⬜ | - |
| M3: Token 管理 | Day 6 | ⬜ | - |
| M4: API 代理 | Day 8 | ⬜ | - |
| M5: 上线部署 | Day 12 | ⬜ | - |
