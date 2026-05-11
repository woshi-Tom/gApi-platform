# 支付模块设计文档

**版本**: v2.2  
**日期**: 2026-03-31  
**状态**: 已实现

---

## 1. 概述

本文档描述支付模块的完整设计方案，包括订单状态机、超时机制、幂等性保证、错误处理和后台任务。

### 1.1 支持的支付场景

| 场景 | 类型 | 说明 |
|------|------|------|
| VIP 套餐购买 | 订阅 | 购买后获得 VIP 资格和配额 |
| 充值套餐购买 | 永久 | 购买后获得永久配额 |

### 1.2 支付流程

```
用户下单 → 生成订单(pending) → 发起支付 → 扫码支付 → 支付成功回调 → 发放配额 → 订单完成
                ↓
           取消订单 → 订单取消
                ↓
           超时未付 → 订单过期
```

### 1.3 用户交互流程

#### 1.3.1 完整支付流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          用户端支付交互流程                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  [商品列表] ──点击购买──→ [创建订单] ──跳转──→ [支付页面]               │
│                                         │                               │
│                                         ├── 扫码支付 ──→ [支付成功]       │
│                                         │         ↓                      │
│                                         │    [订单记录] (自动刷新)        │
│                                         │         ↓                      │
│                                         │    [配额到账]                  │
│                                         │                               │
│                                         ├── 取消订单 ──→ [返回上一页]      │
│                                         │         ↓                      │
│                                         │    [商品列表]                  │
│                                         │                               │
│                                         └── 超时过期 ──→ [订单已过期]     │
│                                                      ↓                  │
│                                                 [重新购买]               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 1.3.2 页面状态与操作

| 页面 | 订单状态 | 显示内容 | 可用操作 |
|------|----------|----------|----------|
| 支付页面 | pending | 二维码 + 倒计时 | 取消订单 |
| 支付页面 | paid | 支付成功提示 | 查看配额 |
| 支付页面 | cancelled | 订单已取消 | 返回商品列表 |
| 支付页面 | expired | 订单已过期 | 重新购买 |
| 订单记录 | pending | 待支付 | 去支付 / 详情 |
| 订单记录 | completed | 已完成 | 详情 |

#### 1.3.3 页面跳转规则

| 操作 | 跳转目标 | 说明 |
|------|----------|------|
| 购买 → 创建订单 | /payment?order_no=xxx | 跳转支付页面 |
| 订单记录 → 去支付 | /payment?order_no=xxx | 跳转支付页面 |
| 取消订单 | router.back() | 返回上一页（商品列表/订单记录） |
| 支付成功 | /profile | 查看配额页面 |
| 订单已过期 → 重新购买 | /products | 返回商品列表 |
| 订单已取消 → 重新购买 | /products | 返回商品列表 |

---

## 2. 订单状态机

### 2.1 状态定义

| 状态 | 说明 | 订单颜色 |
|------|------|----------|
| `pending` | 待支付 | 黄色 |
| `paid` | 已付款(待发货) | 蓝色 |
| `completed` | 已完成(配额已发放) | 绿色 |
| `cancelled` | 已取消 | 灰色 |
| `expired` | 已过期 | 红色 |
| `refunded` | 已退款 | 紫色 |

