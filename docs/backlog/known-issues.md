# 已知问题

> 已确认但未排期修复的问题。按严重度排序。

---

## 🔴 高

| # | 问题 | 来源 | 说明 | 状态 |
|---|------|------|------|------|
| K14 | Gemini 适配器首条消息 panic（C-01） | #016 全面审查 | `gemini.go:52-64` content 为空时 `contents[-1]` 触发 index-out-of-range | ⬜ 待修复 |
| K15 | 审计日志 goroutine gin.Context 数据竞争（H-01） | #016 全面审查 | `audit.go:94-162` go func 读取已回收的 gin.Context | ⬜ 待修复 |
| K16 | API 访问日志 goroutine gin.Context 数据竞争（H-02） | #016 全面审查 | `api_access_log.go:29-96` go func 读取已回收的 gin.Context | ⬜ 待修复 |
| K17 | 7 个适配器静默吞没错误（H-03） | #016 全面审查 | DeepSeek/Zhipu/Baidu/Yi/Ollama/LocalAI/Groq 适配器所有 error 返回 `_` | ⬜ 待修复 |

## 🟡 中

| # | 问题 | 来源 | 说明 | 状态 |
|---|------|------|------|------|
| K01 | 微信支付 UI 字段待移除 | v1.2.0 审查 | settings 页面有微信支付表单但后端不接入，需清理 | ✅ 已修复 (清理 WechatConfig/死代码) |
| K02 | AES 密钥长度未校验 | v1.2.0 审查 | `crypto/aes.go` 弱密钥被静默填充 | ✅ 已修复 (a235b95) |
| K03 | Dockerfile 以 root 运行 | v1.2.0 审查 | 容器逃逸风险 | ✅ 已修复 (a235b95) |
| K04 | Release 命名与实际内容不匹配 | v1.2.0 审查 | linux-amd64 包内含 windows 二进制 | ✅ 已修复 (分平台打包) |
| K05 | Audit skipPaths 精确匹配 | v1.2.0 审查 | 审计日志详情页仍会记日志 | ✅ 已修复 (前缀匹配) |
| K06 | settings_handler 忽略 error | v1.2.0 审查 | 忽略 DB error 存在数据丢失风险 | ✅ 已修复 (e28396c) |
| K07 | 审计日志状态标注矛盾 | v1.2.0 审查 | features/README 与 CHANGELOG 不一致 | ✅ 已修复 |
| K08 | 无升级指南 | v1.2.0 审查 | 从旧版本升级缺少步骤文档 | ✅ 已修复 (upgrade-v1.2.0.md) |
| K09 | 无发布检查清单 | v1.2.0 审查 | 发布流程缺少可执行 checklist | ✅ 已修复 (git-workflow.md 第8章) |
| K10 | 无回滚策略文档 | v1.2.0 审查 | git-workflow 缺少回滚章节 | ✅ 已修复 (git-workflow.md 第7章) |
| K18 | VIP 状态不一致（H-04） | #016 全面审查 | `isVIPUser` 与 `isValidVIP` 对 nil VIPExpiredAt 判断相反 | ⬜ 待修复 |
| K19 | InitHandler 始终报告 RabbitMQ 断开（M-03） | #016 全面审查 | router.go:75 始终传 nil MQ client | ⬜ 待修复 |
| K20 | Admin UpdateChannel 重置 Status/Priority（M-04） | #016 全面审查 | 发送 `{"name": "..."}` 会将 status 重置为 0 禁用 | ⬜ 待修复 |
| K21 | SSEReader 不处理 CRLF 行尾（M-01） | #016 全面审查 | Windows 本地 AI 服务器可能发 CRLF 导致 SSE 解析失败 | ⬜ 待修复 |
| K22 | SIGTERM Windows 静默忽略（M-02） | #016 全面审查 | taskkill 无法优雅关闭，仅 Ctrl+C 有效 | ⬜ 待修复 |
| K23 | deploy/release: JWT/ENCRYPT_KEY 无默认值（H-05） | #016 全面审查 | `.env` 缺失时替换为空字符串 | ⬜ 待修复 |
| K24 | deploy/release: Redis requirepass 不可配（H-06） | #016 全面审查 | 改 REDIS_PASSWORD 不会改 Redis 密码 | ⬜ 待修复 |
| K25 | deploy/release: HEALTHCHECK 用 curl 但未安装（H-09） | #016 全面审查 | 容器健康检查和 depends_on 均失败 | ⬜ 待修复 |
| K26 | CI: release-readme.txt 不存在（H-07） | #016 全面审查 | release 流水线 sed 失败 | ⬜ 待修复 |
| K27 | deploy/docker: config.yaml 卷挂载源不存在（H-08） | #016 全面审查 | Docker 创建空目录覆盖目标路径 | ⬜ 待修复 |
| K28 | deploy-nginx.sh: 调用未定义函数（H-10） | #016 全面审查 | `create_ssl_certs` 不存在，`set -euo pipefail` 脚本崩溃 | ⬜ 待修复 |

## 🟢 低

| # | 问题 | 来源 | 说明 | 状态 |
|---|------|------|------|------|
| K11 | Redis 密码泄露于进程列表 | v1.2.0 审查 | docker-compose.yml 命令行传密码 | ✅ 已修复 (redis.conf 配置文件) |
| K12 | AutoMigrate 约束操作非幂等 | #005 部署 | GORM AutoMigrate 对已迁移数据库重跑时，尝试 DROP 不存在的约束（`uni_model_groups_name`）导致启动失败 | ✅ 已修复 (database.go 日志降级) |
| K13 | 模型定价 ability_types 类型不匹配 | 用户测试 | 编辑模型定价保存时，前端用 `<el-input>` 以字符串发送 `ability_types`（如 `"chat,completion"`），但后端 handler（`model_pricing_handler.go:48`）用 `[]string` 接收期望 JSON 数组（如 `["chat","completion"]`）。报错 `json: cannot unmarshal string into Go struct field .ability_types of type []string` | ✅ 已修复 (parseAbilityTypes 兼容两种格式) |
| K29 | `os.IsNotExist` 不兼容 wrapped errors（L-01） | #016 全面审查 | Go 1.13+ 应使用 `errors.Is` | ⬜ 待修复 |
| K30 | `BodySizeLimit` 死代码（L-02） | #016 全面审查 | const + 函数从未注册到 router | ⬜ 待修复 |
| K31 | `Log.Path` 配置字段从未使用（L-05） | #016 全面审查 | 文件日志从未实现 | ⬜ 待修复 |
| K32 | 已完成订单用叉号标记（L-08） | #016 全面审查 | U+2717 (✗) 应该用 U+2713 (✓) | ⬜ 待修复 |

---

## 关联文档

fix-plan-2026.md 中的修复任务（F-001~F-006, N-001~N-003, D-001）以 **fix-plan 为唯一数据源**，不在此处重复。  
本文件仅记录审查发现的增量问题。fix-plan 条目请直接查阅 [plans/fix-plan-2026.md](../plans/fix-plan-2026.md)。
