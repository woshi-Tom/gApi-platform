<template>
  <AuthLayout :showLoginLink="true">
    <!-- 品牌展示区 -->
    <template #brand>
      <h1 class="brand-title">
        <span class="gradient-text">加入我们</span>
        <br />开启您的 AI 之旅
      </h1>
      <p class="brand-desc">注册即可获得免费额度，体验多渠道 AI API 中转服务</p>

      <!-- 用户概览面板 -->
      <StatsPanel
        title="用户概览"
        live-label="今日状态"
        :stats="registerStats"
      />

      <!-- 特性亮点 -->
      <div class="feature-list animate-fade-up" style="animation-delay: 700ms">
        <div class="feature-item">
          <div class="feature-icon icon-purple">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M16 8l-8 8"/><path d="M8 8l8 8"/></svg>
          </div>
          <div class="feature-text">
            <strong>免费额度</strong>
            注册即送体验额度，即刻开始使用
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-blue">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
          </div>
          <div class="feature-text">
            <strong>安全可靠</strong>
            数据加密传输，隐私安全保障
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-orange">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
          </div>
          <div class="feature-text">
            <strong>极速接入</strong>
            标准 OpenAI 接口，无需改动代码
          </div>
        </div>
      </div>
    </template>

    <!-- 注册卡片内容 -->
    <div class="card-header">
      <h2 class="animate-fade-up" style="animation-delay: 400ms">创建账号</h2>
      <p class="card-desc animate-fade-up" style="animation-delay: 450ms">注册您的 gAPI Platform 账户</p>
    </div>

    <el-form @submit.prevent="handleRegister" class="auth-form animate-fade-up" style="animation-delay: 500ms">
      <!-- 用户名 -->
      <el-form-item>
        <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" size="large" />
      </el-form-item>

      <!-- 密码 -->
      <el-form-item>
        <div class="password-row">
          <el-input v-model="form.password" type="password" placeholder="请输入密码（至少8位）" :prefix-icon="Lock" size="large" show-password />
          <div v-if="form.password" class="password-strength-inline">
            <div class="password-strength-bar">
              <div class="bar-fill" :class="passwordStrengthClass"></div>
            </div>
            <span class="password-strength-text" :class="passwordStrengthClass">
              {{ passwordStrengthText }}
            </span>
          </div>
        </div>
      </el-form-item>

      <!-- 确认密码 -->
      <el-form-item class="confirm-password-item">
        <div class="confirm-password-row">
          <el-input v-model="form.confirmPassword" type="password" placeholder="请确认密码" :prefix-icon="Lock" size="large" show-password />
          <span v-if="form.confirmPassword && form.confirmPassword !== form.password" class="confirm-password-error">
            两次输入的密码不一致
          </span>
        </div>
      </el-form-item>

      <!-- 邮箱 -->
      <el-form-item>
        <el-input v-model="form.email" placeholder="请输入邮箱" :prefix-icon="Message" size="large" @blur="checkEmailFormat" />
      </el-form-item>

      <!-- 验证码 -->
      <el-form-item>
        <div class="captcha-row">
          <div class="code-input-group code-input-inline">
            <input
              v-for="(_, i) in 6"
              :key="i"
              :ref="el => codeInputRefs[i] = el as HTMLInputElement"
              type="text"
              maxlength="1"
              class="code-input"
              :class="{ filled: codeDigits[i] }"
              :value="codeDigits[i]"
              @input="onCodeInput(i, $event)"
              @keydown="onCodeKeydown(i, $event)"
              @paste="onCodePaste"
              placeholder="·"
            />
          </div>
          <el-button
            class="send-code-btn"
            @click="handleSendCode"
            :disabled="countdown > 0 || sendingCode"
            :loading="sendingCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
          </el-button>
        </div>
      </el-form-item>

      <!-- 用户协议 -->
      <el-form-item>
        <el-checkbox v-model="agreeTerms" class="agree-checkbox">
          <span class="agree-text">
            我已阅读并同意
            <a href="#" @click.prevent class="agree-link">《用户协议》</a>
          </span>
        </el-checkbox>
      </el-form-item>

      <!-- 注册按钮 -->
      <el-form-item>
        <el-button
          type="primary"
          native-type="submit"
          :loading="loading"
          :disabled="!canSubmit"
          @click="handleRegister"
          size="large"
          class="submit-btn"
        >
          <span v-if="!loading">注 册</span>
          <span v-else>注册中...</span>
        </el-button>
      </el-form-item>
    </el-form>

    <div class="card-footer animate-fade-up" style="animation-delay: 600ms">
      <span>已有账号？</span>
      <router-link to="/login" class="footer-link">立即登录</router-link>
    </div>

    <SlideCaptcha v-model:visible="showCaptcha" @success="onCaptchaSuccess" ref="captchaRef" />
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { User, Message, Lock } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import StatsPanel from '@/components/auth/StatsPanel.vue'
import request from '@/api/request'
import '@/styles/auth.css'

