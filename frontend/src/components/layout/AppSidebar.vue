<template>
  <el-aside :width="width" class="sidebar">
    <!-- Header: Logo + Collapse Button -->
    <div class="sidebar-header">
      <div class="sidebar-logo">
        <el-icon class="sidebar-logo__icon"><Monitor /></el-icon>
        <span v-show="!collapsed" class="sidebar-logo__text">gAPI 平台</span>
      </div>
      <div class="sidebar-collapse-btn" @click="$emit('toggle')">
        <el-icon :size="18">
          <Fold v-if="!collapsed" />
          <Expand v-else />
        </el-icon>
      </div>
    </div>

    <!-- Navigation -->
    <div class="sidebar-nav">
      <el-menu
        :default-active="activePath"
        router
        background-color="transparent"
        text-color="var(--color-sidebar-text)"
        active-text-color="var(--color-sidebar-active-text)"
        :ellipsis="false"
        :collapse="collapsed"
      >
        <template v-for="item in menuItems" :key="item.index">
          <el-tooltip
            v-if="collapsed"
            :content="item.title"
            placement="right"
            :show-after="200"
            effect="dark"
          >
            <el-menu-item :index="item.index">
              <el-icon><component :is="item.icon" /></el-icon>
              <template #title>{{ item.title }}</template>
            </el-menu-item>
          </el-tooltip>
          <el-menu-item v-else :index="item.index">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </div>

    <!-- Footer: User Account + Settings -->
    <div
      class="sidebar-footer"
      :class="collapsed ? 'sidebar-footer--collapsed' : 'sidebar-footer--expanded'"
    >
      <!-- Collapsed: vertical stack (settings on top, avatar below) -->
      <template v-if="collapsed">
        <div class="sidebar-footer__settings-icon" @click="$emit('command', 'settings')">
          <el-tooltip content="设置" placement="right" :show-after="200" effect="dark">
            <el-icon :size="20"><Setting /></el-icon>
          </el-tooltip>
        </div>
        <div class="sidebar-footer__avatar-only" @click="$emit('command', 'profile')">
          <el-tooltip :content="user?.username || '用户'" placement="right" :show-after="200" effect="dark">
            <el-avatar :size="32" class="sidebar-user-card__avatar">
              {{ avatarLetter }}
            </el-avatar>
          </el-tooltip>
        </div>
      </template>

      <!-- Expanded: horizontal row (user info left, settings right) -->
      <template v-else>
        <div class="sidebar-footer__user" @click="$emit('command', 'profile')">
          <el-avatar :size="32" class="sidebar-user-card__avatar">
            {{ avatarLetter }}
          </el-avatar>
          <div class="sidebar-user-card__info">
            <span class="sidebar-user-card__name">{{ user?.username || '用户' }}</span>
            <span v-if="user?.level && user.level !== 'free'" class="sidebar-user-card__badge sidebar-user-card__badge--vip">VIP</span>
            <span v-else class="sidebar-user-card__badge sidebar-user-card__badge--free">Free</span>
          </div>
        </div>
        <div class="sidebar-footer__settings-btn" @click="$emit('command', 'settings')">
          <el-tooltip content="设置" placement="top" :show-after="200" effect="dark">
            <el-icon :size="20"><Setting /></el-icon>
          </el-tooltip>
        </div>
      </template>
    </div>
  </el-aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  HomeFilled, Key, ShoppingCart, List, Star, User, Ticket,
  Monitor, Fold, Expand, Setting
} from '@element-plus/icons-vue'

const props = defineProps<{
  collapsed: boolean
  activePath: string
  width: string
  user?: {
    username?: string
    email?: string
    level?: string
    is_vip?: boolean
  } | null
}>()

defineEmits<{
  toggle: []
  command: [cmd: string]
}>()

