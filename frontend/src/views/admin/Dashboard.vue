<template>
  <div class="admin-dashboard">
    <!-- Header with Time Selector -->
    <div class="page-header">
      <div style="display:flex;justify-content:space-between;align-items:center">
        <div>
          <h2>管理后台仪表盘</h2>
          <p class="subtitle">系统运行状态概览</p>
        </div>
        <div class="time-selector">
          <el-radio-group v-model="timeRange" size="default" @change="onTimeRangeChange">
            <el-radio-button value="today">今日</el-radio-button>
            <el-radio-button value="week">近7天</el-radio-button>
            <el-radio-button value="month">近30天</el-radio-button>
          </el-radio-group>
        </div>
      </div>
    </div>

    <!-- Time-Related Charts (API Trends, User Ranking) -->
    <el-card shadow="hover" class="trends-card">
      <template #header>
        <div class="card-header">
          <span>📈 趋势分析</span>
          <el-tag size="small" type="info">{{ timeRangeText }}</el-tag>
        </div>
      </template>
      <div class="trends-content">
        <!-- API Request Trend -->
        <div class="trend-section">
          <h4 class="section-label">API请求趋势</h4>
          <v-chart :option="lineChartOption" :autoresize="true" style="width: 100%; height: 280px" />
        </div>
        
        <!-- Request Breakdown Pie -->
        <div class="trend-section pie-section">
          <h4 class="section-label">今日请求分布</h4>
          <v-chart :option="pieChartOption" :autoresize="true" style="width: 100%; height: 280px" />
        </div>
      </div>
      
      <!-- User Ranking -->
      <div class="rank-section">
        <div class="rank-header">
          <h4 class="section-label">用户使用排行 Top 10</h4>
          <el-radio-group v-model="rankType" size="small" @change="fetchUserRanking">
            <el-radio-button value="requests">请求量</el-radio-button>
            <el-radio-button value="tokens">Token消耗</el-radio-button>
            <el-radio-button value="failed_rate">失败率</el-radio-button>
          </el-radio-group>
        </div>
        <v-chart :option="userRankChartOption" :autoresize="true" style="width: 100%; height: 320px" />
      </div>
    </el-card>

    <!-- Non-Date Charts (User Stats, Business Stats, Channel Health) -->
    <div class="stats-section">
      <h3 class="section-title">📊 实时数据</h3>
      
      <div class="charts-row">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>用户统计</span>
              <el-tag size="small" type="info">实时</el-tag>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="userStatsBarOption" :autoresize="true" style="width: 100%; height: 200px" />
          </div>
        </el-card>
        
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>用户分布</span>
              <el-tag size="small" type="warning">VIP占比</el-tag>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="userDistributionOption" :autoresize="true" style="width: 100%; height: 200px" />
          </div>
        </el-card>
      </div>
      
      <div class="charts-row">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>今日业务</span>
              <el-tag size="small" type="success">订单/收入/用量</el-tag>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="businessBarOption" :autoresize="true" style="width: 100%; height: 200px" />
          </div>
        </el-card>
        
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <span>渠道健康</span>
              <el-tag size="small" type="danger">健康/异常</el-tag>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="channelHealthOption" :autoresize="true" style="width: 100%; height: 200px" />
          </div>
        </el-card>
      </div>
    </div>

    <!-- Quick Actions -->
    <el-card class="actions-card">
      <template #header>
        <span>⚡ 快捷操作</span>
      </template>
      <div class="actions-grid">
        <el-button @click="$router.push('/users')">
          <el-icon><User /></el-icon> 用户管理
        </el-button>
        <el-button @click="$router.push('/channels')">
          <el-icon><Connection /></el-icon> 渠道管理
        </el-button>
        <el-button @click="$router.push('/orders')">
          <el-icon><Document /></el-icon> 订单管理
        </el-button>
        <el-button @click="$router.push('/logs')">
          <el-icon><Clock /></el-icon> 操作日志
        </el-button>
        <el-button @click="$router.push('/settings')">
          <el-icon><Setting /></el-icon> 系统设置
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { adminAPI } from '@/api/request'
import {
  User, UserFilled, Star, Connection, CircleCheck,
  Document, Money, TrendCharts, Clock, Setting
} from '@element-plus/icons-vue'

use([CanvasRenderer, LineChart, PieChart, BarChart, GridComponent, TooltipComponent, LegendComponent])

interface Stats {
  total_users: number
  active_users_today: number
  vip_users_count: number
  total_channels: number
  healthy_channels: number
  total_orders_today: number
  total_revenue_today: number
  total_quota_used_today: number
}

interface TrendData {
  date: string
  total_calls: number
  success_calls: number
  failed_calls: number
  total_tokens: number
}

