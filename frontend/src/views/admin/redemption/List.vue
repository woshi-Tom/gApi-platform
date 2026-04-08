<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:20px">
      <h2 style="margin:0">兑换码管理</h2>
      <el-button type="primary" @click="showAdd">
        <el-icon><Plus /></el-icon> 生成兑换码
      </el-button>
    </div>
    <el-card>
      <el-form :inline="true" style="margin-bottom:16px">
        <el-form-item label="类型">
          <el-select v-model="filters.code_type" clearable placeholder="全部" style="width:120px" @change="load">
            <el-option v-for="t in CODE_TYPES" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width:100px" @change="load">
            <el-option v-for="s in CODE_STATUS" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="批次">
          <el-input v-model="filters.batch_id" clearable placeholder="批次号" style="width:160px" @change="load" />
        </el-form-item>
        <el-form-item>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="codes" v-loading="ld" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="兑换码" min-width="160">
          <template #default="{ row }">
            <code style="font-size:13px">{{ row.code }}</code>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ getCodeTypeName(row.code_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="奖励内容" min-width="140">
          <template #default="{ row }">
            <span v-if="row.quota > 0">{{ formatQuota(row.quota) }}</span>
            <span v-if="row.vip_days > 0">{{ row.vip_days }}天VIP</span>
            <span v-if="row.is_permanent" class="text-primary">永久VIP</span>
          </template>
        </el-table-column>
        <el-table-column label="使用次数" width="100">
          <template #default="{ row }">{{ row.used_count }} / {{ row.max_uses === 999999 ? '∞' : row.max_uses }}</template>
        </el-table-column>
        <el-table-column label="有效期" width="160">
          <template #default="{ row }">
            <span v-if="row.valid_from || row.valid_until">
              {{ row.valid_from?.substring(0,10) || '-' }} 至 {{ row.valid_until?.substring(0,10) || '-' }}
            </span>
            <span v-else>无限期</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="批次" width="120">
          <template #default="{ row }">
            <span style="font-size:12px;color:#999">{{ row.batch_id || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="viewUsage(row)">使用记录</el-button>
            <el-button size="small" link type="danger" @click="disable(row)" v-if="row.status === 'active'">禁用</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top:16px;display:flex;justify-content:flex-end">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="load"
          @current-change="load"
        />
      </div>
    </el-card>

    <el-dialog v-model="dlgVisible" title="生成兑换码" width="600px">
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="前缀" prop="prefix">
          <el-input v-model="form.prefix" placeholder="如 VIP2024" />
          <div style="color:#909399;font-size:12px;margin-top:4px">生成的兑换码格式: 前缀+时间戳+随机4位</div>
        </el-form-item>
        <el-form-item label="数量" prop="count">
          <el-input-number v-model="form.count" :min="1" :max="1000" style="width:100%" />
        </el-form-item>
        <el-form-item label="类型" prop="code_type">
          <el-select v-model="form.code_type" style="width:100%">
            <el-option v-for="t in CODE_TYPES" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <template v-if="form.code_type === 'recharge' || form.code_type === 'quota'">
          <el-form-item label="配额数量" prop="quota">
            <el-input-number v-model="form.quota" :min="0" :step="10000" style="width:100%" />
          </el-form-item>
          <el-form-item label="配额类型">
            <el-select v-model="form.quota_type" style="width:100%">
              <el-option v-for="q in QUOTA_TYPES" :key="q.value" :label="q.label" :value="q.value" />
            </el-select>
          </el-form-item>
        </template>
        <template v-if="form.code_type === 'vip'">
          <el-form-item label="VIP天数">
            <el-input-number v-model="form.vip_days" :min="1" :max="365" style="width:100%" />
          </el-form-item>
          <el-form-item label="永久VIP">
            <el-switch v-model="form.is_permanent" />
          </el-form-item>
          <el-form-item label="额外配额">
            <el-input-number v-model="form.quota" :min="0" :step="10000" style="width:100%" />
          </el-form-item>
        </template>
        <el-form-item label="使用次数">
          <el-input-number v-model="form.max_uses" :min="1" :max="999999" style="width:100%" />
          <div style="color:#909399;font-size:12px;margin-top:4px">1=一次性, >1=可多次使用</div>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="开始日期">
              <el-date-picker v-model="form.valid_from" type="date" placeholder="可选" style="width:100%" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束日期">
              <el-date-picker v-model="form.valid_until" type="date" placeholder="可选" style="width:100%" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dlgVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">生成</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="usageVisible" title="使用记录" width="700px">
      <el-table :data="usages" v-loading="usageLoading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="user_id" label="用户ID" width="80" />
        <el-table-column label="获得配额" width="100">
          <template #default="{ row }">{{ row.quota_granted > 0 ? formatQuota(row.quota_granted) : '-' }}</template>
        </el-table-column>
        <el-table-column label="VIP" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.vip_granted" type="success" size="small">{{ row.vip_days }}天</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="redeemed_at" label="兑换时间" width="160">
          <template #default="{ row }">{{ row.redeemed_at?.substring(0, 19) }}</template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="130" />
      </el-table>
    </el-dialog>

    <el-dialog v-model="codesVisible" title="生成的兑换码" width="600px">
      <el-alert type="success" :closable="false" style="margin-bottom:16px">
        成功生成 {{ generatedCodes.length }} 个兑换码
      </el-alert>
      <div style="max-height:300px;overflow-y:auto">
        <div v-for="code in generatedCodes" :key="code.code" style="margin-bottom:8px">
          <code style="font-size:14px;background:#f5f7fa;padding:4px 8px;border-radius:4px">{{ code.code }}</code>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="copyAllCodes">复制全部</el-button>
        <el-button @click="codesVisible=false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { redemptionApi, CODE_TYPES, CODE_STATUS, QUOTA_TYPES, formatCode } from '@/api/redemption'
import type { RedemptionCode, RedemptionUsage } from '@/api/redemption'

const codes = ref<RedemptionCode[]>([])
const ld = ref(false)
const dlgVisible = ref(false)
const saving = ref(false)
const usageVisible = ref(false)
const usageLoading = ref(false)
const usages = ref<RedemptionUsage[]>([])
const codesVisible = ref(false)
const generatedCodes = ref<RedemptionCode[]>([])
const formRef = ref<FormInstance>()
const formRef2 = ref<FormInstance>()

const filters = reactive({
  code_type: '',
  status: '',
  batch_id: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  prefix: 'VIP',
  count: 10,
  code_type: 'vip',
  quota: 0,
  quota_type: 'permanent',
  vip_days: 30,
  is_permanent: false,
  max_uses: 1,
  valid_from: '',
  valid_until: '',
})

const rules: FormRules = {
  prefix: [{ required: true, message: '请输入前缀', trigger: 'blur' }],
  count: [{ required: true, message: '请输入数量', trigger: 'blur' }],
  code_type: [{ required: true, message: '请选择类型', trigger: 'change' }],
}

const getCodeTypeName = (type: string) => {
  const t = CODE_TYPES.find(t => t.value === type)
  return t?.label || type
}

const getStatusName = (status: string) => {
  const s = CODE_STATUS.find(s => s.value === status)
  return s?.label || status
}

const getStatusType = (status: string) => {
  const s = CODE_STATUS.find(s => s.value === status)
  return s?.type || 'info'
}

const formatQuota = (quota: number): string => {
  if (quota >= 1000000) {
    return (quota / 1000000).toFixed(1) + 'M'
  }
  if (quota >= 1000) {
    return (quota / 1000).toFixed(0) + 'K'
  }
  return quota.toString()
}

const resetFilters = () => {
  filters.code_type = ''
  filters.status = ''
  filters.batch_id = ''
  load()
}

const showAdd = () => {
  form.prefix = 'VIP'
  form.count = 10
  form.code_type = 'vip'
  form.quota = 0
  form.quota_type = 'permanent'
  form.vip_days = 30
  form.is_permanent = false
  form.max_uses = 1
  form.valid_from = ''
  form.valid_until = ''
  dlgVisible.value = true
}

const save = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const res = await redemptionApi.create({
        prefix: form.prefix,
        count: form.count,
        code_type: form.code_type,
        quota: form.quota ?? undefined,
        quota_type: form.quota_type ?? undefined,
        vip_days: form.vip_days ?? undefined,
        max_uses: form.max_uses,
        valid_from: form.valid_from || undefined,
        valid_until: form.valid_until || undefined,
      })
      generatedCodes.value = res.data.data?.codes || []
      dlgVisible.value = false
      codesVisible.value = true
      load()
    } catch (e: any) {
      ElMessage.error(e.response?.data?.error?.message || '生成失败')
    } finally {
      saving.value = false
    }
  })
}

const copyAllCodes = () => {
  const text = generatedCodes.value.map(c => c.code).join('\n')
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

const viewUsage = async (row: RedemptionCode) => {
  usageLoading.value = true
  usageVisible.value = true
  try {
    const res = await redemptionApi.getUsage(row.id)
    usages.value = res.data.data || []
  } catch (e: any) {
    ElMessage.error('获取使用记录失败')
  } finally {
    usageLoading.value = false
  }
}

const disable = async (row: RedemptionCode) => {
  try {
    await redemptionApi.disable(row.id)
    ElMessage.success('已禁用')
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '操作失败')
  }
}

const load = async () => {
  ld.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    if (filters.code_type) params.code_type = filters.code_type
    if (filters.status) params.status = filters.status
    if (filters.batch_id) params.batch_id = filters.batch_id

    const res = await redemptionApi.list(params)
    if (res.data.data) {
      codes.value = res.data.data.list || res.data.data
      pagination.total = res.data.data.pagination?.total || codes.value.length
    }
  } catch (e: any) {
    ElMessage.error('加载失败')
  } finally {
    ld.value = false
  }
}

onMounted(load)
</script>
<style scoped>
.text-primary {
  color: #409eff;
}
</style>
