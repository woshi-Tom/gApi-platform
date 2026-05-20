<template>
  <div class="auth-page">
    <!-- 动态星空背景 -->
    <div class="bg-decoration">
      <!-- 渐变光晕 -->
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
      <div class="orb orb-3"></div>
      <!-- 网格线 -->
      <div class="grid-lines"></div>
      <!-- 噪点纹理 -->
      <div class="noise-overlay"></div>
      <!-- 星空粒子 -->
      <div class="stars">
        <div v-for="n in 120" :key="n" class="star" :style="getStarStyle(n)"></div>
      </div>
    </div>

    <!-- 导航栏 -->
    <nav class="navbar">
      <div class="nav-left">
        <router-link :to="navLogoTo" class="nav-logo animate-fade-up" style="animation-delay: 0ms">
          <span class="logo-icon-wrapper">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2L2 7V17L12 22L22 17V7L12 2Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/>
              <path d="M12 22V12" stroke="currentColor" stroke-width="1.5"/>
              <path d="M22 7L12 12L2 7" stroke="currentColor" stroke-width="1.5"/>
              <path d="M17 4.5L7 9.5" stroke="currentColor" stroke-width="1.2" opacity="0.5"/>
            </svg>
          </span>
          gAPI Platform
        </router-link>
      </div>

      <!-- 桌面端导航链接（仅 Login 页显示） -->
      <div class="nav-center" v-if="showNavLinks">
        <a
          v-for="(link, i) in navLinks"
          :key="link.label"
          :href="link.href"
          class="nav-link animate-fade-up"
          :style="{ animationDelay: (100 + i * 50) + 'ms' }"
        >{{ link.label }}</a>
      </div>

      <div class="nav-right">
        <slot name="nav-right">
          <router-link
            v-if="showRegisterLink"
            to="/register"
            class="nav-register animate-fade-up"
            style="animation-delay: 350ms"
          >注册</router-link>
          <router-link
            v-if="showLoginLink"
            to="/login"
            class="nav-register animate-fade-up"
            style="animation-delay: 150ms"
          >已有账号？登录</router-link>
        </slot>

        <!-- 汉堡菜单按钮 -->
        <button
          v-if="showNavLinks"
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
    <div v-if="showNavLinks" class="mobile-menu" :class="{ open: menuOpen }">
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
      <div class="auth-wrapper animate-fade-up" style="animation-delay: 200ms">
        <!-- 左侧品牌展示（可选） -->
        <div v-if="$slots.brand" class="brand-section animate-fade-up" :style="{ animationDelay: brandDelay + 'ms' }">
          <slot name="brand" />
        </div>

        <!-- 右侧卡片 -->
        <div class="glass-card animate-fade-up" :style="{ animationDelay: cardDelay + 'ms' }">
          <!-- Shimmer 高光条 -->
          <div class="card-shimmer"></div>
          <div class="card-inner">
            <slot />
          </div>
        </div>
      </div>
    </main>

    <!-- 版权信息 -->
    <footer class="page-footer animate-fade-up" style="animation-delay: 700ms">
      © {{ new Date().getFullYear() }} gAPI Platform. All rights reserved.
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  showNavLinks?: boolean
  showRegisterLink?: boolean
  showLoginLink?: boolean
  navLogoTo?: string
  brandDelay?: number
  cardDelay?: number
}>(), {
  showNavLinks: false,
  showRegisterLink: false,
  showLoginLink: false,
  navLogoTo: '/login',
  brandDelay: 400,
  cardDelay: 500,
})

const menuOpen = ref(false)

const navLinks = [
  { label: '功能', href: '#' },
  { label: '模型', href: '#' },
  { label: '定价', href: '#' },
  { label: '文档', href: '#' }
]

