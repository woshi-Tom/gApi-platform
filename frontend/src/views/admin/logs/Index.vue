<template>
  <div class="audit-logs">
    <div class="page-header">
      <h2>操作日志</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon> 导出
        </el-button>
      </div>
    </div>

    <el-card class="logs-card">
      <div class="filters">
        <el-select v-model="filters.log_type" placeholder="日志类型" clearable style="width: 120px" @change="handleFilter">
          <el-option v-for="g in logTypes" :key="g.value" :label="g.label" :value="g.value" />
        </el-select>
        <el-select v-model="filters.action_group" placeholder="操作分组" clearable style="width: 120px" @change="handleFilter">
          <el-option v-for="g in actionGroups" :key="g.value" :label="g.label" :value="g.value" />
        </el-select>
        <el-select v-model="filters.success" placeholder="状态" clearable style="width: 100px" @change="handleFilter">
          <el-option label="成功" :value="true" />
          <el-option label="失败" :value="false" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          @change="handleDateChange"
          style="width: 240px"
        />
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table :data="logs" v-loading="ld" stripe class="logs-table">
        <el-table-column prop="id" label="ID" width="50" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="60">
          <template #default="{ row }">
            <el-tag :type="row.log_type === 'operation' ? 'success' : 'info'" size="small">
              {{ row.log_type === 'operation' ? '操作' : '访问' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分组" width="70">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.action_group }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户" width="80">
          <template #default="{ row }">
            <span v-if="row.username">{{ row.username }}</span>
            <span v-else class="text-muted">系统</span>
          </template>
        </el-table-column>
        <el-table-column prop="request_ip" label="IP" width="100" />
        <el-table-column label="状态" width="55">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" size="small">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="140">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="详情" width="50" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text @click="handleShowDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="日志详情" width="800px" class="log-detail-dialog" v-loading="detailLoading">
      <div class="detail-scroll">
      <el-descriptions :column="2" border v-if="currentLog">
        <el-descriptions-item label="ID">{{ currentLog.id }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="currentLog.log_type === 'operation' ? 'success' : 'info'" size="small">
            {{ currentLog.log_type === 'operation' ? '业务操作' : '访问记录' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作">{{ currentLog.action }}</el-descriptions-item>
        <el-descriptions-item label="分组">{{ currentLog.action_group }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ currentLog.username || '系统' }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">
          <el-tag size="small">{{ currentLog.request_method }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="资源类型">{{ currentLog.resource_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="资源ID">{{ currentLog.resource_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP地址" :span="2">{{ currentLog.request_ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">
          <code class="path-code">{{ currentLog.request_path }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog.success ? 'success' : 'danger'">
            {{ currentLog.success ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="HTTP状态码">{{ currentLog.status_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="响应时间">{{ currentLog.response_time_ms }}ms</el-descriptions-item>
        <el-descriptions-item label="时间">{{ formatDate(currentLog.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="User-Agent" :span="2">
          <div class="text-scroll">{{ currentLog.user_agent }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="错误信息" :span="2" v-if="currentLog.error_message">
          <span class="text-danger">{{ currentLog.error_message }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="请求内容" :span="2" v-if="currentLog.request_body">
          <div class="json-wrapper">{{ formatJson(currentLog.request_body) }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="响应内容" :span="2" v-if="currentLog.response_body">
          <div class="json-wrapper">{{ formatJson(currentLog.response_body) }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="变更前" :span="2" v-if="currentLog.old_value">
          <div class="json-wrapper">{{ formatJson(currentLog.old_value) }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="变更后" :span="2" v-if="currentLog.new_value">
          <div class="json-wrapper">{{ formatJson(currentLog.new_value) }}</div>
        </el-descriptions-item>
      </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { auditLogApi, ACTION_GROUPS, LOG_TYPES } from '@/api/log'
import type { AuditLogBrief, AuditLog, AuditLogQuery } from '@/api/log'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'

const logs = ref<AuditLogBrief[]>([])
const ld = ref(false)
const exporting = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const dateRange = ref<string[]>([])
const detailVisible = ref(false)
const detailLoading = ref(false)
const currentLog = ref<AuditLog | null>(null)

const filters = reactive({
  log_type: '',
  action_group: '',
  success: null as boolean | null
})

const actionGroups = ACTION_GROUPS
const logTypes = LOG_TYPES

function formatDate(dateStr: string): string {
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

function formatJson(str: string): string {
  if (!str) return ''
  try {
    // 递归解析嵌套转义的JSON直到解析成功或无法继续
    let current = str
    let maxDepth = 3
    for (let i = 0; i < maxDepth; i++) {
      try {
        const obj = JSON.parse(current)
        return JSON.stringify(obj, null, 2)
      } catch {
        // 尝试解开一层转义
        if (current.includes('\\"')) {
          current = current.replace(/\\"/g, '"')
        } else if (current.includes('\\\\')) {
          current = current.replace(/\\\\/g, '\\')
        } else {
          break
        }
      }
    }
    return current
  } catch {
    return str
  }
}

async function load() {
  ld.value = true
  try {
    const params: AuditLogQuery = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (filters.log_type) params.log_type = filters.log_type
    if (filters.action_group) params.action_group = filters.action_group
    if (filters.success !== null) params.success = filters.success
    if (dateRange.value?.length === 2) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }

    const res = await auditLogApi.list(params)
    logs.value = res.data.data?.list || []
    total.value = res.data.data?.pagination?.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    ld.value = false
  }
}

async function handleShowDetail(row: AuditLogBrief) {
  detailLoading.value = true
  detailVisible.value = true
  try {
    const res = await auditLogApi.getDetail(row.id)
    currentLog.value = res.data.data
  } catch (e: any) {
    ElMessage.error(e.message || '加载详情失败')
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

function handleFilter() {
  currentPage.value = 1
  load()
}

function handleDateChange() {
  currentPage.value = 1
  load()
}

function handlePageChange() {
  load()
}

function handleSizeChange() {
  currentPage.value = 1
  load()
}

function resetFilters() {
  filters.log_type = ''
  filters.action_group = ''
  filters.success = null
  dateRange.value = []
  handleFilter()
}

async function handleExport() {
  exporting.value = true
  try {
    const params: AuditLogQuery = {}
    if (filters.log_type) params.log_type = filters.log_type
    if (filters.action_group) params.action_group = filters.action_group
    if (filters.success !== null) params.success = filters.success
    if (dateRange.value?.length === 2) {
      params.start_time = dateRange.value[0]
      params.end_time = dateRange.value[1]
    }

    const res = await auditLogApi.export(params)
    const blob = new Blob([res.data], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit_logs_${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e: any) {
    ElMessage.error(e.message || '导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.audit-logs {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.logs-card {
  border-radius: 10px;
}

.filters {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
}

.logs-table {
  margin-top: 16px;
  width: 100% !important;
}

.logs-table :deep(.el-table__inner-wrapper) {
  width: 100% !important;
}

.logs-table :deep(.el-table__header-wrapper),
.logs-table :deep(.el-table__body-wrapper) {
  width: 100% !important;
}

.logs-table :deep(table) {
  width: 100% !important;
}

.logs-table :deep(.el-table__header) {
  width: 100% !important;
}

.logs-table :deep(.el-table__body) {
  width: 100% !important;
}

.logs-table :deep(.el-table__header) th {
  background-color: var(--el-fill-color-light) !important;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
}

.text-muted {
  color: var(--el-text-color-secondary);
}

.text-danger {
  color: var(--el-color-danger);
}

.text-truncate {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.text-scroll {
  display: inline-block;
  white-space: nowrap;
  overflow-x: auto;
  overflow-y: hidden;
  font-size: 12px;
  max-width: 100%;
  scrollbar-width: thin;
}

.json-view {
  background: var(--el-fill-color-light);
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
  max-height: 200px;
  overflow: auto;
  margin: 0;
}

.path-code {
  background: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  word-break: break-all;
}

/* 对话框整体样式 - 固定宽度，禁止横向溢出 */
.log-detail-dialog :deep(.el-dialog) {
  width: 800px;
  max-width: 90vw;
  max-height: 85vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 对话框主体 - 唯一滚动区域 */
.log-detail-dialog :deep(.el-dialog__body) {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0;
}

.detail-scroll {
  padding: 20px;
  max-height: 70vh;
  overflow-y: auto;
  box-sizing: border-box;
}

/* 单元格固定宽度 */
.log-detail-dialog :deep(.el-descriptions__cell) {
  max-width: 150px;
  word-break: break-word;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* JSON内容显示 - 跟随父容器滚动 */
.json-wrapper {
  background: var(--el-fill-color-light);
  border-radius: 4px;
  padding: 8px;
  font-size: 11px;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}
</style>
