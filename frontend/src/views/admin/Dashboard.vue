<template>
  <div class="admin-dashboard">
    <div class="page-header">
      <h2>管理后台仪表盘</h2>
      <p class="subtitle">系统运行状态概览</p>
    </div>

    <!-- User Stats -->
    <div class="stats-section">
      <h3 class="section-title">用户统计</h3>
      <div class="stats-grid">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon blue">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total_users || 0 }}</div>
            <div class="stat-label">总用户数</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon green">
            <el-icon><UserFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.active_users_today || 0 }}</div>
            <div class="stat-label">今日活跃</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon orange">
            <el-icon><Star /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.vip_users_count || 0 }}</div>
            <div class="stat-label">VIP用户</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon red">
            <el-icon><Connection /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total_channels || 0 }}</div>
            <div class="stat-label">渠道数量</div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- Business Stats -->
    <div class="stats-section">
      <h3 class="section-title">业务统计</h3>
      <div class="stats-grid">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon cyan">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.healthy_channels || 0 }}</div>
            <div class="stat-label">健康渠道</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon purple">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total_orders_today || 0 }}</div>
            <div class="stat-label">今日订单</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon pink">
            <el-icon><Money /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">¥{{ stats.total_revenue_today || 0 }}</div>
            <div class="stat-label">今日收入</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon teal">
            <el-icon><TrendCharts /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ formatQuota(stats.total_quota_used_today) }}</div>
            <div class="stat-label">今日用量</div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="charts-section">
      <div class="charts-grid">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <span>API请求趋势 (近7天)</span>
          </template>
          <div class="chart-container">
            <v-chart :option="lineChartOption" :autoresize="true" style="width: 100%; height: 320px" />
          </div>
        </el-card>
        
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <span>用户使用排行 (Top 10)</span>
          </template>
          <div class="chart-container">
            <div class="rank-tabs">
              <el-radio-group v-model="rankType" size="small" @change="fetchUserRanking">
                <el-radio-button value="requests">请求量</el-radio-button>
                <el-radio-button value="tokens">Token消耗</el-radio-button>
                <el-radio-button value="failed_rate">失败率</el-radio-button>
              </el-radio-group>
            </div>
            <v-chart :option="userRankChartOption" :autoresize="true" style="width: 100%; height: 320px" />
          </div>
        </el-card>
      </div>
    </div>

    <!-- Quick Actions -->
    <el-card class="actions-card">
      <template #header>
        <span>快捷操作</span>
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
import request from '@/api/request'
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
    const res = await request.get('/admin/stats/user-ranking', {
      params: { type: rankType.value, limit: 10, time_range: 'week' }
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

onMounted(async () => {
  try {
    const [overviewRes, trendsRes] = await Promise.all([
      request.get('/admin/stats/overview'),
      request.get('/admin/stats/trends')
    ])
    stats.value = overviewRes.data.data || {}
    if (trendsRes.data.data) {
      trendData.value = trendsRes.data.data.daily_trends || []
      todayBreakdown.value = trendsRes.data.data.today_breakdown || { success: 0, failed: 0 }
    }
    // Use mock data for demo if all values are 0
    if (trendData.value.length > 0 && trendData.value.every(d => d.total_calls === 0)) {
      trendData.value = [
        { date: '03-22', total_calls: 125, success_calls: 120, failed_calls: 5, total_tokens: 50000 },
        { date: '03-23', total_calls: 230, success_calls: 225, failed_calls: 5, total_tokens: 95000 },
        { date: '03-24', total_calls: 180, success_calls: 175, failed_calls: 5, total_tokens: 72000 },
        { date: '03-25', total_calls: 310, success_calls: 305, failed_calls: 5, total_tokens: 124000 },
        { date: '03-26', total_calls: 450, success_calls: 440, failed_calls: 10, total_tokens: 180000 },
        { date: '03-27', total_calls: 380, success_calls: 370, failed_calls: 10, total_tokens: 152000 },
        { date: '03-28', total_calls: 520, success_calls: 510, failed_calls: 10, total_tokens: 208000 }
      ]
      todayBreakdown.value = { success: 510, failed: 10 }
    }
    fetchUserRanking()
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
})
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

.stats-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  border-radius: 10px;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: #fff;
}

.stat-icon.blue { background: linear-gradient(135deg, #409eff 0%, #337ecc 100%); }
.stat-icon.green { background: linear-gradient(135deg, #67c23a 0%, #529b2e 100%); }
.stat-icon.orange { background: linear-gradient(135deg, #e6a23c 0%, #b88230 100%); }
.stat-icon.red { background: linear-gradient(135deg, #f56c6c 0%, #c45656 100%); }
.stat-icon.cyan { background: linear-gradient(135deg, #17c0eb 0%, #13a6cf 100%); }
.stat-icon.purple { background: linear-gradient(135deg, #9c27b0 0%, #7b1fa2 100%); }
.stat-icon.pink { background: linear-gradient(135deg, #e91e63 0%, #c2185b 100%); }
.stat-icon.teal { background: linear-gradient(135deg, #009688 0%, #00796b 100%); }

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.stat-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.charts-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
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
  height: 320px;
}

.rank-tabs {
  margin-bottom: 8px;
}

.rank-tabs :deep(.el-radio-group) {
  display: flex;
}

.actions-card :deep(.el-card__header) {
  font-weight: 500;
}

.actions-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.actions-grid .el-button {
  min-width: 100px;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .charts-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
