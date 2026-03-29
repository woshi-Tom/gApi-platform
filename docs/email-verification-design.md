# 邮箱验证码与忘记密码功能设计文档

**版本**: 1.0  
**日期**: 2026-03-30  
**状态**: 待实现

---

## 1. 功能概述

本设计文档详细说明两个功能的实现方案：

1. **邮箱验证码注册** - 用户注册时验证邮箱有效性
2. **忘记密码** - 用户通过邮箱重置密码

### 1.1 业务流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         邮箱验证码注册流程                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. 用户输入: 用户名、邮箱、密码                                            │
│     │                                                                      │
│     ▼                                                                      │
│  2. 前端: 校验邮箱格式 → 显示滑动验证                                      │
│     │                                                                      │
│     ▼                                                                      │
│  3. 后端: 校验验证码发送限制 (邮箱/IP/设备)                               │
│     │                                                                      │
│     ▼                                                                      │
│  4. 生成6位数字验证码 → 存储(哈希) → 发送邮件                             │
│     │                                                                      │
│     ▼                                                                      │
│  5. 用户输入验证码 → 后端验证(单次使用)                                     │
│     │                                                                      │
│     ▼                                                                      │
│  6. 创建用户账户 → 赠送配额 → 注册成功                                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         忘记密码流程                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. 用户输入邮箱 → 显示滑动验证                                            │
│     │                                                                      │
│     ▼                                                                      │
│  2. 后端: 校验发送限制 → 生成重置Token → 发送重置链接邮件                   │
│     │                                                                      │
│     ▼                                                                      │
│  3. 用户点击邮件链接 → 进入重置密码页面                                     │
│     │                                                                      │
│     ▼                                                                      │
│  4. 输入新密码 → 验证Token有效性 → 更新密码 → 完成                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. 安全设计

### 2.1 验证码安全参数

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| **验证码长度** | 6位数字 | 足够_entropy，短而快 |
| **有效期** | 10分钟 | 足够填写，避免过长 |
| **最大尝试次数** | 3次 | 防止暴力破解 |
| **发送冷却** | 60秒 | 防止刷验证码 |
| **每邮箱/小时** | 5次 | 合理限制 |
| **每IP/小时** | 10次 | 防止IP滥用 |
| **每设备/小时** | 3次 | 防止设备滥用 |

### 2.2 重置Token安全参数

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| **Token长度** | 32字节随机 | 128位熵，UUID4级别 |
| **有效期** | 1小时 | OWASP推荐 |
| **单次使用** | 是 | 使用后立即失效 |
| **每邮箱/小时** | 3次 | 更严格的限制 |

### 2.3 防攻击措施

| 防护类型 | 措施 | 说明 |
|----------|------|------|
| **防暴力破解** | 验证码3次尝试后失效 | 单次使用+过期 |
| **防枚举攻击** | 统一返回"验证码已发送" | 不暴露邮箱是否注册 |
| **防刷验证码** | 多维度限流 | 邮箱+IP+设备三重限制 |
| **防CSRF** | Token验证 | 重置链接含一次性Token |
| **防钓鱼** | HTTPS强制 | 邮件链接使用HTTPS |

---

## 3. 数据库设计

### 3.1 邮箱验证码表 (已有，需优化)

```sql
CREATE TABLE email_verifications (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(100) NOT NULL,
    code_hash       VARCHAR(64) NOT NULL,           -- 验证码哈希存储
    purpose         VARCHAR(20) NOT NULL DEFAULT 'register',  -- register|reset
    ip_address      VARCHAR(50),
    user_agent      VARCHAR(500),
    device_hash     VARCHAR(64),
    is_used         BOOLEAN DEFAULT FALSE,
    used_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    attempt_count   INTEGER DEFAULT 0,             -- 尝试次数
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_email ON email_verifications(email);
CREATE INDEX idx_email_verifications_token ON email_verifications(token);
CREATE INDEX idx_email_verifications_expires ON email_verifications(expires_at);
```

### 3.2 密码重置表 (新建)

```sql
CREATE TABLE password_resets (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(100) NOT NULL,
    token_hash      VARCHAR(64) NOT NULL UNIQUE,   -- 重置Token哈希
    user_id         BIGINT REFERENCES users(id),
    ip_address      VARCHAR(50),
    is_used         BOOLEAN DEFAULT FALSE,
    used_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_resets_email ON password_resets(email);
CREATE INDEX idx_password_resets_token ON password_resets(token_hash);
CREATE INDEX idx_password_resets_expires ON password_resets(expires_at);
```

---

## 4. API 设计

### 4.1 发送验证码

