<template>
  <div class="vip-page">
    <div class="page-header">
      <h2>VIP 会员</h2>
      <p class="subtitle">开通 VIP 享受更多特权，提升 API 调用体验</p>
    </div>

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

    <div class="packages-grid" v-if="pkgs.length">
      <el-card 
        v-for="pkg in pkgs" 
        :key="pkg.id" 
        class="package-card"
        :class="{ recommended: pkg.is_recommended, popular: pkg.is_popular }"
        shadow="hover"
      >
        <div class="package-badge" v-if="pkg.is_recommended || pkg.is_popular">
          {{ pkg.is_recommended ? '推荐' : '热销' }}
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
            <span>有效期 {{ pkg.duration_days }} 天</span>
          </li>
          <li>
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>配额 {{ formatQuota(pkg.vip_quota) }} Token</span>
          </li>
          <li v-if="pkg.rpm_limit">
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>RPM 限制 {{ pkg.rpm_limit }}</span>
          </li>
          <li v-if="pkg.tpm_limit">
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>TPM 限制 {{ formatQuota(pkg.tpm_limit) }}</span>
          </li>
          <li v-if="pkg.concurrent_limit">
            <el-icon color="#67c23a"><Check /></el-icon>
            <span>并发数 {{ pkg.concurrent_limit }}</span>
          </li>
        </ul>
        
        <el-button 
          :type="pkg.is_recommended ? 'warning' : 'primary'" 
          size="large" 
          class="buy-btn"
          @click="buy(pkg)"
        >
          立即开通
        </el-button>
      </el-card>
    </div>

    <el-empty v-else description="暂无 VIP 套餐" />

    <!-- FAQ Section -->
    <el-card class="faq-card" v-if="pkgs.length">
      <template #header>
        <span>常见问题</span>
      </template>
      <el-collapse>
        <el-collapse-item title="VIP 配额和永久配额有什么区别？" name="1">
          <p>VIP 配额在会员有效期内使用，过期后清零。永久配额不会过期，可长期使用。</p>
        </el-collapse-item>
        <el-collapse-item title="VIP 过期后会怎样？" name="2">
          <p>VIP 过期后，您将回到普通用户等级。VIP 配额将清零，但永久配额不受影响。</p>
        </el-collapse-item>
        <el-collapse-item title="可以升级或降级 VIP 吗？" name="3">
          <p>可以随时购买新的 VIP 套餐。新套餐将延长或替换现有 VIP 有效期。</p>
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
import { getProducts } from '@/api/product'
import request from '@/api/request'

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
  is_popular: boolean
}

const pkgs = ref<VIPPackage[]>([])
const router = useRouter()

function formatQuota(n: number): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
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

onMounted(load)
</script>

<style scoped>
.vip-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 10px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.subtitle {
  color: var(--el-text-color-secondary);
  margin: 8px 0 0;
}

/* Benefits Bar */
.benefits-bar {
  display: flex;
  justify-content: center;
  gap: 40px;
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #ecf5ff 100%);
  border-radius: 12px;
}

.benefit {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-primary);
  font-size: 14px;
}

.benefit-icon {
  font-size: 18px;
  color: var(--el-color-primary);
}

/* Packages Grid */
.packages-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
}

.package-card {
  position: relative;
  text-align: center;
  border-radius: 12px;
  transition: transform 0.2s, box-shadow 0.2s;
}

.package-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
}

.package-card.recommended {
  border: 2px solid var(--el-color-warning);
}

.package-card.popular {
  border: 2px solid var(--el-color-primary);
}

.package-badge {
  position: absolute;
  top: -1px;
  right: 20px;
  background: var(--el-color-warning);
  color: #fff;
  padding: 4px 12px;
  border-radius: 0 0 8px 8px;
  font-size: 12px;
  font-weight: 500;
}

.package-card.popular .package-badge {
  background: var(--el-color-primary);
}

.package-header {
  margin-bottom: 16px;
}

.package-name {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.package-desc {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

/* Price Section */
.price-section {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
}

.currency {
  font-size: 20px;
  font-weight: 500;
  color: var(--el-color-warning);
}

.price {
  font-size: 42px;
  font-weight: 700;
  color: var(--el-color-warning);
  line-height: 1;
}

.original-price {
  font-size: 14px;
  color: var(--el-text-color-placeholder);
  text-decoration: line-through;
  margin-left: 8px;
}

.period {
  font-size: 14px;
  color: var(--el-text-color-secondary);
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
  gap: 8px;
  padding: 8px 0;
  font-size: 14px;
  color: var(--el-text-color-primary);
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
  color: var(--el-text-color-secondary);
  padding-bottom: 16px;
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
}
</style>
