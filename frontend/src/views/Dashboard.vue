<template>
  <div class="dashboard">
    <!-- Stats Cards -->
    <div class="stats-grid">
      <el-card shadow="hover" class="stat-card">
        <div class="stat-icon blue">
          <el-icon size="24"><Coin /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">剩余配额</div>
          <div class="stat-value">{{ formatQuota(quota?.remain_quota) }}</div>
        </div>
      </el-card>
      
      <el-card shadow="hover" class="stat-card">
        <div class="stat-icon green">
          <el-icon size="24"><TrendCharts /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">今日用量</div>
          <div class="stat-value">{{ formatQuota(usageStats?.used_today) }}</div>
        </div>
      </el-card>
      
      <el-card shadow="hover" class="stat-card">
        <div class="stat-icon orange">
          <el-icon size="24"><Star /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">会员状态</div>
          <div class="stat-value">{{ quota?.is_vip ? 'VIP会员' : '免费用户' }}</div>
        </div>
      </el-card>
      
      <el-card shadow="hover" class="stat-card">
        <div class="stat-icon red">
          <el-icon size="24"><Key /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">API密钥</div>
          <div class="stat-value">{{ tokenCount }} 个</div>
        </div>
      </el-card>
    </div>

    <!-- Charts Section -->
    <div class="charts-grid">
      <el-card shadow="hover" class="chart-card">
        <template #header>
          <span>Token消耗趋势 (近7天)</span>
        </template>
        <div class="chart-container">
          <v-chart :option="tokenChartOption" autoresize />
        </div>
      </el-card>
      
      <el-card shadow="hover" class="chart-card">
        <template #header>
          <span>API调用统计 (近7天)</span>
        </template>
        <div class="chart-container">
          <v-chart :option="callsChartOption" autoresize />
        </div>
      </el-card>
    </div>

    <!-- Main Content -->
    <div class="main-grid">
      <!-- Quick Start -->
      <el-card class="quickstart-card">
        <template #header>
          <div class="card-header">
            <span>快速开始</span>
            <el-button text size="small" @click="copyCode">
              <el-icon><CopyDocument /></el-icon> 复制代码
            </el-button>
          </div>
        </template>
        
        <p class="intro-text">使用以下方式调用 API：</p>
        
        <div class="code-box">
          <div class="code-header">
            <span class="lang-badge">bash</span>
            <span class="title">调用示例</span>
          </div>
          <pre class="code-content"><code><span class="hljs-comment"># 使用 cURL 调用 API</span>
curl http://localhost:8080/api/v1/chat/completions \
  <span class="hljs-operator">-H</span> <span class="hljs-string">"Authorization: Bearer sk-ap-your-key"</span> \
  <span class="hljs-operator">-H</span> <span class="hljs-string">"Content-Type: application/json"</span> \
  <span class="hljs-operator">-d</span> <span class="hljs-string">'{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello"}]
  }'</span></code></pre>
        </div>
        
        <el-divider />
        
        <div class="model-list">
          <span class="list-label">支持的模型：</span>
          <el-tag v-for="m in supportedModels" :key="m" type="info" size="small" class="model-tag">
            {{ m }}
          </el-tag>
        </div>
      </el-card>

      <!-- Quota Details -->
      <el-card class="quota-card">
        <template #header>
          <span>配额详情</span>
        </template>
        
        <div class="quota-item">
          <span class="item-label">账户等级</span>
          <el-tag :type="quota?.is_vip ? 'warning' : 'info'" size="small">
            {{ quota?.level || 'free' }}
          </el-tag>
        </div>
        
        <el-divider style="margin: 12px 0" />
        
        <div class="quota-item">
          <span class="item-label">永久配额</span>
          <span class="item-value">{{ formatQuota(quota?.remain_quota) }}</span>
        </div>
        
        <div class="quota-item">
          <span class="item-label">VIP 配额</span>
          <span class="item-value vip">{{ formatQuota(quota?.vip_quota) }}</span>
        </div>
        
        <div class="quota-item">
          <span class="item-label">累计Token</span>
          <span class="item-value">{{ formatQuota(usageStats?.total_tokens) }}</span>
        </div>
        
        <el-divider style="margin: 12px 0" />
        
        <div class="actions">
          <el-button type="primary" size="small" @click="$router.push('/products')">
            <el-icon><ShoppingCart /></el-icon> 购买配额
          </el-button>
          <el-button size="small" @click="$router.push('/vip')">
            <el-icon><Star /></el-icon> 开通VIP
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- Recent Activity -->
    <el-card class="activity-card">
      <template #header>
        <div class="card-header">
          <span>最近活动</span>
          <el-button text size="small" @click="$router.push('/logs')">
            查看全部 <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </template>
      
      <div class="activity-list">
        <div class="activity-item" v-for="item in recentActivity" :key="item.id">
          <div class="activity-icon" :class="item.type">
            <el-icon>
              <Key v-if="item.type === 'token'" />
              <ShoppingCart v-else-if="item.type === 'order'" />
              <Star v-else />
            </el-icon>
          </div>
          <div class="activity-info">
            <div class="activity-title">{{ item.title }}</div>
            <div class="activity-desc">{{ item.description }}</div>
          </div>
          <div class="activity-time">{{ formatTime(item.time) }}</div>
        </div>
        
        <div class="empty-state" v-if="!recentActivity.length">
          <el-empty description="暂无活动记录" :image-size="80" />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { 
  Coin, TrendCharts, Star, Key, CopyDocument, 
  ShoppingCart, ArrowRight 
} from '@element-plus/icons-vue'
import request from '@/api/request'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent])

