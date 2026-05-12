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
  enable_captcha: boolean
  new_user_quota: number
  trial_vip_days: number
  allowed_domains: string
  max_accounts_per_ip: number
  min_password_length: number
  signup_reward_type: string
  signup_reward_amount: number
}

export interface PaymentConfig {
  enabled: boolean
  app_id: string
  public_key: string
  sandbox: boolean
}

export interface UpdatePaymentRequest {
  enabled: boolean
  app_id: string
  private_key: string
  public_key: string
  encrypt_key: string
  sandbox: boolean
}

export interface GeneralSettings {
  site_name: string
  site_logo: string
  site_description: string
}

export interface RateLimitSettings {
  free_rpm: number
  free_tpm: number
  vip_rpm: number
  vip_tpm: number
}

export interface SecuritySettings {
  jwt_secret: string
  jwt_expire_hours: number
  password_min_length: number
  password_expire_days: number
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
  },

  getPaymentConfig: () => {
    return adminAPI.get<PaymentConfig>('/settings/payment')
  },

  updatePaymentConfig: (data: UpdatePaymentRequest) => {
    return adminAPI.put('/settings/payment', data)
  },

  getGeneralSettings: () => {
    return adminAPI.get<GeneralSettings>('/settings/general')
  },

  updateGeneralSettings: (data: GeneralSettings) => {
    return adminAPI.put('/settings/general', data)
  },

  getRateLimitSettings: () => {
    return adminAPI.get<RateLimitSettings>('/settings/rate-limit')
  },

  updateRateLimitSettings: (data: RateLimitSettings) => {
    return adminAPI.put('/settings/rate-limit', data)
  },

  getSecuritySettings: () => {
    return adminAPI.get<SecuritySettings>('/settings/security')
  },

  updateSecuritySettings: (data: SecuritySettings) => {
    return adminAPI.put('/settings/security', data)
  },
}
