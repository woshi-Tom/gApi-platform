# Issue #003: 操作日志显示无数据

## 问题描述

管理后台 → 操作日志页面显示 "no data"，即使添加筛选器也无数据。

## 根本原因

前端响应解析错误：
- 后端返回: `{ success: true, data: [...array...], pagination: {...} }`
- 前端期望: `{ success: true, data: { list: [...], pagination: {...} } }`

前端代码尝试访问 `res.data.data.list`，但实际数据在 `res.data.data` 直接返回。

## 影响

- 操作日志页面无法显示数据
- 显示 "no data"

## 修复方案

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

- `frontend/src/views/admin/logs/Index.vue`

## 发现日期

2026-04-01

## 状态

✅ 已修复
