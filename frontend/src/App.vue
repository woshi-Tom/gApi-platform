<template>
  <el-container v-if="showLayout" class="app-container">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon class="logo-icon"><Monitor /></el-icon>
        <span>gAPI 平台</span>
      </div>
      <el-menu 
        :default-active="route.path" 
        router 
        background-color="#1e1e1e" 
        text-color="#a0a0a0" 
        active-text-color="#409eff"
        :ellipsis="false"
      >
        <el-menu-item index="/">
          <el-icon><HomeFilled /></el-icon>
          <span>控制台</span>
        </el-menu-item>
        <el-menu-item index="/tokens">
          <el-icon><Key /></el-icon>
          <span>API 密钥</span>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon><ShoppingCart /></el-icon>
          <span>商品列表</span>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><List /></el-icon>
          <span>订单记录</span>
        </el-menu-item>
        <el-menu-item index="/vip">
          <el-icon><Star /></el-icon>
          <span>VIP 会员</span>
        </el-menu-item>
        <el-menu-item index="/redeem">
          <el-icon><Ticket /></el-icon>
          <span>兑换码</span>
        </el-menu-item>
        <el-menu-item index="/profile">
          <el-icon><User /></el-icon>
          <span>个人中心</span>
        </el-menu-item>
      </el-menu>
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
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <router-view v-else />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import {
  HomeFilled, Key, ShoppingCart, List, Star, User, Ticket,
  ArrowDown, Avatar, Monitor
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

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
</style>
