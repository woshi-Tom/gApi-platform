<template>
  <div class="profile-page">
    <el-row :gutter="20">
      <!-- User Info Card -->
      <el-col :span="8">
        <el-card class="user-card">
          <div class="user-header">
            <el-avatar :size="80" class="user-avatar">
              {{ user?.username?.[0]?.toUpperCase() || 'U' }}
            </el-avatar>
            <h3 class="user-name">{{ user?.username || '用户' }}</h3>
            <el-tag :type="computeIsVip(user?.level) ? 'warning' : 'info'" size="small">
              {{ computeIsVip(user?.level) ? 'VIP会员' : '普通用户' }}
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
              <div class="stat-value">{{ formatDate(user?.created_at) }}</div>
              <div class="stat-label">注册时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <!-- Settings -->
      <el-col :span="16">
        <el-card class="settings-card">
          <template #header>
            <span>账户设置</span>
          </template>
          
          <el-tabs>
            <!-- Basic Info -->
            <el-tab-pane label="基本信息">
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
                  <el-button type="primary" @click="saveBasic">保存修改</el-button>
                </el-form-item>
              </el-form>
            </el-tab-pane>
            
            <!-- Change Password -->
            <el-tab-pane label="修改密码">
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
            <el-tab-pane label="安全设置">
              <div class="security-items">
                <div class="security-item">
                  <div class="security-info">
                    <h4>登录密码</h4>
                    <p>已设置，建议定期更换</p>
                  </div>
                  <el-button text type="primary" @click="activeTab = '1'">修改</el-button>
                </div>
                
                <el-divider />
                
                <div class="security-item">
                  <div class="security-info">
                    <h4>两步验证</h4>
                    <p>未开启，建议开启以提高安全性</p>
                  </div>
                  <el-button text type="primary" disabled>即将上线</el-button>
                </div>
                
                <el-divider />
                
                <div class="security-item">
                  <div class="security-info">
                    <h4>登录历史</h4>
                    <p>查看最近的登录记录</p>
                  </div>
                  <el-button text type="primary">查看</el-button>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import request from '@/api/request'
import { ElMessage } from 'element-plus'

const authStore = useAuthStore()

interface User {
  id: number
  username: string
  email: string
  phone?: string
  level: string
  is_vip: boolean
  remain_quota: number
  token_count: number
  created_at: string
}

const user = ref<User | null>(null)

// Compute is_vip from level field
function computeIsVip(level: string | undefined): boolean {
  if (!level) return false
  return level !== 'free' && level.startsWith('vip')
}
const activeTab = ref('0')
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

function formatQuota(n: number | undefined): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

async function saveBasic() {
  try {
    await request.put('/user/profile', { phone: basicForm.phone })
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '保存失败')
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
  await authStore.fetchProfile()
  user.value = authStore.user as User
  basicForm.username = user.value?.username || ''
  basicForm.email = user.value?.email || ''
  basicForm.phone = user.value?.phone || ''
})
</script>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* User Card */
.user-card {
  text-align: center;
}

.user-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
  font-size: 32px;
  font-weight: 600;
}

.user-name {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
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
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

/* Settings Card */
.settings-card :deep(.el-card__header) {
  font-weight: 500;
}

.settings-card :deep(.el-tabs__item) {
  font-weight: 500;
}

.settings-form {
  max-width: 500px;
  padding-top: 10px;
}

/* Security Items */
.security-items {
  padding: 10px 0;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.security-info h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
}

.security-info p {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* Responsive */
@media (max-width: 768px) {
  .profile-page :deep(.el-col) {
    width: 100%;
  }
}
</style>
