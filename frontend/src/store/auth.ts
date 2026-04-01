import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/api/request'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userData = ref(localStorage.getItem('user') || null)
  const user = computed(() => {
    if (userData.value) {
      try {
        return JSON.parse(userData.value)
      } catch {
        return null
      }
    }
    return null
  })

  const isLoggedIn = computed(() => !!token.value)

  async function login(email: string, password: string) {
    const { data } = await request.post('/user/login', { email, password })
    token.value = data.data.token
    userData.value = JSON.stringify(data.data.user)
    localStorage.setItem('token', token.value)
    localStorage.setItem('user', userData.value)
    return data
  }

  async function register(username: string, email: string, password: string) {
    const { data } = await request.post('/user/register', { username, email, password })
    return data
  }

  async function fetchProfile() {
    if (!token.value) return
    try {
      const { data } = await request.get('/user/profile')
      const user = data.data
      // Compute user status
      user.is_vip = isVIPUser(user)
      user.account_status = getAccountStatus(user)
      userData.value = JSON.stringify(user)
      localStorage.setItem('user', userData.value)
    } catch {
    }
  }

  function isVIPUser(user: any): boolean {
    if (!user) return false
    if (user.level && user.level !== 'free' && user.level.startsWith('vip')) {
      // Check if VIP not expired
      if (user.v_ip_expired_at) {
        const expiry = new Date(user.v_ip_expired_at)
        if (expiry > new Date()) return true
      }
      // Has VIP level without expiry = permanent VIP
      if (!user.v_ip_expired_at) return true
    }
    return false
  }

  function getAccountStatus(user: any): string {
    if (!user) return 'unknown'
    // VIP users (bronze/silver/gold with valid expiry or permanent)
    if (user.level && user.level !== 'free' && user.level.startsWith('vip')) {
      if (user.v_ip_expired_at) {
        const expiry = new Date(user.v_ip_expired_at)
        if (expiry > new Date()) return 'vip'
        return 'vip_expired'
      }
      return 'vip' // No expiry = permanent VIP
    }
    // Free users with quota = recharge users
    if ((user.free_quota > 0) || (user.v_ip_quota > 0)) {
      return 'recharge'
    }
    return 'free'
  }

  function logout() {
    token.value = ''
    userData.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isLoggedIn, login, register, fetchProfile, logout }
})

export const useAdminStore = defineStore('admin', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const isLoggedIn = computed(() => !!token.value)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('admin_token', t)
  }

  function logout() {
    token.value = ''
    localStorage.removeItem('admin_token')
  }

  return { token, isLoggedIn, setToken, logout }
})
