<template>
  <div class="admin-models">
    <div class="page-header" style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;gap:12px;">
      <div>
        <h2 style="margin:0">模型管理</h2>
        <p class="subtitle" style="margin:4px 0 0; color:#909399; font-size:12px">分组、定价和用户分组管理</p>
      </div>
      <el-tag type="info" hollow>Admin</el-tag>
    </div>

    <el-tabs v-model="activeTab" type="card" lazy>
      <el-tab-pane label="模型分组" name="groups" >
          <el-card>
          <div style="display:flex;justify-content:flex-end;margin-bottom:12px;gap:8px;">
            <el-button type="primary" @click="openGroupDialog(false)"><el-icon><Plus /></el-icon> 添加分组</el-button>
          </div>
          <el-table :data="modelGroups" v-loading="ldGroups" stripe>
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="display_name" label="显示名称" min-width="180" />
            <el-table-column prop="description" label="描述" min-width="260" show-overflow-tooltip />
            <el-table-column label="排序" width="80">
              <template #default="{ row }">{{ row.sort_order }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status === 'active' ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="openGroupDialog(true, row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteGroup(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
        <!-- Group Dialog -->
        <el-dialog :title="groupDialogMode ? '编辑分组' : '创建分组'" v-model:visible="groupDialogVisible" width="700px">
          <el-form :model="groupForm" label-width="120px" ref="groupFormRef">
            <el-form-item label="名称" prop="name" required>
              <el-input v-model="groupForm.name" placeholder="分组英文名，唯一标识" />
            </el-form-item>
            <el-form-item label="显示名称" prop="display_name" required>
              <el-input v-model="groupForm.display_name" placeholder="用户看到的显示名称" />
            </el-form-item>
            <el-form-item label="描述" prop="description">
              <el-input v-model="groupForm.description" type="textarea" rows="3" placeholder="描述信息" />
            </el-form-item>
            <el-form-item label="排序" prop="sort_order">
              <el-input-number v-model="groupForm.sort_order" :min="0" :max="9999" />
            </el-form-item>
            <el-form-item v-if="groupDialogMode" label="状态" prop="status">
              <el-switch v-model="groupForm.status" active-value="active" inactive-value="inactive" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="groupDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="saveGroup" :loading="groupSaving">保存</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="模型定价" name="pricing">
        <el-card>
          <div style="display:flex;justify-content:flex-end;margin-bottom:12px;gap:8px;">
            <el-button type="primary" @click="openPricingDialog(false)">添加定价</el-button>
          </div>
          <el-table :data="modelPricings" v-loading="ldPricing" stripe>
            <el-table-column prop="model" label="模型" min-width="120" />
            <el-table-column prop="provider" label="提供商" min-width="120" />
            <el-table-column prop="display_name" label="显示名称" min-width="180" />
            <el-table-column label="价格输入" width="120">
              <template #default="{ row }">{{ row.price_input }}</template>
            </el-table-column>
            <el-table-column label="价格输出" width="120">
              <template #default="{ row }">{{ row.price_output }}</template>
            </el-table-column>
            <el-table-column label="上下文长度" width="120">
              <template #default="{ row }">{{ row.context_length }}</template>
            </el-table-column>
            <el-table-column label="最大输出" width="100">
              <template #default="{ row }">{{ row.max_output }}</template>
            </el-table-column>
            <el-table-column label="分组" width="120">
              <template #default="{ row }">{{ row.group_id ?? '-' }}</template>
            </el-table-column>
            <el-table-column label="启用" width="90">
              <template #default="{ row }">
                <el-tag :type="row.is_enabled ? 'success' : 'info'" size="small">{{ row.is_enabled ? '是' : '否' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="排序" width="60">
              <template #default="{ row }">{{ row.sort_order }}</template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="openPricingDialog(true, row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deletePricing(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
        <!-- Pricing Dialog -->
        <el-dialog :title="pricingDialogMode ? '编辑定价' : '添加定价'" v-model:visible="pricingDialogVisible" width="800px">
          <el-form :model="pricingForm" label-width="120px" ref="pricingFormRef">
            <el-form-item label="模型" required>
              <el-input v-model="pricingForm.model" placeholder="如：gpt-4" />
            </el-form-item>
            <el-form-item label="提供商" required>
              <el-input v-model="pricingForm.provider" placeholder="provider" />
            </el-form-item>
            <el-form-item label="显示名称">
              <el-input v-model="pricingForm.display_name" placeholder="显示名称" />
            </el-form-item>
            <el-form-item label="价格输入" required>
              <el-input-number v-model="pricingForm.price_input" :min="0" :step="0.01" style="width: 180px" />
            </el-form-item>
            <el-form-item label="价格输出">
              <el-input-number v-model="pricingForm.price_output" :min="0" :step="0.01" style="width: 180px" />
            </el-form-item>
            <el-form-item label="上下文长度">
              <el-input-number v-model="pricingForm.context_length" :min="0" style="width:180px" />
            </el-form-item>
            <el-form-item label="最大输出">
              <el-input-number v-model="pricingForm.max_output" :min="0" style="width:180px" />
            </el-form-item>
            <el-form-item label="分组">
              <el-select v-model="pricingForm.group_id" placeholder="无分组时留空" style="width: 100%">
                <el-option v-for="g in modelGroupsAll" :key="g.id" :label="g.display_name || g.name" :value="g.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="启用">
              <el-switch v-model="pricingForm.is_enabled" true-value="true" false-value="false" />
            </el-form-item>
            <el-form-item label="是否特色">
              <el-switch v-model="pricingForm.is_featured" true-value="true" false-value="false" />
            </el-form-item>
            <el-form-item label="排序">
              <el-input-number v-model="pricingForm.sort_order" :min="0" style="width:180px" />
            </el-form-item>
            <el-form-item label="能力类型" prop="ability_types">
              <el-input v-model="pricingForm.ability_types" placeholder="如：text-davinci-003" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="pricingDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="savePricing" :loading="pricingSaving">保存</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="用户分组" name="userGroups">
        <el-card>
          <el-table :data="users" v-loading="ldUsers" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="username" label="用户名" min-width="140">
              <template #default="{ row }">{{ row.username }}</template>
            </el-table-column>
            <el-table-column prop="email" label="邮箱" min-width="200" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="openUserGroupDialog(row)">设置分组</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
        <!-- User Group Dialog -->
        <el-dialog title="设置用户分组" v-model:visible="userGroupDialogVisible" width="680px">
          <div style="max-height:320px;overflow:auto;">
            <el-checkbox-group v-model="selectedGroupIds" style="display:flex;flex-direction:column;gap:8px;">
              <el-checkbox v-for="g in modelGroupsAll" :key="g.id" :label="g.id">{{ g.display_name || g.name }}</el-checkbox>
            </el-checkbox-group>
          </div>
          <template #footer>
            <el-button @click="userGroupDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="saveUserGroups" :loading="userGroupSaving">保存</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { modelGroupAPI } from '@/api/model'
import type { ModelGroup, ModelPricing } from '@/api/model'
import { modelPricingAPI, userGroupAPI } from '@/api/model'
import { adminUserApi } from '@/api/user'
import { ElMessage } from 'element-plus'

// Tabs
const activeTab = ref('groups')

// Group (模型分组)
const modelGroups = ref<ModelGroup[]>([])
const ldGroups = ref(false)
const groupDialogVisible = ref(false)
const groupDialogMode = ref(false)
const groupSaving = ref(false)
const groupForm = reactive({ id: 0, name: '', display_name: '', description: '', sort_order: 0, status: 'active' as 'active'|'inactive' })
const groupFormRef = ref<any>(null)

async function loadGroups() {
  ldGroups.value = true
  try {
    const res = await modelGroupAPI.list({ page: 1, page_size: 9999 })
    modelGroups.value = res.data?.data?.list ?? []
  } catch (e) {
    ElMessage.error('加载模型分组失败')
  } finally {
    ldGroups.value = false
  }
}

function openGroupDialog(edit = false, row?: ModelGroup) {
  groupDialogMode.value = edit
  if (edit && row) {
    Object.assign(groupForm, {
      id: row.id,
      name: row.name,
      display_name: row.display_name,
      description: row.description,
      sort_order: row.sort_order,
      status: row.status as any
    })
  } else {
    Object.assign(groupForm, { id: 0, name: '', display_name: '', description: '', sort_order: 0, status: 'active' })
  }
  groupDialogVisible.value = true
}

async function saveGroup() {
  const payload = {
    display_name: groupForm.display_name,
    description: groupForm.description,
    sort_order: groupForm.sort_order,
    status: groupForm.status,
  }
  try {
    if (groupDialogMode.value && groupForm.id) {
      await modelGroupAPI.update(groupForm.id, payload)
    } else {
      await modelGroupAPI.create({ name: groupForm.name, ...payload })
    }
    ElMessage.success('保存成功')
    groupDialogVisible.value = false
    await loadGroups()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '保存失败')
  }
}

async function deleteGroup(row: ModelGroup) {
  try {
    await modelGroupAPI.delete(row.id)
    ElMessage.success('删除成功')
    await loadGroups()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '删除失败')
  }
}

