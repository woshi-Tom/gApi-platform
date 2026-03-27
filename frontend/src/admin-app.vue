<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  User, Avatar, Setting, ArrowDown, Back, Clock, Connection,
  Document, DataAnalysis, Lock, Goods
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const adminUsername = ref('管理员')
const menuActive = computed(() => route.path)
const isLoginPage = computed(() => route.path === '/login')

function handleCommand(command: string) {
  if (command === 'logout') {
    handleLogout()
  } else if (command === 'change-password') {
    router.push('/change-password')
  }
}

function handleLogout() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_user')
  ElMessage.success('已退出管理后台')
  router.push('/login')
}

onMounted(() => {
  const admin = localStorage.getItem('admin_user')
  if (admin) {
    try {
      const user = JSON.parse(admin)
      adminUsername.value = user.username || '管理员'
    } catch {
    }
  }
})
</script>

<template>
  <router-view v-if="isLoginPage" />
  <el-container v-else class="app-container admin-layout">
    <el-aside width="220px" class="sidebar admin-sidebar">
      <div class="logo admin-logo">
        <el-icon class="logo-icon"><Setting /></el-icon>
        <span>管理后台</span>
      </div>
      <el-menu 
        :default-active="menuActive" 
        router 
        background-color="#1e1e1e" 
        text-color="#a0a0a0" 
        active-text-color="#409eff"
        :ellipsis="false"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon><Goods /></el-icon>
          <span>商品管理</span>
        </el-menu-item>
        <el-menu-item index="/channels">
          <el-icon><Connection /></el-icon>
          <span>渠道管理</span>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><Document /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon><Clock /></el-icon>
          <span>操作日志</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
        <el-menu-item index="/change-password">
          <el-icon><Lock /></el-icon>
          <span>修改密码</span>
        </el-menu-item>
        <el-divider style="margin: 10px 0; border-color: #333" />
        <el-menu-item @click="handleLogout">
          <el-icon><Back /></el-icon>
          <span>退出登录</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header admin-header">
        <span class="page-title">{{ route.meta.title || '管理后台' }}</span>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info admin-user">
              <el-icon><Avatar /></el-icon>
              {{ adminUsername }}
              <el-icon class="arrow"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="change-password">修改密码</el-dropdown-item>
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
</template>

<style>
/* Admin-specific styles */
.admin-layout {
  width: 100vw;
  height: 100vh;
  background: #f5f7fa;
}

.admin-sidebar {
  height: 100vh;
  background-color: #1e1e1e !important;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.admin-logo {
  height: 60px;
  background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
}

.admin-logo .logo-icon {
  color: #fff;
  font-size: 24px;
}

.admin-header {
  height: 60px !important;
  background: #fff !important;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.admin-header .page-title {
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.admin-user {
  color: #606266;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.admin-user:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

/* 内容区域 */
.main-content {
  height: calc(100vh - 60px);
  overflow-y: auto;
  padding: 24px;
  background: #f5f7fa;
}

/* 修复 el-container 布局 */
.el-container {
  width: 100%;
  height: 100%;
}
</style>
