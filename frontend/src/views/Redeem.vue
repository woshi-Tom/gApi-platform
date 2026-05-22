<template>
  <div class="redeem-page">
    <PageHeader title="兑换码" description="使用兑换码获取配额或 VIP 会员权益" />

    <el-card class="redeem-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon class="card-header-icon" :size="20"><Present /></el-icon>
          <span>兑换码兑换</span>
        </div>
      </template>

      <el-form :model="form" label-width="100px" class="redeem-form">
        <el-form-item label="兑换码">
          <el-input
            v-model="form.code"
            placeholder="请输入兑换码"
            size="large"
            clearable
            :prefix-icon="Tickets"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            :disabled="!form.code.trim()"
            @click="redeem"
            class="redeem-btn"
          >
            <el-icon><Present /></el-icon>
            立即兑换
          </el-button>
        </el-form-item>
      </el-form>

      <el-divider />

      <div class="redeem-tips">
        <h4 class="tips-title">兑换说明</h4>
        <ul class="tips-list">
          <li>每个兑换码仅限使用一次</li>
          <li>兑换码有有效期限，请在有效期内使用</li>
          <li>兑换码一经使用不可退还</li>
        </ul>
      </div>
    </el-card>

    <!-- 兑换结果 -->
    <el-card v-if="result" class="result-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon v-if="result.success" class="success-icon"><CircleCheck /></el-icon>
          <el-icon v-else class="error-icon"><CircleClose /></el-icon>
          <span>{{ result.success ? '兑换成功' : '兑换失败' }}</span>
        </div>
      </template>

      <div v-if="result.success">
        <el-alert type="success" :closable="false" show-icon class="result-alert">
          <template #title>恭喜！您已成功兑换以下权益</template>
        </el-alert>
        <el-descriptions :column="1" border class="result-desc">
          <el-descriptions-item label="配额奖励" v-if="result.quota_granted && result.quota_granted > 0">
            <span class="quota-highlight">{{ formatQuota(result.quota_granted) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="VIP奖励" v-if="result.vip_granted">
            <el-tag type="warning" effect="plain">VIP {{ result.vip_days }} 天</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <div v-else>
        <el-alert type="error" :closable="false" show-icon>
          <template #title>{{ result.error }}</template>
        </el-alert>
      </div>
    </el-card>

    <!-- 兑换历史 -->
    <el-card class="history-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <el-icon class="card-header-icon" :size="20"><Clock /></el-icon>
          <span>兑换历史</span>
        </div>
      </template>

      <el-table :data="history" v-loading="historyLoading" stripe class="history-table">
        <el-table-column label="兑换时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.redeemed_at) }}
          </template>
        </el-table-column>
        <el-table-column label="获得权益" min-width="160">
          <template #default="{ row }">
            <div class="reward-cell">
              <span v-if="row.quota_granted && row.quota_granted > 0" class="reward-quota">
                配额: {{ formatQuota(row.quota_granted) }}
              </span>
              <el-tag v-if="row.vip_granted" type="warning" size="small" effect="plain">
                VIP {{ row.vip_days }} 天
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="150" />
      </el-table>

      <div v-if="!historyLoading && history.length === 0" class="empty-state">
        <el-empty description="暂无兑换记录" :image-size="80" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Present, Tickets, CircleCheck, CircleClose, Clock } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { userRedemptionApi } from '@/api/redemption'
import { formatQuota, formatDate } from '@/composables/useFormat'

interface RedeemResult {
  success: boolean
  quota_granted?: number
  vip_granted?: boolean
  vip_days?: number
  error?: string
}

const form = reactive({ code: '' })
const loading = ref(false)
const historyLoading = ref(false)
const result = ref<RedeemResult | null>(null)
const history = ref<any[]>([])

const redeem = async () => {
  if (!form.code.trim()) {
    ElMessage.warning('请输入兑换码')
    return
  }

  loading.value = true
  result.value = null
  try {
    const res = await userRedemptionApi.redeem(form.code.trim())
    result.value = {
      success: true,
      quota_granted: res.data.data?.quota_granted || 0,
      vip_granted: res.data.data?.vip_granted || false,
      vip_days: res.data.data?.vip_days || 0,
    }
    form.code = ''
    loadHistory()
    ElMessage.success('兑换成功')
  } catch (e: any) {
    result.value = {
      success: false,
      error: e.response?.data?.error?.message || e.response?.data?.error || '兑换失败',
    }
  } finally {
    loading.value = false
  }
}

const loadHistory = async () => {
  historyLoading.value = true
  try {
    const res = await userRedemptionApi.getHistory()
    history.value = res.data.data || []
  } catch {
    // ignore
  } finally {
    historyLoading.value = false
  }
}

onMounted(loadHistory)
</script>

<style scoped>
.redeem-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

/* Card Header */
.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-weight: var(--font-weight-semibold);
}

.card-header-icon {
  color: var(--color-primary);
}

/* Redeem Card */
.redeem-card {
  border-radius: var(--radius-xl);
}

.redeem-form {
  max-width: 500px;
}

.redeem-btn {
  width: 200px;
}

/* Tips */
.redeem-tips {
  padding: var(--spacing-xs) 0;
}

.tips-title {
  margin: 0 0 var(--spacing-sm);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.tips-list {
  margin: 0;
  padding-left: 20px;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.8;
}

/* Result Card */
.result-card {
  border-radius: var(--radius-xl);
}

.success-icon {
  color: var(--color-success);
  font-size: 20px;
}

.error-icon {
  color: var(--color-danger);
  font-size: 20px;
}

.result-alert {
  margin-bottom: var(--spacing-base);
}

.result-desc {
  margin-top: var(--spacing-md);
}

.quota-highlight {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
}

/* History Card */
.history-card {
  border-radius: var(--radius-xl);
}

.history-table {
  border-radius: var(--radius-md);
}

.reward-cell {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.reward-quota {
  font-weight: var(--font-weight-semibold);
  color: var(--color-primary);
}

.empty-state {
  padding: 40px 0;
}

/* Responsive */
@media (max-width: 640px) {
  .redeem-btn {
    width: 100%;
  }
}
</style>