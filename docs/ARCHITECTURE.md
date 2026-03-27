# gAPI Platform 开发设计文档

> API 代理平台，支持多渠道 AI 模型接入，多租户架构

**文档版本**: v1.6  
**更新日期**: 2026-03-27  
**状态**: 开发中 (已完成：商品管理、操作日志、管理后台布局修复)

> 配套文档：[业务设计文档](./BUSINESS.md)

---

## 1. 系统架构

### 1.1 整体架构

```
┌──────────────────────────────────────────────────────────────────────┐
│                         前端层 (Frontend)                            │
├─────────────────────────────┬────────────────────────────────────────┤
│   消费者前端 (5173)         │   管理后台 (5174)                       │
│   http://IP:5173           │   http://IP:5174/admin.html            │
│   - 用户注册/登录           │   - 仪表盘                             │
│   - API 密钥管理           │   - 用户管理                            │
│   - 充值/购买VIP           │   - 渠道管理                            │
│   - AI 接口调用            │   - 订单管理                            │
│   - 个人中心               │   - 操作日志                            │
└─────────────────────────────┴────────────────────────────────────────┘
                                  │
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         统一 API 网关层 (8080)                        │
├──────────────────────────────────────────────────────────────────────┤
│  /api/v1/user/*    → 用户接口   (Bearer Token 认证)               │
│  /api/v1/admin/*   → 管理接口   (账号密码认证)                     │
│  /api/v1/email/*   → 邮件验证   (无认证)                           │
│  /api/v1/captcha/* → 验证码     (无认证)                           │
└──────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         服务层 (Service)                              │
├─────────────┬─────────────┬─────────────┬───────────────────────────┤
│ UserService │ TokenService│OrderService │ ChannelService             │
│ 配额服务    │ 计费服务    │ AI路由服务  │ 健康检查服务               │
└─────────────┴─────────────┴─────────────┴───────────────────────────┘
                                  │
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         数据层                                        │
├─────────────────┬─────────────────┬─────────────────────────────────┤
│  PostgreSQL     │     Redis       │       RabbitMQ                   │
│ - 用户数据       │ - 会话缓存       │ - 异步任务                      │
│ - 渠道配置       │ - 限流计数       │ - 邮件队列                      │
│ - 订单记录       │ - Token缓存     │ - 日志队列                      │
│ - 配额记录       │ - 渠道健康状态   │ - VIP过期通知                   │
└─────────────────┴─────────────────┴─────────────────────────────────┘
```

### 1.2 技术栈

| 层级 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 前端 | Vue 3 + TypeScript | ^3.4 | 用户界面 |
| 前端 | Vite | ^6.0 | 构建工具 |
| 前端 | Element Plus | ^2.5 | UI 组件库 |
| 前端 | Pinia | ^2.1 | 状态管理 |
| 后端 | Go | 1.24 | API 服务 |
| 后端 | Gin | ^1.10 | HTTP 框架 |
| 数据库 | PostgreSQL | 15+ | 主数据存储 |
| 缓存 | Redis | 7+ | 会话/缓存 |
| 消息队列 | RabbitMQ | 3.12 | 异步任务 |

### 1.3 端口规划

| 端口 | 服务 | 说明 |
|------|------|------|
| 5173 | 消费者前端 | 普通用户访问 |
| 5174 | 管理后台前端 | 管理员访问 |
| 8080 | 统一 API | 用户+管理接口 |
| 5432 | PostgreSQL | 数据库 |
| 6379 | Redis | 缓存 |
| 5672 | RabbitMQ | 消息队列 |
| 15672 | RabbitMQ Admin | 管理界面 |

### 1.4 测试账号

#### 管理员账号 (config.yaml 配置)

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | super_admin |
| operator | operator123 | operator |

#### API 认证方式

- **用户接口**: Bearer Token (登录后获取)
- **管理接口**: 登录接口返回 token

---

## 2. 数据库设计

### 2.1 ER 关系图

