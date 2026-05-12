<template>
  <div class="channel-management">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <h2>渠道管理</h2>
        <p class="subtitle">共 {{ pagination.total }} 个渠道</p>
      </div>
      <div class="header-actions">
        <el-button type="info" :loading="batchChecking" @click="batchCheckAll">
          <el-icon><Refresh /></el-icon> 批量检测
        </el-button>
        <el-button type="warning" @click="importDialogVisible = true">
          批量导入
        </el-button>
        <el-button type="primary" @click="showAdd">
          <el-icon><Plus /></el-icon> 添加渠道
        </el-button>
      </div>
    </div>

    <!-- Filters -->
    <el-card class="filter-card">
      <el-form :inline="true" class="filter-form">
        <el-form-item label="渠道类型">
          <el-select v-model="filters.type" clearable placeholder="全部" style="width:140px" @change="load">
            <el-option v-for="t in CHANNEL_TYPES" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" style="width:100px" @change="load">
            <el-option v-for="s in CHANNEL_STATUS" :key="s.value" :label="s.label" :value="s.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" clearable placeholder="名称/地址" style="width:160px" @change="load" />
        </el-form-item>
        <el-form-item>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Table -->
    <el-card class="table-card">
      <el-table :data="channels" v-loading="ld" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column label="类型" width="130">
          <template #default="{ row }">
            <el-tag size="small">{{ getChannelTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="base_url" label="地址" min-width="220" show-overflow-tooltip />
        <el-table-column label="模型" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="Array.isArray(row.models)">{{ row.models.slice(0, 3).join(', ') }}<span v-if="row.models.length > 3">...</span></span>
            <span v-else-if="row.models">{{ String(row.models).split(',').slice(0, 3).join(', ') }}<span v-if="String(row.models).split(',').length > 3">...</span></span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="权重" width="70">
          <template #default="{ row }">{{ row.weight }}</template>
        </el-table-column>
        <el-table-column label="优先级" width="80">
          <template #default="{ row }">{{ row.priority }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status===1?'success':row.status===2?'warning':'danger'" size="small">
              {{ row.status===1?'启用':row.status===2?'维护':'禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="健康状态" width="180">
          <template #default="{ row }">
            <div class="health-cell">
              <span :class="'health-dot ' + (row.is_healthy ? 'healthy' : row.failure_count > 0 ? 'degraded' : 'unhealthy')"></span>
              <div class="health-info">
                <span>{{ row.is_healthy ? '正常' : '异常' }}</span>
                <span v-if="row.last_check_at" class="health-time">{{ formatLastCheck(row.last_check_at) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="失败" width="70">
          <template #default="{ row }">
            <span :style="{ color: getFailureColor(row.failure_count) }">{{ row.failure_count }}</span>
          </template>
        </el-table-column>
        <el-table-column label="响应时间" width="90">
          <template #default="{ row }">{{ row.response_time_avg > 0 ? row.response_time_avg + 'ms' : '-' }}</template>
        </el-table-column>
        <el-table-column label="代理" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.proxy_enabled" type="warning" size="small">{{ row.proxy_type || 'proxy' }}</el-tag>
            <span v-else style="color:#c0c4cc">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="(cmd: string) => handleAction(cmd, row)">
              <el-button size="small" link type="primary">
                操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="test">测试</el-dropdown-item>
                  <el-dropdown-item command="history">测试历史</el-dropdown-item>
                  <el-dropdown-item command="health" :disabled="healthChecking[row.id]">
                    {{ healthChecking[row.id] ? '检测中...' : '检测' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="toggle">
                    {{ row.status===1?'禁用':'启用' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided style="color: var(--el-color-danger)">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
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

    <el-dialog v-model="dlgVisible" :title="isEdit?'编辑渠道':'添加渠道'" width="700px">
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="渠道名称" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" style="width:100%">
            <el-option v-for="t in CHANNEL_TYPES" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="地址" prop="base_url">
          <el-input v-model="form.base_url" placeholder="https://api.openai.com/v1" />
        </el-form-item>
        <el-form-item label="API Key" prop="api_key">
          <el-input v-model="form.api_key" show-password placeholder="sk-xxxx" />
        </el-form-item>
        <el-form-item label="模型列表">
          <el-select v-model="form.models" multiple filterable allow-create default-first-option style="width:100%" placeholder="输入或选择模型">
            <el-option v-for="m in commonModels" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型映射">
          <el-input v-model="form.model_mapping" type="textarea" :rows="2" placeholder='{"gpt-4": "gpt-4-0613"}' />
          <div style="color:#909399;font-size:12px;margin-top:4px">JSON格式，将请求模型映射到渠道支持的模型</div>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="权重">
              <el-input-number v-model="form.weight" :min="1" :max="100" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="优先级">
              <el-input-number v-model="form.priority" :min="0" :max="100" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="分组">
              <el-input v-model="form.group_name" placeholder="分组名称，如: default, pro" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代理设置">
              <el-switch v-model="form.proxy_enabled" active-text="启用" inactive-text="关闭" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16" v-if="form.proxy_enabled">
          <el-col :span="8">
            <el-form-item label="代理类型">
              <el-select v-model="form.proxy_type" style="width:100%">
                <el-option label="SOCKS5" value="socks5" />
                <el-option label="HTTP" value="http" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="16">
            <el-form-item label="代理地址">
              <el-input v-model="form.proxy_url" placeholder="socks5://user:pass@host:port" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dlgVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="testVisible" title="测试渠道" width="500px">
      <el-form :model="testForm" label-width="100px">
        <el-form-item label="测试类型">
          <el-radio-group v-model="testForm.test_type">
            <el-radio label="models">获取模型列表</el-radio>
            <el-radio label="chat">聊天测试</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="模型" v-if="testForm.test_type==='chat'">
          <el-select v-model="testForm.model" style="width:100%" placeholder="请先获取模型列表">
            <el-option v-for="m in testChannelModels" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item label="测试消息" v-if="testForm.test_type==='chat'">
          <el-input v-model="testForm.messages" type="textarea" :rows="3" placeholder="输入测试消息" />
        </el-form-item>
      </el-form>
      <div v-if="testResult" style="margin-top:16px">
        <el-alert :type="testResult.success?'success':'error'" :title="testResult.success?'测试成功':'测试失败'" show-icon>
          <template #default>
            <div>响应时间: {{ testResult.response_time_ms }}ms</div>
            <div v-if="testResult.error">错误: {{ testResult.error }}</div>
            <div v-if="testResult.content">响应: {{ testResult.content }}</div>
            <div v-if="testResult.models && testResult.models.length > 0" style="margin-top:8px">
              <div style="font-weight:500;margin-bottom:4px">模型列表 ({{ testResult.models.length }}个):</div>
              <div style="max-height:200px;overflow-y:auto;font-size:12px;">
                <el-tag v-for="m in testResult.models" :key="m" size="small" style="margin:2px">{{ m }}</el-tag>
              </div>
            </div>
          </template>
        </el-alert>
      </div>
      <template #footer>
        <el-button @click="testVisible=false">关闭</el-button>
        <el-button type="primary" :loading="testing" @click="runTest">运行测试</el-button>
      </template>
    </el-dialog>

    <!-- 测试历史抽屉 -->
    <el-drawer v-model="historyVisible" :title="historyChannelName + ' - 测试历史'" size="600px">
      <el-table :data="testHistory" v-loading="historyLoading" stripe size="small">
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.test_type==='models'?'':row.test_type==='chat'?'success':'warning'">{{ row.test_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.success?'success':'danger'" size="small">{{ row.success?'成功':'失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status_code" label="HTTP" width="70" />
        <el-table-column label="响应时间" width="90">
          <template #default="{ row }">{{ row.response_time_ms }}ms</template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ row.error_message || '-' }}</template>
        </el-table-column>
      </el-table>
      <div v-if="testHistory.length === 0 && !historyLoading" style="text-align:center;padding:40px;color:#909399">
        暂无测试记录
      </div>
    </el-drawer>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="importDialogVisible" title="批量导入渠道" width="500px">
      <div style="margin-bottom:16px">
        <p style="color:#606266;font-size:14px;margin-bottom:12px">上传 CSV 文件批量导入渠道，文件格式要求：</p>
        <div style="background:#f5f7fa;padding:12px;border-radius:4px;font-size:13px;font-family:monospace;word-break:break-all">
          <div style="font-weight:600;margin-bottom:4px">CSV 表头（必填列标 *）：</div>
          <div>name(*), type(*), base_url(*), api_key(*), models, weight, priority, group_name</div>
          <div style="margin-top:8px;color:#909399">models 列用 | 分隔多个模型，如: gpt-4|gpt-3.5-turbo</div>
        </div>
      </div>
      <el-upload
        :auto-upload="false"
        :limit="1"
        accept=".csv"
        :on-change="handleFileChange"
        :on-remove="() => importFile = null"
        drag
      >
        <div style="padding:20px">
          <div style="font-size:28px;margin-bottom:8px"> </div>
          <div>将 CSV 文件拖到此处，或 <em>点击上传</em></div>
        </div>
      </el-upload>
      <div v-if="importResult" style="margin-top:16px">
        <el-alert :type="importResult.fail_count === 0 ? 'success' : 'warning'" :closable="false">
          <div>成功导入: <strong>{{ importResult.success_count }}</strong> 条</div>
          <div v-if="importResult.fail_count > 0">失败: <strong>{{ importResult.fail_count }}</strong> 条</div>
          <div v-if="importResult.errors" style="margin-top:8px;font-size:12px;max-height:120px;overflow-y:auto">
            <div v-for="e in importResult.errors" :key="e" style="color:#f56c6c">{{ e }}</div>
          </div>
        </el-alert>
      </div>
      <template #footer>
        <el-button @click="importDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="importing" :disabled="!importFile" @click="doImport">开始导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { channelApi, CHANNEL_TYPES, CHANNEL_STATUS } from '@/api/channel'
import type { Channel, ChannelTestResult, ChannelTestHistory } from '@/api/channel'

const api = channelApi

const channels = ref<Channel[]>([])
const ld = ref(false)
const dlgVisible = ref(false)
const testVisible = ref(false)
const saving = ref(false)
const testing = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const testChannel = ref<Channel | null>(null)
const testResult = ref<ChannelTestResult | null>(null)

// Test history
const historyVisible = ref(false)
const historyLoading = ref(false)
const historyChannelName = ref('')
const testHistory = ref<ChannelTestHistory[]>([])

// Import
const importDialogVisible = ref(false)
const importing = ref(false)
const importFile = ref<File | null>(null)
const importResult = ref<{ success_count: number; fail_count: number; errors?: string[] } | null>(null)

// Health check loading state
const healthChecking = ref<Record<number, boolean>>({})
const batchChecking = ref(false)

// Batch health check all channels
async function batchCheckAll() {
  if (channels.value.length === 0) {
    ElMessage.warning('暂无渠道数据')
    return
  }
  batchChecking.value = true
  let successCount = 0
  let failCount = 0
  
  for (const channel of channels.value) {
    healthChecking.value[channel.id] = true
    try {
      const res = await api.triggerHealthCheck(channel.id)
      const data = res.data.data || res.data
      if (data.is_healthy) {
        successCount++
      } else {
        failCount++
      }
    } catch {
      failCount++
    } finally {
      healthChecking.value[channel.id] = false
    }
  }
  
  ElMessage.success(`检测完成: ${successCount} 正常, ${failCount} 异常`)
  load()
  batchChecking.value = false
}

const commonModels = [
  'gpt-3.5-turbo', 'gpt-4', 'gpt-4-turbo', 'gpt-4o',
  'claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku',
  'gemini-pro', 'gemini-1.5-pro', 'deepseek-chat', 'deepseek-coder',
  'nvidia/llama-3.1-nemotron-70b-instruct', 'nvidia/llama-3.3-70b-instruct',
]

const DEFAULT_BASE_URLS: Record<string, string> = {
  openai: 'https://api.openai.com/v1',
  nvidia: 'https://integrate.api.nvidia.com/v1',
  azure: 'https://{your-resource}.openai.azure.com',
  claude: 'https://api.anthropic.com',
  gemini: 'https://generativelanguage.googleapis.com/v1beta',
  deepseek: 'https://api.deepseek.com/v1',
  zhipu: 'https://open.bigmodel.cn/api/paas/v4',
  baidu: 'https://qianfan.baidubce.com/v2',
  yi: 'https://api.01.ai/v1',
  groq: 'https://api.groq.com/openai/v1',
  ollama: 'http://localhost:11434/v1',
  localai: 'http://localhost:8080/v1',
  custom: '',
}

const filters = reactive({
  type: '',
  status: '',
  group: '',
  keyword: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  id: 0,
  name: '',
  type: 'openai',
  base_url: '',
  api_key: '',
  models: [] as string[],
  model_mapping: '',
  weight: 100,
  priority: 0,
  status: 1,
  group_name: '',
  timeout: 60000,
  proxy_enabled: false,
  proxy_type: 'socks5',
  proxy_url: '',
})

const testForm = reactive({
  test_type: 'models' as 'models' | 'chat',
  model: '',
  messages: 'Hello, world!',
})

const testChannelModels = computed(() => {
  // 首先从 testResult 获取（测试刚获取的）
  if (testResult.value?.models && testResult.value.models.length > 0) {
    console.log('Using testResult models:', testResult.value.models.length)
    return testResult.value.models
  }
  // 其次从 testChannel 获取（可能是数组或 JSON 字符串）
  const m = testChannel.value?.models
  console.log('testChannel.models:', m, typeof m)
  if (!m) return []
  if (Array.isArray(m)) {
    console.log('Using testChannel as array, length:', m.length)
    return m
  }
  if (typeof m === 'string') {
    try {
      const parsed = JSON.parse(m)
      if (Array.isArray(parsed)) {
        console.log('Using parsed string, length:', parsed.length)
        return parsed
      }
    } catch (e) {
      console.log('Parse failed')
    }
  }
  return []
})

// 当测试结果更新时，同步到 testChannel.models 以便持久化
watch(() => testResult.value?.models, (newModels) => {
  if (newModels && newModels.length > 0 && testChannel.value) {
    testChannel.value.models = newModels
  }
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择渠道类型', trigger: 'change' }],
  base_url: [{ required: true, message: '请输入API地址', trigger: 'blur' }],
  api_key: [{ required: true, message: '请输入API Key', trigger: 'blur' }],
}

const getChannelTypeName = (type: string) => {
  const t = CHANNEL_TYPES.find(t => t.value === type)
  return t?.label || type
}

const formatLastCheck = (timestamp: string) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000)
  
  if (diff < 60) return `${diff}秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  return date.toLocaleString('zh-CN', { hour12: false })
}

const getFailureColor = (count: number) => {
  if (count === 0) return '#67c23a'
  if (count <= 2) return '#e6a23c'
  return '#f56c6c'
}

const resetFilters = () => {
  filters.type = ''
  filters.status = ''
  filters.group = ''
  filters.keyword = ''
  load()
}

const showAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    id: 0, name: '', type: 'openai', base_url: '', api_key: '',
    models: [], model_mapping: '', weight: 100, priority: 0, status: 1, group_name: '', timeout: 60000,
    proxy_enabled: false, proxy_type: 'socks5', proxy_url: '',
  })
  autoFillBaseURL()
  dlgVisible.value = true
}

const autoFillBaseURL = () => {
  if (!isEdit.value && form.type && DEFAULT_BASE_URLS[form.type]) {
    form.base_url = DEFAULT_BASE_URLS[form.type]
  }
}

const edit = (c: Channel) => {
  isEdit.value = true
  Object.assign(form, c)
  form.api_key = ''
  if (typeof form.model_mapping === 'object') {
    form.model_mapping = JSON.stringify(form.model_mapping, null, 2)
  } else if (!form.model_mapping) {
    form.model_mapping = ''
  }
  dlgVisible.value = true
}

const save = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const data = { ...form }
      if (!data.api_key) delete data.api_key
      if (isEdit.value) {
        await api.update(form.id, data)
        ElMessage.success('更新成功')
      } else {
        await api.create(data)
        ElMessage.success('添加成功')
      }
      dlgVisible.value = false
      load()
    } catch (e: any) {
      ElMessage.error(e.response?.data?.error?.message || '保存失败')
    } finally {
      saving.value = false
    }
  })
}

const test = async (c: Channel) => {
  // 如果是打开不同渠道，才重置所有数据
  if (!testChannel.value || testChannel.value.id !== c.id) {
    testChannel.value = { ...c }
    testResult.value = null
    testForm.test_type = 'models'
    testForm.model = ''
  }
  testVisible.value = true

  // 如果 testChannel 没有模型，从 c 复制（但保持不覆盖已获取的模型）
  if (!testChannel.value.models && c.models) {
    testChannel.value.models = c.models
  }

  // 解析模型列表（可能是 JSON 字符串）
  const parseModels = (m: any): string[] => {
    if (!m) return []
    if (Array.isArray(m)) return m
    if (typeof m === 'string') {
      try {
        const parsed = JSON.parse(m)
        return Array.isArray(parsed) ? parsed : []
      } catch {
        return []
      }
    }
    return []
  }

  // 自动获取模型列表（如果表格数据中没有）
  let existingModels = parseModels(testChannel.value?.models)
  
  if (existingModels.length === 0) {
    // 从表格数据获取并解析
    existingModels = parseModels(c.models)
    if (existingModels.length > 0) {
      // 存回 testChannel 以便后续使用
      testChannel.value!.models = existingModels
    }
  }
  
  if (existingModels.length > 0) {
    testForm.model = existingModels[0]
  } else {
    try {
      const res = await api.test(c.id, { test_type: 'models' })
      const models = res.data.data?.models || []
      if (models.length > 0) {
        testChannel.value!.models = models
        testForm.model = models[0]
      }
    } catch {
      // 静默失败，用户可手动点击运行测试
    }
  }
}

const runTest = async () => {
  if (!testChannel.value) return
  testing.value = true
  try {
    const data: any = { test_type: testForm.test_type }
    if (testForm.test_type === 'chat') {
      data.model = testForm.model
      data.messages = [{ role: 'user', content: testForm.messages }]
    }
    const res = await api.test(testChannel.value.id, data)
    testResult.value = res.data.data
    if (testResult.value?.models && testResult.value.models.length > 0) {
      testChannel.value.models = testResult.value.models
      if (!testForm.model) {
        testForm.model = testResult.value.models[0]
      }
    }
  } catch (e: any) {
    testResult.value = {
      success: false,
      error: e.response?.data?.error?.message || '测试失败',
      response_time_ms: 0,
      status_code: e.response?.status || 0,
    }
  } finally {
    testing.value = false
  }
}

const toggleStatus = async (c: Channel) => {
  try {
    if (c.status === 1) {
      await api.disable(c.id)
    } else {
      await api.enable(c.id)
    }
    ElMessage.success('操作成功')
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '操作失败')
  }
}

const checkHealth = async (c: Channel) => {
  healthChecking.value[c.id] = true
  try {
    const res = await api.triggerHealthCheck(c.id)
    const data = res.data.data || res.data
    if (data.is_healthy) {
      ElMessage.success(`检测成功，响应时间: ${data.response_time_ms}ms`)
    } else {
      ElMessage.warning(`检测失败: ${data.last_error || '连接失败'}`)
    }
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || e.response?.data?.error?.message || '检测失败')
  } finally {
    healthChecking.value[c.id] = false
  }
}

const del = async (id: number) => {
  try {
    await api.delete(id)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '删除失败')
  }
}

const handleAction = (cmd: string, row: Channel) => {
  switch (cmd) {
    case 'edit':
      edit(row)
      break
    case 'test':
      test(row)
      break
    case 'history':
      showHistory(row)
      break
    case 'health':
      checkHealth(row)
      break
    case 'toggle':
      toggleStatus(row)
      break
    case 'delete':
      del(row.id)
      break
  }
}

const formatTime = (ts: string) => {
  if (!ts) return '-'
  const d = new Date(ts)
  return d.toLocaleString('zh-CN', { hour12: false })
}

const showHistory = async (c: Channel) => {
  historyChannelName.value = c.name
  historyVisible.value = true
  historyLoading.value = true
  try {
    const res = await channelApi.getTestHistory(c.id)
    testHistory.value = res.data.data || []
  } catch {
    testHistory.value = []
  } finally {
    historyLoading.value = false
  }
}

const handleFileChange = (file: any) => {
  importFile.value = file.raw
  importResult.value = null
}

const doImport = async () => {
  if (!importFile.value) return
  importing.value = true
  try {
    const res = await channelApi.importChannels(importFile.value)
    importResult.value = res.data.data
    if (importResult.value && importResult.value.success_count > 0) {
      load()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '导入失败')
  } finally {
    importing.value = false
  }
}

const load = async () => {
  ld.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    if (filters.type) params.type = filters.type
    if (filters.status) params.status = filters.status
    if (filters.group) params.group = filters.group
    if (filters.keyword) params.keyword = filters.keyword
    
    const res = await api.list(params)
    if (res.data.data) {
      channels.value = res.data.data.list || res.data.data
      pagination.total = res.data.data.pagination?.total || channels.value.length
    }
  } catch (e: any) {
    ElMessage.error('加载失败: ' + (e.response?.data?.error?.message || e.message))
  } finally {
    ld.value = false
  }
}

onMounted(load)

watch(() => form.type, () => {
  if (!isEdit.value) {
    autoFillBaseURL()
  }
})
</script>

<style scoped>
.channel-management {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px;
}

.header-left h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-left .subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: #909399;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* Filter Card */
.filter-card {
  border-radius: 8px;
}

.filter-card :deep(.el-card__body) {
  padding: 16px 20px;
}

.filter-form {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-form :deep(.el-form-item) {
  margin-bottom: 0;
  margin-right: 12px;
}

/* Table Card */
.table-card {
  border-radius: 8px;
}

.table-card :deep(.el-card__body) {
  padding: 0;
}

/* Health Cell */
.health-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.health-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.health-info span:first-child {
  font-size: 13px;
  color: #303133;
}

.health-time {
  font-size: 11px;
  color: #909399;
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  padding: 4px 8px;
  font-size: 12px;
}

/* Pagination */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid #ebeef5;
}

/* Health Dot */
.health-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}
.health-dot.healthy {
  background: #67c23a;
}
.health-dot.degraded {
  background: #e6a23c;
}
.health-dot.unhealthy {
  background: #f56c6c;
}
</style>
