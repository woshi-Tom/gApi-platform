# v1.2.0 Release 审查报告 + 整改措施

> **审查日期**: 2026-05-12  
> **审查范围**: CI/CD Workflow / 后端代码 / 前端代码 / 文档 / 安全 / 测试覆盖  
> **审查角色**: 安全 · 产品经理 · 项目经理 · 开发 · QA/测试 · UI/前端

---

## 总体结论

| 审查角色 | 评分 | 关键问题数 |
|----------|------|-----------|
| 🔒 安全 | ⚠️ 条件通过 | 1 严重 + 5 高 + 8 中 |
| 📋 产品经理 | ❌ 不通过 | 1 阻塞 + 3 高 |
| 📊 项目经理 | ⚠️ 条件通过 | 5 中（缺失流程） |
| 💻 开发 | ⚠️ 条件通过 | 2 阻塞 + 6 重要 |
| 🧪 QA/测试 | ❌ 不通过 | 2 阻塞 + 4 高 |
| 🎨 UI/前端 | ⚠️ 条件通过 | 2 阻塞 + 2 重要（微信支付不接入，问题降级） |

**综合决议**: ❌ **当前状态不建议发布**。需优先修复 4 个阻塞级问题后重新评估。

---

## 一、CI/CD Workflow 问题

### 1.1 [阻塞] CI pipeline 不运行任何测试/lint

| 来源 | 文件 | 行号 |
|------|------|------|
| 开发、QA、PM | `.github/workflows/build.yml` | 全局 |

**问题**: Pipeline 仅包含 build 步骤，没有：
- `go test ./...` / `go vet ./...`
- `golangci-lint` / `npm lint` / `vue-tsc --noEmit`
- `npm test:e2e`
- 安全扫描（`npm audit`、`govulncheck`）

任何编译通过但功能错误的代码都可以无障碍发布。

**整改措施**:
```yaml
# 在 build-backend job 中添加
- name: Run Tests
  run: go test -count=1 -v ./...

- name: Run Vet
  run: go vet ./...

# 在 build-frontend/build-admin job 中添加
- name: Lint
  run: npm run lint   # 需要先配置 eslint

- name: Type Check
  run: npx vue-tsc --noEmit
```

**责任人**: 开发 | **预估工时**: 0.5d | **优先级**: P0

---

### 1.2 [严重] GITHUB_TOKEN 权限过大

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全 | `.github/workflows/build.yml` | 全局 |

**问题**: `build-backend`、`build-frontend`、`build-admin`、`build-docs` 四个作业没有 `permissions` 声明，继承默认的 `contents: write` 权限。这些作业只需要读取代码和上传 artifact。

**整改措施**:
```yaml
build-backend:
  permissions:
    contents: read
  # ... 对 build-* 每个作业加
```

**责任人**: 开发 | **预估工时**: 0.1d | **优先级**: P0

---

### 1.3 [高] Action 版本未锁定到不可变 SHA

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全、开发 | `.github/workflows/build.yml` | 多处 |

**问题**: 所有 action 使用可变 major version tag（`@v6` / `@v5` / `@v3`），存在 supply chain 攻击风险。

**整改措施**: 将 `@v6` 等替换为具体 commit SHA（用 Dependabot 自动更新）：
```yaml
uses: actions/checkout@<full-commit-sha>
```

**责任人**: 开发 | **预估工时**: 0.3d | **优先级**: P1

---

### 1.4 [高] Release 产物无完整性校验

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全、QA | `.github/workflows/build.yml` | 213-227 |

**问题**: Release 产物没有 SHA256SUMS、GPG 签名或 SBOM。用户无法验证下载的文件是否被篡改。

**整改措施**:
```yaml
- name: Generate Checksums
  run: |
    sha256sum gapi-platform-*.tar.gz gapi-platform-*.zip > SHA256SUMS
```

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P1

---

### 1.5 [中] Release 命名与实际内容不匹配

| 来源 | 文件 | 行号 |
|------|------|------|
| 开发 | `.github/workflows/build.yml` | 219-227 |

