<template>
  <div class="login-page">
    <!-- 背景装饰：渐变光晕 -->
    <div class="bg-decoration">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
      <div class="orb orb-3"></div>
      <div class="grid-lines"></div>
    </div>

    <!-- 导航栏 -->
    <nav class="navbar">
      <div class="nav-left">
        <span class="nav-logo animate-fade-up" style="animation-delay: 0ms">gAPI Platform</span>
      </div>

      <div class="nav-center">
        <a
          v-for="(link, i) in navLinks"
          :key="link.label"
          :href="link.href"
          class="nav-link animate-fade-up"
          :style="{ animationDelay: (100 + i * 50) + 'ms' }"
        >{{ link.label }}</a>
      </div>

      <div class="nav-right">
        <router-link
          to="/register"
          class="nav-register animate-fade-up"
          style="animation-delay: 350ms"
        >注册</router-link>

        <button
          class="nav-hamburger animate-fade-up"
          style="animation-delay: 400ms"
          @click="menuOpen = !menuOpen"
          aria-label="菜单"
        >
          <div class="hamburger-icon" :class="{ active: menuOpen }">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </button>
      </div>
    </nav>

    <!-- 移动端菜单 -->
    <div class="mobile-menu" :class="{ open: menuOpen }">
      <a
        v-for="(link, i) in navLinks"
        :key="link.label"
        :href="link.href"
        class="mobile-link"
        :style="{ transitionDelay: menuOpen ? (i * 50) + 'ms' : '0ms' }"
        @click="menuOpen = false"
      >{{ link.label }}</a>
      <div class="mobile-actions">
        <router-link to="/register" class="mobile-action-btn" @click="menuOpen = false">注册账号</router-link>
      </div>
    </div>

    <!-- 主内容区 -->
    <main class="auth-content">
      <div class="auth-wrapper animate-fade-up" style="animation-delay: 300ms">
        <!-- 品牌展示 -->
        <div class="brand-section animate-fade-up" style="animation-delay: 400ms">
          <h1 class="brand-title">
            <span class="gradient-text">智能 AI API</span>
            <br />中转与管理平台
          </h1>
          <p class="brand-desc">支持 OpenAI、Claude、DeepSeek 等多渠道无缝接入</p>
        </div>

        <!-- 登录卡片 -->
        <div class="glass-card animate-fade-up" style="animation-delay: 500ms">
          <div class="card-inner">
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
                  @click="showCaptcha = true"
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
          </div>
        </div>
      </div>
    </main>

    <!-- 版权信息 -->
    <footer class="page-footer animate-fade-up" style="animation-delay: 700ms">
      © {{ new Date().getFullYear() }} gAPI Platform. All rights reserved.
    </footer>

    <!-- 滑块验证码 -->
    <SlideCaptcha
      v-model:visible="showCaptcha"
      @success="onCaptchaSuccess"
      ref="captchaRef"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'
import {
  Message,
  Lock,
  Picture,
  CircleCheck
} from '@element-plus/icons-vue'
import SlideCaptcha from '@/components/SlideCaptcha.vue'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const showCaptcha = ref(false)
const captchaVerified = ref(false)
const captchaRef = ref()
const menuOpen = ref(false)

const form = reactive({ email: '', password: '' })

const navLinks = [
  { label: '功能', href: '#' },
  { label: '模型', href: '#' },
  { label: '定价', href: '#' },
  { label: '文档', href: '#' }
]

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

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
</style>

<style scoped>
/* ===== 页面布局 ===== */
.login-page {
  position: relative;
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #000;
  font-family: 'Inter', sans-serif;
  overflow: hidden;
  color: #fff;
}

/* ===== 背景装饰 ===== */
.bg-decoration {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

/* 渐变光晕 */
.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.15;
}

.orb-1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, #6366f1, transparent 70%);
  top: -200px;
  left: -150px;
  animation: orb-float-1 25s ease-in-out infinite;
}

.orb-2 {
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, #8b5cf6, transparent 70%);
  bottom: -200px;
  right: -100px;
  animation: orb-float-2 30s ease-in-out infinite;
}

