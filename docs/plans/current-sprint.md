# 当前 Sprint

> 最后更新: 2026-05-12
> 目标: v1.2.0 发布
> 状态: ✅ 整改完成，正在合入 main

---

## 背景

v1.2.0 通过 6 角色并行审查 + 1 轮外部评审，汇总发现 **5 个 P0 阻塞问题**（原始 4 个 + audit goroutine recover 从 P1 升级）。

**结果**: 5 个 P0 全部修复完成，全角色重新审查已通过。测试环境已搭建验证通过。

---

## 已完成工作项

### P0 — 全部修复 ✅

| # | 任务 | 文件 | 负责人 | 状态 |
|---|------|------|--------|------|
| 1 | CI pipeline 添加测试/lint | `.github/workflows/build.yml` | 开发 | ✅ |
| 2 | settings 页面存根替换为真实 API | `settings/Index.vue` + router + handler | 开发+前端 | ✅ |
| 3 | 新增 handler/middleware/service 测试 | `*_test.go` (38 tests) | 开发 | ✅ |
| 4 | build 作业添加 `permissions: contents: read` | `.github/workflows/build.yml` | 开发 | ✅ |
| 5 | Audit goroutine 添加 panic recover | `audit.go` | 开发 | ✅ |

### P1 — 已修复 ✅

| # | 任务 | 优先级 | 状态 |
|---|------|--------|------|
| 06 | Action 版本锁 SHA | P1 | ✅ |
| 07 | Release 产物 checksum | P1 | ✅ |
| 08 | Settings handler 输入验证 | P1 | ✅ |
| 09 | maskSensitiveData 递归脱敏 | P1 | ✅ |
| 10 | CHANGELOG 模型管理页面条目 | P1 | ✅ |
| 11 | CHANGELOG Breaking Changes 章节 | P1 | ✅ |
| 12 | settings_handler 忽略 error | P1 | ✅ |

### 过程中发现的修复

| # | 任务 | 文件 | 状态 |
|---|------|------|------|
| 13 | 用户分组列表因 `data?.list` 取值错误导致 No Data | `admin/models/Index.vue` | ✅ |

---

## ✅ 已决策

| # | 事项 | 决策 | 决策人 |
|---|------|------|--------|
| D1 | P0 修复后是否重新触发全角色审查 | ✅ 全角色重新审查一遍 | 用户 |
| D2 | 本轮范围 | ✅ P0 硬目标 + P1 随行就市 | 用户 |

---

## 重新审查结论

| 角色 | 结论 |
|------|------|
| 🔒 安全 | ✅ 通过 |
| 📋 产品经理 | ✅ 通过 |
| 📊 项目经理 | ✅ 通过 |
| 💻 开发 | ✅ 通过 |
| 🧪 QA/测试 | ⚠️ 条件通过 |
| 🎨 UI/前端 | ✅ 通过 |

---

## 完成标准

- [x] 5 个 P0 问题全部修复
- [x] 7 个 P1 问题全部修复
- [x] 通过全角色重新审查（6 角色再评一轮）
- [x] v1.2.0 正式发布（合入 main）
