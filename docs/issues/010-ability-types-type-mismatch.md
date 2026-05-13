# Issue #010: 模型定价 ability_types 类型不匹配

> 发现日期: 2026-05-13
> 来源: 用户测试 — 模型管理编辑保存报错
> 状态: ⬜ 待修复

---

## 一、问题描述

### 现象

模型管理 → 编辑定价 → 点击「保存」按钮 → 返回 400 错误：

```
json: cannot unmarshal string into Go struct field .ability_types of type []string
```

### 复现步骤

1. 管理后台 → 模型管理 → 模型定价
2. 点击任意定价记录的「编辑」按钮
3. 在弹出的对话框中修改或直接保存
4. 点击「保存」→ 报错

---

## 二、根因分析

### 数据流

```
前端 <el-input>  →  JSON: "ability_types": "chat,completion"     ← 字符串
                         ↓
后端 handler 接收 →  struct: AbilityTypes []string                ← 期望数组
                         ↓
                        json.Unmarshal 失败 ❌
```

### 具体代码

**前端** — `frontend/src/views/admin/models/Index.vue`

```vue
<!-- 第 142 行：用 el-input 输入字符串 -->
<el-form-item label="能力类型" prop="ability_types">
  <el-input v-model="pricingForm.ability_types" placeholder="如：text-davinci-003" />
</el-form-item>

<!-- 第 270 行：初始化为空字符串 -->
const pricingForm = reactive({
  ...
  ability_types: ''        // ← string 类型
})

<!-- 第 333 行：发送时仍为字符串 -->
ability_types: pricingForm.ability_types,  // ← 发送 "chat,completion"
```

**后端 handler** — `backend/internal/handler/model_pricing_handler.go`

```go
// 第 48 行：期望接收 JSON 数组
type CreatePricingRequest struct {
    ...
    AbilityTypes  []string `json:"ability_types"`   // ← 期望 ["chat","completion"]
}

// 第 74 行：调用 SetAbilityTypes
if len(req.AbilityTypes) > 0 {
    p.SetAbilityTypes(req.AbilityTypes)
}
```

**后端模型** — `backend/internal/model/model_group.go`

```go
// 第 66 行：数据库存储为 JSON 字符串
AbilityTypes  string  `json:"ability_types" gorm:"type:jsonb"`

// 第 84-100 行：辅助方法做 string ↔ []string 转换
func (m *ModelPricing) SetAbilityTypes(types []string) {
    b, _ := json.Marshal(types)
    m.AbilityTypes = string(b)     // 存为 '["chat","completion"]'
}
```

### 矛盾点

| 层级 | 类型 | 示例 |
|------|------|------|
| 数据库存储 (jsonb) | JSON 字符串 | `'["chat","completion"]'` |
| 模型 GetAbilityTypes | `[]string` | `["chat", "completion"]` |
| 后端 handler 接收 | `[]string` | 期望 `["chat","completion"]` |
| 前端 el-input 发送 | `string` | 发送 `"chat,completion"` |

前端用普通文本输入框，用户输入 `chat,completion`，作为字符串发送。后端期望 JSON 数组，导致反序列化失败。

---

## 三、修复方案

### 方案 A：前端改用多选组件（推荐）

**思路**：将 `<el-input>` 改为 `<el-select multiple>`，以数组形式发送。

**前端改动** — `frontend/src/views/admin/models/Index.vue`

```vue
<!-- 改前 -->
<el-input v-model="pricingForm.ability_types" />

<!-- 改后 -->
<el-select v-model="pricingForm.ability_types" multiple placeholder="选择能力类型">
  <el-option label="聊天 (chat)" value="chat" />
  <el-option label="补全 (completion)" value="completion" />
  <el-option label="嵌入 (embedding)" value="embedding" />
  <el-option label="审核 (moderation)" value="moderation" />
  <el-option label="语音 (audio)" value="audio" />
  <el-option label="视觉 (vision)" value="vision" />
</el-select>
```

同时修改初始化：
```ts
// 改前
ability_types: ''

// 改后
ability_types: [] as string[]
```

