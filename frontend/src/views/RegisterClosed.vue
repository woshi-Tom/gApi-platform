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
          <div class="closed-content">
            <el-icon class="closed-icon animate-fade-up" style="animation-delay: 400ms"><CircleClose /></el-icon>
            <h2 class="animate-fade-up" style="animation-delay: 500ms">注册已关闭</h2>
            <p class="animate-fade-up" style="animation-delay: 550ms">目前暂不支持新用户注册，感谢您的理解。</p>
            <div class="closed-actions animate-fade-up" style="animation-delay: 600ms">
              <el-button type="primary" class="submit-btn" @click="router.push('/login')">前往登录</el-button>
              <el-button class="ghost-btn" @click="router.push('/')">返回首页</el-button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <footer class="page-footer animate-fade-up" style="animation-delay: 700ms">© {{ new Date().getFullYear() }} gAPI Platform. All rights reserved.</footer>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { CircleClose } from '@element-plus/icons-vue'

const router = useRouter()
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
.closed-content { display: flex; flex-direction: column; align-items: center; padding: 20px 0; text-align: center; }
.closed-icon { font-size: 56px; color: rgba(255,255,255,0.25); margin-bottom: 16px; }
.closed-content h2 { margin: 0 0 12px; font-size: 20px; font-weight: 600; color: #fff; }
.closed-content p { margin: 0 0 28px; font-size: 14px; color: rgba(255,255,255,0.35); line-height: 1.6; }
.closed-actions { display: flex; gap: 12px; width: 100%; }
.submit-btn { flex: 1; height: 42px; border-radius: 10px; font-size: 14px; font-weight: 500; background: #fff; color: #000; border: none; transition: all 0.2s; }
.submit-btn:hover { background: rgba(229,229,229,1); }
.ghost-btn { flex: 1; height: 42px; border-radius: 10px; font-size: 14px; font-weight: 500; color: rgba(255,255,255,0.7); background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1); transition: all 0.2s; }
.ghost-btn:hover { background: rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.2); }
.page-footer { position: relative; z-index: 10; text-align: center; padding: 0 16px 16px; font-size: 12px; color: rgba(255,255,255,0.2); }
</style>