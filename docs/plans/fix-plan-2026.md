# gAPI Platform 修复计划文档

> 版本: v1.2
> 日期: 2026-05-14
> 状态: 全部完成（N-003 基础支持已存在，D-001 已同步）

---

## 文档说明

本文档基于 2026-05-05 源代码审查结果编制，记录所有待修复问题、新增功能需求和文档更新任务。

**使用说明**:
1. 每个任务独立编号，便于追踪
2. 每个任务标注风险等级、影响范围、前后端配合要求
3. 实施顺序按依赖关系和风险度排列
4. 完成后更新"状态"列并记录实施日期

---

## 一、紧急修复项 (🔴 高优先级 - 影响功能正确性)

### F-001: GORM FOR UPDATE 语法错误

| 项目 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/redemption_handler.go` |
| **行号** | 283 |
| **问题** | `tx.Set("gorm:query", "FOR UPDATE")` 语法错误，GORM 正确写法是 `tx.Set("gorm:query_option", "FOR UPDATE")` |
| **影响** | 行级锁可能不生效，并发兑换时可能出现数据竞争 |
| **风险等级** | 🔴 高 |
| **前后端配合** | 否 (仅后端) |
| **依赖项** | 无 |
| **状态** | ✅ 已修复（v1.1.1） |
| **修复方案** | 已在第 283 行使用 `gorm:query_option` |

**验证方法**:
1. 已使用 `gorm:query_option`，行级锁正常生效

---

### F-002: 兑换码操作未记录审计日志

| 项目 | 内容 |
|------|------|
| **文件** | `backend/internal/handler/redemption_handler.go` |
| **行号** | Create: 145-170, Redeem: 367-387 |
| **问题** | 兑换码兑换和创建操作未调用审计日志中间件，与设计文档不符 |
| **影响** | 无法追溯兑换码操作，违反合规要求 |
| **风险等级** | 🟡 中 |
| **前后端配合** | 否 (仅后端) |
| **依赖项** | 无 |
| **状态** | ✅ 已修复（v1.1.1） |

**修复方案**:
在 `redemption_handler.go` 中，Redeem 和 Create 方法执行成功后，手动记录审计日志:

```go
// Redeem 方法成功后添加 (约345行附近)
if h.auditRepo != nil {
    h.auditRepo.Create(&model.AuditLog{
        UserID:       userID,
        Username:     user.Username,
        Action:       model.AuditActionRedemptionUse,
        ActionGroup:  model.AuditGroupRedemption,
        ResourceType: "redemption_code",
        ResourceID:   redemptionCode.ID,
        RequestIP:    c.ClientIP(),
        Success:      true,
        NewValue: map[string]interface{}{
            "code":         redemptionCode.Code,
            "quota_granted": quotaGranted,
            "vip_granted":  vipGranted,
            "vip_days":     vipDays,
        },
    })
}