**问题**: `gapi-platform-*-linux-amd64.tar.gz` 文件名暗示只包含 Linux 二进制，但实际包内同时包含 Linux + Windows 两种二进制。

**整改措施**: 方案 A：按平台分别打包；方案 B：改文件名为 `gapi-platform-*-full.tar.gz`。

**责任人**: 开发 | **预估工时**: 0.3d | **优先级**: P2

---

### 1.6 [中] 无 Smoke Test 步骤

| 来源 | 文件 | 行号 |
|------|------|------|
| PM、QA | `.github/workflows/build.yml` | release job |

**问题**: Release 前没有任何冒烟测试（启动二进制、health check、API 基本验证）。

**整改措施**: 在 `release` job 前增加 smoke test：
```yaml
- name: Smoke Test
  run: |
    ./gapi-server -config config.yaml &
    sleep 2
    curl -f http://localhost:8080/health
    kill %1
```

**责任人**: 开发 | **预估工时**: 0.3d | **优先级**: P1

---

## 二、后端代码问题

### 2.1 [高] Audit goroutine 无 panic recovery

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全、开发 | `backend/internal/middleware/audit.go` | 87-150 |

**问题**: 审计日志写入在 `go func() { ... auditRepo.Create(log) }()` 中异步执行，没有 `recover()`。如果 DB 连接断开等导致 panic，会压崩整个 Go 进程。

**整改措施**:
```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[audit] panic recovered: %v", r)
        }
    }()
    if err := auditRepo.Create(log); err != nil {
        log.Printf("[audit] failed to create audit log: %v", err)
    }
}()
```

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P0

---

### 2.2 [高] Settings handler 无输入验证

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全、开发 | `backend/internal/handler/settings_handler.go` | 31-70, 177-211 |

**问题**: `UpdateSMTPConfig`、`UpdatePaymentConfig` 等端点没有输入验证：
- Host/Username/FromEmail 无格式校验
- Port 无范围校验
- 支付密钥无格式校验

**整改措施**: 添加 Gin binding validation tags：
```go
type SMTPRequest struct {
    Host      string `json:"host" binding:"required,hostname"`
    Port      int    `json:"port" binding:"min=1,max=65535"`
    Username  string `json:"username" binding:"required"`
    FromEmail string `json:"from_email" binding:"required,email"`
}
```

**责任人**: 开发 | **预估工时**: 0.5d | **优先级**: P1

---

### 2.3 [高] maskSensitiveData 脱敏不完整

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全 | `backend/internal/middleware/audit.go` | 263-281 |

**问题**: `maskSensitiveData` 只处理顶层 JSON key，嵌套对象和非 JSON 请求体（form-data、multipart）会被原样记录到审计日志。

**整改措施**: 实现递归脱敏：
```go
func maskSensitiveData(data string) string {
    var obj interface{}
    if err := json.Unmarshal([]byte(data), &obj); err != nil {
        return data
    }
    maskRecursive(obj)
    result, _ := json.Marshal(obj)
    return string(result)
}

func maskRecursive(v interface{}) {
    // 递归遍历 map/slice，对匹配 sensitiveFields 的 key 做脱敏
}
```

**责任人**: 开发 | **预估工时**: 0.5d | **优先级**: P1

---

### 2.4 [中] Audit skipPaths 精确匹配，详情页仍会记日志

| 来源 | 文件 | 行号 |
|------|------|------|
| 开发 | `backend/internal/middleware/audit.go` | 28-34 |

**问题**: `/api/v1/admin/logs/operation` 被跳过，但 `/api/v1/admin/logs/operation/123`（详情页）不会。查看审计日志本身也会产生审计日志。

**整改措施**: 改为前缀匹配：
```go
func isSkipPath(path string) bool {
    for p := range skipPaths {
        if strings.HasPrefix(path, p) {
            return true
        }
    }
    return false
}
```

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P2

---

### 2.5 [中] `UpdateRegisterSettings` 忽略 error 存在数据丢失风险

