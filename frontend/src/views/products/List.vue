<template>
  <div>
    <h2>商品列表</h2>
    <el-tabs v-model="tab">
      <el-tab-pane label="充值套餐" name="recharge">
        <el-row :gutter="20"><el-col :span="8" v-for="p in rc" :key="p.id">
          <el-card shadow="hover" class="pc"><template #header><div style="display:flex;justify-content:space-between"><span>{{ p.name }}</span><el-tag v-if="p.is_popular" type="danger" size="small">热销</el-tag></div></template>
            <div class="price">¥{{ p.price }}<span v-if="p.original_price" class="org">¥{{ p.original_price }}</span></div>
            <div class="info">配额：{{ (p.quota ?? 0).toLocaleString() }} Token</div>
            <div class="info" v-if="p.bonus_quota">赠送：{{ (p.bonus_quota ?? 0).toLocaleString() }} Token</div>
            <div class="info-limit">
              <span class="limit-item">RPM: {{ (p.rpm_limit ?? 0) > 0 ? p.rpm_limit : '0' }}</span>
              <span class="limit-item">TPM: {{ (p.tpm_limit ?? 0) > 0 ? (p.tpm_limit/1000)+'k' : '0' }}</span>
            </div>
            <el-button type="primary" style="width:100%;margin-top:16px" @click="buy(p,'recharge')">立即购买</el-button>
          </el-card>
        </el-col></el-row>
      </el-tab-pane>
      <el-tab-pane label="VIP套餐" name="vip">
        <el-row :gutter="20"><el-col :span="8" v-for="p in vp" :key="p.id">
          <el-card shadow="hover" class="pc" :class="{rec:p.is_recommended}"><template #header><div style="display:flex;justify-content:space-between"><span>{{ p.name }}</span><el-tag v-if="p.is_recommended" type="warning" size="small">推荐</el-tag></div></template>
            <div class="price">¥{{ p.price }}<span v-if="p.original_price" class="org">¥{{ p.original_price }}</span></div>
            <div class="info">有效期：{{ p.vip_days || p.duration_days || 30 }} 天</div>
            <div class="info">配额：{{ ((p.vip_quota ?? p.quota) || 0).toLocaleString() }} Token</div>
            <div class="info-limit">
              <span class="limit-item">RPM: {{ (p.rpm_limit ?? 0) > 0 ? p.rpm_limit : '0' }}</span>
              <span class="limit-item">TPM: {{ (p.tpm_limit ?? 0) > 0 ? (p.tpm_limit/1000)+'k' : '0' }}</span>
            </div>
            <el-button type="warning" style="width:100%;margin-top:16px" @click="buy(p,'vip')">开通VIP</el-button>
          </el-card>
        </el-col></el-row>
      </el-tab-pane>
    </el-tabs>
    <el-empty v-if="!rc.length && !vp.length" description="暂无商品" />
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProducts } from '@/api/product'
import request from '@/api/request'
import { ElMessage } from 'element-plus'
const tab = ref('recharge')
const rc = ref<any[]>([])
const vp = ref<any[]>([])
async function load() { const {data}=await getProducts(); const a=data.data||[]; rc.value=a.filter((p:any)=>p.product_type==='recharge'); vp.value=a.filter((p:any)=>p.product_type==='vip') }
async function buy(p:any,t:string) { try { await request.post('/orders',{package_id:p.id,package_type:t,payment_method:'alipay'}); ElMessage.success('订单已创建') } catch(e:any) { ElMessage.error(e.response?.data?.error?.message||'创建失败') } }
onMounted(load)
</script>
<style scoped>
.pc { text-align:center;margin-bottom:20px }.pc.rec { border:2px solid #e6a23c }
.price { font-size:28px;font-weight:bold;color:#f56c6c;margin:16px 0 }.org { font-size:14px;color:#999;text-decoration:line-through;margin-left:8px }
.info { color:#606266;margin:6px 0;font-size:14px }
.info-limit { 
  display:flex; 
  justify-content:center; 
  gap:16px; 
  margin-top:8px; 
  padding:8px; 
  background:#f5f7fa; 
  border-radius:4px;
  font-size:13px;
  color:#606266;
}
.limit-item { font-weight:500 }
</style>