### 2.2 状态转换图

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│    ┌─────────┐                                          │
│    │ pending │ ← 创建订单                                 │
│    └────┬────┘                                          │
│         │                                               │
│    ┌────┼────┬────────────────┐                         │
│    │    │    │                │                         │
│    ↓    ↓    ↓                ↓                         │
│ ┌──────┐ ┌──────────┐    ┌──────────┐                  │
│ │ paid │ │cancelled │    │ expired  │                  │
│ └──┬───┘ └──────────┘    └──────────┘                  │
│    │                                                     │
│    ├── 配额发放成功 ──→ ┌──────────┐                     │
│    │                   │completed │                     │
│    │                   └──────────┘                     │
│    │                       │                            │
│    └── 退款 ──→ ┌──────────┐                            │
│                │ refunded │                             │
│                └──────────┘                            │
└─────────────────────────────────────────────────────────┘
```

### 2.3 有效状态转换

| 当前状态 | 目标状态 | 触发条件 | 操作 |
|----------|----------|----------|------|
| pending | paid | 支付成功回调/轮询确认 | 更新 paid_at |
| pending | cancelled | 用户取消 | 更新 cancel_reason |
| pending | expired | 超时未支付 | 标记为过期 |
| paid | completed | 配额发放成功 | 标记为完成 |
| paid | refunded | 退款成功 | 退款处理 |
| completed | refunded | 退款 | 扣除配额/取消VIP |

---

## 3. 超时机制

### 3.1 超时时间表

| 操作 | 超时时间 | 说明 |
|------|----------|------|
| 二维码有效期 | 15 分钟 | Alipay 标准，不可调整 |
| 订单待支付过期 | 4 小时 | 建议值，可后台配置 |
| API 调用超时 | 30 秒 | 10s 连接 + 20s 响应 |
| 取消重试间隔 | 5s → 10s → 20s | 指数退避 |
| 取消最大重试次数 | 3 次 | 超过后标记待人工处理 |
| 幂等性窗口 | 10 分钟 | 防止重复下单 |

### 3.2 过期订单处理

```
每 5 分钟定时任务扫描:
  SELECT * FROM orders 
  WHERE status = 'pending' 
    AND expires_at <= NOW()
  
  → 更新 status = 'expired'
  → 记录审计日志
  → 发送通知(可选)
```

---

## 4. 幂等性设计

### 4.1 幂等键设计

```go
type IdempotencyKey struct {
    Key        string    // user_id + ":" + action + ":" + timestamp
    OrderID    uint      // 关联订单
    CreatedAt  time.Time
    ExpiresAt  time.Time
}
```

### 4.2 幂等处理流程

```
1. 客户端生成幂等键: sha256(user_id + order_no + timestamp)[:16]
2. 请求携带 Header: X-Idempotency-Key: <key>
3. 服务端检查:
   - 键存在且未过期 → 返回原订单信息
   - 键不存在 → 创建订单并记录键
   - 键过期 → 删除旧键，创建新订单
```

### 4.3 数据库约束

```sql
-- 幂等键表
CREATE TABLE idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(64) UNIQUE NOT NULL,
    order_id VARCHAR(50),
    user_id BIGINT NOT NULL,
    action VARCHAR(32),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '10 minutes')
);

CREATE INDEX idx_idempotency_keys_expires ON idempotency_keys(expires_at);
```

---

## 5. 错误处理

### 5.1 错误分类

| 错误类型 | 错误码 | 是否可重试 | 处理方式 |
|----------|--------|------------|----------|
| 参数错误 | `INVALID_PARAMS` | 否 | 返回 400 |
| 订单不存在 | `ORDER_NOT_FOUND` | 否 | 返回 404 |
| 订单已支付 | `ORDER_ALREADY_PAID` | 否 | 返回已支付状态 |
| 支付超时 | `PAYMENT_TIMEOUT` | 是 | 重试 3 次 |
| 支付宝服务不可用 | `ALIPAY_UNAVAILABLE` | 是 | 降级/重试 |
| 签名验证失败 | `SIGNATURE_INVALID` | 否 | 返回 401 |
| 金额不匹配 | `AMOUNT_MISMATCH` | 否 | 人工处理 |

### 5.2 重试策略

```go
// 指数退避重试
retryConfig := retry.Config{
    MaxAttempts: 3,
    InitialDelay: 5 * time.Second,
    MaxDelay: 60 * time.Second,
    Multiplier: 2.0,
}

// 不可重试错误
nonRetryable := []string{
    "INVALID_PARAMS",
    "ORDER_NOT_FOUND", 
    "ORDER_ALREADY_PAID",
    "SIGNATURE_INVALID",
}
```

### 5.3 补偿机制

```
支付成功但配额发放失败:
1. 标记订单为 partial_completed
2. 记录失败原因
3. 后台任务重试发放
4. 超过 3 次失败 → 人工介入
```

---

## 6. 数据库设计

### 6.1 orders 表改动

```sql
-- 新增字段
ALTER TABLE orders ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN version INT DEFAULT 1;
ALTER TABLE orders ADD COLUMN payment_id BIGINT REFERENCES payments(id);

-- 添加索引
CREATE INDEX idx_orders_expires ON orders(status, expires_at) 
    WHERE status = 'pending';
