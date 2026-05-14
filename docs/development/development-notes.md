# gAPI Platform - Development Notes & Implementation Checklist

> Last Updated: 2026-05-15
> Purpose: Capture all detail issues and pending items for development session
>
> API 接口清单（第 3 节）已于 2026-05-14 与 router.go 逐行核对。

---

## 1. Environment & Configuration Requirements

### 1.1 Required Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `GAPI_MODE` | Run mode: `development`, `production` | `production` | Yes |
| `GAPI_DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `GAPI_DB_PORT` | PostgreSQL port | `5432` | Yes |
| `GAPI_DB_USER` | PostgreSQL user | `gapi_user` | Yes |
| `GAPI_DB_PASSWORD` | PostgreSQL password | `***` | Yes |
| `GAPI_DB_NAME` | Database name | `gapi` | Yes |
| `GAPI_REDIS_HOST` | Redis host | `localhost` | Yes |
| `GAPI_REDIS_PORT` | Redis port | `6379` | Yes |
| `GAPI_REDIS_PASSWORD` | Redis password | `***` | Optional |
| `GAPI_RABBITMQ_HOST` | RabbitMQ host | `localhost` | Yes |
| `GAPI_RABBITMQ_PORT` | RabbitMQ port | `5672` | Yes |
| `GAPI_RABBITMQ_USER` | RabbitMQ user | `guest` | Yes |
| `GAPI_RABBITMQ_PASSWORD` | RabbitMQ password | `***` | Yes |
| `GAPI_LOG_PATH` | Log directory | `/var/log/gapi` | Yes |
| `GAPI_JWT_SECRET` | JWT signing key | Random 32+ chars | Yes |
| `GAPI_ADMIN_SECRET` | Admin page access key | Random string | Yes |
| `GAPI_SERVER_PORT` | HTTP server port | `8080` | Yes |
| `GAPI_FRONTEND_URL` | Frontend URL | `http://localhost:5173` | Yes |

### 1.2 Directory Structure Requirements

```
/var/log/gapi/              # Log directory (writable)
├── access.log              # Access logs
├── error.log               # Error logs
├── operation.log           # Operation logs (database)
└── audit.log               # Audit logs

/etc/gapi/                  # Config directory (optional)
└── config.yaml             # YAML config file (optional, env vars preferred)

/var/lib/gapi/              # Data directory
├── uploads/                # File uploads
└── temp/                   # Temporary files
```

### 1.3 First-Run Initialization Flow

```
1. Detect no admin user exists → Show initialization wizard
2. Step 1: Database connection test → Create tables from DDL
3. Step 2: Redis connection test → Initialize cache
4. Step 3: RabbitMQ connection test → Initialize queues
5. Step 4: Create admin account (username, password, email)
6. Step 5: Configure system settings (log path, JWT secret)
7. Complete → Redirect to admin login
```

---

## 2. Database Implementation Notes

### 2.1 Table Creation Order (Foreign Key Constraints)

```sql
-- Core tables (no dependencies)
1. channels (base channel info)
2. models (model definitions)
3. channel_models (channel -> model mapping)

-- User-related tables
4. users (user accounts)
5. user_tokens (API tokens)
6. user_quota (quota records)

-- Product & Order tables
7. products (products/subscriptions)
8. product_pricing (pricing tiers)
9. orders (order records)
10. order_items (order line items)

-- VIP & Channel Access
11. vip_subscriptions (VIP subscriptions)
12. channel_access_logs (access logs for billing)

-- Audit & Logs
13. operation_logs (database operation logs)
14. api_logs (API call logs)
15. login_logs (login attempt logs)
16. payment_logs (payment transaction logs)

-- Settings & Maintenance
17. system_settings (system configuration)
18. channel_test_results (health check results)
```

### 2.2 Indexes Summary