以及从行数据读取时：
```ts
// 改前
ability_types: row.ability_types ?? ''

// 改后
ability_types: row.ability_types ? JSON.parse(row.ability_types) : []
```

**优点**：用户从预定义选项中选择，体验好，无输入错误
**缺点**：需要前端改动

---

### 方案 B：后端兼容字符串输入

**思路**：handler 中先判断 `ability_types` 类型，如果是字符串则转为 `[]string`。

**后端改动** — `backend/internal/handler/model_pricing_handler.go`

```go
// 接收时用 interface{} 兼容两种格式
AbilityTypes  interface{} `json:"ability_types"`
```

然后在处理逻辑中：
```go
func parseAbilityTypes(v interface{}) []string {
    switch t := v.(type) {
    case []interface{}:
        result := make([]string, 0, len(t))
        for _, item := range t {
            if s, ok := item.(string); ok {
                result = append(result, s)
            }
        }
        return result
    case string:
        if t == "" {
            return []string{"chat"}
        }
        return strings.Split(t, ",")
    }
    return []string{"chat"}
}
```

**优点**：前端不改也能工作，兼容旧数据
**缺点**：handler 逻辑变复杂，字符串转数组的分隔符策略不明确（逗号？空格？）

---

### 方案 C：前后端同时改（推荐组合）

前端用多选组件（方案 A）+ 后端清理默认值逻辑。

**额外优化**：后端在 ability_types 为空时使用默认值 `["chat"]`，当前代码已有此逻辑：

```go
if len(req.AbilityTypes) > 0 {
    p.SetAbilityTypes(req.AbilityTypes)
} else {
    p.SetAbilityTypes([]string{"chat"})  // 默认值
}
```

所以只需前端发对格式，后端无需改动。

---

## 四、各角色评估

| 角色 | 观点 |
|------|------|
| 📋 **产品经理** | 能力类型应该让用户选择而非手动输入。预定义 6 种类型（chat/completion/embedding/moderation/audio/vision）覆盖所有场景。方案 A 最佳。 |
| 📊 **项目经理** | 方案 A 只改前端 1 个文件，无后端风险。与 #009 的 el-switch 修复同属模型管理页面，可一起出。 |
| 💻 **开发** | 方案 A 实现简单，el-select multiple 是 Element Plus 标准用法。注意回显时 JSON 字符串解析。 |
| 🧪 **QA/测试** | 验证点：多选 → 保存 → 刷新回显 ✅；单选 ✅；清空 → 默认 "chat" ✅；空数据兼容 ✅ |
| 🎨 **UI/前端** | 多选下拉比文本框更直观，用户不需要记忆类型名称。 |
| 🔒 **安全** | 无安全影响。预定义选项限制了输入范围，反而更安全。 |

---

## 五、建议执行方案

**推荐：方案 A（仅前端改动）**

| 文件 | 改动 |
|------|------|
| `frontend/src/views/admin/models/Index.vue` | el-input → el-select multiple；初始化改为 `[]`；回显时 JSON.parse |

**工时**：开发 0.3h + 测试 0.3h

---

## 六、验证清单

| # | 场景 | 操作 | 预期 |
|---|------|------|------|
| T1 | 多选保存 | 选 2 个类型 → 保存 → 刷新 | 回显正确 |
| T2 | 单选保存 | 选 1 个类型 → 保存 → 刷新 | 回显正确 |
| T3 | 清空保存 | 清空所有 → 保存 → 刷新 | 默认 chat |
| T4 | 新增记录 | 新建 → 选类型 → 保存 | 成功创建 |
| T5 | 列表回显 | 查看模型定价列表 | ability_types 列显示正确 |

---

## 七、决策

| 决策项 | 结论 |
|--------|------|
| 是否修复 | ⬜ 待定 |
| 采用方案 | ⬜ 方案 A / 方案 B / 方案 C |
| 责任人 | ⬜ |
| 排期 | ⬜ 当前 Sprint / 后续 Sprint |
