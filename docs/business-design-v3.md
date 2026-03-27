# gAPI Platform - 业务设计方案 v3.0

**版本**: 3.0  
**日期**: 2026-03-26  
**状态**: 设计讨论中

---

## 1. 管理员安全设计

### 1.1 独立端口方案

```
架构：
┌─────────────────┐
│   用户端 (8080)  │  ── 对外开放，支持公网访问
└─────────────────┘
         ↓
┌─────────────────┐
│   管理端 (9000)  │  ── 仅内网访问，独立认证
└─────────────────┘
```

**实现方案**:
- 用户端保持 8080 端口
- 管理端使用 9000 端口，独立 Gin 实例
- 管理端只绑定内网 IP (127.0.0.1 或 10.x.x.x)
- Nginx 反向代理仅暴露 8080

**配置文件** (`config.yaml`):
```yaml
server:
  user_port: 8080
  admin_port: 9000
  admin_bind: "127.0.0.1"  # 只允许内网访问
```

### 1.2 管理员密码安全

#### 数据库变更
```sql
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    password_expire_days INTEGER DEFAULT 90,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ;
```

#### 功能设计
1. **密码修改**
   - 管理员可以修改自己的密码
   - 需要输入旧密码验证
   - 新密码强度检查（8+ 位，包含大小写字母、数字、特殊字符）

2. **密码过期策略**
   - 默认 90 天过期
   - 过期前 7 天开始提醒
   - 过期后强制修改

3. **登录安全**
   - 5 次失败后锁定 15 分钟
   - 支持 Google Authenticator（后期实现）

---

## 2. 注册防护设计

### 2.1 注册流程（更新）

```
注册流程：
用户请求 → [滑块验证] → [邮箱验证] → [限流检查] → 创建账号
                ↓            ↓            ↓
           验证码token    6位验证码     多维度限流
           有效期可配      有效期可配    IP+指纹+邮箱
```

**流程说明**：
1. 用户填写基本信息后，先完成滑块验证
2. 滑块验证通过后，系统向邮箱发送验证码
3. 用户输入邮箱验证码
4. 验证通过后，系统进行多维度限流检查
5. 检查通过，创建账号

### 2.2 多维度限流设计

#### 限流维度
```go
// 限流维度（按优先级）
type RateLimitDimension struct {
    // 1. IP限流（辅助）
    IP string
    
    // 2. 设备指纹（主要）
    Fingerprint string
    
    // 3. 邮箱域名（辅助）
    EmailDomain string
    
    // 4. 浏览器指纹（辅助）
    BrowserFingerprint string
}
```

#### 限流规则
| 维度 | 限制次数 | 时间窗口 | 封禁时间 |
|------|---------|---------|---------|
| IP地址 | 10次/天 | 24小时 | 24小时 |
| 设备指纹 | 3次/天 | 24小时 | 7天 |
| 邮箱域名 | 20次/小时 | 1小时 | 1小时 |
| 组合指纹 | 5次/天 | 24小时 | 24小时 |

#### 设备指纹收集
```typescript
// 前端收集设备指纹
interface DeviceFingerprint {
    // 基础信息
    userAgent: string
    screenResolution: string
    timezone: string
    language: string
    platform: string
    
    // 高级指纹
    canvasFingerprint: string    // Canvas指纹
    webGLFingerprint: string     // WebGL指纹
    audioFingerprint: string     // 音频指纹
    fontFingerprint: string      // 字体指纹
    
    // 行为特征
    mouseMovements: number[]     // 鼠标移动轨迹
    keyboardTimings: number[]    // 键盘输入间隔
    scrollBehavior: number[]     // 滚动行为
}

// 后端指纹验证
type FingerprintVerification struct {
    IsValid bool
    RiskScore float64  // 风险评分 0-1
    Reasons []string   // 风险原因
}
```

#### Redis存储结构
```go
// 指纹存储
// key: "fingerprint:{hash}"
// value: {
//   "count": 3,
//   "first_attempt": "2026-03-26T10:00:00Z",
//   "last_attempt": "2026-03-26T10:30:00Z",
//   "blocked_until": "2026-03-27T10:00:00Z"
// }
// TTL: 7天

// IP存储
// key: "ratelimit:register:ip:{ip}"
// value: 类似结构

// 邮箱域名存储
// key: "ratelimit:register:domain:{domain}"
// value: 类似结构
```

### 2.3 验证码有效期配置

