# Issue #003: 操作日志显示无数据 / 数据膨胀

## 问题描述

1. 管理后台 → 操作日志页面显示 "no data"
2. API 响应异常大（首次1MB，后续19MB+）

## 根本原因

1. **显示无数据**：后端响应格式不统一
2. **数据膨胀**：审计中间件记录所有API响应，包括审计日志本身

## 解决方案

采用**方案A：数据分类 + 前端优化**

| 改动 | 说明 |
|------|------|
| 添加 `log_type` 字段 | 区分 operation(操作) 和 access(访问) |
| 添加 `response_time_ms` | 记录响应时间 |
| 限制 body 大小 | 最多50KB，防止再次膨胀 |
| 列表不返回 body | 使用 ListBrief() 只返回概要 |
| 点击详情加载 | GetAuditLogDetail() 获取完整数据 |

## 修改文件

- `backend/internal/model/audit.go` - 添加 LogType 常量
- `backend/internal/repository/user_repo.go` - 添加 ListBrief, GetByID
- `backend/internal/handler/admin_handler.go` - 修改 GetAuditLogs, 添加 GetAuditLogDetail
- `backend/internal/middleware/audit.go` - 设置 log_type，限制 body
- `backend/internal/router/router.go` - 添加详情路由
- `frontend/src/api/log.ts` - 添加详情接口
- `frontend/src/views/admin/logs/Index.vue` - 优化列表和详情

## 状态

✅ 已修复并验证

## 提交记录

- `9260d05` - 审计日志优化完整实现
