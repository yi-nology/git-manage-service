<template>
  <div class="webhook-rules">
    <PageHeader title="Webhook 规则管理">
      <template #actions>
        <el-button type="primary" @click="openCreate" :icon="Plus">新建规则</el-button>
      </template>
    </PageHeader>

    <el-card class="filter-card">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input v-model="filters.keyword" placeholder="搜索规则名称" clearable :prefix-icon="Search" />
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="已启用" value="enabled" />
            <el-option label="已禁用" value="disabled" />
          </el-select>
        </el-col>
        <el-col :span="14" style="text-align: right">
          <el-button @click="resetFilters" :icon="Refresh">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="filteredRules" v-loading="loading" class="rules-table">
      <el-table-column prop="name" label="规则名称" min-width="180">
        <template #default="{ row }">
          <div class="name-cell">
            <el-tag v-if="row.enabled" type="success" size="small" effect="plain">启用</el-tag>
            <el-tag v-else size="small" effect="plain">禁用</el-tag>
            <span class="name">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="Git 平台" width="160">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ providerName(row.provider_config_id) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="触发事件" min-width="200">
        <template #default="{ row }">
          <div class="event-tags">
            <el-tag v-for="evt in splitEvents(row.event_type_pattern)" :key="evt" size="small" effect="plain">{{ evt }}</el-tag>
            <span v-if="!row.event_type_pattern" class="text-muted">全部事件</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="repo_pattern" label="仓库过滤" width="200">
        <template #default="{ row }">
          <span v-if="row.repo_pattern" class="filter-text">{{ row.repo_pattern }}</span>
          <span v-else class="text-muted">全部仓库</span>
        </template>
      </el-table-column>
      <el-table-column label="动作" width="120">
        <template #default="{ row }">
          <el-tag size="small" :type="actionTagType(row.action)">{{ actionLabel(row.action) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          <span v-if="row.created_at">{{ row.created_at }}</span>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" :icon="Edit" @click="editRule(row)">编辑</el-button>
          <el-button size="small" type="danger" :icon="Delete" @click="removeRule(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <span class="text-muted">共 {{ filteredRules.length }} 条</span>
    </div>

    <el-dialog v-model="showCreateDialog" :title="editingId ? '编辑规则' : '新建规则'" width="700px" destroy-on-close>
      <el-form :model="ruleForm" label-width="120px">
        <el-form-item label="规则名称" required>
          <el-input v-model="ruleForm.name" placeholder="输入规则名称" />
        </el-form-item>

        <el-form-item label="Git 平台" required>
          <el-select v-model="ruleForm.provider_config_id" placeholder="选择已配置的 Git 平台" style="width: 100%">
            <el-option v-for="p in providers" :key="p.id" :label="`${p.name} (${p.platform})`" :value="p.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="触发事件">
          <el-checkbox-group v-model="ruleForm.eventTypes">
            <el-checkbox label="push">Push</el-checkbox>
            <el-checkbox label="create">分支创建</el-checkbox>
            <el-checkbox label="delete">分支删除</el-checkbox>
            <el-checkbox label="pull_request">PR/MR</el-checkbox>
            <el-checkbox label="tag">Tag</el-checkbox>
          </el-checkbox-group>
          <div class="form-tip">留空表示匹配所有事件，多选为「或」关系（后端按逗号分隔匹配，支持通配符 *）</div>
        </el-form-item>

        <el-form-item label="仓库过滤">
          <el-input v-model="ruleForm.repo_pattern" placeholder="owner/repo 格式，支持通配符，如 frontend/*" />
          <div class="form-tip">留空表示匹配所有仓库</div>
        </el-form-item>

        <el-form-item label="触发动作" required>
          <el-select v-model="ruleForm.action" placeholder="选择触发动作" style="width: 100%">
            <el-option label="代码审查" value="code_review" />
            <el-option label="触发同步任务" value="sync" />
            <el-option label="发送通知" value="notify" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="ruleForm.action === 'sync'" label="关联同步任务">
          <el-select v-model="ruleForm.syncTaskKey" placeholder="选择同步任务" style="width: 100%" filterable>
            <el-option v-for="t in syncTasks" :key="t.key" :label="t.name" :value="t.key" />
          </el-select>
          <div v-if="!syncTasks.length" class="form-tip">暂无同步任务，请先在「同步管理」创建</div>
        </el-form-item>

        <el-form-item label="启用规则">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { Plus, Search, Refresh, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '@/components/common/PageHeader.vue'
import {
  listWebhookRules,
  createWebhookRule,
  updateWebhookRule,
  deleteWebhookRule,
  type WebhookRule,
} from '@/api/modules/webhook-rule'
import { listProviders, type ProviderConfigDTO } from '@/api/modules/provider'
import { getSyncTasks } from '@/api/modules/sync_v2'

const loading = ref(false)
const saving = ref(false)
const showCreateDialog = ref(false)
const editingId = ref<number | null>(null)

const rules = ref<WebhookRule[]>([])
const providers = ref<ProviderConfigDTO[]>([])
const syncTasks = ref<{ key: string; name: string }[]>([])

const filters = reactive({ keyword: '', status: '' })

const filteredRules = computed(() =>
  rules.value.filter((r) => {
    if (filters.keyword && !r.name.toLowerCase().includes(filters.keyword.toLowerCase())) return false
    if (filters.status === 'enabled' && !r.enabled) return false
    if (filters.status === 'disabled' && r.enabled) return false
    return true
  })
)

function resetFilters() {
  filters.keyword = ''
  filters.status = ''
}

const ruleForm = reactive({
  name: '',
  provider_config_id: 0 as number,
  eventTypes: [] as string[],
  repo_pattern: '',
  action: 'code_review',
  syncTaskKey: '',
  enabled: true,
})

function resetForm() {
  ruleForm.name = ''
  ruleForm.provider_config_id = providers.value[0]?.id ?? 0
  ruleForm.eventTypes = []
  ruleForm.repo_pattern = ''
  ruleForm.action = 'code_review'
  ruleForm.syncTaskKey = ''
  ruleForm.enabled = true
  editingId.value = null
}

function openCreate() {
  resetForm()
  showCreateDialog.value = true
}

function editRule(row: WebhookRule) {
  editingId.value = row.id
  ruleForm.name = row.name
  ruleForm.provider_config_id = row.provider_config_id
  ruleForm.eventTypes = row.event_type_pattern
    ? row.event_type_pattern.split(',').map((s) => s.trim()).filter(Boolean)
    : []
  ruleForm.repo_pattern = row.repo_pattern
  ruleForm.action = row.action
  ruleForm.syncTaskKey = (row.action_config?.sync_task_key as string) || ''
  ruleForm.enabled = row.enabled
  showCreateDialog.value = true
}

async function loadRules() {
  loading.value = true
  try {
    const res = await listWebhookRules()
    rules.value = res.items || []
  } catch {
    /* toast handled by interceptor */
  } finally {
    loading.value = false
  }
}

async function saveRule() {
  if (!ruleForm.name) {
    ElMessage.warning('请输入规则名称')
    return
  }
  if (!ruleForm.provider_config_id) {
    ElMessage.warning('请选择 Git 平台')
    return
  }
  const payload = {
    name: ruleForm.name,
    provider_config_id: ruleForm.provider_config_id,
    event_type_pattern: ruleForm.eventTypes.join(','),
    repo_pattern: ruleForm.repo_pattern,
    action: ruleForm.action,
    action_config: ruleForm.action === 'sync' && ruleForm.syncTaskKey ? { sync_task_key: ruleForm.syncTaskKey } : {},
    enabled: ruleForm.enabled,
  }
  saving.value = true
  try {
    if (editingId.value) {
      await updateWebhookRule(editingId.value, payload)
      ElMessage.success('规则已更新')
    } else {
      await createWebhookRule(payload)
      ElMessage.success('规则已创建')
    }
    showCreateDialog.value = false
    await loadRules()
  } catch {
    /* toast handled by interceptor */
  } finally {
    saving.value = false
  }
}

async function removeRule(row: WebhookRule) {
  try {
    await ElMessageBox.confirm(`确认删除规则「${row.name}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteWebhookRule(row.id)
    ElMessage.success('规则已删除')
    await loadRules()
  } catch {
    /* toast handled by interceptor */
  }
}

function providerName(id: number) {
  const p = providers.value.find((x) => x.id === id)
  return p ? `${p.name}` : id ? `#${id}` : '-'
}

function splitEvents(pattern: string) {
  return pattern ? pattern.split(',').map((s) => s.trim()).filter(Boolean) : []
}

function actionLabel(action: string) {
  switch (action) {
    case 'sync':
      return '同步'
    case 'notify':
      return '通知'
    case 'code_review':
      return '代码审查'
    default:
      return action
  }
}

function actionTagType(action: string) {
  switch (action) {
    case 'sync':
      return 'success'
    case 'notify':
      return 'warning'
    case 'code_review':
      return 'primary'
    default:
      return 'info'
  }
}

onMounted(async () => {
  await Promise.all([
    listProviders().then((res) => (providers.value = res || [])),
    getSyncTasks().then((res) => (syncTasks.value = (res || []).map((t) => ({ key: t.key, name: t.name })))),
  ])
  await loadRules()
})
</script>

<style scoped lang="scss">
.webhook-rules {
  .filter-card {
    margin-bottom: 16px;
  }

  .rules-table {
    .name-cell {
      display: flex;
      align-items: center;
      gap: 8px;

      .name {
        font-weight: 500;
      }
    }

    .event-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }

    .filter-text {
      font-family: 'SF Mono', Monaco, monospace;
      font-size: 12px;
    }

    .text-muted {
      color: var(--text-color-placeholder);
    }
  }

  .form-tip {
    font-size: 12px;
    color: var(--text-color-placeholder);
    margin-top: 4px;
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
