# 项目功能迭代技能

> 当用户要求实现新功能或修复功能时使用此技能

## 触发条件

以下任一场景时激活：
- 用户说"实现XXX功能"
- 用户说"开发XXX模块"
- 用户说"加上XXX功能"
- 用户说"这个功能有问题，修复一下"
- 用户说"添加XXX到XXX"

## 执行流程

### 1. 理解需求

```markdown
1. 明确用户要实现的具体功能
2. 确认功能边界（什么做，什么不做）
3. 检查是否需要设计文档
```

### 2. 检查现有代码

```bash
# 搜索相关模块
grep -r "keyword" backend/internal/
grep -r "keyword" frontend/src/

# 查看模块结构
ls -la backend/internal/handler/
ls -la frontend/src/views/
```

### 3. 创建Todo列表

在 `.sisyphus/plans/功能名-todo.md` 创建：

```markdown
# [功能名] 实现计划

## 功能需求
[用户需求描述]

## 实现步骤
- [ ] 步骤1：xxx
- [ ] 步骤2：xxx

## 验证清单
- [ ] 构建成功
- [ ] 自动化测试通过
- [ ] 功能体验测试通过
- [ ] 无bug
- [ ] Git提交完成
```

### 4. 迭代实现

每个功能点必须经历：
```
实现 → 构建 → 测试 → 发现bug → 修复 → 重新构建 → 测试 → ... → 完成
```

### 5. 自动化测试

```bash
# 后端测试
cd backend && go test ./...

# 前端构建
cd frontend && npm run build

# 前端类型检查
cd frontend && npm run type-check
```

### 6. 功能体验测试

必须验证：
- 页面能正常加载
- 数据能正确保存
- 状态能正确更新
- 错误能正确提示

### 7. Git提交

```bash
git add .
git commit -m "feat: 实现xxx功能"
```

### 8. 完成标准

功能从Todo移除的条件：
- [x] 所有代码修改已提交
- [x] 构建通过
- [x] 功能体验测试通过
- [x] 无已知bug

## 文档更新

功能完成后更新相关文档：
- 文档状态从"待实现"改为"已实现"
- 添加实现说明

## 示例流程

用户："实现渠道管理的新增功能"

1. 检查现有代码：
   - 查看 channel_handler.go
   - 查看 channels/List.vue

2. 创建Todo：
   ```markdown
   # 渠道管理-新增渠道 实现计划
   
   ## 实现步骤
   - [ ] 创建后端接口 (POST /api/v1/admin/channels)
   - [ ] 创建前端表单组件
   - [ ] 添加路由和API调用
   - [ ] 测试功能
   ```

3. 迭代实现每个步骤

4. 完成验证后提交
