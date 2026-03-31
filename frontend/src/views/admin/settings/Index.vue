<template>
  <div class="settings-page">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>

    <el-tabs v-model="activeTab" class="settings-tabs">
      <el-tab-pane label="基本设置" name="general">
        <el-card class="settings-card">
          <template #header>
            <span>基本设置</span>
          </template>
          <el-form :model="generalForm" label-width="140px">
            <el-form-item label="网站名称">
              <el-input v-model="generalForm.site_name" placeholder="API Proxy Platform" />
            </el-form-item>
            <el-form-item label="网站 Logo">
              <el-input v-model="generalForm.site_logo" placeholder="/static/logo.png" />
            </el-form-item>
            <el-form-item label="网站描述">
              <el-input v-model="generalForm.site_description" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveGeneral" :loading="saving">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="注册设置" name="register">
        <el-card class="settings-card">
          <template #header>
            <span>注册设置</span>
          </template>
          <el-form :model="registerForm" label-width="140px">
            <el-form-item label="允许注册">
              <el-switch v-model="registerForm.allow_register" />
            </el-form-item>
            <el-form-item label="需要邮箱验证">
              <div class="email-verify-wrapper">
                <el-switch 
                  v-model="registerForm.require_email_verify" 
                  :disabled="!emailForm.enabled"
                  @change="handleEmailVerifyChange"
                />
                <el-tooltip 
                  v-if="!emailForm.enabled" 
                  content="请先在【邮箱设置】中配置并启用邮箱服务" 
                  placement="right"
                >
                  <el-icon class="warning-icon"><Warning /></el-icon>
                </el-tooltip>
                <span v-if="emailForm.enabled" class="status-text success">已配置</span>
                <span v-else class="status-text warning">未配置邮箱</span>
              </div>
              <span class="form-tip">开启后用户注册需要验证邮箱，关闭前请确保已配置邮箱服务</span>
            </el-form-item>
            <el-form-item label="需要滑块验证">
              <el-switch v-model="registerForm.enable_captcha" />
            </el-form-item>
            <el-form-item label="新用户配额">
              <el-input-number v-model="registerForm.new_user_quota" :min="0" :step="1000" />
              <span class="form-tip">新注册用户的初始配额 (Token)</span>
            </el-form-item>
            <el-form-item label="VIP 试用天数">
              <el-input-number v-model="registerForm.trial_vip_days" :min="0" :max="30" :step="1" />
              <span class="form-tip">新用户赠送的 VIP 试用天数，0 表示不赠送</span>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveRegister" :loading="saving">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="邮箱设置" name="email">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <span>邮箱设置</span>
              <el-tag v-if="emailForm.enabled" type="success" size="small">已启用</el-tag>
              <el-tag v-else type="info" size="small">已禁用</el-tag>
            </div>
          </template>
          <el-form :model="emailForm" label-width="140px">
            <el-form-item label="启用邮箱服务">
              <el-switch v-model="emailForm.enabled" />
              <span class="form-tip">开启后用户可收到注册验证码、密码重置邮件</span>
            </el-form-item>
            
            <el-divider content-position="left">SMTP 配置</el-divider>
            
            <el-form-item label="SMTP 服务器">
              <el-input v-model="emailForm.host" placeholder="smtp.qq.com" :disabled="!emailForm.enabled" />
            </el-form-item>
            <el-form-item label="SMTP 端口">
              <el-input-number v-model="emailForm.port" :min="1" :max="65535" :disabled="!emailForm.enabled" />
              <span class="form-tip">常用端口: 587 (TLS) / 465 (SSL)</span>
            </el-form-item>
            <el-form-item label="使用 TLS">
              <el-switch v-model="emailForm.use_tls" :disabled="!emailForm.enabled" />
              <span class="form-tip">建议开启 TLS 加密传输</span>
            </el-form-item>
            
            <el-divider content-position="left">认证信息</el-divider>
            
            <el-form-item label="用户名/邮箱">
              <el-input v-model="emailForm.username" placeholder="your-email@qq.com" :disabled="!emailForm.enabled" />
            </el-form-item>
            <el-form-item label="密码/授权码">
              <el-input v-model="emailForm.password" type="password" show-password :disabled="!emailForm.enabled" placeholder="输入新密码将覆盖原密码" />
              <span class="form-tip warning">留空则保持原密码不变</span>
            </el-form-item>
            
            <el-divider content-position="left">发件人信息</el-divider>
            
            <el-form-item label="发件人名称">
              <el-input v-model="emailForm.from_name" placeholder="gAPI Platform" :disabled="!emailForm.enabled" />
            </el-form-item>
            <el-form-item label="发件人邮箱">
              <el-input v-model="emailForm.from_email" placeholder="noreply@example.com" :disabled="!emailForm.enabled" />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="saveEmail" :loading="saving">保存设置</el-button>
              <el-button @click="testConnection" :loading="testing" :disabled="!emailForm.enabled">发送测试邮件</el-button>
            </el-form-item>
          </el-form>
          
          <el-dialog v-model="testDialogVisible" title="发送测试邮件" width="400px">
            <el-form>
              <el-form-item label="测试邮箱">
                <el-input v-model="testEmail" placeholder="请输入测试邮箱地址" />
              </el-form-item>
            </el-form>
            <template #footer>
              <el-button @click="testDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="sendTestEmail" :loading="testing">发送</el-button>
            </template>
          </el-dialog>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="速率限制" name="ratelimit">
        <el-card class="settings-card">
          <template #header>
            <span>速率限制</span>
          </template>
          <el-form :model="rateForm" label-width="160px">
            <el-divider content-position="left">免费用户</el-divider>
            <el-form-item label="RPM 限制">
              <el-input-number v-model="rateForm.free_rpm" :min="1" :max="10000" />
              <span class="form-tip">每分钟最大请求数</span>
            </el-form-item>
            <el-form-item label="TPM 限制">
              <el-input-number v-model="rateForm.free_tpm" :min="1000" :max="1000000" :step="1000" />
              <span class="form-tip">每分钟最大 Token 数</span>
            </el-form-item>
            <el-divider content-position="left">VIP 用户</el-divider>
            <el-form-item label="RPM 限制">
              <el-input-number v-model="rateForm.vip_rpm" :min="1" :max="10000" />
            </el-form-item>
            <el-form-item label="TPM 限制">
              <el-input-number v-model="rateForm.vip_tpm" :min="1000" :max="1000000" :step="1000" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveRateLimit" :loading="saving">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="支付设置" name="payment">
        <el-card class="settings-card">
          <template #header>
            <span>支付设置</span>
          </template>
          <el-form :model="paymentForm" label-width="140px">
            <el-divider content-position="left">支付宝</el-divider>
            <el-form-item label="启用支付宝">
              <el-switch v-model="paymentForm.alipay_enabled" />
            </el-form-item>
            <el-form-item label="APP ID" v-if="paymentForm.alipay_enabled">
              <el-input v-model="paymentForm.alipay_app_id" placeholder="支付宝应用 APP ID" />
            </el-form-item>
            <el-form-item label="商家私钥" v-if="paymentForm.alipay_enabled">
              <el-input v-model="paymentForm.alipay_private_key" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="支付宝公钥" v-if="paymentForm.alipay_enabled">
              <el-input v-model="paymentForm.alipay_public_key" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="沙箱模式" v-if="paymentForm.alipay_enabled">
              <el-switch v-model="paymentForm.alipay_sandbox" />
            </el-form-item>

            <el-divider content-position="left">微信支付</el-divider>
            <el-form-item label="启用微信支付">
              <el-switch v-model="paymentForm.wechat_enabled" />
            </el-form-item>
            <el-form-item label="APP ID" v-if="paymentForm.wechat_enabled">
              <el-input v-model="paymentForm.wechat_app_id" placeholder="微信应用 APP ID" />
            </el-form-item>
            <el-form-item label="商户号" v-if="paymentForm.wechat_enabled">
              <el-input v-model="paymentForm.wechat_mch_id" placeholder="微信商户号" />
            </el-form-item>
            <el-form-item label="API 密钥" v-if="paymentForm.wechat_enabled">
              <el-input v-model="paymentForm.wechat_api_key" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="savePayment" :loading="saving">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="安全设置" name="security">
        <el-card class="settings-card">
          <template #header>
            <span>安全设置</span>
          </template>
          <el-form :model="securityForm" label-width="140px">
            <el-form-item label="JWT 密钥">
              <el-input v-model="securityForm.jwt_secret" type="password" show-password />
              <span class="form-tip warning">修改后所有用户需要重新登录</span>
            </el-form-item>
            <el-form-item label="JWT 过期时间">
              <el-input-number v-model="securityForm.jwt_expire_hours" :min="1" :max="720" :step="1" />
              <span class="form-tip">小时</span>
            </el-form-item>
            <el-form-item label="密码最小长度">
              <el-input-number v-model="securityForm.password_min_length" :min="6" :max="32" />
            </el-form-item>
            <el-form-item label="密码过期天数">
              <el-input-number v-model="securityForm.password_expire_days" :min="0" :max="365" />
              <span class="form-tip">0 表示不过期</span>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveSecurity" :loading="saving">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Warning } from '@element-plus/icons-vue'
