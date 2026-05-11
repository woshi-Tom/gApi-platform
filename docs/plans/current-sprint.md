# 当前 Sprint

> 最后更新: 2026-05-12  
> 目标: v1.2.0 发布  
> 状态: 🟡 整改中

---

## 背景

v1.2.0 通过 6 角色并行审查，发现 4 个 P0 阻塞问题。当前处于整改阶段。

---

## 当前工作项

### P0 — 发布前必须完成

| # | 任务 | 文件 | 负责人 | 状态 |
|---|------|------|--------|------|
| 1 | CI pipeline 添加测试/lint | `.github/workflows/build.yml` | 开发 | ⬜ |
| 2 | settings 页面存根替换为真实 API | `settings/Index.vue` + router + handler | 开发+前端 | ⬜ |
| 3 | 新增 handler/middleware/service 测试 | `*_test.go` | 开发 | ⬜ |
| 4 | build 作业添加 `permissions: contents: read` | `.github/workflows/build.yml` | 开发 | ⬜ |

### 待决策

| # | 事项 | 提给谁 | 状态 |
|---|------|--------|------|
| D1 | 审查报告已生成，需要推送到远程供其他智能体评审 | 用户 | ⏳ 等待 push |
| D2 | 确认 P0 修复后是否重新触发全角色审查 | 用户 | ⬜ |

---

## 完成标准

- [ ] 4 个 P0 问题全部修复
- [ ] review-report.md 状态：🔴 → ✅
- [ ] v1.2.0 正式发布
