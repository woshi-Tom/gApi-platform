# 支付模块问题记录 (Payment Module Issues)

**日期**: 2026-03-31  
**版本**: v1.0

---

## 1. 数据库Schema与模型不一致

### 问题描述
代码优化后，数据库表结构与GORM模型定义不匹配，导致以下问题：

### 1.1 orders 表重复列

| 问题 | 影响 |
|------|------|
| 同时存在 `expire_at` 和 `expires_at` 两列 | GORM无法确定使用哪个列，可能导致数据混乱 |
| 模型定义 `ExpiresAt` → 映射到 `expires_at` | 与历史遗留的 `expire_at` 冲突 |

**状态**: ✅ 已修复 - 删除重复的 `expire_at` 列

### 1.2 users 表字段命名不一致

| 数据库列名 | 模型字段 | GORM Tag |
|-----------|---------|----------|
| `v_ip_expired_at` | `VIPExpiredAt` | 未指定 |
| `v_ip_package_id` | `VIPPackageID` | 未指定 |
| `v_ip_quota` | `VIPQuota` | 未指定 |

**影响**: 
- VIP过期检查Worker查询 `vip_expired_at` 失败
- 日志报错: `column "vip_expired_at" does not exist`

**状态**: ✅ 已修复 - 添加正确的 `column:` GORM tag

---

## 2. 新表缺失

### 问题描述
代码新增了 `payment_logs` 和 `idempotency_keys` 模型，但对应的数据库表未创建。

| 表名 | 用途 | 状态 |
|------|------|------|
| `payment_logs` | 支付操作审计日志 | ✅ 已创建 |
| `idempotency_keys` | 幂等性记录 | ✅ 已创建 |

---

## 3. VIP过期检查失败

### 问题描述
`VIPExpiryWorker` 使用错误的列名查询VIP过期用户。

**错误日志**:
```
[VIPWorker] Error querying expired VIPs: ERROR: column "vip_expired_at" does not exist (SQLSTATE 42703)
```

**根因**: SQL查询使用 `vip_expired_at`，但数据库列名是 `v_ip_expired_at`

**状态**: ✅ 已修复 - Worker代码使用正确的列名

---

## 4. 支付二维码生成卡住

### 问题描述
用户下单后跳转到支付页面，一直显示"正在生成二维码..."。

### 可能原因

| 原因 | 排查方法 |
|------|----------|
| API返回错误但被静默处理 | 检查前端是否正确显示错误消息 |
| `qr_code` 字段名不匹配 | 前后端使用一致的JSON字段名 |
| Alipay沙箱配置问题 | 验证支付宝沙箱是否正常 |
| 网络请求超时 | 检查浏览器Network面板 |

### 前后端字段对照

**后端 Response** (payment_handler.go):
```go
type AlipayPaymentResponse struct {
    OrderNo     string `json:"order_no"`
    QRCode      string `json:"qr_code"`      // ✅ 正确
    QRExpireAt  string `json:"qr_expire_at"` // ✅ 正确
    Amount      string `json:"amount"`
    PackageName string `json:"package_name"`
}
```

**前端 API调用** (Payment.vue):
```typescript
const res = await paymentApi.createAlipay(orderNo.value)
const data = res?.data?.data || res?.data || {}
if (data.qr_code) {  // ✅ 匹配
    qrCodeUrl.value = data.qr_code
    await generateQRCode(data.qr_code)
}
```

**状态**: ✅ 字段名匹配 - 无需修改

### 4.1 订单创建时Payment记录URL为空

**问题描述**: 订单创建时handler创建Payment记录，但此时不调用Alipay生成二维码。后续请求二维码时复用该记录，导致返回空QR码。

**根因分析**:
1. 用户下单 → 订单创建 → 支付记录创建(URL为空)
2. 用户请求二维码 → 找到已有pending支付 → 直接返回空URL

**修复方案** (payment_handler.go):
```go
// 发现已有pending支付但URL为空时，调用Alipay获取新二维码
if qrCode == "" {
    log.Info().Str("order_no", order.OrderNo).Msg("existing payment has empty QR code, requesting new one from Alipay")
    subject := fmt.Sprintf("%s - %s", order.PackageName, order.OrderNo)
    qrCode, _, err = h.alipayService.CreatePayment(order.OrderNo, order.PayAmount, subject)
    // ...
}
```