// Create 方法成功后添加 (约149行附近)
adminID := getAdminIDFromContext(c)
if h.auditRepo != nil && adminID != nil {
    h.auditRepo.Create(&model.AuditLog{
        UserID:       *adminID,
        Action:       model.AuditActionRedemptionCreate,
        ActionGroup:  model.AuditGroupRedemption,
        ResourceType: "redemption_code",
        RequestIP:    c.ClientIP(),
        Success:      true,
        NewValue: map[string]interface{}{
            "batch_id": batchID,
            "count":   len(codes),
            "code_type": req.CodeType,
        },
    })
}
```

**验证方法**:
1. 创建兑换码后，查询 `audit_logs` 表确认记录存在
2. 用户兑换兑换码后，查询 `audit_logs` 表确认记录存在

---

### F-003: 注册流程完整实现与Signup Config对接

#### 3.1 问题描述

用户注册时未读取 `signup_configs` 表的以下配置：
- `registration_enabled` - 注册开关
- `allowed_domains` - 邮箱域名限制
- `max_accounts_per_ip` - IP注册数量限制
- `default_signup_reward` - 注册奖励
- `require_email_verification` - 是否强制邮箱验证
- `min_password_length` - 密码最小长度

#### 3.2 业务策略设计

| 场景 | 前端行为 | 后端行为 | 备注 |
|------|---------|---------|------|
| 注册开启 | 显示正常注册页面 | 正常处理注册 | - |
| 注册关闭 | 显示"注册已关闭"页面，禁止访问 | 返回错误: "REGISTRATION_CLOSED" | **需新增页面** |
| 邮箱域名限制 | 输入时实时提示"仅支持xxx域名" | 注册时返回: "EMAIL_DOMAIN_NOT_ALLOWED" | 前端配合 |
| IP注册超限 | 显示"注册过于频繁" | 返回: "IP_REGISTRATION_LIMIT_EXCEEDED" | 24小时内限制 |
| 密码强度不足 | 输入时实时提示 | 注册时返回: "PASSWORD_TOO_WEAK" | 需前端配合 |
| 强制邮箱验证 | 注册后跳转验证页 | 用户状态=unverified | 需前端配合 |
| 无需邮箱验证 | 注册后直接登录 | 用户状态=active，直接发Token | 需前端配合 |

#### 3.3 后端修复

| 项目 | 内容 |
|------|------|
| **文件** | `backend/internal/service/auth_service.go` |
| **文件** | `backend/internal/handler/user_handler.go` |
| **文件** | `backend/internal/model/response.go` |
| **问题** | Register方法未读取和应用注册配置 |
| **风险等级** | 🔴 高 |
| **前后端配合** | 是 (见3.5节) |
| **依赖项** | F-004 (先完成设置服务验证) |
| **状态** | ✅ 已修复（v1.2.0） |

**后端错误码定义** (`model/response.go`):

```go
type ErrorCode string

const (
    ErrCodeRegistrationClosed    ErrorCode = "REGISTRATION_CLOSED"
    ErrCodePasswordTooWeak       ErrorCode = "PASSWORD_TOO_WEAK"
    ErrCodeEmailDomainNotAllowed  ErrorCode = "EMAIL_DOMAIN_NOT_ALLOWED"
    ErrCodeIPRegistrationLimit    ErrorCode = "IP_REGISTRATION_LIMIT_EXCEEDED"
)
```

**后端修复代码** (`auth_service.go`):

```go
func (s *AuthService) Register(c *gin.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // ==================== 第一阶段: 配置检查 ====================
    
    // 1. 读取注册配置
    settingsService := service.NewSettingsService(s.db)
    registerSettings, err := settingsService.GetRegisterSettings()
    if err != nil {
        return nil, fmt.Errorf("failed to get register settings: %w", err)
    }

    // 2. 检查注册是否开启
    if !registerSettings.RegistrationEnabled {
        return nil, ErrRegistrationClosed
    }

    // 3. 验证密码强度
    if len(req.Password) < registerSettings.MinPasswordLength {
        return nil, ErrPasswordTooWeak
    }

    // 4. 检查邮箱域名限制
    if registerSettings.AllowedDomains != "" {
        if !isDomainAllowed(req.Email, registerSettings.AllowedDomains) {
            return nil, ErrEmailDomainNotAllowed
        }
    }

    // 5. 检查IP注册数量限制 (24小时内)
    clientIP := c.ClientIP()
    ipRegCount, err := s.userRepo.CountByIPLast24Hours(clientIP)
    if err != nil {
        return nil, fmt.Errorf("failed to check IP: %w", err)
    }
    if ipRegCount >= registerSettings.MaxAccountsPerIP {
        return nil, ErrIPRegistrationLimit
    }

    // ==================== 第二阶段: 用户创建 ====================
    
    // 6. 创建用户
    user := &model.User{
        Email:        req.Email,
        Username:     req.Username,
        PasswordHash: bcrypt hash,
        Level:        "free",
        Status:       "active",
    }

    // 7. 处理邮箱验证状态
    if registerSettings.RequireEmailVerification {
        user.Status = "unverified"
        user.EmailVerified = false
    }

    // 8. 发放注册奖励
    if registerSettings.DefaultSignupReward.Type != "none" {
        if registerSettings.DefaultSignupReward.Type == "quota" {
            user.FreeQuota += registerSettings.DefaultSignupReward.Amount
        } else if registerSettings.DefaultSignupReward.Type == "trial_vip" {
            // 开通试用VIP...
        }
    }

    // ==================== 第三阶段: 响应处理 ====================
    
    if registerSettings.RequireEmailVerification {
        return &RegisterResponse{
            Message:             "注册成功，请查收验证邮件",
            RequireVerification: true,
            UserID:              user.ID,
        }, nil
    } else {
        token, _ := s.GenerateToken(user)
        return &RegisterResponse{
            Message:             "注册成功",
            RequireVerification: false,
            Token:               token,
            User:                user,
        }, nil
    }
}

