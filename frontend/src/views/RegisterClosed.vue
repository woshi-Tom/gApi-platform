<template>
  <AuthLayout :showLoginLink="true">
    <!-- 品牌展示区 -->
    <template #brand>
      <h1 class="brand-title">
        <span class="gradient-text">gAPI Platform</span>
        <br />智能 AI API 中转平台
      </h1>
      <p class="brand-desc">感谢您对 gAPI Platform 的关注，注册功能暂时关闭中</p>

      <!-- 特性亮点 -->
      <div class="feature-list animate-fade-up" style="animation-delay: 600ms">
        <div class="feature-item">
          <div class="feature-icon icon-purple">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          </div>
          <div class="feature-text">
            <strong>敬请期待</strong>
            我们正在完善系统，即将重新开放
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon icon-blue">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
          </div>
          <div class="feature-text">
            <strong>获取通知</strong>
            如有邀请码可直接使用
          </div>
        </div>
      </div>
    </template>

    <!-- 注册关闭卡片内容 -->
    <div class="closed-content">
      <div class="closed-icon-wrapper animate-fade-up" style="animation-delay: 400ms">
        <el-icon class="closed-icon"><CircleClose /></el-icon>
      </div>
      <h2 class="animate-fade-up" style="animation-delay: 500ms">注册已关闭</h2>
      <p class="animate-fade-up" style="animation-delay: 550ms">目前暂不支持新用户注册，感谢您的理解。</p>
      <div class="closed-actions animate-fade-up" style="animation-delay: 600ms">
        <el-button type="primary" class="submit-btn" @click="router.push('/login')">前往登录</el-button>
        <el-button class="ghost-btn" @click="router.push('/')">返回首页</el-button>
      </div>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { CircleClose } from '@element-plus/icons-vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import '@/styles/auth.css'

const router = useRouter()
</script>

<style scoped>
/* ===== 品牌展示区 ===== */
.brand-title { font-size: 28px; font-weight: 600; letter-spacing: 0; line-height: 1.2; margin: 0 0 16px; color: #fff; }
@media (min-width: 640px) { .brand-title { font-size: 36px; } }
@media (min-width: 768px) { .brand-title { font-size: 44px; } }
@media (min-width: 1024px) { .brand-title { font-size: 52px; } }

.gradient-text { background: linear-gradient(135deg, #fff 0%, rgba(167, 139, 250, 0.7) 50%, rgba(129, 140, 248, 0.8) 100%); background-size: 200% 200%; animation: gradient-shift 6s ease infinite; -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }

.brand-desc { font-size: 15px; color: rgba(255, 255, 255, 0.35); margin: 0 0 24px; line-height: 1.6; max-width: 400px; }
@media (min-width: 768px) { .brand-desc { font-size: 16px; } }

/* ===== 关闭内容 ===== */
.closed-content { display: flex; flex-direction: column; align-items: center; padding: 20px 0; text-align: center; }
.closed-icon-wrapper { position: relative; display: inline-flex; align-items: center; justify-content: center; margin-bottom: 16px; }
.closed-icon { font-size: 56px; color: rgba(255, 255, 255, 0.2); animation: pulse-icon 3s ease-in-out infinite; }
.closed-icon-wrapper::before { content: ''; position: absolute; width: 80px; height: 80px; border-radius: 50%; background: rgba(255, 255, 255, 0.03); animation: pulse-ring 3s ease-in-out infinite; }
.closed-content h2 { margin: 0 0 12px; font-size: 22px; font-weight: 600; color: #fff; }
.closed-content p { margin: 0 0 28px; font-size: 14px; color: rgba(255, 255, 255, 0.3); line-height: 1.6; }
.closed-actions { display: flex; gap: 12px; width: 100%; }

.submit-btn { flex: 1; height: 44px; border-radius: 10px; font-size: 14px; font-weight: 500; background: var(--c-primary); color: #fff; border: none; transition: all 0.3s ease; position: relative; overflow: hidden; }
.submit-btn::after { content: ''; position: absolute; top: 0; left: -100%; width: 100%; height: 100%; background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent); display: none; }
.submit-btn:hover { background: var(--c-primary-hover); box-shadow: none; transform: translateY(-1px); }

.ghost-btn { flex: 1; height: 44px; border-radius: 10px; font-size: 14px; font-weight: 500; color: rgba(255, 255, 255, 0.6); background: rgba(255, 255, 255, 0.03); border: 1px solid rgba(255, 255, 255, 0.08); transition: all 0.3s ease; }
.ghost-btn:hover { background: rgba(255, 255, 255, 0.06); border-color: rgba(255, 255, 255, 0.15); color: #fff; transform: translateY(-1px); }
</style>