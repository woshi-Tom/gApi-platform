# v1.2.0 整改跟踪

> 基于审查报告提取的 P0/P1/P2 问题跟踪表。  
> 来源: [review-report.md](review-report.md)  
> 状态: ✅ P0 全部关闭，等待发布

---

## P0 — 发布前必须修复 ✅

| # | 问题 | 领域 | 措施 | 责任人 | 状态 | 完成日期 |
|---|------|------|------|--------|------|---------|
| 01 | CI pipeline 无测试/lint 步骤 | DevOps | 添加 `go test`、`go vet` 到 build.yml | 开发 | ✅ | 2026-05-12 |
| 02 | settings 页面 3 个 save 是存根 | Frontend | 实现真实 API 调用（general/ratelimit/security） | 开发+前端 | ✅ | 2026-05-12 |
| 03 | 核心变更文件零测试覆盖 | Backend | 新增 settings_handler/audit/settings_service 测试 | 开发 | ✅ | 2026-05-12 |
| 04 | GITHUB_TOKEN 权限过大 | DevOps | build 作业添加 `permissions: contents: read` | 开发 | ✅ | 2026-05-12 |
| 05 | Audit goroutine 无 recover | Backend | 添加 `defer recover()` | 开发 | ✅ | 2026-05-12 |

## P1 — 建议发布前修复

| # | 问题 | 领域 | 措施 | 责任人 | 状态 | 完成日期 |
|---|------|------|------|--------|------|---------|
| 06 | Action 版本未锁 SHA | DevOps | 锁定到 commit SHA | 开发 | ⬜ | - |
| 07 | Release 产物无 checksum | DevOps | 生成 SHA256SUMS | 开发 | ⬜ | - |
| 08 | Settings handler 无输入验证 | Backend | 添加 binding validation | 开发 | ⬜ | - |
| 09 | maskSensitiveData 脱敏不完整 | Backend | 实现递归脱敏 | 开发 | ⬜ | - |
| 10 | CHANGELOG 遗漏模型管理页面 | Docs | 追加条目 | 开发 | ⬜ | - |
| 11 | CHANGELOG 无 Breaking Changes | Docs | 增加章节 | 开发 | ⬜ | - |
| 12 | settings_handler 忽略 error | Backend | 处理 error（nil 指针风险） | 开发 | ⬜ | - |

## P2 — 后续版本修复

| # | 问题 | 领域 | 措施 | 责任人 | 状态 |
|---|------|------|------|--------|------|
| 13 | 微信支付 UI 字段待移除 | Frontend | 从 settings 页面删除微信支付字段 | 前端 | ✅ |
| 14 | 表单无验证规则 | Frontend | 添加 `:rules` | 前端 | ⬜ |
| 15 | CI 无前端 lint/typecheck | DevOps | 添加 `vue-tsc --noEmit` | 前端 | ⬜ |
| 16 | AES 密钥未校验 | Backend | Init 增加密钥长度检查 | 开发 | ⬜ |
| 17 | Dockerfile 以 root 运行 | Deploy | 添加 USER 指令 | 开发 | ⬜ |
| 18 | Release 命名与实际不符 | DevOps | 改文件名或分平台打包 | 开发 | ⬜ |
| 19 | 无 Smoke Test | DevOps | 增加冒烟测试步骤 | 开发 | ⬜ |
| 20 | skipPaths 精确匹配 | Backend | 改为前缀匹配 | 开发 | ⬜ |
| 21 | Audit 日志状态矛盾 | Docs | 统一 features/README 标注 | 开发 | ✅ |
| 22 | 无升级指南 | Docs | 创建 upgrade-v1.2.0.md | 开发 | ⬜ |
| 23 | 无发布检查清单 | Docs | git-workflow 增加 checklist | 开发 | ⬜ |
| 24 | 无回滚策略文档 | Docs | git-workflow 新增第 7 章 | 开发 | ⬜ |

## P3 — 低优先级

| # | 问题 | 领域 | 措施 | 责任人 | 状态 |
|---|------|------|------|--------|------|
| 25 | Redis 密码泄露于进程列表 | Deploy | 改用配置文件 | 开发 | ⬜ |