| 来源 | 文件 | 行号 |
|------|------|------|
| 开发 | `backend/internal/handler/settings_handler.go` | 128 |

**问题**: `current, _ := h.settingsSvc.GetRegisterSettings()` — error 被忽略。若 DB 错误返回 nil，partial-update 字段全部归零。

**整改措施**: 处理 error，或确保 `GetRegisterSettings` 契约保证永不返回 nil。

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P2

---

## 三、前端代码问题

### 3.1 [阻塞] 3 个 save 函数是存根（Stub）

| 来源 | 文件 | 行号 |
|------|------|------|
| UI/前端 | `frontend/src/views/admin/settings/Index.vue` | 433, 470, 519 |

**问题**: `saveGeneral()`、`saveRateLimit()`、`saveSecurity()` 使用 `new Promise(resolve => setTimeout(resolve, 500))` 模拟异步，没有真实 API 调用。用户点击保存会看到"保存成功"提示，但数据根本没有提交。

**整改措施**: 实现真实 API 调用：
1. 后端添加 `PUT /api/v1/admin/settings/general`、`rate-limit`、`security` 路由
2. 前端 `api/settings.ts` 添加对应 API 函数
3. 替换存根为真实调用

相关后端路由参考：
- `router.go` 中 `adminAuth.PUT("/settings/email", ...)` 等模式

**责任人**: 开发 + 前端 | **预估工时**: 1d | **优先级**: P0

---

### 3.2 [已关闭] 微信支付不接入，移除 UI 字段

| 来源 | 文件 | 行号 |
|------|------|------|
| UI/前端 | `frontend/src/views/admin/settings/Index.vue` | 481-515 |

**决策**: ❌ **微信支付暂不接入**（接入成本高，优先级低）。

**问题**: `loadPaymentConfig()` 只加载支付宝字段，`savePayment()` 只保存支付宝字段。WeChat 字段存在表单中但不加载/保存。用户填入微信配置后保存不会报错但数据也不会保存。

**整改措施**: 
1. 从前端 `settings/Index.vue` 的支付设置 tab 中移除微信支付相关字段（WeChat 的 AppID、Secret、商户号等）
2. 保持支付设置 tab 仅展示支付宝配置
3. 将来需要接入时再重新添加

**责任人**: 前端 | **预估工时**: 0.2d | **优先级**: P2（降级）

---

### 3.3 [重要] 所有表单无验证规则

| 来源 | 文件 | 行号 |
|------|------|------|
| UI/前端 | `frontend/src/views/admin/settings/Index.vue` | 全局 |

**问题**: 6 个 `el-form` 均未配置 `:rules`。对比项目中 ChangePassword.vue、channels/List.vue 均有验证规则。

**整改措施**: 为每个 form 添加 Element Plus 验证规则：
```vue
<el-form :model="emailForm" :rules="emailRules" ref="emailFormRef">
```
```ts
const emailRules = {
  host: [{ required: true, message: '请输入 SMTP 服务器', trigger: 'blur' }],
  port: [{ type: 'number', min: 1, max: 65535, message: '端口范围 1-65535', trigger: 'blur' }],
}
```

**责任人**: 前端 | **预估工时**: 0.5d | **优先级**: P1

---

### 3.4 [重要] CI 中无前端 lint/typecheck

| 来源 | 文件 | 行号 |
|------|------|------|
| UI/前端 | `.github/workflows/build.yml` | frontend build jobs |

**问题**: 前端构建前无 `vue-tsc --noEmit` 类型检查或 eslint 检查。

**整改措施**: 参照 1.1。需先在项目中配置 `vue-tsc` 和 eslint。

**责任人**: 前端 | **预估工时**: 0.5d | **优先级**: P1

---

## 四、测试覆盖问题

### 4.1 [阻塞] 核心变更文件零测试覆盖

| 来源 | 文件 |
|------|------|
| QA | `backend/internal/handler/settings_handler.go` |
| QA | `backend/internal/middleware/audit.go` |
| QA | `backend/internal/service/settings_service.go` |

