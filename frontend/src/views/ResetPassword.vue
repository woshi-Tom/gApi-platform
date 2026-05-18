<template>
  <div class="auth-page">
    <div class="bg-decoration">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
      <div class="orb orb-3"></div>
      <div class="grid-lines"></div>
    </div>

    <nav class="navbar">
      <router-link to="/login" class="nav-logo animate-fade-up" style="animation-delay: 0ms">gAPI Platform</router-link>
      <router-link to="/login" class="nav-link-btn animate-fade-up" style="animation-delay: 150ms">返回登录</router-link>
    </nav>

    <main class="auth-content">
      <div class="glass-card animate-fade-up" style="animation-delay: 300ms">
        <div class="card-inner">
          <div class="card-header">
            <h2 class="animate-fade-up" style="animation-delay: 400ms">重置密码</h2>
            <p class="card-desc animate-fade-up" style="animation-delay: 450ms">为您的账户设置新密码</p>
          </div>

          <div v-if="loading" class="animate-fade-up" style="animation-delay: 500ms">
            <el-skeleton :rows="3" animated />
          </div>

          <div v-else-if="tokenValid" class="animate-fade-up" style="animation-delay: 500ms">
            <p class="email-hint">为 <strong>{{ email }}</strong> 设置新密码</p>
            <el-form @submit.prevent="handleSubmit" class="auth-form">
              <el-form-item>
                <el-input v-model="form.password" type="password" show-password placeholder="新密码（至少8位）" :prefix-icon="Lock" size="large" />
                <div v-if="form.password" class="password-strength">
                  <span>密码强度：</span>
                  <el-tag :type="passwordStrengthType" size="small">{{ passwordStrengthText }}</el-tag>
                </div>
              </el-form-item>
              <el-form-item>
                <el-input v-model="form.confirmPassword" type="password" show-password placeholder="再次输入新密码" :prefix-icon="Lock" size="large" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" native-type="submit" :loading="submitting" :disabled="!canSubmit" @click="handleSubmit" size="large" class="submit-btn">重置密码</el-button>
              </el-form-item>
            </el-form>
          </div>

          <div v-else class="error-content animate-fade-up" style="animation-delay: 500ms">
            <el-icon class="error-icon"><CircleClose /></el-icon>
            <h3>链接已失效</h3>
            <p>重置链接无效或已过期，请重新申请</p>
            <el-button type="primary" @click="$router.push('/forgot-password')" size="large" class="submit-btn" style="margin-top: 20px">重新申请</el-button>
          </div>

          <div class="card-footer animate-fade-up" style="animation-delay: 600ms">
            <router-link to="/login" class="footer-link">返回登录</router-link>
          </div>
        </div>
      </div>
    </main>

    <footer class="page-footer animate-fade-up" style="animation-delay: 700ms">© {{ new Date().getFullYear() }} gAPI Platform. All rights reserved.</footer>
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

const form = reactive({ password: '', confirmPassword: '' })

const canSubmit = computed(() => form.password.length >= 8 && form.password === form.confirmPassword)

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
  if (!token) { error.value = 'Missing token'; tokenValid.value = false; loading.value = false; return }
  try {
    const res = await request.get('/auth/reset-password', { params: { token } })
    if (res.data?.success) { email.value = res.data.data.email; tokenValid.value = true }
    else { error.value = res.data?.error?.message || 'Invalid token'; tokenValid.value = false }
  } catch (e: any) { error.value = e.response?.data?.error?.message || 'Link expired'; tokenValid.value = false }
  finally { loading.value = false }
})