// 为每个星星生成固定的随机样式
function getStarStyle(index: number) {
  // 使用伪随机（基于 index 的确定性计算）
  const seed = index * 7919 + 1
  const x = (seed * 13) % 100
  const y = (seed * 17) % 100
  const size = 1 + ((seed * 23) % 3)
  const duration = 3 + ((seed * 29) % 8)
  const delay = ((seed * 31) % 5)

  return {
    left: `${x}%`,
    top: `${y}%`,
    width: `${size}px`,
    height: `${size}px`,
    animationDuration: `${duration}s`,
    animationDelay: `${delay}s`,
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');

/* ===== 页面布局 ===== */
.auth-page {
  position: relative;
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #06060a;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  color: #fff;
}

/* ===== 背景装饰 ===== */
.bg-decoration {
  position: fixed;
  top: -100%;
  left: -100%;
  width: 300%;
  height: 300%;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
  animation: bg-rotate 180s linear infinite;
}

@keyframes bg-rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 渐变光晕 */
.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.12;
}

.orb-1 {
  width: 650px;
  height: 650px;
  background: radial-gradient(circle, #6366f1, transparent 70%);
  top: -250px;
  left: -200px;
  animation: orb-float-1 25s ease-in-out infinite;
}

.orb-2 {
  width: 550px;
  height: 550px;
  background: radial-gradient(circle, #8b5cf6, transparent 70%);
  bottom: -250px;
  right: -150px;
  animation: orb-float-2 30s ease-in-out infinite;
}

.orb-3 {
  width: 450px;
  height: 450px;
  background: radial-gradient(circle, #3b82f6, transparent 70%);
  top: 35%;
  left: 50%;
  animation: orb-float-3 20s ease-in-out infinite;
}

/* 网格线 */
.grid-lines {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.015) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.015) 1px, transparent 1px);
  background-size: 80px 80px;
  animation: grid-scroll 40s linear infinite;
}

@keyframes grid-scroll {
  from { background-position: 0 0; }
  to { background-position: 80px 80px; }
}

/* 噪点纹理 */
.noise-overlay {
  position: absolute;
  inset: 0;
  opacity: 0.03;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='1'/%3E%3C/svg%3E");
  background-repeat: repeat;
  background-size: 256px 256px;
  animation: noise-drift 12s linear infinite;
}

@keyframes noise-drift {
  from { background-position: 0 0; }
  to { background-position: 256px 256px; }
}

/* 星空粒子 */
.stars {
  position: absolute;
  inset: -100px;
  animation: stars-drift 80s linear infinite;
}

@keyframes stars-drift {
  0% { transform: translate(0, 0); }
  25% { transform: translate(-40px, -30px); }
  50% { transform: translate(-15px, -50px); }
  75% { transform: translate(25px, -20px); }
  100% { transform: translate(0, 0); }
}

.star {
  position: absolute;
  background: #fff;
  border-radius: 50%;
  animation: twinkle ease-in-out infinite;
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
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 22px;
  font-weight: 600;
  letter-spacing: -0.04em;
  color: #fff;
  text-decoration: none;
  transition: opacity 0.2s;
}

@media (min-width: 768px) {
  .nav-logo {
    font-size: 26px;
  }
}

.nav-logo:hover {
  opacity: 0.85;
}

.logo-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #818cf8;
  opacity: 0.9;
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
  color: rgba(255, 255, 255, 0.4);
  text-decoration: none;
  transition: color 0.2s ease;
  position: relative;
}

.nav-link::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 0;
  width: 0;
  height: 1px;
  background: rgba(255, 255, 255, 0.3);
  transition: width 0.3s ease;
}

.nav-link:hover {
  color: rgba(255, 255, 255, 0.85);
}

.nav-link:hover::after {
  width: 100%;
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
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: all 0.2s ease;
  background: rgba(255, 255, 255, 0.02);
}

@media (min-width: 640px) {
  .nav-register {
    display: inline-flex;
    align-items: center;
  }
}

.nav-register:hover {
  border-color: rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.04);
  color: #fff;
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
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: border-color 0.2s ease;
}

.nav-hamburger:hover {
  border-color: rgba(255, 255, 255, 0.18);
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
  background: rgba(255, 255, 255, 0.6);
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
  background: rgba(6, 6, 10, 0.95);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
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
  color: rgba(255, 255, 255, 0.5);
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.3s ease-out;
  transform: translateX(-8px);
}

.mobile-menu.open .mobile-link {
  transform: translateX(0);
}

.mobile-link:hover {
  background: rgba(255, 255, 255, 0.04);
  color: #fff;
}

.mobile-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
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
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.06);
  transition: background 0.2s ease;
}

.mobile-action-btn:hover {
  background: rgba(255, 255, 255, 0.08);
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
  margin-top: -100px;
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

/* ===== 品牌展示区 ===== */
.brand-section {
  flex: 1;
  text-align: center;
  min-width: 0;
}

@media (min-width: 768px) {
  .brand-section {
    text-align: left;
  }
}

/* ===== Glass Card ===== */
.glass-card {
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  position: relative;
  overflow: hidden;
  width: 100%;
  max-width: 420px;
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

.glass-card:hover {
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 8px 40px rgba(99, 102, 241, 0.06);
}

/* Shimmer 高光条 — 渐变色流动（无割裂感） */
.card-shimmer {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(
    90deg,
    rgba(139, 92, 246, 0.08),
    rgba(99, 102, 241, 0.35),
    rgba(129, 140, 248, 0.25),
    rgba(99, 102, 241, 0.35),
    rgba(139, 92, 246, 0.08)
  );
  background-size: 200% 100%;
  animation: shimmer-flow 4s linear infinite;
  z-index: 1;
}

/* 内边框渐变 */
.glass-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.08) 0%,
    rgba(255, 255, 255, 0.01) 30%,
    rgba(255, 255, 255, 0) 60%,
    rgba(255, 255, 255, 0.01) 80%,
    rgba(255, 255, 255, 0.05) 100%
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
    padding: 36px 32px 28px;
  }
}

/* ===== 版权信息 ===== */
.page-footer {
  position: relative;
  z-index: 10;
  text-align: center;
  padding: 0 16px 16px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.15);
}
</style>