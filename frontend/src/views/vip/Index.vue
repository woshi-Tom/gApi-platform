<template>
  <div class="vip-page">
    <PageHeader title="VIP 会员" description="开通 VIP 享受更多特权，提升 API 调用体验" centered />

    <div class="benefits-bar">
      <div class="benefit">
        <el-icon class="benefit-icon"><Star /></el-icon>
        <span>专属配额</span>
      </div>
      <div class="benefit">
        <el-icon class="benefit-icon"><TrendCharts /></el-icon>
        <span>更高 RPM</span>
      </div>
      <div class="benefit">
        <el-icon class="benefit-icon"><Headset /></el-icon>
        <span>优先支持</span>
      </div>
      <div class="benefit">
        <el-icon class="benefit-icon"><Timer /></el-icon>
        <span>优先响应</span>
      </div>
    </div>

    <el-alert
      v-if="isVIP"
      type="success"
      :closable="false"
      show-icon
      class="vip-status-alert"
    >
      <template #title>
        <span>您已是 {{ getTierName(vipStatus?.level) }} 会员</span>
        <span class="days-remaining">剩余 {{ vipStatus?.days_remaining || 0 }} 天</span>
      </template>
      <template #default>
        <div class="vip-info-row">
          <span>VIP附赠额度：{{ formatQuota(vipStatus?.vip_quota || 0) }}</span>
          <span class="quota-tip">（30天有效，续费可累加）</span>
        </div>
      </template>
    </el-alert>

    <div class="packages-grid" v-if="pkgs.length">
      <el-card 
        v-for="pkg in pkgs" 
        :key="pkg.id" 
        class="package-card"
        :class="{ recommended: pkg.is_recommended, popular: pkg.is_hot }"
        shadow="hover"
      >
        <div class="package-badge" v-if="pkg.is_recommended || pkg.is_hot">
          {{ pkg.is_recommended ? '推荐' : '热门' }}
        </div>
        
        <div class="package-header">
          <h3 class="package-name">{{ pkg.name }}</h3>
          <p class="package-desc">{{ pkg.description || 'VIP 会员套餐' }}</p>
        </div>
        
        <div class="price-section">
          <span class="currency">¥</span>
          <span class="price">{{ pkg.price }}</span>
          <span class="original-price" v-if="pkg.original_price">¥{{ pkg.original_price }}</span>
          <span class="period">/ {{ pkg.duration_days }}天</span>
        </div>
        
        <el-divider style="margin: 16px 0" />
        
        <ul class="features-list">
          <li>
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>月度订阅，有效期 {{ pkg.duration_days }} 天</span>
          </li>
          <li>
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>附赠配额 {{ formatQuota(pkg.vip_quota) }} Token（当月有效）</span>
          </li>
          <li>
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>可创建 {{ pkg.concurrent_limit || 3 }} 个 API 密钥</span>
          </li>
          <li v-if="pkg.rpm_limit">
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>RPM 限制 {{ pkg.rpm_limit }}</span>
          </li>
          <li v-if="pkg.tpm_limit">
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>TPM 限制 {{ formatQuota(pkg.tpm_limit) }}</span>
          </li>
        </ul>
        
        <el-button 
          v-if="!isVIP"
          :type="pkg.is_recommended ? 'warning' : 'primary'" 
          size="large" 
          class="buy-btn"
          @click="buy(pkg)"
        >
          立即开通
        </el-button>
        <el-button
          v-else
          type="warning"
          size="large"
          class="buy-btn"
          @click="renew(pkg)"
        >
          续费
        </el-button>
      </el-card>
    </div>

    <el-empty v-else description="暂无 VIP 套餐" />

    <el-card class="info-card">
      <template #header>
        <span>套餐说明</span>
      </template>
      <div class="info-content">
        <h4>额度消耗规则</h4>
        <div class="consumption-flow">
          <div class="flow-step">
            <span class="step-num">1</span>
            <span class="step-text">优先消耗免费额度（7天有效）</span>
          </div>
          <div class="flow-arrow">↓</div>
          <div class="flow-step">
            <span class="step-num">2</span>
            <span class="step-text">消耗充值套餐（按购买顺序，FIFO）</span>
          </div>
          <div class="flow-arrow">↓</div>
          <div class="flow-step">
            <span class="step-num">3</span>
            <span class="step-text">最后消耗 VIP 额度</span>
          </div>
        </div>
        <h4>VIP 附赠额度说明</h4>
        <ul>
          <li>VIP 附赠额度有效期 30 天</li>
          <li>到期前续费，有效期自动累加</li>
          <li>VIP 结束时（时间到期或额度用完），附赠额度清零</li>
        </ul>
        <h4>VIP 专属优惠</h4>
        <ul>
          <li>VIP 用户购买充值套餐享受 9 折优惠</li>
          <li>可叠加充值套餐作为额度补充</li>
        </ul>
      </div>
    </el-card>

    <el-card class="faq-card" v-if="pkgs.length">
      <template #header>
        <span>常见问题</span>
      </template>
      <el-collapse>
        <el-collapse-item title="VIP 什么情况下会结束？" name="1">
          <p>VIP 结束条件（满足任一即结束）：</p>
          <ul>
            <li><strong>时间到期</strong>：VIP 开通后 30 天到期</li>
            <li><strong>额度用完</strong>：VIP 附赠额度消耗完毕</li>
          </ul>
          <p>例如：开通 VIP 30天/1M额度，第20天用完1M额度 → VIP 立即结束，剩余10天不继承</p>
        </el-collapse-item>
        <el-collapse-item title="免费额度和充值额度有什么区别？" name="2">
          <p><strong>免费额度</strong>：注册即得，7天有效，到期清零。</p>
          <p><strong>充值额度</strong>：购买获得，3-7天有效，可叠加购买，按购买顺序消耗（FIFO）。</p>
          <p><strong>VIP 附赠额度</strong>：开通 VIP 获得，30天有效，续费可累加。</p>
        </el-collapse-item>
        <el-collapse-item title="到期前可以续费吗？" name="3">
          <p>可以。在 VIP 到期前随时续费，续费后：</p>
          <ul>
            <li>有效期 = 当前到期时间 + 30 天</li>
            <li>VIP 额度 = 当前剩余额度 + 套餐额度</li>
          </ul>
        </el-collapse-item>
        <el-collapse-item title="VIP 用户可以购买充值套餐吗？" name="4">
          <p>可以。VIP 用户可以购买充值套餐作为额度补充。</p>
          <p><strong>VIP 专属优惠</strong>：VIP 用户购买充值套餐享受 9 折优惠。</p>
        </el-collapse-item>
        <el-collapse-item title="多个充值套餐如何消耗？" name="5">
          <p>按购买时间顺序先进先出（FIFO）消耗：</p>
          <p>例如：购买套餐A → 购买套餐B → 消耗顺序：A(先用) → 用完/过期 → B(后用)</p>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check, Star, TrendCharts, Headset, Timer } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { getProducts } from '@/api/product'
