# gAPI Platform 部署指南

## 系统要求

- Docker 20.10+
- Docker Compose v2.0+
- 最低配置: 2 CPU, 4GB RAM
- 推荐配置: 4 CPU, 8GB RAM

## 快速部署

### 1. 克隆代码

```bash
git clone https://your-repo/gapi-platform.git
cd gapi-platform
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置以下必填项:

```bash
DB_PASSWORD=your_secure_database_password
REDIS_PASSWORD=your_secure_redis_password
RABBITMQ_PASSWORD=your_secure_rabbitmq_password
JWT_SECRET=your_jwt_secret_at_least_32_characters
ADMIN_SECRET=your_admin_panel_secret
```

### 3. 启动服务

```bash
# 拉取镜像并启动所有服务
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f
```

### 4. 验证部署

```bash
# 检查后端健康状态
curl http://localhost:8080/health

# 检查前端
curl http://localhost

# 检查管理后台
curl http://localhost/admin/
```

## 服务访问

| 服务 | 端口 | URL |
|------|------|-----|
| 用户前端 | 80 | http://your-domain.com |
| 管理后台 | 80 | http://your-domain.com/admin/ |
| API 服务 | 8080 | http://your-domain.com/api/ |
| RabbitMQ 管理 | 15672 | http://your-domain.com:15672 |

## 默认账户

| 类型 | 账户 | 密码 |
|------|------|------|
| 管理员 | admin | admin123 |
| 用户 | admin@example.com | admin123 |

**重要**: 首次部署后请立即修改默认密码!

## HTTPS 配置

### 使用 Let's Encrypt (推荐)

```bash
# 安装 certbot
apt install certbot python3-certbot-nginx

# 申请证书
certbot certonly --nginx -d your-domain.com

# 复制证书到项目目录
mkdir -p nginx/ssl
cp /etc/letsencrypt/live/your-domain.com/fullchain.pem nginx/ssl/cert.pem
cp /etc/letsencrypt/live/your-domain.com/privkey.pem nginx/ssl/key.pem

# 重启 nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

### 自定义证书

将证书文件放入 `nginx/ssl/` 目录:
- `cert.pem` - 证书
- `key.pem` - 私钥

## 数据备份

### 备份 PostgreSQL

```bash
# 创建备份目录
mkdir -p backups

# 备份数据库
docker exec gapi-postgres pg_dump -U gapi gapi > backups/gapi_$(date +%Y%m%d).sql

# 恢复数据库
cat backups/gapi_20260328.sql | docker exec -i gapi-postgres psql -U gapi gapi
```

### 备份 Redis

```bash
# 备份 Redis 数据
docker exec gapi-redis redis-cli -a $REDIS_PASSWORD SAVE
docker cp gapi-redis:/data/dump.rdb backups/redis_$(date +%Y%m%d).rdb
```

### 自动备份脚本

创建 `scripts/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

# 备份数据库
docker exec gapi-postgres pg_dump -U gapi gapi > $BACKUP_DIR/db_$DATE.sql

# 备份 Redis
docker exec gapi-redis redis-cli -a $REDIS_PASSWORD SAVE
docker cp gapi-redis:/data/dump.rdb $BACKUP_DIR/redis_$DATE.rdb

# 保留最近 7 天的备份
find $BACKUP_DIR -mtime +7 -delete

echo "Backup completed: $DATE"
```

## 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建并启动
docker-compose -f docker-compose.prod.yml up -d --build

# 查看更新后的版本
docker-compose -f docker-compose.prod.yml logs backend | grep "Starting"
```

## 停止服务

```bash
# 停止所有服务 (保留数据卷)
docker-compose -f docker-compose.prod.yml down

# 停止并删除数据卷 (危险! 会删除所有数据)
docker-compose -f docker-compose.prod.yml down -v
```

## 故障排除

### 查看服务日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f backend
docker-compose -f docker-compose.prod.yml logs -f postgres
```

### 检查服务状态

```bash
# 检查容器状态
docker-compose -f docker-compose.prod.yml ps

# 检查容器健康状态
docker inspect gapi-backend | grep -A 10 "Health"
```

### 常见问题

**Q: 502 Bad Gateway**
```bash
# 检查后端是否启动
docker-compose -f docker-compose.prod.yml logs backend

# 重启后端
docker-compose -f docker-compose.prod.yml restart backend
```

**Q: 数据库连接失败**
```bash
# 检查数据库容器
docker-compose -f docker-compose.prod.yml logs postgres

# 等待数据库就绪后重启后端
docker-compose -f docker-compose.prod.yml restart backend
```

**Q: 前端无法访问 API**
```bash
# 检查网络连接
docker network inspect gapi-platform_gapi-network

# 重启所有服务
docker-compose -f docker-compose.prod.yml restart
```

## 生产环境优化

### 1. 配置防火墙

```bash
# 只开放必要端口
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

### 2. 配置日志轮转

编辑 `docker-compose.prod.yml` 添加:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### 3. 资源限制

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
```

### 4. 监控 (使用 Prometheus)

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

## 目录结构

```
gapi-platform/
├── docker-compose.prod.yml    # 生产环境配置
├── .env.example             # 环境变量模板
├── nginx/
│   ├── nginx.conf           # Nginx 配置
│   └── ssl/                # SSL 证书目录
├── backend/
│   ├── Dockerfile          # 后端镜像构建
│   └── ...
├── frontend/
│   ├── Dockerfile          # 用户前端镜像构建
│   ├── Dockerfile.admin    # 管理后台镜像构建
│   └── ...
└── backups/                # 备份文件目录 (需手动创建)
```

## 技术支持

- 文档: `/docs/`
- 问题反馈: GitHub Issues
