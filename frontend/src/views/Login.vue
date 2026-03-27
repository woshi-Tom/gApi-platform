<template>
  <div class="login-page">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">
          <el-icon class="logo-icon"><Monitor /></el-icon>
          <h2>用户登录</h2>
        </div>
      </template>
      <el-form @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="form.email" placeholder="请输入邮箱" prefix-icon="Message" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="请输入密码" prefix-icon="Lock" size="large" show-password />
        </el-form-item>
        <el-form-item>
          <div class="captcha-wrapper" @click="showCaptcha = true">
            <el-icon class="captcha-icon"><Picture /></el-icon>
            <span class="captcha-text">{{ captchaVerified ? '安全验证已通过' : '点击进行安全验证' }}</span>
            <el-icon v-if="captchaVerified" class="captcha-check"><CircleCheck /></el-icon>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" :disabled="!captchaVerified" @click="handleLogin" size="large" style="width:100%">登 录</el-button>
        </el-form-item>
      </el-form>
      <div class="login-footer">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </div>
    </el-card>
    
    <SlideCaptcha 
      v-model:visible="showCaptcha" 
      @success="onCaptchaSuccess" 
      ref="captchaRef"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { Message, Lock, Monitor, Picture, CircleCheck } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const showCaptcha = ref(false)
const captchaVerified = ref(false)
const captchaRef = ref()

const form = reactive({ email: '', password: '' })

function onCaptchaSuccess() {
  captchaVerified.value = true
}

async function handleLogin() {
  if (!captchaVerified.value) {
    ElMessage.warning('请先完成安全验证')
    showCaptcha.value = true
    return
  }
  if (!form.email || !form.password) { 
    ElMessage.warning('请填写邮箱和密码'); 
    return 
  }
  loading.value = true
  try {
    await authStore.login(form.email, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e: any) { 
    ElMessage.error(e.response?.data?.error?.message || '登录失败')
    captchaVerified.value = false
    captchaRef.value?.reset()
  } finally { loading.value = false }
}
</script>

<style scoped>
.login-page {
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

.captcha-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
  background: #f5f7fa;
}
.captcha-wrapper:hover {
  border-color: #409eff;
  background: #fff;
}
.captcha-icon {
  font-size: 20px;
  color: #409eff;
}
.captcha-text {
  flex: 1;
  font-size: 14px;
  color: #909399;
}
.captcha-check {
  color: #67c23a;
  font-size: 18px;
}

.login-footer {
  text-align: center;
  color: #909399;
  margin-top: 16px;
}

.login-footer a {
  color: #409eff;
  text-decoration: none;
}

.login-footer a:hover {
  text-decoration: underline;
}
</style>