#### 配置项
```sql
-- 系统配置表新增
INSERT INTO system_configs (config_key, config_value, config_group, description) VALUES 
-- 滑块验证码
('captcha_slider_expire', '300', 'captcha', '滑块验证码有效期（秒）'),
('captcha_slider_max_attempts', '3', 'captcha', '滑块最大尝试次数'),

-- 邮箱验证码
('captcha_email_expire', '600', 'captcha', '邮箱验证码有效期（秒）'),
('captcha_email_max_send', '5', 'captcha', '邮箱验证码每小时最大发送次数'),
('captcha_email_length', '6', 'captcha', '邮箱验证码长度'),

-- 限流配置
('ratelimit_register_ip_day', '10', 'ratelimit', 'IP每天注册限制'),
('ratelimit_register_fingerprint_day', '3', 'ratelimit', '设备指纹每天注册限制'),
('ratelimit_register_domain_hour', '20', 'ratelimit', '邮箱域名每小时注册限制'),
('ratelimit_register_block_hours', '24', 'ratelimit', '注册封禁时间（小时）');
```

#### 配置管理
```go
// 配置服务
type CaptchaConfigService struct {
    db *gorm.DB
}

func (s *CaptchaConfigService) GetConfig(key string) string {
    var config SystemConfig
    s.db.Where("config_key = ?", key).First(&config)
    return config.ConfigValue
}

func (s *CaptchaConfigService) GetCaptchaExpire() time.Duration {
    expireStr := s.GetConfig("captcha_email_expire")
    expire, _ := strconv.Atoi(expireStr)
    return time.Duration(expire) * time.Second
}
```

### 2.4 滑动验证码（5秒盾）

#### 方案选择（更新）
| 方案 | 优点 | 缺点 | 推荐度 | 实施阶段 |
|------|------|------|--------|---------|
| 自研简单滑块 | 完全自主、成本低 | 效果一般 | ★★★ | 第一阶段 |
| 极验 | 成熟稳定、效果好 | 收费 | ★★★★★ | 后期集成 |
| 腾讯云验证码 | 免费额度、效果好 | 依赖外部 | ★★★★ | 后期可选 |

**实施策略**：
- 第一阶段：自研简单滑块
- 第二阶段：根据业务需要集成极验或腾讯云验证码

#### 自研滑块实现
```typescript
// 前端组件
interface SliderCaptcha {
    bgImage: string        // 背景图
    sliderImage: string    // 滑块图
    sliderX: number        // 滑块位置
    success: boolean       // 验证结果
    token: string          // 验证token（后端生成）
    expiresAt: number      // 过期时间戳
}

// 后端验证
type CaptchaResult struct {
    Token      string    `json:"token"`
    ExpiresAt  time.Time `json:"expires_at"`
    Track      []int     `json:"track"`    // 滑动轨迹
    Duration   int64     `json:"duration"` // 滑动时长
    RiskScore  float64   `json:"risk_score"` // 风险评分
}

// 验证逻辑
func (s *CaptchaService) VerifySlider(token string, track []int, duration int64) (bool, float64) {
    // 1. 验证token有效性
    if !s.verifyToken(token) {
        return false, 1.0
    }
    
    // 2. 分析滑动轨迹
    riskScore := s.analyzeTrack(track, duration)
    
    // 3. 判断是否通过
    return riskScore < 0.7, riskScore
}

// 轨迹分析
func (s *CaptchaService) analyzeTrack(track []int, duration int64) float64 {
    // 检查滑动速度是否合理
    // 检查轨迹是否平滑
    // 检查是否存在异常行为
    // 返回风险评分 0-1
}
```

### 2.5 邮箱验证

#### 验证流程（更新）
```
1. 滑块验证通过 → 获得captcha_token
2. 用户填写邮箱 → 调用发送验证码API（携带captcha_token）
3. 后端验证captcha_token有效 → 发送6位验证码
4. 验证码有效期：可配置（默认10分钟）
5. 验证码最多发送：可配置（默认5次/小时）
6. 用户输入验证码 → 验证通过
7. 继续注册流程
```

#### 数据库变更
```sql
-- 邮箱验证表
CREATE TABLE email_verifications (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(100) NOT NULL,
    code            VARCHAR(10) NOT NULL,
    ip_address      VARCHAR(50),
    user_agent      VARCHAR(500),
    device_hash     VARCHAR(64),                -- 设备指纹哈希
    
    -- 验证码信息
    captcha_token   VARCHAR(100),               -- 关联的滑块验证码
    
    -- 状态
    is_used         BOOLEAN DEFAULT FALSE,
    used_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    
    -- 审计
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_email ON email_verifications(email);
CREATE INDEX idx_email_verifications_expires ON email_verifications(expires_at);
CREATE INDEX idx_email_verifications_device ON email_verifications(device_hash);
```