```
POST /api/v1/email/send-code

Request:
{
    "email": "user@example.com",
    "purpose": "register",        // register | reset
    "captcha_token": "xxx"
}

Response (Success):
{
    "success": true,
    "message": "验证码已发送到您的邮箱",
    "data": {
        "expires_in": 600         // 10分钟有效期
    }
}

Response (Rate Limited):
{
    "success": false,
    "error": {
        "code": "RATE_LIMITED",
        "message": "验证码发送过于频繁，请稍后再试"
    }
}
```

### 4.2 验证注册验证码

```
POST /api/v1/email/verify-code

Request:
{
    "email": "user@example.com",
    "code": "123456",
    "purpose": "register"
}

Response (Success):
{
    "success": true,
    "message": "验证成功",
    "data": {
        "verification_token": "xxx"    // 验证通过的临时Token
    }
}
```

### 4.3 发送重置密码邮件

```
POST /api/v1/auth/forgot-password

Request:
{
    "email": "user@example.com",
    "captcha_token": "xxx"
}

Response:
{
    "success": true,
    "message": "如果该邮箱已注册，重置链接已发送"
}
```

### 4.4 验证重置Token

```
GET /api/v1/auth/reset-password?token=xxx

Response (Valid):
{
    "success": true,
    "data": {
        "email": "user@example.com",
        "expires_at": "2026-03-30T15:00:00Z"
    }
}

Response (Invalid/Expired):
{
    "success": false,
    "error": {
        "code": "INVALID_TOKEN",
        "message": "重置链接无效或已过期"
    }
}
```

### 4.5 重置密码

```
POST /api/v1/auth/reset-password

Request:
{
    "token": "xxx",
    "password": "newPassword123",
    "confirm_password": "newPassword123"
}

Response:
{
    "success": true,
    "message": "密码重置成功"
}
```

---

## 5. 邮件模板设计

### 5.1 注册验证码邮件

**主题**: `[gAPI] 您的注册验证码`

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>注册验证码</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #409eff 0%, #337ecc 100%); padding: 30px; text-align: center; border-radius: 12px 12px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">gAPI Platform</h1>
    </div>
    
    <div style="background: #fff; padding: 30px; border: 1px solid #e4e7ed; border-top: none; border-radius: 0 0 12px 12px;">
        <h2 style="color: #303133; margin-top: 0;">验证码</h2>
        
        <p style="color: #606266; font-size: 16px; line-height: 1.6;">
            您好，<br><br>
            您的注册验证码是：
        </p>
        
        <div style="background: #f5f7fa; padding: 20px; text-align: center; border-radius: 8px; margin: 20px 0;">
            <span style="font-size: 32px; font-weight: bold; color: #409eff; letter-spacing: 8px;">{{code}}</span>
        </div>
        
        <p style="color: #909399; font-size: 14px;">
            验证码有效期为 <strong>10分钟</strong>，请尽快完成验证。<br>
            如果您没有注册 gAPI 账号，请忽略此邮件。
        </p>
        
        <hr style="border: none; border-top: 1px solid #ebeef5; margin: 20px 0;">
        
        <p style="color: #c0c4cc; font-size: 12px;">
            此邮件由系统自动发送，请勿回复。<br>
            为保护您的账户安全，请勿将验证码告知他人。
        </p>
    </div>
</body>
</html>
```

### 5.2 密码重置邮件

**主题**: `[gAPI] 密码重置请求`

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>密码重置</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #67c23a 0%, #529b2e 100%); padding: 30px; text-align: center; border-radius: 12px 12px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">gAPI Platform</h1>
    </div>
    
    <div style="background: #fff; padding: 30px; border: 1px solid #e4e7ed; border-top: none; border-radius: 0 0 12px 12px;">
        <h2 style="color: #303133; margin-top: 0;">密码重置</h2>
        
        <p style="color: #606266; font-size: 16px; line-height: 1.6;">
            您好，<br><br>
            我们收到了您的密码重置请求。请点击下方按钮重置密码：
        </p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{reset_link}}" style="display: inline-block; background: linear-gradient(135deg, #67c23a 0%, #529b2e 100%); color: white; padding: 14px 40px; text-decoration: none; border-radius: 6px; font-weight: bold;">
                重置密码
            </a>
        </div>
        
        <p style="color: #909399; font-size: 14px;">
            此链接有效期为 <strong>1小时</strong>。<br>
            如果您没有请求重置密码，请忽略此邮件，您的账户安全不受影响。
        </p>
        
        <div style="background: #fdf6ec; border: 1px solid #f5dab1; border-radius: 6px; padding: 15px; margin: 20px 0;">
            <p style="color: #e6a23c; font-size: 14px; margin: 0;">
                <strong>安全提示：</strong>链接有效期结束后将自动失效，请妥善保管。
            </p>
        </div>
        
        <hr style="border: none; border-top: 1px solid #ebeef5; margin: 20px 0;">
        
        <p style="color: #c0c4cc; font-size: 12px;">
            此邮件由系统自动发送，请勿回复。<br>
            为保护您的账户安全，请勿将重置链接分享给他人。
        </p>
    </div>
</body>
</html>
```