interface Quota {
  remain_quota: number
  used_quota_today: number
  used_quota_month: number
  vip_quota: number
  is_vip: boolean
  level: string
}

interface DailyUsage {
  date: string
  total_calls: number
  total_tokens: number
  avg_response_ms: number
}

interface UsageStats {
  daily_usage: DailyUsage[]
  total_tokens_all: number
  total_calls_all: number
  used_today: number
}

const quota = ref<Quota | null>(null)
const tokenCount = ref(0)
const usageStats = ref<UsageStats | null>(null)
const dailyUsage = ref<DailyUsage[]>([])

const supportedModels = [
  'GPT-3.5-Turbo', 'GPT-4', 'GPT-4-Turbo', 
  'Claude-3-Opus', 'Claude-3-Sonnet'
]

const recentActivity = ref<Array<{
  id: number
  type: 'token' | 'order' | 'vip'
  title: string
  description: string
  time: Date
}>>([])

function formatQuota(n: number | undefined): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

function formatTime(date: Date): string {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  return `${days}天前`
}

const tokenChartOption = computed(() => {
  const data = dailyUsage.value.map(d => d.total_tokens)
  const maxValue = Math.max(...data, 1000)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: any[]) => {
        const p = params[0]
        const val = p.value >= 1000 ? (p.value / 1000).toFixed(2) + 'k' : p.value.toLocaleString()
        return `${p.axisValue}<br/>${p.marker} Token消耗: ${val} k`
      }
    },
    grid: {
      left: 50,
      right: 20,
      bottom: 25,
      top: 30,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dailyUsage.value.map(d => d.date),
      boundaryGap: false
    },
    yAxis: {
      type: 'value',
      name: 'Token(k)',
      nameLocation: 'middle',
      nameGap: 35,
      nameTextStyle: {
        align: 'center',
        verticalAlign: 'bottom'
      },
      min: 0,
      max: Math.ceil(maxValue / 5) * 5 + 1000,
      splitNumber: 5,
      axisLabel: {
        formatter: (v: number) => v >= 1000 ? (v / 1000).toFixed(1) + 'k' : v
      }
    },
    series: [{
      name: 'Token消耗',
      type: 'line',
      data: data,
      smooth: true,
      itemStyle: { color: '#409eff' },
      areaStyle: { color: 'rgba(64, 158, 255, 0.1)' },
      lineStyle: { width: 3 },
      symbol: 'circle',
      symbolSize: 8
    }]
  }
})

const callsChartOption = computed(() => {
  const data = dailyUsage.value.map(d => d.total_calls)
  const maxValue = Math.max(...data, 30)
  
  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: any[]) => {
        const p = params[0]
        return `${p.axisValue}<br/>${p.marker} API调用: ${p.value.toLocaleString()} 次`
      }
    },
    grid: {
      left: 50,
      right: 20,
      bottom: 25,
      top: 30,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dailyUsage.value.map(d => d.date)
    },
    yAxis: {
      type: 'value',
      name: '调用次数',
      nameLocation: 'middle',
      nameGap: 30,
      nameTextStyle: {
        align: 'center',
        verticalAlign: 'bottom'
      },
      min: 0,
      max: Math.ceil(maxValue / 5) * 5 + 5,
      splitNumber: 5,
      axisLabel: {
        formatter: (v: number) => v >= 1000 ? (v / 1000).toFixed(1) + 'k' : v
      }
    },
    series: [{
      name: 'API调用',
      type: 'bar',
      data: data,
      itemStyle: { color: '#67c23a', borderRadius: [4, 4, 0, 0] },
      barMaxWidth: 40
    }]
  }
})

async function copyCode() {
  const code = `curl http://localhost:8080/api/v1/chat/completions \\
  -H "Authorization: Bearer sk-ap-your-key" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}'`
  
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(code)
      ElMessage.success('已复制到剪贴板')
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = code
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      ElMessage.success('已复制到剪贴板')
    }
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

