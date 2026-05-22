<template>
  <div class="page-header" :class="{ centered }">
    <div class="page-header__left">
      <div class="page-header__title-row">
        <div class="page-header__accent"></div>
        <h2 class="page-header__title">{{ title }}</h2>
      </div>
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
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-base);
  border-bottom: 1px solid var(--color-border-light);
}

.page-header.centered {
  flex-direction: column;
  text-align: center;
  gap: var(--spacing-sm);
  border-bottom: none;
}

.page-header__left {
  min-width: 0;
}

.page-header__title-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.page-header__accent {
  width: 4px;
  height: 22px;
  background: var(--gradient-primary);
  border-radius: var(--radius-full);
  flex-shrink: 0;
}

.centered .page-header__accent {
  display: none;
}

.page-header__title {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  line-height: 1.4;
  letter-spacing: -0.02em;
}

.page-header__desc {
  margin: var(--spacing-xs) 0 0;
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.page-header__actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

/* 响应式 */
@media (max-width: 640px) {
  .page-header:not(.centered) {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .page-header__title {
    font-size: var(--font-size-xl);
  }
}
</style>