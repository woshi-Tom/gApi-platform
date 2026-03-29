<template>
  <div class="forgot-password-page">
    <el-card class="forgot-card">
      <template #header>
        <div class="forgot-header">
          <el-icon class="logo-icon"><Key /></el-icon>
          <h2>忘记密码</h2>
        </div>
      </template>
      
      <div v-if="!codeSent">
        <p class="description">输入您的注册邮箱，我们将发送密码重置链接</p>
        <el-form @submit.prevent="handleSubmit">
          <el-form-item>
            <el-input 
              v-model="form.email" 
              placeholder="请输入邮箱" 
              prefix-icon="Message" 
              size="large"
              @blur="checkEmailFormat"
            />
          </el-form-item>
          
          <el-form-item v-if="form.email && isValidEmail">
            <div class="captcha-wrapper" @click="showCaptcha = true">
              <el-icon class="captcha-icon"><Picture /></el-icon>
              <span class="captcha-text">{{ captchaVerified ? '安全验证已通过' : '点击进行安全验证' }}</span>
              <el-icon v-if="captchaVerified" class="captcha-check"><CircleCheck /></el-icon>
            </div>
          </el-form-item>
          
          <el-form-item>
            <el-button 
              type="primary" 
              :loading="loading" 
              :disabled="!canSubmit"
              @click="handleSubmit"
              size="large"
              style="width:100%"
            >
              发送重置链接
            </el-button>
          </el-form-item>
        </el-form>
      </div>
      
      <div v-else class="success-content">
        <el-icon class="success-icon"><CircleCheck /></el-icon>
        <h3>重置链接已发送</h3>
        <p>我们已发送密码重置链接到 <strong>{{ form.email }}</strong></p>
        <p class="hint">请查收邮件并点击链接重置密码，链接有效期为1小时</p>
        <el-button type="primary" @click="$router.push('/login')" size="large" style="width:100%; margin-top: 20px">
          返回登录
        </el-button>
      </div>
      
      <div class="back-link">
        <router-link to="/login">返回登录</router-link>
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
import { ElMessage } from 'element-plus'
import { Key, Message, Picture, CircleCheck } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'
import request from '@/api/request'

const router = useRouter()

const loading = ref(false)
const showCaptcha = ref(false)
const captchaVerified = ref(false)
const captchaRef = ref()
const codeSent = ref(false)
const isValidEmail = ref(false)

const form = reactive({ 
  email: ''
})

const canSubmit = computed(() => {
  return isValidEmail.value && captchaVerified.value
})

function checkEmailFormat() {
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
  isValidEmail.value = emailRe.test(form.email)
}

function onCaptchaSuccess() {
  captchaVerified.value = true
}

async function handleSubmit() {
  if (!isValidEmail.value) {
    ElMessage.warning('请输入有效的邮箱')
    return
  }
  if (!captchaVerified.value) {
    ElMessage.warning('请先完成安全验证')
    showCaptcha.value = true
    return
  }
  
  loading.value = true
  try {
    await request.post('/auth/forgot-password', {
      email: form.email,
      captcha_token: 'verified'
    })
    codeSent.value = true
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '发送失败')
    captchaVerified.value = false
    captchaRef.value?.reset()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.forgot-password-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.forgot-card {
  width: 420px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.forgot-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.forgot-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.logo-icon {
  font-size: 36px;
  color: #409eff;
}

.description {
  color: #606266;
  text-align: center;
  margin-bottom: 20px;
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

.success-content {
  text-align: center;
  padding: 20px 0;
}

.success-icon {
  font-size: 64px;
  color: #67c23a;
  margin-bottom: 16px;
}

.success-content h3 {
  margin: 0 0 16px 0;
  color: #303133;
}

.success-content p {
  color: #606266;
  margin: 8px 0;
}

.success-content strong {
  color: #409eff;
}

.hint {
  color: #909399 !important;
  font-size: 13px;
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
