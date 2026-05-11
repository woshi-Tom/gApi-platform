# 功能文档总览

> 所有功能设计文档（含已实现、待完善、未实现），按模块分组。

## 状态说明

| 图标 | 含义 |
|------|------|
| ✅ | 已完整实现（后端 + 前端） |
| ⚠️ | 部分实现或功能不完整 |
| ❌ | 未实现 |
| 🚧 | 设计中/开发中 |

---

## 支付模块

| 功能 | 文档 | 后端 | 前端 | 说明 |
|------|------|------|------|------|
| 支付宝支付 | [alipay-payment-design.md](alipay-payment-design.md) | ✅ | ✅ | 完整集成 |
| 支付模块 | [payment-module-design.md](payment-module-design.md) | ✅ | ✅ | 订单+支付流程 |
| 支付修复日志 | [payment-module-fix-log.md](payment-module-fix-log.md) | ✅ | ✅ | 历史修复记录 |
| 支付问题记录 | [payment-module-issues.md](payment-module-issues.md) | ✅ | ✅ | 已知问题 |
| 商品/套餐 | [business-package-spec.md](business-package-spec.md) | ✅ | ✅ | 免费/充值/VIP 规格 |
| 微信支付 | — | ❌ | ❌ | 代码中 `wechat_enabled: false` |

## 用户模块

| 功能 | 文档 | 后端 | 前端 | 说明 |
|------|------|------|------|------|
| 注册/登录/Token | — | ✅ | ✅ | `auth_service.go` + `token_service.go` |
| 邮箱验证 | [email-verification-design.md](email-verification-design.md) | ✅ | ✅ | 注册邮箱验证 |
| SMTP 配置 | [smtp-config-design.md](smtp-config-design.md) | ✅ | ✅ | 邮件服务配置 |
| 滑块验证码 | — | ✅ | ✅ | `slider_captcha_service.go` |
| 用户 API 监控 | [user-api-monitor-design.md](user-api-monitor-design.md) | ✅ | ✅ | 用户调用统计 |
| 注册配置 | [signup-config-design.md](signup-config-design.md) | ✅ | ✅ | 域名/IP/密码/奖励配置 |

## 渠道模块

| 功能 | 文档 | 后端 | 前端 | 说明 |
|------|------|------|------|------|
| 渠道 CRUD | [channel-management-design.md](channel-management-design.md) | ✅ | ✅ | 完整管理界面 |
| 渠道健康检测 | [channel-health-check-design.md](channel-health-check-design.md) | ✅ | ⚠️ | 状态显示+手动检测已有，缺少统计页面和配置 |

## 模型管理

| 功能 | 文档 | 后端 | 前端 | 说明 |
|------|------|------|------|------|
| 模型分组/定价/权限 | [model-group-pricing-permission-design.md](model-group-pricing-permission-design.md) | ✅ | ✅ | 完整 CRUD + 定价 + 用户组 |

## 兑换码

| 功能 | 文档 | 后端 | 前端 | 说明 |
|------|------|------|------|------|
| 兑换码 | [redemption-code-design.md](redemption-code-design.md) | ✅ | ✅ | 生成/兑换/禁用/批次管理 |

## 系统/运维

| 功能 | 文档 | 后端 | 前端 | 说明 |
|------|------|------|------|------|
| 审计日志 | [audit-log-optimization-design.md](audit-log-optimization-design.md) | ⚠️ | ⚠️ | 轻量列表已做，GET过滤已加，数据清理/分表未做 |
| CI/CD 构建 | — | ✅ | — | GitHub Actions 自动构建 + Release |
| 跨平台发布 | — | ✅ | — | Linux tgz + Windows zip |