#### 发送限制
```go
// 检查发送限制
func (s *EmailService) CheckSendLimit(email string, ip string, deviceHash string) error {
    // 1. 检查邮箱发送频率
    emailKey := fmt.Sprintf("email_send:%s", email)
    if s.isRateLimited(emailKey, 5, time.Hour) {
        return errors.New("邮箱验证码发送过于频繁")
    }
    
    // 2. 检查IP发送频率
    ipKey := fmt.Sprintf("email_send:ip:%s", ip)
    if s.isRateLimited(ipKey, 10, time.Hour) {
        return errors.New("IP验证码发送过于频繁")
    }
    
    // 3. 检查设备发送频率
    deviceKey := fmt.Sprintf("email_send:device:%s", deviceHash)
    if s.isRateLimited(deviceKey, 3, time.Hour) {
        return errors.New("设备验证码发送过于频繁")
    }
    
    return nil
}
```

### 2.6 实施优先级

#### 第一阶段（必做）
- [ ] 自研简单滑块验证码
- [ ] 邮箱验证码（带有效期配置）
- [ ] IP + 设备指纹限流

#### 第二阶段（可选）
- [ ] 集成极验或腾讯云验证码
- [ ] 高级行为分析
- [ ] 机器学习风险评估

---

## 3. 限速业务设计

### 3.1 业务逻辑

```
限速策略（从宽到严）：
┌─────────────────────────────────────────────────────────┐
│  VIP订阅用户：低价（严格）→ 中价（宽松）→ 高价（基本不限）  │
│  充值用户：统一宽松限速                                    │
│  普通用户：初期不限速，后期限制                             │
└─────────────────────────────────────────────────────────┘
```

### 3.2 数据库变更

#### 3.2.1 修改 VIP 套餐表
```sql
ALTER TABLE vip_packages ADD COLUMN IF NOT EXISTS
    rpm_limit_free      INTEGER DEFAULT 0,      -- 免费RPM（赠送给免费用户）
    rpm_limit_recharge  INTEGER DEFAULT 500,    -- 充值用户RPM
    rpm_limit_vip       INTEGER DEFAULT 2000,   -- VIP用户RPM
    tpm_limit_free      INTEGER DEFAULT 10000,
    tpm_limit_recharge  INTEGER DEFAULT 100000,
    tpm_limit_vip       INTEGER DEFAULT 500000;
```

#### 3.2.2 修改充值套餐表
```sql
ALTER TABLE recharge_packages ADD COLUMN IF NOT EXISTS
    rpm_limit           INTEGER DEFAULT 200,    -- RPM限制
    tpm_limit           INTEGER DEFAULT 50000,  -- TPM限制
    concurrent_limit    INTEGER DEFAULT 5;      -- 并发限制
```

#### 3.2.3 新增限速配置表
```sql
CREATE TABLE rate_limit_packages (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 套餐信息
    name            VARCHAR(100) NOT NULL,
    package_type    VARCHAR(20) NOT NULL,       -- free|recharge|vip
    price           DECIMAL(10,2) DEFAULT 0,
    
    -- 限速配置
    rpm_limit       INTEGER NOT NULL,           -- Requests Per Minute
    tpm_limit       INTEGER NOT NULL,           -- Tokens Per Minute
    concurrent      INTEGER DEFAULT 10,         -- 并发限制
    
    -- 特殊配置
    burst_limit     INTEGER,                    -- 突发限制（允许短时间超速）
    burst_duration  INTEGER DEFAULT 60,         -- 突发时长（秒）
    
    -- 显示
    sort_order      INTEGER DEFAULT 0,
    is_recommended  BOOLEAN DEFAULT FALSE,
    is_visible      BOOLEAN DEFAULT TRUE,
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'active',
    
    -- 审计
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_rate_limit_tenant ON rate_limit_packages(tenant_id);
CREATE INDEX idx_rate_limit_type ON rate_limit_packages(package_type);
```

#### 3.2.4 修改用户表
```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS
    current_rpm_limit   INTEGER DEFAULT 60,     -- 当前RPM限制
    current_tpm_limit   INTEGER DEFAULT 10000,  -- 当前TPM限制
    rate_limit_id       BIGINT REFERENCES rate_limit_packages(id),
    rate_limit_reset_at TIMESTAMPTZ;            -- 限速重置时间
```

### 3.3 限速逻辑实现

