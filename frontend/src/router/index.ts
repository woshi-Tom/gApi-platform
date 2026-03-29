import { createRouter, createWebHistory } from 'vue-router'
import axios from 'axios'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
    { path: '/register', name: 'Register', component: () => import('@/views/Register.vue'), meta: { title: '注册' } },
    { path: '/forgot-password', name: 'ForgotPassword', component: () => import('@/views/ForgotPassword.vue'), meta: { title: '忘记密码' } },
    { path: '/reset-password', name: 'ResetPassword', component: () => import('@/views/ResetPassword.vue'), meta: { title: '重置密码' } },
    { path: '/', name: 'Dashboard', component: () => import('@/views/Dashboard.vue'), meta: { requiresAuth: true, title: '控制台' } },
    { path: '/tokens', name: 'Tokens', component: () => import('@/views/tokens/List.vue'), meta: { requiresAuth: true, title: 'API 密钥' } },
    { path: '/products', name: 'Products', component: () => import('@/views/products/List.vue'), meta: { requiresAuth: true, title: '商品列表' } },
    { path: '/orders', name: 'Orders', component: () => import('@/views/orders/List.vue'), meta: { requiresAuth: true, title: '订单记录' } },
    { path: '/logs', name: 'APILogs', component: () => import('@/views/logs/ApiLogs.vue'), meta: { requiresAuth: true, title: 'API 调用记录' } },
    { path: '/vip', name: 'VIP', component: () => import('@/views/vip/Index.vue'), meta: { requiresAuth: true, title: 'VIP 会员' } },
    { path: '/profile', name: 'Profile', component: () => import('@/views/Profile.vue'), meta: { requiresAuth: true, title: '个人中心' } }
  ]
})

const apiBase = import.meta.env.VITE_API_BASE_URL || '/api'

function getAdminUrl() {
  const currentHost = window.location.host
  const [hostname, port] = currentHost.split(':')
  return `http://${hostname}:5174`
}

async function checkInitialization() {
  try {
    const response = await axios.get(`${apiBase}/v1/init/status`)
    return response.data?.data?.needs_init === false || response.data?.needs_init === false
  } catch {
    return true
  }
}

let initChecked = false

router.beforeEach(async (to, _from, next) => {
  const token = localStorage.getItem('token')
  const path = to.path
  
  if (path === '/login' || path === '/register') {
    if (!initChecked) {
      const initialized = await checkInitialization()
      initChecked = true
      if (!initialized) {
        window.location.href = `${getAdminUrl()}/#/init`
        return
      }
    }
    if (token && path === '/login') {
      next('/')
    } else {
      next()
    }
    return
  }
  
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
