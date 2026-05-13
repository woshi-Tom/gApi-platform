# Changelog

所有重要更新都会记录在此文件。

---

## [v1.3.0] - 2026-05-14

### Features
- **Anthropic Messages API 兼容**: 新增 `POST /v1/messages` 端点，支持 Claude SDK 直接对接；x-api-key 认证中间件兼容 Anthropic 格式；Claude 流式 SSE 响应完整支持
- **SSE 流式首包延迟修复**: 优化流式响应首包时间，保留 channel failover 能力
- **Provider Adapter 扩展**: 新增 Gemini、Azure adapter，factory 模式统一管理 6 种 AI 提供商（OpenAI/Claude/DeepSeek/NVIDIA/Gemini/Azure）

### Fixes
- **K01 微信支付死代码清理**: 移除 settings 页面微信支付表单和后端 WechatConfig 残留
- **K04 Release 分平台打包**: linux-amd64 包不再包含 windows 二进制
- **K05 审计日志 skipPaths**: 改为前缀匹配，避免详情页重复记录
- **K06 settings_handler 错误处理**: 修复忽略 DB error 的问题
- **K07~K10 文档补全**: 审计日志状态同步、升级指南、发布检查清单、回滚策略
- **K11 Redis 密码泄露**: docker-compose 改用 redis.conf 配置文件
- **K12 AutoMigrate 约束**: 已迁移数据库不再报约束错误（日志降级）
- **K13 ability_types 类型不匹配**: `parseAbilityTypes()` 兼容字符串和数组两种格式
- **TokenAuth nil panic**: x-api-key 测试中 tokenService 为 nil 时添加 recover 保护

### License
- 添加 GPLv3 许可证

### Tests
- Anthropic 兼容测试：转换函数 + SSE 流式 + x-api-key 认证（#014 两轮全部通过）

---

## [v1.2.1] - 2026-05-12

### Fixes
- **CORS 环境变量加载**: 修复 `GAPI_ADMIN_FRONTEND_URL` 环境变量未加载导致管理后台跨域请求失败
- **Model 表自动迁移**: `AutoMigrate` 注册 `ModelGroup`、`ChannelGroupRelation`、`UserGroupRelation`、`ModelPricing` 四个模型，修复新部署时模型管理功能不可用
- **Response 类型清理**: 删除 `model/response.go` 中未使用的 `Pagination`/`PageData` 类型，添加文件注释说明类型职责
- **配置文件清理**: `.env.example` 移除 `config.go` 不支持的微信支付环境变量
- **死代码清理**: 删除未使用的前端组件 `SliderCaptcha.vue`、`HelloWorld.vue`；移除 `router.go` 中冗余的 `_ = idempRepo`
- **日志中间件去重**: `gin.Default()` 改为 `gin.New()` + `gin.Recovery()`，避免与审计日志重复输出请求日志

---

## [v1.2.0] - 2026-05-12

### Breaking Changes
- **审计日志 GET 过滤**: 审计中间件默认跳过 GET 请求的记录（重要路径如 `/api/v1/payment/` 除外），以减少数据膨胀。如需记录所有 GET 请求，需修改中间件配置。
- **Settings 页面字段调整**: 移除注册设置中的 `smtp_enabled` 字段（后端无此字段）；微信支付配置字段从 UI 中移除。

### Features
- **系统设置全栈 API**: 新增通用设置、速率限制、安全设置 3 组设置项，前后端完整实现
- **模型管理前端页面**: 新增模型分组/定价/权限的前端管理界面
- **审计日志增强**: 新增递归敏感字段脱敏（支持嵌套 JSON 和数组）

### CI/CD
- **GitHub Actions 升级**: 修复因Node.js 20弃用导致的release打包失败
- **Release 包名优化**: 添加平台后缀（`*-linux-amd64.tar.gz` / `*-windows-amd64.zip`）
- **Release Notes 优化**: 改为内联说明，展示文件适用平台和快速部署指引
- **修复非tag触发时的VERSION提取**: 支持branch push时安全获取版本号
- **Git 协作规范**: 新增Release发布流程与Tag管理规范
- **Action 版本锁定**: 所有 GitHub Actions 锁定到不可变 commit SHA
- **Release 产物校验**: 新增 SHA256SUMS 校验文件

### Tests
- **新增单元测试**: 审计中间件 22 个测试 + Settings Service 8 个测试 + Settings Handler 8 个测试，共 38 个

### Fixes
- **Audit goroutine 稳定性**: 异步审计日志写入添加 panic recover，防止进程崩溃
- **用户分组 No Data**: 修复模型管理页面用户分组列表因 `data?.list` 取值错误导致的数据不显示

---

## [v1.1.1] - 2026-05-05

### Added
- **注册配置增强**: 支持邮箱域名限制、IP注册限制（24小时内）、密码最小长度配置
- **注册奖励**: 支持 quota（免费配额）和 vip（VIP配额）两种奖励类型
- **用户IP追踪**: 记录用户注册时的IP地址，用于安全审计
- **渠道BaseURL自动填充**: 选择渠道类型时自动填充默认API地址
- **渠道健康状态**: 列表页显示健康状态指示灯、最后检测时间、响应时间
- **API密钥删除防抖**: 防止快速点击导致误删其他密钥

### Fixed
- **GORM语法修复**: FOR UPDATE 使用正确的 `gorm:"query_option"` 标签
- **兑换码审计日志**: 修复类型错误，补全 Create/Redeem 操作审计记录
- **注册错误处理**: 添加错误码映射，前端正确显示注册关闭提示
- **配置文件安全**: admin密码改用bcrypt哈希存储

### Changed
- **数据库迁移**: signup_configs 表新增字段（allowed_domains, min_password_length, signup_reward_type, signup_reward_amount）

---

## [v1.1.0] - 2026-04-13

### Added
- **渠道SOCKS代理支持**: 渠道配置支持通过SOCKS5/HTTP代理访问API
- **渠道测试历史记录**: 保存每次渠道测试的请求/响应记录
- **自动健康检测**: 服务启动时自动启动健康检测定时任务，每5分钟检测一次

### Fixed
- 渠道健康检测定时任务未启动的问题
- API使用量日志未实际记录的问题

### Refactored
- VIP过期检查Worker日志改进为结构化日志

---

## [v1.0.0] - 2026-03-23

### Added
- 多租户架构
- VIP用户体系（30天过期）
- 商品购买（支付宝支付）
- 渠道管理（多渠道负载均衡）
- 完整审计日志
- 兑换码功能
- 邮箱验证注册