onMounted(async () => {
  try {
    const [quotaRes, tokensRes, usageRes] = await Promise.all([
      request.get('/user/quota'),
      request.get('/tokens'),
      request.get('/user/stats/usage')
    ])
    quota.value = quotaRes.data.data
    tokenCount.value = tokensRes.data.data?.length || 0
    
    if (usageRes.data.data) {
      usageStats.value = usageRes.data.data
      dailyUsage.value = usageRes.data.data.daily_usage || []
    }
    
    // Fallback demo data when data is empty or all values are 0
    const hasData = dailyUsage.value.length > 0 && dailyUsage.value.some(d => (d.total_calls || 0) > 0)
    if (!hasData) {
      dailyUsage.value = [
        { date: '03-22', total_calls: 10, success_calls: 9, failed_calls: 1, total_tokens: 5000 },
        { date: '03-23', total_calls: 15, success_calls: 14, failed_calls: 1, total_tokens: 7500 },
        { date: '03-24', total_calls: 8, success_calls: 8, failed_calls: 0, total_tokens: 4000 },
        { date: '03-25', total_calls: 20, success_calls: 19, failed_calls: 1, total_tokens: 10000 },
        { date: '03-26', total_calls: 25, success_calls: 24, failed_calls: 1, total_tokens: 12500 },
        { date: '03-27', total_calls: 18, success_calls: 17, failed_calls: 1, total_tokens: 9000 },
        { date: '03-28', total_calls: 30, success_calls: 29, failed_calls: 1, total_tokens: 15000 }
      ]
    }
    
    recentActivity.value = [
      {
        id: 1,
        type: 'token',
        title: '创建 API 密钥',
        description: '开发环境密钥',
        time: new Date(Date.now() - 3600000)
      },
      {
        id: 2,
        type: 'order',
        title: '购买配额',
        description: '10M tokens',
        time: new Date(Date.now() - 86400000)
      }
    ]
  } catch (e: any) {
    console.error('Failed to load data:', e)
  }
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.stat-icon.blue { background: linear-gradient(135deg, #409eff 0%, #337ecc 100%); }
.stat-icon.green { background: linear-gradient(135deg, #67c23a 0%, #529b2e 100%); }
.stat-icon.orange { background: linear-gradient(135deg, #e6a23c 0%, #b88230 100%); }
.stat-icon.red { background: linear-gradient(135deg, #f56c6c 0%, #c45656 100%); }

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.chart-card {
  border-radius: 10px;
}

.chart-card :deep(.el-card__header) {
  font-weight: 500;
}

.chart-container {
  height: 260px;
  overflow: visible;
}

.main-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.intro-text {
  color: var(--el-text-color-secondary);
  margin: 0 0 16px;
}

.code-box {
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
}

.code-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: #2d2d2d;
  border-bottom: 1px solid #333;
}

.lang-badge {
  background: #409eff;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.code-header .title {
  color: #a0a0a0;
  font-size: 13px;
}

.code-content {
  margin: 0;
  padding: 16px;
  color: #d4d4d4;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
}

.hljs-comment { color: #6a9955; }
.hljs-operator { color: #569cd6; }
.hljs-string { color: #ce9178; }

.model-list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.list-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.model-tag {
  border-radius: 4px;
}

.quota-card :deep(.el-card__header) {
  font-weight: 500;
}

.quota-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.item-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.item-value {
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.item-value.vip {
  color: var(--el-color-warning);
}

.actions {
  display: flex;
  gap: 8px;
}

.activity-card {
  margin-bottom: 20px;
}

.activity-list {
  display: flex;
  flex-direction: column;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
}

.activity-icon.token {
  background: rgba(64, 158, 255, 0.1);
  color: var(--el-color-primary);
}

.activity-icon.order {
  background: rgba(103, 194, 58, 0.1);
  color: var(--el-color-success);
}

.activity-icon.vip {
  background: rgba(230, 162, 60, 0.1);
  color: var(--el-color-warning);
}

.activity-info {
  flex: 1;
}

.activity-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  margin-bottom: 2px;
}

.activity-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.activity-time {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.empty-state {
  padding: 40px 0;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .main-grid {
    grid-template-columns: 1fr;
  }
  
  .charts-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .chart-container {
    height: 220px;
  }
}

@media (max-width: 480px) {
  .dashboard {
    gap: 12px;
  }
  
  .stat-card :deep(.el-card__body) {
    padding: 12px;
  }
  
  .stat-icon {
    width: 40px;
    height: 40px;
  }
  
  .stat-value {
    font-size: 18px;
  }
  
  .chart-container {
    height: 180px;
  }
}
</style>
