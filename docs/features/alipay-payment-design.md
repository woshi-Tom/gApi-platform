# 支付宝当面付扫码支付设计方案

**版本**: v1.0  
**日期**: 2026-03-30  
**状态**: 设计中

---

## 1. 需求概述

### 1.1 业务流程

```
用户选择套餐 → 创建订单 → 支付宝预下单(生成二维码) → 用户扫码支付 → 支付宝异步通知 → 系统发放配额
```

### 1.2 核心功能

| 功能 | 说明 |
|------|------|
| 创建支付订单 | 生成支付宝预支付订单，返回二维码 |
| 扫码支付 | 用户用支付宝APP扫码支付 |
| 异步回调 | 支付宝通知支付结果 |
| 配额发放 | 支付成功后自动增加用户配额 |
| 订单状态查询 | 用户查看支付状态 |

---

## 2. 团队评审

### 2.1 产品经理

**核心流程确认**：
- ✅ 用户选择套餐 → 点击购买
- ✅ 生成支付二维码（有效期15分钟）
- ✅ 用户手机支付宝扫码支付
- ✅ 支付成功自动发放配额
- ✅ 订单页面实时显示支付状态

**用户体验优化**：
- 倒计时显示二维码剩余有效期
- 支付成功后跳转提示
- 订单列表显示支付状态

### 2.2 安全工程师

**安全审查**：

| 安全项 | 方案 | 状态 |
|--------|------|------|
| 签名验证 | RSA2 签名，支付宝公钥验签 | ✅ |
| 幂等性 | 根据 trade_no 查询已处理，防止重复发放 | ✅ |
| 回调验签 | 使用支付宝SDK验签 | ✅ |
| 敏感数据 | 私钥不返回前端，配置仅服务端存储 | ✅ |
| 金额校验 | 订单金额与支付金额必须一致 | ✅ |
| 权限控制 | 用户只能查询自己的订单 | ✅ |

**威胁模型**：
- **伪造支付成功**：必须验签回调 + 金额校验
- **重放攻击**：trade_no 记录已处理，防止重复发放
- **恶意查询**：用户只能查询自己订单

### 2.3 后端开发工程师

**技术选型**：

| 组件 | 选型 | 说明 |
|------|------|------|
| SDK | `github.com/smartwalle/alipay` | Star最多，维护活跃 |
| 签名 | RSA2 | 支付宝推荐 |
| 语言 | Go | 项目现有语言 |