| Table | Index | Type | Columns |
|-------|-------|------|---------|
| users | idx_users_email | unique | email |
| users | idx_users_username | unique | username |
| user_tokens | idx_tokens_user_id | normal | user_id |
| user_tokens | idx_tokens_key | unique | token_key |
| channels | idx_channels_status | normal | status |
| channels | idx_channels_priority | normal | priority |
| models | idx_models_type | normal | model_type |
| orders | idx_orders_user_id | normal | user_id |
| orders | idx_orders_status | normal | status |
| orders | idx_orders_created | normal | created_at |
| api_logs | idx_api_logs_user_id | normal | user_id |
| api_logs | idx_api_logs_created | normal | created_at |
| operation_logs | idx_ops_logs_level | normal | level |
| operation_logs | idx_ops_logs_user_id | normal | user_id |

### 2.3 Migration Strategy

- **Version**: Add `schema_version` table to track migrations
- **Rollback**: Keep DDL comments with version numbers
- **Seeding**: Run seed data after schema creation

### 2.4 Database Column Naming Conventions (IMPORTANT)

⚠️ **数据库列名规范 - 必须遵守**

| 字段类型 | 列名格式 | 示例 |
|---------|---------|------|
| VIP相关 | `v_ip_xxx` | `v_ip_quota`, `v_ip_expired_at`, `v_ip_package_id` |
| 普通字段 | `xxx_xxx` | `created_at`, `updated_at`, `remain_quota` |

⚠️ **代码中引用数据库列名时必须使用正确的下划线格式！**

| 错误写法 ❌ | 正确写法 ✅ |
|-------------|-------------|
| `Update("vip_quota", value)` | `Update("v_ip_quota", value)` |
| `Update("vip_expired_at", value)` | `Update("v_ip_expired_at", value)` |
| `Update("vip_package_id", value)` | `Update("v_ip_package_id", value)` |

### 2.5 Code Review Checklist

提交代码前必须检查：

- [ ] **数据库列名**：所有 SQL/GORM 更新语句中的列名必须与 schema.sql 一致
- [ ] **VIP字段**：确认使用 `v_ip_xxx` 格式
- [ ] **支付流程测试**：创建订单 → 完成支付 → 验证订单状态变为 `completed`
- [ ] **VIP激活验证**：验证用户 level/quota/expired_at 正确更新

---

## 3. API Implementation Details

> 状态标记：✅ 已实现 | ⬜ 未实现（按设计规划）
> 最后核对日期: 2026-05-14（与 router.go 逐行比对）

### 3.1 Northbound API (User-Facing)

