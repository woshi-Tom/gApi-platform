import { adminAPI, userAPI } from './request'

export type RedemptionCode = {
  id: number
  tenant_id: number
  code: string
  code_type: string
  quota: number
  quota_type: string
  vip_days: number
  is_permanent: boolean
  max_uses: number
  used_count: number
  valid_from?: string
  valid_until?: string
  batch_id: string
  status: string
  created_by: number
  created_at: string
}

export type RedemptionUsage = {
  id: number
  code_id: number
  user_id: number
  quota_granted: number
  vip_granted: boolean
  vip_days: number
  redeemed_at: string
  ip_address: string
  user_agent: string
}

export type CreateCodeRequest = {
  prefix: string
  count: number
  code_type: string
  quota?: number
  quota_type?: string
  vip_days?: number
  max_uses?: number
  valid_from?: string
  valid_until?: string
}

export const redemptionApi = {
  list: (params?: {
    page?: number
    page_size?: number
    code_type?: string
    status?: string
    batch_id?: string
  }) => adminAPI.get<{
    data: { list: RedemptionCode[]; pagination: { total: number; page: number; page_size: number } }
  }>('/redemption/codes', { params }),

  create: (data: CreateCodeRequest) => adminAPI.post('/redemption/codes', data),

  disable: (id: number) => adminAPI.post(`/redemption/codes/${id}/disable`),

  getUsage: (id: number) => adminAPI.get<{ data: RedemptionUsage[] }>(`/redemption/codes/${id}/usage`),
}

export const userRedemptionApi = {
  redeem: (code: string) => userAPI.post('/redemption/redeem', { code }),

  getHistory: () => userAPI.get<{ data: RedemptionUsage[] }>('/redemption/history'),
}

export const CODE_TYPES = [
  { label: '配额充值', value: 'recharge' },
  { label: 'VIP开通', value: 'vip' },
  { label: '免费配额', value: 'quota' },
]

export const CODE_STATUS = [
  { label: '激活', value: 'active', type: 'success' },
  { label: '禁用', value: 'disabled', type: 'danger' },
  { label: '已过期', value: 'expired', type: 'warning' },
  { label: '已用完', value: 'used', type: 'info' },
]

export const QUOTA_TYPES = [
  { label: '永久配额', value: 'permanent' },
  { label: 'VIP配额', value: 'vip' },
]

export const formatCode = (code: string): string => {
  if (code.length <= 8) return code
  const prefix = code.substring(0, 4)
  const suffix = code.substring(code.length - 4)
  return `${prefix}****${suffix}`
}
