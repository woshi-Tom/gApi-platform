<template>
  <div class="admin-dashboard">
    <div class="page-header">
      <h2>管理后台仪表盘</h2>
      <p class="subtitle">系统运行状态概览</p>
    </div>

    <!-- User Stats -->
    <div class="stats-section">
      <h3 class="section-title">用户统计</h3>
      <div class="stats-grid">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon blue">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total_users || 0 }}</div>
            <div class="stat-label">总用户数</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon green">
            <el-icon><UserFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.active_users_today || 0 }}</div>
            <div class="stat-label">今日活跃</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon orange">
            <el-icon><Star /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.vip_users_count || 0 }}</div>
            <div class="stat-label">VIP用户</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon red">
            <el-icon><Connection /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total_channels || 0 }}</div>
            <div class="stat-label">渠道数量</div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- Business Stats -->
    <div class="stats-section">
      <h3 class="section-title">业务统计</h3>
      <div class="stats-grid">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon cyan">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.healthy_channels || 0 }}</div>
            <div class="stat-label">健康渠道</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon purple">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.total_orders_today || 0 }}</div>
            <div class="stat-label">今日订单</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon pink">
            <el-icon><Money /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">¥{{ stats.total_revenue_today || 0 }}</div>
            <div class="stat-label">今日收入</div>
          </div>
        </el-card>
        
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon teal">
            <el-icon><TrendCharts /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ formatQuota(stats.total_quota_used_today) }}</div>
            <div class="stat-label">今日用量</div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- Quick Actions -->
    <el-card class="actions-card">
      <template #header>
        <span>快捷操作</span>
      </template>
      <div class="actions-grid">
        <el-button @click="$router.push('/users')">
          <el-icon><User /></el-icon> 用户管理
        </el-button>
        <el-button @click="$router.push('/channels')">
          <el-icon><Connection /></el-icon> 渠道管理
        </el-button>
        <el-button @click="$router.push('/orders')">
          <el-icon><Document /></el-icon> 订单管理
        </el-button>
        <el-button @click="$router.push('/logs')">
          <el-icon><Clock /></el-icon> 操作日志
        </el-button>
        <el-button @click="$router.push('/settings')">
          <el-icon><Setting /></el-icon> 系统设置
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/api/request'
import {
  User, UserFilled, Star, Connection, CircleCheck,
  Document, Money, TrendCharts, Clock, Setting
} from '@element-plus/icons-vue'

interface Stats {
  total_users: number
  active_users_today: number
  vip_users_count: number
  total_channels: number
  healthy_channels: number
  total_orders_today: number
  total_revenue_today: number
  total_quota_used_today: number
}

const stats = ref<Partial<Stats>>({})

function formatQuota(n: number | undefined): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

onMounted(async () => {
  try {
    const res = await request.get('/admin/stats/overview')
    stats.value = res.data.data || {}
  } catch (e) {
    console.error('Failed to load stats:', e)
  }
})
</script>

<style scoped>
.admin-dashboard {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  margin-bottom: 10px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

/* Stats Section */
.stats-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  border-radius: 10px;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: #fff;
}

.stat-icon.blue { background: linear-gradient(135deg, #409eff 0%, #337ecc 100%); }
.stat-icon.green { background: linear-gradient(135deg, #67c23a 0%, #529b2e 100%); }
.stat-icon.orange { background: linear-gradient(135deg, #e6a23c 0%, #b88230 100%); }
.stat-icon.red { background: linear-gradient(135deg, #f56c6c 0%, #c45656 100%); }
.stat-icon.cyan { background: linear-gradient(135deg, #17c0eb 0%, #13a6cf 100%); }
.stat-icon.purple { background: linear-gradient(135deg, #9c27b0 0%, #7b1fa2 100%); }
.stat-icon.pink { background: linear-gradient(135deg, #e91e63 0%, #c2185b 100%); }
.stat-icon.teal { background: linear-gradient(135deg, #009688 0%, #00796b 100%); }

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.stat-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

/* Actions Card */
.actions-card :deep(.el-card__header) {
  font-weight: 500;
}

.actions-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.actions-grid .el-button {
  min-width: 120px;
}

/* Responsive */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