const registerStats = [
  { value: '12,860', label: '已注册用户', fillWidth: '82%' },
  { value: '128', label: '今日新增', color: 'green' as const, fillWidth: '65%' },
  { value: '2,430', label: '已创建 Key', color: 'purple' as const, fillWidth: '72%' },
]

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const showCaptcha = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
const countdownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const captchaRef = ref()
const isValidEmail = ref(false)
const captchaToken = ref('')
const agreeTerms = ref(false)
const codeSent = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
})
const codeDigits = ref<string[]>(['', '', '', '', '', ''])
const codeInputRefs = ref<(HTMLInputElement | null)[]>([])

// 同步 codeDigits 到 code 字符串
const code = computed(() => codeDigits.value.join(''))

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

const canSubmit = computed(() => {
  return (
    form.username &&
    isValidEmail.value &&
    form.password.length >= 8 &&
    form.confirmPassword === form.password &&
    code.value.length === 6 &&
    agreeTerms.value &&
    codeSent.value
  )
})

function checkEmailFormat() {
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
  isValidEmail.value = emailRe.test(form.email)
  if (!isValidEmail.value && form.email) {
    ElMessage.warning('请输入有效的邮箱格式')
  }
}

// 验证码输入处理
function onCodeInput(index: number, event: Event) {
  const input = event.target as HTMLInputElement
  const value = input.value.replace(/\D/g, '')

  if (value) {
    codeDigits.value[index] = value[0]
    if (index < 5) {
      codeInputRefs.value[index + 1]?.focus()
    }
  } else {
    codeDigits.value[index] = ''
  }
}

function onCodeKeydown(index: number, event: KeyboardEvent) {
  if (event.key === 'Backspace' && !codeDigits.value[index] && index > 0) {
    codeDigits.value[index - 1] = ''
    codeInputRefs.value[index - 1]?.focus()
  }
}

function onCodePaste(event: ClipboardEvent) {
  event.preventDefault()
  const pasted = event.clipboardData?.getData('text')?.replace(/\D/g, '') || ''
  for (let i = 0; i < 6; i++) {
    codeDigits.value[i] = pasted[i] || ''
  }
  const lastFilled = Math.min(pasted.length, 6) - 1
  if (lastFilled >= 0) {
    codeInputRefs.value[Math.min(lastFilled + 1, 5)]?.focus()
  }
}

// 滑块验证成功回调
function onCaptchaSuccess(token: string) {
  captchaToken.value = token
  // 滑块验证通过后自动发送验证码
  sendCode()
}

// 点击获取验证码：先弹滑块验证
function handleSendCode() {
  if (!form.email || !isValidEmail.value) {
    ElMessage.warning('请输入有效的邮箱')
    return
  }
  if (captchaToken.value) {
    // 已有滑块 token，直接发送
    sendCode()
  } else {
    // 弹出滑块验证
    showCaptcha.value = true
  }
}

// 发送验证码
async function sendCode() {
  sendingCode.value = true
  try {
    const payload: any = { email: form.email }
    if (captchaToken.value) payload.captcha_token = captchaToken.value
    await request.post('/email/send-code', payload)
    ElMessage.success('验证码已发送到您的邮箱')
    codeSent.value = true
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
    captchaToken.value = ''
    captchaRef.value?.reset()
  } finally { sendingCode.value = false }
}

