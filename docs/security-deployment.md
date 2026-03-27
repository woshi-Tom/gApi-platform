# API Proxy Platform - 安全与部署设计文档 v1.0

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 安全设计

### 1.1 认证与授权

#### JWT 认证流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      JWT 认证流程                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. 登录请求                                                     │
│     ┌──────────┐    POST /api/v1/user/auth/login    ┌──────────┐│
│     │  Client  │ ──────────────────────────────────▶│  Server  ││
│     └──────────┘                                     └────┬─────┘│
│                                                          │      │
│  2. 验证并生成Token                                       │      │
│                                                          ▼      │
│     ┌─────────────────────────────────────────────────────┐   │
│     │  1. 验证密码                                         │   │
│     │  2. 生成 JWT Access Token (24h)                     │   │
│     │  3. 生成 JWT Refresh Token (7d)                     │   │
│     │  4. 存储 Refresh Token 到 Redis                     │   │
│     └─────────────────────────────────────────────────────┘   │
│                                                          │      │
│  3. 返回Token                                             │      │
│     ┌──────────┐    {token, refresh_token}         ┌──┴─────┐│
│     │  Client  │ ◀─────────────────────────────────│ Server ││
│     └────┬─────┘                                     └────────┘│
│          │                                                       │
│  4. 携带Token访问                                              │
│     ┌────▼─────┐    GET /api/v1/user/profile           ┌──────┐│
│     │  Client  │ ─── Authorization: Bearer <token> ──▶│Server ││
│     └──────────┘                                      └───────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### JWT 结构

```go
// Token 声明
type Claims struct {
    UserID    int64  `json:"user_id"`
    Username  string `json:"username"`
    Level     string `json:"level"`      // free|premium|vip
    TokenType string `json:"token_type"` // access|refresh
    jwt.RegisteredClaims
}

// Access Token 有效期: 24小时
// Refresh Token 有效期: 7天
```

#### 权限控制

```go
// 权限定义
const (
    PermissionChannelCreate   = "channel:create"
    PermissionChannelUpdate   = "channel:update"
    PermissionChannelDelete   = "channel:delete"
    PermissionChannelTest     = "channel:test"
    PermissionUserView       = "user:view"
    PermissionUserManage     = "user:manage"
    PermissionOrderView      = "order:view"
    PermissionOrderProcess   = "order:process"
    PermissionAuditView      = "audit:view"
    PermissionConfigManage   = "config:manage"
)

// 角色权限映射
var rolePermissions = map[string][]string{
    "super_admin": {
        PermissionChannelCreate, PermissionChannelUpdate, PermissionChannelDelete,
        PermissionChannelTest, PermissionUserView, PermissionUserManage,
        PermissionOrderView, PermissionOrderProcess, PermissionAuditView,
        PermissionConfigManage,
    },
    "admin": {
        PermissionChannelCreate, PermissionChannelUpdate, PermissionChannelDelete,
        PermissionChannelTest, PermissionUserView, PermissionUserManage,
        PermissionOrderView, PermissionOrderProcess, PermissionAuditView,
    },
    "operator": {
        PermissionChannelCreate, PermissionChannelUpdate, PermissionChannelTest,
        PermissionUserView, PermissionOrderView,
    },
    "viewer": {
        PermissionChannelTest, PermissionUserView, PermissionOrderView,
    },
}
```

### 1.2 敏感数据加密

#### API Key 加密

```go
// 使用 AES-256-GCM 加密 API Key
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

type APIKeyEncryptor struct {
    key []byte
}

func NewAPIKeyEncryptor(key string) (*APIKeyEncryptor, error) {
    // Key 应该是 32 字节 (256 bit)
    keyBytes := []byte(key)
    if len(keyBytes) != 32 {
        return nil, errors.New("key must be 32 bytes")
    }
    return &APIKeyEncryptor{key: keyBytes}, nil
}

func (e *APIKeyEncryptor) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }

    aead, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, aead.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *APIKeyEncryptor) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }

    aead, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := aead.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
    plaintext, err := aead.Open(nil, nonce, ciphertextBytes, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

#### 密码加密

```go
// 使用 bcrypt 加密密码
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### 1.3 输入验证

