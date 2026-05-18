<template>
  <el-container v-if="showLayout" class="app-container">
    <el-aside :width="sidebarWidth" class="sidebar">
      <div class="logo">
        <el-icon class="logo-icon"><Monitor /></el-icon>
        <span v-show="!collapsed">gAPI 平台</span>
      </div>
      <el-menu 
        :default-active="route.path" 
        router 
        background-color="#1e1e1e" 
        text-color="#a0a0a0" 
        active-text-color="#409eff"
        :ellipsis="false"
        :collapse="collapsed"
      >
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <template #title>控制台</template>
        </el-menu-item>
        <el-menu-item index="/tokens">
          <el-icon><Key /></el-icon>
          <template #title>API 密钥</template>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon><ShoppingCart /></el-icon>
          <template #title>商品列表</template>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><List /></el-icon>
          <template #title>订单记录</template>
        </el-menu-item>
        <el-menu-item index="/vip">
          <el-icon><Star /></el-icon>
          <template #title>VIP 会员</template>
        </el-menu-item>
        <el-menu-item index="/redeem">
          <el-icon><Ticket /></el-icon>
          <template #title>兑换码</template>
        </el-menu-item>
        <el-menu-item index="/profile">
          <el-icon><User /></el-icon>
          <template #title>个人中心</template>
        </el-menu-item>
      </el-menu>
      <div class="sidebar-toggle" @click="toggleSidebar">
        <el-icon :size="18">
          <Fold v-if="!collapsed" />
          <Expand v-else />
        </el-icon>
      </div>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span class="page-title">{{ route.meta.title || '控制台' }}</span>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><Avatar /></el-icon>
              {{ authStore.user?.username || '用户' }}
              <el-icon class="arrow"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>

  <router-view v-else />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import {
  HomeFilled, Key, ShoppingCart, List, Star, User, Ticket,
  ArrowDown, Avatar, Monitor, Fold, Expand
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 侧边栏折叠状态（记忆到 localStorage）
const collapsed = ref(localStorage.getItem('sidebar-collapsed') === 'true')
const sidebarWidth = computed(() => collapsed.value ? '64px' : '220px')

function toggleSidebar() {
  collapsed.value = !collapsed.value
  localStorage.setItem('sidebar-collapsed', String(collapsed.value))
}

const showLayout = computed(() => {
  return route.meta.requiresAuth
})

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/profile')
  }
}

onMounted(async () => {
  if (authStore.isLoggedIn && !authStore.user) {
    await authStore.fetchProfile()
  }
})
</script>

<style>
:root {
  --sidebar-bg: #1e1e1e;
  --sidebar-hover: #2a2a2a;
  --sidebar-active: #409eff;
  --content-bg: #f5f7fa;
  --card-bg: #ffffff;
  --border-color: #e4e7ed;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: var(--content-bg);
}

#app {
  width: 100%;
  height: 100%;
}

.app-container {
  width: 100%;
  height: 100vh;
  min-width: 100%;
  background-color: var(--content-bg);
}

.sidebar {
  background-color: var(--sidebar-bg) !important;
  border-right: 1px solid #333;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  background-color: #252525;
  border-bottom: 1px solid #333;
}

.logo-icon {
  font-size: 22px;
  color: #409eff;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--card-bg);
  border-bottom: 1px solid var(--border-color);
  padding: 0 24px;
  height: 60px;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.2s;
  color: var(--el-text-color-primary);
}

.user-info:hover {
  background-color: var(--el-fill-color-light);
}

.arrow {
  margin-left: 4px;
  font-size: 12px;
}

.main-content {
  padding: 24px;
  overflow-y: auto;
}

.el-menu {
  border-right: none !important;
}

.el-menu-item {
  height: 50px;
  line-height: 50px;
  margin: 4px 8px;
  border-radius: 8px;
}

.el-menu-item:hover {
  background-color: var(--sidebar-hover) !important;
}

.el-menu-item.is-active {
  background-color: rgba(64, 158, 255, 0.15) !important;
}

.el-card {
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  border: 1px solid var(--border-color);
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #c0c4cc;
}

/* ===== 页面切换过渡动画 ===== */
.page-fade-enter-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.page-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ===== 卡片统一 hover 效果 ===== */
.el-card {
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
}

.el-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  border-color: var(--el-border-color);
}

/* 排除表格内的 card 不应用 hover 上浮 */
.el-card:has(.el-table):hover {
  transform: none;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

/* ===== 侧边栏折叠 ===== */
.sidebar {
  transition: width var(--transition-base, 0.25s ease) !important;
  overflow-x: hidden;
  position: relative;
}

.logo {
  white-space: nowrap;
  overflow: hidden;
}

.logo span {
  transition: opacity 0.2s ease;
}

.el-menu--collapse .el-menu-item {
  padding: 0;
  display: flex;
  justify-content: center;
}

.el-menu--collapse .el-menu-item .el-icon {
  margin: 0;
}

.sidebar-toggle {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #a0a0a0;
  border-top: 1px solid #333;
  background-color: #1e1e1e;
  transition: color 0.2s, background-color 0.2s;
}

.sidebar-toggle:hover {
  color: #fff;
  background-color: #2a2a2a;
}
</style>
