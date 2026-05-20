·<template>
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
      <el-form-item>
        <div
          class="captcha-wrapper"
          :class="{ 'captcha-verified': captchaVerified }"
          @click="!captchaVerified && (showCaptcha = true)"
        >
          <el-icon class="captcha-icon">
            <CircleCheck v-if="captchaVerified" />
            <Picture v-else />
          </el-icon>
          <span class="captcha-text">
            {{ captchaVerified ? '安全验证已通过' : '点击进行安全验证' }}
          </span>
        </div>
      </el-form-item>
      <el-form-item>
        <el-button
          native-type="submit"
          type="primary"
          :loading="loading"
          :disabled="!captchaVerified"
          @click="handleLogin"
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

    <!-- 滑块验证码 -->
    <SlideCaptcha
      v-model:visible="showCaptcha"
      @success="onCaptchaSuccess"
      ref="captchaRef"
    />
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import { Message, Lock, Picture, CircleCheck } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'
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
    ElMessage.warning('请填写邮箱和密码')
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

/* ===== 验证码 ===== */
.captcha-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 16px;
  border: 1px solid var(--c-input-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: var(--c-input-bg);
  width: 100%;
  user-select: none;
  -webkit-user-select: none;
}

.captcha-wrapper:hover {
  border-color: var(--c-input-hover);
  background: rgba(255, 255, 255, 0.05);
}

.captcha-wrapper.captcha-verified {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.06);
  cursor: default;
}

.captcha-wrapper.captcha-verified:hover {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.06);
}

.captcha-icon {
  font-size: 18px;
  color: var(--c-icon-active);
  flex-shrink: 0;
}

.captcha-verified .captcha-icon {
  color: #22c55e;
}

.captcha-text {
  flex: 1;
  font-size: 14px;
  color: var(--c-text-sub);
}

.captcha-verified .captcha-text {
  color: #22c55e;
}

/* ===== 登录按钮 ===== */
.login-btn {
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

.login-btn::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent);
  animation: btn-glow 3s ease-in-out infinite;
}

.login-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #818cf8, #a78bfa);
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.3);
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
