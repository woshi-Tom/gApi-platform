# 系统初始化向导设计

> 最后更新: 2026-05-13
> 状态: ✅ 已实现

---

## 一、功能概述

首次部署时，系统处于未初始化状态（无管理员用户、无数据库配置等）。
初始化向导引导管理员通过 Web 界面完成系统初始化配置。

---

## 二、触发条件

| 条件 | 行为 |
|------|------|
| 数据库中不存在 `admin_users` 记录 | 系统标记为 "未初始化" |
| 已存在 `admin_users` 记录 | 系统标记为 "已初始化" |

---

## 三、前端路由逻辑

**文件**: `frontend/src/admin-router.ts`

```
每次路由跳转前:
  1. 调用 GET /api/v1/init/status 检查初始化状态
  2. 如果未初始化 → 强制跳转到 /init 路径
  3. 如果已初始化 → 正常导航

访问 /init 路径时:
  1. 检查初始化状态
  2. 如果已初始化 → 跳转到 /login
  3. 如果未初始化 → 展示初始化配置页面
```

**设计要点**:
- 每次路由跳转都检查 init 状态（而非仅 `/init` 路径），确保未初始化时任何操作都被拦截
- `/init` 路径本身也做反向检查，已初始化后不再展示配置向导

---

## 四、后端 API

所有接口位于 `/api/v1/init/` 路由组，受 `InitProtection` 中间件保护（已初始化后拒绝访问）。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/init/status` | 返回初始化状态 `{initialized: true/false}` |
| POST | `/api/v1/init/test-db` | 测试数据库连接（当前配置） |
| POST | `/api/v1/init/test-db-with-config` | 使用指定参数测试数据库连接 |
| POST | `/api/v1/init/init-db` | 初始化数据库（运行 AutoMigrate） |
| POST | `/api/v1/init/test-redis` | 测试 Redis 连接 |
| POST | `/api/v1/init/create-admin` | 创建初始管理员用户 |

---

## 五、配置项

在 `settings.system_configs` 表中存储的通用初始化参数（非敏感信息可通过 API 读取）：

| 字段 | 说明 |
|------|------|
| `site_name` | 站点名称 |
| `site_logo` | 站点 Logo URL |
| `site_description` | 站点描述 |

---

## 六、安全

- 已初始化后 `InitProtection` 中间件阻止所有 `/api/v1/init/*` 请求
- 初始化完成后不可再次进入向导（除非清空数据库）
