# API Proxy Platform - 项目结构设计文档 v1.0

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 项目概览

```
api-proxy/
├── backend/                    # Go 后端
│   ├── cmd/                   # 入口
│   ├── internal/              # 内部包
│   ├── pkg/                   # 公共包
│   ├── scripts/               # 脚本
│   └── config/                # 配置
│
├── frontend/                   # Vue 3 前端
│   ├── src/
│   │   ├── api/              # API 调用
│   │   ├── components/        # 组件
│   │   ├── views/            # 页面
│   │   ├── stores/           # 状态管理
│   │   └── utils/            # 工具函数
│   └── public/               # 静态资源
│
├── scripts/                    # 数据库脚本
└── docs/                       # 文档
```

---

## 2. Go 后端项目结构

### 2.1 目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
│
├── internal/
│   ├── config/                 # 配置
│   │   ├── config.go           # 配置结构体
│   │   ├── database.go         # 数据库配置
│   │   ├── redis.go            # Redis配置
│   │   └── jwt.go              # JWT配置
│   │
│   ├── model/                   # 数据模型
│   │   ├── user.go             # 用户模型
│   │   ├── channel.go          # 渠道模型
│   │   ├── token.go            # Token模型
│   │   ├── order.go            # 订单模型
│   │   ├── audit.go            # 审计日志模型
│   │   └── response.go         # 响应模型
│   │
│   ├── repository/             # 数据访问层
│   │   ├── user_repo.go
│   │   ├── channel_repo.go
│   │   ├── token_repo.go
│   │   ├── order_repo.go
│   │   └── audit_repo.go
│   │
│   ├── service/                 # 业务逻辑层
│   │   ├── user_service.go
│   │   ├── channel_service.go
│   │   ├── token_service.go
│   │   ├── auth_service.go
│   │   ├── payment_service.go
│   │   ├── quota_service.go
│   │   └── audit_service.go
│   │
│   ├── handler/                  # HTTP处理器
│   │   ├── user_handler.go
│   │   ├── channel_handler.go
│   │   ├── token_handler.go
│   │   ├── auth_handler.go
│   │   ├── order_handler.go
│   │   ├── admin_handler.go
│   │   └── openai_handler.go   # OpenAI兼容API
│   │
│   ├── middleware/               # 中间件
│   │   ├── auth.go             # 认证中间件
│   │   ├── cors.go             # 跨域中间件
│   │   ├── ratelimit.go        # 限流中间件
│   │   ├── logger.go           # 日志中间件
│   │   ├── recovery.go         # 恢复中间件
│   │   └── audit.go            # 审计中间件
│   │
│   ├── router/                   # 路由
│   │   ├── router.go           # 主路由
│   │   ├── user.go             # 用户路由
│   │   ├── admin.go            # 管理路由
│   │   └── openai.go           # OpenAI路由
│   │
│   ├── pkg/                     # 公共包
│   │   ├── response/           # 响应封装
│   │   ├── errors/             # 错误处理
│   │   ├── crypto/             # 加密解密
│   │   ├── validator/          # 参数验证
│   │   └── utils/              # 工具函数
│   │
│   └── third/                   # 第三方集成
│       ├── openai/             # OpenAI客户端
│       ├── alipay/             # 支付宝
│       ├── wechat/             # 微信支付
│       └── email/              # 邮件服务
│
├── config/
│   └── config.yaml             # 配置文件
│
├── scripts/
│   ├── migrate/                 # 数据库迁移
│   │   ├── 001_init.sql
│   │   └── 002_seed.sql
│   └── init.sh                 # 初始化脚本
│
├── go.mod                       # Go模块
├── go.sum                       # 依赖锁定
└── Makefile                     # 构建脚本
```

### 2.2 核心文件示例

#### main.go

```go
package main

import (
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "api-proxy/internal/config"
    "api-proxy/internal/router"
    "api-proxy/internal/repository"
    "api-proxy/internal/service"
    "api-proxy/internal/handler"
    "api-proxy/internal/middleware"
)

