import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '@/api/request'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<any>(null)

  const isLoggedIn = computed(() => !!token.value)

  async function login(email: string, password: string) {
    const { data } = await request.post('/user/login', { email, password })
    token.value = data.data.token
    user.value = data.data.user
    localStorage.setItem('token', token.value)
    return data
  }

  async function register(username: string, email: string, password: string) {
    const { data } = await request.post('/user/register', { username, email, password })
    return data
  }

  async function fetchProfile() {
    if (!token.value) return
    const { data } = await request.get('/user/profile')
    user.value = data.data
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
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
