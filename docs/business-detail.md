# API Proxy Platform - 业务详细设计文档 v1.0

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 业务总览

### 1.1 核心业务流程

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           核心业务流程图                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐ │
│  │  访客       │───▶│  注册       │───▶│  免费体验   │───▶│  充值/购买  │ │
│  │  (游客)     │    │  (获赠配额)  │    │  (体验额度)  │    │  (VIP/套餐)  │ │
│  └─────────────┘    └─────────────┘    └─────────────┘    └──────┬──────┘ │
│                                                                    │        │
│                                                                    ▼        │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐ │
│  │  使用API    │◀───│  消耗配额   │◀───│  配额不足   │◀───│  订单支付   │ │
│  │  (调用LLM)  │    │  (Token扣减) │    │  (提示充值)  │    │  (支付成功)  │ │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 用户体系

| 角色 | 说明 | 访问入口 | IP限制 |
|------|------|---------|--------|
| 超级管理员 | 系统最高权限 | 专用登录页 | **仅内网** |
| 普通管理员 | 管理后台操作 | 专用登录页 | **仅内网** |
| 普通用户 | 前台购买/使用 | 用户登录页 | 无限制 |
| 访客 | 未登录用户 | 首页浏览 | 无限制 |

---

## 2. 新用户注册赠送机制

### 2.1 赠送策略配置

```sql
-- 新用户配置表
CREATE TABLE signup_configs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 赠送配置
    enabled         BOOLEAN DEFAULT TRUE,                  -- 是否启用
    quota_amount    BIGINT DEFAULT 100000,                 -- 赠送配额数量
    quota_type      VARCHAR(10) DEFAULT 'permanent',       -- permanent(永久)|vip(体验VIP)
    
    -- VIP体验配置 (可选)
    trial_vip_days  INTEGER DEFAULT 0,                     -- VIP体验天数
    trial_quota     BIGINT DEFAULT 0,                      -- VIP体验配额
    
    -- 限制
    per_ip_limit    INTEGER DEFAULT 3,                     -- 同一IP最大注册数
    per_email_verification BOOLEAN DEFAULT TRUE,           -- 是否需要邮箱验证
    
    -- 有效期
    valid_from      TIMESTAMPTZ,                          -- 生效开始时间
    valid_until     TIMESTAMPTZ,                          -- 生效结束时间
    description     VARCHAR(200),                         -- 描述
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    
    CONSTRAINT uk_signup_config_tenant UNIQUE (tenant_id)
);
```

### 2.2 注册流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      新用户注册流程                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. 用户提交注册                                                 │
│     POST /api/v1/user/auth/register                            │
│     {username, email, password}                                │
│                         │                                      │
│                         ▼                                      │
│  2. 验证注册配置                                                 │
│     ├── 检查是否启用注册                                        │
│     ├── 检查IP注册限制                                          │
│     ├── 检查邮箱格式                                            │
│     └── 检查密码强度                                            │
│                         │                                      │
│                         ▼                                      │
│  3. 创建用户                                                     │
│     ├── 写入 users 表                                          │
│     ├── 生成邮箱验证Token                                       │
│     └── 记录注册来源                                            │
│                         │                                      │
│                         ▼                                      │
│  4. 赠送配额                                                     │
│     ├── 检查赠送配置                                            │
│     ├── 当前用户配额 = quota_amount                            │
│     ├── 记录 quota_transactions                                │
│     └── 记录 audit_log                                         │
│                         │                                      │
│                         ▼                                      │
│  5. 发送验证邮件                                                │
│     ├── 发送验证链接                                            │
│     └── 记录 signup_bonus (防止重复赠送)                        │
│                         │                                      │
│                         ▼                                      │
│  6. 返回结果                                                     │
│     {user_id, username, email, quota: 100000}                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 实现代码

```go
// 注册请求
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50,alphanumunicode"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8,max=128"`
}

type RegisterResponse struct {
    UserID      int64  `json:"user_id"`
    Username    string `json:"username"`
    Email       string `json:"email"`
    Quota       int64  `json:"quota"`        // 获赠的配额
    QuotaType   string `json:"quota_type"`   // permanent | vip
    TrialVIPDays int   `json:"trial_vip_days"` // VIP体验天数
    NeedVerify  bool   `json:"need_verify"`  // 是否需要邮箱验证
}