func main() {
    // 解析命令行参数
    configPath := flag.String("config", "config/config.yaml", "config file path")
    flag.Parse()
    
    // 加载配置
    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 初始化数据库
    db, err := repository.NewDB(cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect database: %v", err)
    }
    defer db.Close()
    
    // 初始化Redis
    rdb, err := repository.NewRedis(cfg.Redis)
    if err != nil {
        log.Fatalf("Failed to connect redis: %v", err)
    }
    defer rdb.Close()
    
    // 初始化存储库
    repos := repository.NewRepositories(db, rdb)
    
    // 初始化服务
    services := service.NewServices(repos, cfg)
    
    // 初始化处理器
    handlers := handler.NewHandlers(services)
    
    // 初始化中间件
    middlewares := middleware.NewMiddlewares(services, cfg)
    
    // 设置路由
    r := router.Setup(handlers, middlewares, cfg)
    
    // 启动服务器
    go func() {
        log.Printf("Server starting on %s", cfg.Server.Addr)
        if err := r.Run(cfg.Server.Addr); err != nil {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")
}
```

#### config.go

```go
package config

import (
    "fmt"
    "os"
    
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis   RedisConfig    `yaml:"redis"`
    JWT     JWTConfig     `yaml:"jwt"`
    Payment PaymentConfig  `yaml:"payment"`
}

type ServerConfig struct {
    Addr    string `yaml:"addr"`
    Mode    string `yaml:"mode"`
    Timeout int    `yaml:"timeout"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Database string `yaml:"database"`
    MaxOpen int    `yaml:"max_open"`
    MaxIdle int    `yaml:"max_idle"`
}

type RedisConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

type JWTConfig struct {
    Secret     string `yaml:"secret"`
    ExpireHour int    `yaml:"expire_hour"`
}

type PaymentConfig struct {
    Alipay  AlipayConfig  `yaml:"alipay"`
    Wechat  WechatConfig  `yaml:"wechat"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }
    
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse yaml: %w", err)
    }
    
    return &cfg, nil
}
```

### 2.3 依赖管理

```go
// go.mod
module api-proxy

go 1.22

require (
    // Web框架
    github.com/gin-gonic/gin v1.9.1
    
    // 数据库
    github.com/go-pg/pg/v10 v10.1.0
    github.com/go-redis/redis/v8 v8.11.5
    
    // JWT
    github.com/golang-jwt/jwt/v5 v5.2.0
    
    // 配置
    gopkg.in/yaml.v3 v3.0.1
    
    // 验证
    github.com/go-playground/validator/v10 v10.18.0
    
    // 加密
    golang.org/x/crypto v0.21.0
    
    // UUID
    github.com/google/uuid v1.6.0
    
    // 日志
    github.com/rs/zerolog v1.32.0
    
    // 第三方支付
    github.com/smartwalle/alipay/v3 v3.0.1
    github.com/yfd11/wxpay v1.0.0
    
    // HTTP客户端
    github.com/imroc/req/v3 v3.35.0
)
```

### 2.4 Makefile

```makefile
.PHONY: build run test clean lint fmt

# 编译
build:
    go build -o bin/api-proxy ./cmd/server

# 运行
run:
    go run ./cmd/server -config config/config.yaml

# 测试
test:
    go test -v -cover ./...

# 清理
clean:
    rm -rf bin/

# 格式化
fmt:
    go fmt ./...

# 依赖
deps:
    go mod tidy
    go mod download

# Lint
lint:
    golangci-lint run

# 数据库迁移
migrate:
    psql -h localhost -U postgres -d api_proxy -f scripts/migrate/001_init.sql

# Docker构建
docker-build:
    docker build -t api-proxy:latest -f Dockerfile .

# Docker运行
docker-run:
    docker run -p 8080:8080 api-proxy:latest
