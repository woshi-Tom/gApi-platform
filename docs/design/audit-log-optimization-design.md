# 审计日志优化设计方案

> 版本: v1.0
> 日期: 2026-04-01
> 状态: 待评估
> 负责人: 团队评审

---

## 问题现状

### 当前问题

| 问题 | 描述 | 影响 |
|------|------|------|
| 数据膨胀 | `audit_logs.response_body` 存储完整API响应 | 单条记录可达19MB |
| 分类混乱 | `system` 组包含797条记录，大部分是HTTP访问 | 难以区分业务操作 |
| 响应缓慢 | 列表接口返回大量无用数据 | 页面加载慢 |
| 存储浪费 | 无用数据占用大量数据库空间 | 存储成本高 |

### 当前数据分布

```
action_group | count
-------------+-------
 system       |   797  ← 主要是 HTTP 访问日志（无用）
 payment      |   567  ← 真实业务操作
 auth         |   519  ← 真实业务操作
 order        |   375  ← 真实业务操作
 token        |   292  ← 真实业务操作
 vip          |     2
```

### 问题根源

审计中间件 `AuditLog()` 记录了**所有HTTP请求**，包括：
- ✅ 应该记录的：用户创建订单、修改渠道、登录等业务操作
- ❌ 不应该记录的：GET /products、GET /logs/operation 等查询请求

```go
// 当前行为：所有请求都被记录
c.Next()  // 处理请求
// ... 记录审计日志（包含完整 response_body）
```

---

## 解决方案评估

### 方案A：分表存储（推荐）

#### 设计思路

| 表名 | 用途 | 存储内容 |
|------|------|----------|
| `operation_logs` | 业务操作日志 | 用户的管理操作（创建、更新、删除、支付等） |
| `access_logs` | HTTP访问日志 | 所有API请求（可选，用于调试） |

#### 表结构

```sql
-- 业务操作日志（只记录重要操作）
CREATE TABLE operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    username VARCHAR(100),
    action VARCHAR(100) NOT NULL,      -- user.create, channel.update
    action_group VARCHAR(50) NOT NULL,  -- user, channel, order
    resource_type VARCHAR(50),          -- user, channel, order
    resource_id BIGINT,
    request_ip VARCHAR(50),
    request_method VARCHAR(10),
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    old_value JSONB,                   -- 变更前的值
    new_value JSONB,                   -- 变更后的值
    created_at TIMESTAMP DEFAULT NOW()
);

-- HTTP访问日志（可选，轻量级）
CREATE TABLE access_logs (
    id BIGSERIAL PRIMARY KEY,
    request_path VARCHAR(500),
    request_method VARCHAR(10),
    request_ip VARCHAR(50),
    status_code INT,
    response_time_ms INT,
    user_id BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 不要存储完整的 request_body 和 response_body
```

#### 优点

- ✅ 职责分离，结构清晰
- ✅ `operation_logs` 轻量，只存储业务操作
- ✅ `access_logs` 可按需启用/禁用
- ✅ 易于扩展和优化

#### 缺点

- ⚠️ 需要数据迁移
- ⚠️ 需要修改审计中间件
- ⚠️ 改动较大

---

### 方案B：过滤不必要请求（快速修复）

#### 设计思路

在审计中间件中过滤掉：
- 所有 GET 请求
- 列表/查询类请求

只记录写操作：
- POST, PUT, DELETE 请求
- 支付等重要操作

#### 代码修改

```go
// AuditLog 中间件优化
func AuditLog(auditRepo *repository.AuditRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 跳过不需要记录的路径
        if skipPaths[c.Request.URL.Path] {
            c.Next()
            return
        }

        // 2. 只记录写操作和重要操作
        method := c.Request.Method
        if method == "GET" && !isImportantGet(c.Request.URL.Path) {
            c.Next()
            return
        }

        // 3. 不存储完整的 response_body
        // ...
    }
}

func isImportantGet(path string) bool {
    // 这些 GET 请求需要记录
    importantPaths := []string{
        "/api/v1/payment/callback",
    }
    for _, p := range importantPaths {
        if strings.HasSuffix(path, p) {
            return true
        }
    }
    return false
}
```

#### 优点

- ✅ 改动小，可以快速实施
- ✅ 解决问题根源
- ✅ 不需要分表

#### 缺点

- ⚠️ 仍然可能漏掉一些重要的GET操作
- ⚠️ 分类不够清晰

