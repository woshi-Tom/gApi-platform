# gAPI Platform - Development Notes & Implementation Checklist

> Last Updated: 2026-03-27
> Purpose: Capture all detail issues and pending items for development session

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

---

## 3. API Implementation Details

### 3.1 Northbound API (User-Facing)

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/user/register` | POST | No | User registration |
| `/api/v1/user/login` | POST | No | User login |
| `/api/v1/user/profile` | GET | JWT | Get profile |
| `/api/v1/user/profile` | PUT | JWT | Update profile |
| `/api/v1/user/change-password` | POST | JWT | Change password |
| `/api/v1/tokens` | GET | JWT | List tokens |
| `/api/v1/tokens` | POST | JWT | Create token |
| `/api/v1/tokens/:id` | DELETE | JWT | Delete token |
| `/api/v1/products` | GET | No | List products |
| `/api/v1/products/:id` | GET | No | Product detail |
| `/api/v1/orders` | GET | JWT | List orders |
| `/api/v1/orders` | POST | JWT | Create order |
| `/api/v1/orders/:id` | GET | JWT | Order detail |
| `/api/v1/payment/alipay` | POST | JWT | Alipay payment |
| `/api/v1/payment/wechat` | POST | JWT | WeChat payment |
| `/api/v1/payment/callback/alipay` | POST | No | Alipay webhook |
| `/api/v1/payment/callback/wechat` | POST | No | WeChat webhook |
| `/api/v1/vip/status` | GET | JWT | VIP status |
| `/api/v1/quota` | GET | JWT | Quota info |
| `/api/v1/chat/completions` | POST | Token | OpenAI-compatible |
| `/api/v1/models` | GET | Token | List models |
| `/api/v1/embeddings` | POST | Token | Embeddings API |

### 3.2 Southbound API (Internal)

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/internal/channels` | GET | Internal | List channels |
| `/api/v1/internal/channels` | POST | Internal | Create channel |
| `/api/v1/internal/channels/:id` | PUT | Internal | Update channel |
| `/api/v1/internal/channels/:id` | DELETE | Internal | Delete channel |
| `/api/v1/internal/channels/:id/test` | POST | Internal | Test channel |
| `/api/v1/internal/models` | GET | Internal | List models |
| `/api/v1/internal/models/sync` | POST | Internal | Sync models |
| `/api/v1/internal/health` | GET | Internal | Health check |
| `/api/v1/internal/balance/:channel_id` | GET | Internal | Get balance |

