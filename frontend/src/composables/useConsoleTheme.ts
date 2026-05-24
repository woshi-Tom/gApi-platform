import { ref, watch } from 'vue'

type ConsoleTheme = 'dark' | 'light'

const STORAGE_KEY = 'gapi-console-theme'
const theme = ref<ConsoleTheme>((localStorage.getItem(STORAGE_KEY) as ConsoleTheme) || 'dark')

// 初始化时立即应用
applyTheme(theme.value)

function applyTheme(t: ConsoleTheme) {
  const el = document.querySelector('.console-layout')
  if (el) {
    el.classList.toggle('console-theme-light', t === 'light')
  }
  // 也存到 localStorage，供后续刷新读取
  localStorage.setItem(STORAGE_KEY, t)
}

watch(theme, (val) => {
  applyTheme(val)
})

export function useConsoleTheme() {
  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
    // watch 会自动触发 applyTheme
  }

  // 确保在组件挂载时也能正确应用（处理首次加载时 .console-layout 还不存在的情况）
  function ensureApplied() {
    applyTheme(theme.value)
  }

  return {
    theme,
    toggleTheme,
    ensureApplied,
  }
}