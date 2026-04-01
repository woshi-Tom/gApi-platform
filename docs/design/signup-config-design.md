# 注册配置功能设计方案

> 版本: v1.0
> 日期: 2026-04-01
> 状态: 待审核

---

## 1. 功能概述

注册配置允许管理员控制用户注册行为，包括开放/关闭注册、验证码方式、注册奖励等。适用于控制用户增长、防止恶意注册、运营活动等场景。

## 2. 业务模型

### 2.1 注册开关

| 状态 | 说明 | 行为 |
|------|------|------|
| open | 开放注册 | 用户可正常注册 |
| closed | 关闭注册 | 注册页面显示"注册已关闭"，禁止注册 |

### 2.2 验证码方式

> ⚠️ **重要更新 (v1.1)**: 邮箱验证和滑块验证可以同时启用，互不冲突

| 方式 | 说明 | 适用场景 |
|------|------|----------|
| email | 邮箱验证 | 标准注册流程（必选） |
| captcha | 滑块验证 | 防机器人注册（可选） |

**配置说明**：
- **邮箱验证**: 必须启用，防止恶意注册和账号冒用
- **滑块验证**: 可选启用，增强防机器人能力
- 两者可以同时启用：先通过滑块验证，再填写邮箱注册

### 2.3 注册奖励

| 类型 | 说明 |
|------|------|
| quota | 赠送配额数量 |
| trial_vip | 试用VIP天数 |
| none | 无奖励 |

## 3. 数据模型

### 3.1 signup_configs 表

```sql
CREATE TABLE signup_configs (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(50) UNIQUE NOT NULL,
    config_value TEXT,
    description VARCHAR(200),
    updated_by INTEGER,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 默认配置项
INSERT INTO signup_configs (config_key, config_value, description) VALUES
('registration_enabled', 'true', '是否开放注册'),
('require_email_verification', 'true', '是否强制邮箱验证 (必填)'),
('require_captcha', 'false', '是否启用滑块验证 (可选)'),
('default_signup_reward', '{"type":"quota","amount":1000000}', '注册奖励配置'),
('max_accounts_per_ip', '5', '同一IP最大注册账号数'),
('min_password_length', '8', '密码最小长度'),
('allowed_domains', '', '允许注册的邮箱域名，逗号分隔，空表示全部');
```

### 3.2 模型定义

```go
type SignupConfig struct {
    ID          uint
    ConfigKey   string
    ConfigValue string
    Description string
    UpdatedBy   *uint
    UpdatedAt   time.Time
}

// 配置结构体
type SignupSettings struct {
    RegistrationEnabled     bool   `json:"registration_enabled"`
    VerificationMethod      string `json:"verification_method"` // none/email/captcha
    DefaultReward           SignupReward `json:"default_signup_reward"`
    MaxAccountsPerIP       int    `json:"max_accounts_per_ip"`
    MinPasswordLength      int    `json:"min_password_length"`
    RequireEmailVerification bool  `json:"require_email_verification"`
    AllowedDomains         string `json:"allowed_domains"` // 逗号分隔
}

type SignupReward struct {
    Type   string `json:"type"`   // none/quota/trial_vip
    Amount int64  `json:"amount"` // 配额数量或VIP天数
}
```

## 4. 功能模块

### 4.1 注册配置（管理员）

**功能**：
- 开启/关闭注册
- 选择验证码方式
- 配置注册奖励
- 设置IP注册限制
- 设置密码强度要求
- 设置邮箱域名限制

**表单设计**：

```
┌─────────────────────────────────────────────────────────┐
│                    注册配置                              │
├─────────────────────────────────────────────────────────┤
│ 基础设置                                                  │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ ● 允许新用户注册                                      │ │
│ │   ○ 关闭注册（新用户无法注册）                         │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                         │
│ 验证方式                                                  │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ ✓ 邮箱验证 (发送验证链接) - 必填                      │ │
│ │ ✓ 滑块验证 (行为验证码) - 可选                       │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                         │
│ 注册奖励                                                  │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ 奖励类型: [赠送配额 ▼]                               │ │
│ │ 奖励数量: [1000000] Tokens                         │ │
│ │                                                         │ │
│ │ ○ 试用VIP                                             │ │
│ │ VIP天数: [7] 天                                     │ │
│ │                                                         │ │
│ │ ● 无奖励                                              │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                         │
│ 安全设置                                                  │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ 同一IP最大注册: [5] 个账号                           │ │
│ │ 密码最小长度: [8] 个字符                             │ │
│ │ ○ 强制邮箱验证                                       │ │
│ │                                                         │ │
│ │ 允许的邮箱域名 (留空表示全部):                         │ │
│ │ ┌─────────────────────────────────────────────────┐   │ │
│ │ │ @company.com, @partner.com                      │   │ │
│ │ └─────────────────────────────────────────────────┘   │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                         │
│                            [保存配置]                   │
└─────────────────────────────────────────────────────────┘
```

