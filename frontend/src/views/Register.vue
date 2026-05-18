<template>
  <div class="auth-page">
    <!-- 背景装饰：渐变光晕 -->
    <div class="bg-decoration">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
      <div class="orb orb-3"></div>
      <div class="grid-lines"></div>
    </div>

    <!-- 导航栏 -->
    <nav class="navbar">
      <router-link to="/login" class="nav-logo animate-fade-up" style="animation-delay: 0ms">
        gAPI Platform
      </router-link>
      <router-link to="/login" class="nav-link-btn animate-fade-up" style="animation-delay: 150ms">
        已有账号？登录
      </router-link>
    </nav>

    <!-- 主内容区 -->
    <main class="auth-content">
      <div class="glass-card animate-fade-up" style="animation-delay: 300ms">
        <div class="card-inner">
          <div class="card-header">
            <h2 class="animate-fade-up" style="animation-delay: 400ms">创建账号</h2>
            <p class="card-desc animate-fade-up" style="animation-delay: 450ms">注册您的 gAPI Platform 账户</p>
          </div>

          <el-form @submit.prevent="handleRegister" class="auth-form animate-fade-up" style="animation-delay: 500ms">
            <el-form-item>
              <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" size="large" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="form.email" placeholder="请输入邮箱" :prefix-icon="Message" size="large" @blur="checkEmailFormat" />
            </el-form-item>
            <el-form-item v-if="form.email && isValidEmail">
              <div class="captcha-wrapper" :class="{ 'captcha-verified': captchaVerified }" @click="showCaptcha = true">
                <el-icon class="captcha-icon"><Picture /></el-icon>
                <span class="captcha-text">{{ captchaVerified ? '安全验证已通过' : '点击进行安全验证' }}</span>
                <el-icon v-if="captchaVerified" class="captcha-check"><CircleCheck /></el-icon>
              </div>
            </el-form-item>
            <el-form-item v-if="form.email && isValidEmail && captchaVerified">
              <el-input v-model="form.code" placeholder="请输入邮箱验证码" :prefix-icon="Key" size="large" maxlength="6">
                <template #append>
                  <el-button @click="sendCode" :disabled="countdown > 0 || sendingCode" :loading="sendingCode">
                    {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
                  </el-button>
                </template>
              </el-input>
            </el-form-item>
            <el-form-item>
              <el-input v-model="form.password" type="password" placeholder="请输入密码（至少8位）" :prefix-icon="Lock" size="large" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" native-type="submit" :loading="loading" :disabled="!canSubmit" @click="handleRegister" size="large" class="submit-btn">
                <span v-if="!loading">注 册</span>
                <span v-else>注册中...</span>
              </el-button>
            </el-form-item>
          </el-form>

          <div class="card-footer animate-fade-up" style="animation-delay: 600ms">
            <span>已有账号？</span>
            <router-link to="/login" class="footer-link">立即登录</router-link>
          </div>
        </div>
      </div>
    </main>

    <footer class="page-footer animate-fade-up" style="animation-delay: 700ms">
      © {{ new Date().getFullYear() }} gAPI Platform. All rights reserved.
    </footer>

    <SlideCaptcha v-model:visible="showCaptcha" @success="onCaptchaSuccess" ref="captchaRef" />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { User, Message, Lock, Key, Picture, CircleCheck } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'
import request from '@/api/request'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const showCaptcha = ref(false)
const captchaVerified = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
const countdownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const captchaRef = ref()
const isValidEmail = ref(false)
const captchaToken = ref('')

const form = reactive({ username: '', email: '', password: '', code: '' })

const canSubmit = computed(() => {
  return form.username && isValidEmail.value && form.code.length === 6 && form.password.length >= 8
})

function checkEmailFormat() {
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
  isValidEmail.value = emailRe.test(form.email)
  if (!isValidEmail.value && form.email) {
    ElMessage.warning('请输入有效的邮箱格式')
  }
}

function onCaptchaSuccess(token: string) {
  captchaVerified.value = true
  captchaToken.value = token
}

async function sendCode() {
  if (!form.email || !isValidEmail.value) {
    ElMessage.warning('请输入有效的邮箱')
    return
  }
  sendingCode.value = true
  try {
    const payload: any = { email: form.email }
    if (captchaToken.value) payload.captcha_token = captchaToken.value
    await request.post('/email/send-code', payload)
    ElMessage.success('验证码已发送到您的邮箱')
    countdown.value = 60
    if (countdownTimer.value) clearInterval(countdownTimer.value)
    countdownTimer.value = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        if (countdownTimer.value) { clearInterval(countdownTimer.value); countdownTimer.value = null }
      }
    }, 1000)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '发送失败')
    captchaVerified.value = false
    captchaRef.value?.reset()
  } finally { sendingCode.value = false }
}

