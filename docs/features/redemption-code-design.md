# 兑换码功能设计方案

> 版本: v1.0
> 日期: 2026-04-01
> 状态: 待审核

---

## 1. 功能概述

兑换码功能允许管理员生成不同类型的兑换码，用户通过输入兑换码获取相应的配额或VIP权益。适用于营销活动、用户奖励、线下推广等场景。

## 2. 业务模型

### 2.1 兑换码类型

| 类型 | 说明 | 用途 |
|------|------|------|
| quota | 配额兑换 | 赠送额度 |
| vip | VIP兑换 | 开通会员 |
| recharge | 充值兑换 | 直接充值 |
| mixed | 混合类型 | 配额+VIP组合 |

### 2.2 兑换码使用限制

| 限制类型 | 说明 |
|----------|------|
| single_use | 一次性，使用后失效 |
| multi_use | 多次使用，指定次数 |
| unlimited | 无限次使用 |

### 2.3 有效期控制

- 发行日期 - 开始可用时间
- 截止日期 - 到期自动失效
- 有效天数 - 发行后N天内有效

## 3. 数据模型

### 3.1 redemption_codes 表

```sql
CREATE TABLE redemption_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,          -- 兑换码
    code_type VARCHAR(20) NOT NULL,            -- quota/vip/recharge/mixed
    usage_limit INTEGER DEFAULT 1,             -- 使用次数限制
    usage_count INTEGER DEFAULT 0,             -- 已使用次数
    quota_amount BIGINT DEFAULT 0,            -- 配额数量
    vip_package_id INTEGER,                   -- VIP套餐ID（关联vip_packages）
    vip_days INTEGER,                        -- VIP天数
    valid_from TIMESTAMP,                     -- 有效期开始
    valid_until TIMESTAMP,                    -- 有效期结束
    batch_id VARCHAR(50),                     -- 批次号
    status VARCHAR(20) DEFAULT 'active',      -- active/disabled/expired
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### 3.2 redemption_usage 表

```sql
CREATE TABLE redemption_usage (
    id SERIAL PRIMARY KEY,
    code_id INTEGER REFERENCES redemption_codes(id),
    user_id INTEGER REFERENCES users(id),
    quota_granted BIGINT DEFAULT 0,
    vip_granted BOOLEAN DEFAULT false,
    vip_days INTEGER DEFAULT 0,
    redeemed_at TIMESTAMP DEFAULT NOW(),
    ip_address VARCHAR(50),
    user_agent TEXT
);
```

### 3.3 模型定义

```go
type RedemptionCode struct {
    ID           uint
    Code         string    // 兑换码
    CodeType     string    // quota/vip/recharge/mixed
    UsageLimit   int       // 使用次数限制
    UsageCount   int       // 已使用次数
    QuotaAmount  int64     // 配额数量
    VIPPackageID *uint    // VIP套餐ID
    VipDays     int       // VIP天数
    ValidFrom   *time.Time // 有效期开始
    ValidUntil  *time.Time // 有效期结束
    BatchID     string    // 批次号
    Status      string    // active/disabled/expired
    CreatedBy   uint
    CreatedAt   time.Time
}

