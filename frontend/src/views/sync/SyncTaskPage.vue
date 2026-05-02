<template>
  <div class="sync-page">
    <PageHeader title="同步任务" showBack :backRoute="`/local-repos/${repoKey}`">
      <template #actions>
        <ActionPill variant="green" :icon="Refresh" @click="handleBatchSync" :disabled="selectedTasks.length === 0">
          同步选中 ({{ selectedTasks.length }})
        </ActionPill>
        <ActionPill variant="primary" :icon="Plus" @click="showQuickPanel = !showQuickPanel">快速同步</ActionPill>
        <ActionPill variant="outline" :icon="Setting" @click="openAddTask">新建规则</ActionPill>
      </template>
    </PageHeader>

    <QuickSyncPanel v-model="showQuickPanel" :repo-key="repoKey" :remote-names="remoteNames" />

    <SectionTitle title="同步任务列表" />
    <div v-loading="loading" class="task-list">
      <el-empty v-if="tasks.length === 0 && !loading" description="暂无同步规则">
        <el-button type="primary" @click="openAddTask">创建第一条规则</el-button>
      </el-empty>

      <el-card v-for="task in tasks" :key="task.key" class="task-card" :class="{ disabled: !task.enabled }">
        <div class="task-content">
          <el-checkbox v-model="selectedTasks" :value="task.key" class="task-checkbox" />

          <div class="direction-flow">
            <div class="endpoint source">
              <span class="label">{{ task.source_remote }}</span>
              <span class="branch">{{ task.source_branch }}</span>
            </div>
            <div class="flow-arrow">
              <el-icon><Right /></el-icon>
              <el-tag v-if="task.sync_mode === 'all-branch'" size="small" type="info">全分支</el-tag>
            </div>
            <div class="endpoint target">
              <span class="label">{{ task.target_remote }}</span>
              <span class="branch">{{ task.target_branch }}</span>
            </div>
          </div>

          <div class="task-status">
            <el-tag :type="task.enabled ? 'success' : 'info'" size="small">
              {{ task.enabled ? '✅ 已启用' : '⏸️ 已暂停' }}
            </el-tag>
            <span v-if="task.cron" class="cron">
              <el-icon><AlarmClock /></el-icon> {{ task.cron }}
            </span>
          </div>

          <div class="git-options">
            <el-tag v-if="task.git_tags" size="small" effect="plain">--tags</el-tag>
            <el-tag v-if="task.git_force" size="small" type="warning" effect="plain">--force</el-tag>
            <el-tag v-if="task.git_prune" size="small" effect="plain">--prune</el-tag>
            <el-tag v-if="task.git_no_verify" size="small" effect="plain">--no-verify</el-tag>
          </div>

          <div class="task-actions">
            <el-button size="small" type="success" @click="handleRun(task.key)" :icon="CaretRight" round>执行</el-button>
            <el-button size="small" @click="showHistory(task.key)" :icon="Clock" round>历史</el-button>
            <el-dropdown trigger="click">
              <el-button size="small" :icon="More" round />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="openEditTask(task)">编辑</el-dropdown-item>
                  <el-dropdown-item @click="toggleEnabled(task)">
                    {{ task.enabled ? '暂停' : '启用' }}
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="handleDelete(task.key)" style="color: #f56c6c">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-card>
    </div>

    <el-dialog v-model="showTaskDialog" :title="editingTask ? '编辑同步规则' : '新建同步规则'" width="700px" destroy-on-close>
      <el-form :model="taskForm" label-width="100px">
        <el-form-item label="同步模式">
          <el-radio-group v-model="taskForm.sync_mode">
            <el-radio value="single">单分支同步</el-radio>
            <el-radio value="all-branch">全分支同步</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="源 (Source)">
              <el-select v-model="taskForm.source_remote" style="width: 100%" @change="onSourceRemoteChange">
                <el-option label="Local (本地)" value="local" />
                <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="taskForm.sync_mode !== 'all-branch'" label="源分支">
              <el-select
                v-model="taskForm.source_branch"
                filterable
                style="width: 100%"
                placeholder="选择源分支"
                :loading="branchLoading"
              >
                <el-option v-for="b in sourceBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="目标 (Target)">
              <el-select v-model="taskForm.target_remote" style="width: 100%">
                <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="taskForm.sync_mode !== 'all-branch'" label="目标分支">
              <el-select
                v-model="taskForm.target_branch"
                filterable
                allow-create
                default-first-option
                style="width: 100%"
                placeholder="选择或输入目标分支（回车新建）"
                :loading="branchLoading"
              >
                <el-option v-for="b in targetBranches" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-alert v-if="taskForm.sync_mode === 'all-branch'" title="全分支模式将自动同步源端所有分支到目标端对应分支" type="info" :closable="false" show-icon class="mb-3" />
        <el-alert v-if="taskForm.source_remote === taskForm.target_remote" title="源和目标不能相同" type="warning" :closable="false" show-icon class="mb-3" />

        <el-divider content-position="left">Git 选项</el-divider>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item>
              <el-checkbox v-model="taskForm.git_tags">--tags 推送所有标签</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="taskForm.git_prune">--prune 清理已删除分支</el-checkbox>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item>
              <el-checkbox v-model="taskForm.git_force">--force 强制推送 ⚠️</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="taskForm.git_no_verify">--no-verify 跳过钩子</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">定时任务</el-divider>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Cron">
              <el-input v-model="taskForm.cron" placeholder="0 2 * * * (留空禁用)" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用">
              <el-switch v-model="taskForm.enabled" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="showTaskDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTask" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showHistoryDialog" title="同步历史" width="900px">
      <el-table :data="historyList" size="small" border>
        <el-table-column prop="start_time" label="时间" width="160">
          <template #default="{ row }">{{ formatDate(row.start_time) }}</template>
        </el-table-column>
        <el-table-column prop="trigger_source" label="触发" width="100">
          <template #default="{ row }">
            <el-tag :type="getTriggerTagType(row.trigger_source)" size="small">{{ getTriggerLabel(row.trigger_source) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusColor(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            {{ row.end_time ? (new Date(row.end_time).getTime() - new Date(row.start_time).getTime()) + 'ms' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="详情">
          <template #default="{ row }">
            <el-button size="small" link @click="showLog(row.details)">日志</el-button>
            <span v-if="row.error_message" class="error-msg">{{ row.error_message }}</span>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="showLogDialog" title="执行详情" width="700px">
      <pre class="log-content">{{ logContent }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Setting, Refresh, Right, CaretRight, Clock,
  AlarmClock, More
} from '@element-plus/icons-vue'
import {
  getSyncTasks, createSyncTask, updateSyncTask, deleteSyncTask,
  runSyncTask, getSyncHistory, batchSync
} from '@/api/modules/sync'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getBranchList } from '@/api/modules/branch'
import type { BranchInfo } from '@/types/branch'
import type { SyncTaskDTO, SyncRunDTO } from '@/types/sync'
import { formatDate, getStatusColor } from '@/utils/format'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import QuickSyncPanel from '@/components/sync/QuickSyncPanel.vue'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const saving = ref(false)
const branchLoading = ref(false)
const tasks = ref<SyncTaskDTO[]>([])
const remoteNames = ref<string[]>([])
const allBranches = ref<BranchInfo[]>([])
const selectedTasks = ref<string[]>([])
const showQuickPanel = ref(false)

const sourceBranches = computed(() => {
  const remote = taskForm.value.source_remote
  if (remote === 'local') {
    return allBranches.value.filter(b => b.type === 'local').map(b => b.name)
  }
  const prefix = remote + '/'
  return allBranches.value
    .filter(b => b.type === 'remote' && b.name.startsWith(prefix))
    .map(b => b.name.slice(prefix.length))
})

const targetBranches = computed(() => {
  return allBranches.value.filter(b => b.type === 'local').map(b => b.name)
})

const showTaskDialog = ref(false)
const editingTask = ref<SyncTaskDTO | null>(null)
const taskForm = ref({
  source_remote: 'local',
  source_branch: 'main',
  target_remote: '',
  target_branch: 'main',
  cron: '',
  enabled: true,
  sync_mode: 'single',
  git_tags: false,
  git_force: false,
  git_prune: false,
  git_no_verify: false,
})

const showHistoryDialog = ref(false)
const historyList = ref<SyncRunDTO[]>([])
const showLogDialog = ref(false)
const logContent = ref('')

function getTriggerTagType(source: string) {
  switch (source) {
    case 'cron': return 'warning'
    case 'webhook': return 'success'
    case 'manual': return 'primary'
    default: return 'info'
  }
}

function getTriggerLabel(source: string) {
  switch (source) {
    case 'cron': return '定时'
    case 'webhook': return 'Webhook'
    case 'manual': return '手动'
    default: return source || '手动'
  }
}

function onSourceRemoteChange() {
  taskForm.value.source_branch = ''
}

onMounted(async () => {
  await loadTasks()
  try {
    const repo = await getRepoDetail(repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      remoteNames.value = (scan.remotes || []).map((r) => r.name)
      taskForm.value.target_remote = remoteNames.value[0] || ''
    }
  } catch { /* ignore */ }
  branchLoading.value = true
  try {
    const result = await getBranchList(repoKey, { page_size: 500 })
    allBranches.value = result?.list || []
  } catch { /* ignore */ } finally {
    branchLoading.value = false
  }
})

async function loadTasks() {
  loading.value = true
  try {
    tasks.value = (await getSyncTasks(repoKey)) || []
  } finally {
    loading.value = false
  }
}

function openAddTask() {
  editingTask.value = null
  taskForm.value = {
    source_remote: 'local',
    source_branch: 'main',
    target_remote: remoteNames.value[0] || '',
    target_branch: 'main',
    cron: '',
    enabled: true,
    sync_mode: 'single',
    git_tags: false,
    git_force: false,
    git_prune: false,
    git_no_verify: false,
  }
  showTaskDialog.value = true
}

function openEditTask(task: SyncTaskDTO) {
  editingTask.value = task
  taskForm.value = {
    source_remote: task.source_remote,
    source_branch: task.source_branch,
    target_remote: task.target_remote,
    target_branch: task.target_branch,
    cron: task.cron,
    enabled: task.enabled,
    sync_mode: task.sync_mode || 'single',
    git_tags: task.git_tags,
    git_force: task.git_force,
    git_prune: task.git_prune,
    git_no_verify: task.git_no_verify,
  }
  showTaskDialog.value = true
}

async function handleSaveTask() {
  if (taskForm.value.source_remote === taskForm.value.target_remote) {
    ElMessage.warning('源和目标不能相同')
    return
  }
  saving.value = true
  try {
    if (editingTask.value) {
      await updateSyncTask({
        key: editingTask.value.key,
        source_repo_key: repoKey,
        target_repo_key: repoKey,
        ...taskForm.value,
      })
    } else {
      await createSyncTask({
        source_repo_key: repoKey,
        target_repo_key: repoKey,
        ...taskForm.value,
      })
    }
    ElMessage.success('保存成功')
    showTaskDialog.value = false
    await loadTasks()
  } finally {
    saving.value = false
  }
}

async function handleRun(key: string) {
  try {
    await runSyncTask(key)
    ElMessage.success('任务已触发')
  } catch { /* handled */ }
}

async function handleBatchSync() {
  if (selectedTasks.value.length === 0) return

  try {
    await ElMessageBox.confirm(`确定同步 ${selectedTasks.value.length} 个规则？`, '批量同步', { type: 'info' })
    await batchSync(selectedTasks.value)
    ElMessage.success('批量同步已触发')
    selectedTasks.value = []
  } catch { /* cancelled */ }
}

async function toggleEnabled(task: SyncTaskDTO) {
  try {
    await updateSyncTask({
      key: task.key,
      source_repo_key: repoKey,
      target_repo_key: repoKey,
      source_remote: task.source_remote,
      source_branch: task.source_branch,
      target_remote: task.target_remote,
      target_branch: task.target_branch,
      enabled: !task.enabled,
    })
    ElMessage.success(task.enabled ? '已暂停' : '已启用')
    await loadTasks()
  } catch { /* handled */ }
}

async function handleDelete(key: string) {
  try {
    await ElMessageBox.confirm('确定删除该同步规则吗？', '确认删除', { type: 'warning' })
    await deleteSyncTask(key)
    ElMessage.success('删除成功')
    await loadTasks()
  } catch { /* cancelled */ }
}

async function showHistory(taskKey: string) {
  try {
    const all = await getSyncHistory()
    historyList.value = all.filter((h) => h.task_key === taskKey)
    showHistoryDialog.value = true
  } catch { /* handled */ }
}

function showLog(details: string) {
  logContent.value = details || '无详情'
  showLogDialog.value = true
}
</script>

<style scoped>
.sync-page {
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: 100vh;
  background: var(--bg-color);
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-card {
  transition: all var(--transition-normal);
  border-left: 4px solid var(--primary-color);
  border-radius: var(--border-radius-md);
}

.task-card.disabled {
  opacity: 0.6;
  border-left-color: var(--text-color-secondary);
}

.task-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm) 0;
}

.task-checkbox {
  flex-shrink: 0;
}

.direction-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.endpoint {
  display: flex;
  flex-direction: column;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-sm);
  min-width: 100px;
}

.endpoint.source {
  background: #ECFDF5;
  border: 1px solid #A7F3D0;
}

.endpoint.target {
  background: #FFFBEB;
  border: 1px solid #FDE68A;
}

.endpoint .label {
  font-weight: 600;
  font-size: var(--font-size-md);
  color: var(--text-color-primary);
}

.endpoint .branch {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  font-family: monospace;
}

.flow-arrow {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: 18px;
  color: var(--primary-color);
}

.task-status {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  align-items: flex-end;
}

.cron {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
}

.git-options {
  display: flex;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}

.task-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.mb-3 {
  margin-bottom: 12px;
}

.log-content {
  background: var(--bg-color);
  padding: 12px;
  border-radius: var(--border-radius-sm);
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  font-size: var(--font-size-sm);
  font-family: monospace;
}

.error-msg {
  color: var(--danger-color);
  font-size: var(--font-size-xs);
  margin-left: var(--spacing-sm);
}

@media (max-width: 768px) {
  .sync-page {
    padding: var(--spacing-md);
  }

  .task-content {
    flex-wrap: wrap;
  }
}
</style>
