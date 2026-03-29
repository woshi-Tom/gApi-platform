<template>
  <div class="init-wizard">
    <el-card class="wizard-card">
      <template #header>
        <div class="wizard-header">
          <el-icon class="logo-icon"><Setting /></el-icon>
          <h2>gAPI Platform 初始化向导</h2>
        </div>
      </template>

      <el-steps :active="step - 1" finish-status="success" align-center style="margin-bottom: 30px">
        <el-step title="数据库" />
        <el-step title="Redis" />
        <el-step title="管理员" />
        <el-step title="完成" />
      </el-steps>

      <!-- Step 1: Database Connection -->
      <div v-if="step === 1" class="step-content">
        <h3>配置数据库连接</h3>
        <p class="text-muted">输入 PostgreSQL 数据库连接信息</p>
        
        <el-form :model="dbForm" :rules="dbRules" ref="dbFormRef" label-position="top" style="margin-top: 20px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="主机地址" prop="host">
                <el-input v-model="dbForm.host" placeholder="localhost 或 IP 地址" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="端口" prop="port">
                <el-input-number v-model="dbForm.port" :min="1" :max="65535" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="用户名" prop="user">
                <el-input v-model="dbForm.user" placeholder="数据库用户名" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="密码" prop="password">
                <el-input v-model="dbForm.password" type="password" show-password placeholder="数据库密码" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="数据库名称" prop="dbname">
            <el-input v-model="dbForm.dbname" placeholder="数据库名称" />
          </el-form-item>
        </el-form>

        <el-alert v-if="dbStatus" :type="dbStatusType" show-icon :closable="false" style="margin: 16px 0">
          <template #title>{{ dbStatus }}</template>
        </el-alert>

        <div class="step-actions">
          <el-button type="primary" :loading="dbTesting" @click="testAndInitDB">测试连接并初始化</el-button>
        </div>
      </div>

      <!-- Step 2: Redis Connection -->
      <div v-if="step === 2" class="step-content">
        <h3>配置 Redis 连接</h3>
        <p class="text-muted">输入 Redis 连接信息（可选）</p>
        
        <el-form :model="redisForm" label-position="top" style="margin-top: 20px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="主机地址">
                <el-input v-model="redisForm.host" placeholder="localhost" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="端口">
                <el-input-number v-model="redisForm.port" :min="1" :max="65535" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="密码">
            <el-input v-model="redisForm.password" type="password" show-password placeholder="留空表示无密码" />
          </el-form-item>
        </el-form>

        <el-alert v-if="redisStatus" :type="redisStatusType" show-icon :closable="false" style="margin: 16px 0">
          <template #title>{{ redisStatus }}</template>
        </el-alert>

        <div class="step-actions">
          <el-button @click="step = 1">上一步</el-button>
          <el-button type="primary" @click="step = 3">下一步</el-button>
        </div>
      </div>

      <!-- Step 3: Admin Account -->
      <div v-if="step === 3" class="step-content">
        <h3>创建管理员账户</h3>
        <p class="text-muted">设置系统管理员登录凭证</p>
        
        <el-form :model="adminForm" :rules="adminRules" ref="adminFormRef" label-position="top" style="margin-top: 20px">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="adminForm.username" placeholder="admin" />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="adminForm.email" placeholder="admin@example.com" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="adminForm.password" type="password" show-password placeholder="至少6位" />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input v-model="adminForm.confirmPassword" type="password" show-password placeholder="再次输入密码" />
          </el-form-item>
          
          <div v-if="adminForm.password" class="password-strength">
            <span>密码强度：</span>
            <el-tag :type="passwordStrengthType" size="small">{{ passwordStrengthText }}</el-tag>
          </div>
        </el-form>

        <el-alert v-if="adminStatus" :type="adminStatusType" show-icon :closable="false" style="margin: 16px 0">
          <template #title>{{ adminStatus }}</template>
        </el-alert>

        <div class="step-actions">
          <el-button @click="step = 2">上一步</el-button>
          <el-button type="primary" :loading="creatingAdmin" @click="createAdmin">创建管理员</el-button>
        </div>
      </div>

      <!-- Step 4: Complete -->
      <div v-if="step === 4" class="step-content">
        <div class="success-content">
          <el-icon class="success-icon"><CircleCheckFilled /></el-icon>
          <h3>初始化完成！</h3>
          <p class="text-muted">系统已准备就绪，可以开始使用了</p>
        </div>
        <div class="step-actions">
          <el-button type="primary" size="large" @click="goToAdmin">前往管理后台</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Setting, CircleCheckFilled } from '@element-plus/icons-vue'
import axios from 'axios'

const router = useRouter()
const step = ref(1)
const dbFormRef = ref<FormInstance>()
const adminFormRef = ref<FormInstance>()

const apiBase = import.meta.env.VITE_API_BASE_URL || ''

