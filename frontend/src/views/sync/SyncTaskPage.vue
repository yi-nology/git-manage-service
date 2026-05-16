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

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon total">
          <el-icon :size="24"><List /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalTasks }}</div>
          <div class="stat-label">总任务数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon running">
          <el-icon :size="24"><VideoPlay /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.runningTasks }}</div>
          <div class="stat-label">运行中</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon :size="24"><CircleCheck /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.todayRuns }}</div>
          <div class="stat-label">今日执行</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon failed">
          <el-icon :size="24"><WarningFilled /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.failedRuns }}</div>
          <div class="stat-label">执行失败</div>
        </div>
      </div>
    </div>

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

    <el-card class="task-list-card">
      <template #header>
        <span class="list-title">任务列表</span>
        <span class="list-count">({{ filteredTasks.length }})</span>
      </template>

      <el-table :data="filteredTasks" v-loading="loading" stripe>
        <el-table-column prop="name" label="任务名称" min-width="180">
          <template #default="{ row }">
            <div class="task-name-cell">
              <span class="task-name">{{ row.name }}</span>
              <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                {{ row.enabled ? '已启用' : '已禁用' }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="同步配置" min-width="300">
          <template #default="{ row }">
            <div class="sync-config-cell">
              <div class="repo-badge source">
                <el-icon><Bottom /></el-icon>
                {{ row.sourceRepoKey }}:{{ row.sourceBranch }}
              </div>
              <el-icon class="arrow-icon"><Right /></el-icon>
              <div class="repo-badge target">
                <el-icon><Top /></el-icon>
                {{ row.targetRepoKey }}:{{ row.targetBranch }}
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="syncMode" label="同步模式" width="100">
          <template #default="{ row }">
            <el-tag type="primary" size="small">
              {{ row.syncMode === 'single' ? '单向' : '双向' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="cron" label="定时规则" width="160">
          <template #default="{ row }">
            <span v-if="row.cron" class="cron-text">
              <el-icon class="cron-icon"><Clock /></el-icon>
              {{ row.cron }}
            </span>
            <span v-else class="no-cron">手动触发</span>
          </template>
        </el-table-column>

        <el-table-column label="上次执行" width="180">
          <template #default="{ row }">
            <div v-if="row.lastRunAt" class="last-run-cell">
              <span :class="['status-dot', row.lastStatus]"></span>
              <span class="run-time">{{ formatTime(row.lastRunAt) }}</span>
            </div>
            <span v-else class="no-run">未执行</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="runTask(row)" :loading="runningKeys.has(row.key)">
              <el-icon><VideoPlay /></el-icon>
              运行
            </el-button>
            <el-button size="small" @click="editTask(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button size="small" type="danger" @click="deleteTask(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && filteredTasks.length === 0" description="暂无同步任务">
        <el-button type="primary" @click="openCreateModal">创建第一个任务</el-button>
      </el-empty>
    </el-card>

    <el-dialog
      v-model="modalVisible"
      :title="isEditMode ? '编辑同步任务' : '新建同步任务'"
      width="600px"
      @close="handleClose"
    >
      <el-form :model="taskForm" :rules="formRules" label-width="100px" ref="formRef">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="taskForm.name" placeholder="输入任务名称" />
        </el-form-item>
        <el-form-item label="源仓库" prop="sourceRepoKey">
          <el-input v-model="taskForm.sourceRepoKey" placeholder="源仓库标识" />
        </el-form-item>
        <el-form-item label="源分支" prop="sourceBranch">
          <el-input v-model="taskForm.sourceBranch" placeholder="如: main, develop" />
        </el-form-item>
        <el-form-item label="目标仓库" prop="targetRepoKey">
          <el-input v-model="taskForm.targetRepoKey" placeholder="目标仓库标识" />
        </el-form-item>
        <el-form-item label="目标分支" prop="targetBranch">
          <el-input v-model="taskForm.targetBranch" placeholder="如: master, main" />
        </el-form-item>
        <el-form-item label="同步模式" prop="syncMode">
          <el-radio-group v-model="taskForm.syncMode">
            <el-radio value="single">单向同步</el-radio>
            <el-radio value="bidirectional">双向同步</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="定时规则" prop="cron">
          <el-input v-model="taskForm.cron" placeholder="Cron 表达式，如: 0 * * * *" />
          <div class="form-tip">留空则仅支持手动触发</div>
        </el-form-item>
        <el-form-item label="同步选项">
          <el-checkbox v-model="taskForm.gitTags">同步标签</el-checkbox>
          <el-checkbox v-model="taskForm.gitForce">强制推送</el-checkbox>
          <el-checkbox v-model="taskForm.gitPrune">清理远程已删除分支</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modalVisible = false">取消</el-button>
        <el-button type="primary" @click="submitTask" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Edit, Delete, VideoPlay, Clock, Right, Top, Bottom } from '@element-plus/icons-vue'
import { syncV2Api } from '@/api/modules/sync_v2'
import type { SyncTask, SyncStats } from '@/types/sync_v2'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const isEditMode = ref(false)
const searchKeyword = ref('')
const statusFilter = ref<string | undefined>()
const runningKeys = ref<Set<string>>(new Set())
const formRef = ref()

const stats = reactive<SyncStats>({
  totalTasks: 0,
  enabledTasks: 0,
  todayRuns: 0,
  failedRuns: 0,
  runningTasks: 0,
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
  sourceRepoKey: '',
  sourceBranch: '',
  targetRepoKey: '',
  targetBranch: '',
  syncMode: 'single',
  cron: '',
  enabled: true,
  gitTags: false,
  gitForce: false,
  gitPrune: false,
  gitNoVerify: false,
  pushOptions: '',
})

const formRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  sourceRepoKey: [{ required: true, message: '请输入源仓库标识', trigger: 'blur' }],
  sourceBranch: [{ required: true, message: '请输入源分支', trigger: 'blur' }],
  targetRepoKey: [{ required: true, message: '请输入目标仓库标识', trigger: 'blur' }],
  targetBranch: [{ required: true, message: '请输入目标分支', trigger: 'blur' }],
}

function formatTime(timeStr: string) {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadStats() {
  try {
    const data = await syncV2Api.getStats()
    Object.assign(stats, data)
  } catch (e) {
    console.error('加载统计数据失败', e)
  }
}

async function loadTasks() {
  loading.value = true
  try {
    const data = await syncV2Api.listTasks()
    tasks.value = data || []
  } catch (e) {
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  isEditMode.value = false
  Object.assign(taskForm, {
    name: '',
    sourceRepoKey: '',
    sourceBranch: '',
    targetRepoKey: '',
    targetBranch: '',
    syncMode: 'single',
    cron: '',
    enabled: true,
    gitTags: false,
    gitForce: false,
    gitPrune: false,
    gitNoVerify: false,
    pushOptions: '',
    key: undefined,
  })
  modalVisible.value = true
}

function editTask(row: SyncTask) {
  isEditMode.value = true
  Object.assign(taskForm, { ...row })
  modalVisible.value = true
}

function handleClose() {
  formRef.value?.resetFields()
}

async function submitTask() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    if (isEditMode.value) {
      await syncV2Api.updateTask(taskForm as SyncTask)
      ElMessage.success('任务更新成功')
    } else {
      await syncV2Api.createTask(taskForm as SyncTask)
      ElMessage.success('任务创建成功')
    }
    modalVisible.value = false
    loadTasks()
    loadStats()
  } catch (e) {
    ElMessage.error(isEditMode.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

async function runTask(row: SyncTask) {
  runningKeys.value.add(row.key)
  try {
    await syncV2Api.runTask(row.key)
    ElMessage.success('任务已触发执行')
  } catch (e) {
    ElMessage.error('任务触发失败')
  } finally {
    runningKeys.value.delete(row.key)
  }
}

async function deleteTask(row: SyncTask) {
  try {
    await ElMessageBox.confirm(`确定要删除任务「${row.name}」吗？`, '确认删除', {
      type: 'warning',
    })
    await syncV2Api.deleteTask(row.key)
    ElMessage.success('删除成功')
    loadTasks()
    loadStats()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;

  .stat-card {
    background: #fff;
    border-radius: 12px;
    padding: 20px;
    display: flex;
    align-items: center;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    transition: all 0.2s;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
    }

    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 16px;

      &.total {
        background: linear-gradient(135deg, #ecf5ff 0%, #d9ecff 100%);
        color: #409eff;
      }

      &.running {
        background: linear-gradient(135deg, #f0f9eb 0%, #e1f3d8 100%);
        color: #67c23a;
      }

      &.success {
        background: linear-gradient(135deg, #fdf6ec 0%, #faecd8 100%);
        color: #e6a23c;
      }

      &.failed {
        background: linear-gradient(135deg, #fef0f0 0%, #fde2e2 100%);
        color: #f56c6c;
      }
    }

    .stat-content {
      .stat-value {
        font-size: 28px;
        font-weight: 600;
        color: #303133;
        line-height: 1.2;
      }

      .stat-label {
        font-size: 14px;
        color: #909399;
        margin-top: 4px;
      }
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

.task-list-card {
  background: #fff;
  border-radius: 12px;

  :deep(.el-card__header) {
    border-bottom: 1px solid #ebeef5;
    padding: 16px 20px;

    .list-title {
      font-weight: 600;
      font-size: 16px;
      color: #303133;
    }

    .list-count {
      font-size: 14px;
      color: #909399;
      margin-left: 8px;
    }
  }

  .task-name-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .task-name {
      font-weight: 500;
      color: #303133;
    }
  }

  .sync-config-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .repo-badge {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 4px 8px;
      border-radius: 4px;
      font-size: 12px;

      &.source {
        background: #ecf5ff;
        color: #409eff;
      }

      &.target {
        background: #f0f9eb;
        color: #67c23a;
      }
    }

    .arrow-icon {
      color: #909399;
    }
  }

  .cron-text {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #606266;
    font-size: 13px;

    .cron-icon {
      color: #909399;
    }
  }

  .no-cron {
    color: #909399;
    font-size: 13px;
  }

  .last-run-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;

      &.success {
        background: #67c23a;
      }

      &.failed {
        background: #f56c6c;
      }

      &.running {
        background: #409eff;
        animation: pulse 1.5s infinite;
      }
    }

    .run-time {
      font-size: 13px;
      color: #606266;
    }
  }

  .no-run {
    color: #909399;
    font-size: 13px;
  }
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
