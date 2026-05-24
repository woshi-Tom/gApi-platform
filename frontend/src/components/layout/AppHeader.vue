<template>
  <el-header class="header">
    <span class="page-title">{{ title }}</span>
    <div class="header-right">
      <button
        class="theme-toggle-btn"
        @click="$emit('toggle-theme')"
        :title="consoleTheme === 'dark' ? '切换到浅色模式' : '切换到深色模式'"
      >
        <transition name="theme-icon-fade" mode="out-in">
          <!-- 太阳图标：深色模式下显示，点击切换到浅色 -->
          <svg
            v-if="consoleTheme === 'dark'"
            key="sun"
            class="theme-icon"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="5" />
            <line x1="12" y1="1" x2="12" y2="3" />
            <line x1="12" y1="21" x2="12" y2="23" />
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
            <line x1="1" y1="12" x2="3" y2="12" />
            <line x1="21" y1="12" x2="23" y2="12" />
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
          </svg>
          <!-- 月亮图标：浅色模式下显示，点击切换到深色 -->
          <svg
            v-else
            key="moon"
            class="theme-icon"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
        </transition>
      </button>
    </div>
  </el-header>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  consoleTheme: 'dark' | 'light'
}>()

defineEmits<{
  'toggle-theme': []
}>()
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--color-bg-elevated);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--color-border-light);
  padding: 0 var(--spacing-xl);
  height: var(--header-height);
}

.page-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
}

.header-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.theme-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.22s ease;
  outline: none;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.theme-toggle-btn:hover {
  border-color: rgba(99, 102, 241, 0.3);
  background: var(--color-bg-hover);
  color: var(--color-text-primary);
}

.theme-toggle-btn:active {
  transform: scale(0.92);
}

.theme-icon {
  display: block;
}

/* 图标切换过渡 */
.theme-icon-fade-enter-active {
  transition: all 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}

.theme-icon-fade-leave-active {
  transition: all 0.15s cubic-bezier(0.55, 0, 1, 0.45);
}

.theme-icon-fade-enter-from {
  opacity: 0;
  transform: rotate(-90deg) scale(0.5);
}

.theme-icon-fade-leave-to {
  opacity: 0;
  transform: rotate(90deg) scale(0.5);
}

/* 响应式 */
@media (max-width: 640px) {
  .header {
    padding: 0 var(--spacing-md);
  }

  .page-title {
    font-size: var(--font-size-lg);
  }
}
</style>