// Pricing
const modelPricings = ref<ModelPricing[]>([])
const ldPricing = ref(false)
const pricingDialogVisible = ref(false)
const pricingDialogMode = ref(false)
const pricingForm = reactive({ id:0, model:'', provider:'', display_name:'', price_input:0, price_output:0, context_length:0, max_output:0, group_id:null as number|null, is_enabled:true, is_featured:false, sort_order:0, ability_types:'' })
const pricingFormRef = ref<any>(null)
const modelGroupsAll = ref<ModelGroup[]>([])

async function loadPricing() {
  ldPricing.value = true
  try {
    const res = await modelPricingAPI.list({ page:1, page_size:9999 })
    modelPricings.value = res.data?.data?.list ?? []
  } catch (e) {
    ElMessage.error('加载模型定价失败')
  } finally {
    ldPricing.value = false
  }
}

async function loadModelGroupsAll() {
  try {
    const res = await modelGroupAPI.listAll()
    modelGroupsAll.value = res.data?.data ?? []
  } catch {
    modelGroupsAll.value = []
  }
}

function openPricingDialog(edit = false, row?: ModelPricing) {
  pricingDialogMode.value = edit
  if (edit && row) {
    Object.assign(pricingForm, {
      id: row.id,
      model: row.model,
      provider: row.provider,
      display_name: row.display_name,
      price_input: row.price_input,
      price_output: row.price_output,
      context_length: row.context_length,
      max_output: row.max_output,
      group_id: row.group_id,
      is_enabled: row.is_enabled,
      is_featured: row.is_featured,
      sort_order: row.sort_order,
      ability_types: row.ability_types ?? ''
    })
  } else {
    Object.assign(pricingForm, { id:0, model:'', provider:'', display_name:'', price_input:0, price_output:0, context_length:0, max_output:0, group_id:null, is_enabled:true, is_featured:false, sort_order:0, ability_types:'' })
  }
  pricingDialogVisible.value = true
}

