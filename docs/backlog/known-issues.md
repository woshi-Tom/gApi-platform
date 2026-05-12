# 已知问题

> 已确认但未排期修复的问题。按严重度排序。

---

## 🔴 高

（暂无未排期的高严重度问题）

## 🟡 中

| # | 问题 | 来源 | 说明 | 状态 |
|---|------|------|------|------|
| K01 | 微信支付 UI 字段待移除 | v1.2.0 审查 | settings 页面有微信支付表单但后端不接入，需清理 | ⬜ |
| K02 | AES 密钥长度未校验 | v1.2.0 审查 | `crypto/aes.go` 弱密钥被静默填充 | ⬜ |
| K03 | Dockerfile 以 root 运行 | v1.2.0 审查 | 容器逃逸风险 | ⬜ |
| K04 | Release 命名与实际内容不匹配 | v1.2.0 审查 | linux-amd64 包内含 windows 二进制 | ⬜ |
| K05 | Audit skipPaths 精确匹配 | v1.2.0 审查 | 审计日志详情页仍会记日志 | ⬜ |
| K06 | settings_handler 忽略 error | v1.2.0 审查 | 忽略 DB error 存在数据丢失风险 | ⬜ |
| K07 | 审计日志状态标注矛盾 | v1.2.0 审查 | features/README 与 CHANGELOG 不一致 | ⬜ |
| K08 | 无升级指南 | v1.2.0 审查 | 从旧版本升级缺少步骤文档 | ⬜ |
| K09 | 无发布检查清单 | v1.2.0 审查 | 发布流程缺少可执行 checklist | ⬜ |
| K10 | 无回滚策略文档 | v1.2.0 审查 | git-workflow 缺少回滚章节 | ⬜ |

## 🟢 低

| # | 问题 | 来源 | 说明 | 状态 |
|---|------|------|------|------|
| K11 | Redis 密码泄露于进程列表 | v1.2.0 审查 | docker-compose.yml 命令行传密码 | ⬜ |
| K12 | AutoMigrate 约束操作非幂等 | #005 部署 | GORM AutoMigrate 对已迁移数据库重跑时，尝试 DROP 不存在的约束（`uni_model_groups_name`）导致启动失败 | ⬜ |

---

## 关联文档

fix-plan-2026.md 中的修复任务（F-001~F-006, N-001~N-003, D-001）以 **fix-plan 为唯一数据源**，不在此处重复。  
本文件仅记录审查发现的增量问题。fix-plan 条目请直接查阅 [plans/fix-plan-2026.md](../plans/fix-plan-2026.md)。
