import { userAPI, adminAPI } from './request'

export interface Token {
  id: number
  name: string
  token_key: string
  token_key_full?: string
  allowed_models: string[]
  allowed_ips: string[]
  expires_at?: string
  status: string
  remain_quota: number
  used_quota: number
  created_at: string
  last_used_at?: string
}

export interface CreateTokenRequest {
  name: string
  allowed_models?: string[]
  allowed_ips?: string[]
  expires_at?: string
  quota?: number
}

export const userTokenApi = {
  list: () => userAPI.get<{ data: Token[] }>('/tokens'),
  create: (data: CreateTokenRequest) => userAPI.post<{ data: Token }>('/tokens', data),
  delete: (id: number) => userAPI.delete(`/tokens/${id}`),
}

export interface AdminToken {
  id: number
  user_id: number
  username?: string
  name: string
  token_key: string
  status: string
  remain_quota: number
  used_quota: number
  created_at: string
}

export const adminTokenApi = {
  list: (params?: { user_id?: number; page?: number; page_size?: number }) =>
    adminAPI.get<{ data: { list: AdminToken[]; pagination: { total: number } } }>('/tokens', { params }),
  create: (userId: number, data: CreateTokenRequest) =>
    adminAPI.post('/tokens', { user_id: userId, ...data }),
  update: (id: number, data: Partial<Token>) =>
    adminAPI.put(`/tokens/${id}`, data),
  delete: (id: number) =>
    adminAPI.delete(`/tokens/${id}`),
  resetQuota: (id: number, quota: number) =>
    adminAPI.post(`/tokens/${id}/reset-quota`, { quota }),
}