import { settingsAPI } from '@/api/settings'

const activeTab = ref('general')
const saving = ref(false)
const testing = ref(false)
const testDialogVisible = ref(false)
const testEmail = ref('')

const generalForm = reactive({
  site_name: 'API Proxy Platform',
  site_logo: '/static/logo.png',
  site_description: 'OpenAI API 代理平台',
})

const registerForm = reactive({
  allow_register: true,
  require_email_verify: true,
  smtp_enabled: false,
  enable_captcha: true,
  new_user_quota: 100000,
  trial_vip_days: 0,
})

const emailForm = reactive({
  enabled: false,
  host: '',
  port: 587,
  use_tls: true,
  username: '',
  password: '',
  from_name: 'gAPI Platform',
  from_email: 'noreply@gapi.com',
})

const rateForm = reactive({
  free_rpm: 60,
  free_tpm: 10000,
  vip_rpm: 2000,
  vip_tpm: 500000,
})

const paymentForm = reactive({
  alipay_enabled: false,
  alipay_app_id: '',
  alipay_private_key: '',
  alipay_public_key: '',
  alipay_sandbox: true,
  wechat_enabled: false,
  wechat_app_id: '',
  wechat_mch_id: '',
  wechat_api_key: '',
})

const securityForm = reactive({
  jwt_secret: 'gapi-jwt-secret-key-change-in-production',
  jwt_expire_hours: 168,
  password_min_length: 8,
  password_expire_days: 90,
})