| Endpoint | Method | Auth | Description | Status |
|----------|--------|------|-------------|--------|
| `/api/v1/user/register` | POST | No | User registration | ✅ |
| `/api/v1/user/login` | POST | No | User login | ✅ |
| `/api/v1/user/profile` | GET | JWT | Get profile | ✅ |
| `/api/v1/user/profile` | PUT | JWT | Update profile | ✅ |
| `/api/v1/user/change-password` | POST | JWT | Change password | ✅ |
| `/api/v1/tokens` | GET | JWT | List tokens | ✅ |
| `/api/v1/tokens` | POST | JWT | Create token | ✅ |
| `/api/v1/tokens/:id` | DELETE | JWT | Delete token | ✅ |
| `/api/v1/products` | GET | No | List products | ✅ |
| `/api/v1/products/:id` | GET | No | Product detail | ✅ |
| `/api/v1/orders` | GET | JWT | List orders | ✅ |
| `/api/v1/orders` | POST | JWT | Create order | ✅ |
| `/api/v1/orders/:id` | GET | JWT | Order detail | ✅ |
| `/api/v1/payment/alipay` | POST | JWT | Alipay payment | ✅ |
| `/api/v1/payment/wechat` | POST | JWT | WeChat payment | ⬜ |
| `/api/v1/payment/callback/alipay` | POST | No | Alipay webhook | ✅ |
| `/api/v1/payment/callback/wechat` | POST | No | WeChat webhook | ⬜ |
| `/api/v1/vip/status` | GET | JWT | VIP status | ✅（实际路径 `/user/vip/status`） |
| `/api/v1/quota` | GET | JWT | Quota info | ✅（实际路径 `/user/quota`） |
| `/api/v1/chat/completions` | POST | Token | OpenAI Chat Completions | ✅ |
| `/api/v1/models` | GET | Token | List models | ✅ |
| `/api/v1/embeddings` | POST | Token | Embeddings API | ✅ |
| `/api/v1/completions` | POST | Token | Text Completions | ✅ v1.3.0 |
| `/api/v1/messages` | POST | Token | Anthropic Messages API | ✅ v1.3.0 |
| `/api/v1/user/stats/usage` | GET | JWT | Usage stats | ✅ v1.2.0 |
| `/api/v1/user/activities` | GET | JWT | Recent activities | ✅ v1.2.0 |
| `/api/v1/orders/no/:order_no` | GET | JWT | Order by order number | ✅ |
| `/api/v1/payment/alipay/query/:order_no` | GET | JWT | Query Alipay order | ✅ |
| `/api/v1/payment/alipay/cancel/:order_no` | POST | JWT | Cancel Alipay order | ✅ |
| `/api/v1/payment/refund/:order_no` | POST | JWT | Refund order | ✅ |
| `/api/v1/payment/config` | GET | JWT | Payment config | ✅ |
| `/api/v1/redemption/redeem` | POST | JWT | Redeem code | ✅ v1.0.0 |
| `/api/v1/redemption/history` | GET | JWT | Redemption history | ✅ v1.0.0 |
| `/api/v1/logs` | GET | JWT | User API access logs | ✅ |
| `/api/v1/email/send-code` | POST | No | Send verification code | ✅ v1.0.0 |
| `/api/v1/email/verify-code` | POST | No | Verify code | ✅ v1.0.0 |
| `/api/v1/auth/forgot-password` | POST | No | Request password reset | ✅ v1.0.0 |
| `/api/v1/auth/reset-password` | GET | No | Verify reset token | ✅ v1.0.0 |
| `/api/v1/auth/reset-password` | POST | No | Reset password | ✅ v1.0.0 |
| `/api/v1/captcha/generate` | GET | No | Generate captcha | ✅ v1.0.0 |
| `/api/v1/captcha/verify` | POST | No | Verify captcha | ✅ v1.0.0 |
| `/api/v1/captcha/validate` | GET | No | Validate captcha token | ✅ v1.0.0 |
| `/api/v1/init/status` | GET | No | Init status | ✅ |
| `/api/v1/init/test-db` | POST | No | Test database | ✅ |
| `/api/v1/init/test-db-with-config` | POST | No | Test DB with config | ✅ |
| `/api/v1/init/init-db` | POST | No | Initialize database | ✅ |
| `/api/v1/init/test-redis` | POST | No | Test Redis | ✅ |
| `/api/v1/init/create-admin` | POST | No | Create admin | ✅ |

### 3.2 Southbound API (Internal)

> 需要 AdminSecret 认证。`internal/channels` 实际挂在 router v1 group 下。

| Endpoint | Method | Auth | Description | Status |
|----------|--------|------|-------------|--------|
| `/api/v1/internal/channels` | GET | Internal | List channels | ✅ |
| `/api/v1/internal/channels` | POST | Internal | Create channel | ✅ |
| `/api/v1/internal/channels/:id` | PUT | Internal | Update channel | ✅ |
| `/api/v1/internal/channels/:id` | DELETE | Internal | Delete channel | ✅ |
| `/api/v1/internal/channels/:id/test` | POST | Internal | Test channel | ✅ |
| `/api/v1/internal/models` | GET | Internal | List models | ⬜ |
| `/api/v1/internal/models/sync` | POST | Internal | Sync models | ⬜ |
| `/api/v1/internal/health` | GET | Internal | Health check | ✅ |
| `/api/v1/internal/balance/:channel_id` | GET | Internal | Get balance | ⬜ |

