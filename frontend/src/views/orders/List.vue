<template>
  <div class="orders-page">
    <PageHeader title="订单记录" description="查看您的购买和支付记录" />

    <el-card class="orders-card">
      <el-table v-if="orders.length > 0 || ld" :data="orders" v-loading="ld" stripe>
        <el-table-column prop="order_no" label="订单号" min-width="200" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.order_type === 'vip' ? 'warning' : 'success'" size="small">
              {{ row.order_type === 'vip' ? 'VIP' : row.order_type === 'recharge' ? '充值' : '套餐' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="package_name" label="商品" min-width="150" />
        <el-table-column label="金额" width="100">
          <template #default="{ row }">
            <span class="amount">¥{{ row.pay_amount.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getOrderStatusType(row.status)" size="small">
              {{ getOrderStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="支付时间" width="170">
          <template #default="{ row }">
            {{ row.paid_at ? formatDate(row.paid_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === 'pending'" 
              type="primary" 
              size="small"
              @click="handlePay(row)"
            >
              去支付
            </el-button>
            <el-button
              v-else
              size="small"
              text
              @click="showDetail(row)"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="!ld && orders.length === 0" class="empty-state">
        <el-empty description="暂无订单记录">
          <el-button type="primary" @click="$router.push('/products')">
            <el-icon><ShoppingCart /></el-icon> 去看看商品
          </el-button>
        </el-empty>
      </div>

      <div class="pagination" v-if="orders.length > 0">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="订单详情" width="500px">
      <el-descriptions :column="2" border v-if="currentOrder">
        <el-descriptions-item label="订单号">{{ currentOrder.order_no }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="currentOrder.order_type === 'vip' ? 'warning' : 'success'" size="small">
            {{ currentOrder.order_type }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="商品">{{ currentOrder.package_name }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getOrderStatusType(currentOrder.status)" size="small">
            {{ getOrderStatusName(currentOrder.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="原价">¥{{ currentOrder.total_amount.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="优惠">-¥{{ currentOrder.discount_amount.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实付" :span="2">
          <span class="amount-large">¥{{ currentOrder.pay_amount.toFixed(2) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentOrder.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">
          {{ currentOrder.paid_at ? formatDate(currentOrder.paid_at) : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="过期时间" v-if="currentOrder.expire_at">
          {{ formatDate(currentOrder.expire_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="取消原因" :span="2" v-if="currentOrder.cancel_reason">
          {{ currentOrder.cancel_reason }}
        </el-descriptions-item>
        <el-descriptions-item label="退款原因" :span="2" v-if="currentOrder.refund_reason">
          {{ currentOrder.refund_reason }}
        </el-descriptions-item>
      </el-descriptions>
      <template #footer v-if="currentOrder?.status === 'pending'">
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="primary" @click="handlePay(currentOrder)">去支付</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { userOrderApi } from '@/api/order'
import PageHeader from '@/components/PageHeader.vue'
import { ShoppingCart } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { formatDate, getOrderStatusName, getOrderStatusType } from '@/composables/useFormat'

const router = useRouter()

interface Order {
  id: number
  order_no: string
  order_type: 'vip' | 'recharge' | 'package'
  package_id: number
  package_name: string
  total_amount: number
  discount_amount: number
  pay_amount: number
  status: 'pending' | 'paid' | 'completed' | 'cancelled' | 'refunded' | 'expired'
  paid_at?: string
  created_at: string
  expire_at?: string
  cancel_reason?: string
  refund_reason?: string
}

const orders = ref<Order[]>([])
const ld = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const detailVisible = ref(false)
const currentOrder = ref<Order | null>(null)

async function load() {
  ld.value = true
  try {
    const res = await userOrderApi.list({
      page: currentPage.value,
      page_size: pageSize.value
    })
    orders.value = res.data.data?.list || []
    total.value = res.data.data?.pagination?.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    ld.value = false
  }
}

function handlePageChange() {
  load()
}

function showDetail(order: Order) {
  currentOrder.value = order
  detailVisible.value = true
}

function handlePay(order: Order) {
  router.push({ path: '/payment', query: { order_no: order.order_no } })
}

onMounted(load)
</script>

<style scoped>
.orders-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.orders-card {
  border-radius: 12px;
}

.orders-card :deep(.el-table) {
  border-radius: 8px;
}

.orders-card :deep(.el-table th.el-table__cell) {
  background: var(--el-fill-color-lighter);
  font-weight: 600;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.orders-card :deep(.el-table td.el-table__cell) {
  font-size: 14px;
}

.amount {
  font-weight: 600;
  color: var(--el-color-primary);
  font-size: 15px;
}

.amount-large {
  font-size: 20px;
  font-weight: 700;
  color: var(--el-color-primary);
}

.orders-card :deep(.el-button--primary) {
  border-radius: 6px;
}

.orders-card :deep(.el-button.is-link) {
  font-weight: 500;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.empty-state {
  padding: 40px 0;
}
</style>
