<template>
  <div class="sync-task-page">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <el-icon class="title-icon"><RefreshRight /></el-icon>
          同步任务管理
        </h1>
        <p class="page-desc">管理多仓库分支同步任务，支持定时同步和手动触发</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="openCreateModal">
          <el-icon><Plus /></el-icon>
          新建任务
        </el-button>
      </div>
    </div>

    <SyncStatsCards :stats="stats" />

    <div class="toolbar">
      <div class="toolbar-left">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索任务名称、仓库..."
          clearable
          style="width: 280px"
          :prefix-icon="Search"
        />
        <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 140px" clearable>
          <el-option label="已启用" value="enabled" />
          <el-option label="已禁用" value="disabled" />
        </el-select>
      </div>
      <div class="toolbar-right">
        <el-button @click="loadTasks">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <SyncTaskTable
      :tasks="filteredTasks"
      :loading="loading"
      :running-keys="runningKeys"
      @run="run_task"
      @edit="editTask"
      @delete="delete_task"
      @create="openCreateModal"
    />

    <SyncTaskFormDialog
      v-model:visible="modalVisible"
      :is-edit="isEditMode"
      :form="taskForm"
      :submitting="submitting"
      @submit="submitTask"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { toastApiError } from '@/composables/useNotification'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, RefreshRight } from '@element-plus/icons-vue'
import { syncV2Api } from '@/api/modules/sync_v2'
import type { SyncTask, SyncStats } from '@/types/sync_v2'
import SyncStatsCards from '@/components/sync/SyncStatsCards.vue'
import SyncTaskTable from '@/components/sync/SyncTaskTable.vue'
import SyncTaskFormDialog from '@/components/sync/SyncTaskFormDialog.vue'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const isEditMode = ref(false)
const searchKeyword = ref('')
const statusFilter = ref<string | undefined>()
const runningKeys = ref<Set<string>>(new Set())

const stats = reactive<SyncStats>({
  total_tasks: 0,
  enabled_tasks: 0,
  today_runs: 0,
  failed_runs: 0,
  running_tasks: 0,
})

const tasks = ref<SyncTask[]>([])

const filteredTasks = computed(() => {
  let result = [...tasks.value]
  if (searchKeyword.value) {
    result = result.filter(t => t.name.toLowerCase().includes(searchKeyword.value.toLowerCase()))
  }
  if (statusFilter.value) {
    result = result.filter(t => statusFilter.value === 'enabled' ? t.enabled : !t.enabled)
  }
  return result
})

const taskForm = reactive<Partial<SyncTask>>({
  name: '',
  source_repo_key: '',
  source_branch: '',
  target_repo_key: '',
  target_branch: '',
  sync_mode: 'single',
  cron: '',
  enabled: true,
  git_tags: false,
  git_force: false,
  git_prune: false,
  git_no_verify: false,
  push_options: '',
})

async function loadStats() {
  try {
    const data = await syncV2Api.get_stats()
    Object.assign(stats, data)
  } catch (e) {
    console.error('加载统计数据失败', e)
  }
}

async function loadTasks() {
  loading.value = true
  try {
    const data = await syncV2Api.list_tasks()
    tasks.value = data || []
  } catch (e) { toastApiError(e, '加载任务列表失败') } finally {
    loading.value = false
  }
}

function openCreateModal() {
  isEditMode.value = false
  Object.assign(taskForm, {
    name: '',
    source_repo_key: '',
    source_branch: '',
    target_repo_key: '',
    target_branch: '',
    sync_mode: 'single',
    cron: '',
    enabled: true,
    git_tags: false,
    git_force: false,
    git_prune: false,
    git_no_verify: false,
    push_options: '',
    key: undefined,
  })
  modalVisible.value = true
}

function editTask(row: SyncTask) {
  isEditMode.value = true
  Object.assign(taskForm, { ...row })
  modalVisible.value = true
}

async function submitTask() {
  submitting.value = true
  try {
    if (isEditMode.value) {
      await syncV2Api.update_task(taskForm as SyncTask)
      ElMessage.success('任务更新成功')
    } else {
      await syncV2Api.create_task(taskForm as SyncTask)
      ElMessage.success('任务创建成功')
    }
    modalVisible.value = false
    loadTasks()
    loadStats()
  } catch (e) {
    toastApiError(e, isEditMode.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

async function run_task(row: SyncTask) {
  runningKeys.value.add(row.key)
  try {
    await syncV2Api.run_task(row.key)
    ElMessage.success('任务已触发执行')
  } catch (e) { toastApiError(e, '任务触发失败') } finally {
    runningKeys.value.delete(row.key)
  }
}

async function delete_task(row: SyncTask) {
  try {
    await ElMessageBox.confirm(`确定要删除任务「${row.name}」吗？`, '确认删除', {
      type: 'warning',
    })
    await syncV2Api.delete_task(row.key)
    ElMessage.success('删除成功')
    loadTasks()
    loadStats()
  } catch (e) {
    if (e !== 'cancel') {
      toastApiError(e, '删除失败')
    }
  }
}

onMounted(() => {
  loadStats()
  loadTasks()
})
</script>

<style scoped lang="less">
.sync-task-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  .header-left {
    .page-title {
      display: flex;
      align-items: center;
      font-size: 24px;
      font-weight: 600;
      color: #1d2129;
      margin: 0 0 8px 0;

      .title-icon {
        margin-right: 12px;
        color: #409eff;
      }
    }

    .page-desc {
      margin: 0;
      color: #909399;
      font-size: 14px;
    }
  }
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .toolbar-left {
    display: flex;
    gap: 12px;
    align-items: center;
  }
}
</style>
