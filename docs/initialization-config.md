# gAPI Platform - 项目初始化与配置设计文档

**版本**: 1.0  
**日期**: 2026-03-23  
**状态**: 待实现

---

## 1. 初始化系统设计

### 1.1 初始化流程

参考 OneAPI/NewAPI 的开箱即用设计，我们的初始化流程如下：

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           首次启动初始化流程                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. 启动服务                                                               │
│     ├── 检查配置文件是否存在                                               │
│     ├── 检查数据库连接                                                     │
│     └── 如果未初始化 → 进入引导页                                          │
│                          │                                                 │
│                          ▼                                                 │
│  2. 引导页面 (Web界面)                                                     │
│     ├── 选择数据库类型 (SQLite/MySQL/PostgreSQL)                           │
│     ├── 配置数据库连接                                                     │
│     ├── 配置Redis (可选)                                                   │
│     ├── 设置管理员账号                                                     │
│     └── 初始化系统                                                         │
│                          │                                                 │
│                          ▼                                                 │
│  3. 初始化完成                                                             │
│     ├── 生成配置文件                                                       │
│     ├── 创建数据库表                                                       │
│     ├── 创建默认管理员                                                     │
│     ├── 初始化默认配置                                                     │
│     └── 跳转到登录页                                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 初始化API

```go
// 初始化状态检查
GET /api/v1/init/status

响应:
{
    "code": 0,
    "data": {
        "initialized": true,        // 是否已初始化
        "database_type": "postgres", // 数据库类型
        "redis_configured": true    // Redis是否配置
    }
}

// 开始初始化
POST /api/v1/init/setup
Content-Type: application/json

请求:
{
    "database": {
        "type": "postgres",         // sqlite | mysql | postgres
        "host": "localhost",
        "port": 5432,
        "username": "postgres",
        "password": "xxx",          // 首次设置
        "database": "gapi_db",
        // SQLite无需其他字段
    },
    "redis": {
        "enabled": true,
        "host": "localhost",
        "port": 6379,
        "password": "xxx",          // 首次设置
        "db": 0
    },
    "admin": {
        "username": "admin",
        "password": "xxx",          // 首次设置
        "confirm_password": "xxx"
    },
    "system": {
        "site_name": "gAPI Platform",
        "timezone": "Asia/Shanghai"
    }
}

响应:
{
    "code": 0,
    "message": "初始化完成",
    "data": {
        "admin_id": 1,
        "redirect": "/admin/login"
    }
}
```

### 1.3 环境变量配置

参考 OneAPI，支持通过环境变量或配置文件进行预配置：

```yaml
# 配置文件 config.yaml
server:
  addr: ":8080"
  mode: "debug"

# 数据库配置
database:
  # 支持: sqlite, mysql, postgres
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "${DB_PASSWORD}"        # 从环境变量读取
  database: "gapi_db"
  
# Redis配置 (可选)
redis:
  enabled: true
  host: "localhost"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0

# JWT配置
jwt:
  secret: "${JWT_SECRET}"
  expire_hour: 24

# 日志配置
log:
  level: "info"                    # debug | info | warn | error
  format: "json"                   # text | json
  output: "file"                   # stdout | file
  path: "/var/log/gapi"           # 日志目录 (可配置)
  max_size: 100                    # 单文件最大MB
  max_backups: 30                  # 保留文件数
  max_age: 90                      # 保留天数
  compress: true                   # 压缩
  
# 支付配置
payment:
  alipay:
    enabled: false
    app_id: "${ALIPAY_APP_ID}"
    private_key: "${ALIPAY_PRIVATE_KEY}"
  wechat:
    enabled: false
    app_id: "${WECHAT_APP_ID}"
    mch_id: "${WECHAT_MCH_ID}"

# 系统配置
system:
  site_name: "gAPI Platform"
  timezone: "Asia/Shanghai"
  allow_register: true
  signup_quota: 100000            # 注册赠送配额
```

```bash
# 环境变量示例
export DB_PASSWORD="安全的数据库密码"
export REDIS_PASSWORD="安全的Redis密码"
export JWT_SECRET="安全的JWT密钥"
export ALIPAY_APP_ID=""
export ALIPAY_PRIVATE_KEY=""
export WECHAT_APP_ID=""
export WECHAT_MCH_ID=""
export WECHAT_API_KEY=""
```

---

## 2. 日志系统设计

