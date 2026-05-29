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
      <router-link
        to="/"
        class="sidebar-brand"
        :aria-hidden="collapsed"
        :tabindex="collapsed ? -1 : 0"
      >
        <span class="sidebar-brand__icon">G</span>
        <span class="sidebar-brand__text">gAPI</span>
      </router-link>
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
        >
          <el-tooltip
            :disabled="!collapsed"
            :content="item.title"
            placement="right"
            :offset="14"
            popper-class="sidebar-collapsed-tooltip"
            :show-after="200"
            effect="dark"
          >
            <span class="sidebar-menu-item__content">
              <span class="sidebar-menu-item__icon">
                <el-icon :size="24">
                  <component :is="item.icon" />
                </el-icon>
              </span>
              <span class="sidebar-menu-item__label">
                {{ item.title }}
              </span>
            </span>
          </el-tooltip>
        </el-menu-item>
      </el-menu>
    </div>

    <!-- Footer: User Account + Settings -->
    <div
      class="sidebar-footer"
      :class="collapsed ? 'sidebar-footer--collapsed' : 'sidebar-footer--expanded'"
    >
      <div class="sidebar-footer__inner">
        <div
          class="sidebar-footer__user"
          @click="$emit('command', 'profile')"
        >
          <el-tooltip
            :disabled="!collapsed"
            :content="accountDisplay"
            placement="right"
            :offset="14"
            popper-class="sidebar-collapsed-tooltip"
            :show-after="200"
            effect="dark"
          >
            <el-avatar :size="32" class="sidebar-user-card__avatar">
              {{ avatarLetter }}
            </el-avatar>
          </el-tooltip>
          <div class="sidebar-user-card__info" :aria-hidden="collapsed">
            <span class="sidebar-user-card__name">{{ accountDisplay }}</span>
            <span
              class="sidebar-user-card__badge"
              :class="isVipUser ? 'sidebar-user-card__badge--vip' : 'sidebar-user-card__badge--free'"
            >
              {{ isVipUser ? 'VIP' : 'Free' }}
            </span>
          </div>
        </div>
        <div
          class="sidebar-footer__settings-btn"
          :class="{ 'sidebar-footer__settings-btn--active': settingsOpen }"
          @click.stop="$emit('toggle-settings')"
        >
          <div class="sidebar-footer__settings-surface">
            <el-tooltip
              content="设置"
              :disabled="settingsOpen"
              :placement="collapsed ? 'right' : 'top'"
              :offset="collapsed ? 14 : 8"
              :popper-class="collapsed ? 'sidebar-collapsed-tooltip' : undefined"
              :show-after="200"
              effect="dark"
            >
              <el-icon :size="20"><Setting /></el-icon>
            </el-tooltip>
          </div>
        </div>
      </div>
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
  settingsOpen: boolean
  user?: {
    username?: string
    email?: string
    level?: string
    is_vip?: boolean
    account_status?: string
    vip_expired_at?: string | null
  } | null
}>()