import request from '@/api/request'
import { formatQuota } from '@/composables/useFormat'

interface VIPPackage {
  id: number
  name: string
  description: string
  price: number
  original_price?: number
  duration_days: number
  vip_quota: number
  quota: number
  rpm_limit?: number
  tpm_limit?: number
  concurrent_limit?: number
  is_recommended: boolean
  is_hot: boolean
}

interface VIPStatus {
  level: string
  vip_expired_at: string
  vip_quota: number
  is_vip: boolean
  days_remaining: number
}

const pkgs = ref<VIPPackage[]>([])
const isVIP = ref(false)
const vipStatus = ref<VIPStatus | null>(null)
const router = useRouter()

function formatExpiry(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getTierName(level: string | undefined): string {
  if (!level) return 'VIP'
  const tierNames: Record<string, string> = {
    'vip_bronze': 'VIP青铜',
    'vip_silver': 'VIP白银',
    'vip_gold': 'VIP黄金',
    'vip': 'VIP'
  }
  return tierNames[level] || 'VIP'
}

async function loadVIPStatus() {
  try {
    const res = await request.get('/user/vip/status')
    if (res.data?.data) {
      vipStatus.value = res.data.data
      isVIP.value = res.data.data.is_vip || false
      if (isVIP.value && vipStatus.value?.vip_expired_at) {
        const expiry = new Date(vipStatus.value.vip_expired_at)
        const now = new Date()
        const diff = expiry.getTime() - now.getTime()
        vipStatus.value.days_remaining = Math.ceil(diff / (1000 * 60 * 60 * 24))
      }
    }
  } catch (e) {
    ElMessage.error('加载VIP状态失败')
  }
}

async function load() {
  try {
    const res = await getProducts('vip')
    pkgs.value = res.data.data || []
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '加载失败')
  }
}

async function buy(pkg: VIPPackage) {
  try {
    const res = await request.post('/orders', {
      package_id: pkg.id,
      package_type: 'vip',
      payment_method: 'alipay',
    })
    const data = res.data?.data || res.data
    ElMessage.success('订单已创建，正在跳转支付页面...')
    router.push(`/payment?order_no=${data.order_no}`)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '创建订单失败')
  }
}