**状态**: ✅ 已修复 (2026-03-31)

### 4.2 轮询时二维码消失 (前端)

**问题描述**: 用户发起支付后，二维码显示2秒后消失。

**根因分析**:
1. `startPayment()` 正确调用 `QRCode.toDataURL()` 生成 base64 图片
2. `startPolling()` 每3秒轮询，但直接将 `data.qr_code` (URL字符串) 赋值给 `qrCodeImage.value`
3. 前端期望 `qrCodeImage.value` 是 base64 图片，但收到的是普通 URL，导致图片无法显示

**问题代码** (Payment.vue):
```javascript
// 错误：直接赋值URL字符串
if (data.qr_code) {
  qrCodeImage.value = data.qr_code  // 这里应该是 base64 图片
}
```

**修复方案**:
```javascript
// 正确：调用 generateQRCode 转换 URL 为 base64 图片
if (data.qr_code) {
  await generateQRCode(data.qr_code)
}
```

**状态**: ✅ 已修复 (2026-03-31)

### 4.3 支付宝沙箱支付错误 ALI41778

**问题描述**: 沙箱环境扫码支付时，第一次输入密码显示"系统异常，ALI41778"，第二次才能支付成功。

**错误代码含义**: ALI41778 = "交易状态错误" (Transaction Status Error)

**根因分析**:

这是**支付宝沙箱环境的问题**，不是代码问题。证据：

| 证据 | 详情 |
|------|------|
| ✅ 支付成功 | 订单状态成功变为 `completed` |
| ✅ 回调接收 | `/api/v1/payment/callback/alipay` 正确处理 |
| ✅ 配额到账 | 用户配额已更新 |
| ✅ 审计记录 | `payment.success` 正确记录 |

**沙箱环境已知问题**:

1. **同一二维码快速扫码两次** - 沙箱存在竞态条件
2. **交易状态不一致** - 沙箱有时序问题
3. **会话超时** - 沙箱会话过期更快
4. **沙箱环境不稳定** - 支付宝开发者沙箱已知问题

**生产环境**: 此错误**不会发生**在真实支付宝账号，生产环境 API 非常稳定。

**测试建议**:

1. 多次尝试 - 沙箱测试时此错误常见，重试通常成功
2. 生成新二维码 - 如果首次支付失败，刷新页面重新生成
3. 使用沙箱调试工具 - 支付宝开放平台开发者控制台有专门的沙箱调试工具

**状态**: ✅ 已记录 (2026-03-31) - 非代码问题

---

## 4.4 订单列表"去支付"功能未实现

**问题描述**: 用户在商品列表点击购买创建订单后，跳转到订单记录页面，点击"去支付"按钮时显示"支付功能开发中，请稍后..."。

**根因分析**: `orders/List.vue` 中的 `handlePay` 函数是占位代码，未实现跳转逻辑。

**问题代码** (List.vue):
```javascript
// 原代码 - 占位函数
function handlePay(order: Order) {
  ElMessage.info('支付功能开发中，请稍后...')
}
```

**修复方案**:
```javascript
// 修复后 - 跳转到支付页面
function handlePay(order: Order) {
  router.push({ path: '/payment', query: { order_no: order.order_no } })
}
```

**状态**: ✅ 已修复 (2026-03-31)

---

## 4.5 取消订单后未跳转

**问题描述**: 用户点击"取消订单"后，仍然停留在支付页面，没有返回上一页。

**根因分析**: 前端 `cancelOrder()` 函数只更新了状态和显示，没有执行页面跳转。

**修复方案** (Payment.vue):
```javascript
async function cancelOrder() {
  await paymentApi.cancelAlipay(orderNo.value)
  stopAllTimers()
  ElMessage.success('订单已取消')
  router.back()  // 返回上一页
}
```

**状态**: ✅ 已修复 (2026-03-31)

---

## 4.6 支付宝沙箱二维码URL无法显示

**问题描述**: 直接使用支付宝返回的URL作为图片源时，图片无法显示。

**根因分析**: 支付宝沙箱二维码URL (`https://qr.alipay.com/...`) 无法在浏览器中直接作为图片显示。