async function savePricing() {
  try {
    const payload: any = {
      model: pricingForm.model,
      provider: pricingForm.provider,
      display_name: pricingForm.display_name,
      price_input: pricingForm.price_input,
      price_output: pricingForm.price_output,
      context_length: pricingForm.context_length,
      max_output: pricingForm.max_output,
      group_id: pricingForm.group_id,
      is_enabled: pricingForm.is_enabled,
      is_featured: pricingForm.is_featured,
      sort_order: pricingForm.sort_order,
      ability_types: pricingForm.ability_types,
    }
    if (pricingForm.id) {
      await modelPricingAPI.update(pricingForm.id, payload)
    } else {
      await modelPricingAPI.create(payload)
    }
    pricingDialogVisible.value = false
    await loadPricing()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '保存失败')
  }
}

async function deletePricing(row: ModelPricing) {
  try {
    await modelPricingAPI.delete(row.id)
    ElMessage.success('删除成功')
    await loadPricing()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '删除失败')
  }
}

// Users (用户分组)
const users = ref<any[]>([])
const ldUsers = ref(false)
const userGroupDialogVisible = ref(false)
const selectedUser = ref<any | null>(null)
const selectedGroupIds = ref<number[]>([])
const userGroupSaving = ref(false)

async function loadUsers() {
  ldUsers.value = true
  try {
    const res = await adminUserApi.listUsers({ page: 1, page_size: 50 })
    users.value = res.data?.data?.list ?? []
  } catch {
    users.value = []
  } finally {
    ldUsers.value = false
  }
}

async function openUserGroupDialog(u: any) {
  selectedUser.value = u
  // Load groups for assignment
  await loadModelGroupsAll()
  // Load existing groups for user
  try {
    const res = await userGroupAPI.getUserGroups(u.id)
    selectedGroupIds.value = (res.data?.data ?? []).map((g: any) => g.group_id)
  } catch {
    selectedGroupIds.value = []
  }
  userGroupDialogVisible.value = true
}

async function saveUserGroups() {
  if (!selectedUser.value) return
  userGroupSaving.value = true
  try {
    await userGroupAPI.setUserGroups(selectedUser.value.id, selectedGroupIds.value)
    userGroupDialogVisible.value = false
    await loadUsers()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error?.message || '设置分组失败')
  } finally {
    userGroupSaving.value = false
  }
}

onMounted(async () => {
  await loadGroups()
  await loadPricing()
  await loadModelGroupsAll()
  await loadUsers()
})

</script>

<style scoped>
.admin-models { padding: 8px 0; }
.page-header { margin-bottom: 12px; }
.subtitle { color: #909399; font-size: 12px; }
</style>
