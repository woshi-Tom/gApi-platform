# Issue #006: 前端全面修复 16 个 Bug

> 提交: `63ce677` | 分支: `dev-tom` | 修改: 9 个文件

---

## 修复清单

### Critical

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| 1 | `settings/Index.vue` | `onMounted` 调用 `loadGeneralSettings()` 但函数未定义，页面直接报错 | 添加函数，调用 `settingsAPI.getGeneralSettings()` 加载基本设置 |
| 2 | `channels/List.vue` | `del()` 直接删除无二次确认，违反设计文档要求 | 添加 `ElMessageBox.confirm`，导入 `ElMessageBox` |
| 3 | `products/List.vue` | 表单用 `is_popular`，后端字段 `is_hot`，热门标签永远无法保存 | 统一为 `is_hot`（defaultForm/handleEdit/handleSave/template） |

### High

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| 4 | `Dashboard.vue` | `ElMessage.error()` 未导入 | 添加 `import { ElMessage } from 'element-plus'` |
| 5 | `Payment.vue` | `status` 类型缺少 `completed`/`refunded`，后端可能返回这些值 | 补全为 6 种状态 |
| 6 | `Register.vue` | `captcha_token: 'bypass'` 硬编码，无验证令牌时发送无效字符串 | 仅在有 token 时才传字段 |
| 7 | `request.ts` | adminAPI 401 也跳 `/login`，管理后台应跳 `/admin/login` | `createRequest` 增加 `loginPath` 参数 |

### High (续)

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| 8 | `models/Index.vue` | `el-switch` 用 `true-value="true"` 字符串，后端期望布尔值 | 移除字符串 true-value/false-value |

### Medium

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| 9 | `models/Index.vue` | `el-checkbox` 废弃 `:label` 在新版 Element Plus 不兼容 | 改为 `:value` |
| 10 | `Dashboard.vue` | `Math.random()` 伪造趋势百分比 | 改用后端 `growth_rate` |
| 11 | `Dashboard.vue` | `tokenData` 计算后未在图表展示 | 添加第二条 series 展示 Token 消耗趋势 |
| 12 | `redemption/List.vue` | `formRef2` 声明后从未使用 | 删除 |
| 13 | `redemption/List.vue` | `create` 缺 `is_permanent`，永久 VIP 兑换码失效 | 补传字段 |

### Low

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| 14 | `products/List.vue` | `var` 声明 + 块级 scoping 问题 | 改为 `let` 并移到 if 前 |
| 15 | `products/List.vue` | `<script setup>` 缺 `lang="ts"` | 添加 |
| 16 | `channels/List.vue` | `testChannelModels` 有 5 处 `console.log` | 全部移除 |

---

## 变更文件清单

```
frontend/src/api/request.ts                  — createRequest 增加 loginPath 参数
frontend/src/views/Register.vue              — 移除 bypass 验证码
frontend/src/views/admin/Dashboard.vue       — ElMessage 导入 + token 趋势线 + 真实 growth_rate
frontend/src/views/admin/channels/List.vue   — 删除确认 + 清除 console.log
frontend/src/views/admin/models/Index.vue    — el-switch/el-checkbox 修复
frontend/src/views/admin/products/List.vue   — is_hot 字段统一 + var→let + lang="ts"
frontend/src/views/admin/redemption/List.vue — is_permanent 补传 + formRef2 清理
frontend/src/views/admin/settings/Index.vue  — loadGeneralSettings 函数
frontend/src/views/user/Payment.vue          — status 类型补全
```

---

## 待验证清单（给执行测试的 Agent）

### 1. 编译验证

```bash
cd frontend
npm run build       # 用户前端编译
npm run build:admin # 管理后台编译
```

**预期**: 两个 build 均无 TypeScript 错误、无 Vite 构建错误。

### 2. TypeScript 类型检查

```bash
cd frontend
npx vue-tsc --noEmit
```

**预期**: 无类型错误。重点关注：
- `Payment.vue` 的 status 类型联合
- `products/List.vue` 新增 `lang="ts"` 后的隐式 any 报错
- `models/Index.vue` 的 el-switch 布尔值类型

### 3. 功能验证（如可启动 dev server）

| 场景 | 步骤 | 预期 |
|------|------|------|
| Settings 页面加载 | 管理后台 → 系统设置 | 不报 JS 错误，基本设置三个字段有值 |
| 商品管理热门标签 | 编辑商品 → 勾选热门 → 保存 → 刷新 | `is_hot` 正确保存和回显 |
| 渠道删除确认 | 操作 → 删除 | 弹出确认对话框，取消不删除 |
| 模型定价开关 | 编辑定价 → 切换启用/特色 → 保存 | 后端收到布尔值 true/false |
| 兑换码永久VIP | 生成 VIP 兑换码 → 勾选永久VIP → 生成 | 生成的码包含 is_permanent |
| 注册流程 | 注册页 → 填写邮箱 → 点发送验证码 | 网络请求不包含 `captcha_token: "bypass"` |
| Dashboard | 管理后台仪表盘 | 无 console 报错，趋势图有两条线 |
| 401 重定向 | 管理后台 token 过期 | 跳转到 `/admin/login` 而非 `/login` |

### 4. 回归检查

```bash
# 确认无遗留调试代码
cd frontend/src
grep -r "console\.log" views/admin/channels/List.vue   # 应为空
grep -r "is_popular" views/admin/products/List.vue     # 应为空（已全部改为 is_hot）
grep -r "formRef2" views/admin/redemption/List.vue     # 应为空
grep -r "bypass" views/Register.vue                    # 应为空
```

---

## 设计依据

- 渠道删除二次确认 → `docs/features/channel-management-design.md` 第 242 行
- Payment 状态枚举 → `docs/features/payment-module-design.md` 定义 6 种状态
- 商品字段 `is_hot` → `docs/features/business-package-spec.md`
- admin 登录页 → 路由配置 `/admin/login`
