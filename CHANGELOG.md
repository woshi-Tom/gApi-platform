# Changelog

所有重要更新都会记录在此文件。

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
