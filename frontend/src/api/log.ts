import { adminAPI } from './request'

export { adminAPI }

export interface AuditLog {
  id: number
  action: string
  action_group: string
  resource_type?: string
  resource_id?: number
  user_id?: number
  username?: string
  request_ip?: string
  request_path?: string
  request_method?: string
  success: boolean
  error_message?: string
  old_value?: Record<string, any>
  new_value?: Record<string, any>
  created_at: string
}

export interface AuditLogQuery {
  page?: number
  page_size?: number
  action?: string
  action_group?: string
  user_id?: number
  resource_type?: string
  resource_id?: number
  start_time?: string
  end_time?: string
  success?: boolean
}

export const auditLogApi = {
  list: (params?: AuditLogQuery) =>
    adminAPI.get<{ 
      data: { 
        list: AuditLog[]; 
        pagination: { total: number; page: number; page_size: number } 
      } 
    }>('/logs/operation', { params }),
  
  export: (params?: AuditLogQuery) =>
    adminAPI.get('/logs/operation/export', { params, responseType: 'blob' }),
}

export const ACTION_GROUPS = [
  { label: '全部', value: '' },
  { label: '认证', value: 'auth' },
  { label: '用户', value: 'user' },
  { label: '渠道', value: 'channel' },
  { label: 'Token', value: 'token' },
  { label: '订单', value: 'order' },
  { label: '支付', value: 'payment' },
  { label: '配额', value: 'quota' },
  { label: 'VIP', value: 'vip' },
  { label: '系统', value: 'system' },
]

export const ACTIONS = {
  auth: ['user.login', 'user.logout', 'user.register', 'admin.login'],
  user: ['user.create', 'user.update', 'user.delete', 'user.enable', 'user.disable', 'user.quota_add'],
  channel: ['channel.create', 'channel.update', 'channel.delete', 'channel.enable', 'channel.disable', 'channel.test'],
  token: ['token.create', 'token.update', 'token.delete', 'token.reset_quota'],
  order: ['order.create', 'order.paid', 'order.cancelled', 'order.refunded'],
  payment: ['payment.init', 'payment.success', 'payment.failed', 'payment.callback'],
  quota: ['quota.recharge', 'quota.deduct', 'quota.expire'],
  vip: ['vip.activate', 'vip.expired', 'vip.cancelled'],
  system: ['system.config', 'system.backup'],
}