.orb-3 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #3b82f6, transparent 70%);
  top: 40%;
  left: 50%;
  animation: orb-float-3 20s ease-in-out infinite;
}

@keyframes orb-float-1 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(40px, 30px) scale(1.08); }
}

@keyframes orb-float-2 {
  0%, 100% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(-30px, -20px) scale(1.05); }
}

@keyframes orb-float-3 {
  0%, 100% { transform: translate(-50%, 0) scale(1); }
  50% { transform: translate(-50%, 30px) scale(0.95); }
}

/* 网格线 */
.grid-lines {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px);
  background-size: 80px 80px;
}

/* ===== 渐入动画 ===== */
@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(20px);
    filter: blur(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
    filter: blur(0);
  }
}

.animate-fade-up {
  animation: fadeUp 0.8s ease-out forwards;
  opacity: 0;
}

/* ===== 导航栏 ===== */
.navbar {
  position: relative;
  z-index: 50;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
}

@media (min-width: 768px) {
  .navbar {
    padding: 16px 48px;
  }
}

.nav-left {
  flex-shrink: 0;
}

.nav-logo {
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.04em;
  color: #fff;
}

@media (min-width: 768px) {
  .nav-logo {
    font-size: 28px;
  }
}

/* 桌面导航链接 */
.nav-center {
  display: none;
}

@media (min-width: 1024px) {
  .nav-center {
    display: flex;
    align-items: center;
    gap: 28px;
  }
}

.nav-link {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.5);
  text-decoration: none;
  transition: color 0.2s ease;
}

.nav-link:hover {
  color: rgba(255, 255, 255, 0.9);
}

/* 导航右侧 */
.nav-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav-register {
  display: none;
  padding: 8px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.2s ease;
}

@media (min-width: 640px) {
  .nav-register {
    display: inline-flex;
    align-items: center;
  }
}

.nav-register:hover {
  border-color: rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.05);
}

/* 汉堡按钮 */
.nav-hamburger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  cursor: pointer;
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: border-color 0.2s ease;
}

.nav-hamburger:hover {
  border-color: rgba(255, 255, 255, 0.25);
}

@media (min-width: 1024px) {
  .nav-hamburger {
    display: none;
  }
}

.hamburger-icon {
  display: flex;
  flex-direction: column;
  gap: 5px;
  width: 18px;
}

.hamburger-icon span {
  display: block;
  width: 18px;
  height: 1.5px;
  background: rgba(255, 255, 255, 0.7);
  border-radius: 1px;
  transition: all 0.4s ease-out;
  transform-origin: center;
}

.hamburger-icon.active span:nth-child(1) {
  transform: translateY(6.5px) rotate(45deg);
}

.hamburger-icon.active span:nth-child(2) {
  opacity: 0;
  transform: scaleX(0.5);
}

.hamburger-icon.active span:nth-child(3) {
  transform: translateY(-6.5px) rotate(-45deg);
}

/* ===== 移动端菜单 ===== */
.mobile-menu {
  position: absolute;
  top: 72px;
  left: 0;
  right: 0;
  z-index: 40;
  background: rgba(10, 10, 10, 0.95);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  padding: 16px;
  transform: translateY(-16px);
  opacity: 0;
  pointer-events: none;
  transition: all 0.4s ease-out;
}

.mobile-menu.open {
  transform: translateY(0);
  opacity: 1;
  pointer-events: auto;
}

@media (min-width: 1024px) {
  .mobile-menu {
    display: none;
  }
}

.mobile-link {
  display: block;
  padding: 12px;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.6);
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.3s ease-out;
  transform: translateX(-8px);
}

.mobile-menu.open .mobile-link {
  transform: translateX(0);
}

.mobile-link:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
}

.mobile-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

@media (min-width: 640px) {
  .mobile-actions {
    display: none;
  }
}

.mobile-action-btn {
  display: block;
  padding: 12px;
  font-size: 15px;
  font-weight: 500;
  color: #fff;
  text-decoration: none;
  border-radius: 8px;
  text-align: center;
  background: rgba(255, 255, 255, 0.08);
  transition: background 0.2s ease;
}

