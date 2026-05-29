<!--
  App.vue - 应用根组件（编排层）
  仅负责：布局切换判断、全局状态初始化、组件编排

  侧边栏 → src/components/layout/AppSidebar.vue
  设置面板 → src/components/layout/SidebarSettingsPanel.vue
  主内容 → src/components/layout/AppMain.vue
  全局布局样式 → src/styles/layout.css

  禁止在此文件中添加：页面 UI、表格、表单、API 请求、业务逻辑
-->
<template>
  <el-container
    v-if="showLayout"
    ref="layoutRef"
    class="app-container console-layout"
    :class="collapsed ? 'console-layout--sidebar-collapsed' : 'console-layout--sidebar-expanded'"
    :style="{ '--console-sidebar-current-width': sidebarWidth }"
  >
    <AppSidebar
      :collapsed="collapsed"
      :active-path="route.path"
      :width="sidebarWidth"
      :user="authStore.user"
      :settings-open="settingsPanelOpen"
      @toggle="toggleSidebar"
      @toggle-settings="toggleSettingsPanel"
      @command="handleCommand"
    />
    <SidebarSettingsPanel
      v-if="settingsPanelOpen"
      :sidebar-width="sidebarWidth"
      :console-theme="consoleTheme"
      @toggle-theme="toggleConsoleTheme"
      @close="closeSettingsPanel"
    />
    <el-container
      class="console-layout__main-shell"
      :class="{ 'console-layout__main-shell--sidebar-collapsed': collapsed }"
    >
      <AppMain />
    </el-container>
  </el-container>

  <router-view v-else />
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useConsoleTheme } from '@/composables/useConsoleTheme'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppMain from '@/components/layout/AppMain.vue'
import SidebarSettingsPanel from '@/components/layout/SidebarSettingsPanel.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const layoutRef = ref<HTMLElement | null>(null)

// 控制台主题
const { theme: consoleTheme, toggleTheme: toggleConsoleTheme, ensureApplied } = useConsoleTheme()

// 侧边栏折叠状态（记忆到 localStorage）
const collapsed = ref(localStorage.getItem('sidebar-collapsed') === 'true')
const sidebarWidth = computed(() => collapsed.value ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)')
const settingsPanelOpen = ref(false)

function toggleSidebar() {
  collapsed.value = !collapsed.value
  localStorage.setItem('sidebar-collapsed', String(collapsed.value))
}

function toggleSettingsPanel() {
  settingsPanelOpen.value = !settingsPanelOpen.value
}

function closeSettingsPanel() {
  settingsPanelOpen.value = false
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

function handleDocumentClick() {
  if (settingsPanelOpen.value) closeSettingsPanel()
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && settingsPanelOpen.value) closeSettingsPanel()
}

onMounted(async () => {
  document.addEventListener('click', handleDocumentClick)
  document.addEventListener('keydown', handleDocumentKeydown)

  if (authStore.isLoggedIn && !authStore.user) {
    await authStore.fetchProfile()
  }
  // 确保主题在 .console-layout 挂载后正确应用
  await nextTick()
  ensureApplied()
})

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
</script>

<style scoped>
.console-layout__main-shell {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  transition: width var(--sidebar-transition);
}
</style>
