<template>
  <div class="slider-captcha">
    <div class="captcha-container" v-if="captchaData">
      <div class="captcha-bg" ref="bgRef">
        <div 
          class="slider-track"
          :style="{ left: sliderX + 'px' }"
        >
          <div 
            class="slider-button"
            @mousedown="startDrag"
            @touchstart="startDrag"
          >
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
        <div class="captcha-tip" v-if="!verified">
          请拖动滑块完成验证
        </div>
        <div class="captcha-success" v-else>
          <el-icon><Check /></el-icon>
          验证成功
        </div>
      </div>
    </div>
    <div class="captcha-loading" v-else>
      <el-icon class="loading-icon"><Loading /></el-icon>
      加载中...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowRight, Check, Loading } from '@element-plus/icons-vue'
import axios from 'axios'

interface CaptchaData {
  token: string
  bg_image: string
  slider_image: string
  x_position: number
  expires_at: string
}

const emit = defineEmits<{
  (e: 'success', token: string): void
  (e: 'fail'): void
}>()

const bgRef = ref<HTMLElement>()
const captchaData = ref<CaptchaData | null>(null)
const sliderX = ref(0)
const isDragging = ref(false)
const startX = ref(0)
const track = ref<number[]>([])
const startTime = ref(0)
const verified = ref(false)

const loadCaptcha = async () => {
  try {
    const response = await axios.get('/api/v1/captcha/generate')
    captchaData.value = response.data.data
    sliderX.value = 0
    track.value = []
    verified.value = false
  } catch (error) {
    ElMessage.error('加载验证码失败')
  }
}

const startDrag = (e: MouseEvent | TouchEvent) => {
  if (verified.value) return
  
  isDragging.value = true
  startX.value = e instanceof MouseEvent ? e.clientX : e.touches[0].clientX
  startTime.value = Date.now()
  track.value = []
  
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', endDrag)
  document.addEventListener('touchmove', onDrag)
  document.addEventListener('touchend', endDrag)
}

const onDrag = (e: MouseEvent | TouchEvent) => {
  if (!isDragging.value || !bgRef.value) return
  
  const currentX = e instanceof MouseEvent ? e.clientX : e.touches[0].clientX
  const diff = currentX - startX.value
  
  const maxX = bgRef.value.offsetWidth - 40
  sliderX.value = Math.max(0, Math.min(diff, maxX))
  
  track.value.push(sliderX.value)
}

const endDrag = async () => {
  if (!isDragging.value) return
  
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
  document.removeEventListener('touchmove', onDrag)
  document.removeEventListener('touchend', endDrag)
  
  const duration = Date.now() - startTime.value
  
  try {
    const response = await axios.post('/api/v1/captcha/verify', {
      token: captchaData.value?.token,
      track: track.value,
      duration
    })
    
    if (response.data.data.passed) {
      verified.value = true
      emit('success', captchaData.value?.token || '')
    } else {
      ElMessage.error('验证失败，请重试')
      emit('fail')
      await loadCaptcha()
    }
  } catch (error) {
    ElMessage.error('验证失败')
    emit('fail')
    await loadCaptcha()
  }
}

onMounted(() => {
  loadCaptcha()
})

onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
  document.removeEventListener('touchmove', onDrag)
  document.removeEventListener('touchend', endDrag)
})

defineExpose({
  refresh: loadCaptcha
})
</script>

<style scoped>
.slider-captcha {
  width: 100%;
}

.captcha-container {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
}

.captcha-bg {
  position: relative;
  height: 40px;
  background: linear-gradient(90deg, #f5f7fa 0%, #e4e7ed 100%);
}

.slider-track {
  position: absolute;
  top: 0;
  left: 0;
  width: 40px;
  height: 100%;
  background: #409eff;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s;
}

.slider-track:hover {
  background: #337ecc;
}

.slider-button {
  color: #fff;
  font-size: 18px;
}

.captcha-tip {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: #909399;
  font-size: 14px;
  pointer-events: none;
}

.captcha-success {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: #67c23a;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.captcha-loading {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  background: #f5f7fa;
  border-radius: 8px;
}

.loading-icon {
  animation: rotating 2s linear infinite;
  margin-right: 8px;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
