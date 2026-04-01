# SMTP 配置管理功能设计方案

**版本**: v1.0  
**日期**: 2026-03-30  
**状态**: 设计中

---

## 1. 需求背景

### 1.1 问题陈述
当前 SMTP 配置通过环境变量存储，存在以下问题：
- 安全性：配置信息明文存储在配置文件中
- 灵活性：修改配置需要重启服务
- 管理性：无法在管理后台统一管理

### 1.2 解决方案
在管理后台添加 SMTP 配置功能，将配置存储到数据库，并支持加密存储敏感信息。

---

## 2. 团队评审

### 2.1 产品经理 (PM)

**需求确认**：
| 需求项 | 确认 |
|--------|------|
| 在管理后台系统设置中添加邮箱设置 Tab | ✅ |
| 支持配置 SMTP 服务器、端口、账号密码 | ✅ |
| 支持测试连接功能 | ✅ |
| 敏感信息（密码）加密存储 | ✅ |
| 支持配置发送邮箱的显示名称 | ✅ |

**用户故事**：
> 作为系统管理员，我希望在管理后台配置邮箱设置，以便：
> - 启用用户注册邮箱验证功能
> - 启用忘记密码邮件发送功能
> - 不需要修改代码或重启服务即可更新配置

### 2.2 安全工程师

**安全审查**：

| 安全项 | 方案 | 状态 |
|--------|------|------|
| 密码加密存储 | 使用 AES-256-GCM 加密 | ✅ |
| API 访问控制 | 仅管理员可访问 | ✅ |
| 日志脱敏 | 敏感字段不记录日志 | ✅ |
| 密码不回显 | GET 接口不返回明文密码 | ✅ |
| 测试连接 | 发送测试邮件验证 | ✅ |

**威胁模型**：
- **敏感信息泄露**：密码使用 `IsSensitive=true`，GET 接口自动过滤
- **SQL 注入**：使用 GORM 参数化查询 ✅
- **权限提升**：检查管理员认证中间件 ✅

### 2.3 前端 UI/UX 工程师

**界面设计**：

```
系统设置页面
├── 基本设置
├── 注册设置
├── 速率限制
├── 支付设置
├── 安全设置
└── 邮箱设置 ← 新增
    ├── 启用邮箱服务 [开关]
    ├── SMTP 服务器 [输入框]
    ├── SMTP 端口 [数字输入] (默认 587)
    ├── 使用 TLS [开关]
    ├── 用户名/邮箱 [输入框]
    ├── 密码/授权码 [密码输入] (加密存储)
    ├── 发件人名称 [输入框]
    ├── 发件人邮箱 [输入框]
    ├── 发送测试邮件 [按钮]
    └── 保存 [按钮]
```

**UX 考量**：
- 密码输入使用 `show-password` 属性
- 保存后显示成功/失败提示
- 测试连接提供加载状态
- 敏感字段在表单中为空时不更新密码

### 2.4 后端开发工程师

**技术方案**：

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| 存储 | 复用 `SystemConfig` 模型 | 复用现有 key-value 存储 |
| 加密 | AES-256-GCM | 复用现有 crypto 包 |
| 缓存 | 内存缓存 | 避免频繁查询数据库 |
| API | RESTful | 遵循现有 API 风格 |

**配置项设计**：

| ConfigKey | ValueType | IsSensitive | 说明 |
|-----------|-----------|-------------|------|
| `smtp_enabled` | boolean | false | 启用开关 |
| `smtp_host` | string | false | SMTP 服务器地址 |
| `smtp_port` | number | false | 端口号 |
| `smtp_use_tls` | boolean | false | 是否使用 TLS |
| `smtp_username` | string | false | 用户名 |
| `smtp_password` | string | true | 密码 (加密) |
| `smtp_from_name` | string | false | 发件人名称 |
| `smtp_from_email` | string | false | 发件人邮箱 |

---

## 3. 数据库设计

### 3.1 复用 SystemConfig 表

```sql
-- 已有表结构，复用存储 SMTP 配置
INSERT INTO system_configs (config_key, config_value, value_type, config_group, description, is_sensitive)
VALUES
  ('smtp_enabled', 'false', 'boolean', 'email', '启用邮箱服务', false),
  ('smtp_host', '', 'string', 'email', 'SMTP 服务器地址', false),
  ('smtp_port', '587', 'number', 'email', 'SMTP 端口', false),
  ('smtp_use_tls', 'true', 'boolean', 'email', '使用 TLS 加密', false),
  ('smtp_username', '', 'string', 'email', 'SMTP 用户名', false),
  ('smtp_password', '', 'string', 'email', 'SMTP 密码 (加密存储)', true),
  ('smtp_from_name', 'gAPI Platform', 'string', 'email', '发件人名称', false),
  ('smtp_from_email', 'noreply@gapi.com', 'string', 'email', '发件人邮箱', false);
```

