<template>
  <el-aside :width="width" class="sidebar" :class="{ 'sidebar--collapsed': collapsed }">
    <!-- Header: Collapse Button + Brand -->
    <div class="sidebar-header">
      <div class="sidebar-collapse-btn" @click="$emit('toggle')">
        <el-icon :size="18">
          <Fold v-if="!collapsed" />
          <Expand v-else />
        </el-icon>
      </div>
      <transition name="brand-fade">
        <router-link v-show="!collapsed" to="/" class="sidebar-brand">
          <span class="sidebar-brand__icon">G</span>
          <span class="sidebar-brand__text">gAPI</span>
        </router-link>
      </transition>
    </div>

    <!-- Navigation -->
    <div class="sidebar-nav">
      <el-menu
        class="sidebar-menu"
        :default-active="activePath"
        router
        background-color="transparent"
        text-color="var(--color-sidebar-text)"
        active-text-color="var(--color-sidebar-active-text)"
        :ellipsis="false"
      >
        <el-menu-item
          v-for="item in menuItems"
          :key="item.index"
          :index="item.index"
          class="sidebar-menu-item"
          :title="collapsed ? item.title : undefined"
        >
          <span class="sidebar-menu-item__icon">
            <el-icon :size="24">
              <component :is="item.icon" />
            </el-icon>
          </span>
          <span class="sidebar-menu-item__label">
            {{ item.title }}
          </span>
        </el-menu-item>
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
  Fold, Expand, Setting
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
/* ===== Base Sidebar ===== */
.sidebar {
  background:
    radial-gradient(circle at 80% 0%, rgba(99, 102, 241, 0.10), transparent 35%),
    radial-gradient(circle at 50% 100%, rgba(139, 92, 246, 0.06), transparent 35%),
    var(--gradient-sidebar);
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
  padding: 0 12px;
  border-bottom: 1px solid var(--color-sidebar-border);
  background: var(--color-sidebar-header-bg);
  backdrop-filter: blur(10px);
  flex-shrink: 0;
  transition: justify-content var(--transition-sidebar), padding var(--transition-sidebar);
}

.sidebar-collapse-btn {
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

.sidebar-collapse-btn:hover {
  color: #fff;
  background: rgba(99, 102, 241, 0.15);
}

/* ===== Brand area ===== */
.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
  margin-right: 8px;
  text-decoration: none;
  cursor: pointer;
  transition: opacity var(--transition-fast);
}

.sidebar-brand:hover {
  opacity: 0.85;
}

/* Brand icon: gradient square with "G" */
.sidebar-brand__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: #fff;
  font-size: 18px;
  font-weight: 800;
  flex-shrink: 0;
  line-height: 1;
}

.sidebar-brand__text {
  color: #fff;
  font-size: 26px;
  font-weight: 800;
  white-space: nowrap;
  letter-spacing: 0.02em;
}

/* Brand text fade transition */
.brand-fade-enter-active {
  transition: opacity 0.2s ease 0.05s, transform 0.2s ease 0.05s;
}

.brand-fade-leave-active {
  transition: opacity 0.1s ease, transform 0.1s ease;
}

.brand-fade-enter-from {
  opacity: 0;
  transform: translateX(-6px);
}

.brand-fade-leave-to {
  opacity: 0;
  transform: translateX(-6px);
}

/* ===== Navigation ===== */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 12px 0 8px;
}

/* Reset el-menu default border */
.sidebar-nav :deep(.el-menu) {
  border-right: none;
  padding: 0;
  background: transparent;
}

/* ============================================================
 * Sidebar Menu Item
 * One stable DOM structure for both expanded and collapsed states.
 * Expanded: icon + label.
 * Collapsed: same icon position, label hidden.
 * ============================================================ */
.sidebar-menu-item {
  box-sizing: border-box;
  display: flex !important;
  align-items: center;
  height: 54px;
  line-height: 54px;
  margin: 3px 8px;
  padding: 0 18px 0 24px !important;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.68) !important;
  transition:
    background var(--transition-fast),
    color var(--transition-fast);
  position: relative;
  overflow: hidden;
}

.sidebar-menu-item__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 54px;
  flex: 0 0 24px;
  line-height: 1;
}

.sidebar-menu-item__label {
  display: inline-block;
  margin-left: 14px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: inherit;
  transition:
    opacity var(--transition-fast),
    transform var(--transition-fast);
}

/* Hover */
.sidebar-menu-item:hover {
  background: rgba(255, 255, 255, 0.06) !important;
  color: rgba(255, 255, 255, 0.82) !important;
}

/* Active: accent background + left bar */
.sidebar-menu-item.is-active {
  background: rgba(99, 102, 241, 0.18) !important;
  color: #c7d2fe !important;
}

.sidebar-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 24px;
  background: linear-gradient(180deg, #6366f1, #8b5cf6);
  border-radius: 999px;
}

/* ============================================================
 * Collapsed State
 * Use the same menu item DOM.
 * Only change width and hide label.
 * Do not move the icon column vertically.
 * ============================================================ */

/* Header: collapse button stays at same left position */
.sidebar--collapsed .sidebar-header {
  justify-content: flex-start;
  padding: 0 12px;
}

/* Nav: center items */
.sidebar--collapsed .sidebar-nav :deep(.el-menu) {
  align-items: center;
  width: 100%;
}

.sidebar--collapsed .sidebar-menu-item {
  width: 56px;
  margin: 3px auto;
  padding: 0 !important;
  justify-content: center;
}

.sidebar--collapsed .sidebar-menu-item__icon {
  width: 24px;
  height: 54px;
  flex: 0 0 24px;
}

.sidebar--collapsed .sidebar-menu-item__label {
  display: none;
}

.sidebar--collapsed .sidebar-menu-item.is-active {
  background: rgba(99, 102, 241, 0.22) !important;
}

/* ===== Footer ===== */
.sidebar-footer {
  flex-shrink: 0;
  margin-top: auto;
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
  border-radius: 18px;
  cursor: pointer;
  color: var(--color-sidebar-text);
  white-space: nowrap;
  overflow: hidden;
  min-width: 0;
  flex: 1;
  background: rgba(255, 255, 255, 0.035);
  border: 1px solid rgba(255, 255, 255, 0.06);
  transition: color var(--transition-fast), background var(--transition-fast), border-color var(--transition-fast);
}

.sidebar-footer__user:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.10);
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