```

---

## 3. Vue 3 前端项目结构

### 3.1 目录结构

```
frontend/
├── public/
│   ├── favicon.ico
│   └── index.html
│
├── src/
│   ├── api/                     # API 调用
│   │   ├── index.ts            # API 入口
│   │   ├── user.ts             # 用户 API
│   │   ├── channel.ts          # 渠道 API
│   │   ├── token.ts            # Token API
│   │   ├── order.ts            # 订单 API
│   │   └── admin.ts            # 管理 API
│   │
│   ├── assets/                  # 静态资源
│   │   ├── images/
│   │   └── styles/
│   │       ├── variables.scss  # 变量
│   │       └── global.scss     # 全局样式
│   │
│   ├── components/               # 组件
│   │   ├── common/              # 通用组件
│   │   │   ├── Pagination.vue
│   │   │   ├── Search.vue
│   │   │   └── Empty.vue
│   │   │
│   │   ├── form/                # 表单组件
│   │   │   ├── ChannelForm.vue
│   │   │   ├── TokenForm.vue
│   │   │   └── UserForm.vue
│   │   │
│   │   └── chart/               # 图表组件
│   │       ├── UsageChart.vue
│   │       └── TrendChart.vue
│   │
│   ├── composables/              # 组合式函数
│   │   ├── useApi.ts
│   │   ├── useAuth.ts
│   │   └── usePagination.ts
│   │
│   ├── hooks/                    # 钩子函数
│   │   ├── useLoading.ts
│   │   └── useDebounce.ts
│   │
│   ├── layouts/                  # 布局
│   │   ├── DefaultLayout.vue
│   │   ├── AdminLayout.vue
│   │   └── UserLayout.vue
│   │
│   ├── router/                   # 路由
│   │   ├── index.ts
│   │   ├── routes.ts
│   │   └── guards.ts
│   │
│   ├── stores/                   # 状态管理 (Pinia)
│   │   ├── user.ts
│   │   ├── channel.ts
│   │   ├── token.ts
│   │   └── admin.ts
│   │
│   ├── types/                    # TypeScript 类型
│   │   ├── api.ts
│   │   ├── user.ts
│   │   ├── channel.ts
│   │   └── response.ts
│   │
│   ├── utils/                    # 工具函数
│   │   ├── request.ts           # Axios 封装
│   │   ├── storage.ts           # 本地存储
│   │   ├── format.ts            # 格式化
│   │   └── validate.ts          # 验证
│   │
│   ├── views/                    # 页面
│   │   ├── home/                # 首页
│   │   │   └── Home.vue
│   │   │
│   │   ├── auth/                # 认证
│   │   │   ├── Login.vue
│   │   │   ├── Register.vue
│   │   │   └── ForgotPassword.vue
│   │   │
│   │   ├── user/                # 用户端
│   │   │   ├── Dashboard.vue
│   │   │   ├── Tokens.vue
│   │   │   ├── Recharge.vue
│   │   │   ├── VIP.vue
│   │   │   └── Profile.vue
│   │   │
│   │   └── admin/               # 管理后台
│   │       ├── Login.vue
│   │       ├── Dashboard.vue
│   │       ├── channel/
│   │       │   ├── List.vue
│   │       │   ├── Form.vue
│   │       │   ├── Test.vue
│   │       │   └── TestHistory.vue
│   │       ├── user/
│   │       │   ├── List.vue
│   │       │   └── Form.vue
│   │       ├── token/
│   │       │   └── List.vue
│   │       ├── vip/
│   │       │   ├── Packages.vue
│   │       │   └── Orders.vue
│   │       ├── order/
│   │       │   └── List.vue
│   │       ├── audit/           # 审计日志 ⭐
│   │       │   ├── List.vue
│   │       │   ├── Detail.vue
│   │       │   └── Statistics.vue
│   │       ├── login-logs/      # 登录日志 ⭐
│   │       │   └── List.vue
│   │       └── settings/
│   │           └── Config.vue
│   │
│   ├── App.vue                   # 根组件
│   └── main.ts                   # 入口文件
│
├── .env                          # 环境变量
├── .env.development              # 开发环境
├── .env.production               # 生产环境
│
├── package.json
├── tsconfig.json
├── vite.config.ts
└── vue.config.js
```

### 3.2 核心文件示例

#### request.ts (Axios 封装)

```typescript
import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

