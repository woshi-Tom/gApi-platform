# gAPI Platform

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待开发

---

## 项目概述

全新设计的 gAPI Platform，类似 OneAPI/NewAPI，支持：
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
| 消息队列 | RabbitMQ (可选) |

---

## 目录结构

```
gapi-platform/
├── backend/            # Go 后端
│   ├── cmd/           # 入口
│   ├── internal/      # 内部包
│   ├── config/        # 配置
│   └── scripts/       # 脚本
│
├── frontend/           # Vue 3 前端
│   ├── src/           # 源码
│   └── public/        # 静态资源
│
├── docs/              # 设计文档
│   ├── system-design.md
│   ├── database-design-v2.md
│   ├── interface-design-south-north.md
│   ├── project-structure.md
│   ├── security-deployment.md
│   └── business-detail.md
│
├── config/            # 配置文件
└── scripts/           # 工具脚本
```

---

## 接口划分

| 接口类型 | 路径前缀 | 说明 |
|---------|---------|------|
| **北向** | `/api/v1/` | 用户端：注册、充值、调用AI API |
| **南向** | `/api/v1/internal/` | 内部：渠道管理、健康检查 |
| **管理后台** | `/api/v1/admin/` | 管理员：商品上下架、用户管理 (内网) |

---

## 核心设计文档

1. **system-design.md** - 系统设计概览
2. **database-design-v2.md** - 数据库完整DDL
3. **interface-design-south-north.md** - 北向/南向/管理后台接口分离
4. **business-detail.md** - 业务细节：注册赠送、商品管理
5. **security-deployment.md** - 安全与部署
6. **development-notes.md** - 开发实现清单 (含环境变量、接口清单、检查项)

---

## 开发前必读

> 新开开发会话前，请先阅读 `docs/development-notes.md`，包含：
> - 环境变量配置清单
> - 数据库表创建顺序
> - API接口完整清单
> - 敏感数据脱敏规则
> - 支付回调处理
> - VIP系统细节
> - 通道测试类型
> - 日志统计展示格式
> - 边缘 case 处理
> - 前后端实现检查项

---

## 开始开发

详见各设计文档