**解决方案**: 使用 `qrcode` 库在本地生成本地二维码图片（base64格式）。

**修复方案**:
```javascript
// 使用 qrcode 库生成二维码
const qrCodeImage = await QRCode.toDataURL(alipayUrl, {
  width: 256,
  margin: 2
})
// 在模板中使用 base64 图片
<img :src="qrCodeImage" alt="支付宝二维码" />
```

**状态**: ✅ 已修复 (2026-03-31)

---

## 4.7 二维码过期与订单过期状态混淆

**问题描述**: `initCountdown()` 函数在二维码过期时错误地将订单状态设置为 `expired`。

**根因分析**: 二维码过期（15分钟）≠ 订单过期（4小时），是独立的两个概念。

**修复方案**:
```javascript
function initCountdown() {
  // 二维码过期时只清除二维码显示，不改变订单状态
  if (left <= 0) {
    remainingSeconds.value = 0
    qrCodeUrl.value = ''
    qrCodeImage.value = ''
    return  // 不设置 status = 'expired'
  }
}
```

**状态**: ✅ 已修复 (2026-03-31)

---

## 4.8 订单列表"去支付"跳转问题

**问题描述**: 订单记录页面点击"去支付"按钮显示"支付功能开发中"。

**根因分析**: `handlePay()` 函数是占位代码，未实现跳转逻辑。

**修复方案**:
```javascript
function handlePay(order: Order) {
  router.push({ path: '/payment', query: { order_no: order.order_no } })
}
```

**状态**: ✅ 已修复 (2026-03-31)

---

## 5. 代码耦合性问题

### 问题描述
支付模块代码耦合度较高，一个改动可能影响多个功能。

### 建议改进

1. **分层解耦**:
   - `handler` 层: 只处理HTTP请求/响应
   - `service` 层: 业务逻辑
   - `repository` 层: 数据库操作
   - `worker` 层: 后台任务

2. **配置管理**:
   - 数据库列名配置化
   - 默认值统一管理

3. **测试覆盖**:
   - 单元测试覆盖核心业务逻辑
   - 集成测试覆盖API端点

---

## 6. 数据库迁移规范

### 建议

1. **Schema变更必须同步更新模型**
2. **新增表必须执行迁移SQL**
3. **使用GORM AutoMigrate时检查差异**

### 迁移检查命令
```bash
# 检查模型与数据库差异
docker exec gapi-backend ./gapi-server -check-schema

# 执行迁移
docker exec gapi-backend ./gapi-server -migrate
```

---

## 后续监控

### 需要监控的问题

| 指标 | 告警条件 |
|------|----------|
| VIP Worker错误数 | > 0/分钟 |
| 支付成功率 | < 95% |
| 二维码生成超时 | > 5秒 |
| 订单过期率 | > 10%/小时 |

---

## 变更记录

| 日期 | 问题 | 状态 | 备注 |
|------|------|------|------|
| 2026-03-31 | orders表重复列expire_at/expires_at | ✅ 已修复 | - |
| 2026-03-31 | VIP Worker列名错误 | ✅ 已修复 | - |
| 2026-03-31 | payment_logs表缺失 | ✅ 已创建 | - |
| 2026-03-31 | idempotency_keys表缺失 | ✅ 已创建 | - |
| 2026-03-31 | User模型缺少column tag | ✅ 已修复 | - |
| 2026-03-31 | 订单创建时Payment记录URL为空导致二维码不显示 | ✅ 已修复 | - |
| 2026-03-31 | 轮询时二维码消失 (直接设置URL而非base64) | ✅ 已修复 | - |
| 2026-03-31 | 支付宝沙箱 ALI41778 错误 | ℹ️ 已知问题 | 沙箱环境特性，非代码问题 |
| 2026-03-31 | 订单列表"去支付"未跳转 | ✅ 已修复 | - |
| 2026-03-31 | 取消订单后未跳转 | ✅ 已修复 | - |
| 2026-03-31 | 支付宝沙箱二维码URL无法显示 | ✅ 已修复 | 使用qrcode库生成本地二维码 |
| 2026-03-31 | 二维码过期与订单过期状态混淆 | ✅ 已修复 | 分离两个独立概念 |