defineEmits<{
  toggle: []
  'toggle-settings': []
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

const accountDisplay = computed(() => {
  return props.user?.email || props.user?.username || '用户'
})

const isVipUser = computed(() => {
  const user = props.user
  if (!user) return false

  if (user.vip_expired_at) {
    const expiresAt = Date.parse(user.vip_expired_at)
    if (!Number.isNaN(expiresAt)) return expiresAt > Date.now()
  }

  if (user.account_status === 'vip') return true
  if (['vip_expired', 'free', 'recharge'].includes(user.account_status || '')) {
    return false
  }

  if (typeof user.is_vip === 'boolean') return user.is_vip
  return Boolean(user.level && user.level !== 'free' && user.level.startsWith('vip'))
})
</script>

<style scoped>
/* ===== Base Sidebar ===== */
.sidebar {
  --sidebar-menu-icon-size: 24px;
  --sidebar-menu-item-height: 54px;
  --sidebar-menu-item-expanded-margin-x: 8px;
  --sidebar-menu-item-expanded-width-offset: 16px;
  --sidebar-menu-item-collapsed-width: 56px;
  --sidebar-menu-item-collapsed-margin-x: 14px;
  --sidebar-menu-item-collapsed-padding-x: 16px;
  --sidebar-menu-label-width: 150px;
  --sidebar-footer-expanded-height: 66px;
  --sidebar-footer-collapsed-height: 96px;
  --sidebar-footer-inner-expanded-height: 50px;
  --sidebar-footer-inner-collapsed-height: 80px;
  --sidebar-footer-settings-size: 36px;
  --sidebar-footer-settings-expanded-top: 7px;
  --sidebar-footer-settings-collapsed-left: 24px;
  --sidebar-footer-settings-expanded-padding-total: 16px;
  --sidebar-footer-settings-path-x: calc(
    var(--sidebar-width) -
    var(--sidebar-footer-settings-expanded-padding-total) -
    var(--sidebar-footer-settings-size) -
    var(--sidebar-footer-settings-collapsed-left)
  );
  --sidebar-footer-settings-path-y: var(--sidebar-footer-settings-expanded-top);
  --sidebar-footer-settings-turn-delay: 72ms;
  --sidebar-footer-settings-x-duration: calc(
    var(--sidebar-transition-duration) - var(--sidebar-footer-settings-turn-delay)
  );
  --sidebar-footer-action-offset: 44px;
  --sidebar-footer-user-collapsed-size: 40px;
  --sidebar-footer-user-collapsed-x: 22px;
  --sidebar-footer-user-collapsed-y: 40px;

  background:
    radial-gradient(circle at 80% 0%, rgba(99, 102, 241, 0.10), transparent 35%),
    radial-gradient(circle at 50% 100%, rgba(139, 92, 246, 0.06), transparent 35%),
    var(--gradient-sidebar);
  border-right: 1px solid var(--color-sidebar-border);
  transition:
    width var(--sidebar-transition),
    inline-size var(--sidebar-transition),
    flex-basis var(--sidebar-transition) !important;
  overflow-x: hidden;
  overflow-y: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
  will-change: width;
}

/* ===== Header ===== */
.sidebar-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  border-bottom: 1px solid var(--color-sidebar-border);
  background: var(--color-sidebar-header-bg);
  backdrop-filter: blur(10px);
  flex-shrink: 0;
  transition: padding var(--sidebar-transition), gap var(--sidebar-transition);
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
  color: var(--color-sidebar-active-text);
  background: var(--color-sidebar-active-bg);
}

/* ===== Brand area ===== */
.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: 128px;
  min-width: 0;
  margin-left: 4px;
  text-decoration: none;
  cursor: pointer;
  overflow: hidden;
  opacity: 1;
  transform: translateX(0);
  user-select: none;
  -webkit-user-select: none;
  transition:
    max-width var(--sidebar-transition),
    margin var(--sidebar-transition),
    opacity 160ms var(--sidebar-transition-easing),
    transform var(--sidebar-transition);
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
  background: var(--gradient-primary);
  color: #fff;
  font-size: 18px;
  font-weight: 800;
  flex-shrink: 0;
  line-height: 1;
  user-select: none;
  -webkit-user-select: none;
}

.sidebar-brand__text {
  color: var(--color-text-primary);
  font-size: 26px;
  font-weight: 800;
  white-space: nowrap;
  letter-spacing: 0.02em;
  user-select: none;
  -webkit-user-select: none;
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
  width: 100%;
}

/* ============================================================
 * Sidebar Menu Item
 * One stable DOM structure for both expanded and collapsed states.
 * Expanded: icon + label.
 * Collapsed: same icon position, label hidden.
 * ============================================================ */
.sidebar .sidebar-menu .sidebar-menu-item {
  box-sizing: border-box;
  display: grid !important;
  grid-template-columns: var(--sidebar-menu-icon-size) minmax(0, 1fr);
  column-gap: 14px;
  align-items: center;
  width: calc(100% - var(--sidebar-menu-item-expanded-width-offset));
  height: var(--sidebar-menu-item-height);
  min-height: var(--sidebar-menu-item-height);
  line-height: 1;
  margin: 3px var(--sidebar-menu-item-expanded-margin-x) !important;
  padding: 0 18px 0 22px !important;
  border-radius: 14px !important;
  font-size: 15px;
  font-weight: 500;
  color: var(--color-sidebar-text) !important;
  background-color: transparent !important;
  transition:
    width var(--sidebar-transition),
    margin var(--sidebar-transition),
    padding var(--sidebar-transition),
    grid-template-columns var(--sidebar-transition),
    column-gap var(--sidebar-transition),
    border-radius var(--sidebar-transition),
    background-color var(--transition-fast),
    color var(--transition-fast);
  position: relative;
  overflow: hidden;
}

