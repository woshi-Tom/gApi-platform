<template>
  <div class="admin-login">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">
          <el-icon class="logo-icon"><Setting /></el-icon>
          <h2>管理后台</h2>
        </div>
      </template>
      <el-form @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="form.username" placeholder="请输入管理员用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="请输入密码" prefix-icon="Lock" size="large" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleLogin" size="large" style="width:100%">登 录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { User, Lock, Setting } from '@element-plus/icons-vue'
import { adminAPI } from '@/api/request'

const router = useRouter()
const adminStore = useAdminStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

async function handleLogin() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  
  loading.value = true
  try {
    const payload = {
      username: form.username,
      password: form.password
    }
    console.log('Admin login request:', payload)
    const { data } = await adminAPI.post('/login', payload)
    
    if (data.success) {
      localStorage.setItem('admin_token', data.data.token)
      localStorage.setItem('admin_user', JSON.stringify(data.data))
      adminStore.setToken(data.data.token)
      ElMessage.success('登录成功')
      router.push('/dashboard')
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.admin-login {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.login-card {
  width: 400px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.login-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.logo-icon {
  font-size: 36px;
  color: #409eff;
}
</style>
