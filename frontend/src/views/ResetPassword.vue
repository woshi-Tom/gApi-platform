<template>
  <div class="reset-password-page">
    <el-card class="reset-card">
      <template #header>
        <div class="reset-header">
          <el-icon class="logo-icon"><Lock /></el-icon>
          <h2>重置密码</h2>
        </div>
      </template>
      
      <div v-if="loading">
        <el-skeleton :rows="3" animated />
      </div>
      
      <div v-else-if="tokenValid">
        <p class="description">为 <strong>{{ email }}</strong> 设置新密码</p>
        <el-form @submit.prevent="handleSubmit">
          <el-form-item label="新密码">
            <el-input 
              v-model="form.password" 
              type="password" 
              show-password
              placeholder="至少8位"
              size="large"
            />
            <div v-if="form.password" class="password-strength">
              <span>密码强度：</span>
              <el-tag :type="passwordStrengthType" size="small">{{ passwordStrengthText }}</el-tag>
            </div>
          </el-form-item>
          
          <el-form-item label="确认密码">
            <el-input 
              v-model="form.confirmPassword" 
              type="password" 
              show-password
              placeholder="再次输入新密码"
              size="large"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button 
              type="primary" 
              :loading="submitting"
              :disabled="!canSubmit"
              @click="handleSubmit"
              size="large"
              style="width:100%"
            >
              重置密码
            </el-button>
          </el-form-item>
        </el-form>
      </div>
      
      <div v-else class="error-content">
        <el-icon class="error-icon"><CircleClose /></el-icon>
        <h3>链接已失效</h3>
        <p>重置链接无效或已过期，请重新申请</p>
        <el-button type="primary" @click="$router.push('/forgot-password')" size="large" style="width:100%; margin-top: 20px">
          重新申请
        </el-button>
      </div>
      
      <div class="back-link">
        <router-link to="/login">返回登录</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, CircleClose } from '@element-plus/icons-vue'
import request from '@/api/request'

const router = useRouter()
const route = useRoute()

const loading = ref(true)
const submitting = ref(false)
const tokenValid = ref(false)
const email = ref('')
const error = ref('')

const form = reactive({ 
  password: '',
  confirmPassword: ''
})

const canSubmit = computed(() => {
  return form.password.length >= 8 && 
         form.password === form.confirmPassword
})

const passwordStrength = computed(() => {
  const pwd = form.password
  if (!pwd) return 0
  let score = 0
  if (pwd.length >= 6) score++
  if (pwd.length >= 8) score++
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++
  return score
})

const passwordStrengthType = computed(() => {
  const s = passwordStrength.value
  if (s <= 2) return 'danger'
  if (s <= 3) return 'warning'
  return 'success'
})

const passwordStrengthText = computed(() => {
  const s = passwordStrength.value
  if (s <= 1) return '弱'
  if (s <= 2) return '中等'
  if (s <= 3) return '良好'
  return '强'
})

onMounted(async () => {
  const token = route.query.token as string
  if (!token) {
    error.value = 'Missing token'
    tokenValid.value = false
    loading.value = false
    return
  }
  
  try {
    const res = await request.get('/auth/reset-password', {
      params: { token }
    })
    if (res.data?.success) {
      email.value = res.data.data.email
      tokenValid.value = true
    } else {
      error.value = res.data?.error?.message || 'Invalid token'
      tokenValid.value = false
    }
  } catch (e: any) {
    error.value = e.response?.data?.error?.message || 'Link expired'
    tokenValid.value = false
  } finally {
    loading.value = false
  }
})

async function handleSubmit() {
  if (!canSubmit.value) {
    ElMessage.warning('请填写所有字段')
    return
  }
  if (form.password !== form.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  
  const token = route.query.token as string
  submitting.value = true
  try {
    await request.post('/auth/reset-password', {
      token,
      password: form.password,
      confirm_password: form.confirmPassword
    })
    ElMessage.success('密码重置成功')
    router.push('/login')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '重置失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.reset-password-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.reset-card {
  width: 420px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.reset-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.reset-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.logo-icon {
  font-size: 36px;
  color: #67c23a;
}

.description {
  color: #606266;
  text-align: center;
  margin-bottom: 20px;
}

.description strong {
  color: #409eff;
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #606266;
  margin-top: 8px;
}

.error-content {
  text-align: center;
  padding: 20px 0;
}

.error-icon {
  font-size: 64px;
  color: #f56c6c;
  margin-bottom: 16px;
}

.error-content h3 {
  margin: 0 0 16px 0;
  color: #303133;
}

.error-content p {
  color: #606266;
  margin: 8px 0;
}

.back-link {
  text-align: center;
  margin-top: 16px;
}

.back-link a {
  color: #409eff;
  text-decoration: none;
}

.back-link a:hover {
  text-decoration: underline;
}
</style>
