<template>
  <AuthLayout :showLoginLink="true">
    <!-- 品牌展示区 -->
    <template #brand>
      <h1 class="brand-title">
        <span class="gradient-text">重置密码</span>
        <br />设置您的新密码
      </h1>
      <p class="brand-desc">为了您的账户安全，建议使用包含大小写字母、数字和特殊字符的强密码</p>

      <!-- 特性亮点 -->
      <div class="feature-list animate-fade-up" style="animation-delay: 600ms">
        <div class="feature-item">
          <div class="feature-icon icon-green">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
          </div>
          <div class="feature-text">
            <strong>安全加密</strong>
            密码加密存储，传输全程 HTTPS
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-purple">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
          </div>
          <div class="feature-text">
            <strong>即时生效</strong>
            重置后立即使用新密码登录
          </div>
        </div>
      </div>
    </template>

    <!-- 重置密码卡片内容 -->
    <div class="card-header">
      <h2 class="animate-fade-up" style="animation-delay: 400ms">重置密码</h2>
      <p class="card-desc animate-fade-up" style="animation-delay: 450ms">为您的账户设置新密码</p>
    </div>

    <div v-if="loading" class="animate-fade-up" style="animation-delay: 500ms">
      <div class="skeleton-wrapper">
        <div class="skeleton-line"></div>
        <div class="skeleton-line short"></div>
        <div class="skeleton-line"></div>
      </div>
    </div>

    <div v-else-if="tokenValid" class="animate-fade-up" style="animation-delay: 500ms">
      <p class="email-hint">为 <strong>{{ email }}</strong> 设置新密码</p>
      <el-form @submit.prevent="handleSubmit" class="auth-form">
        <el-form-item>
          <el-input v-model="form.password" type="password" show-password placeholder="新密码（至少8位）" :prefix-icon="Lock" size="large" />
          <div v-if="form.password" class="password-strength-bar">
            <div class="bar-fill" :class="passwordStrengthClass"></div>
          </div>
          <div v-if="form.password" class="password-strength-info">
            <span class="password-strength-text" :class="passwordStrengthClass">{{ passwordStrengthText }}</span>
          </div>
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.confirmPassword" type="password" show-password placeholder="再次输入新密码" :prefix-icon="Lock" size="large" />
          <span v-if="form.confirmPassword && form.password !== form.confirmPassword" class="mismatch-hint">两次输入的密码不一致</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="submitting" :disabled="!canSubmit" @click="handleSubmit" size="large" class="submit-btn">重置密码</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div v-else class="error-content animate-fade-up" style="animation-delay: 500ms">
      <div class="error-icon-wrapper">
        <el-icon class="error-icon"><CircleClose /></el-icon>
      </div>
      <h3>链接已失效</h3>
      <p>重置链接无效或已过期，请重新申请</p>
      <el-button type="primary" @click="$router.push('/forgot-password')" size="large" class="submit-btn" style="margin-top: 20px">重新申请</el-button>
    </div>

    <div class="card-footer animate-fade-up" style="animation-delay: 600ms">
      <router-link to="/login" class="footer-link">返回登录</router-link>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, CircleClose } from '@element-plus/icons-vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import request from '@/api/request'
import '@/styles/auth.css'

const router = useRouter()
const route = useRoute()

const loading = ref(true)
const submitting = ref(false)
const tokenValid = ref(false)
const email = ref('')

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

const passwordStrengthClass = computed(() => {
  const s = passwordStrength.value
  if (s <= 1) return 'strength-weak'
  if (s <= 2) return 'strength-fair'
  if (s <= 3) return 'strength-good'
  return 'strength-strong'
})

const passwordStrengthText = computed(() => {
  const s = passwordStrength.value
  if (s <= 1) return '弱'
  if (s <= 2) return '一般'
  if (s <= 3) return '良好'
  return '强'
})