### 3.3 Admin API (Intranet Only)

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/admin/login` | POST | No | Admin login |
| `/api/v1/admin/users` | GET | Admin | User list |
| `/api/v1/admin/users/:id` | PUT | Admin | Update user |
| `/api/v1/admin/users/:id/quota` | POST | Admin | Adjust quota |
| `/api/v1/admin/tokens` | GET | Admin | Token list |
| `/api/v1/admin/tokens/:id` | DELETE | Admin | Revoke token |
| `/api/v1/admin/products` | GET | Admin | Product list |
| `/api/v1/admin/products` | POST | Admin | Create product |
| `/api/v1/admin/products/:id` | PUT | Admin | Update product |
| `/api/v1/admin/products/:id/publish` | POST | Admin | Publish product |
| `/api/v1/admin/products/:id/unpublish` | POST | Admin | Unpublish product |
| `/api/v1/admin/orders` | GET | Admin | Order list |
| `/api/v1/admin/orders/:id` | GET | Admin | Order detail |
| `/api/v1/admin/orders/:id/process` | POST | Admin | Process order |
| `/api/v1/admin/channels` | GET | Admin | Channel list |
| `/api/v1/admin/channels` | POST | Admin | Create channel |
| `/api/v1/admin/channels/:id` | PUT | Admin | Update channel |
| `/api/v1/admin/channels/:id/test` | POST | Admin | Test channel |
| `/api/v1/admin/channels/batch-import` | POST | Admin | Batch import |
| `/api/v1/admin/vip-subscriptions` | GET | Admin | VIP list |
| `/api/v1/admin/vip-subscriptions/:id` | PUT | Admin | Update VIP |
| `/api/v1/admin/logs/operation` | GET | Admin | Operation logs |
| `/api/v1/admin/logs/api` | GET | Admin | API logs |
| `/api/v1/admin/logs/login` | GET | Admin | Login logs |
| `/api/v1/admin/logs/payment` | GET | Admin | Payment logs |
| `/api/v1/admin/logs/stats` | GET | Admin | Log statistics |
| `/api/v1/admin/settings` | GET | Admin | System settings |
| `/api/v1/admin/settings` | PUT | Admin | Update settings |
| `/api/v1/admin/stats/overview` | GET | Admin | Dashboard stats |

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

### 10.1 Backend (Go)

- [ ] Project initialization (go.mod, main.go)
- [ ] Config loading (environment variables)
- [ ] Database connection (GORM)
- [ ] Redis connection
- [ ] RabbitMQ connection
- [ ] Logging setup (zap + rotation)
- [ ] JWT middleware
- [ ] Rate limiting middleware
- [ ] Admin auth middleware
- [ ] User CRUD
- [ ] Token management
- [ ] Channel CRUD
- [ ] Model sync
- [ ] Product CRUD
- [ ] Order management
- [ ] Payment integration (Alipay, WeChat)
- [ ] VIP subscription
- [ ] Quota management
- [ ] API proxy (Chat, Embeddings)
- [ ] Channel testing
- [ ] Health check worker
- [ ] Operation logging
- [ ] API logging
- [ ] Login logging
- [ ] Admin APIs
- [ ] First-run initialization

### 10.2 Frontend (Vue 3)

- [ ] Project initialization
- [ ] Vue Router setup
- [ ] Pinia store
- [ ] Element Plus integration
- [ ] Login page (user)
- [ ] Register page
- [ ] Dashboard (user)
- [ ] Token management page
- [ ] Product list page
- [ ] Order history page
- [ ] VIP purchase page
- [ ] Profile page
- [ ] Admin login page
- [ ] Admin dashboard
- [ ] User management (admin)
- [ ] Product management (admin)
- [ ] Channel management (admin)
- [ ] Order management (admin)
- [ ] Log viewer (admin)
- [ ] Settings page (admin)
- [ ] Initialization wizard

### 10.3 Testing

- [ ] Unit tests (services)
- [ ] Integration tests (API endpoints)
- [ ] E2E tests (critical flows)
- [ ] Load testing (10K concurrent users)
- [ ] Security testing (penetration)

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

### 12.1 Completed Features (2026-03-27)
- ✅ Go 后端核心服务 (端口 8080)
- ✅ Vue 3 管理后台 (端口 5173/5174)
- ✅ PostgreSQL / Redis / RabbitMQ Docker 容器
- ✅ 用户注册/登录/Token 管理
- ✅ 渠道管理与健康检查
- ✅ 商品管理 (VIP套餐/充值套餐)
- ✅ 操作日志 (审计中间件 + 登录日志)
- ✅ 登录日志 API (`/api/v1/admin/logs/login`)
- ✅ 管理后台布局修复 (无刷新重复)
- ✅ 限流功能 (RPM/TPM 限制)
- ✅ 局域网访问支持 (绑定 0.0.0.0)

### 12.2 Running Services
| Service | Port | Command |
|---------|------|---------|
| Consumer Frontend | 5173 | `npm run dev` |
| Admin Frontend | 5174 | `npm run preview:admin` |
| Go Backend | 8080 | `./gapi-server -config config.yaml` |

### 12.3 Test Accounts
| Type | Username | Password |
|------|----------|----------|
| Admin | `admin` | `admin123` |
| User | `admin@example.com` | `admin123` |

### 12.4 API Endpoints (Implemented)
- `POST /api/v1/user/login` - 用户登录
- `POST /api/v1/user/register` - 用户注册
- `GET /api/v1/products` - 商品列表
- `GET /api/v1/admin/login` - 管理员登录
- `GET /api/v1/admin/logs/login` - 登录日志
- `GET /api/v1/admin/logs/operation` - 操作日志
- `GET /api/v1/admin/stats/overview` - 仪表盘统计
- `GET /api/v1/admin/users` - 用户列表
- `GET /api/v1/admin/channels` - 渠道列表
- `GET /api/v1/admin/orders` - 订单列表
- `GET /api/v1/admin/products` - 商品列表
- `POST /api/v1/admin/products` - 创建商品
- `PUT /api/v1/admin/products/:id` - 更新商品
- `POST /api/v1/admin/products/:id/enable` - 启用商品
- `POST /api/v1/admin/products/:id/disable` - 禁用商品

### 12.5 Remaining Tasks
- [ ] 支付集成 (支付宝/微信)
- [ ] 部署文档

### 12.6 Recently Fixed (2026-03-28)
- [x] 商品管理 RPM/TPM 限制显示 bug
- [x] 用户 API 密钥复制显示不完整
- [x] 用户 VIP 配额显示为 0
- [x] 用户控制台"最近活动"跳转错误

### 12.7 Recently Added (2026-03-28)
- [x] 用户 API 调用日志功能
  - 后端中间件记录 API 调用
  - 前端 /logs 页面查看 API 调用记录
  - Dashboard "查看全部" 链接到日志页面
- [x] VIP 过期处理后台任务
  - 每分钟检查过期 VIP 用户
  - 自动降级为 free 等级

### 12.6 Known Issues (Resolved)
- [x] 管理后台布局重复 - 已修复
- [x] 局域网访问 502 - 已修复
- [x] 登录日志 API 404 - 已添加
- [x] 数据库 jsonb 类型错误 - 已修复为 text

---

*Document Version: 1.1*
*Last Updated: 2026-03-27*