### 4.2 注册页面适配

**开放注册 + 邮箱验证**：
```
┌─────────────────────────────────────────┐
│            用户注册                       │
├─────────────────────────────────────────┤
│ 用户名: [____________]                   │
│ 邮箱:   [____________]                   │
│ 密码:   [____________]                   │
│ 确认:   [____________]                   │
│                                         │
│ [注册]                                   │
│                                         │
│ 已收到验证邮件？[重新发送]                │
└─────────────────────────────────────────┘
```

**关闭注册**：
```
┌─────────────────────────────────────────┐
│            用户注册                       │
├─────────────────────────────────────────┤
│                                         │
│         🔒 注册已关闭                     │
│                                         │
│    抱歉，当前不支持新用户注册              │
│    如需账号请联系管理员                   │
│                                         │
│         [返回首页]  [联系客服]             │
└─────────────────────────────────────────┘
```

**域名限制**：
```
┌─────────────────────────────────────────┐
│            用户注册                       │
├─────────────────────────────────────────┤
│ 用户名: [____________]                   │
│ 邮箱:   [____________] ✗                │
│         仅支持 @company.com 域名          │
│ 密码:   [____________]                   │
│                                         │
│ [注册]                                   │
└─────────────────────────────────────────┘
```

### 4.3 注册流程适配

**无验证流程**：
```
1. 填写注册信息
2. 提交注册
3. 发放注册奖励
4. 注册成功
```

**邮箱验证流程**：
```
1. 填写注册信息
2. 提交注册
3. 创建用户（状态=未验证）
4. 发送验证邮件
5. 用户点击验证链接
6. 验证成功，状态=正常
7. 发放注册奖励
```

**注册奖励发放**：
```
注册奖励类型: quota
→ 增加用户 remain_quota += 奖励数量

注册奖励类型: trial_vip  
→ 更新用户:
  level = "vip_bronze" (或配置的等级)
  vip_expired_at = now + 奖励天数
  vip_quota = 套餐默认配额
```

## 5. API设计

### 5.1 管理端API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/settings/signup` | GET | 获取注册配置 |
| `/api/v1/admin/settings/signup` | PUT | 更新注册配置 |

### 5.2 公开API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/init/status` | GET | 获取系统初始化状态（包含注册开关） |
| `/api/v1/user/register` | POST | 用户注册 |
| `/api/v1/user/verify-email` | POST | 验证邮箱 |

### 5.3 请求/响应示例

**获取注册配置**：
```json
GET /api/v1/admin/settings/signup

{
  "success": true,
  "data": {
    "registration_enabled": true,
    "verification_method": "email",
    "default_signup_reward": {
      "type": "quota",
      "amount": 1000000
    },
    "max_accounts_per_ip": 5,
    "min_password_length": 8,
    "require_email_verification": true,
    "allowed_domains": ""
  }
}
```

**更新配置**：
```json
PUT /api/v1/admin/settings/signup
{
  "registration_enabled": true,
  "verification_method": "email",
  "default_signup_reward": {
    "type": "trial_vip",
    "amount": 7
  },
  "max_accounts_per_ip": 3,
  "min_password_length": 8,
  "require_email_verification": true,
  "allowed_domains": "@company.com,@partner.com"
}
```

**用户注册**：
```json
POST /api/v1/user/register
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123"
}

// 响应 - 需要邮箱验证
{
  "success": true,
  "data": {
    "message": "注册成功，请查收验证邮件",
    "user_id": 123,
    "require_verification": true
  }
}
```

## 6. 业务规则

### 6.1 IP注册限制

```sql
-- 检查同一IP注册数量
SELECT COUNT(*) FROM users 
WHERE ip_registered_from = :client_ip 
AND created_at > NOW() - INTERVAL '24 hours';

-- 如果超过限制，返回错误
if (count >= config.max_accounts_per_ip) {
    return error("注册过于频繁，请稍后再试")
}
```

### 6.2 邮箱域名限制

```go
func isAllowedDomain(email string, allowed string) bool {
    if allowed == "" {
        return true // 空表示全部允许
    }
    domains := strings.Split(allowed, ",")
    userDomain := email[strings.Index(email, "@"):]
    for _, d := range domains {
        if userDomain == d {
            return true
        }
    }
    return false
}
```

### 6.3 验证状态流转

```
用户注册 → status: unverified
    ↓
验证邮箱 → status: active (如果 require_email_verification=true)
    或
注册即激活 → status: active (如果 require_email_verification=false)
```

## 7. 安全考虑

- IP注册频率限制（Redis计数）
- 密码强度验证
- 邮箱域名白名单
- 注册请求速率限制
- 验证码防止机器人注册
- 注册日志记录

## 8. 后续扩展

- [ ] 邀请注册（邀请码）
- [ ] 注册审批流程
- [ ] 注册来源追踪
- [ ] 黑名单管理

---

## 审核意见

（待填写）