```
┌──────────┐       ┌──────────┐       ┌──────────┐
│  Tenant  │       │   User   │       │  Channel  │
│  租户    │──1:N──│  用户    │──N:N──│  渠道    │
└──────────┘       └──────────┘       └──────────┘
                          │                   │
                          │ N:1               │ 1:N
                          ▼                   ▼
                   ┌──────────┐       ┌──────────┐
                   │  Order   │       │ API Token │
                   │  订单    │       │   密钥    │
                   └──────────┘       └──────────┘
                          │
                          │ N:1
                          ▼
                   ┌──────────┐
                   │ Product  │
                   │  商品    │
                   └──────────┘
```

### 2.2 核心表结构

#### tenants (租户表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| name | VARCHAR(100) | 租户名称 |
| code | VARCHAR(50) | 租户代码(唯一) |
| max_users | INT | 最大用户数 |
| max_channels | INT | 最大渠道数 |
| max_tokens | BIGINT | 最大Token配额 |
| features | JSONB | 功能开关 |
| status | SMALLINT | 状态:0禁用/1启用 |

#### users (用户表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| tenant_id | BIGINT | 租户ID |
| username | VARCHAR(50) | 用户名 |
| email | VARCHAR(255) | 邮箱(唯一) |
| phone | VARCHAR(20) | 手机号 |
| password_hash | VARCHAR(255) | 密码哈希 |
| email_verified | BOOLEAN | 邮箱已验证 |
| verify_token | VARCHAR(64) | 验证Token |
| verify_expired | TIMESTAMP | Token过期时间 |
| level | SMALLINT | 用户等级 |
| vip_expired_at | TIMESTAMP | VIP过期时间 |
| vip_package_id | BIGINT | VIP套餐ID |
| remain_quota | BIGINT | 剩余配额 |
| vip_quota | BIGINT | VIP配额 |
| status | SMALLINT | 状态 |

#### channels (渠道表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| tenant_id | BIGINT | 租户ID |
| name | VARCHAR(100) | 渠道名称 |
| type | VARCHAR(20) | 类型:openai/nvidia/azure... |
| base_url | TEXT | API地址 |
| api_key_encrypted | TEXT | 加密的API密钥 |
| key_version | INT | 密钥版本 |
| models | JSONB | 支持的模型列表 |
| model_mapping | JSONB | 模型映射 |
| weight | INT | 权重(加权随机) |
| priority | INT | 优先级 |
| rpm_limit | INT | 每分钟请求限制 |
| tpm_limit | INT | 每分钟Token限制 |
| cost_factor | DECIMAL | 成本系数 |
| price_per_1k_input | DECIMAL | 输入价格 |
| price_per_1k_output | DECIMAL | 输出价格 |
| group_name | VARCHAR(50) | 分组名称 |
| status | SMALLINT | 状态 |
| is_healthy | BOOLEAN | 健康状态 |
| failure_count | INT | 连续失败次数 |

#### api_tokens (API密钥表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| user_id | BIGINT | 用户ID |
| token | VARCHAR(64) | 密钥(唯一) |
| name | VARCHAR(100) | 密钥名称 |
| quota_limit | BIGINT | 配额上限 |
| quota_used | BIGINT | 已用配额 |
| rpm_limit | INT | RPM限制 |
| status | SMALLINT | 状态 |
| last_used_at | TIMESTAMP | 最后使用时间 |

#### orders (订单表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| order_no | VARCHAR(32) | 订单号(唯一) |
| user_id | BIGINT | 用户ID |
| order_type | VARCHAR(20) | 类型:recharge/vip/package |
| package_id | BIGINT | 商品ID |
| total_amount | DECIMAL | 总额 |
| discount_amount | DECIMAL | 优惠金额 |
| pay_amount | DECIMAL | 实付金额 |
| status | VARCHAR(20) | 状态 |
| paid_at | TIMESTAMP | 支付时间 |
| expire_at | TIMESTAMP | 过期时间 |

