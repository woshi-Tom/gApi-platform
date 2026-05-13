# #011 编译测试指南

> 日期: 2026-05-13
> 关联: [010-ability-types-type-mismatch.md](010-ability-types-type-mismatch.md) K13 修复 + 代码清理
> 状态: 待编译验证
> 目标: 验证 K13 后端兼容修复、死代码清理、不安全类型断言修复的编译通过和基础功能正确性

---

## 一、需要编译验证的变更

| 文件 | 变更内容 |
|------|---------|
| `backend/internal/handler/model_pricing_handler.go` | 新增 `parseAbilityTypes()` 辅助函数；Create/Update 中 `AbilityTypes` 类型从 `[]string` 改为 `interface{}`，运行时兼容字符串和数组两种格式 |
| `backend/internal/handler/api_handler.go` | 删除 6 个死代码函数：`handleStream()`、`logUsage()`、`parseSSEStream()`、`getChannelID()`、`setChannelID()`、`getModelName()`、`parseIntParam()`；`ListModels()` 中 `userID.(uint)` 不安全断言改为安全断言；清理 `bufio`、`strconv`、`strings` 无用 import |

---

## 二、编译步骤

### 2.1 Go 编译

```bash
cd backend
go build ./...
```

**预期结果**: 无编译错误，无 warning。

### 2.2 Go Vet 静态检查

```bash
cd backend
go vet ./...
```

**预期结果**: 无问题报告。

### 2.3 现有单元测试

```bash
cd backend
go test ./... -count=1 -timeout 120s
```

**预期结果**: 所有现有测试通过（不应有 regression）。

---

## 三、功能验证清单

### 3.1 T-K13: ability_types 后端兼容

```
验证点 1: 字符串格式（模拟当前前端行为）
  POST /api/v1/admin/model-pricings
  Body: {"model":"gpt-4","ability_types":"chat,completion"}
  预期: 200，ability_types 存储为 ["chat","completion"]

验证点 2: JSON 数组格式（标准格式）
  POST /api/v1/admin/model-pricings
  Body: {"model":"gpt-3.5-turbo","ability_types":["chat","completion"]}
  预期: 200，ability_types 存储为 ["chat","completion"]

验证点 3: 空字符串
  POST /api/v1/admin/model-pricings
  Body: {"model":"test","ability_types":""}
  预期: 200，默认 ability_types 为 ["chat"]

验证点 4: null 值
  POST /api/v1/admin/model-pricings
  Body: {"model":"test","ability_types":null}
  预期: 200，默认 ability_types 为 ["chat"]

验证点 5: Update 兼容（PUT）
  PUT /api/v1/admin/model-pricings/{id}
  Body: {"ability_types":"embedding,chat"}
  预期: 200，更新成功

验证点 6: 回显正确
  GET /api/v1/admin/model-pricings/{id}
  预期: ability_types 返回 JSON 数组格式
```

### 3.2 T-CLEAN: 死代码清理

```
验证点 1: 编译无引用错误
  go build ./...
  预期: 无 "undefined" 错误

验证点 2: Go Vet 无警告
  go vet ./...
  预期: 无 "unused" 警告
```

### 3.3 T-SAFE: 不安全类型断言

```
验证点 1: ListModels 正常调用（有 user_id）
  GET /v1/models
  Authorization: Bearer sk-ap-xxxxx
  预期: 200，返回模型列表

验证点 2: ListModels 无 user_id 不 panic
  GET /v1/models（无认证或 user_id 类型异常）
  预期: 200，返回可用模型列表（降级为全量），不 panic
```

---

## 四、验证结论模板

```
编译结果: ✅ 通过 / ❌ 失败（错误信息: ...）
Go Vet:   ✅ 通过 / ❌ 失败
现有测试: ✅ 全部通过 / ❌ 有失败（用例: ...）
T-K13 字符串:  ✅ / ❌
T-K13 数组:    ✅ / ❌
T-K13 空/null: ✅ / ❌
T-K13 Update:  ✅ / ❌
T-CLEAN:       ✅ / ❌
T-SAFE:        ✅ / ❌
结论:     可合入 / 需修复后重新验证
```
