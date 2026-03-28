# gAPI Platform Docker Setup

一键启动 gAPI Platform 完整开发环境。

## 架构

```
┌─────────────────────────────────────────┐
│              Docker Network              │
├─────────────────────────────────────────┤
│  postgres:5432   - PostgreSQL 16         │
│  redis:6379     - Redis 7               │
│  backend:8080  - Go API Server         │
│  frontend:80   - User Dashboard (5173)  │
│  admin:80      - Admin Panel (5174)     │
└─────────────────────────────────────────┘
```

## 快速启动

```bash
cd deploy/docker

# 复制环境变量配置
cp .env.example .env

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f backend
```

## 服务地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 用户前端 | http://localhost:5173 | 用户仪表盘 |
| 管理后台 | http://localhost:5174/admin.html | 管理员面板 |
| API 服务 | http://localhost:8080 | 后端 API |
| Swagger | http://localhost:8080/swagger/index.html | API 文档 |
| PostgreSQL | localhost:5432 | 数据库 |
| Redis | localhost:6379 | 缓存 |

## 测试账号

- **管理员**: admin / admin123
- **操作员**: operator / operator123

## 常用命令

```bash
# 启动
docker-compose up -d

# 停止
docker-compose down

# 停止并删除数据卷
docker-compose down -v

# 重建服务
docker-compose up -d --force-recreate

# 进入后端容器
docker exec -it gapi-backend sh

# 进入数据库
docker exec -it gapi-postgres psql -U gapi -d gapi

# 查看后端日志
docker logs -f gapi-backend

# 重启后端
docker-compose restart backend
```

## 开发模式

修改代码后自动重建：

```bash
docker-compose up -d --build backend
```

## 环境变量

在 `.env` 文件中配置：

```env
# 数据库
POSTGRES_DB=gapi
POSTGRES_USER=gapi
POSTGRES_PASSWORD=gapi123

# Redis
REDIS_PASSWORD=redis123

# JWT 密钥 (生产环境必须修改!)
GAPI_JWT_SECRET=your-super-secret-jwt-key-at-least-32-characters

# 管理员密钥
GAPI_ADMIN_SECRET=gapi-admin-secret-key-2026
```

## 生产部署注意

1. 修改所有默认密码
2. 使用强 JWT 密钥
3. 配置 HTTPS (使用 nginx + letsencrypt)
4. 启用防火墙
5. 考虑使用 Docker Swarm 或 Kubernetes

## 清理

```bash
# 删除所有容器和卷
docker-compose down -v --rmi all

# 完全清理
docker system prune -a
```
