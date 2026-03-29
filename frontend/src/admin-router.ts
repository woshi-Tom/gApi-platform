import { createRouter, createWebHashHistory } from 'vue-router'
import axios from 'axios'
import LoginVue from './views/admin/Login.vue'

const apiBase = import.meta.env.VITE_API_BASE_URL || '/api'

const adminRouter = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/init', name: 'InitWizard', component: () => import('@/views/admin/InitWizard.vue') },
    { path: '/login', name: 'AdminLogin', component: LoginVue },
    { path: '/dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/Dashboard.vue') },
    { path: '/users', name: 'AdminUsers', component: () => import('@/views/admin/users/List.vue') },
    { path: '/products', name: 'AdminProducts', component: () => import('@/views/admin/products/List.vue') },
    { path: '/channels', name: 'AdminChannels', component: () => import('@/views/admin/channels/List.vue') },
    { path: '/orders', name: 'AdminOrders', component: () => import('@/views/admin/orders/List.vue') },
    { path: '/logs', name: 'AdminLogs', component: () => import('@/views/admin/logs/Index.vue') },
    { path: '/settings', name: 'AdminSettings', component: () => import('@/views/admin/settings/Index.vue') },
    { path: '/change-password', name: 'AdminChangePassword', component: () => import('@/views/admin/ChangePassword.vue') }
  ]
})

async function checkInitStatus() {
  try {
    const response = await axios.get(`${apiBase}/v1/init/status`)
    return response.data?.data?.needs_init === false || response.data?.needs_init === false
  } catch {
    return false
  }
}

let initStatusChecked = false

adminRouter.beforeEach(async (to, from, next) => {
  const path = to.path
  const adminToken = localStorage.getItem('admin_token')
  
  if (path === '/init') {
    if (!initStatusChecked) {
      const initialized = await checkInitStatus()
      initStatusChecked = true
      if (initialized) {
        next('/login')
        return
      }
    }
    next()
    return
  }
  
  if (path === '/login') {
    if (adminToken) {
      next('/dashboard')
    } else {
      next()
    }
    return
  }
  
  if (path === '/') {
    next('/dashboard')
    return
  }
  
  if (!adminToken) {
    next('/login')
    return
  }
  
  next()
})

export default adminRouter