// 注册服务
func (s *AuthService) Register(ctx *gin.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // 1. 验证注册配置
    config := s.getSignupConfig()
    if !config.Enabled {
        return nil, ErrRegistrationClosed
    }
    
    // 2. 检查IP注册限制
    if err := s.checkIPLimit(ctx.ClientIP(), config.PerIPLimit); err != nil {
        return nil, err
    }
    
    // 3. 创建用户
    user := &model.User{
        Username:      req.Username,
        Email:         req.Email,
        PasswordHash:  s.hashPassword(req.Password),
        Level:         "free",
        Status:        "active",
    }
    
    // 4. 赠送配额
    if config.QuotaAmount > 0 {
        user.RemainQuota = config.QuotaAmount
    }
    
    // 5. VIP体验
    if config.TrialVIPDays > 0 {
        user.Level = "vip"
        user.VIPExpiredAt = time.Now().AddDate(0, 0, config.TrialVIPDays)
        user.VIPQuota = config.TrialQuota
    }
    
    // 6. 保存用户
    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // 7. 记录配额流水
    if user.RemainQuota > 0 {
        s.quotaService.RecordTransaction(user.ID, "signup_bonus", "permanent", 
            int64(config.QuotaAmount), user.RemainQuota, "新用户注册赠送")
    }
    
    // 8. 发送验证邮件
    if config.PerEmailVerification {
        s.sendVerificationEmail(user)
    }
    
    return &RegisterResponse{
        UserID:        user.ID,
        Username:      user.Username,
        Email:         user.Email,
        Quota:         user.RemainQuota,
        QuotaType:     "permanent",
        TrialVIPDays:  config.TrialVIPDays,
        NeedVerify:    config.PerEmailVerification,
    }, nil
}
```

---

## 3. 页面隔离设计

### 3.1 前端路由结构

```
frontend/src/
├── views/
│   ├── admin/                    # 管理后台 (仅内网访问)
│   │   ├── Login.vue             # 管理员登录
│   │   ├── Dashboard.vue         # 仪表盘
│   │   ├── channels/
│   │   ├── users/
│   │   ├── orders/
│   │   ├── products/             # 商品管理 ⭐
│   │   │   ├── List.vue          # 商品列表
│   │   │   ├── Form.vue          # 创建/编辑商品
│   │   │   └── Category.vue      # 分类管理
│   │   ├── vip/
│   │   ├── recharge/             # 充值套餐
│   │   ├── audit/                # 审计日志
│   │   └── settings/
│   │
│   └── user/                     # 用户前台 (公网访问)
│       ├── Login.vue             # 用户登录
│       ├── Register.vue          # 用户注册 (含赠送)
│       ├── Dashboard.vue         # 用户仪表盘
│       ├── Tokens.vue            # API Key管理
│       ├── Products.vue          # 商品列表 ⭐
│       ├── ProductDetail.vue     # 商品详情
│       ├── Recharge.vue          # 充值
│       ├── VIP.vue               # VIP购买
│       ├── Orders.vue            # 我的订单
│       ├── Usage.vue             # 用量明细
│       └── Profile.vue           # 个人资料
```

### 3.2 路由守卫

```typescript
// router/guards.ts

// 用户端路由守卫
const userRouterGuard = (to: RouteLocationNormalized, from: RouteLocationNormalizedNormalized) => {
  const userStore = useUserStore()
  
  // 需要登录的页面
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    return '/user/login'
  }
  
  // 已登录禁止访问
  if (to.meta.guestOnly && userStore.isLoggedIn) {
    return '/user/dashboard'
  }
}

// 管理员端路由守卫
const adminRouterGuard = (to: RouteLocationNormalized, from: RouteLocationNormalizedNormalized) => {
  const adminStore = useAdminStore()
  const userStore = useUserStore()
  
  // 检查IP是否在 内网范围内
  const clientIP = getClientIP()
  if (!isIntranetIP(clientIP) && to.path !== '/admin/login') {
    return '/user/login' // 重定向到用户端
  }
  
  // 需要管理员登录
  if (to.meta.requiresAdminAuth && !adminStore.isLoggedIn) {
    return '/admin/login'
  }
  
  // 禁止普通用户访问管理后台
  if (userStore.isLoggedIn && to.path.startsWith('/admin')) {
    return '/user/dashboard'
  }
}