#### products (商品表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| name | VARCHAR(100) | 商品名称 |
| product_type | VARCHAR(20) | 类型:recharge/vip/package |
| price | DECIMAL | 价格 |
| quota | BIGINT | 配额 |
| vip_days | INT | VIP天数 |
| sort_order | INT | 排序 |
| is_recommended | BOOLEAN | 推荐 |
| status | VARCHAR(20) | 状态 |

#### quota_transactions (配额流水表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| user_id | BIGINT | 用户ID |
| type | VARCHAR(20) | 类型:recharge/deduct/expire/refund |
| amount | BIGINT | 变动金额 |
| balance | BIGINT | 变动后余额 |
| source | VARCHAR(20) | 来源 |
| order_id | BIGINT | 关联订单 |
| description | TEXT | 描述 |

#### admin_users (管理员表)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| username | VARCHAR(50) | 用户名(唯一) |
| password_hash | VARCHAR(255) | 密码哈希 |
| email | VARCHAR(255) | 邮箱 |
| role | VARCHAR(20) | 角色 |
| status | SMALLINT | 状态 |

---

## 3. API 设计

### 3.1 用户接口 (/api/v1/user/*)

#### 认证
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /register | 用户注册 | 无 |
| POST | /login | 用户登录 | 无 |
| POST | /logout | 退出登录 | Bearer |
| GET | /verify-email | 邮箱验证 | Token |
| POST | /send-verify-email | 发送验证邮件 | Bearer |

#### 用户
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /info | 获取用户信息 | Bearer |
| PUT | /info | 更新用户信息 | Bearer |
| PUT | /password | 修改密码 | Bearer |

#### Token (API密钥)
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /tokens | 获取密钥列表 | Bearer |
| POST | /tokens | 创建密钥 | Bearer |
| DELETE | /tokens/:id | 删除密钥 | Bearer |

#### 订单
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /products | 商品列表 | Bearer |
| POST | /orders | 创建订单 | Bearer |
| GET | /orders | 订单列表 | Bearer |

#### 配额
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /quota | 配额信息 | Bearer |
| GET | /quota/transactions | 配额流水 | Bearer |

### 3.2 AI代理接口 (/api/v1/*)

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /chat/completions | ChatGPT兼容接口 | Bearer |
| POST | /completions | 文本补全 | Bearer |
| POST | /embeddings | 向量嵌入 | Bearer |
| GET | /models | 模型列表 | Bearer |

### 3.3 管理接口 (/api/v1/admin/*)

#### 认证
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /login | 管理员登录 (用户名+密码) | 无 |
| POST | /change-password | 修改密码 | Bearer Token |

**登录流程**:
1. 管理员访问 `http://IP:5174/admin.html`
2. 输入用户名/密码 (配置于 `backend/config/config.yaml` 的 `admin_users`)
3. 后端验证后返回 Bearer Token
4. 前端存储 Token，跳转管理后台

**注意**: 管理后台是独立入口，使用 `admin.html`，不是 `/admin/login` 路径

#### 用户管理
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /users | 用户列表 | Bearer Token |
| PUT | /users/:id | 更新用户 | Bearer Token |
| PUT | /users/:id/quota | 调整配额 | Bearer Token |

#### 渠道管理
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /channels | 渠道列表 | Bearer Token |
| POST | /channels | 创建渠道 | Bearer Token |
| PUT | /channels/:id | 更新渠道 | Bearer Token |
| DELETE | /channels/:id | 删除渠道 | Bearer Token |
| POST | /channels/:id/test | 测试渠道 | Bearer Token |
| POST | /channels/:id/enable | 启用渠道 | Bearer Token |
| POST | /channels/:id/disable | 禁用渠道 | Bearer Token |

#### 订单管理
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /orders | 订单列表 | Bearer Token |
| PUT | /orders/:id/status | 更新状态 | Bearer Token |

#### 统计
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /stats/overview | 概览统计 | Bearer Token |

#### 日志
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /logs/operation | 操作日志 | Bearer Token |

---

## 4. 前端架构

