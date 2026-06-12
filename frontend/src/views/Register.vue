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

    <div class="step-summary animate-fade-up" style="animation-delay: 475ms">
      <span>{{ stepTitle }} · {{ currentStep }} / 3</span>
      <div class="step-progress" aria-hidden="true">
        <span :style="{ width: `${stepProgress}%` }"></span>
      </div>
    </div>

    <el-form @submit.prevent="handlePrimaryAction" class="auth-form animate-fade-up" style="animation-delay: 500ms">
      <div class="step-panel">
        <template v-if="currentStep === 1">
          <!-- 用户名 -->
          <el-form-item>
            <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" size="large" />
          </el-form-item>

          <!-- 邮箱 -->
          <el-form-item>
            <el-input v-model="form.email" placeholder="请输入邮箱" :prefix-icon="Message" size="large" @blur="checkEmailFormat" />
          </el-form-item>

          <el-form-item class="action-item">
            <el-button type="primary" native-type="submit" size="large" class="submit-btn">
              下一步
            </el-button>
          </el-form-item>
        </template>

        <template v-else-if="currentStep === 2">
          <div class="email-review">
            <div class="email-review-content">
              <span>当前邮箱</span>
              <strong>{{ form.email }}</strong>
            </div>
            <el-button link type="primary" native-type="button" class="change-email-btn" @click="handleEditEmail">
              修改邮箱
            </el-button>
          </div>

          <el-form-item class="turnstile-form-item">
            <TurnstileWidget
              ref="turnstileRef"
              v-show="!isWaitingTurnstileReset"
              v-model="turnstileToken"
              @expired="handleTurnstileExpired"
              @error="handleTurnstileError"
              @timeout="handleTurnstileTimeout"
            />
            <p v-if="isWaitingTurnstileReset" class="turnstile-sent-message">
              验证码已发送，倒计时结束后可重新验证并再次发送
            </p>
          </el-form-item>

          <!-- 验证码 -->
          <el-form-item>
            <div class="captcha-row">
              <el-input
                v-model="code"
                placeholder="请输入 6 位验证码"
                :prefix-icon="Message"
                size="large"
                inputmode="numeric"
                autocomplete="one-time-code"
                @input="onCodeInput"
              />
              <el-button
                native-type="button"
                class="send-code-btn"
                @click="handleSendCode"
                :disabled="countdown > 0 || sendingCode || !turnstileToken"
                :loading="sendingCode"
              >
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </el-button>
            </div>
          </el-form-item>

          <el-form-item class="action-item">
            <div class="step-actions">
              <el-button native-type="button" size="large" class="back-step-btn" @click="handleEditEmail">
                上一步
              </el-button>
              <el-button type="primary" native-type="submit" size="large" class="submit-btn step-primary-btn">
                下一步
              </el-button>
            </div>
          </el-form-item>
        </template>

        <template v-else>
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

          <!-- 用户协议 -->
          <el-form-item>
            <el-checkbox v-model="agreeTerms" class="agree-checkbox">
              <span class="agree-text">
                我已阅读并同意
                <a href="#" @click.prevent class="agree-link">《用户协议》</a>
              </span>
            </el-checkbox>
          </el-form-item>

          <el-form-item class="action-item">
            <div class="step-actions">
              <el-button native-type="button" size="large" class="back-step-btn" @click="currentStep = 2">
                上一步
              </el-button>
              <el-button
                type="primary"
                native-type="submit"
                :loading="loading"
                :disabled="loading"
                size="large"
                class="submit-btn step-primary-btn"
              >
                <span v-if="!loading">创建账号</span>
                <span v-else>注册中...</span>
              </el-button>
            </div>
          </el-form-item>
        </template>
      </div>
    </el-form>

    <div class="card-footer animate-fade-up" style="animation-delay: 600ms">
      <span>已有账号？</span>
      <router-link to="/login" class="footer-link">立即登录</router-link>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { User, Message, Lock } from '@element-plus/icons-vue'
import { isAxiosError } from 'axios'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import StatsPanel from '@/components/auth/StatsPanel.vue'
import { emailApi } from '@/api/user'
import '@/styles/auth.css'

const registerStats = [
  { value: '12,860', label: '已注册用户', fillWidth: '82%' },
  { value: '128', label: '今日新增', color: 'green' as const, fillWidth: '65%' },
  { value: '2,430', label: '已创建 Key', color: 'purple' as const, fillWidth: '72%' },
]

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
const countdownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const turnstileRef = ref<{ reset: () => void } | null>(null)
const isValidEmail = ref(false)
const turnstileToken = ref('')
const agreeTerms = ref(false)
const codeSent = ref(false)
const currentStep = ref(1)
const verificationVersion = ref(0)
const turnstileConsumed = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
})

