import { ref, watch } from 'vue'

type Theme = 'dark' | 'light'

const theme = ref<Theme>((localStorage.getItem('gapi-theme') as Theme) || 'dark')

// 初始化时立即应用
applyTheme(theme.value)

function applyTheme(t: Theme) {
  document.documentElement.setAttribute('data-theme', t)
}

watch(theme, (val) => {
  localStorage.setItem('gapi-theme', val)
  applyTheme(val)
})

export function useTheme() {
  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  return {
    theme,
    toggleTheme,
  }
}