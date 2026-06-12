<template>
  <div class="profile-page">
    <div class="profile-grid">
      <!-- User Info Card -->
      <el-card v-loading="pageLoading" class="user-card" shadow="hover">
        <div class="user-header">
          <el-avatar :size="80" class="user-avatar">
            {{ user?.username?.[0]?.toUpperCase() || 'U' }}
          </el-avatar>
          <h3 class="user-name">{{ user?.username || '用户' }}</h3>
          <el-tag :type="getStatusType(user?.account_status)" size="small">
            {{ getStatusLabel(user?.account_status) }}
          </el-tag>
        </div>

        <el-divider style="margin: 20px 0" />

        <div class="user-stats">
          <div class="stat-item">
            <div class="stat-value">{{ formatQuota(user?.remain_quota) }}</div>
            <div class="stat-label">剩余配额</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ user?.token_count || 0 }}</div>
            <div class="stat-label">密钥数量</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatRegisterDate(user?.created_at) }}</div>
            <div class="stat-label">注册时间</div>
          </div>
        </div>
      </el-card>

      <!-- Settings -->
      <el-card class="settings-card" shadow="hover">
        <template #header>
          <span>账户设置</span>
        </template>

        <el-tabs v-model="activeTab">
          <!-- Basic Info -->
          <el-tab-pane label="基本信息" name="basic">
            <el-form :model="basicForm" label-width="100px" class="settings-form">
              <el-form-item label="用户名">
                <el-input v-model="basicForm.username" disabled />
              </el-form-item>
              <el-form-item label="邮箱">
                <el-input v-model="basicForm.email" disabled />
              </el-form-item>
              <el-form-item label="手机号">
                <el-input v-model="basicForm.phone" placeholder="请输入手机号" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="saving" @click="saveBasic">保存修改</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- Change Password -->
          <el-tab-pane label="修改密码" name="password">
            <el-form :model="pwdForm" label-width="100px" class="settings-form">
              <el-form-item label="当前密码">
                <el-input
                  v-model="pwdForm.old"
                  type="password"
                  show-password
                  placeholder="请输入当前密码"
                />
              </el-form-item>
              <el-form-item label="新密码">
                <el-input
                  v-model="pwdForm.new"
                  type="password"
                  show-password
                  placeholder="至少8位，包含字母和数字"
                />
              </el-form-item>
              <el-form-item label="确认密码">
                <el-input
                  v-model="pwdForm.confirm"
                  type="password"
                  show-password
                  placeholder="请再次输入新密码"
                />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="changePassword" :loading="pwdLoading">
                  修改密码
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <!-- Security -->
          <el-tab-pane label="安全设置" name="security">
            <div class="security-items">
              <div class="security-item">
                <div class="security-info">
                  <h4>登录密码</h4>
                  <p>已设置，建议定期更换</p>
                </div>
                <el-button text type="primary" @click="activeTab = 'password'">修改</el-button>
              </div>

              <el-divider />

              <div class="security-item">
                <div class="security-info">
                  <h4>两步验证</h4>
                  <p>未开启，建议开启以提高安全性</p>
                </div>
                <el-tooltip content="功能开发中" placement="top">
                  <el-button text type="info" disabled>即将上线</el-button>
                </el-tooltip>
              </div>

              <el-divider />

              <div class="security-item">
                <div class="security-info">
                  <h4>登录历史</h4>
                  <p>查看最近的登录记录</p>
                </div>
                <el-tooltip content="功能开发中" placement="top">
                  <el-button text type="info" disabled>查看</el-button>
                </el-tooltip>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import request from '@/api/request'
import { ElMessage } from 'element-plus'
import { formatQuota, formatDate } from '@/composables/useFormat'

const authStore = useAuthStore()

interface User {
  id: number
  username: string
  email: string
  phone?: string
  level: string
  is_vip: boolean
  account_status: 'vip' | 'vip_expired' | 'recharge' | 'free' | 'unknown'
  v_ip_expired_at?: string
  free_quota?: number
  v_ip_quota?: number
  remain_quota: number
  token_count: number
  created_at: string
}

const user = ref<User | null>(null)
const pageLoading = ref(true)
const saving = ref(false)

function getStatusLabel(status: string | undefined): string {
  switch (status) {
    case 'vip': return 'VIP会员'
    case 'vip_expired': return 'VIP已过期'
    case 'recharge': return '充值用户'
    case 'free': return '普通用户'
    default: return '普通用户'
  }
}

function getStatusType(status: string | undefined): string {
  switch (status) {
    case 'vip': return 'warning'
    case 'vip_expired': return 'info'
    case 'recharge': return 'success'
    default: return 'info'
  }
}
const activeTab = ref('basic')
const pwdLoading = ref(false)

const basicForm = reactive({
  username: '',
  email: '',
  phone: ''
})

const pwdForm = reactive({
  old: '',
  new: '',
  confirm: ''
})

function formatRegisterDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

async function saveBasic() {
  saving.value = true
  try {
    await request.put('/user/profile', { phone: basicForm.phone })
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function changePassword() {
  if (!pwdForm.old || !pwdForm.new) {
    ElMessage.warning('请填写所有字段')
    return
  }
  if (pwdForm.new.length < 8) {
    ElMessage.warning('新密码至少8位')
    return
  }
  if (pwdForm.new !== pwdForm.confirm) {
    ElMessage.warning('两次密码不一致')
    return
  }
  
  pwdLoading.value = true
  try {
    await request.post('/user/change-password', {
      old_password: pwdForm.old,
      new_password: pwdForm.new
    })
    ElMessage.success('密码修改成功')
    pwdForm.old = ''
    pwdForm.new = ''
    pwdForm.confirm = ''
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '修改失败')
  } finally {
    pwdLoading.value = false
  }
}

onMounted(async () => {
  try {
    await authStore.fetchProfile()
    user.value = authStore.user as User
    basicForm.username = user.value?.username || ''
    basicForm.email = user.value?.email || ''
    basicForm.phone = user.value?.phone || ''
  } catch {
    ElMessage.error('加载用户信息失败')
  } finally {
    pageLoading.value = false
  }
})
</script>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
}

.profile-grid {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: var(--spacing-lg);
}

/* User Card */
.user-card {
  text-align: center;
  border-radius: var(--radius-xl);
}

.user-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
}

.user-avatar {
  background: var(--gradient-primary);
  font-size: 32px;
  font-weight: var(--font-weight-semibold);
}

.user-name {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
}

.user-stats {
  display: flex;
  justify-content: space-around;
  text-align: center;
}

.stat-item {
  flex: 1;
}

.stat-value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-xs);
}

/* Settings Card */
.settings-card {
  border-radius: var(--radius-xl);
}

.settings-card :deep(.el-card__header) {
  font-weight: var(--font-weight-medium);
}

.settings-card :deep(.el-tabs__item) {
  font-weight: var(--font-weight-medium);
}

.settings-form {
  max-width: 500px;
  padding-top: var(--spacing-md);
}

/* Security Items */
.security-items {
  padding: var(--spacing-md) 0;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm) 0;
}

.security-info h4 {
  margin: 0;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
}

.security-info p {
  margin: var(--spacing-xs) 0 0;
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

/* Responsive */
@media (max-width: 900px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}
</style>
