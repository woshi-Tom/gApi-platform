<template>
  <div class="page-header" :class="{ centered }">
    <div class="page-header__left">
      <h2 class="page-header__title">{{ title }}</h2>
      <p v-if="description || $slots.description" class="page-header__desc">
        <slot name="description">{{ description }}</slot>
      </p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  centered?: boolean
}>()
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header.centered {
  flex-direction: column;
  text-align: center;
  gap: 8px;
}

.page-header__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.4;
}

.page-header__desc {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.page-header__actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

/* 响应式 */
@media (max-width: 640px) {
  .page-header:not(.centered) {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .page-header__title {
    font-size: 18px;
  }
}
</style>