### 3.3 Admin API

> 所有 Admin 端点需要 JWT + X-Admin-Secret 双重认证（login 除外）。

| Endpoint | Method | Description | Status |
|----------|--------|-------------|--------|
| `/api/v1/admin/login` | POST | Admin login | ✅ |
| `/api/v1/admin/users` | GET | User list | ✅ |
| `/api/v1/admin/users/:id` | PUT | Update user | ✅ |
| `/api/v1/admin/users/:id/quota` | POST | Adjust quota | ✅ v1.4.0 |
| `/api/v1/admin/tokens` | GET | Token list | ✅ v1.4.0 |
| `/api/v1/admin/tokens/:id` | DELETE | Revoke token | ✅ v1.4.0 |
| `/api/v1/admin/products` | GET | Product list | ✅ |
| `/api/v1/admin/products` | POST | Create product | ✅ |
| `/api/v1/admin/products/:id` | PUT | Update product | ✅ |
| `/api/v1/admin/products/:id` | DELETE | Delete product | ✅ |
| `/api/v1/admin/orders` | GET | Order list | ✅ |
| `/api/v1/admin/orders/:id` | GET | Order detail | ✅ v1.4.0 |
| `/api/v1/admin/orders/:id/process` | POST | Process order | ✅ v1.4.0 |
| `/api/v1/admin/channels` | GET | Channel list | ✅ |
| `/api/v1/admin/channels` | POST | Create channel | ✅ |
| `/api/v1/admin/channels/:id` | PUT | Update channel | ✅ |
| `/api/v1/admin/channels/:id` | DELETE | Delete channel | ✅ |
| `/api/v1/admin/channels/:id/test` | POST | Test channel | ✅ |
| `/api/v1/admin/channels/batch-import` | POST | Batch import | ✅ |
| `/api/v1/admin/channels/:id/health` | POST | Trigger health check | ✅ v1.1.0 |
| `/api/v1/admin/channels/:id/test-history` | GET | Test history | ✅ v1.1.0 |
| `/api/v1/admin/channels/import` | POST | Batch import (alias) | ✅ |
| `/api/v1/admin/channels/:id/enable` | POST | Enable channel | ✅ |
| `/api/v1/admin/channels/:id/disable` | POST | Disable channel | ✅ |
| `/api/v1/admin/vip-subscriptions` | GET | VIP list | ⬜ |
| `/api/v1/admin/vip-subscriptions/:id` | PUT | Update VIP | ⬜ |
| `/api/v1/admin/logs/operation` | GET | Operation logs | ✅ |
| `/api/v1/admin/logs/operation/:id` | GET | Audit log detail | ✅ |
| `/api/v1/admin/logs/api` | GET | API logs | ⬜ |
| `/api/v1/admin/logs/login` | GET | Login logs | ✅ |
| `/api/v1/admin/logs/payment` | GET | Payment logs | ⬜ |
| `/api/v1/admin/logs/stats` | GET | Log statistics | ⬜ |
| `/api/v1/admin/stats/overview` | GET | Dashboard stats | ✅ |
| `/api/v1/admin/stats/trends` | GET | API request trends | ✅ |
| `/api/v1/admin/stats/user-overview` | GET | User stats overview | ✅ v1.2.0 |
| `/api/v1/admin/stats/user-ranking` | GET | User ranking | ✅ v1.2.0 |
| `/api/v1/admin/stats/user-list` | GET | User list for stats | ✅ v1.2.0 |
| `/api/v1/admin/stats/abnormal-users` | GET | Abnormal users | ✅ v1.2.0 |
| `/api/v1/admin/stats/user/:id/detail` | GET | User detail stats | ✅ v1.2.0 |
| `/api/v1/admin/change-password` | POST | Change admin password | ✅ |
| `/api/v1/admin/settings/email` | GET/PUT | SMTP config | ✅ v1.2.0 |
| `/api/v1/admin/settings/email/test` | POST | Test SMTP | ✅ v1.2.0 |
| `/api/v1/admin/settings/register` | GET/PUT | Registration settings | ✅ v1.2.0 |
| `/api/v1/admin/settings/payment` | GET/PUT | Payment settings | ✅ v1.2.0 |
| `/api/v1/admin/settings/general` | GET/PUT | General settings | ✅ v1.2.0 |
| `/api/v1/admin/settings/rate-limit` | GET/PUT | Rate limit settings | ✅ v1.2.0 |
| `/api/v1/admin/settings/security` | GET/PUT | Security settings | ✅ v1.2.0 |
| `/api/v1/admin/redemption/codes` | GET | List codes | ✅ v1.0.0 |
| `/api/v1/admin/redemption/codes` | POST | Create codes | ✅ v1.0.0 |
| `/api/v1/admin/redemption/codes/:id/disable` | POST | Disable code | ✅ v1.0.0 |
| `/api/v1/admin/redemption/codes/:id/usage` | GET | Code usage | ✅ v1.0.0 |
| `/api/v1/admin/model-groups` | GET | List model groups | ✅ v1.2.0 |
| `/api/v1/admin/model-groups/all` | GET | List all groups | ✅ v1.2.0 |
| `/api/v1/admin/model-groups` | POST | Create group | ✅ v1.2.0 |
| `/api/v1/admin/model-groups/:id` | PUT | Update group | ✅ v1.2.0 |
| `/api/v1/admin/model-groups/:id` | DELETE | Delete group | ✅ v1.2.0 |
| `/api/v1/admin/model-groups/:id/channels` | GET | Group channels | ✅ v1.2.0 |
| `/api/v1/admin/model-groups/:id/channels` | POST | Add channel | ✅ v1.2.0 |
| `/api/v1/admin/model-groups/:id/channels/:cid` | DELETE | Remove channel | ✅ v1.2.0 |
| `/api/v1/admin/model-pricing` | GET | List pricing | ✅ v1.2.0 |
| `/api/v1/admin/model-pricing/all` | GET | List all pricing | ✅ v1.2.0 |
| `/api/v1/admin/model-pricing/model/:model` | GET | Pricing by model | ✅ v1.2.0 |
| `/api/v1/admin/model-pricing` | POST | Create pricing | ✅ v1.2.0 |
| `/api/v1/admin/model-pricing/:id` | PUT | Update pricing | ✅ v1.2.0 |
| `/api/v1/admin/model-pricing/:id` | DELETE | Delete pricing | ✅ v1.2.0 |
| `/api/v1/admin/users/:id/groups` | GET | User groups | ✅ v1.2.0 |
| `/api/v1/admin/users/:id/groups` | PUT | Set user groups | ✅ v1.2.0 |