**问题**: 本次发布的 3 个核心变更文件测试覆盖率为 0%。

**整改措施**: 新增测试文件：

`backend/internal/handler/settings_handler_test.go`：
- `UpdateRegisterSettings` 部分更新行为
- `UpdateSMTPConfig` 密码加密
- `UpdatePaymentConfig` 缓存失效

`backend/internal/middleware/audit_test.go`：
- GET 请求跳过行为
- `maskSensitiveData` 嵌套/非 JSON
- skipPaths 精确匹配
- goroutine 错误处理

`backend/internal/service/settings_service_test.go`：
- Config group 隔离
- 缓存 TTL
- 加密存储/解密

**责任人**: 开发 | **预估工时**: 1.5d | **优先级**: P0

---

### 4.2 [高] CI 中无测试执行步骤

| 来源 | 文件 |
|------|------|
| QA | `.github/workflows/build.yml` |

**问题**: 已有测试文件（7 个 Go 测试文件 + 3 个 Playwright spec）在 CI 中不执行。

**整改措施**: 见 1.1。在 `build-backend` 中加入 `go test -count=1 ./...`。

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P0

---

## 五、文档问题

### 5.1 [高] CHANGELOG 遗漏模型管理页面

| 来源 | 文件 |
|------|------|
| 产品 | `CHANGELOG.md` |

**问题**: `frontend/src/views/admin/models/Index.vue`（418 行全新页面）在 CHANGELOG v1.2.0 中完全未提及。

**整改措施**: 在 CHANGELOG v1.2.0 追加：
```markdown
### 功能增强
- **模型管理前端页面**: 新增模型分组/定价/权限的前端管理界面
```

**责任人**: 开发 | **预估工时**: 0.1d | **优先级**: P1

---

### 5.2 [高] 无 Breaking Changes 章节

| 来源 | 文件 |
|------|------|
| 产品 | `CHANGELOG.md` |

**问题**: 从 v1.1.x 升级到 v1.2.0 无 Breaking Changes 说明。运维人员无法评估升级风险。

**整改措施**: 在 CHANGELOG v1.2.0 增加：
```markdown
### Breaking Changes
- 无破坏性变更。审计日志默认跳过 GET 请求（重要 GET 路径如支付除外），如需恢复旧行为需修改配置。
```

**责任人**: 开发 | **预估工时**: 0.1d | **优先级**: P1

---

### 5.3 [中] 审计日志功能状态矛盾

| 来源 | 文件 |
|------|------|
| 产品 | `docs/features/README.md` vs `CHANGELOG.md` |

**问题**: `docs/features/README.md` 标注审计日志为 ⚠️（部分实现），但 CHANGELOG 中作为 v1.2.0 新功能列出。读者困惑是否已完成。

**整改措施**: 统一状态：如果确认功能达到发布标准，将 features/README.md 的 ⚠️ 改为 ✅ 并注明"基础功能已完成，数据清理策略将在后续版本完善"。

**责任人**: 开发 | **预估工时**: 0.1d | **优先级**: P2

---

### 5.4 [中] 无升级指南

| 来源 | 文件 |
|------|------|
| 产品 | `docs/deployment/` |

**问题**: 没有从旧版本升级到 v1.2.0 的步骤文档（数据迁移、配置变更）。

**整改措施**: 创建 `docs/deployment/upgrade-v1.2.0.md`，包含：
- 数据库迁移（如果有）
- 环境变量变更
- 部署步骤
- 回滚指南

**责任人**: 开发 | **预估工时**: 0.3d | **优先级**: P2

---

### 5.5 [中] 发布流程无检查清单

| 来源 | 文件 |
|------|------|
| PM | `docs/development/git-workflow.md` |

**问题**: 发布流程缺少可执行的检查清单。全部凭开发者记忆，新人容易漏步骤。

**整改措施**: 在 `git-workflow.md` 第 6 章末尾增加：
```markdown
### 发布检查清单

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
```

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P2