async function handleRegister() {
  if (!form.username || !form.email || !form.password) { ElMessage.warning('请填写所有字段'); return }
  if (!isValidEmail.value) { ElMessage.warning('请输入有效的邮箱'); return }
  if (form.password.length < 8) { ElMessage.warning('密码至少8位'); return }
  if (form.code.length !== 6) { ElMessage.warning('请输入6位验证码'); return }
  loading.value = true
  try {
    await request.post('/email/verify-code', { email: form.email, code: form.code })
    await authStore.register(form.username, form.email, form.password)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (e: any) {
    const errorData = e.response?.data?.error
    if (errorData?.code === 'REGISTRATION_CLOSED') { router.push('/register-closed'); return }
    ElMessage.error(errorData?.message || '注册失败')
    captchaVerified.value = false
    captchaRef.value?.reset()
  } finally { loading.value = false }
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
</style>

<style scoped>
/* ===== 页面布局 ===== */
.auth-page {
  position: relative; width: 100%; min-height: 100vh;
  display: flex; flex-direction: column; background: #000;
  font-family: 'Inter', sans-serif; overflow: hidden; color: #fff;
}

/* ===== 背景装饰 ===== */
.bg-decoration { position: fixed; inset: 0; z-index: 0; pointer-events: none; overflow: hidden; }
.orb { position: absolute; border-radius: 50%; filter: blur(120px); opacity: 0.15; }
.orb-1 { width: 600px; height: 600px; background: radial-gradient(circle, #6366f1, transparent 70%); top: -200px; left: -150px; animation: orb-1 25s ease-in-out infinite; }
.orb-2 { width: 500px; height: 500px; background: radial-gradient(circle, #8b5cf6, transparent 70%); bottom: -200px; right: -100px; animation: orb-2 30s ease-in-out infinite; }
.orb-3 { width: 400px; height: 400px; background: radial-gradient(circle, #3b82f6, transparent 70%); top: 40%; left: 50%; animation: orb-3 20s ease-in-out infinite; }
@keyframes orb-1 { 0%, 100% { transform: translate(0, 0) scale(1); } 50% { transform: translate(40px, 30px) scale(1.08); } }
@keyframes orb-2 { 0%, 100% { transform: translate(0, 0) scale(1); } 50% { transform: translate(-30px, -20px) scale(1.05); } }
@keyframes orb-3 { 0%, 100% { transform: translate(-50%, 0) scale(1); } 50% { transform: translate(-50%, 30px) scale(0.95); } }
.grid-lines { position: absolute; inset: 0; background-image: linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px); background-size: 80px 80px; }

/* ===== 动画 ===== */
@keyframes fadeUp { from { opacity: 0; transform: translateY(20px); filter: blur(10px); } to { opacity: 1; transform: translateY(0); filter: blur(0); } }
.animate-fade-up { animation: fadeUp 0.8s ease-out forwards; opacity: 0; }

/* ===== 导航栏 ===== */
.navbar { position: relative; z-index: 50; display: flex; justify-content: space-between; align-items: center; padding: 16px; }
@media (min-width: 768px) { .navbar { padding: 16px 48px; } }
.nav-logo { font-size: 24px; font-weight: 600; letter-spacing: -0.04em; color: #fff; text-decoration: none; transition: opacity 0.2s; }
@media (min-width: 768px) { .nav-logo { font-size: 28px; } }
.nav-logo:hover { opacity: 0.8; }
.nav-link-btn { padding: 8px 20px; border-radius: 8px; font-size: 14px; font-weight: 500; color: rgba(255,255,255,0.8); text-decoration: none; border: 1px solid rgba(255,255,255,0.1); transition: all 0.2s; }
.nav-link-btn:hover { border-color: rgba(255,255,255,0.25); background: rgba(255,255,255,0.05); }

/* ===== 主内容区 ===== */
.auth-content { position: relative; z-index: 10; flex: 1; display: flex; align-items: center; justify-content: center; padding: 0 16px 32px; }

/* ===== Glass Card ===== */
.glass-card { background: rgba(255,255,255,0.03); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px); border-radius: 16px; border: 1px solid rgba(255,255,255,0.06); position: relative; overflow: hidden; width: 100%; max-width: 400px; }
.glass-card::before { content: ''; position: absolute; inset: 0; border-radius: inherit; padding: 1px; background: linear-gradient(180deg, rgba(255,255,255,0.12) 0%, rgba(255,255,255,0.02) 30%, rgba(255,255,255,0) 60%, rgba(255,255,255,0.02) 80%, rgba(255,255,255,0.08) 100%); -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0); -webkit-mask-composite: xor; mask-composite: exclude; pointer-events: none; }
.card-inner { padding: 28px 24px 24px; }
@media (min-width: 640px) { .card-inner { padding: 32px 28px 28px; } }

/* ===== 卡片头部 ===== */
.card-header { text-align: center; margin-bottom: 24px; }
.card-header h2 { margin: 0 0 6px; font-size: 20px; font-weight: 600; color: #fff; }
.card-desc { margin: 0; font-size: 14px; color: rgba(255,255,255,0.35); }

/* ===== 表单 ===== */
.auth-form :deep(.el-form-item) { margin-bottom: 16px; }
.auth-form :deep(.el-form-item:last-child) { margin-bottom: 0; }
.auth-form :deep(.el-input__wrapper) { background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); border-radius: 10px; box-shadow: none; transition: all 0.3s; padding: 4px 12px; }
.auth-form :deep(.el-input__wrapper:hover) { border-color: rgba(255,255,255,0.15); }
.auth-form :deep(.el-input__wrapper.is-focus) { border-color: rgba(255,255,255,0.25); box-shadow: 0 0 0 3px rgba(255,255,255,0.04); }
.auth-form :deep(.el-input__inner) { color: #fff; font-size: 14px; }
.auth-form :deep(.el-input__inner::placeholder) { color: rgba(255,255,255,0.25); }
.auth-form :deep(.el-input__prefix .el-icon) { color: rgba(255,255,255,0.25); font-size: 16px; }
.auth-form :deep(.el-input__suffix .el-icon) { color: rgba(255,255,255,0.25); }
.auth-form :deep(.el-input-group__append) { background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.08); border-left: none; border-radius: 0 10px 10px 0; box-shadow: none; color: #fff; }
.auth-form :deep(.el-input-group__append .el-button) { color: rgba(255,255,255,0.7); font-size: 13px; }
.auth-form :deep(.el-input-group__append .el-button:disabled) { color: rgba(255,255,255,0.2); }

/* ===== 验证码 ===== */
.captcha-wrapper { display: flex; align-items: center; gap: 10px; padding: 11px 16px; border: 1px solid rgba(255,255,255,0.08); border-radius: 10px; cursor: pointer; transition: all 0.3s; background: rgba(255,255,255,0.04); width: 100%; }
.captcha-wrapper:hover { border-color: rgba(255,255,255,0.15); background: rgba(255,255,255,0.06); }
.captcha-wrapper.captcha-verified { border-color: rgba(34,197,94,0.3); background: rgba(34,197,94,0.06); }
.captcha-icon { font-size: 18px; color: rgba(255,255,255,0.3); flex-shrink: 0; }
.captcha-text { flex: 1; font-size: 14px; color: rgba(255,255,255,0.35); }
.captcha-check { color: #22c55e; font-size: 18px; }

/* ===== 按钮 ===== */
.submit-btn { width: 100%; height: 42px; border-radius: 10px; font-size: 14px; font-weight: 500; background: #fff; color: #000; border: none; transition: all 0.2s; }
.submit-btn:hover:not(:disabled) { background: rgba(229,229,229,1); }
.submit-btn:disabled { opacity: 0.25; cursor: not-allowed; }

/* ===== 底部链接 ===== */
.card-footer { text-align: center; margin-top: 20px; padding-top: 16px; border-top: 1px solid rgba(255,255,255,0.06); color: rgba(255,255,255,0.3); font-size: 13px; }
.footer-link { color: rgba(255,255,255,0.5); text-decoration: none; font-weight: 500; transition: color 0.2s; }
.footer-link:hover { color: #fff; }

/* ===== 版权信息 ===== */
.page-footer { position: relative; z-index: 10; text-align: center; padding: 0 16px 16px; font-size: 12px; color: rgba(255,255,255,0.2); }
</style>