### 2.1 日志级别

| 级别 | 值 | 说明 | 颜色 |
|------|-----|------|------|
| DEBUG | 0 | 调试信息 | 灰色 |
| INFO | 1 | 正常信息 | 蓝色 |
| WARN | 2 | 警告信息 | 黄色 |
| ERROR | 3 | 错误信息 | 红色 |
| FATAL | 4 | 致命错误 | 深红 |

### 2.2 日志脱敏

**原则**: 密码、密钥、Token等敏感信息不能出现在日志中

```go
// 日志脱敏处理
package logger

import (
    "regexp"
    "strings"
)

var sensitivePatterns = map[string]string{
    "password":    `"password"\s*:\s*"[^"]*"`,
    "password":    `"password"\s*:\s*[^,}]*`,
    "api_key":     `"api_key"\s*:\s*"[^"]*"`,
    "token":       `"token"\s*:\s*"[^"]*"`,
    "secret":      `"secret"\s*:\s*"[^"]*"`,
    "private_key": `"private_key"\s*:\s*"[^"]*"`,
    "credit_card": `"credit_card"\s*:\s*"[^"]*"`,
    "auth_header": `Authorization\s*:\s*Bearer[^"]*`,
    "x-api-key":   `X-API-Key\s*:\s*[^"]*`,
}

// MaskSensitiveData 脱敏处理
func MaskSensitiveData(data string) string {
    result := data
    
    // 使用正则替换
    for key, pattern := range sensitivePatterns {
        re := regexp.MustCompile(pattern)
        result = re.ReplaceAllString(result, fmt.Sprintf(`"%s":"***"`, key))
    }
    
    // 特殊处理 Bearer Token
    result = strings.ReplaceAll(result, "Bearer sk-", "Bearer sk-***")
    result = strings.ReplaceAll(result, "Bearer eyJ", "Bearer eyJ***")
    
    return result
}

// 安全日志记录
func (l *Logger) SafeLog(level string, message string, data map[string]interface{}) {
    sanitized := make(map[string]interface{})
    
    for k, v := range data {
        // 检查是否是敏感字段
        if isSensitiveField(k) {
            sanitized[k] = "***"
        } else if str, ok := v.(string); ok {
            sanitized[k] = MaskSensitiveData(str)
        } else {
            sanitized[k] = v
        }
    }
    
    l.Log(level, message, sanitized)
}

func isSensitiveField(field string) bool {
    sensitiveFields := []string{
        "password", "password_hash", "api_key", "token",
        "token_key", "secret", "private_key", "access_token",
        "refresh_token", "credit_card", "bank_account",
    }
    
    for _, s := range sensitiveFields {
        if strings.Contains(strings.ToLower(field), s) {
            return true
        }
    }
    return false
}
```

### 2.3 日志记录动作

虽然日志会脱敏，但为了安全审计，我们会把**重要操作记录到数据库**：

```sql
-- 操作日志表
CREATE TABLE operation_logs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT,
    user_id         BIGINT,
    
    -- 日志信息
    level           VARCHAR(10) NOT NULL,               -- debug|info|warn|error
    category        VARCHAR(50) NOT NULL,              -- auth|channel|order|payment|system
    action          VARCHAR(100) NOT NULL,             # 具体操作
    message         TEXT NOT NULL,                      # 日志消息
    
    -- 上下文 (脱敏后)
    ip              VARCHAR(50),
    user_agent      VARCHAR(500),
    request_method  VARCHAR(10),
    request_path    VARCHAR(200),
    
    -- 关联资源
    resource_type   VARCHAR(50),
    resource_id     BIGINT,
    
    -- 附加数据 (JSON, 脱敏)
    extra           JSONB,
    
    -- 时间
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_op_logs_level ON operation_logs(level);
CREATE INDEX idx_op_logs_category ON operation_logs(category);
CREATE INDEX idx_op_logs_created ON operation_logs(created_at DESC);
CREATE INDEX idx_op_logs_user ON operation_logs(user_id, created_at DESC);

-- 按日期自动清理 (保留30天)
CREATE INDEX idx_op_logs_date ON operation_logs(created_at DESC) 
    WHERE created_at > NOW() - INTERVAL '30 days';
```

### 2.4 日志写入服务

```go
// 日志服务
type LogService struct {
    db *pg.DB
}

