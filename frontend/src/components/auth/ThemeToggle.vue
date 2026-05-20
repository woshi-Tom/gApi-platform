<template>
  <button
    class="theme-toggle"
    :class="{ pressed: isPressed }"
    @mousedown="onPress"
    @mouseup="onRelease"
    @mouseleave="onRelease"
    @click="toggleTheme"
    :aria-label="theme === 'dark' ? '切换到浅色模式' : '切换到深色模式'"
  >
    <transition name="theme-icon" mode="out-in">
      <svg
        v-if="theme === 'dark'"
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
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useTheme } from '@/composables/useTheme'

const { theme, toggleTheme } = useTheme()

const isPressed = ref(false)

function onPress() {
  isPressed.value = true
}

function onRelease() {
  isPressed.value = false
}
</script>

<style scoped>
.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: 1px solid var(--c-border-sub);
  background: var(--c-input-bg);
  color: var(--c-text-sub);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.22, 1, 0.36, 1);
  outline: none;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.theme-toggle:hover {
  border-color: var(--c-border-hover);
  background: rgba(99, 102, 241, 0.06);
  color: var(--c-text);
}

/* 按压反馈 */
.theme-toggle.pressed {
  transform: scale(0.85);
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.15);
}

/* 按钮涟漪效果 */
.theme-toggle::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.2) 0%, transparent 70%);
  opacity: 0;
  transform: scale(0);
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.theme-toggle:active::after {
  opacity: 1;
  transform: scale(2);
  transition: none;
}

.theme-icon {
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1);
}

/* 图标切换过渡 */
.theme-icon-enter-active {
  transition: all 0.3s cubic-bezier(0.22, 1, 0.36, 1);
}

.theme-icon-leave-active {
  transition: all 0.2s cubic-bezier(0.55, 0, 1, 0.45);
}

.theme-icon-enter-from {
  opacity: 0;
  transform: rotate(-90deg) scale(0.5);
}

.theme-icon-leave-to {
  opacity: 0;
  transform: rotate(90deg) scale(0.5);
}

</style>