```

### 6.2 payments 表改动

```sql
-- 新增字段
ALTER TABLE payments ADD COLUMN idempotency_key VARCHAR(64);
ALTER TABLE payments ADD COLUMN retry_count INT DEFAULT 0;
ALTER TABLE payments ADD COLUMN last_retry_at TIMESTAMPTZ;
```

### 6.3 payment_logs 表

```sql
CREATE TABLE payment_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    payment_id BIGINT,
    action VARCHAR(32) NOT NULL,  -- create|notify|query|cancel|refund
    status VARCHAR(16),
    request_data TEXT,
    response_data TEXT,
    error_message TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payment_logs_order ON payment_logs(order_id);
CREATE INDEX idx_payment_logs_created ON payment_logs(created_at);
```

### 6.4 quota_transactions 表

```sql
CREATE TABLE quota_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,  -- vip_activate|quota_add|quota_deduct|refund
    amount BIGINT NOT NULL,
    order_id VARCHAR(50),
    balance_before BIGINT,
    balance_after BIGINT,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_quota_trans_user ON quota_transactions(user_id, created_at DESC);
```

---

## 7. API 设计

### 7.1 创建订单

```
POST /api/v1/orders
Headers:
  Authorization: Bearer <token>
  X-Idempotency-Key: <unique-key>

Request:
{
  "package_id": 1,
  "package_type": "vip",      // vip | recharge
  "payment_method": "alipay"  // alipay | wechat
}

Response (201):
{
  "success": true,
  "data": {
    "order_id": 123,
    "order_no": "ORD20260331abc12345",
    "expires_at": "2026-03-31T12:00:00Z",
    "status": "pending"
  }
}
```

### 7.2 发起支付

```
POST /api/v1/payment/alipay
Headers:
  Authorization: Bearer <token>
  X-Idempotency-Key: <unique-key>

Request:
{
  "order_no": "ORD20260331abc12345"
}

Response (200):
{
  "success": true,
  "data": {
    "order_no": "ORD20260331abc12345",
    "qr_code": "https://qr.alipay.com/...",
    "qr_expire_at": "2026-03-31T11:15:00Z",
    "expires_at": "2026-03-31T16:00:00Z"
  }
}
```

### 7.3 查询订单状态

```
GET /api/v1/payment/alipay/query/:order_no
Headers:
  Authorization: Bearer <token>

Response (200):
{
  "success": true,
  "data": {
    "order_no": "ORD20260331abc12345",
    "status": "pending",  // pending|paid|completed|cancelled|expired
    "qr_code": "...",
    "qr_expire_at": "...",
    "paid_at": null
  }
}
```

### 7.4 取消订单

```
POST /api/v1/payment/alipay/cancel/:order_no
Headers:
  Authorization: Bearer <token>

Response (200):
{
  "success": true,
  "data": {
    "order_no": "ORD20260331abc12345",
    "status": "cancelled",
    "cancel_reason": "user cancelled"
  }
}
```

### 7.5 支付宝回调

```
POST /api/v1/payment/callback/alipay
Content-Type: application/x-www-form-urlencoded