```go
// 限速检查器
type RateLimiter struct {
    redis *redis.Client
    db    *gorm.DB
}

// 检查RPM限制
func (rl *RateLimiter) CheckRPM(userID uint, tokenID uint) error {
    // 获取用户当前RPM限制
    limit := rl.getUserRPMLimit(userID)
    
    // Redis key
    key := fmt.Sprintf("ratelimit:rpm:%d:%d", userID, tokenID)
    
    // INCR + EXPIRE
    count, err := rl.redis.Incr(key).Result()
    if err != nil {
        return err
    }
    
    if count == 1 {
        rl.redis.Expire(key, time.Minute)
    }
    
    if count > int64(limit) {
        return errors.New("RPM limit exceeded")
    }
    
    return nil
}

// 获取用户RPM限制
func (rl *RateLimiter) getUserRPMLimit(userID uint) int {
    var user User
    rl.db.First(&user, userID)
    
    // 优先级：用户自定义 > 套餐限制 > 默认限制
    if user.CurrentRPMLimit > 0 {
        return user.CurrentRPMLimit
    }
    
    // 根据用户等级返回限制
    switch user.Level {
    case "vip":
        return rl.getVIPRPM(user.VIPPackageID)
    case "premium":
        return rl.getRechargeRPM(user.RateLimitID)
    default:
        return 60  // 免费用户默认
    }
}
```

### 3.4 限速套餐设计

#### 免费用户
| 套餐 | RPM | TPM | 价格 |
|------|-----|-----|------|
| 免费 | 60 | 10000 | 0 |

#### 充值套餐（永久配额）
| 套餐 | 配额 | RPM | TPM | 价格 | 并发 |
|------|------|-----|-----|------|------|
| 入门 | 100K | 100 | 20000 | ¥1 | 3 |
| 基础 | 1M | 200 | 50000 | ¥10 | 5 |
| 专业 | 5M | 500 | 100000 | ¥50 | 10 |
| 企业 | 10M | 1000 | 200000 | ¥100 | 20 |

#### VIP订阅套餐（限时配额）
| 套餐 | 配额 | RPM | TPM | 价格 | 并发 | 特点 |
|------|------|-----|-----|------|------|------|
| VIP基础 | 1M | 200 | 50000 | ¥9.9/月 | 5 | 限速严格 |
| VIP专业 | 5M | 1000 | 200000 | ¥29.9/月 | 15 | 限速宽松 |
| VIP企业 | 20M | 5000 | 1000000 | ¥99.9/月 | 50 | 基本不限 |

---

## 4. 管理员功能设计

### 4.1 功能模块

```
管理员后台
├── 仪表盘
│   ├── 实时统计
│   ├── 收入报表
│   └── 用户增长
├── 用户管理
│   ├── 用户列表
│   ├── 配额调整
│   └── 限速调整
├── 商品管理
│   ├── 充值套餐管理
│   │   ├── 上架/下架
│   │   ├── 价格调整
│   │   └── 限速配置
│   ├── VIP套餐管理
│   │   ├── 上架/下架
│   │   ├── 价格调整
│   │   └── 权益配置
│   └── 限速套餐管理（新）
│       ├── 套餐列表
│       ├── 创建套餐
│       └── 限速配置
├── 渠道管理
├── 订单管理
├── 系统设置
│   ├── 基础设置
│   ├── 支付设置
│   ├── 邮件设置
│   └── 限速策略
└── 安全设置
    ├── 修改密码
    ├── 登录日志
    └── 操作日志
```

### 4.2 商品管理 API

```go
// 充值套餐管理
GET    /admin/recharge-packages      // 列表
POST   /admin/recharge-packages      // 创建
PUT    /admin/recharge-packages/:id  // 更新
DELETE /admin/recharge-packages/:id  // 删除
PATCH  /admin/recharge-packages/:id/status  // 上架/下架

// VIP套餐管理
GET    /admin/vip-packages
POST   /admin/vip-packages
PUT    /admin/vip-packages/:id
DELETE /admin/vip-packages/:id
PATCH  /admin/vip-packages/:id/status

// 限速套餐管理（新）
GET    /admin/rate-limit-packages
POST   /admin/rate-limit-packages
PUT    /admin/rate-limit-packages/:id
DELETE /admin/rate-limit-packages/:id
PATCH  /admin/rate-limit-packages/:id/status
```

### 4.3 商品编辑界面