// 错误定义
var (
    ErrRegistrationClosed     = errors.New("registration_closed")
    ErrPasswordTooWeak         = errors.New("password_too_weak")
    ErrEmailDomainNotAllowed  = errors.New("email_domain_not_allowed")
    ErrIPRegistrationLimit    = errors.New("ip_registration_limit_exceeded")
)
```

**用户模型扩展** (`model/user.go`):

```go
type User struct {
    // ... existing fields
    IPRegisteredFrom string `json:"ip_registered_from"`  // 注册时的IP
}
```

**用户仓库方法** (`repository/user_repo.go`):

```go
func (r *UserRepository) CountByIPLast24Hours(ip string) (int64, error) {
    var count int64
    err := r.db.Model(&model.User{}).
        Where("ip_registered_from = ? AND created_at > ?", ip, time.Now().Add(-24*time.Hour)).
        Count(&count).Error
    return count, err
}
```

#### 3.4 前端修复 (需配合后端)

| 项目 | 内容 |
|------|------|
| **文件** | `frontend/src/views/Register.vue` |
| **文件** | `frontend/src/router/index.ts` |
| **新增页面** | `frontend/src/views/RegisterClosed.vue` |
| **问题** | 需适配多种注册配置场景 |
| **状态** | ⬜ 待修复 |

**前端Register.vue改造**:

```typescript
// 1. 实时验证邮箱域名
const validateEmail = () => {
  const allowedDomains = registerConfig.value.allowedDomains
  if (allowedDomains && !isDomainAllowed(form.email, allowedDomains)) {
    formErrors.email = `仅支持 ${allowedDomains} 域名`
  }
}

// 2. 密码强度实时提示
const validatePassword = () => {
  const minLength = registerConfig.value.minPasswordLength || 8
  if (form.password.length < minLength) {
    formErrors.password = `密码长度至少 ${minLength} 位`
  }
}

// 3. 错误处理
catch (err) {
  switch (err.response?.data?.error?.code) {
    case 'REGISTRATION_CLOSED':
      router.push('/register-closed')
      break
    case 'PASSWORD_TOO_WEAK':
      formErrors.password = err.response.data.error.message
      break
    case 'EMAIL_DOMAIN_NOT_ALLOWED':
      formErrors.email = err.response.data.error.message
      break
    case 'IP_REGISTRATION_LIMIT_EXCEEDED':
      ElMessage.error('注册过于频繁，请24小时后再试')
      break
  }
}
```

**新增页面RegisterClosed.vue**:

```vue
<template>
  <div class="register-closed">
    <el-result
      icon="warning"
      title="注册已关闭"
      sub-title="抱歉，当前不支持新用户注册。如需账号请联系管理员。"
    >
      <template #extra>
        <el-button type="primary" @click="goHome">返回首页</el-button>
        <el-button @click="contactAdmin">联系管理员</el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup>