---

### 方案C：前端分页优化（辅助方案）

#### 设计思路

1. 后端：默认只返回必要的字段，不返回 `request_body` 和 `response_body`
2. 前端：支持展开查看详情

#### 代码修改

```go
// 后端：列表接口不返回 body 字段
type AuditLogBrief struct {
    ID            uint      `json:"id"`
    Action        string    `json:"action"`
    ActionGroup   string    `json:"action_group"`
    Username      string    `json:"username"`
    RequestIP     string    `json:"request_ip"`
    Success       bool      `json:"success"`
    ErrorMessage  string    `json:"error_message,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    // 不包含 request_body, response_body, old_value, new_value
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
    // 查询时排除大字段
    logs, total := h.auditRepo.ListBrief(page, pageSize, ...)
    
    response.Success(c, gin.H{
        "list": logs,  // AuditLogBrief 列表，轻量
        "pagination": ...
    })
}

// 前端：点击详情时再请求完整数据
async function showDetail(log) {
    const res = await adminAPI.get(`/logs/operation/${log.id}`)
    // 只在弹窗中显示详情
}
```

#### 优点

- ✅ 改动静默，不影响现有逻辑
- ✅ 显著减少数据传输

#### 缺点

- ⚠️ 需要两次请求（列表 + 详情）
- ⚠️ 不能根本解决问题

---

## 推荐方案

### 组合方案：方案B + 方案C

| 优先级 | 措施 | 理由 |
|--------|------|------|
| 1️⃣ 高 | 方案B：过滤不必要请求 | 快速修复，根本解决数据膨胀 |
| 2️⃣ 中 | 方案C：列表不返回body | 减少传输，加速页面加载 |
| 3️⃣ 低 | 方案A：分表存储 | 未来扩展，预留空间 |

### 实施步骤

#### Phase 1: 快速修复（当天）

1. 修改审计中间件，跳过 GET 请求
2. 不存储 request_body 和 response_body
3. 清理现有数据中的大字段

```go
// 修改点：backend/internal/middleware/audit.go
var skipMethods = map[string]bool{
    "GET": true,  // 跳过所有 GET 请求
}
```

#### Phase 2: 优化展示（1-2天）

1. 修改列表接口，使用 AuditLogBrief
2. 前端支持展开查看详情
3. 添加索引优化查询

#### Phase 3: 长期规划（可选）

1. 设计分表方案
2. 数据迁移
3. 归档策略

---

## 数据清理

### 清理脚本

```sql
-- 清理 response_body 字段（已损坏的大数据）
UPDATE audit_logs SET response_body = NULL 
WHERE length(response_body) > 1000;

-- 清理过大的 request_body
UPDATE audit_logs SET request_body = NULL 
WHERE length(request_body) > 5000;

-- 删除旧数据（保留30天）
DELETE FROM audit_logs 
WHERE created_at < NOW() - INTERVAL '30 days';

-- 查看清理效果
SELECT 
    COUNT(*) as total,
    SUM(length(response_body)) as total_resp_size,
    AVG(length(response_body)) as avg_resp_size
FROM audit_logs;
```

### 预防措施

1. **数据库约束**：添加 CHECK 限制字段大小
```sql
ALTER TABLE audit_logs 
ADD CONSTRAINT chk_response_body_size 
CHECK (length(response_body) <= 10000);
```

2. **配置化**：添加审计日志配置
```yaml
audit:
  enabled: true
  max_body_size: 10000  # bytes
  log_get_requests: false
  log_paths_whitelist: []
```

---

## 评估清单

| 评估项 | 方案A | 方案B | 方案C |
|--------|-------|-------|-------|
| 实施难度 | 高 | 低 | 中 |
| 见效速度 | 慢 | 快 | 快 |
| 代码质量 | 高 | 中 | 高 |
| 扩展性 | 高 | 低 | 中 |
| 风险 | 中 | 低 | 低 |
| 建议 | 长期方案 | **首选** | 辅助 |

---

## 行动项

- [ ] 团队评审此设计方案
- [ ] 确定采用方案（建议 B+C）
- [ ] 实施 Phase 1 快速修复
- [ ] 验证效果
- [ ] 规划 Phase 2

---

## 参考

- 当前审计中间件：`backend/internal/middleware/audit.go`
- 操作日志前端：`frontend/src/views/admin/logs/Index.vue`
- 问题记录：`docs/issues/003-operation-logs-empty.md`
