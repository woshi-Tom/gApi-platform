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
              <el-switch v-model="registerForm.email_verify" />
            </el-form-item>
            <el-form-item label="需要滑块验证">
              <el-switch v-model="registerForm.captcha_verify" />
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
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('general')
const saving = ref(false)

const generalForm = reactive({
  site_name: 'API Proxy Platform',
  site_logo: '/static/logo.png',
  site_description: 'OpenAI API 代理平台',
})

const registerForm = reactive({
  allow_register: true,
  email_verify: true,
  captcha_verify: true,
  new_user_quota: 100000,
  trial_vip_days: 0,
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
  saving.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('注册设置已保存')
  } catch (e) {
    ElMessage.error('保存失败')
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

async function savePayment() {
  saving.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    ElMessage.success('支付设置已保存')
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

.form-tip {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.form-tip.warning {
  color: var(--el-color-warning);
}
</style>