```go
// 使用 go-playground/validator
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50,alphanumunicode"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8,max=128,containsany=uppercase,containsany=lowercase,containsany=digit"`
}

type ChannelCreateRequest struct {
    Name      string   `json:"name" binding:"required,min=1,max=100"`
    Type      string   `json:"type" binding:"required,oneof=openai azure claude gemini anthropic custom"`
    BaseURL   string   `json:"base_url" binding:"required,url"`
    APIKey    string   `json:"api_key" binding:"required"`
    Models    []string `json:"models" binding:"required,min=1,dive,required"`
    Weight    int      `json:"weight" binding:"omitempty,min=1,max=1000"`
    RPMLimit  int      `json:"rpm_limit" binding:"omitempty,min=1"`
    TPMLimit  int      `json:"tpm_limit" binding:"omitempty,min=1"`
}
```

### 1.4 SQL 注入防护

```go
// 使用参数化查询 (go-pg)
func GetUserByEmail(email string) (*User, error) {
    var user User
    // 参数化查询,防止SQL注入
    err := db.Model(&user).Where("email = ?", email).Select()
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 避免字符串拼接
// ❌ 错误
query := "SELECT * FROM users WHERE name = '" + name + "'"

// ✅ 正确
query := db.Model(&user).Where("name = ?", name)
```

### 1.5 速率限制

```go
// Redis + Token Bucket 算法
package ratelimit

import (
    "context"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type RateLimiter struct {
    rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
    return &RateLimiter{rdb: rdb}
}

// RPM (Requests Per Minute)
func (rl *RateLimiter) AllowRPM(ctx context.Context, key string, limit int) (bool, error) {
    return rl.allow(ctx, "rpm:"+key, limit, time.Minute)
}

// TPM (Tokens Per Minute)
func (rl *RateLimiter) AllowTPM(ctx context.Context, key string, limit int, tokens int) (bool, error) {
    return rl.allow(ctx, "tpm:"+key+":"+string(rune(tokens)), limit, time.Minute)
}

func (rl *RateLimiter) allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    // Lua 脚本保证原子性
    script := redis.NewScript(`
        local current = redis.call('INCR', KEYS[1])
        if current == 1 then
            redis.call('EXPIRE', KEYS[1], ARGV[1])
        end
        return current
    `)
    
    result, err := script.Run(ctx, rl.rdb, []string{key}, int(window.Seconds())).Int()
    if err != nil {
        return false, err
    }
    
    return result <= limit, nil
}
```

### 1.6 CORS 配置

```go
// CORS 中间件
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
        c.Header("Access-Control-Max-Age", "86400")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

---

## 2. 审计日志

### 2.1 审计日志中间件

```go
package middleware

type AuditMiddleware struct {
    auditRepo repository.AuditRepository
}

func (m *AuditMiddleware) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        
        // 处理请求
        c.Next()
        
        // 异步记录审计日志
        go m.recordAudit(c, startTime)
    }
}

func (m *AuditMiddleware) recordAudit(c *gin.Context, startTime time.Time) {
    // 只记录写操作
    if c.Request.Method == "GET" {
        return
    }
    
    // 跳过健康检查等
    if shouldSkip(c.Request.URL.Path) {
        return
    }
    
    // 获取用户信息
    userID, _ := c.Get("user_id")
    username, _ := c.Get("username")
    
    // 构建审计日志
    log := &model.AuditLog{
        UserID:       toInt64(userID),
        Username:     toString(username),
        Action:       determineAction(c),
        ActionGroup:  determineActionGroup(c),
        ResourceType: determineResourceType(c.Request.URL.Path),
        ResourceID:   determineResourceID(c),
        RequestMethod: c.Request.Method,
        RequestPath:  c.Request.URL.Path,
        RequestIP:   c.ClientIP(),
        UserAgent:   c.Request.UserAgent(),
        StatusCode:  c.Writer.Status(),
        Success:     c.Writer.Status() < 400,
        TraceID:    c.GetString("trace_id"),
        CreatedAt:  time.Now(),
    }
    
    // 脱敏敏感字段
    log.RequestBody = maskSensitiveData(m.getRequestBody(c))
    log.ResponseBody = maskSensitiveData(m.getResponseBody(c))
    
    m.auditRepo.Create(log)
}

func maskSensitiveData(data string) string {
    sensitiveFields := []string{
        "password", "password_hash", "api_key", "token",
        "credit_card", "bank_account", "secret",
    }
    
    for _, field := range sensitiveFields {
        pattern := regexp.MustCompile(`"` + field + `"\s*:\s*"[^"]*"`)
        data = pattern.ReplaceAllString(data, `"`+field+`":"***"`)
    }
    
    return data
}
```

### 2.2 手动记录审计日志

```go
// 用户注册
func (s *UserService) Register(ctx *gin.Context, req *RegisterRequest) (*User, error) {
    // 业务逻辑...
    user := s.createUser(req)
    
    // 记录审计日志
    audit.Log(&model.AuditLog{
        Action:       model.AuditActionUserRegister,
        ActionGroup:  model.AuditGroupAuth,
        ResourceType: "user",
        ResourceID:   user.ID,
        UserID:       user.ID,
        RequestIP:    ctx.ClientIP(),
        Success:      true,
        NewValue: map[string]interface{}{
            "username": user.Username,
            "email":    user.Email,
        },
    })
    
    return user, nil
}