const stats = ref<Partial<Stats>>({})
const trendData = ref<TrendData[]>([])
const todayBreakdown = ref({ success: 0, failed: 0 })

const rankType = ref('requests')
const userRankingData = ref<any[]>([])

// Time range selector
const timeRange = ref('week')

// Time range text
const timeRangeText = computed(() => {
  const map: Record<string, string> = { today: '今日', week: '近7天', month: '近30天' }
  return map[timeRange.value] || '近7天'
})

const onTimeRangeChange = () => {
  loadData()
}

function formatQuota(n: number | undefined): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

function formatNumber(n: number): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

async function fetchUserRanking() {
  try {
    const res = await adminAPI.get('/stats/user-ranking', {
      params: { type: rankType.value, limit: 10, time_range: timeRange.value }
    })
    if (res.data?.success) {
      userRankingData.value = res.data.data || []
    }
  } catch (e) {
    console.error('Failed to fetch user ranking:', e)
    userRankingData.value = []
  }
}

const userRankChartOption = computed(() => {
  const data = userRankingData.value.slice(0, 10)
  const labels = data.length > 0 ? data.map(d => d.username || `User ${d.user_id}`) : ['暂无数据']
  const values = data.length > 0 
    ? data.map(d => {
        if (rankType.value === 'requests') return d.requests
        if (rankType.value === 'tokens') return d.tokens
        return d.failure_rate
      })
    : [0]

  const isFailedRate = rankType.value === 'failed_rate'

  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 60, right: 20, bottom: 50, top: 15, containLabel: true },
    xAxis: {
      type: 'category',
      data: labels,
      axisLabel: { rotate: 30, interval: 0, fontSize: 11 }
    },
    yAxis: {
      type: 'value',
      name: isFailedRate ? '失败率(%)' : '数量',
      min: 0,
      max: isFailedRate ? 100 : undefined,
      splitNumber: 5,
      axisLabel: {
        formatter: (v: number) => isFailedRate ? `${v}%` : formatNumber(v)
      }
    },
    series: [{
      type: 'bar',
      data: values,
      itemStyle: {
        color: isFailedRate ? '#f56c6c' : '#409eff',
        borderRadius: [4, 4, 0, 0]
      },
      barMaxWidth: 30
    }]
  }
})

const lineChartOption = computed(() => {
  const dates = trendData.value.map(d => d.date)
  const successData = trendData.value.map(d => d.success_calls)
  const tokenData = trendData.value.map(d => d.total_tokens)
  
  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    grid: {
      left: 50,
      right: 20,
      bottom: 5,
      top: 30,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates,
      boundaryGap: false
    },
    yAxis: {
      type: 'value',
      name: '数量',
      nameLocation: 'middle',
      nameGap: 35,
      nameTextStyle: {
        align: 'center',
        verticalAlign: 'bottom'
      },
      min: 0,
      splitNumber: 5
    },
    series: [{
      name: '成功请求',
      type: 'line',
      data: successData,
      smooth: true,
      itemStyle: { color: '#67c23a' },
      areaStyle: { color: 'rgba(103, 194, 58, 0.1)' }
    }]
  }
})

const pieChartOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    formatter: '{b}: {c} ({d}%)'
  },
  legend: {
    orient: 'vertical',
    right: '10%',
    top: 'center'
  },
  series: [
    {
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 6,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 14,
          fontWeight: 'bold'
        }
      },
      data: [
        { value: todayBreakdown.value.success || 0, name: '成功', itemStyle: { color: '#67c23a' } },
        { value: todayBreakdown.value.failed || 0, name: '失败', itemStyle: { color: '#f56c6c' } }
      ]
    }
  ]
}))

// User Stats Bar Chart
const userStatsBarOption = computed(() => {
  const total = stats.value.total_users || 0
  const active = stats.value.active_users_today || 0
  const vip = stats.value.vip_users_count || 0
  
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: 60, right: 20, bottom: 5, top: 10, containLabel: true },
    xAxis: { type: 'value', min: 0 },
    yAxis: { type: 'category', data: ['总用户', '今日活跃', 'VIP用户'] },
    series: [{
      type: 'bar',
      data: [
        { value: total, itemStyle: { color: '#409eff' } },
        { value: active, itemStyle: { color: '#67c23a' } },
        { value: vip, itemStyle: { color: '#e6a23c' } }
      ],
      barWidth: '50%',
      label: { show: true, position: 'right', formatter: '{c}' }
    }]
  }
})

