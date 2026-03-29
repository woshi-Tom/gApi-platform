import { adminAPI } from './request'

export interface SMTPConfig {
  enabled: boolean
  host: string
  port: number
  use_tls: boolean
  username: string
  password: string
  from_name: string
  from_email: string
}

export interface UpdateSMTPRequest {
  enabled: boolean
  host: string
  port: number
  use_tls: boolean
  username: string
  password: string
  from_name: string
  from_email: string
}

export interface TestEmailRequest {
  test_email: string
}

export interface RegisterSettings {
  allow_register: boolean
  require_email_verify: boolean
  smtp_enabled: boolean
  enable_captcha: boolean
  new_user_quota: number
  trial_vip_days: number
}

export const settingsAPI = {
  getSMTPConfig: () => {
    return adminAPI.get<SMTPConfig>('/settings/email')
  },
  
  updateSMTPConfig: (data: UpdateSMTPRequest) => {
    return adminAPI.put('/settings/email', data)
  },
  
  testSMTPConnection: (testEmail: string) => {
    return adminAPI.post<TestEmailRequest>('/settings/email/test', { test_email: testEmail })
  },

  getRegisterSettings: () => {
    return adminAPI.get<RegisterSettings>('/settings/register')
  },

  updateRegisterSettings: (data: Partial<RegisterSettings>) => {
    return adminAPI.put('/settings/register', data)
  }
}
