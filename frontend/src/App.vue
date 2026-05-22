<!--
  App.vue - 应用根组件（编排层）
  仅负责：布局切换判断、全局状态初始化、组件编排

  侧边栏 → src/components/layout/AppSidebar.vue
  顶部栏 → src/components/layout/AppHeader.vue
  主内容 → src/components/layout/AppMain.vue
  全局布局样式 → src/styles/layout.css

  禁止在此文件中添加：页面 UI、表格、表单、API 请求、业务逻辑
-->
<template>
  <el-container v-if="showLayout" class="app-container">
    <AppSidebar
      :collapsed="collapsed"
      :active-path="route.path"
      :width="sidebarWidth"
      :user="authStore.user"
      @toggle="toggleSidebar"
      @command="handleCommand"
    />
    <el-container>
      <AppHeader
        :title="(route.meta.title as string) || '控制台'"
      />
      <AppMain />
    </el-container>
  </el-container>

  <router-view v-else />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppMain from '@/components/layout/AppMain.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 侧边栏折叠状态（记忆到 localStorage）
const collapsed = ref(localStorage.getItem('sidebar-collapsed') === 'true')
const sidebarWidth = computed(() => collapsed.value ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)')

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
  } else if (command === 'profile' || command === 'settings') {
    router.push('/profile')
  }
}

onMounted(async () => {
  if (authStore.isLoggedIn && !authStore.user) {
    await authStore.fetchProfile()
  }
})
</script>