func (s *LogService) Log(ctx context.Context, level, category, action, message string) {
    // 1. 写入数据库 (异步)
    go s.saveToDatabase(ctx, level, category, action, message)
    
    // 2. 写入文件 (同步)
    s.saveToFile(level, message)
}

func (s *LogService) saveToDatabase(ctx context.Context, level, category, action, message string) {
    userID, _ := ctx.Get("user_id")
    
    log := &model.OperationLog{
        Level:      level,
        Category:   category,
        Action:     action,
        Message:    message,
        UserID:    toInt64(userID),
        IP:        ctx.ClientIP(),
        CreatedAt: time.Now(),
    }
    
    s.db.Insert(log)
}

// 日志统计 (用于管理页面)
type LogStats struct {
    Total       int64   `json:"total"`
    DebugCount  int64   `json:"debug_count"`
    InfoCount   int64   `json:"info_count"`
    WarnCount   int64   `json:"warn_count"`
    ErrorCount  int64   `json:"error_count"`
    FatalCount  int64   `json:"fatal_count"`
    
    DebugRatio  float64 `json:"debug_ratio"`
    InfoRatio   float64 `json:"info_ratio"`
    WarnRatio   float64 `json:"warn_ratio"`
    ErrorRatio  float64 `json:"error_ratio"`
    FatalRatio  float64 `json:"fatal_ratio"`
    
    TodayCount  int64   `json:"today_count"`
    WeekCount   int64   `json:"week_count"`
    MonthCount  int64   `json:"month_count"`
}

func (s *LogService) GetStats(days int) (*LogStats, error) {
    stats := &LogStats{}
    
    // 获取各级别数量
    query := `
        SELECT 
            COUNT(*) as total,
            SUM(CASE WHEN level = 'debug' THEN 1 ELSE 0 END) as debug_count,
            SUM(CASE WHEN level = 'info' THEN 1 ELSE 0 END) as info_count,
            SUM(CASE WHEN level = 'warn' THEN 1 ELSE 0 END) as warn_count,
            SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) as error_count,
            SUM(CASE WHEN level = 'fatal' THEN 1 ELSE 0 END) as fatal_count
        FROM operation_logs
        WHERE created_at > NOW() - INTERVAL '? days'
    `
    
    _, err := s.db.Query(query, days).Scan(
        &stats.Total, &stats.DebugCount, &stats.InfoCount,
        &stats.WarnCount, &stats.ErrorCount, &stats.FatalCount,
    )
    
    // 计算比例
    if stats.Total > 0 {
        stats.DebugRatio = float64(stats.DebugCount) / float64(stats.Total) * 100
        stats.InfoRatio = float64(stats.InfoCount) / float64(stats.Total) * 100
        stats.WarnRatio = float64(stats.WarnCount) / float64(stats.Total) * 100
        stats.ErrorRatio = float64(stats.ErrorCount) / float64(stats.Total) * 100
        stats.FatalRatio = float64(stats.FatalCount) / float64(stats.Total) * 100
    }
    
    return stats, err
}
```

---

## 3. 管理后台日志查看页面

### 3.1 日志统计API

```go
// 获取日志统计
GET /api/v1/admin/logs/statistics

响应:
{
    "code": 0,
    "data": {
        "total": 10000,
        "debug": { "count": 1000, "ratio": 10.0 },
        "info": { "count": 7000, "ratio": 70.0 },
        "warn": { "count": 1500, "ratio": 15.0 },
        "error": { "count": 490, "ratio": 4.9 },
        "fatal": { "count": 10, "ratio": 0.1 },
        "today": 500,
        "trend": [
            {"date": "2026-03-17", "error": 20, "warn": 50},
            {"date": "2026-03-18", "error": 15, "warn": 45},
            {"date": "2026-03-19", "error": 25, "warn": 60},
            ...
        ]
    }
}

// 获取日志列表
GET /api/v1/admin/logs