const router = useRouter()
const goHome = () => router.push('/')
const contactAdmin = () => {
  // 可配置管理员联系方式
  ElMessage.info('请联系: admin@example.com')
}
</script>
```

**路由配置** (`router/index.ts`):

```typescript
{
  path: '/register-closed',
  name: 'RegisterClosed',
  component: () => import('@/views/RegisterClosed.vue')
}
```

---

### F-004: 验证Signup Config服务完整性

| 项目 | 内容 |
|------|------|
| **文件** | `backend/internal/service/settings_service.go` |
| **问题** | 需确认 GetRegisterSettings 方法是否返回完整配置对象 |
| **影响** | F-003 修复的前提 |
| **风险等级** | 🟡 中 |
| **前后端配合** | 否 |
| **依赖项** | 无 |
| **状态** | ✅ 已验证（v1.2.0） |

**验证结果**:
`GetRegisterSettings()` 方法存在且返回完整 `RegisterSettings` 结构体：
- `AllowRegister`（等效 RegistrationEnabled）
- `AllowedDomains` ✅
- `MaxAccountsPerIP` ✅
- `SignupRewardType` + `SignupRewardAmount`（等效 DefaultSignupReward） ✅
- `MinPasswordLength` ✅
- `RequireEmailVerify`（等效 RequireEmailVerification） ✅

---

## 二、功能完善项 (🟡 中优先级 - 增强功能完整性)

### F-005: 兑换码永久VIP支持字段持久化

| 项目 | 内容 |
|------|------|
| **文件** | `backend/internal/model/` |
| **问题** | `RedemptionCode.IsPermanent` 字段已存在，但数据库表结构可能缺失 |
| **影响** | 永久VIP兑换码功能不完整 |
| **风险等级** | 🟡 中 |
| **前后端配合** | 否 |
| **依赖项** | 无 |
| **状态** | ✅ 已验证（`is_permanent` 列存在） |

**验证结果**:
数据库 `redemption_codes` 表包含 `is_permanent boolean default false` 列，代码中 `RedemptionCode.IsPermanent` 字段已经正确使用（参见 `redemption_handler.go:300`）

---

### F-006: VIP过期任务日志输出优化

| 项目 | 内容 |
|------|------|
| **文件** | `backend/worker/vip_expiry.go` |
| **问题** | 建议增加日志输出，便于排查问题 |
| **影响** | 运维可观测性 |
| **风险等级** | 🟢 低 |
| **前后端配合** | 否 |
| **依赖项** | 无 |
| **状态** | ✅ 已优化（v1.2.0） |

**当前实现**:
`backend/internal/worker/vip_expiry.go` 已有完整日志输出：
- `logger.Info("VIP worker started", ...)` — 启动日志
- `logger.Infof("VIP worker processed %d expired VIP users", ...)` — 完成统计
- `logger.Errorf("VIP worker error updating user %d: %v", ...)` — 错误追踪
- `logger.Infof("VIP expired for user %d (%s), downgraded to free tier", ...)` — 每个用户降级记录
```

---

## 三、新增功能项 (🟡 中优先级 - 文档中有设计但未实现) ✅ 全部完成

### N-001: 渠道测试历史记录页面

| 项目 | 内容 |
|------|------|
| **描述** | 新增渠道测试历史记录前端页面 |
| **设计文档** | `docs/design/channel-management-design.md` |
| **状态** | ✅ 已实现(v1.2.0) |

**实现内容**:
1. 后端: 新增 `GET /api/v1/admin/channels/:id/test-history` 端点（AdminHandler.GetTestHistory）
2. 前端: 渠道列表页下拉菜单新增"测试历史"选项，打开 el-drawer 展示测试记录
3. 前置修复: AdminHandler.TestChannel 缺少 testHistoryRepo，导致测试历史未写入

**关键文件**:
- `backend/internal/handler/admin_handler.go` — GetTestHistory 方法
- `backend/internal/router/router.go` — 路由注册
- `frontend/src/api/channel.ts` — getTestHistory API
- `frontend/src/views/admin/channels/List.vue` — 测试历史抽屉组件

---

### N-002: 渠道批量导入

| 项目 | 内容 |
|------|------|
| **描述** | 渠道批量导入(CSV) 功能 |
| **设计文档** | `docs/design/channel-management-design.md` |
| **状态** | ✅ 已实现(v1.2.0) |

**实现内容**:
1. 后端: 新增 `POST /api/v1/admin/channels/import` 端点（AdminHandler.ImportChannels）
2. 前端: 渠道列表页新增"批量导入"按钮，上传 CSV 文件后调用导入 API
3. CSV 列: `name, type, base_url, api_key, models(用|分隔), weight, priority, group_name`

