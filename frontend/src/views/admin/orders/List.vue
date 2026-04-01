<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:20px">
      <h2 style="margin:0">订单管理</h2>
    </div>
    <el-card>
      <el-form :inline="true" style="margin-bottom:16px">
        <el-form-item label="订单状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width:120px" @change="load">
            <el-option label="待支付" value="pending" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已退款" value="refunded" />
          </el-select>
        </el-form-item>
        <el-form-item label="订单类型">
          <el-select v-model="filters.order_type" clearable placeholder="全部" style="width:120px" @change="load">
            <el-option label="VIP套餐" value="vip" />
            <el-option label="充值" value="recharge" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input v-model="filters.user_id" clearable placeholder="用户ID" style="width:100px" @change="load" />
        </el-form-item>
      </el-form>
      <el-table :data="orders" v-loading="ld" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="order_no" label="订单号" min-width="180" />
        <el-table-column prop="user_id" label="用户ID" width="80" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="row.order_type==='vip'?'warning':'success'" size="small">
              {{ orderTypeName(row.order_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="package_name" label="商品" min-width="120" />
        <el-table-column label="金额" width="100">
          <template #default="{ row }">¥{{ row.pay_amount?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="showDetail(row)">详情</el-button>
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

    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <el-descriptions :column="2" border v-if="currentOrder">
        <el-descriptions-item label="订单ID">{{ currentOrder.id }}</el-descriptions-item>
        <el-descriptions-item label="订单号">{{ currentOrder.order_no }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ currentOrder.user_id }}</el-descriptions-item>
        <el-descriptions-item label="订单类型">
          <el-tag size="small">{{ orderTypeName(currentOrder.order_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="商品名称">{{ currentOrder.package_name }}</el-descriptions-item>
        <el-descriptions-item label="商品ID">{{ currentOrder.package_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="原价">¥{{ currentOrder.total_amount?.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="折扣">¥{{ currentOrder.discount_amount?.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实付金额">
          <b style="color:#409eff">¥{{ currentOrder.pay_amount?.toFixed(2) }}</b>
        </el-descriptions-item>
        <el-descriptions-item label="订单状态">
          <el-tag :type="statusType(currentOrder.status)" size="small">{{ statusName(currentOrder.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ currentOrder.paid_at ? formatTime(currentOrder.paid_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="过期时间">{{ currentOrder.expire_at ? formatTime(currentOrder.expire_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(currentOrder.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="取消原因" v-if="currentOrder.cancel_reason">{{ currentOrder.cancel_reason }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible=false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminOrderApi } from '@/api/order'
import type { Order } from '@/api/order'

const orders = ref<Order[]>([])
const ld = ref(false)
const detailVisible = ref(false)
const currentOrder = ref<Order | null>(null)

const filters = reactive({
  status: '',
  order_type: '',
  user_id: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const statusType = (s: string) => {
  switch (s) {
    case 'completed': return 'success'
    case 'paid': return 'success'
    case 'pending': return 'warning'
    case 'cancelled': return 'info'
    case 'refunded': return 'danger'
    case 'expired': return 'warning'
    default: return 'info'
  }
}

const statusName = (s: string) => {
  const map: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    completed: '已完成',
    cancelled: '已取消',
    refunded: '已退款',
    expired: '已过期',
  }
  return map[s] || s
}

const orderTypeName = (t: string) => {
  return t === 'vip' ? 'VIP套餐' : t === 'recharge' ? '充值' : t
}

const formatTime = (t: string) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

const showDetail = (order: Order) => {
  currentOrder.value = order
  detailVisible.value = true
}

const load = async () => {
  ld.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    if (filters.status) params.status = filters.status
    if (filters.order_type) params.order_type = filters.order_type
    if (filters.user_id) params.user_id = filters.user_id
    
    const res = await adminOrderApi.list(params)
    if (res.data.data) {
      orders.value = res.data.data.list || res.data.data
      pagination.total = res.data.data.pagination?.total || orders.value.length
    }
  } finally {
    ld.value = false
  }
}

onMounted(load)
</script>
