<template>
  <aside
    class="sidebar-settings-panel"
    :class="{ 'sidebar-settings-panel--light': !isDark }"
    :style="panelStyle"
    role="dialog"
    aria-label="侧边栏设置"
    @click.stop
  >
    <div class="settings-panel__header">
      <div>
        <div class="settings-panel__title">设置</div>
        <div class="settings-panel__subtitle">控制台偏好</div>
      </div>
      <button class="settings-panel__close" type="button" aria-label="关闭设置面板" @click="$emit('close')">
        <el-icon :size="16"><Close /></el-icon>
      </button>
    </div>

    <section class="settings-panel__section">
      <div class="settings-panel__section-title">外观</div>
      <button class="settings-panel__row settings-panel__row--button" type="button" @click="$emit('toggle-theme')">
        <span class="settings-panel__row-icon">
          <el-icon :size="18">
            <Moon v-if="isDark" />
            <Sunny v-else />
          </el-icon>
        </span>
        <span class="settings-panel__row-main">
          <span class="settings-panel__row-label">主题</span>
          <span class="settings-panel__row-desc">{{ isDark ? '当前为深色模式' : '当前为浅色模式' }}</span>
        </span>
        <span class="settings-panel__pill">{{ isDark ? '深色' : '浅色' }}</span>
      </button>
    </section>

    <section class="settings-panel__section">
      <div class="settings-panel__section-title">偏好</div>
      <button class="settings-panel__row" type="button" disabled>
        <span class="settings-panel__row-icon">
          <el-icon :size="18"><Reading /></el-icon>
        </span>
        <span class="settings-panel__row-main">
          <span class="settings-panel__row-label">语言</span>
          <span class="settings-panel__row-desc">简体中文</span>
        </span>
        <span class="settings-panel__soon">即将支持</span>
      </button>

      <button class="settings-panel__row" type="button" disabled>
        <span class="settings-panel__row-icon">
          <el-icon :size="18"><Bell /></el-icon>
        </span>
        <span class="settings-panel__row-main">
          <span class="settings-panel__row-label">通知</span>
          <span class="settings-panel__row-desc">暂无通知</span>
        </span>
        <span class="settings-panel__soon">即将支持</span>
      </button>
    </section>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Bell, Close, Moon, Reading, Sunny } from '@element-plus/icons-vue'

const props = defineProps<{
  sidebarWidth: string
  consoleTheme: 'dark' | 'light'
}>()

defineEmits<{
  close: []
  'toggle-theme': []
}>()

const isDark = computed(() => props.consoleTheme === 'dark')

const panelStyle = computed<Record<string, string>>(() => ({
  '--sidebar-settings-left': `calc(${props.sidebarWidth} + 12px)`,
}))
</script>

<style scoped>
.sidebar-settings-panel {
  position: fixed;
  left: var(--sidebar-settings-left);
  bottom: 18px;
  z-index: 2200;
  width: min(320px, calc(100vw - var(--sidebar-settings-left) - 12px));
  max-height: calc(100vh - 36px);
  overflow: auto;
  padding: 12px;
  border: 1px solid var(--color-sidebar-border);
  border-radius: 18px;
  background: rgba(18, 20, 28, 0.94);
  color: var(--color-text-primary);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.36);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  transition: left var(--sidebar-transition), width var(--sidebar-transition);
}

.settings-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 4px 4px 10px;
}

.settings-panel__title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  line-height: 1.3;
}

.settings-panel__subtitle {
  margin-top: 2px;
  color: var(--color-text-placeholder);
  font-size: var(--font-size-xs);
}

.settings-panel__close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast);
}

.settings-panel__close:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.settings-panel__section {
  padding: 8px 0;
  border-top: 1px solid var(--color-border-light);
}

.settings-panel__section-title {
  padding: 0 4px 6px;
  color: var(--color-text-placeholder);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.settings-panel__row {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 10px;
  min-height: 48px;
  padding: 8px 10px;
  border: none;
  border-radius: 12px;
  background: transparent;
  color: var(--color-text-primary);
  text-align: left;
  transition: background-color var(--transition-fast), color var(--transition-fast);
}

.settings-panel__row--button {
  cursor: pointer;
}

.settings-panel__row--button:hover {
  background: var(--color-bg-hover);
}

.settings-panel__row:disabled {
  cursor: not-allowed;
  opacity: 0.68;
}

.settings-panel__row-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 10px;
  background: var(--color-bg-hover);
  color: var(--color-sidebar-active-text);
  flex-shrink: 0;
}

.settings-panel__row-main {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.settings-panel__row-label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.settings-panel__row-desc {
  color: var(--color-text-placeholder);
  font-size: var(--font-size-xs);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.settings-panel__pill,
.settings-panel__soon {
  flex-shrink: 0;
  border-radius: var(--radius-full);
  padding: 3px 8px;
  font-size: 11px;
  font-weight: var(--font-weight-medium);
}

.settings-panel__pill {
  background: var(--color-primary-bg);
  color: var(--color-sidebar-active-text);
}

.settings-panel__soon {
  background: rgba(148, 163, 184, 0.12);
  color: var(--color-text-placeholder);
}

.sidebar-settings-panel--light {
  background: rgba(255, 255, 255, 0.96);
  border-color: var(--color-border);
  color: var(--color-text-primary);
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.14);
}

.sidebar-settings-panel--light .settings-panel__subtitle,
.sidebar-settings-panel--light .settings-panel__section-title,
.sidebar-settings-panel--light .settings-panel__row-desc {
  color: var(--color-text-secondary);
}

.sidebar-settings-panel--light .settings-panel__close {
  color: var(--color-text-secondary);
}

.sidebar-settings-panel--light .settings-panel__close:hover {
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.sidebar-settings-panel--light .settings-panel__section {
  border-color: var(--color-border-light);
}

.sidebar-settings-panel--light .settings-panel__row {
  color: var(--color-text-primary);
}

.sidebar-settings-panel--light .settings-panel__row--button:hover {
  background: var(--color-bg-hover);
}

.sidebar-settings-panel--light .settings-panel__row:disabled {
  opacity: 0.78;
}

.sidebar-settings-panel--light .settings-panel__row-icon {
  background: rgba(99, 102, 241, 0.08);
  color: var(--color-sidebar-active-text);
}

.sidebar-settings-panel--light .settings-panel__pill {
  background: var(--color-primary-bg);
  color: var(--color-sidebar-active-text);
}

.sidebar-settings-panel--light .settings-panel__soon {
  background: rgba(100, 116, 139, 0.1);
  color: var(--color-text-secondary);
}
</style>