.mobile-action-btn:hover {
  background: rgba(255, 255, 255, 0.12);
}

/* ===== 主内容区 ===== */
.auth-content {
  position: relative;
  z-index: 10;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px 32px;
}

.auth-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 40px;
  width: 100%;
  max-width: 1000px;
}

@media (min-width: 768px) {
  .auth-wrapper {
    flex-direction: row;
    align-items: center;
    gap: 64px;
  }
}

/* ===== 品牌展示 ===== */
.brand-section {
  flex: 1;
  text-align: center;
}

@media (min-width: 768px) {
  .brand-section {
    text-align: left;
  }
}

.brand-title {
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.04em;
  line-height: 1.2;
  margin: 0 0 16px;
  color: #fff;
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
  background: linear-gradient(135deg, #fff 0%, rgba(255, 255, 255, 0.6) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-desc {
  font-size: 15px;
  color: rgba(255, 255, 255, 0.4);
  margin: 0;
  line-height: 1.6;
  max-width: 400px;
}

@media (min-width: 768px) {
  .brand-desc {
    font-size: 16px;
  }
}

/* ===== Glass Card ===== */
.glass-card {
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  position: relative;
  overflow: hidden;
  width: 100%;
  max-width: 400px;
}

.glass-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.12) 0%,
    rgba(255, 255, 255, 0.02) 30%,
    rgba(255, 255, 255, 0) 60%,
    rgba(255, 255, 255, 0.02) 80%,
    rgba(255, 255, 255, 0.08) 100%
  );
  -webkit-mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
}

.card-inner {
  padding: 28px 24px 24px;
}

@media (min-width: 640px) {
  .card-inner {
    padding: 32px 28px 28px;
  }
}

/* ===== 卡片头部 ===== */
.card-header {
  text-align: center;
  margin-bottom: 24px;
}

.card-header h2 {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 600;
  color: #fff;
}

.card-desc {
  margin: 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.35);
}

/* ===== 表单 ===== */
.auth-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.auth-form :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

.auth-form :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  box-shadow: none;
  transition: all 0.3s ease;
  padding: 4px 12px;
}

.auth-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.15);
}

.auth-form :deep(.el-input__wrapper.is-focus) {
  border-color: rgba(255, 255, 255, 0.25);
  box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.04);
}

.auth-form :deep(.el-input__inner) {
  color: #fff;
  font-size: 14px;
}

.auth-form :deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.25);
}

.auth-form :deep(.el-input__prefix .el-icon) {
  color: rgba(255, 255, 255, 0.25);
  font-size: 16px;
}

.auth-form :deep(.el-input__suffix .el-icon) {
  color: rgba(255, 255, 255, 0.25);
}

/* ===== 验证码 ===== */
.captcha-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: rgba(255, 255, 255, 0.04);
  width: 100%;
}

.captcha-wrapper:hover {
  border-color: rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.06);
}

.captcha-wrapper.captcha-verified {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.06);
}

.captcha-icon {
  font-size: 18px;
  color: rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}

.captcha-verified .captcha-icon {
  color: #22c55e;
}

.captcha-text {
  flex: 1;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.35);
}

.captcha-verified .captcha-text {
  color: #22c55e;
}

/* ===== 登录按钮 ===== */
.login-btn {
  width: 100%;
  height: 42px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  background: #fff;
  color: #000;
  border: none;
  transition: all 0.2s ease;
}

.login-btn:hover:not(:disabled) {
  background: rgba(229, 229, 229, 1);
}

.login-btn:disabled {
  opacity: 0.25;
  cursor: not-allowed;
}

/* ===== 底部链接 ===== */
.card-footer {
  text-align: center;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.3);
  font-size: 13px;
}

.footer-link {
  color: rgba(255, 255, 255, 0.5);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: #fff;
}

.separator {
  margin: 0 10px;
  color: rgba(255, 255, 255, 0.1);
}

/* ===== 版权信息 ===== */
.page-footer {
  position: relative;
  z-index: 10;
  text-align: center;
  padding: 0 16px 16px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.2);
}
</style>