async function handleSubmit() {
  if (!canSubmit.value) { ElMessage.warning('请填写所有字段'); return }
  if (form.password !== form.confirmPassword) { ElMessage.warning('两次输入的密码不一致'); return }
  const token = route.query.token as string
  submitting.value = true
  try {
    await request.post('/auth/reset-password', { token, password: form.password, confirm_password: form.confirmPassword })
    ElMessage.success('密码重置成功')
    router.push('/login')
  } catch (e: any) { ElMessage.error(e.response?.data?.error?.message || '重置失败') }
  finally { submitting.value = false }
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
</style>

<style scoped>
.auth-page { position: relative; width: 100%; min-height: 100vh; display: flex; flex-direction: column; background: #000; font-family: 'Inter', sans-serif; overflow: hidden; color: #fff; }
.bg-decoration { position: fixed; inset: 0; z-index: 0; pointer-events: none; overflow: hidden; }
.orb { position: absolute; border-radius: 50%; filter: blur(120px); opacity: 0.15; }
.orb-1 { width: 600px; height: 600px; background: radial-gradient(circle, #6366f1, transparent 70%); top: -200px; left: -150px; animation: orb-1 25s ease-in-out infinite; }
.orb-2 { width: 500px; height: 500px; background: radial-gradient(circle, #8b5cf6, transparent 70%); bottom: -200px; right: -100px; animation: orb-2 30s ease-in-out infinite; }
.orb-3 { width: 400px; height: 400px; background: radial-gradient(circle, #3b82f6, transparent 70%); top: 40%; left: 50%; animation: orb-3 20s ease-in-out infinite; }
@keyframes orb-1 { 0%, 100% { transform: translate(0, 0) scale(1); } 50% { transform: translate(40px, 30px) scale(1.08); } }
@keyframes orb-2 { 0%, 100% { transform: translate(0, 0) scale(1); } 50% { transform: translate(-30px, -20px) scale(1.05); } }
@keyframes orb-3 { 0%, 100% { transform: translate(-50%, 0) scale(1); } 50% { transform: translate(-50%, 30px) scale(0.95); } }
.grid-lines { position: absolute; inset: 0; background-image: linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px); background-size: 80px 80px; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(20px); filter: blur(10px); } to { opacity: 1; transform: translateY(0); filter: blur(0); } }
.animate-fade-up { animation: fadeUp 0.8s ease-out forwards; opacity: 0; }
.navbar { position: relative; z-index: 50; display: flex; justify-content: space-between; align-items: center; padding: 16px; }
@media (min-width: 768px) { .navbar { padding: 16px 48px; } }
.nav-logo { font-size: 24px; font-weight: 600; letter-spacing: -0.04em; color: #fff; text-decoration: none; transition: opacity 0.2s; }
@media (min-width: 768px) { .nav-logo { font-size: 28px; } }
.nav-logo:hover { opacity: 0.8; }
.nav-link-btn { padding: 8px 20px; border-radius: 8px; font-size: 14px; font-weight: 500; color: rgba(255,255,255,0.8); text-decoration: none; border: 1px solid rgba(255,255,255,0.1); transition: all 0.2s; }
.nav-link-btn:hover { border-color: rgba(255,255,255,0.25); background: rgba(255,255,255,0.05); }
.auth-content { position: relative; z-index: 10; flex: 1; display: flex; align-items: center; justify-content: center; padding: 0 16px 32px; }
.glass-card { background: rgba(255,255,255,0.03); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px); border-radius: 16px; border: 1px solid rgba(255,255,255,0.06); position: relative; overflow: hidden; width: 100%; max-width: 400px; }
.glass-card::before { content: ''; position: absolute; inset: 0; border-radius: inherit; padding: 1px; background: linear-gradient(180deg, rgba(255,255,255,0.12) 0%, rgba(255,255,255,0.02) 30%, rgba(255,255,255,0) 60%, rgba(255,255,255,0.02) 80%, rgba(255,255,255,0.08) 100%); -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0); -webkit-mask-composite: xor; mask-composite: exclude; pointer-events: none; }
.card-inner { padding: 28px 24px 24px; }
@media (min-width: 640px) { .card-inner { padding: 32px 28px 28px; } }
.card-header { text-align: center; margin-bottom: 24px; }
.card-header h2 { margin: 0 0 6px; font-size: 20px; font-weight: 600; color: #fff; }
.card-desc { margin: 0; font-size: 14px; color: rgba(255,255,255,0.35); }
.email-hint { color: rgba(255,255,255,0.5); text-align: center; margin-bottom: 20px; font-size: 14px; }
.email-hint strong { color: #fff; }
.auth-form :deep(.el-form-item) { margin-bottom: 16px; }
.auth-form :deep(.el-form-item:last-child) { margin-bottom: 0; }
.auth-form :deep(.el-input__wrapper) { background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); border-radius: 10px; box-shadow: none; transition: all 0.3s; padding: 4px 12px; }
.auth-form :deep(.el-input__wrapper:hover) { border-color: rgba(255,255,255,0.15); }
.auth-form :deep(.el-input__wrapper.is-focus) { border-color: rgba(255,255,255,0.25); box-shadow: 0 0 0 3px rgba(255,255,255,0.04); }
.auth-form :deep(.el-input__inner) { color: #fff; font-size: 14px; }
.auth-form :deep(.el-input__inner::placeholder) { color: rgba(255,255,255,0.25); }
.auth-form :deep(.el-input__prefix .el-icon) { color: rgba(255,255,255,0.25); font-size: 16px; }
.password-strength { display: flex; align-items: center; gap: 8px; font-size: 12px; color: rgba(255,255,255,0.35); margin-top: 8px; }
.error-content { text-align: center; padding: 20px 0; }
.error-icon { font-size: 56px; color: #f56c6c; margin-bottom: 16px; }
.error-content h3 { margin: 0 0 16px 0; font-size: 18px; font-weight: 600; color: #fff; }
.error-content p { color: rgba(255,255,255,0.5); margin: 8px 0; font-size: 14px; }
.submit-btn { width: 100%; height: 42px; border-radius: 10px; font-size: 14px; font-weight: 500; background: #fff; color: #000; border: none; transition: all 0.2s; }
.submit-btn:hover:not(:disabled) { background: rgba(229,229,229,1); }
.submit-btn:disabled { opacity: 0.25; cursor: not-allowed; }
.card-footer { text-align: center; margin-top: 20px; padding-top: 16px; border-top: 1px solid rgba(255,255,255,0.06); font-size: 13px; }
.footer-link { color: rgba(255,255,255,0.5); text-decoration: none; font-weight: 500; transition: color 0.2s; }
.footer-link:hover { color: #fff; }
.page-footer { position: relative; z-index: 10; text-align: center; padding: 0 16px 16px; font-size: 12px; color: rgba(255,255,255,0.2); }
</style>