---

### 5.6 [中] 无回滚策略文档

| 来源 | 文件 |
|------|------|
| PM | `docs/development/git-workflow.md` |

**问题**: 没有定义回滚触发条件、步骤和重新发布流程。

**整改措施**: 在 `git-workflow.md` 新增第 7 章「回滚策略」：
```markdown
## 7. 回滚策略

### 触发条件
- Release 后 24h 内发现关键 bug
- 用户反馈严重影响使用的功能异常
- 安全漏洞发现

### 回滚步骤
1. `git revert <commit>` 在 main 上 revert 问题 commit
2. 打补丁 tag（如 v1.2.1）
3. 推送 tag 触发重新发布

### 注意事项
- 禁止 `git reset --hard` 修改历史
- 删除远程 tag 会删除对应 Release，谨慎操作
```

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P2

---

## 六、安全问题补充

### 6.1 [中] AES 密钥长度未校验

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全 | `backend/internal/pkg/crypto/aes.go` | 16-25 |

**问题**: `NewAESEncryptor` 接受任意长度密钥，弱密钥被静默填充到 32 字节。

**整改措施**: `Init` 中增加密钥长度校验：
```go
func Init(key string) error {
    if len(key) < 32 {
        return errors.New("encryption key must be at least 32 characters")
    }
    // ...
}
```

**责任人**: 开发 | **预估工时**: 0.1d | **优先级**: P2

---

### 6.2 [中] Dockerfile 以 root 运行

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全 | `Dockerfile.backend` | 20 |

**问题**: 容器以 root 运行，二进制被利用时攻击者获得容器 root 权限。

**整改措施**:
```dockerfile
RUN adduser -D -g '' appuser
USER appuser
```

**责任人**: 开发 | **预估工时**: 0.2d | **优先级**: P2

---

### 6.3 [中] Redis 密码泄露于进程列表

| 来源 | 文件 | 行号 |
|------|------|------|
| 安全 | `deploy/release/docker-compose.yml` | 28, 34 |

**问题**: Redis 密码通过命令行参数传递，`docker ps` 暴露进程命令参数。

**整改措施**: 使用 Redis 配置文件设置密码，或使用 Docker secrets。

**责任人**: 开发 | **预估工时**: 0.3d | **优先级**: P3

---

## 七、全部问题汇总