### 4.1 目录结构

```
frontend/
├── src/
│   ├── main.ts              # 消费者前端入口
│   ├── admin-main.ts        # 管理后台入口
│   ├── App.vue              # 消费者应用组件
│   ├── admin-app.vue        # 管理后台组件
│   ├── admin-router.ts      # 管理后台路由
│   ├── router/
│   │   └── index.ts         # 用户路由
│   ├── views/
│   │   ├── Login.vue        # 用户登录页
│   │   ├── Register.vue     # 用户注册页
│   │   ├── Dashboard.vue    # 用户控制台
│   │   ├── Profile.vue      # 个人中心
│   │   ├── tokens/          # API密钥
│   │   ├── orders/          # 订单
│   │   ├── products/        # 商品
│   │   ├── vip/             # VIP
│   │   └── admin/           # 管理后台
│   │       ├── Login.vue    # 管理员登录页
│   │       ├── Dashboard.vue # 仪表盘
│   │       ├── users/       # 用户管理
│   │       ├── channels/    # 渠道管理
│   │       ├── orders/      # 订单管理
│   │       ├── logs/        # 操作日志
│   │       └── settings/    # 系统设置
│   ├── api/                 # API请求封装
│   ├── store/               # Pinia状态管理
│   └── style.css           # 全局样式
├── index.html               # 消费者入口HTML
├── admin.html               # 管理后台HTML
├── vite.config.ts          # 消费者前端Vite配置
└── vite.admin.config.ts    # 管理后台Vite配置
```

### 4.2 多入口配置

```typescript
// vite.config.ts - 消费者前端
export default defineConfig({
  server: {
    port: 5173
  },
  build: {
    rollupOptions: {
      input: './index.html'
    }
  }
})

// vite.admin.config.ts - 管理后台
export default defineConfig({
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist-admin',
    rollupOptions: {
      input: './admin.html'
    }
  }
})
```

### 4.3 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 消费者前端 | http://IP:5173 | 用户注册、登录、充值等 |
| 管理后台 | http://IP:5174/admin.html | 管理员登录和后台管理 |

### 4.4 路由守卫

```typescript
// 用户路由 - 需要登录
{ path: '/', component: Dashboard, meta: { requiresAuth: true } }

// 管理路由 - 需要管理员权限
{ path: '/admin', component: AdminDashboard, meta: { requiresAdmin: true } }
```

---

## 5. 安全设计

### 5.1 认证机制

| 类型 | 说明 | 使用场景 |
|------|------|----------|
| Bearer Token | JWT Token | 用户接口 + 管理接口认证 |
| API Key | 用户Token | AI调用认证 |

**管理接口认证流程**:
1. 管理员登录: `POST /api/v1/admin/login` (用户名+密码)
2. 获取 JWT Token
3. 后续请求携带: `Authorization: Bearer <token>`

### 5.2 密码安全

- 密码使用 bcrypt 加密 (cost=12)
- 密码修改后旧的 Token 失效
- 连续登录失败5次锁定15分钟

### 5.3 邮箱验证

- 注册时生成验证 Token (24小时有效)
- Token 通过邮件链接发送
- 点击链接验证邮箱

### 5.4 API 限流

| 接口 | 限制 | 备注 |
|------|------|------|
| /register | 5次/小时/IP | 防刷注册 |
| /login | 10次/小时/IP | 防暴力破解 |
| /chat/completions | 按用户RPM | 渠道限制叠加 |

---

## 6. AI 渠道代理

### 6.1 支持的渠道类型

| 类型 | 标识 | 说明 |
|------|------|------|
| OpenAI | openai | OpenAI API |
| NVIDIA NIM | nvidia | NVIDIA NIM |
| Azure OpenAI | azure | Azure OpenAI |
| Claude | claude | Anthropic Claude |
| Gemini | gemini | Google Gemini |
| DeepSeek | deepseek | DeepSeek |
| 智谱 ChatGLM | zhipu | 智谱AI |
| 百度千帆 | baidu | 百度云 |
| Groq | groq | Groq |
| Ollama | ollama | 本地Ollama |

