<template>
  <div class="register-page">
    <el-card class="register-card">
      <template #header>
        <div class="register-header">
          <el-icon class="logo-icon"><Monitor /></el-icon>
          <h2>用户注册</h2>
        </div>
      </template>
      <el-form @submit.prevent="handleRegister">
        <el-form-item>
          <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.email" placeholder="请输入邮箱" prefix-icon="Message" size="large" @blur="checkEmailFormat" />
        </el-form-item>
        <el-form-item v-if="form.email && isValidEmail">
          <div class="captcha-wrapper" @click="showCaptcha = true">
            <el-icon class="captcha-icon"><Picture /></el-icon>
            <span class="captcha-text">{{ captchaVerified ? '安全验证已通过' : '点击进行安全验证' }}</span>
            <el-icon v-if="captchaVerified" class="captcha-check"><CircleCheck /></el-icon>
          </div>
        </el-form-item>
        <el-form-item v-if="form.email && isValidEmail && captchaVerified">
          <el-input v-model="form.code" placeholder="请输入邮箱验证码" prefix-icon="Key" size="large" maxlength="6">
            <template #append>
              <el-button 
                @click="sendCode" 
                :disabled="countdown > 0 || sendingCode"
                :loading="sendingCode"
              >
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="请输入密码（至少8位）" prefix-icon="Lock" size="large" show-password />
        </el-form-item>
        <el-form-item>
          <el-button 
            type="primary" 
            :loading="loading" 
            :disabled="!canSubmit" 
            @click="handleRegister" 
            size="large" 
            style="width:100%"
          >
            注 册
          </el-button>
        </el-form-item>
      </el-form>
      <div class="register-footer">
        已有账号？<router-link to="/login">立即登录</router-link>
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
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { User, Message, Lock, Key, Monitor, Picture, CircleCheck } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'
import request from '@/api/request'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const showCaptcha = ref(false)
const captchaVerified = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
const captchaRef = ref()
const isValidEmail = ref(false)

const form = reactive({ 
  username: '', 
  email: '', 
  password: '',
  code: ''
})

const canSubmit = computed(() => {
  return form.username && 
         isValidEmail.value && 
         form.code.length === 6 && 
         form.password.length >= 8
})

function checkEmailFormat() {
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
  isValidEmail.value = emailRe.test(form.email)
  if (!isValidEmail.value && form.email) {
    ElMessage.warning('请输入有效的邮箱格式')
  }
}

function onCaptchaSuccess() {
  captchaVerified.value = true
}

async function sendCode() {
  if (!form.email || !isValidEmail.value) {
    ElMessage.warning('请输入有效的邮箱')
    return
  }
  
  sendingCode.value = true
  try {
    await request.post('/email/send-code', {
      email: form.email,
      captcha_token: 'verified'
    })
    ElMessage.success('验证码已发送到您的邮箱')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '发送失败')
    captchaVerified.value = false
    captchaRef.value?.reset()
  } finally {
    sendingCode.value = false
  }
}

async function handleRegister() {
  if (!form.username || !form.email || !form.password) { 
    ElMessage.warning('请填写所有字段'); 
    return 
  }
  if (!isValidEmail.value) {
    ElMessage.warning('请输入有效的邮箱')
    return
  }
  if (form.password.length < 8) { 
    ElMessage.warning('密码至少8位'); 
    return 
  }
  if (form.code.length !== 6) {
    ElMessage.warning('请输入6位验证码')
    return
  }
  
  loading.value = true
  try {
    await request.post('/email/verify-code', {
      email: form.email,
      code: form.code
    })
    
    await authStore.register(form.username, form.email, form.password)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (e: any) { 
    ElMessage.error(e.response?.data?.error?.message || '注册失败')
    captchaVerified.value = false
    captchaRef.value?.reset()
  } finally { loading.value = false }
}
</script>

<style scoped>
.register-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.register-card {
  width: 420px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.register-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.register-header h2 {
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

.register-footer {
  text-align: center;
  color: #909399;
  margin-top: 16px;
}

.register-footer a {
  color: #409eff;
  text-decoration: none;
}

.register-footer a:hover {
  text-decoration: underline;
}
</style>