const code = ref('')
const stepLabels = ['账户信息', '邮箱验证', '设置密码']
const stepTitle = computed(() => stepLabels[currentStep.value - 1])
const stepProgress = computed(() => Math.round((currentStep.value / stepLabels.length) * 100))
const isWaitingTurnstileReset = computed(() => turnstileConsumed.value && countdown.value > 0)

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

function checkEmailFormat() {
  syncEmailValidity()
  if (!isValidEmail.value && form.email) {
    ElMessage.warning('请输入有效的邮箱格式')
  }
}

function syncEmailValidity() {
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
  isValidEmail.value = emailRe.test(form.email)
  return isValidEmail.value
}

function validateAccountInfo() {
  if (!form.username) {
    ElMessage.warning('请输入用户名')
    return false
  }
  if (!form.email) {
    ElMessage.warning('请输入邮箱')
    return false
  }
  if (!syncEmailValidity()) {
    ElMessage.warning('请输入有效的邮箱格式')
    return false
  }
  return true
}

function validateVerificationCode() {
  if (!code.value) {
    ElMessage.warning('请输入验证码')
    return false
  }
  if (!/^\d{6}$/.test(code.value)) {
    ElMessage.warning('请输入6位验证码')
    return false
  }
  if (!codeSent.value) {
    ElMessage.warning('请先获取验证码')
    return false
  }
  return true
}

function validatePasswordInfo() {
  if (!form.password || !form.confirmPassword) {
    ElMessage.warning('请填写密码和确认密码')
    return false
  }
  if (form.password.length < 8) {
    ElMessage.warning('密码至少8位')
    return false
  }
  if (form.password !== form.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return false
  }
  if (!agreeTerms.value) {
    ElMessage.warning('请阅读并同意用户协议')
    return false
  }
  return true
}

function handlePrimaryAction() {
  if (currentStep.value === 1) {
    if (validateAccountInfo()) currentStep.value = 2
    return
  }

  if (currentStep.value === 2) {
    if (validateVerificationCode()) currentStep.value = 3
    return
  }

  handleRegister()
}

function resetVerificationState() {
  verificationVersion.value++
  code.value = ''
  codeSent.value = false
  turnstileToken.value = ''
  turnstileConsumed.value = false
  sendingCode.value = false
  countdown.value = 0
  clearCountdownTimer()
  turnstileRef.value?.reset()
}

function handleEditEmail() {
  resetVerificationState()
  currentStep.value = 1
}

// 验证码输入处理
function onCodeInput(value: string | number) {
  code.value = String(value ?? '').replace(/\D/g, '').slice(0, 6)
}

function resetTurnstile() {
  turnstileToken.value = ''
  turnstileConsumed.value = false
  turnstileRef.value?.reset()
}

function clearCountdownTimer() {
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value)
    countdownTimer.value = null
  }
}

function finishCodeCountdown(version: number) {
  clearCountdownTimer()
  countdown.value = 0
  if (version === verificationVersion.value) {
    resetTurnstile()
  }
}

function startCodeCountdown(version: number) {
  countdown.value = 60
  clearCountdownTimer()
  countdownTimer.value = setInterval(() => {
    if (version !== verificationVersion.value) {
      clearCountdownTimer()
      return
    }

    countdown.value = Math.max(countdown.value - 1, 0)
    if (countdown.value <= 0) {
      finishCodeCountdown(version)
    }
  }, 1000)
}

function handleTurnstileExpired() {
  turnstileToken.value = ''
  if (isWaitingTurnstileReset.value) return
  ElMessage.warning('人机验证已过期，请重新验证')
}

function handleTurnstileError() {
  turnstileToken.value = ''
  if (isWaitingTurnstileReset.value) return
  ElMessage.error('人机验证失败，请重试')
}

function handleTurnstileTimeout() {
  turnstileToken.value = ''
  if (isWaitingTurnstileReset.value) return
  ElMessage.warning('人机验证已超时，请重新验证')
}

function getApiError(error: unknown) {
  if (isAxiosError<{ error?: { code?: string; message?: string } }>(error)) {
    return error.response?.data?.error
  }
  return undefined
}

function handleSendCode() {
  if (countdown.value > 0 || sendingCode.value) {
    return
  }
  if (!form.email || !syncEmailValidity()) {
    ElMessage.warning('请输入有效的邮箱')
    return
  }
  if (!turnstileToken.value) {
    ElMessage.warning('请先完成人机验证')
    return
  }
  sendCode()
}

// 发送验证码
async function sendCode() {
  const version = verificationVersion.value
  const email = form.email
  const token = turnstileToken.value
  if (!token) {
    ElMessage.warning('请先完成人机验证')
    return
  }

  sendingCode.value = true
  try {
    await emailApi.sendCode({ email, turnstileToken: token })
    if (version !== verificationVersion.value || email !== form.email) return
    ElMessage.success('验证码已发送到您的邮箱')
    codeSent.value = true
    turnstileToken.value = ''
    turnstileConsumed.value = true
    startCodeCountdown(version)
  } catch (error: unknown) {
    if (version !== verificationVersion.value) return
    ElMessage.error(getApiError(error)?.message || '发送失败')
    resetTurnstile()
  } finally {
    if (version === verificationVersion.value) sendingCode.value = false
  }
}