// 创建实例
const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 添加 Token
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    
    // 添加请求 ID
    config.headers['X-Request-ID'] = crypto.randomUUID()
    
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const { code, message, data } = response.data
    
    if (code === 0) {
      return data
    }
    
    // 业务错误
    ElMessage.error(message)
    return Promise.reject(new Error(message))
  },
  (error: AxiosError) => {
    const { response } = error
    
    if (response) {
      switch (response.status) {
        case 401:
          // Token 过期
          ElMessage.error('登录已过期，请重新登录')
          const userStore = useUserStore()
          userStore.logout()
          router.push('/login')
          break
        case 403:
          ElMessage.error('权限不足')
          break
        case 429:
          ElMessage.error('请求过于频繁，请稍后再试')
          break
        default:
          ElMessage.error('请求失败，请稍后再试')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    
    return Promise.reject(error)
  }
)

export default service
```

#### channel.ts (API 调用)

```typescript
import request from '@/utils/request'
import type { Channel, ChannelListParams, ChannelTestRequest } from '@/types/channel'

// 获取渠道列表
export function getChannelList(params: ChannelListParams) {
  return request.get<{
    total: number
    list: Channel[]
  }>('/admin/channels', { params })
}

// 创建渠道
export function createChannel(data: Partial<Channel>) {
  return request.post<Channel>('/admin/channels', data)
}

// 更新渠道
export function updateChannel(id: number, data: Partial<Channel>) {
  return request.put<Channel>(`/admin/channels/${id}`, data)
}

// 删除渠道
export function deleteChannel(id: number) {
  return request.delete<void>(`/admin/channels/${id}`)
}

// 测试渠道 ⭐
export function testChannel(id: number, data: ChannelTestRequest) {
  return request.post(`/admin/channels/${id}/test`, data)
}

// 获取测试历史
export function getTestHistory(id: number, params: { page: number; page_size: number }) {
  return request.get(`/admin/channels/${id}/test-history`, { params })
}
```

#### types/channel.ts (类型定义)

```typescript
export interface Channel {
  id: number
  name: string
  type: string
  base_url: string
  status: number
  is_healthy: boolean
  models: string[]
  weight: number
  priority: number
  rpm_limit: number
  tpm_limit: number
  failure_count: number
  last_success_at: string
  response_time_avg: number
  created_at: string
}

export interface ChannelListParams {
  page: number
  page_size: number
  type?: string
  status?: number
  group?: string
  keyword?: string
}

export interface ChannelTestRequest {
  test_type: 'models' | 'chat' | 'embeddings'
  model?: string
  messages?: { role: string; content: string }[]
  input?: string
  temperature?: number
  max_tokens?: number
}

export interface ChannelTestResponse {
  success: boolean
  response_time_ms: number
  status_code: number
  models?: string[]
  content?: string
  usage?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
  embedding?: number[]
  error?: string
}
```

### 3.3 package.json

```json
{
  "name": "api-proxy-frontend",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts --fix"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.2.0",
    "pinia": "^2.1.0",
    "element-plus": "^2.5.0",
    "@element-plus/icons-vue": "^2.3.0",
    "axios": "^1.6.0",
    "dayjs": "^1.11.0",
    "echarts": "^5.5.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0",
    "vue-tsc": "^1.8.0",
    "sass": "^1.69.0",
    "unplugin-vue-components": "^0.26.0",
    "@types/node": "^20.10.0"
  }
}
```

### 3.4 vite.config.ts

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import path from 'path'

export default defineConfig({
  plugins: [
    vue(),
    Components({
      resolvers: [
        ElementPlusResolver()
      ]
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 1500
  }
})
```

---

## 4. 配置管理

### 4.1 后端配置 (config.yaml)

```yaml
server:
  addr: ":8080"
  mode: "release"
  timeout: 60

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  database: "api_proxy"
  max_open: 100
  max_idle: 10

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-here"
  expire_hour: 24

payment:
  alipay:
    app_id: ""
    private_key: ""
    alipay_public_key: ""
    sandbox: false
  wechat:
    app_id: ""
    mch_id: ""
    api_key: ""
    cert_path: ""

rate_limit:
  free_rpm: 60
  free_tpm: 10000
  vip_rpm: 2000
  vip_tpm: 100000

log:
  level: "info"
  format: "json"
  output: "stdout"
```

### 4.2 前端环境变量 (.env)

```bash
# 开发环境
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=API Proxy Platform
VITE_APP_VERSION=1.0.0

# 生产环境
VITE_API_BASE_URL=https://api.example.com/api/v1
```

---

## 5. 部署架构

### 5.1 Docker Compose

```yaml
version: '3.8'

services:
  api-proxy:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/config/config.yaml
    volumes:
      - ./config:/app/config
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  frontend:
    build: ./frontend
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - api-proxy
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: api_proxy
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

### 5.2 目录权限

```
/opt/api-proxy/
├── config/
│   └── config.yaml           # 600 - 包含敏感信息
├── logs/                     # 755 - 日志目录
├── data/                     # 755 - 数据目录
├── backend                   # 可执行文件
│   └── api-proxy
└── frontend
    └── dist/                 # 前端构建产物
```

---

**文档版本**: 1.0  
**下一步**: 安全与部署设计
