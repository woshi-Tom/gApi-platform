<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>支付宝支付</h2>
      <p class="subtitle">请使用手机支付宝扫描下方二维码完成支付</p>
    </div>

    <el-card class="payment-card" v-loading="loading || loadingOrderInfo">
      <el-result v-if="status === 'paid' || status === 'completed'" icon="success" title="支付成功">
        <template #sub-title>
          <p>您的订单已支付成功，配额已到账</p>
          <el-button type="primary" @click="$router.push('/profile')">查看配额</el-button>
        </template>
      </el-result>

      <el-result v-else-if="status === 'cancelled'" icon="warning" title="订单已取消">
        <template #sub-title>
          <el-button type="primary" @click="$router.push('/products')">重新购买</el-button>
        </template>
      </el-result>

      <template v-else>
        <el-row :gutter="20" align="middle">
          <el-col :span="12" class="qr-col">
            <div class="qr-wrap">
              <div v-if="qrCodeImage" style="text-align:center">
                <img :src="qrCodeImage" alt="支付宝二维码" class="qr-image" />
                <p style="color:green">支付宝扫码支付</p>
              </div>
              <div v-else class="qr-placeholder">
                <span v-if="loading">正在生成二维码...</span>
                <span v-else-if="status === 'expired'">订单已过期</span>
                <span v-else-if="status === 'pending'">二维码已过期</span>
                <span v-else>加载失败</span>
              </div>
              <div class="countdown" v-if="remainingSeconds >= 0 && qrCodeImage && status === 'pending'">
                二维码有效期: {{ countdownDisplay }}
              </div>
              <el-button v-if="!qrCodeImage && status === 'pending'" type="primary" :loading="loading" @click="startPayment" style="margin-top:12px">
                重新生成二维码
              </el-button>
              <el-button v-if="status === 'expired'" type="primary" @click="$router.push('/products')" style="margin-top:12px">
                重新购买
              </el-button>
            </div>
          </el-col>
          <el-col :span="12" class="info-col">
            <el-descriptions border title="订单信息" :column="1">
              <el-descriptions-item label="订单号">{{ orderNo || '-' }}</el-descriptions-item>
              <el-descriptions-item label="商品">{{ packageName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="金额">{{ amountDisplay ? '¥' + amountDisplay : '-' }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusType" size="small">{{ statusLabel }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>
            <el-button type="danger" plain style="margin-top:16px;width:100%" @click="cancelOrder">取消订单</el-button>
          </el-col>
        </el-row>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { paymentApi } from '@/api/order'
import request from '@/api/request'

const route = useRoute()
const router = useRouter()

const orderNo = ref<string>('')
const qrCodeUrl = ref<string>('')
const qrCodeImage = ref<string>('')
const expireAt = ref<string | null>(null)
const packageName = ref<string>('')
const amount = ref<number>(0)
const status = ref<'pending'|'paid'|'expired'|'cancelled'>('pending')
const loading = ref(false)
const loadingOrderInfo = ref(false)

const remainingSeconds = ref<number>(0)
let countdownTimer: number | null = null
let pollTimer: number | null = null

const amountDisplay = computed(() => {
  if (amount.value > 0) return amount.value.toFixed(2)
  return ''
})

const statusLabel = computed(() => {
  switch (status.value) {
    case 'completed':
    case 'paid': return '已支付'
    case 'pending': return '待支付'
    case 'expired': return '已过期'
    case 'cancelled': return '已取消'
    case 'refunded': return '已退款'
  }
  return status.value
})

const statusType = computed(() => {
  switch (status.value) {
    case 'completed':
    case 'paid': return 'success'
    case 'pending': return 'warning'
    case 'expired': return 'danger'
    case 'cancelled':
    case 'refunded': return 'info'
  }
  return 'info'
})

const countdownDisplay = computed(() => {
  const s = remainingSeconds.value
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${String(m).padStart(2,'0')}:${String(sec).padStart(2,'0')}`
})

async function fetchOrderInfo() {
  loadingOrderInfo.value = true
  try {
    const res = await request.get(`/orders/no/${orderNo.value}`)
    const data = res.data?.data || res.data
    packageName.value = data.package_name || ''
    amount.value = data.pay_amount || 0
    status.value = data.status || 'pending'
    if (data.alipay_qr_url) {
      qrCodeUrl.value = data.alipay_qr_url
      await generateQRCode(data.alipay_qr_url)
    }
    if (data.qr_expire_at) {
      expireAt.value = data.qr_expire_at
      initCountdown()
    }
  } catch (e: any) {
    ElMessage.error('获取订单信息失败')
  } finally {
    loadingOrderInfo.value = false
  }
}

async function startPayment() {
  if (!orderNo.value) {
    ElMessage.error('订单信息不完整')
    return
  }
  loading.value = true
  try {
    const res = await paymentApi.createAlipay(orderNo.value)
    const data = res?.data?.data || res?.data || {}
    if (data.qr_code) {
      qrCodeUrl.value = data.qr_code
      await generateQRCode(data.qr_code)
    }
    expireAt.value = data.qr_expire_at || null
    if (expireAt.value) {
      initCountdown()
    }
    startPolling()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '发起支付失败')
  } finally {
    loading.value = false
  }
}

async function generateQRCode(url: string) {
  if (!url) {
    qrCodeImage.value = ''
    return
  }
  try {
    qrCodeImage.value = await QRCode.toDataURL(url, {
      width: 256,
      margin: 2
    })
  } catch (err) {
    ElMessage.error('二维码生成失败')
    qrCodeImage.value = ''
  }
}

function initCountdown() {
  if (countdownTimer) window.clearInterval(countdownTimer)
  if (!expireAt.value) {
    remainingSeconds.value = 0
    return
  }
  const end = new Date(expireAt.value).getTime()
  const tick = () => {
    const left = Math.floor((end - Date.now()) / 1000)
    if (left <= 0) {
      remainingSeconds.value = 0
      if (countdownTimer) window.clearInterval(countdownTimer)
      countdownTimer = null
      qrCodeUrl.value = ''
      return
    }
    remainingSeconds.value = left
  }
  tick()
  countdownTimer = window.setInterval(tick, 1000)
}

function startPolling() {
  if (pollTimer) window.clearInterval(pollTimer)
  pollTimer = window.setInterval(async () => {
    if (!orderNo.value || status.value !== 'pending') return
    try {
      const res = await paymentApi.queryAlipay(orderNo.value)
      const data = res?.data?.data || res?.data || {}
      if (data.status) {
        status.value = data.status
      }
      if (data.qr_code && data.qr_code !== qrCodeUrl.value) {
        qrCodeUrl.value = data.qr_code
        await generateQRCode(data.qr_code)
      }
      if (data.qr_expire_at) {
        expireAt.value = data.qr_expire_at
        initCountdown()
      }
      if (status.value === 'paid' || status.value === 'completed') {
        stopAllTimers()
        ElMessage.success('支付成功！配额已到账')
        setTimeout(() => {
          router.push('/orders')
        }, 1500)
      } else if (status.value === 'expired' || status.value === 'cancelled') {
        stopAllTimers()
        qrCodeUrl.value = ''
        qrCodeImage.value = ''
      }
    } catch (err) {
      ElMessage.error('查询支付状态失败')
      stopAllTimers()
    }
  }, 3000)
}

function stopAllTimers() {
  if (countdownTimer) {
    window.clearInterval(countdownTimer)
    countdownTimer = null
  }
  if (pollTimer) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

async function cancelOrder() {
  try {
    const res = await request.get(`/orders/no/${orderNo.value}`)
    const data = res.data?.data || res.data
    if (data.status !== 'pending') {
      status.value = data.status
      ElMessage.warning('订单状态已变更，无法取消')
      return
    }
    await paymentApi.cancelAlipay(orderNo.value)
    stopAllTimers()
    ElMessage.success('订单已取消')
    router.back()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '取消失败')
  }
}

onMounted(async () => {
  const no = route.query.order_no as string
  if (no) {
    orderNo.value = no
    await fetchOrderInfo()
    if (status.value === 'pending') {
      await startPayment()
    } else if (status.value === 'paid' || status.value === 'completed') {
      // 订单已完成，直接跳转
      ElMessage.success('订单已完成')
      setTimeout(() => {
        router.push('/orders')
      }, 1500)
    }
  } else {
    ElMessage.error('无效的订单')
    router.push('/products')
  }
})

onUnmounted(() => {
  stopAllTimers()
})
</script>

<style scoped>
.payment-page {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}
.page-header {
  text-align: center;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0 0 8px;
}
.subtitle {
  color: var(--el-text-color-secondary);
  margin: 0;
}
.payment-card {
  padding: 10px;
}
.qr-col {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}
.qr-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.qr-image {
  width: 280px;
  height: 280px;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.qr-placeholder {
  width: 280px;
  height: 280px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 8px;
  color: var(--el-text-color-placeholder);
}
.countdown {
  margin-top: 12px;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}
.info-col {
  padding: 20px;
}
</style>
