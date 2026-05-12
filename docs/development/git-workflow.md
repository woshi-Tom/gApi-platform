# Git 协作规范

> 适用版本: v1.1 | 最后更新: 2026-05-12

---

## ⚡ 快速参考卡

```text
【提交格式】 <type>: <祈使语气描述>
【分支结构】 main  ←  dev-a(你)  +  dev-b(朋友)
【每日铁律】 开工前/推送前各 rebase 一次 main
【合入门槛】 对方 Review 通过 + CI 通过
【分支寿命】 理想 1-2 天，最长不超 1 周
【发布流程】 dev 测试 → main 合入 → main 打 tag → 自动 Release
```

---

## 1. 分支策略

```
main                 ← 稳定，随时可部署（两人共同维护）
├── dev-<你的标识>   ← 你随便折腾
└── dev-<朋友标识>   ← 朋友随便折腾
```

**规则**：
- 永远不在 `main` 上直接改代码
- `main` 始终保持绿色（能编译、测试通过）
- 分支存活尽量短，最长 1 周
- 每天至少 rebase 一次 main
- 合入 main 前必须互相 review

**命名规范**：

| 场景 | 格式 | 示例 |
|------|------|------|
| 个人开发 | `dev-<标识>` | `dev-tom` |
| 功能分支 | `feat-<功能名>` | `feat-wechat-payment` |
| 紧急修复 | `hotfix-<描述>` | `hotfix-channel-timeout` |

---

## 2. 提交信息规范

### 2.1 格式

```
<type>(<可选作用域>): <简短描述>

<可选: 详细说明>
<可选: Closes #42>
```

一行原则：大多数提交一句话搞定。一行不够时，空一行再补充。

### 2.2 type 列表

| type | 何时用 | 示例 |
|------|--------|------|
| `feat` | 新功能 | `feat: 添加支付宝扫码支付` |
| `fix` | 修 bug | `fix: 渠道测试对话框关闭后模型列表丢失` |
| `docs` | 文档 | `docs: 添加 Git 协作规范` |
| `style` | 代码格式（不影响逻辑） | `style: 统一 Go 结构体命名` |
| `refactor` | 重构 | `refactor: 抽取渠道健康检查为独立服务` |
| `perf` | 性能优化 | `perf: 缓存渠道列表减少数据库查询` |
| `test` | 测试 | `test: 添加用户登录集成测试` |
| `chore` | 构建/工具/依赖 | `chore: 升级 vite 到 6.x` |
| `ci` | CI/CD | `ci: 修复 GitHub Actions 权限问题` |
| `security` | 安全 | `security: 加固 JWT 密钥生成` |

### 2.3 作用域（可选）

修改涉及特定模块时加作用域：`feat(channel): 批量导入功能`

| 作用域 | 模块 |
|--------|------|
| `channel` | 渠道管理 |
| `payment` | 支付模块 |
| `auth` | 认证/登录 |
| `user` | 用户管理 |
| `product` | 商品管理 |
| `admin` | 管理后台 |
| `dashboard` | 仪表盘 |
| `api` | API 代理/路由 |
| `deploy` | 部署相关 |
| `deps` | 依赖升级 |

### 2.4 反例

```text
❌ "fix bug"               → 太模糊
❌ "修改了一些文件"         → 等于没写
❌ "wip"                   → 合入前 squash 掉
❌ "feat: 添加A顺便修了B"   → 一次只做一件事
```

---

## 3. 工作流程

### 每日流程

```text
① git checkout main && git pull --rebase
② git checkout dev-a && git rebase main     # 开工前同步
③ 开发 → git commit（每个独立功能点一交）
④ git fetch origin main && git rebase origin/main  # 推送前再同步
⑤ 如有 wip 提交 → git rebase -i HEAD~N  squash 掉
⑥ git push origin dev-a
⑦ GitHub 创建 PR (dev-a → main)
⑧ 对方 review → 改 → push 更新 PR → review 通过 → Merge
```

### rebase vs merge

| 场景 | 操作 | 原因 |
|------|------|------|
| 自己分支同步 main | `rebase` | 历史线性，无多余 merge commit |
| PR 合入 main | GitHub Merge 按钮 | 保留每笔提交上下文 |
| 清理本地草稿 | `rebase -i` squash | 合入前把 wip 合并成干净提交 |

### 紧急修复

```bash
git checkout main && git pull --rebase
git checkout -b hotfix-xxx
# 修 → commit
# 提 PR → main，修复合入后记得 rebase 自己分支
```

---

## 4. 代码审查

### PR 描述模板

```markdown
### 改动内容
- 后端添加 ...
- 前端新增 ...

### 测试方法
1. 登录 → 2. 调用 → 3. 验证

### 关联
Closes #42
```

### 提交前自检清单

