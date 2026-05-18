import { ref, watch, onMounted } from 'vue'

/**
 * 数字滚动动画 composable
 * @param target 目标数字（响应式 getter）
 * @param duration 动画持续时间（毫秒）
 */
export function useCountUp(target: () => number, duration = 800) {
  const display = ref(0)
  let animFrame: number | null = null

  function animate(from: number, to: number) {
    if (animFrame) cancelAnimationFrame(animFrame)

    const startTime = performance.now()
    const diff = to - from

    function step(currentTime: number) {
      const elapsed = currentTime - startTime
      const progress = Math.min(elapsed / duration, 1)
      // easeOutCubic 缓动
      const eased = 1 - Math.pow(1 - progress, 3)
      display.value = Math.round(from + diff * eased)

      if (progress < 1) {
        animFrame = requestAnimationFrame(step)
      }
    }

    animFrame = requestAnimationFrame(step)
  }

  watch(target, (newVal, oldVal) => {
    animate(oldVal ?? 0, newVal ?? 0)
  })

  onMounted(() => {
    animate(0, target())
  })

  return display
}