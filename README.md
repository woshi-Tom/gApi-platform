# gAPI Platform

**版本**: 1.2.0  
**日期**: 2026-05-12  
**状态**: 已发布

> ⚠️ **免责声明**: 本项目仅供学习交流使用，禁止用于任何非法用途。使用者需自行承担一切风险和责任。

[![Build and Release](https://github.com/woshi-Tom/gApi-platform/actions/workflows/build.yml/badge.svg)](https://github.com/woshi-Tom/gApi-platform/actions/workflows/build.yml)

---

## 项目概述

gAPI Platform 是一个类似 OneAPI/NewAPI 的 API 代理平台，专为学习和研究 AI API 集成技术而设计。

**主要功能**：
- 🤖 多渠道管理 - 支持 OpenAI、Claude、DeepSeek、NVIDIA 等多种 AI API
- 🔄 智能负载均衡 - 多渠道自动负载均衡和故障转移
- 🌐 SOCKS5/HTTP 代理支持 - 突破网络限制访问海外 API
- 💳 用户体系 - VIP 会员、套餐充值、支付宝支付
- 🎁 兑换码系统 - 兑换码生成、兑换、禁用、批次管理
- 📊 模型管理 - 模型分组、定价、用户组权限控制
- 🔒 安全设计 - API Key 加密存储、滑块验证码、完整权限控制

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
# 克隆项目
git clone https://github.com/woshi-Tom/gApi-platform.git
cd gApi-platform

# 复制环境变量配置
cp .env.example .env
# 编辑 .env 填写你的配置

# 启动服务
cd deploy/docker
docker-compose up -d
```

访问地址：
- 用户前端: http://localhost:5173
- 管理后台: http://localhost:5174
- API: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html

---

## 目录结构

```
gapi-platform/
├── backend/                    # Go 后端
│   ├── cmd/                   # 入口
│   └── internal/              # 内部包
├── frontend/                   # Vue 3 前端
│   └── src/                   # 源码
├── docs/                       # 设计文档
│   ├── development/            # 开发指南 ⚠️ 开发前必读
│   ├── architecture/           # 架构设计
│   ├── api/                    # API设计
│   ├── features/               # 功能文档（含实现状态）
│   ├── deployment/             # 部署运维
│   └── issues/                 # 问题追踪
├── deploy/                     # 部署配置
│   ├── docker/                # Docker Compose 环境
│   ├── release/               # Release 打包模板
│   └── nginx/                 # Nginx 反向代理
├── .github/workflows/          # GitHub Actions CI/CD
└── scripts/                    # 工具脚本
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
| [`docs/development/development-notes.md`](docs/development/development-notes.md) | ⚠️ **开发前必读** - 环境变量、接口清单、检查项 |
| [`docs/development/git-workflow.md`](docs/development/git-workflow.md) | Git 协作规范 + Release 发布流程 |
| [`docs/architecture/system-design.md`](docs/architecture/system-design.md) | 系统设计概览 |
| [`docs/architecture/database-design-v2.md`](docs/architecture/database-design-v2.md) | 数据库完整DDL |
| [`docs/api/interface-design-south-north.md`](docs/api/interface-design-south-north.md) | 北向/南向/管理后台接口 |
| [`docs/features/business-package-spec.md`](docs/features/business-package-spec.md) | 业务套餐规格（免费/充值/VIP） |
| [`docs/features/redemption-code-design.md`](docs/features/redemption-code-design.md) | 兑换码设计（已实现） |
| [`docs/features/model-group-pricing-permission-design.md`](docs/features/model-group-pricing-permission-design.md) | 模型分组/定价/权限（已实现） |
| [`docs/deployment/deployment.md`](docs/deployment/deployment.md) | Docker 部署文档 |
| [`docs/deployment/security-deployment.md`](docs/deployment/security-deployment.md) | 安全与部署指南 |

> 完整文档索引及功能状态总览见 [`docs/README.md`](docs/README.md) | [`docs/features/README.md`](docs/features/README.md)

---

## Release 发布

项目使用 GitHub Actions 自动构建和发布 Release。

### 发布流程

```bash
# 在 main 分支上打 tag 并推送
git checkout main
git tag v1.2.0
git push origin v1.2.0
```

### 下载产物

| 文件 | 适用平台 |
|------|----------|
| `gapi-platform-{VERSION}-linux-amd64.tar.gz` | Linux / macOS |
| `gapi-platform-{VERSION}-windows-amd64.zip` | Windows |

详细发布流程见 [`git-workflow.md`](docs/development/git-workflow.md#6-release-发布流程)。

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

## 更新日志

> 详见 [CHANGELOG.md](./CHANGELOG.md)

---

## 开发前必读

> 新开开发会话前，请先阅读 [`docs/development/development-notes.md`](docs/development/development-notes.md)