- [ ] 能编译（`go build` / `npm run build`）
- [ ] 没有遗留的 `console.log` / `fmt.Println` 等调试代码
- [ ] 没有注释掉的代码块
- [ ] 已 rebase 最新 main，PR 只包含自己的改动
- [ ] commit 信息符合规范

### Review 原则

- **24 小时内 review**，别拖着
- 关注逻辑正确性 > 代码格式（格式靠 linter）
- 大问题直接打回，小问题可以 Approve + comment
- **对代码不对人**
  - ✅ "这里可能有个空指针风险"
  - ❌ "你写错了"

---

## 5. 不该提交的文件

`.env`、`node_modules/`、`dist/`、`bin/`、`*.log`、`.idea/`、`.vscode/`、Docker 构建缓存。

---

## 6. Release 发布流程

### 6.1 核心原则

```
tag 和分支是独立的 —— tag 绑定的是 commit，不是分支。
Release 由 GitHub Actions 自动生成，触发条件是 git tag push。
```

**关键规则**：
- 只在 `main` 分支上打正式发布 tag
- 不要在 dev 分支上打正式 tag（测试 tag 用完即删）
- 同一个 tag 名 push 多次会触发多次 workflow，生成重复 release

### 6.2 标准发布流程

```bash
# 1. 在 dev 分支上测试 release（可选）
git tag v1.2.1-rc.1
git push origin dev-tom --tags
# → GitHub Actions 会触发，生成 pre-release，验证打包是否正确

# 2. 清理测试 tag（避免污染）
git tag -d v1.2.1-rc.1                  # 删除本地 tag
git push origin :refs/tags/v1.2.1-rc.1  # 删除远程 tag

# 3. 合并到 main
git checkout main
git merge dev-tom
git push origin main

# 4. 在 main 上打正式 tag
git tag v1.2.1
git push origin v1.2.1
# → GitHub Actions 自动构建并发布 Release
```

### 6.3 Release 产物说明

| 文件 | 适用平台 | 说明 |
|------|----------|------|
| `gapi-platform-{VERSION}-linux-amd64.tar.gz` | Linux / macOS | Docker Compose 一键部署 |
| `gapi-platform-{VERSION}-windows-amd64.zip` | Windows | 同上，Windows 用 zip 解压 |

### 6.4 Tag 管理注意事项

| 操作 | 命令 | 说明 |
|------|------|------|
| 创建 tag | `git tag v1.0.0` | 打在当前 HEAD |
| 推送 tag | `git push origin v1.0.0` | 只推送单个 tag，避免 `--tags` 误推 |
| 删除本地 tag | `git tag -d v1.0.0` | 安全，不影响远程 |
| 删除远程 tag | `git push origin :refs/tags/v1.0.0` | 会删除已生成的 Release |
| 慎用 `--tags` | `git push origin --tags` | 会推送所有本地 tag，容易把测试 tag 也推上去 |

**推荐做法**：始终用 `git push origin <tag名>` 单独推送 tag，不用 `--tags`。

---

## 7. 回滚策略

### 7.1 触发条件

| 场景 | 响应时间 | 操作人 |
|------|---------|--------|
| Release 后 24h 内发现关键 bug（功能不可用、数据丢失、安全漏洞） | 立即 | 开发 |
| 用户反馈严重影响使用的功能异常 | 2h 内评估 | 产品/开发 |
| 安全漏洞发现 | 立即 | 开发/安全 |

### 7.2 回滚步骤

```bash
# 1. 在 main 上 revert 问题 commit
git checkout main
git log --oneline -10              # 确认要 revert 的 commit
git revert <问题commit的hash>
git push origin main

# 2. 打补丁 tag
git tag v1.2.1                     # 补丁版本号递增
git push origin v1.2.1
# → GitHub Actions 自动构建并发布补丁 Release

# 3. 通知相关人员
# Release notes 中说明回滚原因和修复内容
```

### 7.3 注意事项

- **禁止** `git reset --hard` 修改已推送的历史
- **禁止** `git push --force` 覆盖远程 main 分支
- 删除远程 tag 会删除对应的 GitHub Release，谨慎操作
- revert 后在 dev 分支同步修复，避免下次合入时被重新引入

---

## 8. 发布检查清单

发布前逐项确认：

- [ ] CHANGELOG.md 已更新到当前版本并包含全部变更
- [ ] README.md 版本号已更新
- [ ] docs/README.md 文档索引已同步
- [ ] 所有 PR 已合并到 dev 分支
- [ ] dev 分支 CI 通过（含测试和 lint）
- [ ] 测试环境已验证
- [ ] 测试 tag 已清理
- [ ] main 分支已更新到最新 dev
- [ ] 正式 tag 已创建并推送
- [ ] Release Notes 已审核

---
