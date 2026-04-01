<template>
  <div class="activities-view">
    <el-card shadow="hover" class="activities-card">
      <template #header>
        <div class="card-header">
          <span>最近活动</span>
        </div>
      </template>
      
      <div class="filter-tabs">
        <el-radio-group v-model="filterType" size="default">
          <el-radio-button label="all">全部</el-radio-button>
          <el-radio-button label="order">订单</el-radio-button>
          <el-radio-button label="vip">VIP</el-radio-button>
          <el-radio-button label="token">API调用</el-radio-button>
          <el-radio-button label="login">登录</el-radio-button>
        </el-radio-group>
      </div>

      <el-table :data="paginatedActivities" v-loading="loading" stripe style="width: 100%; margin-top: 16px;">
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <div class="type-cell">
              <span class="type-icon" :class="'icon-' + row.type"></span>
              <span>{{ typeLabel(row.type) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="200">
          <template #default="{ row }">
            <span class="activity-title" :class="{ 'text-success': row.status === 'completed' }">
              {{ row.title }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="时间" width="160">
          <template #default="{ row }">
            <span class="time-cell">{{ formatTime(row.time) }}</span>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!loading && filteredActivities.length === 0" class="empty-state">
        <el-empty description="暂无活动记录" :image-size="80" />
      </div>

      <div class="pagination-wrapper" v-if="filteredActivities.length > 0">
        <el-pagination
          background
          layout="total, prev, pager, next"
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="filteredActivities.length"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface Activity {
  id: number
  type: 'order' | 'vip' | 'token' | 'login'
  title: string
  description: string
  time: string
  status?: string
}

const loading = ref(false)
const activities = ref<Activity[]>([])
const filterType = ref('all')
const currentPage = ref(1)
const pageSize = ref(10)

const filteredActivities = computed(() => {
  if (filterType.value === 'all') {
    return activities.value
  }
  return activities.value.filter(a => a.type === filterType.value)
})

const paginatedActivities = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredActivities.value.slice(start, end)
})

function typeLabel(type: string): string {
  const labels: Record<string, string> = {
    order: '订单',
    vip: 'VIP',
    token: 'API',
    login: '登录'
  }
  return labels[type] || type
}

function formatTime(timeStr: string): string {
  const date = new Date(timeStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return '刚刚'
  if (diffMins < 60) return `${diffMins}分钟前`
  if (diffHours < 24) return `${diffHours}小时前`
  if (diffDays < 7) return `${diffDays}天前`
  
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function loadActivities() {
  loading.value = true
  try {
    const res = await request.get('/user/activities')
    if (res.data.data) {
      activities.value = res.data.data.map((item: any) => ({
        id: item.id,
        type: item.type,
        title: item.title,
        description: item.description,
        time: item.time,
        status: item.status
      }))
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '加载活动记录失败')
  } finally {
    loading.value = false
  }
}

watch(filterType, () => {
  currentPage.value = 1
})

onMounted(loadActivities)
</script>

<style scoped>
.activities-view {
  padding: 20px;
}

.activities-card {
  border-radius: 8px;
}

.card-header {
  font-weight: 600;
  font-size: 16px;
}

.filter-tabs {
  display: flex;
  gap: 8px;
}

.type-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-icon {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.icon-order {
  background: #67c23a;
}

.icon-vip {
  background: #e6a23c;
}

.icon-token {
  background: #409eff;
}

.icon-login {
  background: #909399;
}

.activity-title {
  font-weight: 500;
}

.text-success {
  color: #67c23a;
}

.time-cell {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.empty-state {
  padding: 40px 0;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .activities-view {
    padding: 12px;
  }

  .filter-tabs :deep(.el-radio-button__inner) {
    padding: 8px 12px;
    font-size: 13px;
  }
}
</style>
