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
      // Compute is_vip from level field
      user.is_vip = computeIsVip(user.level)
      userData.value = JSON.stringify(user)
      localStorage.setItem('user', userData.value)
    } catch {
    }
  }

  function computeIsVip(level: string | undefined): boolean {
    if (!level) return false
    return level !== 'free' && level.startsWith('vip')
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
