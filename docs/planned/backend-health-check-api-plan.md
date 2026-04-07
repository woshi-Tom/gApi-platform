# 后端需求计划 - 渠道手动健康检测API

> 版本: v1.0
> 日期: 2026-04-07
> 状态: ✅ 已完成
> 前端依赖: 高 - 需要此API才能完成手动检测功能

---

## 1. 需求概述

前端渠道管理列表需要添加"手动健康检测"功能，用户点击按钮可立即触发对指定渠道的健康检测。

**约束**: 前端不能修改后端代码，需要后端实现以下功能。

---

## 2. 需求详情

### 2.1 功能需求

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 手动触发检测 | 🔴 高 | 管理员点击按钮触发单渠道健康检测 |
| 返回检测结果 | 🔴 高 | 检测完成后返回成功/失败状态和错误信息 |

### 2.2 API 接口需求

#### 2.2.1 POST - 手动触发健康检测

**请求**:
```
POST /api/v1/admin/channels/:id/health
Content-Type: application/json
Authorization: Bearer <admin_jwt_token>
```

**响应 (成功)**:
```json
{
  "success": true,
  "data": {
    "is_healthy": true,
    "failure_count": 0,
    "last_check_at": "2026-04-07T12:34:56Z",
    "last_error": null,
    "response_time_ms": 230
  }
}
```

**响应 (失败)**:
```json
{
  "success": false,
  "message": "渠道连接失败",
  "data": {
    "is_healthy": false,
    "failure_count": 1,
    "last_check_at": "2026-04-07T12:34:56Z",
    "last_error": "connection timeout",
    "response_time_ms": 30000
  }
}
```

**响应 (渠道不存在)**:
```json
{
  "success": false,
  "message": "渠道不存在"
}
```

#### 2.2.2 GET - 获取健康状态 (可选，如列表接口已包含则不需要)

```
GET /api/v1/admin/channels/:id/health
```

**说明**: 如果渠道列表接口 `/api/v1/admin/channels` 已返回完整的健康状态字段（包括 `last_check_at`, `last_error`, `response_time_avg`），则此接口可能不需要。

---

## 3. 后端实现参考

### 3.1 现有健康检测服务

文件: `backend/internal/service/health_check.go`

已有方法:
- `CheckChannelManually(channelID uint)` - 手动检测单渠道
- `GetChannelStatus(channelID uint)` - 获取渠道状态

### 3.2 路由配置

文件: `backend/internal/router/router.go`

需要在 `adminAuth` 路由组中添加:
```go
adminAuth.POST("/channels/:id/health", channelHandler.TriggerHealthCheck)
```

### 3.3 数据库字段 (已存在)

```sql
-- channels 表已包含以下字段:
is_healthy BOOLEAN DEFAULT true
failure_count INTEGER DEFAULT 0
last_success_at TIMESTAMP
last_check_at TIMESTAMP
last_error TEXT
response_time_avg INTEGER DEFAULT 0
```

---

## 4. 前端集成计划

### 4.1 前端需要做的

1. **API方法** (`frontend/src/api/channel.ts`):
   ```typescript
   triggerHealthCheck: (id: number) => 
     adminAPI.post(`/channels/${id}/health`)
   ```

2. **UI按钮** (在 `List.vue` 每行操作列):
   - 点击调用 `triggerHealthCheck`
   - 显示加载状态
   - 成功后刷新列表
   - 失败后显示错误消息

### 4.2 前端展示效果

```
操作列:
[编辑] [测试] [检测] [禁用/启用] [删除]
         ↑ 点击后触发健康检测
```

---

## 5. 验证清单

实现完成后需验证:

- [ ] POST /channels/:id/health 返回正确格式
- [ ] 成功时返回 `is_healthy: true`
- [ ] 失败时返回 `is_healthy: false` 和 `last_error`
- [ ] 响应包含 `last_check_at` 时间戳
- [ ] 响应包含 `response_time_ms`
- [ ] 未授权返回 401
- [ ] 渠道不存在返回 404

---

## 6. 优先级

| 项目 | 优先级 | 说明 |
|------|--------|------|
| POST /channels/:id/health | 🔴 高 | 手动检测核心功能 |

---

## 7. 参考资料

- 健康检测设计文档: `docs/design/channel-health-check-design.md`
- 渠道列表前端: `frontend/src/views/admin/channels/List.vue`
- 渠道API定义: `frontend/src/api/channel.ts`
- 后端路由: `backend/internal/router/router.go`
- 后端健康检测服务: `backend/internal/service/health_check.go`

---

## 8. 备注

- 手动检测应该与定时检测使用相同的检测逻辑
- 检测应该是同步的（等待检测完成后再返回结果），还是异步的？
  - 建议: 同步等待，这样前端可以立即看到结果
- 检测超时时间建议使用渠道配置的 timeout 值