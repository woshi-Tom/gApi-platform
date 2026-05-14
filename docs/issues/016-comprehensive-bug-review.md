# 016 — 全面 Bug 审查报告

**日期**: 2026-05-15
**审查范围**: 整个项目（backend / frontend / deploy / CI / docs）
**审查方法**: 4 路并行静态分析（Go 后端 / Vue 前端 / 部署配置 / 文档）
**审查人**: TOM

---

## 变更概述

用户开发者朋友反馈 Windows 上存在大量 bug。对全项目进行跨平台兼容性和正确性审查，发现 **36 个问题**（3 Critical / 10 High / 10 Medium / 13 Low），其中 **3 个 Windows 特定问题**。

---

## Critical

### C-01. Gemini 适配器：首条消息导致 index-out-of-range panic 🪟

[gemini.go:52-65](backend/internal/pkg/adapter/gemini.go#L52-L65)

消息分组循环存在逻辑 bug。当 `contents` 为空（第一条消息）时，外层 `if len(contents) == 0` 进入，但内层 `if len(contents) > 0` 为 false，不追加新条目。随后 `lastIdx = len(contents) - 1` = `-1`，`contents[-1]` 触发 `index out of range` panic。**任何 Gemini API 调用都会导致服务器崩溃。**

**修复**: 移除内层 `if len(contents) > 0` 守卫，确保 contents 为空或 role 变化时总是追加新条目：

```go
if len(contents) == 0 || contents[len(contents)-1]["role"] != role {
    contents = append(contents, map[string]interface{}{
        "role": role,
        "parts": []map[string]string{},
    })
}
lastIdx := len(contents) - 1
parts := contents[lastIdx]["parts"].([]map[string]string)
parts = append(parts, map[string]string{"text": m["content"]})
contents[lastIdx]["parts"] = parts
```

### C-02. Gemini 适配器：system 消息静默丢失

[gemini.go:49](backend/internal/pkg/adapter/gemini.go#L49)

System 消息被映射为 `"user"` role，且因 C-01 panic 路径完全丢失。修复 C-01 后，system 消息仍会被错误地作为用户消息处理，而非通过 Gemini 的 `system_instruction` 字段传递。

**修复**: 提取 system 消息通过 Gemini API 的 `system_instruction` 顶层字段传递。

### C-03. Gemini 适配器：API Key 暴露在 URL 查询参数中

[gemini.go:40](backend/internal/pkg/adapter/gemini.go#L40)

```go
url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, model, channel.APIKey)
```

原始 API Key 出现在 URL 中，会被服务器访问日志、错误日志、中间代理日志记录，相比其他适配器的 `Authorization: Bearer` 模式存在重大凭证泄露风险。

**修复**: 至少对日志中的 key 做掩码处理（前4位+后4位）。

---

## High

### H-01. 审计日志 goroutine 数据竞争（gin.Context 回收后访问）

[audit.go:94-162](backend/internal/middleware/audit.go#L94-L162)

审计日志在 `go func()` 中异步写入，读取 `c.Get("user_id")`、`c.Get("username")`、`c.Writer.Status()`、`c.Request.URL.Path`。Gin 从 `sync.Pool` 复用 `gin.Context`，goroutine 执行时 context 可能已被回收并用于其他请求。**这是数据竞争。**

**修复**: goroutine 外捕获所有值再传入：

```go
userID, _ := c.Get("user_id")
username, _ := c.Get("username")
statusCode := c.Writer.Status()
path := c.Request.URL.Path
// ...
go func() { /* use captured locals */ }()
```

### H-02. API 访问日志 goroutine 数据竞争（同上）

[api_access_log.go:29-96](backend/internal/middleware/api_access_log.go#L29-L96)

与 H-01 相同模式。goroutine 读取 `c.Get("user_id")`、`c.Get("token_id")`、`c.Get("request_model")`、`c.Writer.Status()` 等。

**修复**: 同 H-01。

### H-03. 7 个适配器静默吞没所有错误（DeepSeek/Zhipu/Baidu/Yi/Ollama/LocalAI/Groq）

多个适配器文件，示例 [deepseek.go:51-66](backend/internal/pkg/adapter/deepseek.go#L51-L66)

所有 error 返回值使用 `_` 空白标识符：
```go
jsonPayload, _ := json.Marshal(payload)
body, _ := io.ReadAll(resp.Body)
json.Unmarshal(body, &result)
```

`json.Unmarshal` 使用空白标识符时，若上游返回无效 JSON，适配器返回零值 `ChatResponse` 且无 error。系统误认为调用成功，可能对零 token 响应扣除配额。

**修复**: 所有 fallible 操作返回 error：
```go
jsonPayload, err := json.Marshal(payload)
if err != nil { return nil, fmt.Errorf("marshal request: %w", err) }
```

### H-04. VIP 状态不一致（profile vs VIP status 端点）

[user_handler.go:188-200 / 317-332](backend/internal/handler/user_handler.go)

- `isVIPUser`（`GetProfile` 使用）：`nil VIPExpiredAt` 视为"未过期"（返回 true）
- `isValidVIP`（`GetVIPStatus` 使用）：`nil VIPExpiredAt` 视为"非 VIP"（返回 false）

同一用户通过 `/user/profile` 获得 `is_vip: true`，而 `/user/vip/status` 获得 `is_vip: false`。

**修复**: 统一两个函数的判断逻辑。

### H-05. deploy/release/docker-compose.yml：JWT_SECRET / ENCRYPT_KEY 无默认值

```yaml
GAPI_JWT_SECRET: ${JWT_SECRET}      # 无 :- 回退
GAPI_ENCRYPT_KEY: ${ENCRYPT_KEY}    # 无 :- 回退
```

若 `.env` 缺失或未设置，Docker Compose 替换为空字符串 `""`，导致 JWT secret 和加密密钥为空。

**修复**: 添加 `:-` 默认值（dev docker-compose 已有此保护）。

### H-06. deploy/release/docker-compose.yml：Redis requirepass 无法通过环境变量配置

- Redis `command` 缺少 `--requirepass` 参数
- `redis.conf` 硬编码 `requirepass CHANGE_ME`
- healthcheck 缺少 `-a <password>`

修改 `.env` 的 `REDIS_PASSWORD` 不会改变 Redis 密码，后端用新密码而 Redis 仍用 `CHANGE_ME`，认证失败。

**修复**: 添加 `--requirepass ${REDIS_PASSWORD:-CHANGE_ME}` 和 healthcheck 的 `-a` 参数。

### H-07. .github/workflows/build.yml：引用不存在的 release-readme.txt

```yaml
sed "s|{VERSION}|${VERSION}|g" release-readme.txt > "${DIR}/README.md"
```

文件 `deploy/release/release-readme.txt` 不存在，导致 release 流水线崩溃。

**修复**: 创建该文件或移除这行改为内联生成。

### H-08. deploy/docker/docker-compose.yml：config 卷挂载指向不存在的文件

```yaml
- ../../backend/config/config.yaml:/app/config/config.yaml:ro
```

`backend/config/config.yaml` 不存在（只有 `.example`）。Docker 会在源不存在时创建空目录挂载，覆盖容器内目标路径，后端无法读取配置。

**修复**: docker-compose up 前检查/复制 `.example` 文件。

### H-09. deploy/release/Dockerfile.backend：HEALTHCHECK 使用 curl 但未安装

`apk add` 只装了 `ca-certificates tzdata`，缺少 `curl`。HEALTHCHECK 和 depends_on 的 condition 均失败。

**修复**: 将 `curl`（或 `wget`）加入 `apk add`。

### H-10. deploy/nginx/deploy-nginx.sh：调用未定义的函数 create_ssl_certs

`--install` 分支调用不存在的 `create_ssl_certs` 函数。由于 `set -euo pipefail`，脚本在此崩溃，导致半安装状态。

**修复**: 删除该行（`copy_ssl_certificates` 已在下一行调用）。

---

## Medium

### M-01. SSEReader 不处理 `\r\n` 行尾 🪟

[nvidia.go:271-286](backend/internal/pkg/adapter/nvidia.go#L271-L286)

`SSEReader.Read()` 只检查 `\n` (LF)。若上游返回 CRLF，`\r` 被包含在行字符串中，`strings.HasPrefix(line, "data: ")` 检查失败，所有 SSE 数据静默丢弃。

**修复**: 返回前去除行尾 `\r`：
```go
if len(result) > 0 && result[len(result)-1] == '\r' {
    result = result[:len(result)-1]
}
```

### M-02. SIGTERM 在 Windows 上静默忽略 🪟

[main.go:123](backend/cmd/server/main.go#L123)

`signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)` — SIGTERM 在 Windows 上存在常量定义但 `signal.Notify` 不会投递。只有 SIGINT (Ctrl+C) 有效。`taskkill /PID` 或服务管理器无法优雅关闭。

**修复**: 至少文档说明此限制；或使用 build tags + Windows `SetConsoleCtrlHandler`。

### M-03. InitHandler 始终报告 RabbitMQ 断开

[router.go:75](backend/internal/router/router.go#L75)

```go
initHandler := handler.NewInitHandler(db.GetDB(), redisClient, nil)
```

始终传 `nil` 给 MQ client，即使 RabbitMQ 正在运行。初始化向导永远显示 `rabbitmq_connected: false`。

**修复**: 传递 `mqClient` 变量。

### M-04. Admin UpdateChannel 始终重置 Status 和 Priority 为零值

[admin_handler.go:342-343](backend/internal/handler/admin_handler.go)

`Name`/`BaseURL` 只在非空时更新，但 `Status`/`Priority` 始终从请求设置（包括零值）。发送 `{"name": "new name"}` 会将 status 重置为 0（禁用），priority 重置为 0。

**修复**: 与 Name/BaseURL 一致的判断模式。

### M-05. 密码暴露在 DSN 和 RabbitMQ URL 中

[config.go:403-416](backend/internal/config/config.go)

数据库密码和 RabbitMQ 密码明文嵌入连接字符串，可能出现在错误日志、堆栈跟踪中。

**修复**: 添加 `Redacted()` 方法返回密码掩码版本（`***`）。

### M-06. parseCIDR 错误处理 —— nil 指针风险

[validator.go:104-106](backend/internal/pkg/validator/validator.go)

`parseCIDR` 忽略 `net.ParseCIDR` 的 error。若 CIDR 字符串无效返回 nil `*IPNet`，`isPrivateIP` 中 `block.Contains(ip)` 触发 nil 指针 panic。当前所有 CIDR 硬编码且有效，但未来添加无效 CIDR 会导致 panic。

**修复**: 检查 error + nil guard。

### M-07. deploy/release/Dockerfile.backend：config.yaml.example 烘焙到镜像

```dockerfile
COPY config.yaml.example /app/config/config.yaml
```

镜像包含 `CHANGE_ME` 占位符配置。若环境变量未正确覆盖，生产以不安全默认值运行。

### M-08. 环境变量命名不一致（3 个 .env.example 文件使用不同约定）

- 根 `.env.example`: `DB_USER`/`DB_PASSWORD`/`DB_NAME`
- `deploy/docker/.env.example`: `POSTGRES_USER`/`POSTGRES_PASSWORD`/`POSTGRES_DB`
- `deploy/release/.env.example`: `POSTGRES_USER`/`POSTGRES_PASSWORD`/`POSTGRES_DB`

用户容易用错 `.env` 文件匹配错 compose。

### M-09. frontend/Dockerfile Node.js 版本与 CI 不匹配

Dockerfile 使用 `node:20-alpine`，CI 使用 `NODE_VERSION: '22'`。版本差异可能导致构建行为差异。

### M-10. deploy/release/docker-compose.yml：缺少 GAPI_ADMIN_SECRET

Release compose 未传递 `GAPI_ADMIN_SECRET`（prod compose 有），可能导致管理员认证回退到不安全默认值。

---

## Low

### L-01. `os.IsNotExist` 不兼容 Go 1.13+ wrapped errors

[config.go:142](backend/internal/config/config.go)

`os.IsNotExist(err)` 不支持 wrapped errors。应使用 `errors.Is(err, os.ErrNotExist)`。

### L-02. `BodySizeLimit` middlware 函数死代码

[security.go:22,30](backend/internal/middleware/security.go)

`maxBodySize` 常量和 `BodySizeLimit()` 函数从未在 router.go 注册。只有 `MaxBodyBytes(10<<20)` 被使用。

### L-03. 参数名 `bytes` 遮蔽 Go 内置类型

[security.go:42](backend/internal/middleware/security.go)

`func MaxBodyBytes(bytes int64)` 中 `bytes` 参数名遮蔽标准库 `[]byte` 类型别名。

### L-04. `Log.Path` 配置字段从未使用

[config.go:128](backend/internal/config/config.go) / [main.go:35](backend/cmd/server/main.go)

`Config.Log.Path` 有 yaml tag 且从配置加载，但 main.go 始终用 `zerolog.New(os.Stdout)`。文件日志从未实现。

### L-05. `Redis.DB = 0` 赋值死代码

[config.go:325-327](backend/internal/config/config.go)

`if c.Redis.DB == 0 { c.Redis.DB = 0 }` —— 零值赋给自己。

### L-06. `Database.Close()` gorm 关闭失败时泄露连接池

[database.go:113-120](backend/internal/repository/database.go)

`if sqlDB, err := d.DB.DB(); err == nil { sqlDB.Close() }` —— 若 `DB()` 返回 error（gorm 部分初始化时），底层连接池不关闭。

### L-07. `AutoMigrate` 始终返回 nil（即使失败）

[database.go:73-78](backend/internal/repository/database.go)

AutoMigrate 只 log warn 不返回 error，init_handler.go 的 `if err := h.DB.AutoMigrate(...)` 永远检测不到失败。

### L-08. 活动列表中已完成订单使用叉号标记 (U+2717)

[user_handler.go:495-497](backend/internal/handler/user_handler.go)

`if order.Status == model.OrderStatusCompleted { title = "✗ " + title }` —— 叉号（表示失败）用于已完成订单，用户期望用勾号。

### L-09. `maxBodySize` 常量与 `10<<20` 硬编码重复

[router.go:81](backend/internal/router/router.go) vs [security.go:22](backend/internal/middleware/security.go)

同一值在两处有 `const maxBodySize` 和 `MaxBodyBytes(10 << 20)` 两个定义。

### L-10. `ListModels` 变量 `err` 遮蔽

[api_handler.go:691](backend/internal/handler/api_handler.go)

`channelIDs, err := ...` 使用 `:=` 遮蔽外层 `err`，代码脆弱。

### L-11. deploy/release/config.yaml.example：admin_users 占位符 bcrypt 哈希混淆

占位符哈希可能被误认为有效密码。

### L-12. deploy/release/start.sh：编辑器建议缺少 vi

脚本建议 `nano .env` 或 `code .env`。无头服务器上两者可能都不可用。

### L-13. frontend/Dockerfile.admin：sed 替换可能不完整

将 `admin.html` 重命名为 `index.html` 的 `sed` 可能遗漏构建后 JS 中的引用。

---

## Windows 特定问题总结

| # | 文件 | 描述 |
|---|------|------|
| C-01 | gemini.go:52-65 | Gemini 适配器 panic — 跨平台但所有平台受影响 |
| M-01 | nvidia.go:271-286 | SSEReader 不处理 CRLF — Windows 本地 AI 服务器可能发送 CRLF |
| M-02 | main.go:123 | SIGTERM 在 Windows 上静默忽略 — 无法优雅关闭 |

Go 代码本身**没有**路径分隔符（`/` vs `\`）、文件权限位、exec.Command 等 Windows 兼容性问题。代码库未使用 `os.Create`、`os.MkdirAll`、`os.Chmod` 或 `os/exec`。GORM 表名均为小写且有显式 `TableName()`，列名通过 `gorm:"column:..."` 显式映射。

前端代码是平台无关的（Vue/TypeScript/Vite）。

---

## 从 CODE_REVIEW_REPORT 追踪到的历史未解决问题

| # | 问题 | 状态 |
|---|------|------|
| 6 | VIP 等级字符串不一致 (`"vip"` vs `"vip_bronze/silver/gold"`) | 未跟踪 |
| 8 | 支付回调未验证来源（无 IP 白名单） | 未跟踪 |
| 9 | 退款逻辑不调用支付宝退款 API | 未跟踪 |
| 12 | ChannelService 使用 NoOpCrypto | 未跟踪 |
| 15 | TokenRateLimit map 内存泄漏 | 未跟踪 |
| 19 | VIP 过期不清零 VIPQuota | 未跟踪 |
| 22 | 管理员登录无暴力破解锁定（有日志但无锁定） | 部分修复 |

---

## 修复优先级建议

| 优先级 | 问题 | 理由 |
|--------|------|------|
| P0 | C-01 Gemini panic | 服务器崩溃 |
| P0 | H-01/H-02 数据竞争 | 生产环境下审计日志损坏 |
| P1 | H-03 适配器静默吞错误 | 配额误扣 + 错误信息缺失 |
| P1 | H-04 VIP 状态不一致 | 用户可见功能 bug |
| P1 | H-05/H-06/H-07 部署配置 | 新用户无法部署 / 安全密钥为空 |
| P2 | M-01 SSEReader CRLF | Windows 流式中断 |
| P2 | M-02 SIGTERM Windows | Windows 生产环境优雅关闭 |
| P2 | M-03 InitHandler MQ nil | 运维困惑 |
| P3 | 其余 Medium/Low | 代码质量改善 |

