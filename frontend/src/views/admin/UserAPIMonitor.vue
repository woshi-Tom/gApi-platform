<template>
  <div class="user-api-monitor">
    <div class="page-header">
      <div class="header-left">
        <h2>用户 API 监控</h2>
        <p class="subtitle">实时监控用户API使用情况，快速发现异常</p>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="refreshAll">刷新</el-button>
        <el-button :icon="Download" @click="exportData">导出</el-button>
      </div>
    </div>

    <el-card shadow="hover" class="filter-card">
      <div class="filter-row">
        <div class="filter-item">
          <span class="filter-label">时间范围</span>
          <el-select v-model="filters.timeRange" @change="onFilterChange">
            <el-option label="今日" value="today" />
            <el-option label="本周" value="week" />
            <el-option label="本月" value="month" />
          </el-select>
        </div>
        <div class="filter-item">
          <span class="filter-label">用户等级</span>
          <el-select v-model="filters.level" @change="onFilterChange">
            <el-option label="全部" value="all" />
            <el-option label="免费" value="free" />
            <el-option label="VIP" value="vip" />
            <el-option label="企业" value="enterprise" />
          </el-select>
        </div>
        <div class="filter-item">
          <span class="filter-label">状态</span>
          <el-select v-model="filters.status" @change="onFilterChange">
            <el-option label="全部" value="all" />
            <el-option label="正常" value="normal" />
            <el-option label="异常" value="abnormal" />
          </el-select>
        </div>
      </div>
    </el-card>

    <div class="stats-overview">
      <div class="stat-card">
        <div class="stat-icon blue">
          <el-icon size="24"><TrendCharts /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">总请求数</div>
          <div class="stat-value">{{ formatNumber(overviewStats.total_requests) }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon green">
          <el-icon size="24"><Coin /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">Token消耗</div>
          <div class="stat-value">{{ formatNumber(overviewStats.total_tokens, true) }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" :class="overviewStats.failure_rate > 5 ? 'red' : 'orange'">
          <el-icon size="24"><Warning /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">失败率</div>
          <div class="stat-value" :class="{ 'text-danger': overviewStats.failure_rate > 10 }">
            {{ overviewStats.failure_rate.toFixed(2) }}%
          </div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" :class="overviewStats.abnormal_users > 0 ? 'red' : 'cyan'">
          <el-icon size="24"><User /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">异常用户</div>
          <div class="stat-value" :class="{ 'text-danger': overviewStats.abnormal_users > 0 }">
            {{ overviewStats.abnormal_users }}
          </div>
        </div>
      </div>
    </div>

    <el-collapse v-model="alertCollapse" class="alert-section">
      <el-collapse-item title="异常告警" name="alerts">
        <template #title>
          <div class="collapse-title">
            <el-icon color="#f56c6c"><Warning /></el-icon>
            <span>异常告警</span>
            <el-tag type="danger" size="small">{{ abnormalUsers.length }} 个异常</el-tag>
          </div>
        </template>
        <div v-if="abnormalUsers.length === 0" class="no-alert">
          <el-icon color="#67c23a"><CircleCheck /></el-icon>
          <span>暂无异常用户</span>
        </div>
        <div v-else class="alert-list">
          <div v-for="user in abnormalUsers" :key="user.user_id" class="alert-item">
            <div class="alert-info">
              <span class="alert-username">{{ user.username }}</span>
              <span class="alert-email">{{ user.email }}</span>
            </div>
            <div class="alert-stats">
              <span class="alert-rate text-danger">{{ user.failure_rate.toFixed(1) }}%</span>
              <span class="alert-requests">{{ formatNumber(user.requests) }} 次请求</span>
            </div>
            <div class="alert-actions">
              <el-button size="small" @click="showUserDetail(user.user_id)">查看</el-button>
              <el-button size="small" type="warning" @click="handleUser(user.user_id)">处理</el-button>
            </div>
          </div>
        </div>
      </el-collapse-item>
    </el-collapse>

    <el-card shadow="hover" class="ranking-card">
      <template #header>
        <div class="card-header">
          <span>排行榜</span>
          <el-radio-group v-model="rankingType" @change="fetchRanking">
            <el-radio-button value="requests">请求量</el-radio-button>
            <el-radio-button value="tokens">Token消耗</el-radio-button>
            <el-radio-button value="failed_rate">失败率</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div class="chart-container">
        <v-chart :option="rankingChartOption" autoresize />
      </div>
    </el-card>

    <el-card shadow="hover" class="table-card">
      <template #header>
        <div class="card-header">
          <span>用户明细</span>
          <span class="table-count">共 {{ totalUsers }} 个用户</span>
        </div>
      </template>
      <el-table :data="userList" stripe @sort-change="onTableSort">
        <el-table-column prop="username" label="用户" min-width="120">
          <template #default="{ row }">
            <div class="user-cell">
              <span class="username">{{ row.username }}</span>
              <span class="email">{{ row.email }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="level" label="等级" width="80">
          <template #default="{ row }">
            <el-tag :type="getLevelType(row.level)" size="small">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="requests" label="请求数" sortable="custom" width="120" align="right">
          <template #default="{ row }">
            {{ formatNumber(row.requests) }}
          </template>
        </el-table-column>
        <el-table-column prop="tokens" label="Token" sortable="custom" width="120" align="right">
          <template #default="{ row }">
            {{ formatNumber(row.tokens, true) }}
          </template>
        </el-table-column>
        <el-table-column prop="failure_rate" label="失败率" sortable="custom" width="100" align="right">
          <template #default="{ row }">
            <span :class="{ 'text-danger': row.failure_rate > 30 }">
              {{ row.failure_rate.toFixed(2) }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.failure_rate > 30 ? 'danger' : 'success'" size="small">
              {{ row.failure_rate > 30 ? '异常' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button size="small" link @click="showUserDetail(row.user_id)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="totalUsers"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchUserList"
          @current-change="fetchUserList"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="用户详情" width="900px" destroy-on-close>
      <div v-if="userDetail" class="user-detail">
        <div class="detail-header">
          <div class="detail-user">
            <h3>{{ userDetail.user?.username }}</h3>
            <p>{{ userDetail.user?.email }}</p>
          </div>
          <el-tag :type="getLevelType(userDetail.user?.level)" size="large">
            {{ userDetail.user?.level }}
          </el-tag>
        </div>

        <div class="detail-stats">
          <div class="detail-stat">
            <span class="label">请求数</span>
            <span class="value">{{ formatNumber(userDetail.stats?.requests) }}</span>
          </div>
          <div class="detail-stat">
            <span class="label">Token消耗</span>
            <span class="value">{{ formatNumber(userDetail.stats?.tokens, true) }}</span>
          </div>
          <div class="detail-stat">
            <span class="label">失败数</span>
            <span class="value">{{ formatNumber(userDetail.stats?.failed) }}</span>
          </div>
          <div class="detail-stat">
            <span class="label">失败率</span>
            <span class="value" :class="{ 'text-danger': userDetail.stats?.failure_rate > 30 }">
              {{ userDetail.stats?.failure_rate?.toFixed(2) }}%
            </span>
          </div>
        </div>

        <div class="detail-time-selector">
          <el-radio-group v-model="detailTimeRange" @change="fetchUserDetail(currentDetailUserId)">
            <el-radio-button value="today">今日</el-radio-button>
            <el-radio-button value="week">本周</el-radio-button>
            <el-radio-button value="month">本月</el-radio-button>
          </el-radio-group>
        </div>

        <div class="detail-charts">
          <div class="chart-wrapper">
            <h4>每日趋势</h4>
            <div class="chart-container-small">
              <v-chart :option="trendChartOption" autoresize />
            </div>
          </div>
          <div class="chart-wrapper">
            <h4>模型分布</h4>
            <div class="chart-container-small">
              <v-chart :option="modelChartOption" autoresize />
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { Refresh, Download, TrendCharts, Coin, Warning, User, CircleCheck } from '@element-plus/icons-vue'
import request from '@/api/request'

use([CanvasRenderer, BarChart, LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

const filters = ref({
  timeRange: 'week',
  level: 'all',
  status: 'all'
})

const pagination = ref({
  page: 1,
  pageSize: 20
})

const sortConfig = ref({
  prop: 'failure_rate',
  order: 'descending'
})

const overviewStats = ref({
  total_requests: 0,
  total_tokens: 0,
  failure_rate: 0,
  abnormal_users: 0
})

const rankingType = ref('requests')
const rankingData = ref<any[]>([])
const userList = ref<any[]>([])
const totalUsers = ref(0)
const abnormalUsers = ref<any[]>([])

const detailVisible = ref(false)
const userDetail = ref<any>(null)
const detailTimeRange = ref('week')
const currentDetailUserId = ref<number>(0)

const alertCollapse = ref(['alerts'])

const timeRangeMap: Record<string, string> = {
  today: 'today',
  week: 'week',
  month: 'month'
}

function formatNumber(n: number, isToken = false): string {
  if (!n) return '0'
  if (isToken) {
    if (n >= 1e9) return (n / 1e9).toFixed(2) + 'B'
    if (n >= 1e6) return (n / 1e6).toFixed(2) + 'M'
    if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
    return n.toString()
  }
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

function getLevelType(level: string): string {
  switch (level) {
    case 'vip': return 'warning'
    case 'enterprise': return 'danger'
    default: return 'info'
  }
}

async function fetchOverview() {
  try {
    const res = await request.get('/admin/stats/user-overview', {
      params: { time_range: filters.value.timeRange }
    })
    if (res.data?.success) {
      overviewStats.value = res.data.data
    }
  } catch (e) {
    ElMessage.error('加载概览数据失败')
  }
}

async function fetchRanking() {
  try {
    const res = await request.get('/admin/stats/user-ranking', {
      params: {
        type: rankingType.value,
        limit: 10,
        time_range: filters.value.timeRange
      }
    })
    if (res.data?.success) {
      rankingData.value = res.data.data
    }
  } catch (e) {
    ElMessage.error('加载排名数据失败')
  }
}

async function fetchUserList() {
  try {
    const res = await request.get('/admin/stats/user-list', {
      params: {
        page: pagination.value.page,
        page_size: pagination.value.pageSize,
        sort_by: sortConfig.value.prop,
        order: sortConfig.value.order === 'ascending' ? 'asc' : 'desc',
        time_range: filters.value.timeRange,
        level: filters.value.level,
        status: filters.value.status
      }
    })
    if (res.data?.success) {
      userList.value = res.data.data
      totalUsers.value = res.data.pagination?.total || 0
    }
  } catch (e) {
    ElMessage.error('加载用户列表失败')
  }
}

async function fetchAbnormalUsers() {
  try {
    const res = await request.get('/admin/stats/abnormal-users', {
      params: {
        threshold: 30,
        limit: 20,
        time_range: filters.value.timeRange
      }
    })
    if (res.data?.success) {
      abnormalUsers.value = res.data.data
    }
  } catch (e) {
    ElMessage.error('加载异常用户失败')
  }
}

async function fetchUserDetail(userId: number) {
  try {
    const res = await request.get(`/admin/stats/user/${userId}/detail`, {
      params: { time_range: detailTimeRange.value }
    })
    if (res.data?.success) {
      userDetail.value = res.data.data
    }
  } catch (e) {
    ElMessage.error('加载用户详情失败')
  }
}

function showUserDetail(userId: number) {
  currentDetailUserId.value = userId
  detailVisible.value = true
  fetchUserDetail(userId)
}

function handleUser(userId: number) {
  ElMessage.info(`处理用户 ${userId} - 功能开发中`)
}

function refreshAll() {
  fetchOverview()
  fetchRanking()
  fetchUserList()
  fetchAbnormalUsers()
  ElMessage.success('刷新成功')
}

function exportData() {
  ElMessage.info('导出功能开发中')
}

function onFilterChange() {
  pagination.value.page = 1
  fetchOverview()
  fetchRanking()
  fetchUserList()
  fetchAbnormalUsers()
}

function onTableSort({ prop, order }: any) {
  sortConfig.value.prop = prop || 'failure_rate'
  sortConfig.value.order = order || 'descending'
  fetchUserList()
}

const rankingChartOption = computed(() => {
  const data = rankingData.value.slice(0, 10)
  const labels = data.map(d => d.username || `User ${d.user_id}`)
  const values = data.map(d => {
    if (rankingType.value === 'requests') return d.requests
    if (rankingType.value === 'tokens') return d.tokens
    return d.failure_rate
  })

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: labels,
      axisLabel: {
        rotate: 45,
        interval: 0
      }
    },
    yAxis: {
      type: 'value',
      name: rankingType.value === 'failed_rate' ? '失败率(%)' : '数量',
      axisLabel: {
        formatter: (v: number) => rankingType.value === 'failed_rate' ? `${v}%` : formatNumber(v, true)
      }
    },
    series: [{
      type: 'bar',
      data: values,
      itemStyle: {
        color: rankingType.value === 'failed_rate' ? '#f56c6c' : '#409eff',
        borderRadius: [4, 4, 0, 0]
      },
      barMaxWidth: 50
    }]
  }
})

const trendChartOption = computed(() => {
  const trends = userDetail.value?.daily_trends || []
  const dates = trends.map((d: any) => d.date)
  const requests = trends.map((d: any) => d.requests)
  const tokens = trends.map((d: any) => d.tokens)

  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['请求数', 'Token'], top: 0 },
    grid: { left: '3%', right: '4%', bottom: '10%', top: '40px', containLabel: true },
    xAxis: { type: 'category', data: dates, boundaryGap: false },
    yAxis: [
      { type: 'value', name: '请求数' },
      { type: 'value', name: 'Token', axisLabel: { formatter: (v: number) => formatNumber(v, true) } }
    ],
    series: [
      { name: '请求数', type: 'line', data: requests, smooth: true, itemStyle: { color: '#409eff' } },
      { name: 'Token', type: 'line', yAxisIndex: 1, data: tokens, smooth: true, itemStyle: { color: '#67c23a' } }
    ]
  }
})

const modelChartOption = computed(() => {
  const models = userDetail.value?.model_distribution || []
  
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', right: '5%', top: 'center' },
    series: [{
      type: 'pie',
      radius: ['35%', '60%'],
      center: ['35%', '50%'],
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14 } },
      data: models.map((m: any, i: number) => ({
        name: m.model || 'Unknown',
        value: m.tokens || 0,
        itemStyle: { color: ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399'][i % 5] }
      }))
    }]
  }
})

onMounted(() => {
  fetchOverview()
  fetchRanking()
  fetchUserList()
  fetchAbnormalUsers()
})
</script>

<style scoped>
.user-api-monitor {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.filter-card {
  border-radius: 10px;
}

.filter-row {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.stats-overview {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
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
.stat-icon.cyan { background: linear-gradient(135deg, #17c0eb 0%, #13a6cf 100%); }

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

.text-danger {
  color: #f56c6c !important;
}

.alert-section {
  border-radius: 10px;
}

.collapse-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.no-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px;
  color: #67c23a;
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.alert-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: rgba(245, 108, 108, 0.05);
  border-radius: 8px;
  border-left: 3px solid #f56c6c;
}

.alert-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.alert-username {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.alert-email {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.alert-stats {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.alert-rate {
  font-weight: 600;
  font-size: 16px;
}

.alert-requests {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.alert-actions {
  display: flex;
  gap: 8px;
}

.ranking-card, .table-card {
  border-radius: 10px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-count {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.chart-container {
  height: 300px;
}

.user-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.username {
  font-weight: 500;
}

.email {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.user-detail {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-user h3 {
  margin: 0;
  font-size: 18px;
}

.detail-user p {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.detail-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 16px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.detail-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.detail-stat .label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.detail-stat .value {
  font-size: 18px;
  font-weight: 600;
}

.detail-time-selector {
  display: flex;
  justify-content: center;
}

.detail-charts {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.chart-wrapper h4 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 500;
}

.chart-container-small {
  height: 200px;
}

@media (max-width: 1200px) {
  .stats-overview {
    grid-template-columns: repeat(2, 1fr);
  }
  .detail-charts {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-overview {
    grid-template-columns: 1fr;
  }
  .filter-row {
    flex-direction: column;
  }
}
</style>