**关键文件**:
- `backend/internal/handler/admin_handler.go` — ImportChannels 方法
- `backend/internal/router/router.go` — 路由注册
- `frontend/src/api/channel.ts` — importChannels API
- `frontend/src/views/admin/channels/List.vue` — 导入对话框组件
- 导出生成CSV文件供下载

**实现步骤**:
1. 后端: 新增 `POST /api/v1/admin/channels/import` 端点
2. 后端: 新增 `GET /api/v1/admin/channels/export` 端点
3. 前端: 渠道列表页添加"导入/导出"按钮

---

### N-003: 渠道分组管理

| 项目 | 内容 |
|------|------|
| **描述** | 渠道分组(默认/高级/备用)的独立管理功能 |
| **设计文档** | `docs/design/channel-management-design.md` |
| **状态** | ✅ 基础支持已存在 |

**当前状态**: 渠道模型有 `group_name` 字段，`channel_group_relations` 表已存在，已支持筛选。如需独立分组CRUD页面则需额外开发。

---

## 四、文档更新项 (🟢 低优先级 - 文档与实际同步)

### D-001: 更新文档实现状态

| 文档 | 当前标注 | 实际状态 | 状态 |
|------|---------|---------|------|
| `docs/features/README.md` | 兑换码 ✅ 已实现 | ✅ 85%完成 | ✅ 已同步 |
| `docs/README.md` (项目) | 渠道前端 ⚠️ 部分 | ✅ 基本完整 | ✅ 已同步 |
| `docs/features/` 支付模块 | 微信支付 ❌ 未集成 | 实际未集成 | ✅ 已标注 |

**更新说明**:
- 兑换码状态已在 `features/README.md` 中标注为 ✅（v1.2.0 重构时同步）
- docs/README.md 的 v5.0 索引已覆盖渠道前端状态
- 微信支付标注已在 `features/README.md` 中更新为 ❌

---

## 五、风险控制措施

### 5.1 修复实施原则

| 原则 | 说明 |
|------|------|
| 独立测试 | 每个修复单独测试后再合并 |
| 回滚准备 | 修复前记录当前代码状态(git commit) |
| 最小改动 | 只改必要行，不做无关重构 |
| 验证完整 | 修复后执行完整功能测试 |
| 前后端配合 | 后端修改时同步确认前端配合方案 |

### 5.2 修复检查清单

| 阶段 | 检查项 |
|------|--------|
| 修复前 | [ ] 确认修复文件路径和行号准确 |
| 修复前 | [ ] 确认前后端配合方案已沟通 |
| 修复中 | [ ] 不删除任何现有代码(注释除外) |
| 修复后(后端) | [ ] 运行 `go build` 无编译错误 |
| 修复后(后端) | [ ] 运行 `lsp_diagnostics` 无警告 |
| 修复后(前端) | [ ] 运行 `npm run build` 无错误 |
| 验证后 | [ ] 功能测试通过 |

---

## 六、任务执行状态

```
第1阶段: 紧急修复 ✅ 全部完成（v1.1.1~v1.2.0）
├── F-001: GORM FOR UPDATE 语法修复 ✅
├── F-002: 兑换码审计日志补全 ✅
├── F-004: 验证Signup Config服务 ✅
└── F-003: 注册流程配置读取 + 前后端配合 ✅
    ├── 后端: auth_service.go 改造 + 错误码定义 ✅
    ├── 后端: 用户模型添加IP字段 + 仓库方法 ✅
    ├── 前端: Register.vue 错误处理改造 ✅
    └── 前端: 新增 RegisterClosed.vue 页面 ✅

第2阶段: 功能完善 ✅ 全部完成（v1.2.0）
├── F-005: 永久VIP字段验证 ✅
└── F-006: VIP任务日志优化 ✅

第3阶段: 新增功能 ✅ v1.2.0
├── N-001: 渠道测试历史页面 (前后端) ✅
├── N-002: 渠道批量导入 (前后端) ✅
└── N-003: 渠道分组管理评估 ✅ 基础支持已存在

第4阶段: 文档同步 ✅ 已完成
└── D-001: 文档状态更新 ✅
```

---

## 七、任务状态追踪表

