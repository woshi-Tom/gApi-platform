# Issue #003: 操作日志显示无数据

## 问题描述

管理后台 → 操作日志页面显示 "no data"，即使添加筛选器也无数据。

## 根本原因

后端响应格式不统一：
- `response.Paginated()` 返回: `{ success: true, data: [...array...], pagination: {...} }`
- 前端期望: `{ success: true, data: { list: [...], pagination: {...} } }`

## 修复方案

### 1. 后端统一响应格式

修改 `backend/internal/handler/admin_handler.go`：

```go
// 修复前
response.Paginated(c, orders, page, pageSize, total)

// 修复后
response.Success(c, gin.H{
    "list": orders,
    "pagination": gin.H{
        "page":      page,
        "page_size": pageSize,
        "total":     total,
    },
})
```

影响范围：
- `ListOrders` - 订单列表
- `GetAuditLogs` - 操作日志
- `GetLoginLogs` - 登录日志

### 2. 前端解析修复

修改 `frontend/src/views/admin/logs/Index.vue`：

```javascript
// 修复前
logs.value = res.data.data?.list || []
total.value = res.data.data?.pagination?.total || 0

// 修复后
logs.value = res.data.data || []
total.value = res.data.pagination?.total || 0
```

## 修改文件

- `backend/internal/handler/admin_handler.go`
- `frontend/src/views/admin/logs/Index.vue`

## 发现日期

2026-04-01

## 状态

✅ 已修复

## 提交记录

- `ae26666` - 统一后端分页响应格式
- `28c6272` - 修复前端响应解析错误
