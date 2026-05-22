<template>
  <el-aside :width="width" class="sidebar">
    <div class="logo">
      <el-icon class="logo-icon"><Monitor /></el-icon>
      <span v-show="!collapsed">gAPI 平台</span>
    </div>
    <el-menu 
      :default-active="activePath" 
      router 
      background-color="transparent" 
      text-color="var(--color-sidebar-text)" 
      active-text-color="var(--color-sidebar-active-text)"
      :ellipsis="false"
      :collapse="collapsed"
    >
      <el-menu-item index="/">
        <el-icon><HomeFilled /></el-icon>
        <template #title>控制台</template>
      </el-menu-item>
      <el-menu-item index="/tokens">
        <el-icon><Key /></el-icon>
        <template #title>API 密钥</template>
      </el-menu-item>
      <el-menu-item index="/products">
        <el-icon><ShoppingCart /></el-icon>
        <template #title>商品列表</template>
      </el-menu-item>
      <el-menu-item index="/orders">
        <el-icon><List /></el-icon>
        <template #title>订单记录</template>
      </el-menu-item>
      <el-menu-item index="/vip">
        <el-icon><Star /></el-icon>
        <template #title>VIP 会员</template>
      </el-menu-item>
      <el-menu-item index="/redeem">
        <el-icon><Ticket /></el-icon>
        <template #title>兑换码</template>
      </el-menu-item>
      <el-menu-item index="/profile">
        <el-icon><User /></el-icon>
        <template #title>个人中心</template>
      </el-menu-item>
    </el-menu>
    <div class="sidebar-toggle" @click="$emit('toggle')">
      <el-icon :size="18">
        <Fold v-if="!collapsed" />
        <Expand v-else />
      </el-icon>
    </div>
  </el-aside>
</template>

<script setup lang="ts">
import {
  HomeFilled, Key, ShoppingCart, List, Star, User, Ticket,
  Monitor, Fold, Expand
} from '@element-plus/icons-vue'

defineProps<{
  collapsed: boolean
  activePath: string
  width: string
}>()

defineEmits<{
  toggle: []
}>()
</script>

<style scoped>
.sidebar {
  background: var(--gradient-sidebar);
  border-right: 1px solid var(--color-sidebar-border);
  transition: width var(--transition-base) !important;
  overflow-x: hidden;
  position: relative;
}

.logo {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff;
  font-size: 18px;
  font-weight: var(--font-weight-semibold);
  background: var(--color-sidebar-header-bg);
  border-bottom: 1px solid var(--color-sidebar-border);
  white-space: nowrap;
  overflow: hidden;
  backdrop-filter: blur(10px);
}

.logo span {
  transition: opacity var(--transition-fast);
}

.logo-icon {
  font-size: 22px;
  color: var(--color-primary-light);
}

.sidebar-toggle {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--color-sidebar-text);
  border-top: 1px solid var(--color-sidebar-border);
  background: rgba(0, 0, 0, 0.1);
  transition: color var(--transition-fast), background var(--transition-fast);
}

.sidebar-toggle:hover {
  color: #fff;
  background: rgba(99, 102, 241, 0.1);
}
</style>