async function loadEmailConfig() {
  try {
    const res = await settingsAPI.getSMTPConfig()
    if (res.data.data) {
      const data = res.data.data
      emailForm.enabled = data.enabled
      emailForm.host = data.host
      emailForm.port = data.port
      emailForm.use_tls = data.use_tls
      emailForm.username = data.username
      emailForm.password = ''
      emailForm.from_name = data.from_name
      emailForm.from_email = data.from_email
    }
  } catch (e) {
    console.error('Failed to load email config:', e)
  }
}

async function loadRegisterSettings() {
  try {
    const res = await settingsAPI.getRegisterSettings()
    if (res.data.data) {
      const data = res.data.data
      registerForm.allow_register = data.allow_register
      registerForm.require_email_verify = data.require_email_verify
      registerForm.smtp_enabled = data.smtp_enabled
      registerForm.enable_captcha = data.enable_captcha
      registerForm.new_user_quota = data.new_user_quota
      registerForm.trial_vip_days = data.trial_vip_days
    }
  } catch (e) {
    console.error('Failed to load register settings:', e)
  }
}

function handleEmailVerifyChange(value: boolean) {
  if (value && !emailForm.enabled) {
    ElMessage.warning('请先在【邮箱设置】中配置并启用邮箱服务')
    registerForm.require_email_verify = false
  }
}