// 配额变更
func (s *QuotaService) AdjustQuota(ctx *gin.Context, userID int64, quotaType string, amount int) error {
    user := s.getUser(userID)
    oldQuota := user.RemainQuota
    
    // 更新配额
    if err := s.updateQuota(userID, quotaType, amount); err != nil {
        return err
    }
    
    // 记录审计日志
    audit.Log(&model.AuditLog{
        Action:       model.AuditActionUserQuotaAdd,
        ActionGroup:  model.AuditGroupQuota,
        ResourceType: "user",
        ResourceID:   userID,
        UserID:       userID,
        RequestIP:    ctx.ClientIP(),
        Success:      true,
        OldValue:     map[string]interface{}{"remain_quota": oldQuota},
        NewValue:     map[string]interface{}{
            "remain_quota": oldQuota + amount,
            "added": amount,
            "source": "manual",
        },
    })
    
    return nil
}
```

---

## 3. 部署设计

### 3.1 部署架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      部署架构                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                        ┌─────────────┐                         │
│                        │   Nginx     │                         │
│                        │  (HTTPS)    │                         │
│                        └──────┬──────┘                         │
│                               │                                 │
│         ┌─────────────────────┼─────────────────────┐          │
│         │                     │                     │          │
│         ▼                     ▼                     ▼          │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐  │
│  │   Frontend  │      │   Backend   │      │   Backend   │  │
│  │  (Vue SPA)  │      │   (Go)      │      │   (Go)      │  │
│  │    :80      │      │   :8080     │      │   :8081     │  │
│  └─────────────┘      └──────┬──────┘      └──────┬──────┘  │
│                               │                     │          │
│                               └─────────┬───────────┘          │
│                                         │                      │
│                               ┌─────────▼─────────┐           │
│                               │    PostgreSQL    │           │
│                               │      :5432        │           │
│                               └─────────┬─────────┘           │
│                                         │                      │
│                               ┌─────────▼─────────┐           │
│                               │      Redis       │           │
│                               │      :6379       │           │
│                               └──────────────────┘           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 二进制构建

```makefile
# 构建脚本
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: build
build:
    GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/api-proxy-linux-amd64 ./cmd/server
    GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/api-proxy-darwin-amd64 ./cmd/server
    GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/api-proxy-darwin-arm64 ./cmd/server

