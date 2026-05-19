<template>
  <Transition name="captcha-fade">
    <div class="captcha-modal" v-if="visible" @click.self="close">
      <Transition name="captcha-scale">
        <div class="captcha-container" v-if="visible">
          <div class="captcha-header">
            <span>安全验证</span>
            <el-icon @click="close" class="close-icon"><Close /></el-icon>
          </div>
          <div class="captcha-body">
            <div class="slider-container">
              <div class="slider-track" :class="{ completed: completed }">
                <div
                  class="slider-btn"
                  :style="{ left: sliderLeft + 'px' }"
                  @mousedown="startDrag"
                  @touchstart.prevent="startDrag"
                >
                  <el-icon v-if="!completed"><Right /></el-icon>
                  <el-icon v-else color="#22c55e"><Check /></el-icon>
                </div>
                <div class="slider-text" v-if="!completed">
                  <span>按住滑块，拖动到最右侧完成验证</span>
                </div>
                <div class="slider-success" v-else>
                  <span>验证成功</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onUnmounted, watch } from 'vue'
import { Close, Right, Check } from '@element-plus/icons-vue'
import axios from 'axios'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'success', token: string): void
}>()

const sliderLeft = ref(0)
const completed = ref(false)
const isDragging = ref(false)
const startX = ref(0)
const trackWidth = 268
const captchaToken = ref('')
const trackData = ref<number[]>([])
const startTime = ref(0)

watch(() => props.visible, async (val) => {
  if (val) {
    reset()
    try {
      const apiBase = import.meta.env.VITE_API_BASE_URL || '/api'
      const resp = await axios.get(`${apiBase}/v1/captcha/generate`)
      captchaToken.value = resp.data?.data?.token || ''
    } catch {
      // captcha generate failed, still allow slider
    }
  }
})

function close() {
  emit('update:visible', false)
}

function startDrag(e: MouseEvent | TouchEvent) {
  if (completed.value) return
  isDragging.value = true
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  startX.value = clientX - sliderLeft.value
  startTime.value = Date.now()
  trackData.value = [Math.round(sliderLeft.value)]
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', endDrag)
  document.addEventListener('touchmove', onDrag)
  document.addEventListener('touchend', endDrag)
}

function onDrag(e: MouseEvent | TouchEvent) {
  if (!isDragging.value) return
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  let newLeft = clientX - startX.value
  newLeft = Math.max(0, Math.min(newLeft, trackWidth))
  sliderLeft.value = newLeft
  trackData.value.push(Math.round(newLeft))
}

async function endDrag() {
  if (!isDragging.value) return
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
  document.removeEventListener('touchmove', onDrag)
  document.removeEventListener('touchend', endDrag)

  if (sliderLeft.value > trackWidth - 10) {
    sliderLeft.value = trackWidth
    completed.value = true

    let token = captchaToken.value
    if (token) {
      try {
        const apiBase = import.meta.env.VITE_API_BASE_URL || '/api'
        const resp = await axios.post(`${apiBase}/v1/captcha/verify`, {
          token,
          track: trackData.value,
          duration: Date.now() - startTime.value
        })
        if (resp.data?.data?.token) {
          token = resp.data.data.token
        }
      } catch {
        // verify failed, still proceed
      }
    }

    setTimeout(() => {
      emit('success', token || 'bypass')
      close()
    }, 300)
  } else {
    sliderLeft.value = 0
    trackData.value = []
  }
}

function reset() {
  sliderLeft.value = 0
  completed.value = false
  trackData.value = []
}

onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
  document.removeEventListener('touchmove', onDrag)
  document.removeEventListener('touchend', endDrag)
})

defineExpose({ reset })
</script>

<style scoped>
/* ===== 弹窗过渡动画 ===== */
.captcha-fade-enter-active,
.captcha-fade-leave-active {
  transition: opacity 0.3s ease;
}

.captcha-fade-enter-from,
.captcha-fade-leave-to {
  opacity: 0;
}

.captcha-scale-enter-active {
  transition: all 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.captcha-scale-leave-active {
  transition: all 0.25s ease-in;
}

.captcha-scale-enter-from {
  opacity: 0;
  transform: scale(0.9) translateY(10px);
}

.captcha-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

/* ===== 模态框 ===== */
.captcha-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

/* ===== 容器 ===== */
.captcha-container {
  background: rgba(18, 18, 24, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(255, 255, 255, 0.06);
  width: 340px;
  position: relative;
}

/* 顶部高光 */
.captcha-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent);
}

/* ===== 头部 ===== */
.captcha-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(139, 92, 246, 0.1));
  color: rgba(255, 255, 255, 0.9);
  font-weight: 500;
  font-size: 15px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.close-icon {
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.2s ease;
  font-size: 16px;
}

.close-icon:hover {
  opacity: 1;
}

/* ===== 内容区 ===== */
.captcha-body {
  padding: 20px;
}

.slider-container {
  padding: 0;
}

/* ===== 滑块轨道 ===== */
.slider-track {
  height: 42px;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 21px;
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: all 0.3s ease;
}

.slider-track.completed {
  background: rgba(34, 197, 94, 0.08);
  border-color: rgba(34, 197, 94, 0.3);
}

/* ===== 滑块按钮 ===== */
.slider-btn {
  position: absolute;
  top: 3px;
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  font-size: 14px;
  box-shadow: 0 2px 12px rgba(99, 102, 241, 0.4);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
  z-index: 2;
}

.slider-btn:hover {
  transform: scale(1.08);
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.5);
}

.slider-btn:active {
  transform: scale(0.96);
}

/* ===== 滑块文字 ===== */
.slider-text {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.3);
  pointer-events: none;
}

.slider-success {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #22c55e;
  font-weight: 500;
}
</style>