onMounted(async () => {
  const token = route.query.token as string
  if (!token) { tokenValid.value = false; loading.value = false; return }
  try {
    const res = await request.get('/auth/reset-password', { params: { token } })
    if (res.data?.success) { email.value = res.data.data.email; tokenValid.value = true }
    else { tokenValid.value = false }
  } catch { tokenValid.value = false }
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

<style scoped>
/* ===== 品牌展示区 ===== */
.brand-title { font-size: 28px; font-weight: 600; letter-spacing: 0; line-height: 1.2; margin: 0 0 16px; color: var(--c-text); }
@media (min-width: 640px) { .brand-title { font-size: 36px; } }
@media (min-width: 768px) { .brand-title { font-size: 44px; } }
@media (min-width: 1024px) { .brand-title { font-size: 52px; } }

.gradient-text { background: linear-gradient(135deg, #fff 0%, rgba(167, 139, 250, 0.7) 50%, rgba(129, 140, 248, 0.8) 100%); background-size: 200% 200%; animation: gradient-shift 6s ease infinite; -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }

.brand-desc { font-size: 15px; color: var(--c-text-sub); margin: 0 0 24px; line-height: 1.6; max-width: 400px; }
@media (min-width: 768px) { .brand-desc { font-size: 16px; } }

/* ===== 卡片头部 ===== */
.card-header { text-align: center; margin-bottom: 24px; }
.card-header h2 { margin: 0 0 6px; font-size: 22px; font-weight: 600; color: var(--c-text); }
.card-desc { margin: 0; font-size: 14px; color: var(--c-text-sub); }

/* ===== 骨架屏 ===== */
.skeleton-wrapper { padding: 8px 0; }
.skeleton-line { height: 14px; background: linear-gradient(90deg, rgba(255, 255, 255, 0.04) 25%, rgba(255, 255, 255, 0.08) 50%, rgba(255, 255, 255, 0.04) 75%); background-size: 200% 100%; animation: shimmer 1.5s ease-in-out infinite; border-radius: 6px; margin-bottom: 16px; }
.skeleton-line.short { width: 60%; }

/* ===== 邮箱提示 ===== */
.email-hint { color: var(--c-text-sub); text-align: center; margin-bottom: 20px; font-size: 14px; }
.email-hint strong { color: var(--c-text-secondary); }

/* ===== 表单 ===== */
.auth-form :deep(.el-form-item) { margin-bottom: 16px; }
.auth-form :deep(.el-form-item:last-child) { margin-bottom: 0; }
.auth-form :deep(.el-input__wrapper) { background: var(--c-input-bg); border: 1px solid var(--c-input-border); border-radius: 10px; box-shadow: none; transition: all 0.3s; padding: 4px 12px; }
.auth-form :deep(.el-input__wrapper:hover) { border-color: var(--c-input-hover); }
.auth-form :deep(.el-input__wrapper.is-focus) { border-color: var(--c-primary); box-shadow: 0 0 0 3px var(--c-focus-ring); }
.auth-form :deep(.el-input__inner) { color: var(--c-text); font-size: 14px; }
.auth-form :deep(.el-input__inner::placeholder) { color: var(--c-text-muted); }
.auth-form :deep(.el-input__prefix .el-icon) { color: var(--c-icon); font-size: 16px; }

/* ===== 密码强度 ===== */
.password-strength-info { display: flex; align-items: center; gap: 8px; margin-top: 6px; }
.password-strength-text { font-size: 12px; transition: color 0.3s ease; }
.password-strength-text.strength-weak { color: #ef4444; }
.password-strength-text.strength-fair { color: #f59e0b; }
.password-strength-text.strength-good { color: #10b981; }
.password-strength-text.strength-strong { color: var(--c-primary); }

/* ===== 密码不一致提示 ===== */
.mismatch-hint { display: block; font-size: 12px; color: #ef4444; margin-top: 6px; }

/* ===== 提交按钮 ===== */
.submit-btn { width: 100%; height: 44px; border-radius: 10px; font-size: 14px; font-weight: 500; background: var(--c-primary); color: #fff; border: none; transition: all 0.3s ease; position: relative; overflow: hidden; }
.submit-btn::after { content: ''; position: absolute; top: 0; left: -100%; width: 100%; height: 100%; background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent); display: none; }
.submit-btn:hover:not(:disabled) { background: var(--c-primary-hover); box-shadow: none; transform: translateY(-1px); }
.submit-btn:active:not(:disabled) { transform: translateY(0); }
.submit-btn:disabled { opacity: 0.3; cursor: not-allowed; }

/* ===== 错误态 ===== */
.error-content { text-align: center; padding: 20px 0; }
.error-icon-wrapper { position: relative; display: inline-flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.error-icon { font-size: 56px; color: rgba(239, 68, 68, 0.8); animation: envelope-fly 0.6s cubic-bezier(0.22, 1, 0.36, 1) forwards; }
.error-icon-wrapper::before { content: ''; position: absolute; width: 80px; height: 80px; border-radius: 50%; background: rgba(239, 68, 68, 0.06); animation: pulse-ring 2s ease-in-out infinite; }
.error-content h3 { margin: 0 0 16px 0; font-size: 18px; font-weight: 600; color: var(--c-text); }
.error-content p { color: var(--c-text-sub); margin: 8px 0; font-size: 14px; }

/* ===== 底部链接 ===== */
.card-footer { text-align: center; margin-top: 20px; padding-top: 16px; border-top: 1px solid var(--c-border); font-size: 13px; }
.footer-link { color: var(--c-text-sub); text-decoration: none; font-weight: 500; transition: color 0.2s ease; }
.footer-link:hover { color: var(--c-text); }
</style>