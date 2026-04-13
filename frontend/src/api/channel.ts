import { adminAPI } from './request'

export type Channel = {
  id: number
  name: string
  type: string
  base_url: string
  api_key?: string
  models: string[]
  weight: number
  priority: number
  group_name: string
  status: number
  is_healthy: boolean
  failure_count: number
  last_success_at?: string
  last_check_at?: string
  last_error?: string
  response_time_avg: number
  timeout?: number
  proxy_enabled?: boolean
  proxy_type?: string
  proxy_url?: string
  created_at: string
  updated_at: string
}

export type ChannelTestRequest = {
  test_type: 'models' | 'chat' | 'embeddings'
  model?: string
  messages?: { role: string; content: string }[]
  input?: string
  temperature?: number
  max_tokens?: number
}

export type ChannelTestResult = {
  success: boolean
  response_time_ms: number
  status_code: number
  models?: string[]
  content?: string
  error?: string
}

export type ChannelHealthResult = {
  is_healthy: boolean
  failure_count: number
  last_check_at: string
  last_error?: string
  response_time_ms: number
}

export const channelApi = {
  list: (params?: { 
    page?: number; 
    page_size?: number; 
    type?: string; 
    status?: string;
    group?: string;
    keyword?: string;
  }) => adminAPI.get<{ 
    data: { list: Channel[]; pagination: { total: number; page: number; page_size: number } } 
  }>('/channels', { params }),
  
  create: (data: Partial<Channel>) => adminAPI.post('/channels', data),
  
  update: (id: number, data: Partial<Channel>) => adminAPI.put(`/channels/${id}`, data),
  
  delete: (id: number) => adminAPI.delete(`/channels/${id}`),
  
  test: (id: number, data: ChannelTestRequest) => 
    adminAPI.post<{ data: ChannelTestResult }>(`/channels/${id}/test`, data),
  
  enable: (id: number) => adminAPI.post(`/channels/${id}/enable`),
  
  disable: (id: number) => adminAPI.post(`/channels/${id}/disable`),
  
  triggerHealthCheck: (id: number) => 
    adminAPI.post<{ data: ChannelHealthResult }>(`/channels/${id}/health`),
}

export const CHANNEL_TYPES = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'NVIDIA NIM', value: 'nvidia' },
  { label: 'Azure OpenAI', value: 'azure' },
  { label: 'Claude (Anthropic)', value: 'claude' },
  { label: 'Google Gemini', value: 'gemini' },
  { label: 'DeepSeek', value: 'deepseek' },
  { label: '智谱 ChatGLM', value: 'zhipu' },
  { label: '百度千帆', value: 'baidu' },
  { label: '零一万物 Yi', value: 'yi' },
  { label: 'Groq', value: 'groq' },
  { label: 'Ollama (本地)', value: 'ollama' },
  { label: 'LocalAI', value: 'localai' },
  { label: '自定义', value: 'custom' },
]

export const CHANNEL_STATUS = [
  { label: '禁用', value: 0, type: 'danger' },
  { label: '启用', value: 1, type: 'success' },
  { label: '维护中', value: 2, type: 'warning' },
]
