import { userAPI, adminAPI } from './request'

export interface Product {
  id: number
  name: string
  description: string
  product_type: 'recharge' | 'vip' | 'package'
  price: number
  original_price?: number
  quota: number
  bonus_quota?: number
  vip_days?: number
  vip_quota?: number
  sort_order: number
  is_recommended: boolean
  is_hot: boolean
  status: string
  created_at: string
}

export type Order = {
  id: number
  order_no: string
  order_type: 'recharge' | 'vip' | 'package'
  package_id: number
  package_name: string
  total_amount: number
  discount_amount: number
  pay_amount: number
  status: 'pending' | 'paid' | 'cancelled' | 'refunded' | 'expired'
  paid_at?: string
  created_at: string
  expire_at?: string
}

export interface CreateOrderRequest {
  package_id: number
  payment_method: 'alipay' | 'wechat'
}

export const productApi = {
  list: () => userAPI.get<{ data: Product[] }>('/products'),
  getById: (id: number) => userAPI.get<{ data: Product }>(`/products/${id}`),
}

export const userOrderApi = {
  list: (params?: { page?: number; page_size?: number }) =>
    userAPI.get<{ data: { list: Order[]; pagination: { total: number } } }>('/orders', { params }),
  create: (data: CreateOrderRequest) => userAPI.post<{ data: Order }>('/orders', data),
  getById: (id: number) => userAPI.get<{ data: Order }>(`/orders/${id}`),
}

export const adminOrderApi = {
  list: (params?: { page?: number; page_size?: number; status?: string; user_id?: number }) =>
    adminAPI.get<{ data: { list: Order[]; pagination: { total: number } } }>('/orders', { params }),
  getById: (id: number) => adminAPI.get<{ data: Order }>(`/orders/${id}`),
  updateStatus: (id: number, status: string) =>
    adminAPI.put(`/orders/${id}/status`, { status }),
}

export const paymentApi = {
  createAlipay: (orderId: number) =>
    userAPI.post<{ data: { payment_url: string; qr_code: string } }>('/payment/alipay', { order_id: orderId }),
  createWechat: (orderId: number) =>
    userAPI.post<{ data: { payment_url: string; qr_code: string } }>('/payment/wechat', { order_id: orderId }),
}