> **Settings 端点说明**: 原始设计为单一 `/api/v1/admin/settings` GET/PUT，v1.2.0 拆分为上述 7 个独立粒度端点。旧版单一端点不再存在。

### 3.4 Response Formats

**Success:**
```json
{
  "success": true,
  "data": { ... },
  "message": "Success"
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "Invalid email format"
  }
}
```

**Pagination:**
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

---

## 4. Security Implementation

### 4.1 JWT Token Structure

```go
type JWTPayload struct {
    UserID    uint   `json:"user_id"`
    TokenKey  string `json:"token_key"`
    ExpiresAt int64  `json:"expires_at"`
    IssuedAt  int64  `json:"issued_at"`
}

// Token header: Authorization: Bearer <token>
```

### 4.2 Sensitive Data Masking

| Field | Display | Storage | Example |
|-------|---------|---------|---------|
| Password | `***` | bcrypt hash | `bcrypt:$2a$10$...` |
| Token | `gsk-***abc` | Full | `gsk-xxxabc123` |
| API Key | `sk-***xyz` | Full | `sk-xxxxyz` |
| Credit Card | `****1234` | Encrypted | AES-256 |
| Phone | `138****5678` | Encrypted | AES-256 |
| Email | `a***@example.com` | Full | `abc@example.com` |

