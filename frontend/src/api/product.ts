import { userAPI } from './request'

export const getProducts = (type?: string) => {
  const params = type ? { type } : {}
  return userAPI.get<{ data: Product[] }>('/products', { params })
}

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
  duration_days?: number
  rpm_limit?: number
  tpm_limit?: number
  concurrent_limit?: number
  sort_order: number
  is_recommended: boolean
  is_hot: boolean
  status: string
  created_at: string
}

export const productApi = {
  list: (type?: string) => {
    const params = type ? { type } : {}
    return userAPI.get<{ data: Product[] }>('/products', { params })
  },
  getById: (id: number) => userAPI.get<{ data: Product }>(`/products/${id}`),
}
