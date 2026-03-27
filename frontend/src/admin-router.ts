import { createRouter, createWebHashHistory } from 'vue-router'
import LoginVue from './views/admin/Login.vue'

const adminRouter = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/login', name: 'AdminLogin', component: LoginVue, meta: { title: '管理员登录', requireAuth: false } },
    { path: '/dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/Dashboard.vue'), meta: { title: '仪表盘' } },
    { path: '/users', name: 'AdminUsers', component: () => import('@/views/admin/users/List.vue'), meta: { title: '用户管理' } },
    { path: '/products', name: 'AdminProducts', component: () => import('@/views/admin/products/List.vue'), meta: { title: '商品管理' } },
    { path: '/channels', name: 'AdminChannels', component: () => import('@/views/admin/channels/List.vue'), meta: { title: '渠道管理' } },
    { path: '/orders', name: 'AdminOrders', component: () => import('@/views/admin/orders/List.vue'), meta: { title: '订单管理' } },
    { path: '/logs', name: 'AdminLogs', component: () => import('@/views/admin/logs/Index.vue'), meta: { title: '操作日志' } },
    { path: '/settings', name: 'AdminSettings', component: () => import('@/views/admin/settings/Index.vue'), meta: { title: '系统设置' } },
    { path: '/change-password', name: 'AdminChangePassword', component: () => import('@/views/admin/ChangePassword.vue'), meta: { title: '修改密码' } },
    { path: '/:pathMatch(.*)*', redirect: '/login' }
  ]
})

adminRouter.beforeEach((to, from, next) => {
  const adminToken = localStorage.getItem('admin_token')
  
  if (to.path === '/login') {
    if (adminToken) {
      next('/dashboard')
    } else {
      next()
    }
  } else if (to.path === '/' || to.path === '#/') {
    next('/dashboard')
  } else if (!adminToken) {
    window.location.href = '/#/login'
    next(false)
  } else {
    next()
  }
})

export default adminRouter