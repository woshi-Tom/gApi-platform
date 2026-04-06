<template>
  <div style="max-width:600px;margin:0 auto;padding:20px">
    <el-card>
      <template #header>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:20px">🎁</span>
          <span style="font-size:18px;font-weight:bold">兑换码兑换</span>
        </div>
      </template>
      
      <el-form :model="form" label-width="100px">
        <el-form-item label="兑换码">
          <el-input v-model="form.code" placeholder="请输入兑换码" size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" :loading="loading" @click="redeem" style="width:100%">
            立即兑换
          </el-button>
        </el-form-item>
      </el-form>

      <el-divider />

      <div style="color:#909399;font-size:13px">
        <p style="margin-bottom:8px">兑换说明：</p>
        <ul style="padding-left:20px;margin:0">
          <li>每个兑换码仅限使用一次</li>
          <li>兑换码有有效期限，请在有效期内使用</li>
          <li>兑换码一经使用不可退还</li>
        </ul>
      </div>
    </el-card>

    <el-card v-if="result" style="margin-top:20px">
      <template #header>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:20px">{{ result.success ? '✅' : '❌' }}</span>
          <span style="font-size:16px;font-weight:bold">{{ result.success ? '兑换成功' : '兑换失败' }}</span>
        </div>
      </template>
      
      <div v-if="result.success">
        <p style="margin-bottom:8px">恭喜！您已成功兑换以下权益：</p>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="配额奖励" v-if="result.quota_granted > 0">
            {{ formatQuota(result.quota_granted) }}
          </el-descriptions-item>
          <el-descriptions-item label="VIP奖励" v-if="result.vip_granted">
            {{ result.vip_days }}天VIP会员
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <div v-else style="color:#f56c6c">
        {{ result.error }}
      </div>
    </el-card>

    <el-card style="margin-top:20px">
      <template #header>
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:18px">📜</span>
          <span style="font-size:16px;font-weight:bold">兑换历史</span>
        </div>
      </template>
      
      <el-table :data="history" v-loading="historyLoading" stripe>
        <el-table-column prop="redeemed_at" label="兑换时间" width="160">
          <template #default="{ row }">{{ row.redeemed_at?.substring(0, 19) }}</template>
        </el-table-column>
        <el-table-column label="获得权益" min-width="120">
          <template #default="{ row }">
            <span v-if="row.quota_granted > 0">配额: {{ formatQuota(row.quota_granted) }}</span>
            <el-tag v-if="row.vip_granted" type="success" size="small" style="margin-left:4px">
              {{ row.vip_days }}天VIP
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="130" />
      </el-table>
      
      <el-empty v-if="!historyLoading && history.length === 0" description="暂无兑换记录" style="padding:40px 0" />
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { userRedemptionApi } from '@/api/redemption'

const form = reactive({
  code: '',
})

const loading = ref(false)
const historyLoading = ref(false)
const result = ref<{success: boolean; quota_granted?: number; vip_granted?: boolean; vip_days?: number; error?: string} | null>(null)
const history = ref<any[]>([])

const formatQuota = (quota: number): string => {
  if (quota >= 1000000) {
    return (quota / 1000000).toFixed(1) + 'M'
  }
  if (quota >= 1000) {
    return (quota / 1000).toFixed(0) + 'K'
  }
  return quota.toString()
}

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
  } catch (e: any) {
    // ignore
  } finally {
    historyLoading.value = false
  }
}

onMounted(loadHistory)
</script>