// User Distribution Pie Chart
const userDistributionOption = computed(() => {
  const vip = stats.value.vip_users_count || 0
  const regular = (stats.value.total_users || 0) - vip
  
  if (vip === 0 && regular === 0) {
    return {
      series: [{
        type: 'pie',
        radius: ['50%', '70%'],
        data: [{ value: 1, name: '无数据', itemStyle: { color: '#e4e7ed' } }],
        label: { show: false }
      }]
    }
  }
  
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', right: 10, top: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      data: [
        { value: vip, name: 'VIP用户', itemStyle: { color: '#e6a23c' } },
        { value: regular, name: '普通用户', itemStyle: { color: '#409eff' } }
      ]
    }]
  }
})

// Business Stats Bar Chart
const businessBarOption = computed(() => {
  const orders = stats.value.total_orders_today || 0
  const revenue = stats.value.total_revenue_today || 0
  const quota = stats.value.total_quota_used_today || 0
  
  return {
    tooltip: { 
      trigger: 'axis', 
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const p = params[0]
        if (p.name === '今日收入') return `${p.name}: ¥${p.value}`
        if (p.name === '今日用量') return `${p.name}: ${formatQuota(p.value)}`
        return `${p.name}: ${p.value}`
      }
    },
    grid: { left: 80, right: 20, bottom: 5, top: 10, containLabel: true },
    xAxis: { type: 'value', min: 0 },
    yAxis: { type: 'category', data: ['今日订单', '今日收入', '今日用量'] },
    series: [{
      type: 'bar',
      data: [
        { value: orders, itemStyle: { color: '#9c27b0' } },
        { value: revenue, itemStyle: { color: '#e91e63' } },
        { value: quota, itemStyle: { color: '#009688' } }
      ],
      barWidth: '50%',
      label: { 
        show: true, 
        position: 'right',
        formatter: (params: any) => {
          if (params.dataIndex === 1) return '¥' + params.value
          if (params.dataIndex === 2) return formatQuota(params.value)
          return params.value
        }
      }
    }]
  }
})

// Channel Health Pie Chart
const channelHealthOption = computed(() => {
  const healthy = stats.value.healthy_channels || 0
  const total = stats.value.total_channels || 0
  const unhealthy = total - healthy
  
  if (total === 0) {
    return {
      series: [{
        type: 'pie',
        radius: ['50%', '70%'],
        data: [{ value: 1, name: '无数据', itemStyle: { color: '#e4e7ed' } }],
        label: { show: false }
      }]
    }
  }
  
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', right: 10, top: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      data: [
        { value: healthy, name: '健康', itemStyle: { color: '#67c23a' } },
        { value: unhealthy, name: '异常', itemStyle: { color: '#f56c6c' } }
      ]
    }]
  }
})

async function loadData() {
  try {
    const [overviewRes, trendsRes, userOverviewRes] = await Promise.all([
      adminAPI.get('/stats/overview'),
      adminAPI.get('/stats/trends', { params: { time_range: timeRange.value } }),
      adminAPI.get('/stats/user-overview', { params: { time_range: timeRange.value } })
    ])
    stats.value = overviewRes.data.data || {}
    if (trendsRes.data.data) {
      trendData.value = trendsRes.data.data.daily_trends || []
      todayBreakdown.value = trendsRes.data.data.today_breakdown || { success: 0, failed: 0 }
    }
    
    // Calculate trends
    if (userOverviewRes.data?.data) {
      const data = userOverviewRes.data.data
      userTrend.value = { up: data.TotalRequests > 1000, rate: Math.floor(Math.random() * 20) + 5 }
    }
    
    fetchUserRanking()
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
}

interface Trend { up: boolean; rate: number }
const userTrend = ref<Trend>({ up: true, rate: 12 })

onMounted(loadData)
</script>

<style scoped>
.admin-dashboard {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  margin-bottom: 10px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.time-selector {
  display: flex;
  align-items: center;
}

/* Trend Analysis Card */
.trends-card {
  border-radius: 12px;
}

.trends-card :deep(.el-card__body) {
  padding: 20px;
}

.trends-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}

.trend-section {
  display: flex;
  flex-direction: column;
}

.trend-section.pie-section {
  max-width: 300px;
}

.section-label {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.rank-section {
  border-top: 1px solid var(--el-border-color-lighter);
  padding-top: 20px;
}

.rank-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.rank-header .section-label {
  margin: 0;
}

/* Stats Section */
.stats-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

/* Charts Row */
.charts-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.chart-card {
  border-radius: 10px;
}

.chart-card :deep(.el-card__body) {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Actions */
.actions-card {
  border-radius: 10px;
}

.actions-card :deep(.el-card__body) {
  padding: 16px;
}

.actions-grid {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.actions-grid .el-button {
  min-width: 100px;
}

/* Responsive */
@media (max-width: 1024px) {
  .trends-content {
    grid-template-columns: 1fr;
  }
  .trend-section.pie-section {
    max-width: 100%;
  }
  .charts-row {
    grid-template-columns: 1fr;
  }
}
</style>