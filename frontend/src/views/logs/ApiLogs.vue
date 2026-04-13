<template>
  <div class="api-logs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>API 调用记录</span>
        </div>
      </template>

      <el-table :data="logs" v-loading="loading" stripe>
        <el-table-column label="时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="接口" prop="endpoint" width="200" />
        <el-table-column label="模型" prop="model" width="150" />
        <el-table-column label="Token用量" width="150">
          <template #default="{ row }">
            {{ row.total_tokens?.toLocaleString() || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="响应时间" width="100">
          <template #default="{ row }">
            {{ row.response_time }}ms
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status_code < 400 ? 'success' : 'danger'" size="small">
              {{ row.status_code }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="IP" prop="request_ip" width="130" />
        <el-table-column label="错误信息" min-width="200">
          <template #default="{ row }">
            <span v-if="row.error_message" class="error-text">{{ row.error_message }}</span>
            <span v-else class="success-text">-</span>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="load"
          @size-change="load"
        />
      </div>

      <el-empty v-if="!loading && !logs.length" description="暂无API调用记录" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface APILog {
  id: number
  endpoint: string
  method: string
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  status_code: number
  response_time: number
  error_message: string
  request_ip: string
  created_at: string
}

const logs = ref<APILog[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

async function load() {
  loading.value = true
  try {
    const res = await request.get('/logs', {
      params: { page: page.value, page_size: pageSize.value }
    })
    logs.value = res.data.data || []
    if (res.data.pagination) {
      total.value = res.data.pagination.total
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.api-logs {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.error-text {
  color: var(--el-color-danger);
  font-size: 13px;
}

.success-text {
  color: var(--el-text-color-placeholder);
}
</style>
