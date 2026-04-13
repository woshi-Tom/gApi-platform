# gAPI Platform v$VERSION$

## 快速开始

### 方式1：Docker Compose（推荐）

```bash
# 1. 复制环境变量配置
cp .env.example .env

# 2. 编辑 .env 填入你的配置
nano .env

# 3. 启动所有服务
docker-compose up -d

# 4. 查看运行状态
docker-compose ps
```

### 方式2：直接运行二进制

```bash
# 需要先准备好 PostgreSQL、Redis、RabbitMQ

# 设置环境变量
export GAPI_DB_HOST=localhost
export GAPI_DB_PORT=5432
export GAPI_DB_USER=your_user
export GAPI_DB_PASSWORD=your_password
export GAPI_DB_NAME=gapi
export GAPI_REDIS_HOST=localhost
export GAPI_REDIS_PORT=6379
export GAPI_REDIS_PASSWORD=your_redis_password
export GAPI_RABBITMQ_HOST=localhost
export GABBITMQ_PORT=5672
export GAPI_RABBITMQ_USER=your_user
export GAPI_RABBITMQ_PASSWORD=your_password
export GAPI_JWT_SECRET=your_jwt_secret_min_32_chars
export GAPI_ENCRYPT_KEY=your_encrypt_key_min_32_chars

# 运行
./gapi-server
```

## 访问地址

| 服务 | 地址 |
|------|------|
| 用户前端 | http://localhost:5173 |
| 管理后台 | http://localhost:5174 |
| API 接口 | http://localhost:8080 |
| RabbitMQ | http://localhost:15672 |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |

## 目录结构

```
.
├── gapi-server          # 后端二进制
├── gapi-frontend/      # 用户前端 (Nginx)
├── gapi-admin/          # 管理后台 (Nginx)
├── docker-compose.yml   # Docker 部署配置
├── .env.example         # 环境变量模板
└── README.md            # 本文件
```

## 环境变量说明

复制 `.env.example` 为 `.env` 后，修改以下关键配置：

| 变量 | 说明 | 必填 |
|------|------|------|
| POSTGRES_PASSWORD | PostgreSQL 密码 | ✅ |
| REDIS_PASSWORD | Redis 密码 | ✅ |
| RABBITMQ_PASSWORD | RabbitMQ 密码 | ✅ |
| JWT_SECRET | JWT密钥（至少32字符） | ✅ |
| ENCRYPT_KEY | 加密密钥（至少32字符） | ✅ |

## 默认账号

首次运行需要初始化管理员：
- 访问 http://localhost:5174/admin/init
- 按照向导创建管理员账号

## 常见问题

### 端口被占用
修改 `docker-compose.yml` 中的端口映射

### 数据库连接失败
确保 PostgreSQL 已启动并创建了数据库

### 前端无法访问后端
检查 `GAPI_FRONTEND_URL` 和 `GAPI_ADMIN_FRONTEND_URL` 配置

## 更多信息

- 项目地址: https://github.com/woshi-Tom/gApi-platform
- 文档: https://github.com/woshi-Tom/gApi-platform#readme