.sidebar .sidebar-menu-item__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--sidebar-menu-icon-size);
  height: var(--sidebar-menu-item-height);
  min-width: var(--sidebar-menu-icon-size);
  flex: 0 0 var(--sidebar-menu-icon-size);
  line-height: 1;
  overflow: hidden;
}

.sidebar .sidebar-menu-item__content {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: var(--sidebar-menu-icon-size) minmax(0, 1fr);
  column-gap: 14px;
  align-items: center;
  width: 100%;
  height: 100%;
  transition:
    grid-template-columns var(--sidebar-transition),
    column-gap var(--sidebar-transition);
}

.sidebar .sidebar-menu-item__icon :deep(.el-icon) {
  width: 24px !important;
  height: 24px !important;
  margin: 0 !important;
  font-size: 24px !important;
  line-height: 1 !important;
}

.sidebar .sidebar-menu-item__icon :deep(svg) {
  display: block;
  width: 24px;
  height: 24px;
}

.sidebar .sidebar-menu-item__label {
  display: block;
  width: max-content;
  min-width: 0;
  max-width: var(--sidebar-menu-label-width);
  margin: 0;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: inherit;
  opacity: 1;
  transform: translateX(0);
  transition:
    max-width var(--sidebar-transition),
    opacity 160ms var(--sidebar-transition-easing),
    transform var(--sidebar-transition);
}

/* Hover */
.sidebar .sidebar-menu .sidebar-menu-item:hover {
  background-color: var(--color-sidebar-hover) !important;
  color: var(--color-text-primary) !important;
}

/* Active: accent background + left bar */
.sidebar .sidebar-menu .sidebar-menu-item.is-active {
  background-color: var(--color-sidebar-active-bg) !important;
  color: var(--color-sidebar-active-text) !important;
}

.sidebar-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 24px;
  background: var(--gradient-primary);
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
  padding: 0 12px;
}

.sidebar--collapsed .sidebar-brand {
  max-width: 0;
  margin-left: 0;
  opacity: 0;
  pointer-events: none;
  transform: translateX(-6px);
}

/* Nav: center items */
.sidebar.sidebar--collapsed .sidebar-menu .sidebar-menu-item {
  grid-template-columns: var(--sidebar-menu-icon-size) 0;
  column-gap: 0;
  width: var(--sidebar-menu-item-collapsed-width);
  margin: 3px var(--sidebar-menu-item-collapsed-margin-x) !important;
  padding: 0 var(--sidebar-menu-item-collapsed-padding-x) !important;
  justify-content: start;
}

.sidebar.sidebar--collapsed .sidebar-menu-item__content {
  grid-template-columns: var(--sidebar-menu-icon-size) 0;
  column-gap: 0;
}

.sidebar.sidebar--collapsed .sidebar-menu-item__icon {
  width: var(--sidebar-menu-icon-size);
  height: var(--sidebar-menu-item-height);
  flex: 0 0 var(--sidebar-menu-icon-size);
}

.sidebar.sidebar--collapsed .sidebar-menu-item__label {
  max-width: 0;
  margin: 0;
  opacity: 0;
  pointer-events: none;
  transform: translateX(-6px);
}

.sidebar.sidebar--collapsed .sidebar-menu .sidebar-menu-item.is-active {
  background-color: var(--color-sidebar-active-bg) !important;
}

/* ===== Footer ===== */
.sidebar-footer {
  flex-shrink: 0;
  margin-top: auto;
  border-top: 1px solid var(--color-sidebar-border);
  padding: 8px;
  min-height: var(--sidebar-footer-expanded-height);
  transition:
    min-height var(--sidebar-transition),
    padding var(--sidebar-transition);
}

.sidebar-footer--expanded {
  min-height: var(--sidebar-footer-expanded-height);
}

.sidebar-footer--collapsed {
  min-height: var(--sidebar-footer-collapsed-height);
  padding: 8px 0;
}

.sidebar-footer__inner {
  position: relative;
  width: 100%;
  height: var(--sidebar-footer-inner-expanded-height);
  transition: height var(--sidebar-transition);
}

.sidebar-footer--collapsed .sidebar-footer__inner {
  height: var(--sidebar-footer-inner-collapsed-height);
}