Response: "success" (只返回 success 字符串)
```

---

## 8. 后台任务

### 8.1 任务列表

| 任务名 | 频率 | 功能 |
|--------|------|------|
| `expiry_worker` | 每 5 分钟 | 过期待支付订单 |
| `reconcile_worker` | 每 5 分钟 | 同步订单与支付宝状态 |
| `quota_retry_worker` | 每 10 分钟 | 重试失败的配额发放 |
| `cleanup_worker` | 每天凌晨 | 清理过期日志 |

### 8.2 expiry_worker

```go
func (w *ExpiryWorker) Run() error {
    var orders []model.Order
    
    // 查询超时的待支付订单
    w.db.Where("status = ? AND expires_at <= ?", 
        "pending", time.Now()).Find(&orders)
    
    for _, order := range orders {
        w.db.Model(&order).Updates(map[string]interface{}{
            "status": "expired",
        })
        
        // 记录审计日志
        w.auditLog.Create(&model.AuditLog{
            Action: "order.expired",
            OrderID: order.ID,
        })
    }
    
    return nil
}
```

### 8.3 reconcile_worker

```go
func (w *ReconcileWorker) Run() error {
    var orders []model.Order
    
    // 查询已支付但未完成的订单
    w.db.Where("status = ?", "paid").Find(&orders)
    
    for _, order := range orders {
        result, err := w.alipay.Query(order.OrderNo)
        if err != nil {
            w.handleError(order, err)
            continue
        }
        
        switch result.TradeStatus {
        case "TRADE_SUCCESS", "TRADE_FINISHED":
            if order.Status != "completed" {
                w.deliverQuota(order)
            }
        case "TRADE_CLOSED":
            w.db.Model(&order).Update("status", "cancelled")
        }
    }
    
    return nil
}
```

---

## 9. 安全措施

### 9.1 签名验证

```go
func VerifyAlipayNotify(params map[string]string, publicKey string) bool {
    sign := params["sign"]
    signType := params["sign_type"]
    
    // 移除 sign 和 sign_type
    delete(params, "sign")
    delete(params, "sign_type")
    
    // 按字典序排序
    keys := make([]string, 0, len(params))
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    // 拼接待签名字符串
    var str bytes.Buffer
    for _, k := range keys {
        str.WriteString(k)
        str.WriteString("=")
        str.WriteString(params[k])
        str.WriteString("&")
    }
    str.Truncate(str.Len() - 1)
    
    // 验证签名
    return verify(str.String(), sign, publicKey)
}
```

### 9.2 限流

```go
// 每用户每分钟限制
rateLimitConfig := ratelimit.Config{
    Requests: 10,
    Window: time.Minute,
    KeyFunc: func(c *gin.Context) string {
        userID, _ := c.Get("user_id")
        return fmt.Sprintf("payment:%d", userID)
    },
}
```

### 9.3 金额校验

```go
func ValidateAmount(order *model.Order, paidAmount float64) error {
    tolerance := 0.01 // 1分钱容差
    diff := math.Abs(order.PayAmount - paidAmount)
    if diff > tolerance {
        return fmt.Errorf("amount mismatch: expected %.2f, got %.2f", 
            order.PayAmount, paidAmount)
    }
    return nil
}
```

---

## 10. 监控指标

### 10.1 Prometheus 指标

```go
var (
    ordersTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "payment_orders_total",
            Help: "Total number of orders by status",
        },
        []string{"status", "type"},
    )
    
    paymentLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "payment_processing_seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"action"},
    )
    
    alipayErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "alipay_errors_total",
            Help: "Total Alipay API errors",
        },
        []string{"method", "error_code"},
    )
)
```

### 10.2 告警规则

```yaml
groups:
  - name: payment_alerts
    rules:
      - alert: HighPendingOrderExpiryRate
        expr: rate(payment_orders_expired_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
          
      - alert: AlipayHighErrorRate
        expr: rate(alipay_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
          
      - alert: QuotaDeliveryFailure
        expr: rate(quota_delivery_failures_total[5m]) > 0
        for: 2m
        labels:
          severity: warning
```

---

## 11. 实施计划

### 11.1 P0 阶段 (核心功能)

| 任务 | 优先级 | 工作量 | 状态 |
|------|--------|--------|------|
| 添加 expires_at 字段 | P0 | 1h | pending |
| 实现订单过期逻辑 | P0 | 2h | pending |
| 实现幂等性保证 | P0 | 3h | pending |
| 实现 expiry_worker | P0 | 2h | pending |
| 订单状态枚举完善 | P0 | 1h | pending |

### 11.2 P1 阶段 (可靠性)

| 任务 | 优先级 | 工作量 | 状态 |
|------|--------|--------|------|
| 实现 reconcile_worker | P1 | 3h | pending |
| 完善取消订单重试 | P1 | 2h | pending |
| 支付宝错误处理优化 | P1 | 2h | pending |

### 11.3 P2 阶段 (完善)

| 任务 | 优先级 | 工作量 | 状态 |
|------|--------|--------|------|
| 添加 payment_logs 表 | P2 | 1h | pending |
| 添加 quota_transactions 表 | P2 | 2h | pending |
| 添加监控指标 | P2 | 2h | pending |
| 清理任务 | P2 | 1h | pending |

---

## 12. 变更记录

| 版本 | 日期 | 修改内容 | 作者 |
|------|------|----------|------|
| v1.0 | 2026-03-23 | 初始版本 | - |
| v2.0 | 2026-03-31 | 完善超时、幂等、后台任务设计 | - |