### 3.2 加密策略

```
明文密码 → AES-256-GCM 加密 → Base64 编码 → 存储到 ConfigValue

读取时：
ConfigValue (加密) → Base64 解码 → AES-256-GCM 解密 → 明文密码 → 用于 SMTP 认证
```

---

## 4. API 设计

### 4.1 后端 API

#### GET /api/v1/admin/settings/email
获取邮箱配置（密码返回空字符串）

**响应示例**：
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "host": "smtp.qq.com",
    "port": 587,
    "use_tls": true,
    "username": "123456@qq.com",
    "password": "",  // 不返回实际密码
    "from_name": "gAPI Platform",
    "from_email": "noreply@gapi.com"
  }
}
```

#### PUT /api/v1/admin/settings/email
更新邮箱配置

**请求示例**：
```json
{
  "enabled": true,
  "host": "smtp.qq.com",
  "port": 587,
  "use_tls": true,
  "username": "123456@qq.com",
  "password": "new-password-or-empty",  // 空字符串表示不更新
  "from_name": "gAPI Platform",
  "from_email": "noreply@gapi.com"
}
```

#### POST /api/v1/admin/settings/email/test
发送测试邮件

**请求示例**：
```json
{
  "test_email": "admin@example.com"
}
```

**响应示例**：
```json
{
  "success": true,
  "message": "测试邮件发送成功"
}
```

### 4.2 前端 API 模块

```typescript
// src/api/admin/settings.ts
export const emailSettingsApi = {
  get: () => axios.get('/admin/settings/email'),
  update: (data: EmailSettings) => axios.put('/admin/settings/email', data),
  test: (email: string) => axios.post('/admin/settings/email/test', { test_email: email }),
}
```

---

## 5. 组件设计

### 5.1 后端组件

```
backend/
├── internal/
│   ├── handler/
│   │   └── settings_handler.go    # 新增：邮箱配置 API
│   ├── service/
│   │   └── settings_service.go    # 新增：配置管理服务
│   └── repository/
│       └── settings_repo.go      # 新增：配置仓储
```

### 5.2 前端组件

```
frontend/src/
├── api/
│   └── admin/
│       └── settings.ts           # 新增：API 接口
├── views/admin/settings/
│   └── Index.vue               # 修改：添加邮箱设置 Tab
```

---

## 6. 实现步骤

### 6.1 后端实现

| 步骤 | 任务 | 优先级 |
|------|------|--------|
| 1 | 创建 `settings_service.go` - 配置管理服务 | P0 |
| 2 | 创建 `settings_handler.go` - 配置 API | P0 |
| 3 | 在 `router.go` 添加路由 | P0 |
| 4 | 修改 `email_mailer.go` 支持动态配置 | P0 |
| 5 | 添加配置缓存机制 | P1 |

### 6.2 前端实现

| 步骤 | 任务 | 优先级 |
|------|------|--------|
| 1 | 创建 API 接口模块 | P0 |
| 2 | 添加邮箱设置 Tab | P0 |
| 3 | 实现表单和数据绑定 | P0 |
| 4 | 实现测试连接功能 | P1 |

### 6.3 测试验证

| 步骤 | 任务 | 优先级 |
|------|------|--------|
| 1 | API 单元测试 | P0 |
| 2 | 前端功能测试 | P0 |
| 3 | 实际邮件发送测试 | P0 |

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 加密密钥丢失 | 无法解密密码 | 使用 `GAPI_ENCRYPT_KEY` 作为加密密钥，该密钥必须持久化 |
| 数据库迁移 | 现有配置丢失 | 提供数据迁移脚本 |
| SMTP 服务不可用 | 邮件发送失败 | 提供连接测试功能，记录错误日志 |

---

## 8. 配置优先级

```
环境变量 > 数据库配置 > 默认值

即：
1. 如果设置了环境变量，使用环境变量
2. 否则使用数据库配置
3. 最后使用默认值
```

这样保证：
- 容器启动时可以使用环境变量
- 运行时可以动态修改数据库配置
- 兼容现有部署方式

---

## 9. 验收标准

- [ ] 管理员可以在后台配置 SMTP 设置
- [ ] SMTP 密码加密存储到数据库
- [ ] API 不返回明文密码
- [ ] 可以发送测试邮件验证配置
- [ ] 邮件发送功能正常工作
- [ ] 单元测试覆盖关键逻辑

---

## 10. 后续规划

- [ ] 邮件模板管理界面
- [ ] 邮件发送日志
- [ ] 队列发送机制（大批量邮件）
- [ ] 多种邮件服务商支持（QQ/163/Gmail/企业邮箱）
