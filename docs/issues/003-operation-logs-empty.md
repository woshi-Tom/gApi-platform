# Issue #003: 操作日志显示无数据 / 数据膨胀

## 问题描述

1. 管理后台 → 操作日志页面显示 "no data"
2. API 响应异常大（首次1MB，后续19MB+）

## 问题分析

### 问题1：显示无数据

**原因**：后端响应格式不统一
- `response.Paginated()` 返回数据直接在 `data` 数组
- 前端期望 `data.list` 结构

**修复**：统一响应格式为 `{ success, data: { list, pagination } }`

### 问题2：数据膨胀（根因）

**原因**：审计中间件会记录**所有API响应**，包括审计日志本身

| ID | Action | Response Body 大小 |
|----|--------|-------------------|
| 2550 | GET.logs/operation | **19,619,669 bytes (19MB!)** |
| 2549 | GET.logs/operation | 7,137,166 bytes (7MB) |
| 2548 | GET.logs/operation | 2,675,697 bytes (2.7MB) |

形成恶性循环：
1. 请求 `/logs/operation` 返回大量数据
2. 审计中间件将响应存入 `audit_logs.response_body`
3. 下次请求时数据量更大
4. 最终导致响应超时或页面崩溃

## 修复方案

### 1. 后端统一响应格式

修改 `backend/internal/handler/admin_handler.go`：
- `ListOrders`
- `GetAuditLogs`  
- `GetLoginLogs`

改用 `response.Success()` 返回 `{ list, pagination }` 结构

### 2. 前端解析修复

修改 `frontend/src/views/admin/logs/Index.vue`：
```javascript
logs.value = res.data.data?.list || []
total.value = res.data.data?.pagination?.total || 0
```

### 3. 跳过审计日志本身的记录

修改 `backend/internal/middleware/audit.go`：

```go
var skipPaths = map[string]bool{
	"/api/v1/internal/health":    true,
	"/health":                   true,
	"/ping":                     true,
	"/api/v1/admin/logs/operation": true, // 避免审计日志本身的记录形成数据膨胀
	"/api/v1/admin/logs/login":     true,
}
```

## 修改文件

- `backend/internal/handler/admin_handler.go`
- `frontend/src/views/admin/logs/Index.vue`
- `backend/internal/middleware/audit.go`

## 发现日期

2026-04-01

## 状态

✅ 已修复

## 提交记录

- `ae26666` - 统一后端分页响应格式
- `28c6272` - 修复前端响应解析错误
- 新增 - 跳过审计日志路径防止数据膨胀
