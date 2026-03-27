<template>
  <div>
    <div class="page-header">
      <h2>API 密钥管理</h2>
      <el-button type="primary" @click="dlg = true">
        <el-icon><Plus /></el-icon> 创建密钥
      </el-button>
    </div>

    <el-card class="token-card">
      <el-table :data="list" v-loading="ld" stripe>
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="密钥" min-width="300">
          <template #default="{ row }">
            <code class="token-key">{{ getDisplayKey(row) }}</code>
            <el-button 
              text 
              size="small" 
              @click="toggleShow(row.id)" 
              :title="showKeys[row.id] ? '隐藏密钥' : '显示密钥'"
            >
              <el-icon><View v-if="!showKeys[row.id]" /><Hide v-else /></el-icon>
            </el-button>
            <el-button 
              text 
              size="small" 
              @click="copyToken(row.token_key)" 
              title="复制密钥"
            >
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="剩余配额" width="140">
          <template #default="{ row }">
            <span class="quota-value">{{ formatQuota(row.remain_quota) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-popconfirm 
              title="确定删除此密钥？删除后不可恢复。" 
              @confirm="del(row.id)"
              confirm-button-text="删除"
              cancel-button-text="取消"
            >
              <template #reference>
                <el-button type="danger" link size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dlg" title="创建 API 密钥" width="500px" :close-on-click-modal="false">
      <el-form :model="f" label-width="100px" class="create-form">
        <el-form-item label="密钥名称" required>
          <el-input v-model="f.name" placeholder="例如：开发环境" maxlength="50" />
        </el-form-item>
        <el-form-item label="可用模型">
          <el-select v-model="f.models" multiple placeholder="全部模型可用" style="width: 100%">
            <el-option v-for="m in availableModels" :key="m.value" :label="m.label" :value="m.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="IP白名单">
          <el-input v-model="f.ips" placeholder="留空不限制，逗号分隔多个IP" />
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker 
            v-model="f.exp" 
            type="datetime" 
            placeholder="永不过期" 
            style="width: 100%"
            :shortcuts="expiryShortcuts"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dlg = false">取消</el-button>
        <el-button type="primary" :loading="crt" @click="create">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { View, Hide, Plus, CopyDocument } from '@element-plus/icons-vue'
import request from '@/api/request'

interface Token {
  id: number
  name: string
  token_key: string
  remain_quota: number
  status: string
  created_at: string
}

const list = ref<Token[]>([])
const ld = ref(false)
const crt = ref(false)
const dlg = ref(false)
const showKeys = ref<Record<number, boolean>>({})

const f = reactive({ 
  name: '', 
  models: [] as string[], 
  ips: '', 
  exp: null as Date | null 
})

const availableModels = [
  { label: 'GPT-3.5-Turbo', value: 'gpt-3.5-turbo' },
  { label: 'GPT-4', value: 'gpt-4' },
  { label: 'GPT-4-Turbo', value: 'gpt-4-turbo' },
  { label: 'Claude-3-Opus', value: 'claude-3-opus' },
  { label: 'Claude-3-Sonnet', value: 'claude-3-sonnet' },
]

const expiryShortcuts = [
  { text: '7天后', value: () => new Date(Date.now() + 7 * 24 * 60 * 60 * 1000) },
  { text: '30天后', value: () => new Date(Date.now() + 30 * 24 * 60 * 60 * 1000) },
  { text: '90天后', value: () => new Date(Date.now() + 90 * 24 * 60 * 60 * 1000) },
]

async function copyToken(key: string) {
  try {
    await navigator.clipboard.writeText(key)
    ElMessage.success('已复制到剪贴板')
  } catch (e) {
    // Fallback for older browsers or non-HTTPS
    const textarea = document.createElement('textarea')
    textarea.value = key
    textarea.style.position = 'fixed'
    textarea.style.left = '-9999px'
    textarea.style.top = '-9999px'
    document.body.appendChild(textarea)
    textarea.focus()
    textarea.select()
    try {
      document.execCommand('copy')
      ElMessage.success('已复制到剪贴板')
    } catch (e2) {
      ElMessage.error('复制失败，请手动复制')
    }
    document.body.removeChild(textarea)
  }
}

function toggleShow(id: number) {
  showKeys.value[id] = !showKeys.value[id]
}

function getDisplayKey(row: Token): string {
  if (showKeys.value[row.id]) {
    return row.token_key
  }
  return row.token_key.substring(0, 8) + '••••••••••••'
}

function formatQuota(quota: number | undefined): string {
  if (!quota) return '0'
  if (quota >= 1000000) return (quota / 1000000).toFixed(1) + 'M'
  if (quota >= 1000) return (quota / 1000).toFixed(1) + 'K'
  return quota.toLocaleString()
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function load() { 
  ld.value = true
  try { 
    const res = await request.get('/tokens')
    list.value = res.data.data || []
    // Reset show state when reloading
    showKeys.value = {}
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '加载失败')
  } finally { 
    ld.value = false 
  }
}

async function create() {
  if (!f.name.trim()) { 
    ElMessage.warning('请输入密钥名称')
    return 
  }
  crt.value = true
  try {
    await request.post('/tokens', { 
      name: f.name.trim(), 
      allowed_models: f.models, 
      allowed_ips: f.ips ? f.ips.split(',').map(s => s.trim()).filter(Boolean) : [], 
      expires_at: f.exp ? f.exp.toISOString() : null,
    })
    ElMessage.success('创建成功')
    dlg.value = false
    f.name = ''
    f.models = []
    f.ips = ''
    f.exp = null
    load()
  } catch(e: any) { 
    ElMessage.error(e.response?.data?.error?.message || '创建失败') 
  } finally { 
    crt.value = false 
  }
}

async function del(id: number) { 
  try {
    await request.delete(`/tokens/${id}`)
    ElMessage.success('已删除')
    load()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error?.message || '删除失败')
  }
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.token-card {
  border-radius: 8px;
}

.token-key {
  display: inline-block;
  padding: 4px 8px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: var(--el-text-color-primary);
  margin-right: 8px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quota-value {
  font-weight: 500;
  color: var(--el-color-success);
}

.create-form {
  padding-top: 10px;
}
</style>
