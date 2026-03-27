import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const createRequest = (baseURL: string) => {
  const instance = axios.create({
    baseURL,
    timeout: 30000
  })

  instance.interceptors.request.use(config => {
    const token = localStorage.getItem('token') || localStorage.getItem('admin_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    const adminSecret = localStorage.getItem('admin_secret')
    if (adminSecret) {
      config.headers['X-Admin-Secret'] = adminSecret
    }
    return config
  })

  instance.interceptors.response.use(
    response => response,
    error => {
      if (error.response?.status === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('admin_token')
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
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

export const userAPI = createRequest('/api/v1')
export const adminAPI = createRequest('/api/v1/admin')

export default userAPI