### 6.2 渠道选择算法

```go
// 加权随机 + 健康检查
func SelectChannel(channels []*Channel) *Channel {
    // 1. 过滤不健康的渠道
    // 2. 按权重加权随机
    // 3. 跳过超出RPM/TPM限制的渠道
    return selected
}
```

### 6.3 请求转发

```
用户请求 → API网关 → Token验证 → 配额检查 → 渠道选择 → 请求转发 → 响应返回
```

---

## 7. 消息队列任务

### 7.1 队列列表

| 队列名 | 说明 | 消费者 |
|--------|------|--------|
| usage.log | 用量日志 | 记录用量统计 |
| email.send | 发送邮件 | SMTP发送 |
| order.payment | 订单支付 | 支付回调处理 |
| vip.expire | VIP过期 | 过期通知+状态更新 |
| health.check | 健康检查 | 渠道健康检测 |

### 7.2 邮件模板

- 注册验证邮件
- VIP过期提醒
- 配额不足提醒
- 订单支付成功

---

## 8. 待完成功能

### 8.1 已完成 ✅

- [x] 用户注册/登录
- [x] Token管理
- [x] 渠道CRUD
- [x] 订单管理
- [x] 配额系统
- [x] AI代理转发
- [x] 多渠道支持
- [x] 管理后台
- [x] 前后端分离部署

### 8.2 进行中 🔄

- [x] 滑动验证码 (简化版：拖动滑块)
- [ ] 邮箱验证
- [ ] 支付集成

### 8.3 待开发 📋

- [ ] WebSocket流式响应
- [ ] 用量统计图表
- [ ] 国际化 (i18n)
- [ ] 多语言界面

---

## 9. 部署说明

### 9.1 技术栈声明

**本项目全程使用 Go + Node.js，不使用 Python。**

- 后端：Go + Gin
- 前端：Node.js + Vite + Vue 3
- 数据库：PostgreSQL
- 缓存：Redis
- 消息队列：RabbitMQ

### 9.2 开发环境

```bash
# 1. 启动数据库服务
systemctl start postgresql redis rabbitmq

# 2. 启动后端 (Go)
cd /root/gapi-platform/backend
go run cmd/server/main.go -config ./config/config.yaml

# 3. 启动消费者前端 (Node.js/Vite)
cd /root/gapi-platform/frontend
npm run dev

# 4. 启动管理后台 (Node.js/Vite + API代理)
cd /root/gapi-platform/frontend
npm run build:admin && npm run preview:admin

# 5. 访问地址
# 消费者前端: http://localhost:5173
# 管理后台:   http://localhost:5174/admin.html
```

**注意**：管理后台必须使用 `npm run preview:admin` 而非 Python 或其他静态服务器，因为需要 Vite 代理转发 API 请求到后端。

### 9.3 生产环境构建

```bash
# 构建消费者前端
cd frontend && npm run build

# 构建管理后台
cd frontend && npm run build:admin

# 构建全部
cd frontend && npm run build:all
```

### 9.4 服务管理

```bash
# 查看服务状态
ps aux | grep -E "gapi-server|vite"

# 重启后端
pkill -f gapi-server
cd backend && go run cmd/server/main.go -config ./config/config.yaml

# 重启消费者前端
pkill -f "vite"
cd frontend && npm run dev

# 重启管理后台 (必须使用 Vite 预览支持 API 代理)
pkill -f "vite"
cd frontend && npm run build:admin && npm run preview:admin
```

---

## 10. 环境变量

### 后端 (.env)

```bash
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=gapi
DB_PASSWORD=xxx
DB_NAME=gapi

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# JWT
JWT_SECRET=xxx
JWT_EXPIRE=720h

# 管理员密钥
ADMIN_SECRET=admin123

# 邮件
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASSWORD=xxx
```

### 前端

```bash
VITE_API_BASE=http://localhost:8080
```
