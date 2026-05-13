import { adminAPI } from './request'

export interface ModelGroup {
  id: number
  tenant_id: number
  name: string
  display_name: string
  description: string
  sort_order: number
  status: string
  created_at: string
  updated_at: string
}

export interface ModelPricing {
  id: number
  tenant_id: number
  model: string
  provider: string
  display_name: string
  price_input: number
  price_output: number
  ability_types: string | string[]
  context_length: number
  max_output: number
  group_id: number | null
  is_enabled: boolean
  is_featured: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface UserGroupRelation {
  id: number
  user_id: number
  group_id: number
  created_at: string
}

export const modelGroupAPI = {
  list: (params?: { page?: number; page_size?: number; status?: string }) =>
    adminAPI.get<{ data: { list: ModelGroup[]; pagination: { total: number } } }>('/model-groups', { params }),

  listAll: () =>
    adminAPI.get<{ data: ModelGroup[] }>('/model-groups/all'),

  create: (data: { name: string; display_name: string; description?: string; sort_order?: number }) =>
    adminAPI.post('/model-groups', data),

  update: (id: number, data: { display_name?: string; description?: string; sort_order?: number; status?: string }) =>
    adminAPI.put(`/model-groups/${id}`, data),

  delete: (id: number) =>
    adminAPI.delete(`/model-groups/${id}`),

  getChannels: (id: number) =>
    adminAPI.get(`/model-groups/${id}/channels`),

  addChannel: (id: number, data: { channel_id: number; priority?: number }) =>
    adminAPI.post(`/model-groups/${id}/channels`, data),

  removeChannel: (id: number, cid: number) =>
    adminAPI.delete(`/model-groups/${id}/channels/${cid}`),
}

export const modelPricingAPI = {
  list: (params?: { page?: number; page_size?: number; provider?: string }) =>
    adminAPI.get<{ data: { list: ModelPricing[]; pagination: { total: number } } }>('/model-pricing', { params }),

  listAll: () =>
    adminAPI.get<{ data: ModelPricing[] }>('/model-pricing/all'),

  getByModel: (model: string) =>
    adminAPI.get(`/model-pricing/model/${model}`),

  create: (data: Partial<ModelPricing>) =>
    adminAPI.post('/model-pricing', data),

  update: (id: number, data: Partial<ModelPricing>) =>
    adminAPI.put(`/model-pricing/${id}`, data),

  delete: (id: number) =>
    adminAPI.delete(`/model-pricing/${id}`),
}

export const userGroupAPI = {
  getUserGroups: (userId: number) =>
    adminAPI.get<{ data: UserGroupRelation[] }>(`/users/${userId}/groups`),

  setUserGroups: (userId: number, groupIds: number[]) =>
    adminAPI.put(`/users/${userId}/groups`, { group_ids: groupIds }),
}
