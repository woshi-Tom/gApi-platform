# Issue #004: Code Review 全面修复

## 修复背景

v1.2.0 发布后的全面 code review 发现多个问题，涉及 CORS 配置、数据库迁移、类型一致性、死代码等。

## 修复进度

| # | 优先级 | 问题 | 文件 | 状态 |
|---|--------|------|------|------|
| 1 | P0 | `GAPI_ADMIN_FRONTEND_URL` 环境变量未加载 | `backend/internal/config/config.go` | ✅ 已修复 |
| 2 | P0 | Model 表未注册 AutoMigrate | `backend/internal/repository/database.go` | ✅ 已修复 |
| 3 | P1 | `model/response.go` 存在未使用的 Pagination/PageData | `backend/internal/model/response.go` | ✅ 已清理 |
| 4 | P1 | `.env.example` 包含 config.go 不支持的微信支付变量 | `.env.example` | ✅ 已清理 |
| 5 | P2 | `_ = idempRepo` 冗余赋值 | `backend/internal/router/router.go` | ✅ 已清理 |
| 6 | P2 | 未使用的前端组件 SliderCaptcha.vue、HelloWorld.vue | `frontend/src/components/` | ✅ 已删除 |
| 7 | P2 | `gin.Default()` 改为 `gin.New()` 避免重复日志 | `backend/cmd/server/main.go` | ✅ 已修复 |

## 详细变更说明

### 1. `GAPI_ADMIN_FRONTEND_URL` 环境变量加载

**文件**: `backend/internal/config/config.go`

**问题**: docker-compose.yml 设置了 `GAPI_ADMIN_FRONTEND_URL`，但 `loadFromEnv()` 从未读取该变量。`cfg.Server.AdminFrontend` 始终为空，导致 CORS 中间件无法匹配管理后台前端的 origin。

**变更**: 在 `loadFromEnv()` 的 Server 段新增：
```go
if v := os.Getenv("GAPI_ADMIN_FRONTEND_URL"); v != "" {
    c.Server.AdminFrontend = v
}
```

### 2. Model 表 AutoMigrate

**文件**: `backend/internal/repository/database.go`

**变更**: 取消 `ModelGroup`、`ChannelGroupRelation`、`UserGroupRelation`、`ModelPricing` 四个模型在 `AutoMigrate()` 中的注释。

### 3. 清理 model/response.go

**变更**:
- 删除未使用的 `Pagination` 和 `PageData` 类型（业务分页使用 `pkg/response` 包）
- 添加文件注释说明类型用途

### 4. gin.Default() → gin.New()

**文件**: `backend/cmd/server/main.go`

**变更**: `gin.Default()` 自带 Logger + Recovery 中间件，与审计日志重复。改为 `gin.New()` + 手动添加 `gin.Recovery()`。

---

## 编译测试 Checklist

在合并前，请在有 Go 环境的机器上执行以下步骤：

```bash
# 1. 切到分支
git checkout dev-tom
git pull

# 2. 编译
cd backend
go build ./...
# 期望：无错误输出

# 3. 单元测试
go test -count=1 ./...
# 期望：全部 PASS

# 4. 静态检查
go vet ./...
# 期望：无错误输出

# 5. 前端类型检查
cd ../frontend
npx vue-tsc --noEmit
# 期望：无错误输出

# 6. 前端构建
npm run build
npm run build:admin
# 期望：无错误输出
```

### 特别注意

- **新增 Model 表**: `AutoMigrate` 新增了 4 个表，首次启动会自动建表。确保数据库连接正常。
- **CORS 验证**: 启动后确认管理后台前端（默认 `http://localhost:5174`）能正常请求后端 API。可通过浏览器 DevTools Network 面板检查是否有 CORS 错误。
- **gin 日志**: 启动后确认只有审计日志输出，不再有 gin 默认的请求日志。

---

## 待后续处理（本次未修复）

- Response 类型统一（`model.APIResponse` 与 `pkg/response.APIResponse` 合并，影响面大需单独 PR）
- 内存限流改 Redis-backed 实现
- Swagger 注解补全
- 前端 E2E 测试纳入 CI
- DB SSL 支持
- RateLimit graceful shutdown