.PHONY: docker-build
docker-build:
    docker build --build-arg BUILD_TIME=$(BUILD_TIME) \
                --build-arg GIT_COMMIT=$(GIT_COMMIT) \
                -t api-proxy:latest \
                -f Dockerfile .
```

### 3.3 Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制源码
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 构建
ARG BUILD_TIME
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o api-proxy \
    ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制二进制
COPY --from=builder /app/api-proxy .
COPY --from=builder /app/config ./config

# 创建非root用户
RUN adduser -D -g '' appuser
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动
CMD ["./api-proxy", "-config", "config/config.yaml"]
```

### 3.4 环境变量配置

```bash
# 环境变量文件 (不提交到版本控制)
# .env.production

# 数据库
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=xxx
DATABASE_NAME=api_proxy

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=xxx

# JWT
JWT_SECRET=xxx
JWT_EXPIRE_HOUR=24

# 支付
ALIPAY_APP_ID=xxx
ALIPAY_PRIVATE_KEY=xxx
ALIPAY_PUBLIC_KEY=xxx

WECHAT_APP_ID=xxx
WECHAT_MCH_ID=xxx
WECHAT_API_KEY=xxx

# 日志
LOG_LEVEL=info
LOG_FORMAT=json
```

### 3.5 systemd 服务

```ini
# /etc/systemd/system/api-proxy.service
[Unit]
Description=API Proxy Platform
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=api-proxy
Group=api-proxy
WorkingDirectory=/opt/api-proxy
ExecStart=/opt/api-proxy/api-proxy -config /opt/api-proxy/config/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=api-proxy

# 环境变量
EnvironmentFile=/opt/api-proxy/.env

[Install]
WantedBy=multi-user.target
```

### 3.6 数据库迁移

```bash
#!/bin/bash
# scripts/migrate.sh

set -e

echo "Running database migrations..."

# 迁移目录
MIGRATION_DIR="./scripts/migrate"

# 执行迁移
for file in $(ls $MIGRATION_DIR/*.sql | sort); do
    echo "Applying: $file"
    psql -h $DATABASE_HOST -p $DATABASE_PORT -U $DATABASE_USER -d $DATABASE_NAME -f $file
done

echo "Migrations completed!"
```

---

## 4. 监控与告警

### 4.1 Prometheus 指标

```go
// 指标定义
var (
    // HTTP 请求
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    // HTTP 延迟
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    // API 调用
    apiCallsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_calls_total",
            Help: "Total API calls",
        },
        []string{"channel_id", "model", "status"},
    )
    
    // Token 使用
    tokensUsedTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tokens_used_total",
            Help: "Total tokens used",
        },
        []string{"type", "model"},
    )
    
    // 用户配额
    userQuota = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "user_quota",
            Help: "User quota",
        },
        []string{"user_id", "type"},
    )
)
```

### 4.2 健康检查

```go
// 健康检查端点
func (h *Handler) HealthCheck(c *gin.Context) {
    checks := map[string]bool{
        "database": h.checkDatabase(),
        "redis":    h.checkRedis(),
    }
    
    allHealthy := true
    for _, ok := range checks {
        if !ok {
            allHealthy = false
            break
        }
    }
    
    if allHealthy {
        c.JSON(200, gin.H{
            "status": "healthy",
            "checks": checks,
        })
    } else {
        c.JSON(503, gin.H{
            "status": "unhealthy",
            "checks": checks,
        })
    }
}
```

---

## 5. 备份与恢复

### 5.1 备份策略

```bash
#!/bin/bash
# scripts/backup.sh

set -e

BACKUP_DIR="/var/backups/api-proxy"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_$DATE.sql.gz"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 数据库备份
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME | gzip > $BACKUP_FILE

# 保留最近 30 天的备份
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete

# 上传到对象存储 (可选)
# s3cmd put $BACKUP_FILE s3://my-bucket/backups/

echo "Backup completed: $BACKUP_FILE"
```

### 5.2 恢复流程

```bash
# 从备份恢复
gunzip < backup_20260323_120000.sql.gz | psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
```

---

**文档版本**: 1.0  
**下一步**: 开始实现