// IP检查工具
function isIntranetIP(ip: string): boolean {
  // 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
  const intranet = /^10\.|^172\.(1[6-9]|2[0-9]|3[0-1])\.|^192\.168\./
  return intranet.test(ip) || ip === '127.0.0.1' || ip === 'localhost'
}
```

### 3.3 入口分离

```
┌─────────────────────────────────────────────────────────────────┐
│                         入口文件分离                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用户端入口    │  管理端入口                                     │
│  ─────────    │  ─────────                                     │
│  /            │  /admin                                        │
│  /user/login  │  /admin/login                                  │
│  /user/reg    │  /admin/dashboard                              │
│  /user/*      │  /admin/*                                      │
│                                                                 │
│  前端Nginx配置:                                                 │
│  ────────────                                                  │
│  location / {                   # 用户端                       │
│      root /usr/share/nginx/html/user;                         │
│      index index.html;                                         │
│      try_files $uri $uri/ /user/index.html;                    │
│  }                                                             │
│                                                                 │
│  location /admin {             # 管理端                         │
│      root /usr/share/nginx/html/admin;                         │
│      index index.html;                                         │
│      try_files $uri $uri/ /admin/index.html;                  │
│                                                                 │
│      # 内网IP限制                                               │
│      allow 10.0.0.0/8;                                         │
│      allow 172.16.0.0/12;                                      │
│      allow 192.168.0.0/16;                                     │
│      deny all;                                                  │
│  }                                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. 商品管理系统 (管理员)

### 4.1 商品表设计

```sql
-- ============================================================
-- 商品表 - 管理员上架/下架商品
-- ============================================================
CREATE TABLE products (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    -- 商品信息
    name            VARCHAR(200) NOT NULL,                  -- 商品名称
    description     TEXT,                                  -- 商品描述
    category_id     BIGINT,                                -- 分类ID
    cover_image     VARCHAR(500),                          -- 封面图
    
    -- 商品类型
    product_type    VARCHAR(20) NOT NULL,                   -- recharge(充值)|vip(VIP)|package(套餐)
    
    -- 价格配置
    price           DECIMAL(10,2) NOT NULL,                -- 售价
    original_price  DECIMAL(10,2),                         -- 原价
    discount        DECIMAL(5,2),                          -- 折扣率
    
    -- 配额配置
    quota           BIGINT NOT NULL,                       -- 配额(tokens)
    bonus_quota     BIGINT DEFAULT 0,                      -- 赠送配额
    
    -- VIP配置
    vip_days        INTEGER DEFAULT 0,                     -- VIP天数 (0表示非VIP)
    vip_quota       BIGINT DEFAULT 0,                      -- VIP配额
    
    -- 显示配置
    sort_order      INTEGER DEFAULT 0,                     -- 排序
    is_recommended  BOOLEAN DEFAULT FALSE,                 -- 推荐
    is_hot          BOOLEAN DEFAULT FALSE,                 -- 热门
    tags            JSONB DEFAULT '[]',                   -- 标签
    
    -- 上架状态
    status          VARCHAR(20) DEFAULT 'draft',            -- draft(草稿)|active(上架)|inactive(下架)
    published_at    TIMESTAMPTZ,                           -- 上架时间
    offline_reason  VARCHAR(200),                          -- 下架原因
    
    -- 限制配置
    max_per_user    INTEGER DEFAULT 0,                      -- 每人限购(0不限)
    total_stock     INTEGER DEFAULT 0,                      -- 库存(0不限)
    sold_count      INTEGER DEFAULT 0,                      -- 已售
    
    -- 审计字段
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT,
    updated_by      BIGINT,
    deleted_at      TIMESTAMPTZ
);

-- 索引
CREATE INDEX idx_products_tenant ON products(tenant_id);
CREATE INDEX idx_products_type ON products(product_type);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_status ON products(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_sort ON products(sort_order);

-- 商品分类表
CREATE TABLE product_categories (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT REFERENCES tenants(id),
    
    name            VARCHAR(100) NOT NULL,
    parent_id       BIGINT,                                -- 父分类
    sort_order      INTEGER DEFAULT 0,
    status          VARCHAR(20) DEFAULT 'active',
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);
```

### 4.2 管理员商品管理API

```go
// 商品管理handler

// 获取商品列表
func (h *AdminHandler) GetProducts(c *gin.Context) {
    params := &ProductListParams{
        Page:        pagination.GetPage(c),
        PageSize:    pagination.GetPageSize(c),
        ProductType: c.Query("product_type"),
        CategoryID:  toInt64(c.Query("category_id")),
        Status:      c.Query("status"),
        Keyword:     c.Query("keyword"),
    }
    
    products, total := h.productService.List(params)
    
    response.Success(c, gin.H{
        "total": total,
        "list":  products,
    })
}

// 创建商品
func (h *AdminHandler) CreateProduct(c *gin.Context) {
    var req ProductCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, "参数错误")
        return
    }
    
    // 验证
    if req.Name == "" || req.Price <= 0 || req.Quota <= 0 {
        response.Fail(c, "名称、价格、配额不能为空")
        return
    }
    
    // 创建商品
    product, err := h.productService.Create(&req, getAdminID(c))
    if err != nil {
        response.Fail(c, err.Error())
        return
    }
    
    // 记录审计日志
    audit.Log(&model.AuditLog{
        Action:       "product.create",
        ActionGroup:  "product",
        ResourceType: "product",
        ResourceID:   product.ID,
        UserID:       getAdminID(c),
        RequestIP:    c.ClientIP(),
        NewValue:     map[string]interface{}{"name": product.Name, "price": product.Price},
        Success:      true,
    })
    
    response.Success(c, product)
}

// 上架商品 ⭐
func (h *AdminHandler) PublishProduct(c *gin.Context) {
    productID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    err := h.productService.Publish(productID, getAdminID(c))
    if err != nil {
        response.Fail(c, err.Error())
        return
    }
    
    response.Success(c, gin.H{"message": "商品已上架"})
}

// 下架商品 ⭐
func (h *AdminHandler) UnpublishProduct(c *gin.Context) {
    productID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    reason := c.Query("reason") // 下架原因
    
    err := h.productService.Unpublish(productID, reason, getAdminID(c))
    if err != nil {
        response.Fail(c, err.Error())
        return
    }
    
    response.Success(c, gin.H{"message": "商品已下架"})
}

// 删除商品
func (h *AdminHandler) DeleteProduct(c *gin.Context) {
    productID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    err := h.productService.Delete(productID)
    if err != nil {
        response.Fail(c, err.Error())
        return
    }
    
    response.Success(c, gin.H{"message": "商品已删除"})
}
```

### 4.3 管理员商品页面

```vue
<!-- views/admin/products/List.vue -->
<template>
  <div class="product-list">
    <!-- 工具栏 -->
    <el-card>
      <el-form inline>
        <el-form-item label="商品类型">
          <el-select v-model="query.product_type" clearable>
            <el-option label="充值套餐" value="recharge" />
            <el-option label="VIP套餐" value="vip" />
            <el-option label="组合套餐" value="package" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="已上架" value="active" />
            <el-option label="已下架" value="inactive" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="搜索">
          <el-input v-model="query.keyword" placeholder="商品名称" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
          <el-button type="success" @click="handleCreate">新建商品</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <!-- 商品列表 -->
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="商品名称" min-width="200">
        <template #default="{ row }">
          <div class="product-info">
            <img v-if="row.cover_image" :src="row.cover_image" class="cover" />
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="product_type" label="类型" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.product_type === 'recharge'" type="success">充值</el-tag>
          <el-tag v-else-if="row.product_type === 'vip'" type="warning">VIP</el-tag>
          <el-tag v-else type="info">套餐</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="price" label="价格" width="100">
        <template #default="{ row }">
          ¥{{ row.price }}
        </template>
      </el-table-column>
      <el-table-column prop="quota" label="配额" width="120">
        <template #default="{ row }">
          {{ formatNumber(row.quota) }} tokens
        </template>
      </el-table-column>
      <el-table-column prop="sold_count" label="已售" width="80" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'active'" type="success">已上架</el-tag>
          <el-tag v-else-if="row.status === 'inactive'" type="danger">已下架</el-tag>
          <el-tag v-else type="info">草稿</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status !== 'active'" link type="primary" @click="handlePublish(row)">
            上架
          </el-button>
          <el-button v-if="row.status === 'active'" link type="danger" @click="handleUnpublish(row)">
            下架
          </el-button>
          <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页 -->
    <el-pagination
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[20, 50, 100]"
      layout="total, sizes, prev, pager, next"
      @change="fetchList"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProductList, publishProduct, unpublishProduct, deleteProduct } from '@/api/admin'

const router = useRouter()
const list = ref([])
const loading = ref(false)
const total = ref(0)
const query = ref({
  page: 1,
  page_size: 20,
  product_type: '',
  status: '',
  keyword: ''
})

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getProductList(query.value)
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  router.push('/admin/products/form')
}

const handleEdit = (row) => {
  router.push(`/admin/products/form?id=${row.id}`)
}

const handlePublish = async (row) => {
  await publishProduct(row.id)
  ElMessage.success('商品已上架')
  fetchList()
}

const handleUnpublish = async (row) => {
  await ElMessageBox.confirm('确定要下架该商品吗?', '提示')
  const reason = '手动下架' // 可以添加下架原因输入框
  await unpublishProduct(row.id, reason)
  ElMessage.success('商品已下架')
  fetchList()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定要删除该商品吗?', '提示', { type: 'warning' })
  await deleteProduct(row.id)
  ElMessage.success('商品已删除')
  fetchList()
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.product-info {
  display: flex;
  align-items: center;
  gap: 10px;
}
.cover {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 4px;
}
</style>
```

---

## 5. 用户前台商品展示

### 5.1 用户商品API

```go
// 获取商品列表 (用户端)
func (h *UserHandler) GetProducts(c *gin.Context) {
    params := &ProductListParams{
        Page:        pagination.GetPage(c),
        PageSize:    pagination.GetPageSize(c),
        ProductType: c.Query("product_type"),
        CategoryID:  toInt64(c.Query("category_id")),
    }
    
    // 只返回已上架的商品
    params.Status = "active"
    
    products, total := h.productService.List(params)
    categories := h.productService.GetCategories()
    
    response.Success(c, gin.H{
        "total":     total,
        "list":      products,
        "categories": categories,
    })
}

// 获取商品详情
func (h *UserHandler) GetProductDetail(c *gin.Context) {
    productID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    product, err := h.productService.GetByID(productID)
    if err != nil || product.Status != "active" {
        response.Fail(c, "商品不存在")
        return
    }
    
    response.Success(c, product)
}
```

### 5.2 用户商品页面

```vue
<!-- views/user/Products.vue -->
<template>
  <div class="products-page">
    <!-- 分类导航 -->
    <div class="category-nav">
      <el-radio-group v-model="categoryId" @change="fetchList">
        <el-radio-button :value="0">全部</el-radio-button>
        <el-radio-button 
          v-for="cat in categories" 
          :key="cat.id" 
          :value="cat.id"
        >
          {{ cat.name }}
        </el-radio-button>
      </el-radio-group>
    </div>
    
    <!-- 商品类型 -->
    <div class="type-tabs">
      <el-tabs v-model="productType" @tab-change="fetchList">
        <el-tab-pane label="充值套餐" name="recharge" />
        <el-tab-pane label="VIP会员" name="vip" />
        <el-tab-pane label="热门推荐" name="hot" />
      </el-tabs>
    </div>
    
    <!-- 商品列表 -->
    <div class="product-grid">
      <div 
        v-for="product in list" 
        :key="product.id" 
        class="product-card"
        @click="goDetail(product)"
      >
        <!-- 标签 -->
        <div class="product-tags">
          <el-tag v-if="product.is_hot" type="danger" size="small">热门</el-tag>
          <el-tag v-if="product.is_recommended" type="success" size="small">推荐</el-tag>
          <el-tag v-if="product.discount" type="warning" size="small">
            {{ product.discount }}折
          </el-tag>
        </div>
        
        <!-- 封面 -->
        <div class="product-cover">
          <img v-if="product.cover_image" :src="product.cover_image" />
          <div v-else class="default-cover">
            <el-icon :size="40"><Coin /></el-icon>
          </div>
        </div>
        
        <!-- 信息 -->
        <div class="product-info">
          <h3 class="name">{{ product.name }}</h3>
          <p class="description">{{ product.description }}</p>
          
          <!-- 配额信息 -->
          <div class="quota-info">
            <span class="quota">{{ formatNumber(product.quota) }}</span>
            <span class="unit">tokens</span>
            <span v-if="product.bonus_quota" class="bonus">
              + {{ formatNumber(product.bonus_quota) }} 赠送
            </span>
          </div>
          
          <!-- VIP信息 -->
          <div v-if="product.vip_days" class="vip-info">
            <el-icon><Star /></el-icon>
            <span>{{ product.vip_days }}天VIP</span>
            <span v-if="product.vip_quota"> + {{ formatNumber(product.vip_quota) }} 配额</span>
          </div>
          
          <!-- 价格 -->
          <div class="price-row">
            <span class="price">¥{{ product.price }}</span>
            <span v-if="product.original_price" class="original">
              ¥{{ product.original_price }}
            </span>
          </div>
        </div>
        
        <!-- 购买按钮 -->
        <div class="action-row">
          <el-button type="primary" @click.stop="handleBuy(product)">
            立即购买
          </el-button>
        </div>
      </div>
    </div>
    
    <!-- 空状态 -->
    <el-empty v-if="!loading && list.length === 0" description="暂无商品" />
    
    <!-- 分页 -->
    <div v-if="total > 0" class="pagination">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        :page-sizes="[12, 24, 48]"
        layout="total, sizes, prev, pager, next"
        @change="fetchList"
      />
    </div>
    
    <!-- 购买确认对话框 -->
    <el-dialog v-model="buyDialogVisible" title="确认订单" width="400px">
      <div class="order-summary">
        <img :src="selectedProduct?.cover_image" class="order-cover" />
        <div class="order-info">
          <h4>{{ selectedProduct?.name }}</h4>
          <p>{{ selectedProduct?.description }}</p>
        </div>
      </div>
      
      <el-divider />
      
      <div class="order-price">
        <span>应付金额:</span>
        <span class="price">¥{{ selectedProduct?.price }}</span>
      </div>
      
      <template #footer>
        <el-button @click="buyDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="buying" @click="confirmBuy">
          确认购买
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getUserProductList, createOrder } from '@/api/user'

const router = useRouter()
const list = ref([])
const categories = ref([])
const loading = ref(false)
const total = ref(0)
const productType = ref('recharge')
const categoryId = ref(0)

const query = ref({
  page: 1,
  page_size: 12
})

const buyDialogVisible = ref(false)
const selectedProduct = ref(null)
const buying = ref(false)

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getUserProductList({
      ...query.value,
      product_type: productType.value,
      category_id: categoryId.value || undefined
    })
    list.value = res.list
    categories.value = res.categories
    total.value = res.total
  } finally {
    loading.value = false
  }
}

const goDetail = (product) => {
  router.push(`/user/products/${product.id}`)
}

const handleBuy = (product) => {
  selectedProduct.value = product
  buyDialogVisible.value = true
}

const confirmBuy = async () => {
  buying.value = true
  try {
    const order = await createOrder({
      product_id: selectedProduct.value.id,
      payment_method: 'alipay' // 可以让用户选择
    })
    
    // 跳转支付
    router.push(`/user/pay/${order.id}`)
  } finally {
    buying.value = false
  }
}

const formatNumber = (num) => {
  return new Intl.NumberFormat('zh-CN').format(num)
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.products-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100vh;
}

.category-nav {
  margin-bottom: 20px;
}

.type-tabs {
  margin-bottom: 20px;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.product-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.product-tags {
  position: absolute;
  top: 10px;
  left: 10px;
  display: flex;
  gap: 5px;
  z-index: 1;
}

.product-cover {
  height: 160px;
  background: #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.default-cover {
  color: #c0c4cc;
}

.product-info {
  padding: 15px;
}

.name {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.description {
  margin: 0 0 10px;
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.quota-info {
  margin-bottom: 8px;
}

.quota {
  font-size: 20px;
  font-weight: 700;
  color: #409eff;
}

.unit {
  font-size: 12px;
  color: #909399;
}

.bonus {
  font-size: 12px;
  color: #67c23a;
  margin-left: 5px;
}

.vip-info {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #e6a23c;
  margin-bottom: 8px;
}

.price-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.price {
  font-size: 24px;
  font-weight: 700;
  color: #f56c6c;
}

.original {
  font-size: 14px;
  color: #909399;
  text-decoration: line-through;
}

.action-row {
  padding: 0 15px 15px;
}

.action-row .el-button {
  width: 100%;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.order-summary {
  display: flex;
  gap: 15px;
}

.order-cover {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
}

.order-info h4 {
  margin: 0 0 5px;
}

.order-info p {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.order-price {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
}

.order-price .price {
  font-size: 24px;
  color: #f56c6c;
}

@media (max-width: 1200px) {
  .product-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
```

---

## 6. 订单与支付流程

### 6.1 订单创建流程

```go
// 创建订单
func (h *UserHandler) CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, "参数错误")
        return
    }
    
    userID := getUserID(c)
    
    // 获取商品
    product, err := h.productService.GetByID(req.ProductID)
    if err != nil || product.Status != "active" {
        response.Fail(c, "商品不存在或已下架")
        return
    }
    
    // 检查限购
    if product.MaxPerUser > 0 {
        count := h.orderService.GetUserOrderCount(userID, product.ID)
        if count >= product.MaxPerUser {
            response.Fail(c, "该商品每人限购"+strconv.Itoa(product.MaxPerUser)+"件")
            return
        }
    }
    
    // 生成订单号
    orderNo := generateOrderNo()
    
    // 创建订单
    order := &model.Order{
        OrderNo:     orderNo,
        UserID:      userID,
        ProductID:   product.ID,
        ProductName: product.Name,
        ProductType: product.ProductType,
        TotalAmount: product.Price,
        PayAmount:   product.Price,
        Status:      "pending",
        ExpireAt:    time.Now().Add(30 * time.Minute), // 30分钟过期
    }
    
    if err := h.orderService.Create(order); err != nil {
        response.Fail(c, "创建订单失败")
        return
    }
    
    response.Success(c, order)
}

// 支付订单
func (h *UserHandler) PayOrder(c *gin.Context) {
    orderID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    paymentMethod := c.Query("method") // alipay | wechat
    
    userID := getUserID(c)
    
    // 获取订单
    order, err := h.orderService.GetByID(orderID)
    if err != nil || order.UserID != userID {
        response.Fail(c, "订单不存在")
        return
    }
    
    if order.Status != "pending" {
        response.Fail(c, "订单状态不正确")
        return
    }
    
    // 检查订单过期
    if order.ExpireAt.Before(time.Now()) {
        response.Fail(c, "订单已过期")
        return
    }
    
    // 创建支付
    payment, err := h.paymentService.CreatePayment(order, paymentMethod)
    if err != nil {
        response.Fail(c, "创建支付失败")
        return
    }
    
    response.Success(c, gin.H{
        "payment_id": payment.ID,
        "payment_url": payment.PaymentURL,
        "qr_code":     payment.QRCode,
    })
}
```

---

## 7. 邮箱验证

### 7.1 验证流程

```sql
-- 邮箱验证记录表
CREATE TABLE email_verifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    email           VARCHAR(100) NOT NULL,
    token           VARCHAR(100) NOT NULL UNIQUE,
    
    -- 状态
    status          VARCHAR(20) DEFAULT 'pending',         -- pending|verified|expired
    verified_at     TIMESTAMPTZ,
    
    -- 过期时间
    expired_at      TIMESTAMPTZ NOT NULL,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_token ON email_verifications(token);
CREATE INDEX idx_verification_user ON email_verifications(user_id);
```

### 7.2 验证API

```
GET /api/v1/user/auth/verify-email
Query: token=xxx

响应: 
- 成功: {code: 0, message: "验证成功"}
- 失败: {code: 40001, message: "验证链接无效或已过期"}
```

---

**文档版本**: 1.0  
**下一步**: 可以开始实现了