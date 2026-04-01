<template>
  <div class="admin-users">
    <div class="page-header">
      <h2>用户管理</h2>
      <div class="header-actions">
        <el-select v-model="filters.level" clearable placeholder="用户等级" style="width:120px" @change="load">
          <el-option label="全部" value="" />
          <el-option label="免费" value="free" />
          <el-option label="VIP" value="vip" />
          <el-option label="企业版" value="enterprise" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="用户状态" style="width:120px" @change="load">
          <el-option label="全部" value="" />
          <el-option label="正常" value="active" />
          <el-option label="禁用" value="disabled" />
        </el-select>
        <el-input 
          v-model="filters.keyword" 
          placeholder="搜索用户名或邮箱" 
          prefix-icon="Search" 
          clearable 
          style="width: 240px"
          @change="load"
        />
      </div>
    </div>

    <el-card class="users-card">
      <el-table :data="users" v-loading="ld" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" width="140">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="28">{{ row.username?.[0]?.toUpperCase() }}</el-avatar>
              <span>{{ row.username }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="levelType(row.level)" size="small">
              {{ levelName(row.level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="永久配额" width="120">
          <template #default="{ row }">
            <span class="quota-value">{{ formatQuota(row.remain_quota) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="VIP附赠配额" width="120">
          <template #default="{ row }">
            <span v-if="row.vip_quota > 0" class="quota-value vip">{{ formatQuota(row.vip_quota) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="VIP到期" width="140">
          <template #default="{ row }">
            <template v-if="row.level === 'vip' || row.level === 'enterprise'">
              <el-tag v-if="isVIPActive(row)" type="success" size="small" effect="plain">
                {{ formatDate(row.vip_expired_at) }}
              </el-tag>
              <el-tag v-else type="info" size="small" effect="plain">
                已过期
              </el-tag>
            </template>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="已使用" width="100">
          <template #default="{ row }">{{ formatQuota(row.used_quota) }}</template>
        </el-table-column>
        <el-table-column label="注册时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="edit(row)">编辑</el-button>
            <el-button 
              size="small" 
              link 
              :type="row.status === 'active' ? 'danger' : 'success'"
              @click="toggle(row)"
            >
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="load"
          @current-change="load"
        />
      </div>
    </el-card>

    <el-dialog v-model="dlg" title="编辑用户" width="550px">
      <el-form :model="ef" label-width="100px">
        <el-descriptions :column="2" border style="margin-bottom:20px">
          <el-descriptions-item label="用户ID">{{ ef.id }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ ef.username }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ ef.email }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ ef.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ formatDate(ef.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="最后登录">{{ ef.last_login_at ? formatDate(ef.last_login_at) : '-' }}</el-descriptions-item>
        </el-descriptions>
        
        <el-form-item label="用户等级">
          <el-select v-model="ef.level" style="width: 100%">
            <el-option label="免费用户" value="free" />
            <el-option label="VIP青铜" value="vip_bronze" />
            <el-option label="VIP白银" value="vip_silver" />
            <el-option label="VIP黄金" value="vip_gold" />
            <el-option label="企业版" value="enterprise" />
          </el-select>
        </el-form-item>
        <el-form-item label="VIP到期时间">
          <el-date-picker 
            v-model="ef.vip_expired_at" 
            type="datetime" 
            placeholder="不设置则留空"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="配额调整">
          <el-input-number 
            v-model="ef.quota_adjust" 
            :step="100000"
            style="width: 100%" 
          />
          <div class="form-tip">正数增加，负数减少。当前永久配额: {{ formatQuota(ef.remain_quota) }}</div>
        </el-form-item>
        <el-form-item label="VIP配额调整">
          <el-input-number 
            v-model="ef.vip_quota_adjust" 
            :step="100000"
            style="width: 100%" 
          />
          <div class="form-tip">正数增加，负数减少。当前VIP配额: {{ formatQuota(ef.vip_quota) }}</div>
        </el-form-item>
        <el-form-item label="用户状态">
          <el-select v-model="ef.status" style="width: 100%">
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item label="禁用原因" v-if="ef.status === 'disabled'">
          <el-input v-model="ef.disabled_reason" placeholder="请输入禁用原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dlg = false">取消</el-button>
        <el-button type="primary" @click="save" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminUserApi } from '@/api/user'
import { ElMessage, ElMessageBox } from 'element-plus'

interface User {
  id: number
  username: string
  email: string
  phone?: string
  level: string
  status: string
  remain_quota: number
  vip_quota: number
  used_quota: number
  vip_expired_at?: string
  last_login_at?: string
  created_at: string
  disabled_reason?: string
}

const users = ref<User[]>([])
const ld = ref(false)
const dlg = ref(false)
const saving = ref(false)

const filters = reactive({
  level: '',
  status: '',
  keyword: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const ef = reactive({
  id: 0,
  username: '',
  email: '',
  phone: '',
  level: 'free',
  status: 'active',
  remain_quota: 0,
  vip_quota: 0,
  vip_expired_at: null as Date | null,
  last_login_at: '',
  created_at: '',
  disabled_reason: '',
  quota_adjust: 0,
  vip_quota_adjust: 0,
})

const levelType = (level: string) => {
  switch (level) {
    case 'vip_gold': return 'danger'
    case 'vip_silver': return 'warning'
    case 'vip_bronze': return 'info'
    case 'enterprise': return 'danger'
    case 'free': return 'info'
    default: return 'info'
  }
}

const levelName = (level: string) => {
  switch (level) {
    case 'vip_bronze': return 'VIP青铜'
    case 'vip_silver': return 'VIP白银'
    case 'vip_gold': return 'VIP黄金'
    case 'enterprise': return '企业版'
    case 'free': return '免费用户'
    default: return level || '免费用户'
  }
}

function formatQuota(n: number | undefined): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function isVIPActive(row: User): boolean {
  if (!row.vip_expired_at) return false
  return new Date(row.vip_expired_at) > new Date()
}

async function load() {
  ld.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    if (filters.level) params.level = filters.level
    if (filters.status) params.status = filters.status
    if (filters.keyword) params.keyword = filters.keyword
    
    const res = await adminUserApi.listUsers(params)
    if (res.data.data) {
      users.value = res.data.data.list || res.data.data
      pagination.total = res.data.data.pagination?.total || users.value.length
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '加载失败')
  } finally {
    ld.value = false
  }
}

function edit(u: User) {
  ef.id = u.id
  ef.username = u.username
  ef.email = u.email
  ef.phone = u.phone || ''
  ef.level = u.level
  ef.status = u.status
  ef.remain_quota = u.remain_quota
  ef.vip_quota = u.vip_quota
  ef.vip_expired_at = u.vip_expired_at ? new Date(u.vip_expired_at) : null
  ef.last_login_at = u.last_login_at || ''
  ef.created_at = u.created_at
  ef.disabled_reason = u.disabled_reason || ''
  ef.quota_adjust = 0
  ef.vip_quota_adjust = 0
  dlg.value = true
}

async function save() {
  saving.value = true
  try {
    const data: any = {
      level: ef.level,
      status: ef.status,
    }
    if (ef.quota_adjust !== 0) data.quota_adjust = ef.quota_adjust
    if (ef.vip_quota_adjust !== 0) data.vip_quota_adjust = ef.vip_quota_adjust
    if (ef.vip_expired_at) data.vip_expired_at = ef.vip_expired_at.toISOString()
    if (ef.status === 'disabled' && ef.disabled_reason) data.disabled_reason = ef.disabled_reason
    
    await adminUserApi.updateUser(ef.id, data)
    ElMessage.success('用户信息已更新')
    dlg.value = false
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '更新失败')
  } finally {
    saving.value = false
  }
}

async function toggle(u: User) {
  const newStatus = u.status === 'active' ? 'disabled' : 'active'
  const action = newStatus === 'disabled' ? '禁用' : '启用'
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 ${u.username} 吗？`,
      '确认操作',
      { type: 'warning' }
    )
    
    await adminUserApi.updateUser(u.id, { status: newStatus })
    ElMessage.success(`已${action}`)
    load()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error?.message || '操作失败')
    }
  }
}

onMounted(load)
</script>

<style scoped>
.admin-users {
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

.header-actions {
  display: flex;
  gap: 12px;
}

.users-card {
  border-radius: 10px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.quota-value {
  font-weight: 500;
  color: var(--el-color-success);
}

.quota-value.vip {
  color: var(--el-color-warning);
}

.text-muted {
  color: var(--el-text-color-placeholder);
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
