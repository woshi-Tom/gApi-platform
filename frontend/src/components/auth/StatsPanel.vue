<template>
  <div class="stats-panel animate-fade-up" :style="{ animationDelay: delay + 'ms' }">
    <div class="stats-header">
      <span class="stats-dot"></span>
      <span class="stats-title">{{ title }}</span>
      <span class="stats-live">{{ liveLabel }}</span>
    </div>
    <div class="stats-grid">
      <div v-for="(stat, i) in stats" :key="i" class="stat-item">
        <div class="stat-value" :style="{ animationDelay: (delay + 200 + i * 200) + 'ms' }">
          <span
            class="stat-number"
            :class="{
              'stat-green': stat.color === 'green',
              'stat-purple': stat.color === 'purple'
            }"
          >{{ stat.value }}</span>
        </div>
        <div class="stat-label">{{ stat.label }}</div>
        <div class="stat-bar">
          <div
            class="stat-bar-fill"
            :class="{
              'fill-green': stat.color === 'green',
              'fill-purple': stat.color === 'purple'
            }"
            :style="{
              animationDelay: (delay + 400 + i * 200) + 'ms',
              '--fill-width': (stat.fillWidth || '50%')
            }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface StatItem {
  value: string
  label: string
  color?: 'default' | 'green' | 'purple'
  fillWidth?: string
}

withDefaults(defineProps<{
  title: string
  liveLabel: string
  stats: StatItem[]
  delay?: number
}>(), {
  delay: 600,
})
</script>

<style scoped>
.stats-panel {
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 12px 18px 14px;
  max-width: 460px;
  overflow: hidden;
  position: relative;
}

.stats-panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 28px;
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 12px 12px 0 0;
}

.stats-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  position: relative;
  z-index: 1;
}

.stats-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.4);
  animation: pulse-icon 2s ease-in-out infinite;
}

.stats-title {
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.6);
  flex: 1;
}

.stats-live {
  font-size: 11px;
  color: rgba(34, 197, 94, 0.8);
  background: rgba(34, 197, 94, 0.08);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.stats-grid {
  display: flex;
  gap: 16px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.stat-value {
  opacity: 0;
  animation: terminal-line 0.4s ease-out forwards;
}

.stat-number {
  font-size: 20px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.stat-number.stat-green {
  color: #4ade80;
}

.stat-number.stat-purple {
  background: linear-gradient(135deg, #a5b4fc, #818cf8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.stat-label {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.25);
  font-weight: 400;
}

.stat-bar {
  height: 3px;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.06);
  overflow: hidden;
  margin-top: 4px;
}

.stat-bar-fill {
  height: 100%;
  border-radius: 2px;
  width: 0;
  background: linear-gradient(90deg, #6366f1, #818cf8);
  animation: stats-bar-grow 1s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  opacity: 0;
}

.stat-bar-fill.fill-green {
  background: linear-gradient(90deg, #22c55e, #4ade80);
}

.stat-bar-fill.fill-purple {
  background: linear-gradient(90deg, #6366f1, #a78bfa);
}

@keyframes stats-bar-grow {
  from {
    width: 0;
    opacity: 0;
  }
  to {
    width: var(--fill-width, 50%);
    opacity: 1;
  }
}
</style>