<template>
  <AuthLayout :showNavLinks="true" :showRegisterLink="true">
    <!-- 品牌展示区 -->
    <template #brand>
      <h1 class="brand-title">
        <span class="gradient-text">智能 AI API</span>
        <br />中转与管理平台
      </h1>
      <p class="brand-desc">支持 OpenAI、Claude、DeepSeek 等多渠道无缝接入</p>

      <!-- 产品预览面板 -->
      <StatsPanel
        title="平台概览"
        live-label="实时"
        :stats="loginStats"
      />

      <!-- 特性亮点 -->
      <div class="feature-list animate-fade-up" style="animation-delay: 700ms">
        <div class="feature-item">
          <div class="feature-icon icon-purple">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z"/><line x1="4" y1="22" x2="4" y2="15"/></svg>
          </div>
          <div class="feature-text">
            <strong>多渠道接入</strong>
            OpenAI / Claude / DeepSeek 一站管理
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-blue">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
          </div>
          <div class="feature-text">
            <strong>智能路由</strong>
            负载均衡 · 故障转移 · 延迟最优
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-green">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          </div>
          <div class="feature-text">
            <strong>安全加密</strong>
            API Key 加密存储 · 传输全程 HTTPS
          </div>
        </div>
      </div>
    </template>

    <!-- 登录卡片内容 -->
    <div class="card-header">
      <h2>欢迎回来</h2>
      <p class="card-desc">请登录您的账户以继续使用</p>
    </div>

    <el-form @submit.prevent="handleLogin" class="auth-form">
      <el-form-item>
        <el-input
          v-model="form.email"
          placeholder="请输入邮箱地址"
          :prefix-icon="Message"
          size="large"
        />
      </el-form-item>
      <el-form-item>
        <el-input
          v-model="form.password"
          type="password"
          placeholder="请输入密码"
          :prefix-icon="Lock"
          size="large"
          show-password
        />
      </el-form-item>
      <el-form-item class="turnstile-form-item">
        <TurnstileWidget
          ref="turnstileRef"
          v-model="turnstileToken"
          @expired="handleTurnstileExpired"
          @error="handleTurnstileError"
          @timeout="handleTurnstileTimeout"
        />
      </el-form-item>
      <el-form-item>
        <el-button
          native-type="submit"
          type="primary"
          :loading="loading"
          :disabled="loading || !turnstileToken"
          size="large"
          class="login-btn"
        >
          <span v-if="!loading">登 录</span>
          <span v-else>登录中...</span>
        </el-button>
      </el-form-item>
    </el-form>

    <div class="card-footer">
      <span>还没有账号？</span>
      <router-link to="/register" class="footer-link">立即注册</router-link>
      <span class="separator">|</span>
      <router-link to="/forgot-password" class="footer-link">忘记密码</router-link>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { Message, Lock } from '@element-plus/icons-vue'
import { isAxiosError } from 'axios'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import StatsPanel from '@/components/auth/StatsPanel.vue'
import '@/styles/auth.css'

const loginStats = [
  { value: '128,430', label: '今日请求量', fillWidth: '78%' },
  { value: '99.92%', label: '成功率', color: 'green' as const, fillWidth: '99%' },
  { value: '200+', label: '支持模型', color: 'purple' as const, fillWidth: '85%' },
]

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const turnstileToken = ref('')
const turnstileRef = ref<{ reset: () => void } | null>(null)

const form = reactive({ email: '', password: '' })

function resetTurnstile() {
  turnstileToken.value = ''
  turnstileRef.value?.reset()
}

function handleTurnstileExpired() {
  turnstileToken.value = ''
  ElMessage.warning('人机验证已过期，请重新验证')
}

function handleTurnstileError() {
  turnstileToken.value = ''
  ElMessage.error('人机验证失败，请重试')
}

function handleTurnstileTimeout() {
  turnstileToken.value = ''
  ElMessage.warning('人机验证已超时，请重新验证')
}

function getErrorMessage(error: unknown, fallback: string) {
  if (isAxiosError<{ error?: { message?: string } }>(error)) {
    return error.response?.data?.error?.message || fallback
  }
  return fallback
}

async function handleLogin() {
  if (!turnstileToken.value) {
    ElMessage.warning('请先完成人机验证')
    return
  }
  if (!form.email || !form.password) {
    ElMessage.warning('请填写邮箱和密码')
    return
  }
  loading.value = true
  try {
    await authStore.login(form.email, form.password, turnstileToken.value)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error: unknown) {
    ElMessage.error(getErrorMessage(error, '登录失败'))
    resetTurnstile()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* ===== 品牌展示区 ===== */
.brand-title {
  font-size: 28px;
  font-weight: 600;
  letter-spacing: 0;
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
  margin-bottom: 24px;
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
  transition: all 0.3s ease;
  padding: 4px 12px;
}

.auth-form :deep(.el-input__wrapper:hover) {
  border-color: var(--c-input-hover);
}

.auth-form :deep(.el-input__wrapper.is-focus) {
  border-color: var(--c-primary); box-shadow: 0 0 0 3px var(--c-focus-ring);
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

.turnstile-form-item :deep(.el-form-item__content) {
  line-height: normal;
}

/* ===== 登录按钮 ===== */
.login-btn {
  width: 100%;
  height: 44px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  background: var(--c-primary);
  color: #fff;
  border: none;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.login-btn::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent);
  display: none;
}

.login-btn:hover:not(:disabled) {
  background: var(--c-primary-hover);
  box-shadow: none;
  transform: translateY(-1px);
}

.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.login-btn:disabled {
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

.separator {
  margin: 0 10px;
  color: var(--c-border);
}
</style>
