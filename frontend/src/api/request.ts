import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const createRequest = (baseURL: string, tokenKey: string, clearKeys: string[]) => {
  const instance = axios.create({
    baseURL,
    timeout: 30000
  })

  instance.interceptors.request.use(config => {
    const token = localStorage.getItem(tokenKey)
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })

  instance.interceptors.response.use(
    response => response,
    error => {
      if (error.response?.status === 401) {
        clearKeys.forEach(key => localStorage.removeItem(key))
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
      } else if (error.response?.status === 403 && error.response?.data?.error?.message?.includes('admin')) {
        // Skip showing error for admin auth issues, let the page handle it
      } else if (error.response?.data?.error?.message) {
        ElMessage.error(error.response.data.error.message)
      } else if (error.message) {
        ElMessage.error(error.message)
      }
      return Promise.reject(error)
    }
  )

  return instance
}

export const userAPI = createRequest('/api/v1', 'token', ['token'])
export const adminAPI = createRequest('/api/v1/admin', 'admin_token', ['admin_token', 'admin_secret'])

export default userAPI
