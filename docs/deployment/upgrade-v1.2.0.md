# v1.2.0 升级指南

> 从 v1.1.x 升级到 v1.2.0 的步骤说明。

---

## 升级概要

| 项目 | 说明 |
|------|------|
| 数据库迁移 | ⚠️ 需要执行（无破坏性变更） |
| 环境变量 | 无新增/删减 |
| 配置变更 | 无破坏性变更 |
| 停机时间 | 滚动重启，无需停机 |
| 回滚难度 | 低（revert + 重新打 tag） |

---

## 升级步骤

### 1. 备份数据库

```bash
docker exec gapi-postgres pg_dump -U gapi -d gapi > backup-v1.1.x.sql
```

### 2. 拉取新版本

```bash
git fetch origin
git checkout v1.2.0
```

### 3. 更新服务

```bash
cd deploy/docker

# 后端
docker-compose up -d --build backend

# 数据库 migration（如需要）
docker exec gapi-backend ./gapi-server -config /app/config/config.yaml -migrate

# 前端
docker-compose up -d --build frontend admin
```

### 4. 验证

```bash
# API 健康检查
curl http://localhost:8080/health

# 前端访问
curl -I http://localhost:5173
curl -I http://localhost:5174
```

---

## 主要变更

### 新增功能
- 系统设置 API：通用设置、速率限制、安全设置（前后端完整）
- 审计日志增强：递归敏感字段脱敏、panic 恢复
- CI/CD：自动测试、lint 检查、产物 checksum

### 行为变更
- 审计日志默认跳过 GET 请求（支付路径除外），减少数据量
- 设置页面不再包含微信支付字段

---

## 回滚

如遇到问题，按以下步骤回滚：

```bash
git checkout <上一个版本>
cd deploy/docker
docker-compose up -d --build
```

详细回滚流程见 [`docs/development/git-workflow.md`](../development/git-workflow.md#7-回滚策略)。