// 注册
async function handleRegister() {
  if (!validateAccountInfo()) {
    currentStep.value = 1
    return
  }
  if (!validateVerificationCode()) {
    currentStep.value = 2
    return
  }
  if (!validatePasswordInfo()) return

  loading.value = true
  try {
    // 先验证邮箱验证码
    await emailApi.verifyCode({ email: form.email, code: code.value })
    // 再注册
    await authStore.register(form.username, form.email, form.password)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (error: unknown) {
    const errorData = getApiError(error)
    if (errorData?.code === 'REGISTRATION_CLOSED') { router.push('/register-closed'); return }
    ElMessage.error(errorData?.message || '注册失败')
  } finally { loading.value = false }
}

onBeforeUnmount(() => {
  clearCountdownTimer()
})
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

/* ===== 轻量步骤提示 ===== */
.step-summary {
  margin: -6px 0 18px;
  text-align: center;
}

.step-summary span {
  display: block;
  margin-bottom: 10px;
  font-size: 13px;
  color: var(--c-text-sub);
}

.step-progress {
  width: 100%;
  height: 2px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--c-border-sub);
}

.step-progress span {
  display: block;
  height: 100%;
  margin: 0;
  border-radius: inherit;
  background: linear-gradient(90deg, #6366f1, #8b5cf6);
  transition: width 0.3s ease;
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

/* ===== 分步表单 ===== */
.step-panel {
  display: flex;
  flex-direction: column;
  min-height: 226px;
}

.action-item {
  margin-top: auto;
  padding-top: 4px;
}

.step-actions {
  display: flex;
  width: 100%;
  gap: 12px;
}

.step-actions .submit-btn {
  width: auto;
  flex: 1;
}

.back-step-btn {
  flex: 0 0 108px;
  height: 44px;
  border-radius: 10px;
  font-size: 14px;
  color: var(--c-text-sub);
  background: var(--c-input-bg);
  border: 1px solid var(--c-input-border);
  transition: all 0.3s ease;
}

.back-step-btn:hover {
  color: var(--c-text);
  border-color: var(--c-input-hover);
  background: rgba(255, 255, 255, 0.05);
}

.email-review {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  margin-bottom: 16px;
  padding: 11px 14px;
  border: 1px solid var(--c-input-border);
  border-radius: 10px;
  background: var(--c-input-bg);
}

.email-review-content {
  min-width: 0;
}

.email-review-content span {
  display: block;
  margin-bottom: 3px;
  font-size: 12px;
  color: var(--c-text-muted);
}

.email-review-content strong {
  display: block;
  overflow-wrap: anywhere;
  font-size: 14px;
  font-weight: 500;
  color: var(--c-text-secondary);
}

.change-email-btn {
  flex-shrink: 0;
  padding: 0;
  color: #a5b4fc;
}

.change-email-btn:hover {
  color: #c7d2fe;
}

.turnstile-form-item :deep(.el-form-item__content) {
  line-height: normal;
}

.turnstile-sent-message {
  width: 100%;
  margin: 0;
  padding: 12px 14px;
  border: 1px solid rgba(34, 197, 94, 0.22);
  border-radius: 10px;
  background: rgba(34, 197, 94, 0.08);
  color: #86efac;
  font-size: 13px;
  line-height: 1.5;
}

/* ===== 验证码行 ===== */
.captcha-row {
  display: flex;
  align-items: stretch;
  gap: 10px;
  width: 100%;
}

.captcha-row .el-input {
  flex: 1;
  min-width: 0;
}

.send-code-btn {
  flex-shrink: 0;
  min-width: 104px;
  height: 40px;
  padding: 0 14px;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  border-radius: 10px;
  font-size: 13px;
  transition: all 0.3s ease;
}

.send-code-btn:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.3);
  color: #c7d2fe;
}

.send-code-btn:disabled {
  opacity: 0.65;
  color: var(--c-text-muted);
  background: var(--c-input-bg);
  border-color: var(--c-input-border);
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

@media (max-width: 480px) {
  .step-panel {
    min-height: 252px;
  }

  .captcha-row,
  .step-actions,
  .password-row,
  .confirm-password-row {
    flex-direction: column;
    align-items: stretch;
  }

  .send-code-btn,
  .back-step-btn,
  .step-actions .submit-btn {
    width: 100%;
    flex: none;
  }

  .password-strength-inline {
    align-self: flex-start;
  }

  .confirm-password-error {
    white-space: normal;
  }
}

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
