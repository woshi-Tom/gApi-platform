<template>
  <div class="captcha-modal" v-if="visible" @click.self="close">
    <div class="captcha-container">
      <div class="captcha-header">
        <span>安全验证</span>
        <el-icon @click="close"><Close /></el-icon>
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
              <el-icon v-else color="#67c23a"><Check /></el-icon>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { Close, Right, Check } from '@element-plus/icons-vue'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'success'): void
}>()

const sliderLeft = ref(0)
const completed = ref(false)
const isDragging = ref(false)
const startX = ref(0)
const trackWidth = 268

function close() {
  emit('update:visible', false)
}

function startDrag(e: MouseEvent | TouchEvent) {
  if (completed.value) return
  isDragging.value = true
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  startX.value = clientX - sliderLeft.value
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
}

function endDrag() {
  if (!isDragging.value) return
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
  document.removeEventListener('touchmove', onDrag)
  document.removeEventListener('touchend', endDrag)
  
  if (sliderLeft.value > trackWidth - 10) {
    sliderLeft.value = trackWidth
    completed.value = true
    setTimeout(() => {
      emit('success')
      close()
    }, 300)
  } else {
    sliderLeft.value = 0
  }
}

function reset() {
  sliderLeft.value = 0
  completed.value = false
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
.captcha-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.captcha-container {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
  width: 340px;
}
.captcha-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
  font-weight: 500;
  font-size: 15px;
}
.captcha-header .el-icon {
  cursor: pointer;
  opacity: 0.8;
}
.captcha-header .el-icon:hover {
  opacity: 1;
}
.captcha-body {
  padding: 20px;
}
.slider-container {
  padding: 0;
}
.slider-track {
  height: 40px;
  background: #f0f0f0;
  border-radius: 20px;
  position: relative;
  overflow: hidden;
  border: 1px solid #ddd;
  transition: background 0.3s;
}
.slider-track.completed {
  background: #e8f5e9;
  border-color: #81c784;
}
.slider-btn {
  position: absolute;
  top: 2px;
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #fff;
  font-size: 14px;
  box-shadow: 0 2px 8px rgba(102,126,234,0.4);
  transition: transform 0.1s;
  z-index: 2;
}
.slider-btn:hover {
  transform: scale(1.05);
}
.slider-btn:active {
  transform: scale(0.98);
}
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
  color: #999;
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
  color: #67c23a;
  font-weight: 500;
}
</style>
