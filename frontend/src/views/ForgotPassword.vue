<template>
  <AuthLayout :showLoginLink="true">
    <!-- 品牌展示区 -->
    <template #brand>
      <h1 class="brand-title">
        <span class="gradient-text">找回密码</span>
        <br />别担心，我们来帮您
      </h1>
      <p class="brand-desc">通过注册邮箱验证身份，安全快速地重置您的账户密码</p>

      <!-- 特性亮点 -->
      <div class="feature-list animate-fade-up" style="animation-delay: 600ms">
        <div class="feature-item">
          <div class="feature-icon icon-purple">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          </div>
          <div class="feature-text">
            <strong>安全验证</strong>
            邮箱验证 + 滑块验证双重保障
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-blue">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
          <div class="feature-text">
            <strong>快速重置</strong>
            链接有效期 1 小时，过期可重新申请
          </div>
        </div>
      </div>
    </template>

    <!-- 忘记密码卡片内容 -->
    <div class="card-header">
      <h2 class="animate-fade-up" style="animation-delay: 400ms">忘记密码</h2>
      <p class="card-desc animate-fade-up" style="animation-delay: 450ms">输入您的注册邮箱，我们将发送密码重置链接</p>
    </div>

    <div v-if="!codeSent" class="animate-fade-up" style="animation-delay: 500ms">
      <el-form @submit.prevent="handleSubmit" class="auth-form">
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
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" :disabled="!canSubmit" @click="handleSubmit" size="large" class="submit-btn">发送重置链接</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div v-else class="success-content animate-fade-up" style="animation-delay: 500ms">
      <div class="success-icon-wrapper">
        <el-icon class="success-icon"><CircleCheck /></el-icon>
      </div>
      <h3>重置链接已发送</h3>
      <p>我们已发送密码重置链接到 <strong>{{ form.email }}</strong></p>
      <p class="hint">请查收邮件并点击链接重置密码，链接有效期为1小时</p>
      <el-button type="primary" @click="$router.push('/login')" size="large" class="submit-btn" style="margin-top: 20px">返回登录</el-button>
    </div>

    <div class="card-footer animate-fade-up" style="animation-delay: 600ms">
      <router-link to="/login" class="footer-link">返回登录</router-link>
    </div>

    <SlideCaptcha v-model:visible="showCaptcha" @success="onCaptchaSuccess" ref="captchaRef" />
  </AuthLayout>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Message, Picture, CircleCheck } from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import request from '@/api/request'
import '@/styles/auth.css'

const router = useRouter()
const loading = ref(false)
const showCaptcha = ref(false)
const captchaVerified = ref(false)
const captchaRef = ref()
const codeSent = ref(false)
const isValidEmail = ref(false)
const captchaToken = ref('')

const form = reactive({ email: '' })
const canSubmit = computed(() => isValidEmail.value && captchaVerified.value)

function checkEmailFormat() {
  const emailRe = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
  isValidEmail.value = emailRe.test(form.email)
}

function onCaptchaSuccess(token: string) {
  captchaVerified.value = true
  captchaToken.value = token
}

async function handleSubmit() {
  if (!isValidEmail.value) { ElMessage.warning('请输入有效的邮箱'); return }
  if (!captchaVerified.value) { ElMessage.warning('请先完成安全验证'); showCaptcha.value = true; return }
  loading.value = true
  try {
    await request.post('/auth/forgot-password', { email: form.email, captcha_token: captchaToken.value || 'bypass' })
    codeSent.value = true
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '发送失败')
    captchaVerified.value = false
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
  color: #fff;
}

@media (min-width: 640px) {
  .brand-title { font-size: 36px; }
}

@media (min-width: 768px) {
  .brand-title { font-size: 44px; }
}

@media (min-width: 1024px) {
  .brand-title { font-size: 52px; }
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
  color: rgba(255, 255, 255, 0.35);
  margin: 0 0 24px;
  line-height: 1.6;
  max-width: 400px;
}

@media (min-width: 768px) {
  .brand-desc { font-size: 16px; }
}