type RedemptionUsage struct {
    ID           uint
    CodeID      uint
    UserID      uint
    QuotaGranted int64
    VipGranted  bool
    VipDays     int
    RedeemedAt  time.Time
    IPAddress   string
    UserAgent   string
}
```

## 4. 功能模块

### 4.1 兑换码生成（管理员）

**功能**：
- 批量生成兑换码
- 设置兑换码类型和权益
- 设置使用限制和有效期
- 生成批次管理

**表单字段**：
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 兑换码前缀 | 输入框 | 是 | 如 "VIP2024"，自动生成后缀 |
| 生成数量 | 数字 | 是 | 1-1000 |
| 码类型 | 下拉框 | 是 | 配额/VIP/充值/混合 |
| 配额数量 | 数字 | 配额类型必填 | 赠送的Token数量 |
| VIP套餐 | 下拉框 | VIP类型必填 | 关联VIP套餐 |
| VIP天数 | 数字 | VIP类型选填 | 覆盖套餐默认天数 |
| 使用次数 | 下拉框 | 是 | 一次性/多次/无限 |
| 有效期 | 日期选择 | 是 | 开始-结束日期 |

**生成规则**：
```
前缀 + 时间戳 + 随机字符
示例: VIP20260401A7B3C9D
```

### 4.2 兑换码管理

**功能**：
- 查看所有兑换码
- 按批次/类型/状态筛选
- 查看使用统计
- 禁用/启用兑换码
- 导出兑换码列表

**列表字段**：
| 字段 | 说明 |
|------|------|
| 兑换码 | 脱敏显示：VIP****3D |
| 类型 | 配额/VIP/充值 |
| 配额/权益 | 具体内容 |
| 使用次数 | 已使用/限制 |
| 有效期 | 日期范围 |
| 状态 | 激活/禁用/已过期 |
| 操作 | 查看/禁用/复制 |

### 4.3 兑换码兑换（用户）

**功能**：
- 输入兑换码
- 验证兑换码有效性
- 兑换成功后发放权益

**UI设计**：
```
┌─────────────────────────────────────────┐
│         🎁 兑换码兑换                   │
├─────────────────────────────────────────┤
│  请输入您的兑换码：                      │
│  ┌─────────────────────────────────┐   │
│  │ VIP20260401A7B3C9D            │   │
│  └─────────────────────────────────┘   │
│                                         │
│  [立即兑换]                             │
│                                         │
│  ─────────────────────────────────────  │
│  兑换说明：                              │
│  • 每个兑换码仅限使用一次                │
│  • 兑换码有效期至2024年12月31日         │
│  • 最终解释权归平台所有                  │
└─────────────────────────────────────────┘
```

### 4.4 兑换记录

**管理员查看**：
- 按兑换码查看使用记录
- 按用户查看兑换历史
- 统计各批次使用率

**用户查看**：
- 我的兑换记录
- 显示兑换时间和获得权益

## 5. API设计

### 5.1 管理端API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/admin/redemption/codes` | GET | 获取兑换码列表 |
| `/api/v1/admin/redemption/codes` | POST | 生成兑换码 |
| `/api/v1/admin/redemption/codes/:id` | GET | 获取兑换码详情 |
| `/api/v1/admin/redemption/codes/:id/disable` | POST | 禁用兑换码 |
| `/api/v1/admin/redemption/codes/:id/usage` | GET | 查看使用记录 |
| `/api/v1/admin/redemption/batches` | GET | 获取批次列表 |
| `/api/v1/admin/redemption/export` | POST | 导出兑换码 |

### 5.2 用户端API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/redemption/redeem` | POST | 兑换兑换码 |
| `/api/v1/redemption/history` | GET | 获取兑换历史 |

### 5.3 请求/响应示例

**生成兑换码**：
```json
POST /api/v1/admin/redemption/codes
{
  "prefix": "VIP2024",
  "count": 100,
  "code_type": "vip",
  "vip_package_id": 3,
  "usage_limit": 1,
  "valid_from": "2024-04-01",
  "valid_until": "2024-12-31"
}
```

**用户兑换**：
```json
POST /api/v1/redemption/redeem
{
  "code": "VIP20260401A7B3C9D"
}
```

```json
// 成功响应
{
  "success": true,
  "data": {
    "message": "兑换成功",
    "reward": {
      "type": "vip",
      "vip_level": "vip_gold",
      "vip_days": 30,
      "expire_at": "2024-05-01"
    }
  }
}
```

## 6. 用户流程

### 6.1 管理员生成兑换码

```
1. 进入"兑换码管理"页面
2. 点击"+ 生成兑换码"
3. 选择兑换码类型和权益
4. 设置使用限制和有效期
5. 填写生成数量
6. 点击"生成"
7. 系统批量生成并保存
8. 显示生成的兑换码列表
9. 可选择导出或复制
```

### 6.2 用户兑换流程

```
1. 用户登录平台
2. 进入"兑换码"页面或弹窗
3. 输入兑换码
4. 点击"兑换"
5. 系统验证：
   - 兑换码是否存在
   - 是否在有效期内
   - 是否已被使用
   - 用户是否已兑换过（单用户限制）
6. 验证通过 → 发放权益
7. 记录兑换日志
8. 显示兑换成功
9. 更新用户配额/VIP状态
```

## 7. 业务规则

### 7.1 兑换规则

1. **有效期检查**：必须在 valid_from 和 valid_until 之间
2. **使用次数检查**：usage_count < usage_limit
3. **状态检查**：status 必须为 active
4. **用户限制**：同一用户不能重复兑换（可选配置）

### 7.2 权益发放规则

| 类型 | 发放内容 |
|------|----------|
| quota | 增加用户 remain_quota |
| vip | 更新用户 level/vip_expired_at/vip_package_id |
| recharge | 增加用户 vip_quota |
| mixed | 同时发放配额和VIP |

### 7.3 VIP续费规则

- 如果用户已是VIP：累加VIP天数
- 如果用户不是VIP：开通VIP，设置到期时间
- VIP配额：设置套餐默认配额

## 8. 安全考虑

- 兑换码生成使用加密随机算法
- 兑换操作记录IP地址和UserAgent
- 防止暴力破解：同一IP请求频率限制
- 兑换码脱敏显示在列表中
- 批量生成需要权限验证

## 9. 后续扩展

- [ ] 兑换码分享海报生成
- [ ] 兑换码到期提醒
- [ ] 兑换码使用统计报表
- [ ] 渠道分销码（带渠道追踪）

---

## 审核意见

（待填写）