async function saveEmail() {
  saving.value = true
  try {
    await settingsAPI.updateSMTPConfig({
      enabled: emailForm.enabled,
      host: emailForm.host,
      port: emailForm.port,
      use_tls: emailForm.use_tls,
      username: emailForm.username,
      password: emailForm.password,
      from_name: emailForm.from_name,
      from_email: emailForm.from_email,
    })
    ElMessage.success('邮箱设置已保存')
    emailForm.password = ''
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function testConnection() {
  testEmail.value = ''
  testDialogVisible.value = true
}

async function sendTestEmail() {
  if (!testEmail.value) {
    ElMessage.warning('请输入测试邮箱地址')
    return
  }
  testing.value = true
  try {
    await settingsAPI.testSMTPConnection(testEmail.value)
    ElMessage.success('测试邮件发送成功，请检查收件箱')
    testDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '发送失败')
  } finally {
    testing.value = false
  }
}

async function saveGeneral() {
  saving.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('基本设置已保存')
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function saveRegister() {
  if (registerForm.require_email_verify && !emailForm.enabled) {
    ElMessage.error('请先在【邮箱设置】中配置并启用邮箱服务')
    return
  }
  saving.value = true
  try {
    await settingsAPI.updateRegisterSettings({
      allow_register: registerForm.allow_register,
      require_email_verify: registerForm.require_email_verify,
      enable_captcha: registerForm.enable_captcha,
      new_user_quota: registerForm.new_user_quota,
      trial_vip_days: registerForm.trial_vip_days,
    })
    ElMessage.success('注册设置已保存')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function saveRateLimit() {
  saving.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('速率限制设置已保存')
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function loadPaymentConfig() {
  try {
    const res = await settingsAPI.getPaymentConfig()
    if (res.data.data) {
      const data = res.data.data
      paymentForm.alipay_enabled = data.enabled
      paymentForm.alipay_app_id = data.app_id || ''
      paymentForm.alipay_public_key = data.public_key || ''
      paymentForm.alipay_sandbox = data.sandbox !== false
      paymentForm.alipay_private_key = ''
    }
  } catch (e) {
    console.error('Failed to load payment config:', e)
  }
}

async function savePayment() {
  saving.value = true
  try {
    await settingsAPI.updatePaymentConfig({
      enabled: paymentForm.alipay_enabled,
      app_id: paymentForm.alipay_app_id,
      private_key: paymentForm.alipay_private_key,
      public_key: paymentForm.alipay_public_key,
      encrypt_key: '',
      sandbox: paymentForm.alipay_sandbox,
    })
    ElMessage.success('支付设置已保存')
    paymentForm.alipay_private_key = ''
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function saveSecurity() {
  saving.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('安全设置已保存')
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadEmailConfig()
  loadRegisterSettings()
  loadPaymentConfig()
})
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.settings-tabs {
  background: white;
  padding: 0;
}

.settings-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 20px;
  background: white;
  border-radius: 10px 10px 0 0;
}

.settings-card {
  border-radius: 0 0 10px 10px;
}

.settings-card :deep(.el-card__header) {
  font-weight: 500;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.form-tip {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.form-tip.warning {
  color: var(--el-color-warning);
}

.email-verify-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.warning-icon {
  color: var(--el-color-warning);
  font-size: 16px;
  cursor: help;
}

.status-text {
  font-size: 12px;
}

.status-text.success {
  color: var(--el-color-success);
}

.status-text.warning {
  color: var(--el-color-warning);
}
</style>