// 注册
async function handleRegister() {
  if (!form.username || !form.email || !form.password || !form.confirmPassword) {
    ElMessage.warning('请填写所有字段'); return
  }
  if (!isValidEmail.value) { ElMessage.warning('请输入有效的邮箱'); return }
  if (form.password.length < 8) { ElMessage.warning('密码至少8位'); return }
  if (form.password !== form.confirmPassword) { ElMessage.warning('两次输入的密码不一致'); return }
  if (code.value.length !== 6) { ElMessage.warning('请输入6位验证码'); return }
  if (!agreeTerms.value) { ElMessage.warning('请阅读并同意用户协议'); return }

  loading.value = true
  try {
    // 先验证邮箱验证码
    await request.post('/email/verify-code', { email: form.email, code: code.value })
    // 再注册
    await authStore.register(form.username, form.email, form.password)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (e: any) {
    const errorData = e.response?.data?.error
    if (errorData?.code === 'REGISTRATION_CLOSED') { router.push('/register-closed'); return }
    ElMessage.error(errorData?.message || '注册失败')
    captchaToken.value = ''
    captchaRef.value?.reset()
  } finally { loading.value = false }
}
</script>

<style scoped>
/* ===== 品牌展示区 ===== */
.brand-title {
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.04em;
  line-height: 1.2;
  margin: 0 0 16px;
  color: var(--c-text);
}

@media (min-width: 640px) {
  .brand-title {
    font-size: 36px;
  }
}

@media (min-width: 768px) {
  .brand-title {
    font-size: 44px;
  }
}

@media (min-width: 1024px) {
  .brand-title {
    font-size: 52px;
  }
}

.gradient-text {
  background: linear-gradient(135deg, #fff 0%, rgba(167, 139, 250, 0.7) 50%, rgba(129, 140, 248, 0.8) 100%);
  background-size: 200% 200%;
  animation: gradient-shift 6s ease infinite;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-desc {
  font-size: 15px;
  color: var(--c-text-sub);
  margin: 0 0 24px;
  line-height: 1.6;
  max-width: 400px;
}

@media (min-width: 768px) {
  .brand-desc {
    font-size: 16px;
  }
}

/* ===== 卡片头部 ===== */
.card-header {
  text-align: center;
  margin-bottom: 20px;
}

.card-header h2 {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 600;
  color: var(--c-text);
}

.card-desc {
  margin: 0;
  font-size: 14px;
  color: var(--c-text-sub);
}

/* ===== 表单 ===== */
.auth-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.auth-form :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

.auth-form :deep(.el-input__wrapper) {
  background: var(--c-input-bg);
  border: 1px solid var(--c-input-border);
  border-radius: 10px;
  box-shadow: none;
  transition: all 0.3s;
  padding: 4px 12px;
}

.auth-form :deep(.el-input__wrapper:hover) {
  border-color: var(--c-input-hover);
}

.auth-form :deep(.el-input__wrapper.is-focus) {
  border-color: rgba(99, 102, 241, 0.4);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.08), 0 0 20px rgba(99, 102, 241, 0.04);
}

.auth-form :deep(.el-input__inner) {
  color: var(--c-text);
  font-size: 14px;
}

.auth-form :deep(.el-input__inner::placeholder) {
  color: var(--c-text-muted);
}

.auth-form :deep(.el-input__prefix .el-icon) {
  color: var(--c-icon);
  font-size: 16px;
}

.auth-form :deep(.el-input__suffix .el-icon) {
  color: var(--c-icon);
}

/* ===== 验证码行 ===== */
.captcha-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.code-input-inline {
  flex: 1;
  min-width: 0;
}

.code-input-inline .code-input {
  width: calc((100% - 40px) / 6);
  min-width: 0;
}

.send-code-btn {
  flex-shrink: 0;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  border-radius: 8px;
  font-size: 13px;
  height: 48px;
  transition: all 0.3s ease;
}

.send-code-btn:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.3);
  color: #c7d2fe;
}

.send-code-btn:disabled {
  opacity: 0.4;
  color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.06);
}

/* ===== 确认密码错误提示 ===== */
.confirm-password-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.confirm-password-row .el-input {
  flex: 1;
  min-width: 0;
}

.confirm-password-error {
  flex-shrink: 0;
  font-size: 12px;
  color: #ef4444;
  white-space: nowrap;
}

/* ===== 密码行（输入框 + 强度提示） ===== */
.password-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.password-row .el-input {
  flex: 1;
  min-width: 0;
}

.password-strength-inline {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.password-strength-inline .password-strength-bar {
  width: 50px;
  flex-shrink: 0;
}

/* ===== 密码强度 ===== */
.password-strength-text {
  font-size: 12px;
  white-space: nowrap;
  transition: color 0.3s ease;
}

.password-strength-text.strength-weak { color: #ef4444; }
.password-strength-text.strength-fair { color: #f59e0b; }
.password-strength-text.strength-good { color: #10b981; }
.password-strength-text.strength-strong { color: #8b5cf6; }

/* ===== 用户协议 ===== */
.agree-checkbox {
  display: flex;
  align-items: center;
}

.agree-text {
  font-size: 13px;
  color: var(--c-text-sub);
}

.agree-link {
  color: #a5b4fc;
  text-decoration: none;
  transition: color 0.2s ease;
}

.agree-link:hover {
  color: #c7d2fe;
  text-decoration: underline;
}

.auth-form :deep(.el-checkbox__label) {
  padding-left: 6px;
}

.auth-form :deep(.el-checkbox__inner) {
  background: var(--c-input-bg);
  border-color: var(--c-input-border);
}

.auth-form :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background: rgba(99, 102, 241, 0.8);
  border-color: rgba(99, 102, 241, 0.8);
}

/* ===== 提交按钮 ===== */
.submit-btn {
  width: 100%;
  height: 44px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  border: none;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.submit-btn::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent);
  animation: btn-glow 3s ease-in-out infinite;
}

.submit-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #818cf8, #a78bfa);
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.3);
  transform: translateY(-1px);
}

.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.submit-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

/* ===== 底部链接 ===== */
.card-footer {
  text-align: center;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--c-border);
  color: var(--c-text-muted);
  font-size: 13px;
}

.footer-link {
  color: var(--c-text-sub);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: var(--c-text);
}
</style>