### 4.3 Rate Limiting Rules

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/api/v1/user/register` | 5 | minute |
| `/api/v1/user/login` | 10 | minute |
| `/api/v1/chat/completions` | 60 | minute (token-based) |
| `/api/v1/embeddings` | 60 | minute (token-based) |

### 4.4 Admin Access Control

- Separate admin login page: `/admin/login`
- Admin routes: `/api/v1/admin/*`
- Require `X-Admin-Secret` header for all admin endpoints
- Or use separate admin JWT with role-based access

---

## 5. Payment Integration

### 5.1 Alipay Integration

```
1. Create order → Generate Alipay payment URL
2. Redirect user to Alipay
3. User pays → Alipay sends async callback
4. Verify signature → Update order status
5. Grant quota/VIP to user
```

**Callback Fields:** trade_no, out_trade_no, trade_status, total_amount

### 5.2 WeChat Pay Integration

```
1. Create order → Generate WeChat payment QR code
2. Show QR code to user
3. User scans → WeChat sends async callback
4. Verify signature → Update order status
5. Grant quota/VIP to user
```

**Callback Fields:** transaction_id, out_trade_no, result_code, total_fee

### 5.3 Order Status Flow

```
PENDING → (payment success) → PAID → (admin process) → COMPLETED
                        ↓
                   FAILED/CANCELLED
```

---

## 6. VIP System

### 6.1 VIP Features

| Feature | Free | VIP |
|---------|------|-----|
| Daily quota | 10 | 100 |
| Priority queue | No | Yes |
| Max concurrent | 2 | 10 |
| Model access | Basic | All |
| Support | Community | Priority |
| Expires | Never | 30 days |

### 6.2 VIP Purchase Flow

```
1. User purchases VIP product
2. Payment success callback
3. Create/update VIP subscription
4. Set expiry = now + 30 days
5. Notify user
```

---

## 7. Channel Testing (Southbound)

### 7.1 Test Types

| Test | Purpose | Timeout |
|------|---------|---------|
| Models test | List available models | 10s |
| Chat test | Send test message | 30s |
| Embeddings test | Generate embeddings | 30s |
| Balance test | Check account balance | 10s |

### 7.2 Health Check Schedule

- Every 5 minutes: Check all active channels
- On-demand: Admin can manually test
- Results stored in `channel_test_results` table

---

## 8. Logging & Audit

### 8.1 Log Levels

| Level | Color | Stored | Display |
|-------|-------|--------|---------|
| DEBUG | Gray | Optional | Console |
| INFO | Green | Database | Console + File |
| WARN | Yellow | Database | Console + File |
| ERROR | Red | Database | Console + File + Alert |
| FATAL | Red+ | Database | Console + File + Alert |

### 8.2 Admin Log Viewer Statistics

```json
{
  "total": 1000,
  "by_level": {
    "INFO": 800,
    "WARN": 150,
    "ERROR": 50
  },
  "ratio": {
    "error_rate": "5%",
    "warn_rate": "15%"
  },
  "trends": [
    {"date": "2026-03-24", "error": 5, "warn": 15}
  ]
}
```

### 8.3 Sensitive Action Logging

| Action | Logged | Masked Fields |
|--------|--------|---------------|
| User login | Yes | password |
| Password change | Yes | old_password, new_password |
| Token create | Yes | token (partial) |
| Payment | Yes | card_number, cvv |
| API call | Yes | api_key (partial) |

---

## 9. Edge Cases & Known Issues

### 9.1 Payment Edge Cases

| Scenario | Handling |
|----------|----------|
| Double payment | Check order by out_trade_no, reject duplicates |
| Payment timeout | Auto-cancel after 30 min, release quota hold |
| Partial refund | Create refund order, adjust quota |
| Currency mismatch | Use USD as base, convert at checkout |

### 9.2 Channel Edge Cases

| Scenario | Handling |
|----------|----------|
| Channel timeout | Retry 3 times, then mark as unhealthy |
| Channel response error | Log error, return fallback response |
| No available channel | Return 503 Service Unavailable |
| Rate limit hit | Switch to next channel (if available) |

### 9.3 User Edge Cases

| Scenario | Handling |
|----------|----------|
| Token leaked | User can revoke, regenerate |
| Quota exhausted | Block API calls, show upgrade prompt |
| Account suspended | Block all access, show reason |
| Concurrent login | Allow multiple sessions, track in Redis |

---

## 10. Implementation Checklist

> 全部已完成（v1.0.0 ~ v1.3.0）。仅列出未完成项。

### 10.1 未完成

- [ ] 支付集成 — 微信支付（当前决定：暂不接入）
- [ ] Phase 4 压测回归（`api-key-lifecycle-test-plan.md`，推迟到后续 Sprint）
- [x] 压测自动化断言（`test_api.py` 已加入 assert_eq/assert_true + sys.exit 退出码）

### 10.2 已完成摘要

<details>
<summary>后端 (Go)</summary>

- ✅ Config / DB / Redis / RabbitMQ / JWT / Rate Limit / Admin Auth
- ✅ User CRUD / Token / Channel / Product / Order / Payment (Alipay)
- ✅ VIP / Quota / API Proxy / Health Check / Logging / Init Wizard
- ✅ Anthropic Messages API / SSE Streaming / Provider Adapters (6)
- ✅ 兑换码 / 模型管理 / 邮箱验证 / 滑块验证码 / 幂等性
</details>

<details>
<summary>前端 (Vue 3)</summary>

- ✅ 用户端：注册/登录/Profile/Token/商品/订单/VIP/日志/Dashboard
- ✅ 管理后台：用户/渠道/商品/订单/日志/设置/兑换码/模型管理/Dashboard
</details>

---

## 11. Reference Resources

### 11.1 Open Source References

- **OneAPI**: https://github.com/songquanpeng/one-api
  - Initialization flow
  - Channel management
  - Token-based API proxy
  
- **NewAPI**: https://github.com/QuantumNous/new-api
  - Multi-tenant design
  - Payment integration patterns
  
- **Laisky/one-api**: https://github.com/Laisky/one-api
  - Additional implementation patterns

### 11.2 Documentation Links

- PostgreSQL: https://www.postgresql.org/docs/
- Redis: https://redis.io/docs/
- RabbitMQ: https://www.rabbitmq.com/docs/
- Gin: https://gin-gonic.com/docs/
- Vue 3: https://vuejs.org/guide/
- Element Plus: https://element-plus.org/

---

## 12. Project Status

### 12.1 版本历史

| 版本 | 日期 | 关键变更 |
|------|------|---------|
| v1.0.0 | 2026-03-23 | 核心功能：多租户/VIP/支付/渠道/审计/兑换码 |
| v1.1.0 | 2026-04-13 | SOCKS 代理 / 测试历史 / 健康检测 |
| v1.1.1 | 2026-05-05 | 注册配置增强 / GORM 修复 / 审计日志补全 |
| v1.2.0 | 2026-05-12 | CI/CD / 设置页 / 模型管理前端 / 审计增强 |
| v1.2.1 | 2026-05-12 | CORS / AutoMigrate / 死代码清理 |
| v1.3.0 | 2026-05-14 | Anthropic Messages API / SSE 流式优化 / K01~K13 修复 |

### 12.2 Running Services

| Service | Port | Command |
|---------|------|---------|
| Consumer Frontend | 5173 | `npm run dev` |
| Admin Frontend | 5174 | `npm run preview:admin` |
| Go Backend | 8080 | `./gapi-server -config config.yaml` |

### 12.3 待办事项

- [ ] 微信支付接入（当前决定：暂不接入）
- [ ] Phase 4 压测回归（`api-key-lifecycle-test-plan.md`，推迟到后续 Sprint）
- [ ] #016 全面 Bug 审查修复（3 Critical / 10 High / 10 Medium / 13 Low，详见 [docs/issues/016-comprehensive-bug-review.md](../issues/016-comprehensive-bug-review.md)）
- [x] Phase 2/3 单元测试补充（BillingService 级联扣减 + TokenRateLimit 隔离 + Anthropic 端点）
- [x] 压测自动化断言（`test_api.py` assert_eq/assert_true + sys.exit 退出码）

### 12.4 部署文件

- `deploy/docker/docker-compose.yml` — 开发环境
- `deploy/docker/docker-compose.prod.yml` — 生产环境
- `deploy/docker/nginx/nginx.conf` — Nginx SSL + 反向代理（Docker 生产）
- `deploy/nginx/gapi-platform.conf` — Nginx 裸机部署配置
- `deploy/nginx/deploy-nginx.sh` — 裸机部署脚本
- `deploy/release/` — Release 打包模板

---

## 13. Platform-Specific Notes

### 13.1 Windows 兼容性

> 项目主要开发/部署平台为 Debian Linux (Docker)。Windows 作为**发布目标**（交叉编译产物提供 `.exe`），不推荐裸机运行。

**已知 Windows 限制**：

| 限制 | 说明 | 影响 |
|------|------|------|
| SIGTERM 不工作 | `signal.Notify(quit, syscall.SIGTERM)` 在 Windows 上静默忽略 | `taskkill /PID` 无法优雅关闭，只能用 Ctrl+C |
| CRLF 行尾 | SSEReader 只处理 `\n`，不处理 `\r\n` | Windows 本地 AI 服务器流式响应可能中断 |
| 无裸机部署指南 | 所有部署文档基于 Docker/Linux | Windows 裸机运行需用户自行配置 PostgreSQL/Redis/RabbitMQ |

**Windows 上推荐运行方式**：
1. **Docker Desktop**：使用 `deploy/docker/docker-compose.yml` 部署所有服务
2. **WSL2**：在 WSL2 内按 Linux 方式运行
3. **仅后端交叉编译**：Go 后端交叉编译为 `.exe` 配合外部 PostgreSQL/Redis

**Go 后端 Windows 兼容性良好**：
- 无 `os.Chmod`、`os/exec`、`filepath` 拼接硬编码路径
- GORM 表名均为小写 + 显式 `TableName()`，无大小写问题
- `os.TempDir` 等跨平台 API 使用正确

### 13.2 Debian/Linux

- 主要开发平台，Docker Compose 一键部署
- 生产建议使用 `deploy/docker/docker-compose.prod.yml`
- 裸机部署参考 `deploy/nginx/deploy-nginx.sh`（Nginx 反向代理）

### 13.3 已知部署 Bug

以下部署配置问题在 #016 审查中发现，尚未修复：

| # | 问题 | 影响 |
|---|------|------|
| H-05 | `deploy/release/docker-compose.yml` JWT/ENCRYPT_KEY 无默认值 | `.env` 缺失时密钥为空 |
| H-06 | `deploy/release/docker-compose.yml` Redis 密码不可配 | 改 `REDIS_PASSWORD` 不生效 |
| H-07 | CI `release-readme.txt` 不存在 | release 流水线崩溃 |
| H-08 | `deploy/docker/docker-compose.yml` config 卷源不存在 | 后端启动失败 |
| H-09 | `deploy/release/Dockerfile.backend` HEALTHCHECK 缺 curl | 容器健康检查失败 |
| H-10 | `deploy/nginx/deploy-nginx.sh` 调用未定义函数 | 裸机 nginx 部署脚本崩溃 |

---
	
*Document Version: 1.5*
*Last Updated: 2026-05-15*