<template>
  <el-main class="main-content">
    <router-view v-slot="{ Component }">
      <transition name="page-fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </el-main>
</template>

<script setup lang="ts">
</script>

<style scoped>
.main-content {
  padding: 24px;
  overflow-y: auto;
  background: var(--color-bg-page);
}
</style>