const menuItems = [
  { index: '/', icon: HomeFilled, title: '控制台' },
  { index: '/tokens', icon: Key, title: 'API 密钥' },
  { index: '/products', icon: ShoppingCart, title: '商品列表' },
  { index: '/orders', icon: List, title: '订单记录' },
  { index: '/vip', icon: Star, title: 'VIP 会员' },
  { index: '/redeem', icon: Ticket, title: '兑换码' },
  { index: '/profile', icon: User, title: '个人中心' },
]

const avatarLetter = computed(() => {
  const name = props.user?.username || props.user?.email || 'U'
  return name.charAt(0).toUpperCase()
})
</script>

<style scoped>
.sidebar {
  background: var(--gradient-sidebar);
  border-right: 1px solid var(--color-sidebar-border);
  transition: width var(--transition-sidebar) !important;
  overflow-x: hidden;
  overflow-y: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
}

/* ===== Header ===== */
.sidebar-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  border-bottom: 1px solid var(--color-sidebar-border);
  background: var(--color-sidebar-header-bg);
  backdrop-filter: blur(10px);
  flex-shrink: 0;
  gap: 8px;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
  font-size: 18px;
  font-weight: var(--font-weight-semibold);
  white-space: nowrap;
  overflow: hidden;
  min-width: 0;
}

.sidebar-logo__icon {
  font-size: 22px;
  color: var(--color-primary-light);
  flex-shrink: 0;
}

.sidebar-logo__text {
  transition: opacity var(--transition-fast);
}

.sidebar-collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--color-sidebar-text);
  flex-shrink: 0;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.sidebar-collapse-btn:hover {
  color: #fff;
  background: rgba(99, 102, 241, 0.15);
}

/* ===== Navigation ===== */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
}

/* ===== Footer ===== */
.sidebar-footer {
  flex-shrink: 0;
  border-top: 1px solid var(--color-sidebar-border);
  padding: 8px;
  transition: padding var(--transition-sidebar);
}

/* Expanded: horizontal row — [avatar + name] left, [settings] right */
.sidebar-footer--expanded {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

/* Collapsed: vertical stack — [settings] on top, [avatar] below */
.sidebar-footer--collapsed {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 0;
}

/* ===== Expanded: User info (left side) ===== */
.sidebar-footer__user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-sidebar-text);
  white-space: nowrap;
  overflow: hidden;
  min-width: 0;
  flex: 1;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.sidebar-footer__user:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.06);
}

/* ===== Expanded: Settings button (right side) ===== */
.sidebar-footer__settings-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-sidebar-text);
  flex-shrink: 0;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.sidebar-footer__settings-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.06);
}

/* ===== Collapsed: Settings icon (centered) ===== */
.sidebar-footer__settings-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-sidebar-text);
  transition: color var(--transition-fast), background var(--transition-fast);
}

.sidebar-footer__settings-icon:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.06);
}

/* ===== Collapsed: Avatar only (centered) ===== */
.sidebar-footer__avatar-only {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 4px 0;
  transition: opacity var(--transition-fast);
}

.sidebar-footer__avatar-only:hover {
  opacity: 0.85;
}

/* ===== Shared: Avatar ===== */
.sidebar-user-card__avatar {
  flex-shrink: 0;
  background: var(--gradient-primary);
  color: #fff;
  font-size: 14px;
  font-weight: var(--font-weight-semibold);
}

/* ===== Shared: User info text ===== */
.sidebar-user-card__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
  overflow: hidden;
}

.sidebar-user-card__name {
  font-size: var(--font-size-sm);
  color: #fff;
  font-weight: var(--font-weight-medium);
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-user-card__badge {
  font-size: 11px;
  font-weight: var(--font-weight-medium);
  padding: 1px 6px;
  border-radius: var(--radius-full);
  line-height: 1.4;
  width: fit-content;
}

.sidebar-user-card__badge--vip {
  background: var(--gradient-vip);
  color: #fff;
}

.sidebar-user-card__badge--free {
  background: rgba(148, 163, 184, 0.15);
  color: var(--color-sidebar-text);
}
</style>