| # | 严重度 | 分类 | 领域 | 文件 | 简述 | 措施 | 工时 |
|---|--------|------|------|------|------|------|------|
| 1 | 🔴 阻塞 | CI/CD | DevOps | `build.yml` | Pipeline 无测试/lint 步骤 | 添加 `go test`、`go vet`、lint 步骤 | 0.5d |
| 2 | 🔴 阻塞 | 前端 | Frontend | `settings/Index.vue` | 3 个 save 函数是存根 | 实现真实 API 调用 | 1d |
| 3 | 🟢 已关闭 | 前端 | Frontend | `settings/Index.vue` | 微信支付不接入 | 移除 UI 中微信支付字段 | 0.2d |
| 4 | 🔴 阻塞 | 测试 | Backend | `settings_handler.go` | 核心文件零测试覆盖 | 新增 handler/middleware/service 测试 | 1.5d |
| 5 | 🔴 严重 | 安全 | CI/CD | `build.yml` | GITHUB_TOKEN 权限过大 | 各 job 添加 `permissions: contents: read` | 0.1d |
| 6 | 🟠 高 | 安全 | Backend | `audit.go` | Goroutine 无 recover | 添加 `defer recover()` | 0.2d |
| 7 | 🟠 高 | 安全 | CI/CD | `build.yml` | Action 版本未锁 SHA | 锁定到 commit SHA | 0.3d |
| 8 | 🟠 高 | 安全 | CI/CD | `build.yml` | Release 无 checksum | 生成 SHA256SUMS | 0.2d |
| 9 | 🟠 高 | 安全 | Backend | `settings_handler.go` | 输入验证缺失 | 添加 binding validation | 0.5d |
| 10 | 🟠 高 | 安全 | Backend | `audit.go:263` | maskSensitiveData 不完整 | 递归脱敏 | 0.5d |
| 11 | 🟠 高 | 测试 | CI/CD | `build.yml` | CI 不执行测试 | 见 #1 | - |
| 12 | 🟠 高 | 文档 | Docs | `CHANGELOG.md` | 遗漏模型管理页面 | 追加条目 | 0.1d |
| 13 | 🟠 高 | 文档 | Docs | `CHANGELOG.md` | 无 Breaking Changes | 增加章节 | 0.1d |
| 14 | 🟡 中 | 前端 | Frontend | `settings/Index.vue` | 表单无验证规则 | 添加 `:rules` | 0.5d |
| 15 | 🟡 中 | 前端 | CI/CD | `build.yml` | CI 无前端 lint/typecheck | 添加 `vue-tsc --noEmit` | 0.5d |
| 16 | 🟡 中 | 安全 | Backend | `crypto/aes.go` | AES 密钥未校验 | Init 增加密钥长度检查 | 0.1d |
| 17 | 🟡 中 | 安全 | Deploy | `Dockerfile.backend` | 以 root 运行 | 添加 USER 指令 | 0.2d |
| 18 | 🟡 中 | CI/CD | DevOps | `build.yml` | Release 命名与实际不符 | 改文件名或分平台打包 | 0.3d |
| 19 | 🟡 中 | CI/CD | DevOps | `build.yml` | 无 Smoke Test | 增加冒烟测试步骤 | 0.3d |
| 20 | 🟡 中 | 后端 | Backend | `audit.go:28` | skipPaths 精确匹配 | 改为前缀匹配 | 0.2d |
| 21 | 🟡 中 | 后端 | Backend | `settings_handler.go:128` | 忽略 error 风险 | 处理 error | 0.2d |
| 22 | 🟡 中 | 文档 | Docs | `features/README.md` | 审计日志状态矛盾 | 统一标注 | 0.1d |
| 23 | 🟡 中 | 文档 | Docs | `deployment/` | 无升级指南 | 创建 upgrade 文档 | 0.3d |
| 24 | 🟡 中 | 文档 | Docs | `git-workflow.md` | 无发布检查清单 | 增加 checklist | 0.2d |
| 25 | 🟡 中 | 文档 | Docs | `git-workflow.md` | 无回滚策略 | 新增第 7 章 | 0.2d |
| 26 | 🟢 低 | 安全 | Deploy | `docker-compose.yml` | Redis 密码泄露于进程 | 改用配置文件 | 0.3d |

### 工时估算

| 优先级 | 问题数 | 预估总工时 |
|--------|--------|-----------|
| P0（发布前必须修） | 4 | ~2.3d |
| P1（建议发布前修） | 7 | ~1.8d |
| P2（后续版本修） | 12 | ~2.6d |
| P3（低优先级） | 1 | ~0.3d |

---

## 八、整改后重新评估标准

修复 P0 问题后，按下表重新评估是否达到发布标准：

| 检查项 | 标准 | 验证方式 |
|--------|------|----------|
| CI 测试通过 | `go test ./...` 全部通过 | CI green |
| 代码质量检查 | `go vet` / `vue-tsc` 无报错 | CI green |
| 前端存根修复 | settings 页面 6 个 tab 均可真实保存 | 人工验证 |
| CHANGELOG 完整 | v1.2.0 条目包含所有变更 | 人工审核 |
| GITHUB_TOKEN 权限 | build 作业 `permissions: contents: read` | 代码审查 |
| Audit recover | goroutine 有 `recover()` | 代码审查 |
| 核心功能测试 | settings + audit 关键路径有测试 | CI green |
| 心跳检查 | Release 产物启动后 health check 通过 | Smoke Test |
| 文档一致性 | features/README 与 CHANGELOG 无矛盾 | 人工审核 |
| 发布检查清单 | 所有项 ✅ | 人工执行 |