```
┌─────────────────────────────────────────────────┐
│                 创建/编辑套餐                      │
├─────────────────────────────────────────────────┤
│ 套餐名称: [输入框]                                │
│ 套餐类型: [下拉选择: 充值/VIP/限速]                │
│ 价格: [数字输入] 元                               │
│ 原价: [数字输入] 元                               │
│                                                 │
│ ┌─ 配额配置 ─────────────────────────────────┐   │
│ │ 配额: [数字输入] Token                       │   │
│ │ 赠送配额: [数字输入] Token                    │   │
│ └─────────────────────────────────────────────┘   │
│                                                 │
│ ┌─ 限速配置 ─────────────────────────────────┐   │
│ │ RPM限制: [数字输入] 请求/分钟                 │   │
│ │ TPM限制: [数字输入] Token/分钟                │   │
│ │ 并发限制: [数字输入] 并发数                   │   │
│ │ 突发限制: [数字输入] 请求数                   │   │
│ │ 突发时长: [数字输入] 秒                      │   │
│ └─────────────────────────────────────────────┘   │
│                                                 │
│ ┌─ 显示配置 ─────────────────────────────────┐   │
│ │ 排序: [数字输入]                             │   │
│ │ ☐ 推荐套餐                                  │   │
│ │ ☐ 热门套餐                                  │   │
│ │ ☐ 可见                                      │   │
│ └─────────────────────────────────────────────┘   │
│                                                 │
│ [取消]  [保存]                                   │
└─────────────────────────────────────────────────┘
```

---

## 5. 实施计划

### 5.1 第一阶段：管理员安全（1-2天）

- [ ] 后端：管理员独立端口
- [ ] 后端：管理员修改密码 API
- [ ] 前端：管理员密码修改页面
- [ ] 测试：端口隔离、密码修改

### 5.2 第二阶段：注册防护（2-3天）

- [ ] 后端：邮箱验证服务
- [ ] 后端：滑块验证码
- [ ] 后端：IP限流
- [ ] 前端：滑块组件
- [ ] 前端：邮箱验证流程
- [ ] 测试：防护效果

### 5.3 第三阶段：限速业务（3-4天）

- [ ] 数据库：新增限速表
- [ ] 后端：限速检查器
- [ ] 后端：限速套餐管理 API
- [ ] 前端：限速套餐管理界面
- [ ] 前端：用户限速配置界面
- [ ] 测试：限速效果

### 5.4 第四阶段：商品管理（2-3天）

- [ ] 后端：商品管理 API
- [ ] 前端：商品管理界面
- [ ] 前端：上下架功能
- [ ] 测试：商品管理

---

## 6. 技术依赖

### 6.1 后端依赖
- Redis（限流、会话）
- 邮件服务（SMTP配置）
- 滑块验证码（可选：极验）

### 6.2 前端依赖
- 滑块组件（自研或第三方）
- 验证码组件

---

## 7. 风险评估

| 风险 | 影响 | 解决方案 |
|------|------|----------|
| 滑块验证码被破解 | 注册泛洪 | 定期更新算法、集成专业服务 |
| 邮件发送延迟 | 用户体验差 | 使用专业邮件服务 |
| 限速误判 | 用户投诉 | 设置白名单、提供申诉渠道 |
| 管理员密码泄露 | 系统安全 | 强密码策略、定期更换 |
| 设备指纹收集失败 | 限制不准确 | 降级到IP限制、多维度组合 |
| 验证码过期配置错误 | 用户体验差 | 提供合理的默认值、管理界面配置 |

---

## 8. 配置管理界面

### 8.1 验证码配置

```
┌─────────────────────────────────────────────────┐
│                 验证码配置                        │
├─────────────────────────────────────────────────┤
│ ┌─ 滑块验证码 ─────────────────────────────┐    │
│ │ 验证码有效期: [300] 秒                     │    │
│ │ 最大尝试次数: [3] 次                       │    │
│ └───────────────────────────────────────────┘    │
│                                                 │
│ ┌─ 邮箱验证码 ─────────────────────────────┐    │
│ │ 验证码有效期: [600] 秒                     │    │
│ │ 每小时最大发送: [5] 次                     │    │
│ │ 验证码长度: [6] 位                         │    │
│ └───────────────────────────────────────────┘    │
│                                                 │
│ ┌─ 限流配置 ───────────────────────────────┐    │
│ │ IP每天注册限制: [10] 次                    │    │
│ │ 设备指纹每天限制: [3] 次                   │    │
│ │ 邮箱域名每小时限制: [20] 次                │    │
│ │ 封禁时间: [24] 小时                        │    │
│ └───────────────────────────────────────────┘    │
│                                                 │
│ [重置默认值]  [保存配置]                          │
└─────────────────────────────────────────────────┘
```

### 8.2 API接口

```go
// 获取配置
GET /admin/config/captcha
GET /admin/config/ratelimit

// 更新配置
PUT /admin/config/captcha
PUT /admin/config/ratelimit

// 重置默认值
POST /admin/config/reset
```

---

**文档版本**: 3.1  
**更新日期**: 2026-03-26  
**下一步**: 确认需求后开始实施
