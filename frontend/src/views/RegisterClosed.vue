<template>
  <div class="closed-page">
    <el-card class="closed-card">
      <div class="closed-content">
        <el-icon class="closed-icon"><CircleClose /></el-icon>
        <h2>注册已关闭</h2>
        <p>目前暂不支持新用户注册，感谢您的理解。</p>
        <div class="closed-actions">
          <el-button type="primary" @click="router.push('/login')">前往登录</el-button>
          <el-button @click="router.push('/')">返回首页</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { CircleClose } from '@element-plus/icons-vue'

const router = useRouter()
</script>

<style scoped>
.closed-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.closed-card {
  width: 420px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.closed-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
  text-align: center;
}

.closed-icon {
  font-size: 64px;
  color: #909399;
  margin-bottom: 16px;
}

.closed-content h2 {
  margin: 0 0 12px;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.closed-content p {
  margin: 0 0 24px;
  font-size: 14px;
  color: #909399;
  line-height: 1.6;
}

.closed-actions {
  display: flex;
  gap: 12px;
}
</style>