async function renew(pkg: VIPPackage) {
  try {
    const res = await request.post('/orders', {
      package_id: pkg.id,
      package_type: 'vip',
      payment_method: 'alipay',
    })
    const data = res.data?.data || res.data
    ElMessage.success('续费订单已创建，正在跳转支付页面...')
    router.push(`/payment?order_no=${data.order_no}`)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '创建续费订单失败')
  }
}

onMounted(() => {
  loadVIPStatus()
  load()
})
</script>

<style scoped>
.vip-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}


/* Benefits Bar */
.benefits-bar {
  display: flex;
  justify-content: center;
  gap: 40px;
  padding: var(--spacing-lg);
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.05) 0%, rgba(139, 92, 246, 0.08) 100%);
  border-radius: var(--radius-xl);
  border: 1px solid rgba(99, 102, 241, 0.1);
}

.benefit {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: var(--color-text-primary);
  font-size: var(--font-size-base);
}

.benefit-icon {
  font-size: 18px;
  color: var(--color-primary);
}

/* Packages Grid */
.packages-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--spacing-lg);
}

.package-card {
  position: relative;
  text-align: center;
  border-radius: var(--radius-xl);
  transition: transform var(--transition-base), box-shadow var(--transition-base);
}

.package-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.package-card.recommended {
  border: 2px solid var(--color-warning);
}

.package-card.popular {
  border: 2px solid var(--color-primary);
}

.package-badge {
  position: absolute;
  top: -1px;
  right: 20px;
  background: var(--color-warning);
  color: #fff;
  padding: 4px 12px;
  border-radius: 0 0 var(--radius-md) var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.package-card.popular .package-badge {
  background: var(--color-primary);
}

.package-header {
  margin-bottom: var(--spacing-base);
}

.package-name {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
}

.package-desc {
  margin: var(--spacing-sm) 0 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

/* Price Section */
.price-section {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: var(--spacing-xs);
}

.currency {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-medium);
  color: var(--color-warning);
}

.price {
  font-size: 42px;
  font-weight: var(--font-weight-bold);
  color: var(--color-warning);
  line-height: 1;
}

.original-price {
  font-size: var(--font-size-base);
  color: var(--color-text-placeholder);
  text-decoration: line-through;
  margin-left: var(--spacing-sm);
}

.period {
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
}

/* Features List */
.features-list {
  list-style: none;
  padding: 0;
  margin: 0;
  text-align: left;
}

.features-list li {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) 0;
  font-size: var(--font-size-base);
  color: var(--color-text-primary);
}

/* Buy Button */
.buy-btn {
  width: 100%;
  margin-top: 16px;
}

/* FAQ Card */
.faq-card :deep(.el-collapse) {
  border: none;
}

.faq-card :deep(.el-collapse-item__header) {
  font-weight: 500;
}

.faq-card :deep(.el-collapse-item__content) {
  color: var(--color-text-regular);
  padding-bottom: var(--spacing-base);
}

/* VIP Status Alert */
.vip-status-alert {
  margin-bottom: 10px;
}

.vip-status-alert :deep(.el-alert__title) {
  display: block;
  font-size: 14px;
}

.days-remaining {
  margin-left: 16px;
  color: var(--el-color-warning);
  font-weight: 600;
}

.vip-info-row {
  margin-top: 4px;
}

.quota-tip {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-left: 4px;
}

/* Info Card */
.info-card {
  margin-top: 10px;
}

.info-content h4 {
  margin: var(--spacing-base) 0 var(--spacing-sm);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
}

.info-content h4:first-child {
  margin-top: 0;
}

.info-content ul {
  margin: 0;
  padding-left: 20px;
  color: var(--el-text-color-secondary);
}

.info-content li {
  margin: var(--spacing-xs) 0;
  font-size: var(--font-size-sm);
}

/* Consumption Flow */
.consumption-flow {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-base);
  background: #f8fafc;
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-base);
}

.flow-step {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  justify-content: center;
}

.step-num {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--el-color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}

.step-text {
  font-size: var(--font-size-base);
  color: var(--color-text-primary);
}

.flow-arrow {
  color: var(--el-color-primary);
  font-size: 20px;
  padding: 8px 0;
}

/* Responsive */
@media (max-width: 768px) {
  .benefits-bar {
    flex-wrap: wrap;
    gap: 16px;
  }
  
  .packages-grid {
    grid-template-columns: 1fr;
  }

  .days-remaining {
    display: block;
    margin-left: 0;
    margin-top: 4px;
  }
}
</style>