查询参数:
- level: debug|info|warn|error
- category: auth|channel|order|payment|system
- start_time: 开始时间
- end_time: 结束时间
- page: 页码
- page_size: 每页数量
```

### 3.2 日志查看页面Vue

```vue
<!-- views/admin/logs/Index.vue -->
<template>
  <div class="logs-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ formatNumber(stats.total) }}</div>
          <div class="stat-label">总日志数</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card info">
          <div class="stat-value">{{ stats.info_count }}</div>
          <div class="stat-label">INFO</div>
          <div class="stat-ratio">{{ stats.info_ratio.toFixed(1) }}%</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card warning">
          <div class="stat-value">{{ stats.warn_count }}</div>
          <div class="stat-label">WARN</div>
          <div class="stat-ratio">{{ stats.warn_ratio.toFixed(1) }}%</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card danger">
          <div class="stat-value">{{ stats.error_count }}</div>
          <div class="stat-label">ERROR</div>
          <div class="stat-ratio">{{ stats.error_ratio.toFixed(1) }}%</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ stats.today_count }}</div>
          <div class="stat-label">今日</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ stats.week_count }}</div>
          <div class="stat-label">本周</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <span>日志趋势</span>
      </template>
      <div class="chart-container">
        <v-chart :option="chartOption" autoresize />
      </div>
    </el-card>

    <!-- 日志列表 -->
    <el-card class="list-card">
      <el-form inline>
        <el-form-item label="级别">
          <el-select v-model="query.level" clearable>
            <el-option label="DEBUG" value="debug" />
            <el-option label="INFO" value="info" />
            <el-option label="WARN" value="warn" />
            <el-option label="ERROR" value="error" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="分类">
          <el-select v-model="query.category" clearable>
            <el-option label="认证" value="auth" />
            <el-option label="渠道" value="channel" />
            <el-option label="订单" value="order" />
            <el-option label="支付" value="payment" />
            <el-option label="系统" value="system" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="时间">
          <el-date-picker
            v-model="query.time_range"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="fetchList">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" v-loading="loading">
        <el-table-column prop="level" label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="getLevelType(row.level)" size="small">
              {{ row.level.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="category" label="分类" width="100">
          <template #default="{ row }">
            {{ getCategoryLabel(row.category) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="action" label="操作" width="150" />
        
        <el-table-column prop="message" label="消息" min-width="300">
          <template #default="{ row }">
            <el-tooltip :content="row.message" placement="top">
              <span class="message-text">{{ row.message }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        
        <el-table-column prop="ip" label="IP" width="130" />
        
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @change="fetchList"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="日志详情" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="级别">
          <el-tag :type="getLevelType(detail.level)">
            {{ detail.level }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="分类">
          {{ getCategoryLabel(detail.category) }}
        </el-descriptions-item>
        <el-descriptions-item label="操作">
          {{ detail.action }}
        </el-descriptions-item>
        <el-descriptions-item label="时间">
          {{ formatTime(detail.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="用户ID" v-if="detail.user_id">
          {{ detail.user_id }}
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">
          {{ detail.ip }}
        </el-descriptions-item>
        <el-descriptions-item label="请求方法">
          {{ detail.request_method }}
        </el-descriptions-item>
        <el-descriptions-item label="请求路径">
          {{ detail.request_path }}
        </el-descriptions-item>
        <el-descriptions-item label="消息" :span="2">
          {{ detail.message }}
        </el-descriptions-item>
        <el-descriptions-item label="附加数据" :span="2" v-if="detail.extra">
          <pre>{{ detail.extra }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { getLogStatistics, getLogList } from '@/api/admin'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

const stats = ref({})
const list = ref([])
const loading = ref(false)
const total = ref(0)
const detail = ref({})
const detailVisible = ref(false)

const query = ref({
  page: 1,
  page_size: 20,
  level: '',
  category: '',
  time_range: []
})

const fetchStats = async () => {
  const res = await getLogStatistics()
  stats.value = res
}

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getLogList(query.value)
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

const showDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

const getLevelType = (level) => {
  const map = { debug: 'info', info: '', warn: 'warning', error: 'danger', fatal: 'danger' }
  return map[level] || 'info'
}

const getCategoryLabel = (cat) => {
  const map = { auth: '认证', channel: '渠道', order: '订单', payment: '支付', system: '系统' }
  return map[cat] || cat
}

const formatTime = (t) => new Date(t).toLocaleString()
const formatNumber = (n) => new Intl.NumberFormat().format(n)

const chartOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['ERROR', 'WARN'] },
  xAxis: { type: 'category', data: stats.value.trend?.map(t => t.date) || [] },
  yAxis: { type: 'value' },
  series: [
    { name: 'ERROR', type: 'line', data: stats.value.trend?.map(t => t.error) || [] },
    { name: 'WARN', type: 'line', data: stats.value.trend?.map(t => t.warn) || [] }
  ]
}))

onMounted(() => {
  fetchStats()
  fetchList()
})
</script>

<style scoped>
.stats-row {
  margin-bottom: 20px;
}
.stat-card {
  text-align: center;
}
.stat-value {
  font-size: 28px;
  font-weight: bold;
}
.stat-label {
  font-size: 12px;
  color: #909399;
}
.stat-ratio {
  font-size: 12px;
  color: #67c23a;
}
.stat-card.danger .stat-value { color: #f56c6c; }
.stat-card.warning .stat-value { color: #e6a23c; }
.chart-container { height: 300px; }
.message-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
}
</style>
```

---

## 4. 参考 OneAPI/NewAPI 的优点

### 4.1 单可执行文件部署

```makefile
# 构建单一可执行文件
build:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gapi-server .
    
# Docker 构建
docker:
    docker build -t gapi-platform:latest .
```

### 4.2 默认配置 (开箱即用)

```go
// 默认配置
var DefaultConfig = &Config{
    Server: ServerConfig{
        Addr:   ":8080",
        Mode:   "debug",
    },
    Database: DatabaseConfig{
        Type: "sqlite",
        Path: "./data/gapi.db",
    },
    Log: LogConfig{
        Level:  "info",
        Format: "text",
        Output: "stdout",
    },
    System: SystemConfig{
        SiteName:     "gAPI Platform",
        Timezone:     "Asia/Shanghai",
        AllowRegister: true,
        SignupQuota:  100000,
    },
}
```

### 4.3 自动数据迁移

```go
// 使用 golang-migrate 进行数据库迁移
// or 自定义简易迁移

func autoMigrate(db *pg.DB) error {
    // 检查表是否存在，不存在则创建
    tables := []string{
        "tenants", "users", "admin_users", "channels",
        "abilities", "tokens", "products", "orders",
        "payments", "quota_transactions", "audit_logs",
        "login_logs", "usage_logs", "operation_logs",
    }
    
    for _, table := range tables {
        if !tableExists(db, table) {
            if err := createTable(db, table); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

---

## 5. 完整配置清单

| 配置项 | 环境变量 | 默认值 | 说明 |
|-------|---------|-------|------|
| **数据库** | | | |
| DATABASE_TYPE | DATABASE_TYPE | sqlite | sqlite/mysql/postgres |
| DATABASE_HOST | DB_HOST | localhost | 数据库主机 |
| DATABASE_PORT | DB_PORT | 5432 | 数据库端口 |
| DATABASE_USER | DB_USER | postgres | 数据库用户 |
| DATABASE_PASSWORD | DB_PASSWORD | - | 数据库密码 |
| DATABASE_NAME | DB_NAME | gapi_db | 数据库名 |
| DATABASE_PATH | DB_PATH | ./data/gapi.db | SQLite路径 |
| **Redis** | | | |
| REDIS_ENABLED | REDIS_ENABLED | false | 是否启用Redis |
| REDIS_HOST | REDIS_HOST | localhost | Redis主机 |
| REDIS_PORT | REDIS_PORT | 6379 | Redis端口 |
| REDIS_PASSWORD | REDIS_PASSWORD | - | Redis密码 |
| **JWT** | | | |
| JWT_SECRET | JWT_SECRET | auto-generated | JWT密钥 |
| JWT_EXPIRE_HOUR | JWT_EXPIRE_HOUR | 24 | Token有效期 |
| **日志** | | | |
| LOG_LEVEL | LOG_LEVEL | info | 日志级别 |
| LOG_FORMAT | LOG_FORMAT | text | 日志格式 |
| LOG_OUTPUT | LOG_OUTPUT | stdout | 输出方式 |
| LOG_PATH | LOG_PATH | /var/log/gapi | 日志目录 |
| **系统** | | | |
| SITE_NAME | SITE_NAME | gAPI Platform | 站点名称 |
| TIMEZONE | TIMEZONE | Asia/Shanghai | 时区 |
| ALLOW_REGISTER | ALLOW_REGISTER | true | 是否开放注册 |
| SIGNUP_QUOTA | SIGNUP_QUOTA | 100000 | 注册赠送配额 |
| **支付** | | | |
| ALIPAY_ENABLED | ALIPAY_ENABLED | false | 启用支付宝 |
| WECHAT_ENABLED | WECHAT_ENABLED | false | 启用微信支付 |

---

**文档版本**: 1.0  
**下一步**: 项目代码实现