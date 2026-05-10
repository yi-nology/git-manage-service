<template>
  <div class="event-rules-page">
    <PageHeader title="Webhook 事件规则" subtitle="配置 Webhook 事件的处理规则和过滤策略" show-back back-route="/settings" />

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'rules' }" @click="activeTab = 'rules'">事件规则</button>
      <button class="tab-btn" :class="{ active: activeTab === 'logs' }" @click="activeTab = 'logs'">审查日志</button>
    </div>

    <div v-show="activeTab === 'rules'" class="tab-content">
      <div class="section-header">
        <div class="section-info">
          <h3>规则列表</h3>
          <p class="section-desc">按优先级顺序匹配 Webhook 事件，触发代码审查</p>
        </div>
        <ActionPill variant="primary" :icon="Plus" @click="openAddDialog">添加规则</ActionPill>
      </div>

      <DataTable :columns="columns" :data="rules" :loading="loading" row-key="id">
        <template #cell-name="{ row }">
          <span style="font-weight:500">{{ row.name }}</span>
        </template>
        <template #cell-event_type="{ row }">
          <StatusBadge :variant="eventTypeVariant(row.event_type)" :text="row.event_type" />
        </template>
        <template #cell-is_active="{ row }">
          <StatusBadge :variant="row.is_active ? 'success' : 'default'" :text="row.is_active ? '启用' : '禁用'" />
        </template>
        <template #cell-priority="{ row }">
          <span>{{ row.priority }}</span>
        </template>
        <template #row-actions="{ row }">
          <button class="act-btn act-btn--primary" @click="openEditDialog(row)">编辑</button>
          <button class="act-btn act-btn--danger" @click="handleDelete(row)">删除</button>
        </template>
      </DataTable>
    </div>

    <div v-show="activeTab === 'logs'" class="tab-content">
      <ReviewAuditLogTab />
    </div>

    <el-dialog v-model="showDialog" :title="editing ? '编辑事件规则' : '添加事件规则'" width="520px" destroy-on-close @close="resetForm">
      <el-form label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="例如: 自动审查 MR" />
        </el-form-item>
        <el-form-item label="事件类型">
          <el-select v-model="form.event_type" style="width:100%">
            <el-option label="Merge Request" value="merge_request" />
            <el-option label="Push" value="push" />
            <el-option label="Tag Push" value="tag_push" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="规则描述" />
        </el-form-item>
        <el-form-item label="匹配规则">
          <el-input v-model="form.match_rules" type="textarea" :rows="4" placeholder='JSON 格式匹配条件，例如: {"target_branch": "^main$"}' />
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="form.priority" :min="0" :max="100" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import { listEventRules, createEventRule, updateEventRule, deleteEventRule } from '@/api/modules/review'
import type { WebhookEventRuleDTO } from '@/api/modules/review'
import type { TableColumn } from '@/components/common/DataTable.vue'
import ReviewAuditLogTab from './ReviewAuditLogTab.vue'

const activeTab = ref('rules')
const loading = ref(false)
const rules = ref<WebhookEventRuleDTO[]>([])
const showDialog = ref(false)
const saving = ref(false)
const editing = ref<WebhookEventRuleDTO | null>(null)

const form = ref({
  name: '',
  event_type: 'merge_request',
  description: '',
  match_rules: '',
  priority: 0,
  is_active: true,
})

const columns: TableColumn[] = [
  { key: 'name', label: '名称', width: '160px' },
  { key: 'event_type', label: '事件类型', width: '120px' },
  { key: 'priority', label: '优先级', width: '80px' },
  { key: 'is_active', label: '状态', width: '80px' },
  { key: 'description', label: '描述', flex: 1 },
]

function eventTypeVariant(type: string): string {
  const map: Record<string, string> = {
    merge_request: 'purple',
    push: 'blue',
    tag_push: 'teal',
  }
  return map[type] || 'default'
}

async function loadRules() {
  loading.value = true
  try {
    const res = await listEventRules() as any
    rules.value = res.rules || (Array.isArray(res) ? res : [])
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openAddDialog() {
  editing.value = null
  form.value = { name: '', event_type: 'merge_request', description: '', match_rules: '', priority: 0, is_active: true }
  showDialog.value = true
}

function openEditDialog(row: WebhookEventRuleDTO) {
  editing.value = row
  form.value = {
    name: row.name,
    event_type: row.event_type,
    description: row.description || '',
    match_rules: row.match_rules || '',
    priority: row.priority || 0,
    is_active: row.is_active,
  }
  showDialog.value = true
}

function resetForm() {
  editing.value = null
}

async function handleSave() {
  if (!form.value.name || !form.value.event_type || !form.value.match_rules) {
    ElMessage.warning('请填写名称、事件类型和匹配规则')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateEventRule(editing.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await createEventRule(form.value)
      ElMessage.success('添加成功')
    }
    showDialog.value = false
    await loadRules()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: WebhookEventRuleDTO) {
  try {
    await ElMessageBox.confirm(`确定删除规则 "${row.name}"？`, '确认删除', { type: 'warning' })
  } catch { return }
  try {
    await deleteEventRule(row.id)
    ElMessage.success('已删除')
    await loadRules()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(() => { loadRules() })
</script>

<style scoped>
.event-rules-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.tab-bar {
  display: flex;
  gap: 4px;
  background: var(--surface-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  padding: 4px;
  width: fit-content;
}

.tab-btn {
  padding: 8px 20px;
  border: none;
  background: transparent;
  color: var(--text-color-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: var(--text-color-primary);
  background: var(--border-color-extra-light);
}

.tab-btn.active {
  background: var(--primary-color);
  color: #fff;
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.section-info h3 {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-color-primary);
}

.section-desc {
  margin: 4px 0 0;
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
}
</style>
