# gAPI Platform

**版本**: 1.1.0  
**日期**: 2026-04-13  
**状态**: 开发中

---

## 更新日志

> 详见 [CHANGELOG.md](./CHANGELOG.md)

---

## 项目概述

类似 OneAPI/NewAPI 的 API 代理平台，支持：
- 多租户架构
- VIP 用户体系 (30天过期)
- 商品购买 (支付宝/微信支付)
- 渠道管理 + 健康检查
- 完整审计日志

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go + Gin |
| 前端 | Vue 3 + Element Plus + TypeScript |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 消息队列 | RabbitMQ |

---

## 快速启动

```bash
cd deploy/docker
docker-compose up -d
```

访问地址：
- 用户前端: http://localhost:5173
- 管理后台: http://localhost:5174
- API: http://localhost:8080

---

## 目录结构

```
gapi-platform/
├── backend/                    # Go 后端
│   ├── cmd/                   # 入口
│   └── internal/              # 内部包
├── frontend/                   # Vue 3 前端
│   └── src/                  # 源码
├── docs/                      # 设计文档
│   ├── development-notes.md   # ⚠️ 开发前必读
│   └── *.md                  # 其他设计文档
└── deploy/
    ├── docker/                # Docker 部署
    │   ├── docker-compose.yml      # 开发环境
    │   ├── docker-compose.prod.yml # 生产环境
    │   └── docker-compose.test.yml  # 测试环境
    └── nginx/                # Nginx 配置
```

---

## 接口划分

| 接口类型 | 路径前缀 | 说明 |
|---------|---------|------|
| **北向** | `/api/v1/` | 用户端：注册、充值、调用AI API |
| **南向** | `/api/v1/internal/` | 内部：渠道管理、健康检查 |
| **管理后台** | `/api/v1/admin/` | 管理员：商品上下架、用户管理 |
| **初始化** | `/api/v1/init/` | 系统初始化向导 |

---

## 核心设计文档

| 文档 | 说明 |
|------|------|
| `development-notes.md` | ⚠️ **开发前必读** - 环境变量、接口清单、检查项 |
| `system-design.md` | 系统设计概览 |
| `database-design-v2.md` | 数据库完整DDL |
| `interface-design-south-north.md` | 北向/南向/管理后台接口 |
| `business-detail.md` | 业务细节：注册赠送、商品管理 |
| `security-deployment.md` | 安全与部署 |

---

## 部署

### 开发环境
```bash
cd deploy/docker
cp .env.example .env
docker-compose up -d
```

### 生产环境
```bash
cd deploy/docker
docker-compose -f docker-compose.prod.yml up -d
```

---

## 开发前必读

> 新开开发会话前，请先阅读 `docs/development-notes.md`
