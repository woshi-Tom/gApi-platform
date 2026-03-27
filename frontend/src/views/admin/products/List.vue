<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:20px">
      <h2 style="margin:0">商品管理</h2>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon> 添加商品
      </el-button>
    </div>
    <el-card>
      <el-form :inline="true" style="margin-bottom:16px">
        <el-form-item label="商品类型">
          <el-select v-model="filters.type" clearable placeholder="全部" style="width:140px" @change="load">
            <el-option label="全部" value="" />
            <el-option label="VIP套餐" value="vip" />
            <el-option label="充值套餐" value="recharge" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width:120px" @change="load">
            <el-option label="全部" value="" />
            <el-option label="上架" value="active" />
            <el-option label="下架" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <el-table :data="products" v-loading="ld" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.product_type==='vip'?'warning':'success'" size="small">
              {{ row.product_type==='vip'?'VIP套餐':'充值' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="100">
          <template #default="{ row }">
            ¥{{ row.price.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="额度/天数" width="120">
          <template #default="{ row }">
            <span v-if="row.product_type==='vip'">{{ row.vip_days || 0 }}天</span>
            <span v-else>{{ (row.quota || 0).toLocaleString() }} token</span>
          </template>
        </el-table-column>
        <el-table-column label="RPM限制" width="90">
          <template #default="{ row }">
            <span v-if="row.product_type==='vip'">{{ row.rpm_limit || '-' }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="TPM限制" width="90">
          <template #default="{ row }">
            <span v-if="row.product_type==='vip'">{{ row.tpm_limit ? (row.tpm_limit / 1000) + 'k' : '-' }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status==='active'?'success':'danger'" size="small">
              {{ row.status==='active'?'上架':'下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="60" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button v-if="row.status==='active'" type="danger" size="small" @click="handleDisable(row)">
              下架
            </el-button>
            <el-button v-else type="success" size="small" @click="handleEnable(row)">
              上架
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑商品' : '添加商品'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="商品类型">
          <el-radio-group v-model="form.product_type">
            <el-radio label="vip">VIP套餐</el-radio>
            <el-radio label="recharge">充值套餐</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="请输入商品名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" rows="3" placeholder="请输入商品描述" />
        </el-form-item>
        <el-form-item label="价格" required>
          <el-input-number v-model="form.price" :min="0" :precision="2" style="width:200px" />
        </el-form-item>
        <template v-if="form.product_type === 'vip'">
          <el-form-item label="VIP天数">
            <el-input-number v-model="form.vip_days" :min="1" style="width:200px" />
          </el-form-item>
          <el-form-item label="配额">
            <el-input-number v-model="form.quota" :min="0" style="width:200px" />
          </el-form-item>
          <el-form-item label="RPM限制">
            <el-input-number v-model="form.rpm_limit" :min="0" style="width:200px" />
          </el-form-item>
          <el-form-item label="TPM限制">
            <el-input-number v-model="form.tpm_limit" :min="0" style="width:200px" />
          </el-form-item>
          <el-form-item label="并发限制">
            <el-input-number v-model="form.concurrent_limit" :min="1" style="width:200px" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="配额" required>
            <el-input-number v-model="form.quota" :min="0" style="width:200px" />
          </el-form-item>
          <el-form-item label="赠送配额">
            <el-input-number v-model="form.bonus_quota" :min="0" style="width:200px" />
          </el-form-item>
        </template>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" style="width:200px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio label="active">上架</el-radio>
            <el-radio label="inactive">下架</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { adminAPI } from '@/api/request'

const ld = ref(false)
const products = ref([])
const dialogVisible = ref(false)
const saving = ref(false)
const isEdit = ref(false)
const filters = reactive({
  type: '',
  status: ''
})

const defaultForm = () => ({
  id: null,
  product_type: 'vip',
  name: '',
  description: '',
  price: 0,
  vip_days: 30,
  quota: 1000000,
  bonus_quota: 0,
  rpm_limit: 2000,
  tpm_limit: 100000,
  concurrent_limit: 10,
  sort_order: 0,
  status: 'active'
})

const form = reactive(defaultForm())

async function load() {
  ld.value = true
  try {
    const params = {}
    if (filters.type) params.type = filters.type
    const res = await adminAPI.get('/products', { params })
    let list = res.data.data || []
    if (filters.status) {
      list = list.filter(p => p.status === filters.status)
    }
    products.value = list
  } catch (e) {
    ElMessage.error('加载失败')
  } finally {
    ld.value = false
  }
}

function handleAdd() {
  isEdit.value = false
  Object.assign(form, defaultForm())
  dialogVisible.value = true
}

function handleEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    product_type: row.product_type,
    name: row.name,
    description: row.description || '',
    price: row.price,
    vip_days: row.vip_days || 30,
    quota: row.quota || 0,
    bonus_quota: row.bonus_quota || 0,
    rpm_limit: row.rpm_limit || 2000,
    tpm_limit: row.tpm_limit || 100000,
    concurrent_limit: row.concurrent_limit || 10,
    sort_order: row.sort_order || 0,
    status: row.status
  })
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.name) {
    ElMessage.warning('请输入商品名称')
    return
  }
  if (form.price <= 0) {
    ElMessage.warning('请输入有效的价格')
    return
  }

  saving.value = true
  try {
    if (form.product_type === 'vip') {
      var payload = {
        product_type: 'vip',
        name: form.name,
        description: form.description,
        price: form.price,
        vip_days: form.vip_days,
        quota: form.quota,
        rpm_limit: form.rpm_limit,
        tpm_limit: form.tpm_limit,
        concurrent_limit: form.concurrent_limit,
        sort_order: form.sort_order,
        status: form.status
      }
    } else {
      var payload = {
        product_type: 'recharge',
        name: form.name,
        description: form.description,
        price: form.price,
        quota: form.quota,
        bonus_quota: form.bonus_quota,
        sort_order: form.sort_order,
        status: form.status
      }
    }

    if (isEdit.value) {
      await adminAPI.put(`/products/${form.id}?type=${form.product_type}`, payload)
      ElMessage.success('编辑成功')
    } else {
      await adminAPI.post('/products', payload)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e.response?.data?.error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleEnable(row) {
  try {
    await ElMessageBox.confirm(`确认上架「${row.name}」?`, '提示')
    const type = row.product_type || (row.vip_days ? 'vip' : 'recharge')
    await adminAPI.post(`/products/${row.id}/enable?type=${type}`)
    ElMessage.success('上架成功')
    load()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

async function handleDisable(row) {
  try {
    await ElMessageBox.confirm(`确认下架「${row.name}」?`, '提示')
    const type = row.product_type || (row.vip_days ? 'vip' : 'recharge')
    await adminAPI.post(`/products/${row.id}/disable?type=${type}`)
    ElMessage.success('下架成功')
    load()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

onMounted(() => {
  load()
})
</script>