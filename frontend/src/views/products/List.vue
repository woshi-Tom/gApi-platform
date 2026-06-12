<template>
  <div>
    <PageHeader title="商品列表" description="选择适合您的充值套餐或 VIP 会员" />
    <el-tabs v-model="tab" class="products-tabs">
      <el-tab-pane label="充值套餐" name="recharge">
        <div class="products-grid">
          <el-card v-for="p in rc" :key="p.id" shadow="hover" class="product-card" :class="{ recommended: p.is_recommended, hot: p.is_hot }">
            <div class="product-badge" v-if="p.is_recommended || p.is_hot">
              {{ p.is_recommended ? '推荐' : '热门' }}
            </div>
            <div class="product-header">
              <h3 class="product-name">{{ p.name }}</h3>
            </div>
            <div class="product-price">
              <span class="currency">¥</span>
              <span class="amount">{{ p.price }}</span>
              <span v-if="p.original_price" class="original">¥{{ p.original_price }}</span>
            </div>
            <div class="product-features">
              <div class="feature-item">
                <span class="feature-label">配额</span>
                <span class="feature-value primary">{{ formatQuota(p.quota) }}</span>
              </div>
              <div class="feature-item" v-if="p.bonus_quota">
                <span class="feature-label">赠送</span>
                <span class="feature-value success">+{{ formatQuota(p.bonus_quota) }}</span>
              </div>
              <div class="feature-item">
                <span class="feature-label">RPM</span>
                <span class="feature-value">{{ (p.rpm_limit ?? 0) > 0 ? p.rpm_limit : '无限制' }}</span>
              </div>
              <div class="feature-item">
                <span class="feature-label">TPM</span>
                <span class="feature-value">{{ (p.tpm_limit ?? 0) > 0 ? formatQuota(p.tpm_limit) : '无限制' }}</span>
              </div>
            </div>
            <el-button type="primary" size="large" class="buy-btn" @click="buy(p,'recharge')">
              立即购买
            </el-button>
          </el-card>
        </div>
      </el-tab-pane>
      <el-tab-pane label="VIP套餐" name="vip">
        <div class="products-grid">
          <el-card v-for="p in vp" :key="p.id" shadow="hover" class="product-card" :class="{ recommended: p.is_recommended, hot: p.is_hot }">
            <div class="product-badge" v-if="p.is_recommended || p.is_hot">
              {{ p.is_recommended ? '推荐' : '热门' }}
            </div>
            <div class="product-header">
              <h3 class="product-name">{{ p.name }}</h3>
            </div>
            <div class="product-price">
              <span class="currency">¥</span>
              <span class="amount">{{ p.price }}</span>
              <span v-if="p.original_price" class="original">¥{{ p.original_price }}</span>
            </div>
            <div class="product-features">
              <div class="feature-item">
                <span class="feature-label">有效期</span>
                <span class="feature-value">{{ p.vip_days || p.duration_days || 30 }} 天</span>
              </div>
              <div class="feature-item">
                <span class="feature-label">配额</span>
                <span class="feature-value primary">{{ formatQuota((p.vip_quota ?? p.quota) || 0) }}</span>
              </div>
              <div class="feature-item">
                <span class="feature-label">RPM</span>
                <span class="feature-value">{{ (p.rpm_limit ?? 0) > 0 ? p.rpm_limit : '无限制' }}</span>
              </div>
              <div class="feature-item">
                <span class="feature-label">TPM</span>
                <span class="feature-value">{{ (p.tpm_limit ?? 0) > 0 ? formatQuota(p.tpm_limit) : '无限制' }}</span>
              </div>
            </div>
            <el-button type="warning" size="large" class="buy-btn" @click="buy(p,'vip')">
              开通VIP
            </el-button>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>
    <el-empty v-if="!rc.length && !vp.length && !loading" description="暂无商品" />
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProducts } from '@/api/product'
import PageHeader from '@/components/PageHeader.vue'
import request from '@/api/request'
import { ElMessage } from 'element-plus'
import { formatQuota } from '@/composables/useFormat'

const router = useRouter()
const tab = ref('recharge')
const loading = ref(false)
const rc = ref<any[]>([])
const vp = ref<any[]>([])

async function load() {
  loading.value = true
  try {
    const { data } = await getProducts()
    const a = data.data || []
    rc.value = a.filter((p: any) => p.product_type === 'recharge')
    vp.value = a.filter((p: any) => p.product_type === 'vip')
  } catch {
    ElMessage.error('加载商品失败')
  } finally {
    loading.value = false
  }
}

async function buy(p: any, t: string) {
  try {
    const res = await request.post('/orders', { package_id: p.id, package_type: t, payment_method: 'alipay' })
    const data = res.data?.data || res.data
    ElMessage.success('订单已创建，正在跳转支付页面...')
    router.push(`/payment?order_no=${data.order_no}`)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '创建失败')
  }
}

onMounted(load)
</script>
<style scoped>
.products-tabs {
  margin-top: 0;
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-lg);
}

.product-card {
  position: relative;
  border-radius: var(--radius-2xl);
  text-align: center;
  transition: transform var(--transition-base), box-shadow var(--transition-base);
  overflow: visible;
}

.product-card:hover {
  transform: translateY(-6px);
  box-shadow: var(--shadow-xl);
}

.product-card.recommended {
  border: 2px solid var(--color-warning);
}

.product-card.hot {
  border: 2px solid var(--color-danger);
}

.product-badge {
  position: absolute;
  top: -10px;
  right: 16px;
  background: var(--gradient-orange);
  color: #fff;
  padding: 4px 14px;
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.4);
  z-index: 1;
}

.product-card.hot .product-badge {
  background: var(--gradient-red);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.4);
}

.product-header {
  margin-bottom: var(--spacing-base);
}

.product-name {
  margin: 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.product-price {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 2px;
  margin-bottom: var(--spacing-lg);
}

.product-price .currency {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-medium);
  color: var(--color-primary);
}

.product-price .amount {
  font-size: 40px;
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  line-height: 1;
}

.product-price .original {
  font-size: var(--font-size-base);
  color: var(--color-text-placeholder);
  text-decoration: line-through;
  margin-left: var(--spacing-sm);
}

.product-features {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  padding: var(--spacing-base);
  background: #f8fafc;
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-lg);
}

.feature-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--font-size-base);
}

.feature-label {
  color: var(--color-text-secondary);
}

.feature-value {
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.feature-value.primary {
  color: var(--color-primary);
}

.feature-value.success {
  color: var(--color-success);
  font-weight: var(--font-weight-bold);
}

.buy-btn {
  width: 100%;
  height: 44px;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  border-radius: var(--radius-lg);
}

@media (max-width: 768px) {
  .products-grid {
    grid-template-columns: 1fr;
  }
}
</style>