**API 接口设计**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/payment/alipay/create` | POST | 创建支付订单 |
| `/api/v1/payment/alipay/query/:order_no` | GET | 查询订单状态 |
| `/api/v1/payment/alipay/cancel/:order_no` | POST | 取消订单 |
| `/api/v1/payment/alipay/notify` | POST | 支付宝异步回调 |
| `/api/v1/payment/alipay/page` | GET | 支付页面(返回二维码) |

### 2.4 前端 UI 工程师

**支付页面设计**：

```
┌─────────────────────────────────────┐
│         支付订单 #20260330001        │
├─────────────────────────────────────┤
│                                     │
│         ┌─────────────┐             │
│         │             │             │
│         │   二维码    │             │
│         │             │             │
│         └─────────────┘             │
│                                     │
│    金额: ¥100.00                    │
│    套餐: 100,000 Token              │
│                                     │
│    ⏱ 剩余支付时间: 14:32           │
│                                     │
│    请使用支付宝APP扫码支付            │
│                                     │
│    ┌─────────────────────────┐     │
│    │ [  ] 已完成支付? 查看状态 │     │
│    └─────────────────────────┘     │
│                                     │
└─────────────────────────────────────┘
```

**UX 考量**：
- 倒计时提醒剩余支付时间
- 支付成功/失败的明确提示
- 订单号显示便于用户查询

---

## 3. 数据库设计

### 3.1 复用现有表结构

**orders 表** (已有)：
```sql
- order_no          -- 订单号
- status            -- pending|paid|cancelled|refunded|expired
- pay_amount        -- 支付金额
```

**payments 表** (已有)：
```sql
- payment_no       -- 支付流水号
- channel_order_no  -- 支付宝订单号 (trade_no)
- channel_trade_no  -- 支付宝交易号
- qr_code          -- 二维码 URL
- status           -- pending|success|failed|refunded
```

### 3.2 新增字段

```sql
ALTER TABLE orders ADD COLUMN alipay_trade_no VARCHAR(64);
ALTER TABLE orders ADD COLUMN alipay_qr_url TEXT;
ALTER TABLE orders ADD COLUMN qr_expire_at TIMESTAMP;
```

---

## 4. API 设计

### 4.1 创建支付订单

**请求**：
```json
POST /api/v1/payment/alipay/create
{
  "order_id": 123
}
```

**响应**：
```json
{
  "success": true,
  "data": {
    "order_no": "ORD202603300001",
    "qr_code": "https://qr.alipay.com/xxx",
    "qr_expire_at": "2026-03-30T12:30:00Z",
    "amount": 100.00
  }
}
```

### 4.2 查询订单状态

**请求**：
```json
GET /api/v1/payment/alipay/query/ORD202603300001
```

**响应**：
```json
{
  "success": true,
  "data": {
    "order_no": "ORD202603300001",
    "status": "pending",  // pending|paid|expired
    "amount": 100.00,
    "paid_at": null
  }
}
```

### 4.3 支付宝回调

**支付宝通知格式**：
```json
{
  "trade_no": "支付宝交易号",
  "out_trade_no": "商家订单号",
  "trade_status": "TRADE_SUCCESS",
  "total_amount": "100.00",
  "receipt_amount": "100.00"
}
```

**响应**：
```
success (表示成功接收)
```

---

## 5. 安全实现

### 5.1 签名验证流程

```
1. 接收支付宝回调参数
2. 移除 sign、sign_type 字段
3. 字典序排序所有参数
4. 用 & 连接成字符串
5. 使用支付宝公钥验签
6. 验证通过后处理业务
```

### 5.2 幂等性处理

```go
func HandleAlipayCallback(tradeNo string) error {
    // 1. 查询是否已处理
    payment, _ := GetPaymentByTradeNo(tradeNo)
    if payment.Status == "success" {
        return nil // 已处理，直接返回
    }
    
    // 2. 处理支付逻辑
    // 3. 更新状态为 success
    // 4. 发放配额
}
```

### 5.3 金额校验

```go
func HandleAlipayCallback(params map[string]string) error {
    orderAmount := order.PayAmount
    paidAmount, _ := strconv.ParseFloat(params["total_amount"], 64)
    
    if paidAmount != orderAmount {
        return errors.New("金额不匹配")
    }
    
    // 继续处理...
}
```

---

## 6. 配置设计

### 6.1 管理后台配置

| 字段 | 说明 | 示例 |
|------|------|------|
| 启用支付宝 | 开关 | true/false |
| APP ID | 支付宝应用ID | 2021xxxxxx |
| 商家私钥 | RSA2私钥 | -----BEGIN RSA PRIVATE KEY-----... |
| 支付宝公钥 | 支付宝公钥 | -----BEGIN PUBLIC KEY-----... |
| 回调地址 | 支付回调URL | https://api.example.com/api/v1/payment/alipay/notify |

### 6.2 环境变量 (备选)

```env
ALIPAY_APP_ID=2021xxxxxx
ALIPAY_PRIVATE_KEY=xxx
ALIPAY_PUBLIC_KEY=xxx
ALIPAY_NOTIFY_URL=https://api.example.com/api/v1/payment/alipay/notify
```

---

## 7. SDK 使用

### 7.1 安装 SDK

```bash
go get github.com/smartwalle/alipay/v2
```

### 7.2 初始化

```go
import "github.com/smartwalle/alipay/v2"

alipayClient, _ := alipay.New(
    "APP_ID",
    "商家私钥",
    alipay.RSA2,
    alipay.WithPublicKey("支付宝公钥"),
)
```

### 7.3 预下单

```go
bizContent := map[string]interface{}{
    "out_trade_no": orderNo,
    "total_amount": fmt.Sprintf("%.2f", amount),
    "subject":      subject,
    "product_code": "FACE_TO_FACE_PAYMENT",
}

resp, err := alipayClient.TradePrecreate()
    .WithBizContent(bizContent)
    .Execute()
```

---

## 8. 实施计划

### 阶段1: 后端基础 (P0)
- [ ] 安装 Alipay SDK
- [ ] 创建 AlipayService
- [ ] 实现预下单接口
- [ ] 实现回调验签处理
- [ ] 配额发放逻辑

### 阶段2: 前端页面 (P0)
- [ ] 支付页面组件
- [ ] 订单状态轮询
- [ ] 支付成功提示

### 阶段3: 管理配置 (P1)
- [ ] 支付宝配置管理
- [ ] 沙箱/生产切换

### 阶段4: 测试验证 (P0)
- [ ] 沙箱测试
- [ ] 回调验签测试
- [ ] 幂等性测试

---

## 9. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 回调失败 | 配额未发放 | 幂等性设计 + 订单状态检查 |
| 网络超时 | 用户重复点击 | 防重复提交 + 订单状态检查 |
| 签名伪造 | 资金损失 | 必须验签 |
| 二维码过期 | 需重新下单 | 前端倒计时 + 过期提示 |

---

## 10. 验收标准

- [ ] 用户可创建支付订单
- [ ] 显示可扫码支付的二维码
- [ ] 支付宝回调正确处理
- [ ] 支付成功后配额自动发放
- [ ] 订单状态正确更新
- [ ] 防重复发放
- [ ] 沙箱测试通过

---

## 11. 参考资料

- [支付宝当面付文档](https://opendocs.alipay.com/open/8ad49e4a_alipay.trade.precreate)
- [smartwalle/alipay SDK](https://github.com/smartwalle/alipay)
- [支付宝验签方法](https://opendocs.alipay.com/open/291/105765)