/* ===== 卡片头部 ===== */
.card-header { text-align: center; margin-bottom: 24px; }
.card-header h2 { margin: 0 0 6px; font-size: 22px; font-weight: 600; color: #fff; }
.card-desc { margin: 0; font-size: 14px; color: rgba(255, 255, 255, 0.3); }

/* ===== 表单 ===== */
.auth-form :deep(.el-form-item) { margin-bottom: 16px; }
.auth-form :deep(.el-form-item:last-child) { margin-bottom: 0; }
.auth-form :deep(.el-input__wrapper) { background: rgba(255, 255, 255, 0.03); border: 1px solid rgba(255, 255, 255, 0.06); border-radius: 10px; box-shadow: none; transition: all 0.3s; padding: 4px 12px; }
.auth-form :deep(.el-input__wrapper:hover) { border-color: rgba(255, 255, 255, 0.12); }
.auth-form :deep(.el-input__wrapper.is-focus) { border-color: rgba(99, 102, 241, 0.4); box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.08), 0 0 20px rgba(99, 102, 241, 0.04); }
.auth-form :deep(.el-input__inner) { color: #fff; font-size: 14px; }
.auth-form :deep(.el-input__inner::placeholder) { color: rgba(255, 255, 255, 0.2); }
.auth-form :deep(.el-input__prefix .el-icon) { color: rgba(255, 255, 255, 0.2); font-size: 16px; }

/* ===== 验证码 ===== */
.captcha-wrapper { display: flex; align-items: center; gap: 10px; padding: 11px 16px; border: 1px solid rgba(255, 255, 255, 0.06); border-radius: 10px; cursor: pointer; transition: all 0.3s; background: rgba(255, 255, 255, 0.03); width: 100%; }
.captcha-wrapper:hover { border-color: rgba(255, 255, 255, 0.12); background: rgba(255, 255, 255, 0.05); }
.captcha-wrapper.captcha-verified { border-color: rgba(34, 197, 94, 0.3); background: rgba(34, 197, 94, 0.06); }
.captcha-icon { font-size: 18px; color: rgba(255, 255, 255, 0.25); flex-shrink: 0; }
.captcha-text { flex: 1; font-size: 14px; color: rgba(255, 255, 255, 0.3); }
.captcha-check { color: #22c55e; font-size: 18px; }

/* ===== 提交按钮 ===== */
.submit-btn { width: 100%; height: 44px; border-radius: 10px; font-size: 14px; font-weight: 500; background: linear-gradient(135deg, #6366f1, #8b5cf6); color: #fff; border: none; transition: all 0.3s ease; position: relative; overflow: hidden; }
.submit-btn::after { content: ''; position: absolute; top: 0; left: -100%; width: 100%; height: 100%; background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent); animation: btn-glow 3s ease-in-out infinite; }
.submit-btn:hover:not(:disabled) { background: linear-gradient(135deg, #818cf8, #a78bfa); box-shadow: 0 4px 20px rgba(99, 102, 241, 0.3); transform: translateY(-1px); }
.submit-btn:active:not(:disabled) { transform: translateY(0); }
.submit-btn:disabled { opacity: 0.3; cursor: not-allowed; }

/* ===== 成功态 ===== */
.success-content { text-align: center; padding: 20px 0; }
.success-icon-wrapper { position: relative; display: inline-flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.success-icon { font-size: 56px; color: #22c55e; animation: envelope-fly 0.6s cubic-bezier(0.22, 1, 0.36, 1) forwards; }
.success-icon-wrapper::before { content: ''; position: absolute; width: 80px; height: 80px; border-radius: 50%; background: rgba(34, 197, 94, 0.08); animation: pulse-ring 2s ease-in-out infinite; }
.success-content h3 { margin: 0 0 16px 0; font-size: 18px; font-weight: 600; color: #fff; }
.success-content p { color: rgba(255, 255, 255, 0.5); margin: 8px 0; font-size: 14px; }
.success-content strong { color: #fff; }
.hint { color: rgba(255, 255, 255, 0.25) !important; font-size: 13px; }

/* ===== 底部链接 ===== */
.card-footer { text-align: center; margin-top: 20px; padding-top: 16px; border-top: 1px solid rgba(255, 255, 255, 0.04); font-size: 13px; }
.footer-link { color: rgba(255, 255, 255, 0.4); text-decoration: none; font-weight: 500; transition: color 0.2s ease; }
.footer-link:hover { color: #fff; }
</style>