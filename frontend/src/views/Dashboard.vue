<template>
  <div class="dashboard">
    <!-- Stats Cards -->
    <div class="stats-grid">
      <el-card shadow="hover" class="stat-card" :class="{ 'urgent-card': isFreeExpiringSoon }">
        <div class="stat-icon blue">
          <el-icon size="24"><Coin /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">可用额度</div>
          <div class="stat-value">{{ formatQuota(getTotalAvailableQuota()) }}</div>
          <div class="stat-sub" v-if="!quota?.is_vip && quota?.free_expired_at">
            <el-tag type="danger" size="small" effect="plain">免费额度仅剩 {{ getFreeDaysRemaining() }} 天</el-tag>
          </div>
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
      
      <el-card shadow="hover" class="stat-card" :class="{ 'vip-card': quota?.is_vip }">
        <div class="stat-icon orange">
          <el-icon size="24"><Star /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">{{ quota?.is_vip ? 'VIP剩余天数' : '会员状态' }}</div>
          <div class="stat-value" :class="{ 'vip-value': quota?.is_vip }">
            {{ quota?.is_vip ? getVIPDaysRemaining() + ' 天' : '免费用户' }}
          </div>
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
          <div class="card-header">
            <span>我的额度</span>
            <el-tag v-if="quota?.is_vip" type="warning" size="small" effect="plain">
              <el-icon><Star /></el-icon> VIP会员
            </el-tag>
            <el-tag v-else type="info" size="small" effect="plain">
              免费用户
            </el-tag>
          </div>
        </template>
        
        <div class="quota-main">
          <div class="quota-big">
            <span class="quota-number">{{ formatQuota(getTotalAvailableQuota()) }}</span>
            <span class="quota-unit">Tokens</span>
          </div>
          <div class="quota-hint" :class="{ urgent: !quota?.is_vip && isFreeExpiringSoon() }">
            <el-icon><Timer /></el-icon>
            <span v-if="quota?.is_vip">VIP额度每月重置，还剩 {{ getVIPDaysRemaining() }} 天</span>
            <span v-else>免费额度仅剩 {{ getFreeDaysRemaining() }} 天，请尽快购买续命！</span>
          </div>
        </div>
        
        <el-divider style="margin: 16px 0" />
        
        <div class="quota-breakdown">
          <div class="quota-item-row">
            <span class="quota-label">免费额度</span>
            <span class="quota-value-row">
              <span class="quota-amount">{{ formatQuota(quota?.free_quota) }}</span>
              <el-tag type="info" size="small" effect="plain">{{ getFreeDaysRemaining() }}天后清零</el-tag>
            </span>
          </div>
          <div class="quota-item-row" v-if="quota?.is_vip">
            <span class="quota-label">VIP额度</span>
            <span class="quota-value-row">
              <span class="quota-amount vip">{{ formatQuota(quota?.vip_quota) }}</span>
              <el-tag type="warning" size="small" effect="plain">{{ getVIPDaysRemaining() }}天后重置</el-tag>
            </span>
          </div>
          <div class="quota-item-row" v-if="usageStats?.used_today">
            <span class="quota-label">今日消耗</span>
            <span class="quota-amount danger">{{ formatQuota(usageStats.used_today) }}</span>
          </div>
        </div>
        
        <el-divider style="margin: 16px 0" />
        
        <div class="actions">
          <el-button type="primary" size="default" @click="$router.push('/products')">
            <el-icon><ShoppingCart /></el-icon> 购买额度
          </el-button>
          <el-button v-if="!quota?.is_vip" type="warning" size="default" @click="$router.push('/vip')">
            <el-icon><Star /></el-icon> 开通VIP（更优惠）
          </el-button>
          <el-button v-else type="warning" plain size="default" @click="$router.push('/vip')">
            <el-icon><Star /></el-icon> VIP续费
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- Recent Activity -->
    <el-card class="activity-card">
      <template #header>
        <div class="card-header">
          <span>最近活动</span>
          <el-button text size="small" @click="$router.push('/activities')">
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
  ShoppingCart, ArrowRight, Timer
} from '@element-plus/icons-vue'
import request from '@/api/request'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent])

interface Quota {
  free_quota: number
  free_expired_at: string
  vip_quota: number
  vip_expired_at: string
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

function formatVIPExpiry(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  const now = new Date()
  const diff = date.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const dateDisplay = date.toLocaleDateString('zh-CN', {
    month: 'long',
    day: 'numeric',
  })
  return `${dateDisplay} (${days}天后)`
}

function getVIPDaysRemaining(): number {
  if (!quota.value?.vip_expired_at) return 0
  const expiry = new Date(quota.value.vip_expired_at)
  const now = new Date()
  const diff = expiry.getTime() - now.getTime()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

function getFreeDaysRemaining(): number {
  if (!quota.value?.free_expired_at) return 0
  const expiry = new Date(quota.value.free_expired_at)
  const now = new Date()
  const diff = expiry.getTime() - now.getTime()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

function getTotalAvailableQuota(): number {
  if (!quota.value) return 0
  return (quota.value.free_quota || 0) + (quota.value.vip_quota || 0)
}

function isFreeExpiringSoon(): boolean {
  if (!quota.value?.free_expired_at) return false
  return getFreeDaysRemaining() <= 3
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

    const activitiesRes = await request.get('/user/activities')
    if (activitiesRes.data.data && activitiesRes.data.data.length > 0) {
      recentActivity.value = activitiesRes.data.data.map((item: any) => ({
        id: item.id,
        type: item.type as 'token' | 'order' | 'vip',
        title: item.title,
        description: item.description,
        time: new Date(item.time)
      }))
    }
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

.stat-value.vip-value {
  color: var(--el-color-warning);
}

.vip-card :deep(.stat-icon) {
  background: linear-gradient(135deg, #e6a23c 0%, #b88230 100%);
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