---

## 6. 前端设计

### 6.1 忘记密码页面

**路由**: `/forgot-password`

**UI 流程**:
1. 输入邮箱 → 格式校验
2. 滑动验证
3. 发送重置邮件 → 显示成功提示
4. 引导查看邮箱

### 6.2 重置密码页面

**路由**: `/reset-password?token=xxx`

**前置条件**: Token 有效性验证

**UI 流程**:
1. 验证 Token → 显示邮箱信息
2. 输入新密码 + 确认密码
3. 密码强度提示
4. 提交 → 成功跳转登录页

### 6.3 注册流程优化

**当前流程** (已有):
```
邮箱输入 → 格式校验 → 滑动验证 → 发送验证码 → 输入验证码 → 注册
```

**优化点**:
- 验证码有效期倒计时显示
- 验证码错误后自动刷新滑块
- 注册成功后自动登录选项

---

## 7. 实现清单

### 7.1 后端实现

| 任务 | 优先级 | 状态 |
|------|--------|------|
| 验证码哈希存储 (bcrypt/sha256) | P0 | 待实现 |
| 实际邮件发送 (SMTP) | P0 | 待实现 |
| 邮件模板渲染 | P0 | 待实现 |
| 忘记密码API | P0 | 待实现 |
| 重置密码API | P0 | 待实现 |
| Token验证页面API | P0 | 待实现 |
| 验证码尝试次数限制 | P1 | 待实现 |
| 统一错误消息 (防枚举) | P1 | 待实现 |
| 邮件发送日志 | P2 | 待实现 |

### 7.2 前端实现

| 任务 | 优先级 | 状态 |
|------|--------|------|
| 忘记密码页面 | P0 | 待实现 |
| 重置密码页面 | P0 | 待实现 |
| 登录页添加"忘记密码"链接 | P0 | 待实现 |
| 注册页优化 (倒计时、错误提示) | P1 | 待实现 |
| 密码强度指示器 | P1 | 待实现 |

### 7.3 配置项

```yaml
email:
  smtp:
    host: "smtp.example.com"
    port: 587
    username: "noreply@gapi.com"
    password: "${SMTP_PASSWORD}"
    from: "gAPI <noreply@gapi.com>"
    use_tls: true
  
verification:
  code_length: 6
  code_expires_minutes: 10
  max_attempts: 3
  send_cooldown_seconds: 60
  rate_limit:
    per_email_per_hour: 5
    per_ip_per_hour: 10
    per_device_per_hour: 3

password_reset:
  token_length: 32
  token_expires_minutes: 60
  rate_limit:
    per_email_per_hour: 3
```

---

## 8. 测试用例

### 8.1 验证码发送测试

| 场景 | 预期结果 |
|------|----------|
| 正常发送 | 返回成功，邮件发送 |
| 邮箱格式错误 | 返回参数错误 |
| 60秒内重复发送 | 返回限流错误 |
| 每小时第6次发送 | 返回限流错误 |
| IP被限制 | 返回IP限流错误 |
| SlideCaptcha未验证 | 返回验证失败 |

### 8.2 验证码验证测试

| 场景 | 预期结果 |
|------|----------|
| 正确验证码 | 返回验证成功，Token |
| 错误验证码 | 返回验证失败 |
| 已使用验证码 | 返回已使用 |
| 过期验证码 | 返回已过期 |
| 3次错误尝试 | 验证码失效 |

### 8.3 密码重置测试

| 场景 | 预期结果 |
|------|----------|
| 有效Token重置 | 密码更新成功 |
| 过期Token | 返回Token无效 |
| 已使用Token | 返回Token无效 |
| 新密码不符合要求 | 返回参数错误 |

---

## 9. 部署检查清单

- [ ] SMTP 服务配置完成
- [ ] 发送邮箱域名配置 (SPF/DKIM/DMARC)
- [ ] 邮件发送日志监控
- [ ] 验证码发送限流监控
- [ ] 密码重置使用率监控

---

## 参考资料

- [OWASP Forgot Password Cheat Sheet](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Forgot_Password_Cheat_Sheet.md)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Django Password Reset Implementation](https://github.com/django/django/blob/main/django/contrib/auth/tokens.py)
- [Laravel Password Reset](https://laravel.com/docs/12.x/passwords)