| 编号 | 任务名称 | 类型 | 风险 | 前端配合 | 依赖 | 状态 | 实施日期 | 实施人 |
|------|----------|------|-----|---------|-----|------|---------|--------|
| F-001 | GORM FOR UPDATE语法修复 | 修复 | 🔴高 | 否 | - | ✅ v1.1.1 | 2026-05-05 | - |
| F-002 | 兑换码审计日志补全 | 修复 | 🟡中 | 否 | - | ✅ v1.1.1 | 2026-05-05 | - |
| F-004 | 验证Signup Config服务 | 验证 | 🟡中 | 否 | - | ✅ v1.2.0 | 2026-05-12 | - |
| F-003a | 注册配置后端改造 | 修复 | 🔴高 | 否 | F-004 | ✅ v1.2.0 | 2026-05-12 | - |
| F-003b | 用户模型IP字段+仓库方法 | 修复 | 🟡中 | 否 | F-003a | ✅ v1.2.0 | 2026-05-12 | - |
| F-003c | 注册前端错误处理 | 修复 | 🟡中 | 是 | F-003a | ✅ v1.2.0 | 2026-05-12 | - |
| F-003d | 注册关闭专用页面 | 新增 | 🟡中 | 是 | F-003a | ✅ v1.2.0 | 2026-05-12 | - |
| F-005 | 永久VIP字段验证 | 验证 | 🟡中 | 否 | - | ✅ v1.2.0 | 2026-05-12 | - |
| F-006 | VIP任务日志优化 | 优化 | 🟢低 | 否 | - | ✅ v1.2.0 | 2026-05-12 | - |
| N-001 | 渠道测试历史页面 | 新增 | 🟡中 | 是 | - | ✅ v1.2.0 | 2026-05-12 | TOM |
| N-002 | 渠道批量导入 | 新增 | 🟡中 | 是 | - | ✅ v1.2.0 | 2026-05-12 | TOM |
| N-003 | 渠道分组管理评估 | 评估 | 🟢低 | - | - | ✅ 基础支持已存在（v1.2.0） | 2026-05-12 | - |
| D-001 | 文档状态更新 | 文档 | 🟢低 | 否 | - | ✅ 已同步 | 2026-05-12 | - |

---

## 八、安全与业务策略补充说明

### 8.1 注册关闭策略的业务场景

| 场景 | 触发条件 | 处理策略 |
|------|---------|----------|
| 维护升级 | 系统维护期间 | 临时关闭，显示维护公告 |
| 恶意注册 | 发现机器人/垃圾账号 | 紧急关闭，保留管理员注册 |
| 运营控制 | 达到注册指标上限 | 阶段性关闭，需提前公告 |
| 内测阶段 | 产品内测中 | 仅邀请码开放 |

### 8.2 安全措施清单

| 措施 | 说明 |
|------|------|
| IP限制 | 同一IP 24小时内最多注册N个账号 |
| 邮箱域名限制 | 仅允许特定域名注册(如企业邮箱) |
| 频率限制 | 注册接口限流(5次/分钟) |
| 验证码 | 强制滑块验证，防止机器人 |
| 密码强度 | 最小长度+复杂度要求 |
| 审计日志 | 所有注册操作记录审计日志 |

### 8.3 注册关闭时的用户体验

```
┌─────────────────────────────────────────────────────────────┐
│                    注册已关闭                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🔒 抱歉，当前不支持新用户注册                               │
│                                                             │
│  可能原因:                                                  │
│  • 系统维护升级中                                           │
│  • 已达到注册上限                                           │
│  • 暂时关闭注册                                             │
│                                                             │
│  如您需要账号，请:                                          │
│  1. 联系管理员申请注册                                       │
│  2. 使用已有账号登录                                         │
│                                                             │
│  [返回首页]                        [联系管理员]            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 九、审核确认

| 角色 | 签字 | 日期 |
|------|------|------|
| 审核人 | | |
| 实施人 | | |
| 测试人 | | |

---

**文档版本**: 2.1  
**创建日期**: 2026-05-05  
**更新说明**: N-001/N-002 已实现（v1.2.0）；修复 AdminHandler.TestChannel 缺失 testHistoryRepo bug  
**下次更新**: 渠道导出功能增加后