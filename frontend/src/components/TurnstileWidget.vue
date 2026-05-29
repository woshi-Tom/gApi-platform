<template>
  <div class="turnstile-box">
    <div ref="containerRef" class="turnstile-container"></div>
    <p v-if="errorMessage" class="turnstile-message">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'

const TURNSTILE_SCRIPT_ID = 'cloudflare-turnstile-script'
const TURNSTILE_SCRIPT_SRC = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'

let scriptLoadPromise: Promise<void> | null = null

const props = withDefaults(defineProps<{
  modelValue: string
  theme?: TurnstileTheme
  size?: TurnstileSize
}>(), {
  theme: 'auto',
  size: 'normal',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'verified', token: string): void
  (e: 'expired'): void
  (e: 'error'): void
  (e: 'timeout'): void
}>()

const containerRef = ref<HTMLElement | null>(null)
const widgetId = ref('')
const errorMessage = ref('')
let disposed = false

function loadTurnstileScript() {
  if (window.turnstile) {
    return Promise.resolve()
  }

  if (scriptLoadPromise) {
    return scriptLoadPromise
  }

  scriptLoadPromise = new Promise<void>((resolve, reject) => {
    const existingScript = document.getElementById(TURNSTILE_SCRIPT_ID) as HTMLScriptElement | null
    if (existingScript) {
      existingScript.addEventListener('load', () => resolve(), { once: true })
      existingScript.addEventListener('error', () => {
        scriptLoadPromise = null
        reject(new Error('Turnstile script failed to load'))
      }, { once: true })
      return
    }

    const script = document.createElement('script')
    script.id = TURNSTILE_SCRIPT_ID
    script.src = TURNSTILE_SCRIPT_SRC
    script.async = true
    script.defer = true
    script.onload = () => resolve()
    script.onerror = () => {
      scriptLoadPromise = null
      reject(new Error('Turnstile script failed to load'))
    }
    document.head.appendChild(script)
  })

  return scriptLoadPromise
}

function clearToken() {
  emit('update:modelValue', '')
}

function handleExpired() {
  clearToken()
  errorMessage.value = '人机验证已过期，请重新验证'
  emit('expired')
}

function handleError() {
  clearToken()
  errorMessage.value = '人机验证失败，请重试'
  emit('error')
}

function handleTimeout() {
  clearToken()
  errorMessage.value = '人机验证超时，请重新验证'
  emit('timeout')
}

async function renderWidget() {
  const siteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY
  if (!siteKey) {
    clearToken()
    errorMessage.value = '人机验证配置缺失：请检查 VITE_TURNSTILE_SITE_KEY 是否已注入前端环境，并重启前端服务'
    emit('error')
    return
  }

  try {
    await loadTurnstileScript()
    await nextTick()
  } catch {
    if (!disposed) {
      handleError()
    }
    return
  }

  if (disposed || !containerRef.value || !window.turnstile) {
    return
  }

  if (widgetId.value) {
    window.turnstile.remove(widgetId.value)
    widgetId.value = ''
  }

  widgetId.value = window.turnstile.render(containerRef.value, {
    sitekey: siteKey,
    theme: props.theme,
    size: props.size,
    callback: (token: string) => {
      errorMessage.value = ''
      emit('update:modelValue', token)
      emit('verified', token)
    },
    'expired-callback': handleExpired,
    'error-callback': handleError,
    'timeout-callback': handleTimeout,
  })
}

function reset() {
  clearToken()
  errorMessage.value = ''
  if (widgetId.value && window.turnstile) {
    window.turnstile.reset(widgetId.value)
    return
  }
  renderWidget()
}

onMounted(() => {
  renderWidget()
})

onBeforeUnmount(() => {
  disposed = true
  if (widgetId.value && window.turnstile) {
    window.turnstile.remove(widgetId.value)
    widgetId.value = ''
  }
})

defineExpose({ reset })
</script>

<style scoped>
.turnstile-box {
  width: 100%;
}

.turnstile-container {
  min-height: 65px;
}

.turnstile-message {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.4;
  color: #f87171;
}
</style>
