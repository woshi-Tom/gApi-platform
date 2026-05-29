import { userAPI, adminAPI } from './request'

export interface User {
  id: number
  username: string
  email: string
  phone?: string
  level: string
  vip_expired_at?: string
  vip_package_id?: number
  remain_quota: number
  vip_quota: number
  used_quota?: number
  status: string
  disabled_reason?: string
  last_login_at?: string
  last_login_ip?: string
  email_verified: boolean
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
  turnstileToken: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: string
  user: User
}

export interface SendRegisterEmailCodeRequest {
  email: string
  purpose?: 'register'
  turnstileToken: string
}

export interface SendResetEmailCodeRequest {
  email: string
  purpose: 'reset'
  captcha_token: string
}

export type SendEmailCodeRequest = SendRegisterEmailCodeRequest | SendResetEmailCodeRequest

export interface VerifyEmailCodeRequest {
  email: string
  code: string
  purpose?: 'register' | 'reset'
}

export const authApi = {
  login: (data: LoginRequest) => userAPI.post<{ data: LoginResponse }>('/user/login', data),
  register: (data: RegisterRequest) => userAPI.post('/user/register', data),
  getProfile: () => userAPI.get<{ data: User }>('/user/profile'),
  updateProfile: (data: Partial<User>) => userAPI.put('/user/profile', data),
  changePassword: (oldPassword: string, newPassword: string) => 
    userAPI.post('/user/change-password', { old_password: oldPassword, new_password: newPassword }),
  getQuota: () => userAPI.get<{ data: { remain_quota: number; vip_quota: number; is_vip: boolean; level: string } }>('/user/quota'),
  getVIPStatus: () => userAPI.get('/user/vip/status'),
}

export const emailApi = {
  sendCode: (data: SendEmailCodeRequest) => userAPI.post('/email/send-code', data),
  verifyCode: (data: VerifyEmailCodeRequest) => userAPI.post('/email/verify-code', data),
}

export interface UpdateUserRequest {
  level?: string
  status?: string
  quota_adjust?: number
  vip_quota_adjust?: number
  vip_expired_at?: string
  disabled_reason?: string
}

export const adminUserApi = {
  listUsers: (params?: { 
    page?: number; 
    page_size?: number; 
    level?: string; 
    status?: string; 
    keyword?: string 
  }) => adminAPI.get<{ 
    data: { 
      list: User[]; 
      pagination: { total: number; page: number; page_size: number } 
    } 
  }>('/users', { params }),
  updateUser: (id: number, data: UpdateUserRequest) => adminAPI.put(`/users/${id}`, data),
}

export const adminAuthApi = {
  login: (username: string, password: string) => {
    localStorage.setItem('admin_secret', 'gapi-admin-secret-key-2026')
    return adminAPI.post('/login', { username, password })
  },
  changePassword: (oldPassword: string, newPassword: string) =>
    adminAPI.post('/change-password', { old_password: oldPassword, new_password: newPassword }),
}

export const userManagementApi = adminUserApi