const dbForm = reactive({
  host: 'localhost',
  port: 5432,
  user: 'gapi',
  password: '',
  dbname: 'gapi'
})
const dbTesting = ref(false)
const dbConnected = ref(false)
const dbStatus = ref('')
const dbStatusType = computed(() => dbConnected.value ? 'success' : 'error')

const dbRules: FormRules = {
  host: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  user: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  dbname: [{ required: true, message: '请输入数据库名称', trigger: 'blur' }]
}

const redisForm = reactive({
  host: 'localhost',
  port: 6379,
  password: ''
})
const redisConnected = ref(false)
const redisStatus = ref('')
const redisStatusType = computed(() => redisConnected.value ? 'success' : 'error')

const adminForm = reactive({
  username: 'admin',
  email: 'admin@example.com',
  password: '',
  confirmPassword: ''
})
const creatingAdmin = ref(false)
const adminCreated = ref(false)
const adminStatus = ref('')
const adminStatusType = computed(() => adminCreated.value ? 'success' : 'error')

const adminRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, message: '用户名至少3位', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效邮箱', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: any) => {
        if (value !== adminForm.password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const passwordStrength = computed(() => {
  const pwd = adminForm.password
  if (!pwd) return 0
  let score = 0
  if (pwd.length >= 6) score++
  if (pwd.length >= 8) score++
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++
  return score
})

const passwordStrengthType = computed(() => {
  const s = passwordStrength.value
  if (s <= 2) return 'danger'
  if (s <= 3) return 'warning'
  return 'success'
})

const passwordStrengthText = computed(() => {
  const s = passwordStrength.value
  if (s <= 1) return '弱'
  if (s <= 2) return '中等'
  if (s <= 3) return '良好'
  return '强'
})

async function testAndInitDB() {
  if (!dbFormRef.value) return
  
  try {
    await dbFormRef.value.validate()
  } catch {
    return
  }

  dbTesting.value = true
  dbStatus.value = ''
  
  try {
    const response = await axios.post(`${apiBase}/api/v1/init/test-db-with-config`, {
      host: dbForm.host,
      port: dbForm.port,
      user: dbForm.user,
      password: dbForm.password,
      dbname: dbForm.dbname
    })
    
    if (response.data.success) {
      dbConnected.value = true
      dbStatus.value = '数据库连接成功，正在初始化表结构...'
      
      const initResponse = await axios.post(`${apiBase}/api/v1/init/init-db`, {
        host: dbForm.host,
        port: dbForm.port,
        user: dbForm.user,
        password: dbForm.password,
        dbname: dbForm.dbname
      })
      
      if (initResponse.data.success) {
        dbStatus.value = '数据库连接成功，表结构已初始化'
        setTimeout(() => {
          step.value = 2
        }, 1000)
      }
    }
  } catch (e: any) {
    dbConnected.value = false
    const msg = e.response?.data?.error?.message || e.message || '连接失败'
    dbStatus.value = msg
    ElMessage.error(msg)
  } finally {
    dbTesting.value = false
  }
}

function createAdmin() {
  if (!adminFormRef.value) return
  
  adminFormRef.value.validate(async (valid) => {
    if (!valid) return

    if (adminForm.password !== adminForm.confirmPassword) {
      ElMessage.error('两次输入密码不一致')
      return
    }

    creatingAdmin.value = true
    adminStatus.value = ''
    
    try {
      const response = await axios.post(`${apiBase}/api/v1/init/create-admin`, {
        username: adminForm.username,
        password: adminForm.password,
        email: adminForm.email
      })
      
      if (response.data.success) {
        adminCreated.value = true
        adminStatus.value = '管理员账户创建成功！'
        ElMessage.success('管理员账户创建成功')
        setTimeout(() => {
          step.value = 4
        }, 1000)
      }
    } catch (e: any) {
      adminCreated.value = false
      const msg = e.response?.data?.error?.message || e.message || '创建失败'
      adminStatus.value = msg
      ElMessage.error(msg)
    } finally {
      creatingAdmin.value = false
    }
  })
}

function goToAdmin() {
  router.push('/login')
}
</script>

<style scoped>
.init-wizard {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  padding: 20px;
}

.wizard-card {
  width: 600px;
  max-width: 100%;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.wizard-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.wizard-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.logo-icon {
  font-size: 36px;
  color: #409eff;
}

.step-content {
  padding: 20px 0;
}

.step-content h3 {
  margin: 0 0 8px 0;
  text-align: center;
}

.text-muted {
  color: #909399;
  text-align: center;
  margin-bottom: 0;
}

.success-content {
  text-align: center;
  padding: 20px 0;
}

.success-icon {
  font-size: 64px;
  color: #67c23a;
  margin-bottom: 16px;
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #606266;
  margin-top: -10px;
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}
</style>
