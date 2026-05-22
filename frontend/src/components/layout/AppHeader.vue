<template>
  <el-header class="header">
    <span class="page-title">{{ title }}</span>
    <div class="header-right">
      <el-dropdown @command="(cmd: string) => $emit('command', cmd)">
        <span class="user-info">
          <el-icon><Avatar /></el-icon>
          {{ username }}
          <el-icon class="arrow"><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">个人中心</el-dropdown-item>
            <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>

<script setup lang="ts">
import { ArrowDown, Avatar } from '@element-plus/icons-vue'

defineProps<{
  title: string
  username: string
}>()

defineEmits<{
  command: [cmd: string]
}>()
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--color-border-light);
  padding: 0 var(--spacing-xl);
  height: var(--header-height);
}

.page-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  color: var(--color-text-regular);
  font-size: var(--font-size-base);
}

.user-info:hover {
  background: var(--color-primary-bg);
  color: var(--color-primary);
}

.arrow {
  margin-left: var(--spacing-xs);
  font-size: 12px;
  transition: transform var(--transition-fast);
}
</style>