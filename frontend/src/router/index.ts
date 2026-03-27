import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
    { path: '/register', name: 'Register', component: () => import('@/views/Register.vue'), meta: { title: '注册' } },
    { path: '/', name: 'Dashboard', component: () => import('@/views/Dashboard.vue'), meta: { requiresAuth: true, title: '控制台' } },
    { path: '/tokens', name: 'Tokens', component: () => import('@/views/tokens/List.vue'), meta: { requiresAuth: true, title: 'API 密钥' } },
    { path: '/products', name: 'Products', component: () => import('@/views/products/List.vue'), meta: { requiresAuth: true, title: '商品列表' } },
    { path: '/orders', name: 'Orders', component: () => import('@/views/orders/List.vue'), meta: { requiresAuth: true, title: '订单记录' } },
    { path: '/logs', name: 'APILogs', component: () => import('@/views/logs/ApiLogs.vue'), meta: { requiresAuth: true, title: 'API 调用记录' } },
    { path: '/vip', name: 'VIP', component: () => import('@/views/vip/Index.vue'), meta: { requiresAuth: true, title: 'VIP 会员' } },
    { path: '/profile', name: 'Profile', component: () => import('@/views/Profile.vue'), meta: { requiresAuth: true, title: '个人中心' } }
  ]
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) next('/login')
  else next()
})

export default router
