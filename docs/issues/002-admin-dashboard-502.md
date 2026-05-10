# Issue #002: 管理后台Dashboard 502错误

## 问题描述

刷新管理后台时报错：`Request failed with status code 502`

## 根本原因

Dashboard.vue 使用了错误的 API 客户端：

```javascript
// 错误：导入了 userAPI
import request from '@/api/request'  // 这是 userAPI (baseURL: /api/v1)

// 调用了需要 admin 认证的接口
request.get('/admin/stats/overview')  // 请求 /api/v1/admin/stats/overview，但缺少 X-Admin-Secret header
```

## 影响

- 管理后台 Dashboard 页面无法加载统计数据
- 返回 502 错误（实际是认证失败被网关拒绝）

## 修复方案

修改 `frontend/src/views/admin/Dashboard.vue`：

```javascript
// 修复前
import request from '@/api/request'
request.get('/admin/stats/overview')

// 修复后
import { adminAPI } from '@/api/request'
adminAPI.get('/stats/overview')
```

## 修改文件

- `frontend/src/views/admin/Dashboard.vue`

## 发现日期

2026-04-01

## 状态

✅ 已修复

## 预防措施

- 管理员页面必须使用 `adminAPI` 而不是 `userAPI`
- `adminAPI` 自动添加 `X-Admin-Secret` header