/* ===== Expanded: User info (left side) ===== */
.sidebar-footer__user {
  position: absolute;
  top: 0;
  left: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  width: calc(100% - var(--sidebar-footer-action-offset));
  height: var(--sidebar-footer-inner-expanded-height);
  padding: 8px 10px;
  border-radius: 18px;
  cursor: pointer;
  color: var(--color-sidebar-text);
  white-space: nowrap;
  overflow: hidden;
  min-width: 0;
  background-color: var(--color-bg-hover);
  border: 1px solid var(--color-sidebar-border);
  transform: translate3d(0, 0, 0);
  transition:
    width var(--sidebar-transition),
    height var(--sidebar-transition),
    padding var(--sidebar-transition),
    gap var(--sidebar-transition),
    border-radius var(--sidebar-transition),
    transform var(--sidebar-transition),
    background-color var(--sidebar-transition),
    border-color var(--sidebar-transition),
    color var(--transition-fast);
  will-change: transform, width, height;
}

.sidebar-footer__user:hover {
  color: var(--color-text-primary);
  background-color: var(--color-sidebar-active-bg);
  border-color: var(--color-sidebar-active-bg);
}

/* ===== Expanded: Settings button (right side) ===== */
.sidebar-footer__settings-btn {
  position: absolute;
  top: 0;
  left: var(--sidebar-footer-settings-collapsed-left);
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--sidebar-footer-settings-size);
  height: var(--sidebar-footer-settings-size);
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-sidebar-text);
  flex-shrink: 0;
  transform: translate3d(var(--sidebar-footer-settings-path-x), 0, 0);
  transition:
    transform var(--sidebar-footer-settings-x-duration) var(--sidebar-transition-easing)
      var(--sidebar-footer-settings-turn-delay),
    color var(--transition-fast);
  will-change: transform;
}

.sidebar-footer__settings-surface {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border-radius: inherit;
  transform: translate3d(0, var(--sidebar-footer-settings-path-y), 0);
  transition:
    transform var(--sidebar-footer-settings-turn-delay) var(--sidebar-transition-easing),
    background-color var(--transition-fast);
  will-change: transform;
}

.sidebar-footer__settings-btn:hover {
  color: var(--color-sidebar-active-text);
}

.sidebar-footer__settings-btn:hover .sidebar-footer__settings-surface,
.sidebar-footer__settings-btn--active .sidebar-footer__settings-surface {
  background-color: var(--color-sidebar-active-bg);
}

.sidebar-footer__settings-btn--active {
  color: var(--color-sidebar-active-text);
}

.sidebar-footer--collapsed .sidebar-footer__settings-btn {
  transform: translate3d(0, 0, 0);
}

.sidebar-footer--collapsed .sidebar-footer__settings-surface {
  transform: translate3d(0, 0, 0);
}

.sidebar-footer--collapsed .sidebar-footer__user {
  justify-content: center;
  gap: 0;
  width: var(--sidebar-footer-user-collapsed-size);
  height: var(--sidebar-footer-user-collapsed-size);
  padding: 4px;
  border-radius: var(--radius-md);
  background-color: transparent;
  border-color: transparent;
  transform: translate3d(
    var(--sidebar-footer-user-collapsed-x),
    var(--sidebar-footer-user-collapsed-y),
    0
  );
}

.sidebar-footer--collapsed .sidebar-footer__user:hover {
  background-color: var(--color-sidebar-active-bg);
  border-color: transparent;
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
  max-width: 136px;
  flex: 1;
  overflow: hidden;
  opacity: 1;
  transform: translateX(0);
  transition:
    max-width var(--sidebar-transition),
    opacity 150ms var(--sidebar-transition-easing) 30ms,
    transform var(--sidebar-transition);
}

.sidebar-footer--collapsed .sidebar-user-card__info {
  max-width: 0;
  opacity: 0;
  pointer-events: none;
  transform: translateX(-6px);
  transition:
    max-width var(--sidebar-transition),
    opacity 120ms var(--sidebar-transition-easing),
    transform var(--sidebar-transition);
}

.sidebar-user-card__name {
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
  font-weight: var(--font-weight-medium);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

:global(.sidebar-collapsed-tooltip) {
  margin-left: 4px !important;
  margin-top: -4px !important;
}
</style>
