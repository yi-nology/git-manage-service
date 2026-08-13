<template>
  <div class="sync-dashboard">
    <PageHeader title="同步任务">
      <template #actions>
        <el-button type="primary" @click="$router.push('/sync/new')" :icon="Plus">新建同步任务</el-button>
      </template>
    </PageHeader>

    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon success">
              <el-icon><Check /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.total_tasks }}</div>
              <div class="stat-label">总任务数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon primary">
              <el-icon><VideoPlay /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.running_tasks }}</div>
              <div class="stat-label">运行中</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon warning">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todaySyncs }}</div>
              <div class="stat-label">今日同步</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon danger">
              <el-icon><Close /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.failedTasks }}</div>
              <div class="stat-label">失败任务</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <SectionTitle title="同步任务列表" />
    <el-card class="filter-card">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input v-model="filters.keyword" placeholder="搜索任务名称" clearable :prefix-icon="Search" />
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="已启用" value="enabled" />
            <el-option label="已暂停" value="paused" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.repo" placeholder="仓库" clearable style="width: 100%">
            <el-option label="全部仓库" value="" />
            <el-option v-for="repo in repos" :key="repo.key" :label="repo.name" :value="repo.key" />
          </el-select>
        </el-col>
        <el-col :span="10" style="text-align: right">
          <el-button @click="resetFilters" :icon="Refresh">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="tasks" v-loading="loading" class="tasks-table">
      <el-table-column prop="name" label="任务名称" min-width="180">
        <template #default="{ row }">
          <div class="task-name-cell">
            <el-tag v-if="row.enabled" type="success" size="small" effect="plain">启用</el-tag>
            <el-tag v-else size="small" effect="plain">暂停</el-tag>
            <span class="name">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="repo_name" label="仓库" width="140" />
      <el-table-column prop="sync_mode" label="同步模式" width="120">
        <template #default="{ row }">
          <el-tag type="info" size="small">{{ row.sync_mode === 'all-branch' ? '全分支' : '单分支' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="source" label="源 → 目标" min-width="200">
        <template #default="{ row }">
          <div class="sync-flow">
            <span class="source">{{ row.sourceRemote }}</span>
            <el-icon class="arrow"><Right /></el-icon>
            <span class="target">{{ row.targetRemote }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="cron" label="调度规则" width="160">
        <template #default="{ row }">
          <span v-if="row.cron" class="cron-text">
            <el-icon><Timer /></el-icon>
            {{ row.cron }}
          </span>
          <span v-else class="cron-text">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="lastSync" label="上次同步" width="180">
        <template #default="{ row }">
          <span v-if="row.lastSync" :class="getLastSyncClass(row)">
            {{ formatTime(row.lastSync) }}
          </span>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="success" :icon="VideoPlay" circle @click="run_task(row)" />
          <el-button size="small" :icon="Clock" circle @click="viewHistory(row)" />
          <el-button size="small" :icon="Edit" circle @click="editTask(row)" />
          <el-button size="small" type="danger" :icon="Delete" circle @click="delete_task(row)" />
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        :page-sizes="[10, 20, 50, 100]"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Check, VideoPlay, Clock, Close, Search, Refresh, Right, Timer, Edit, Delete } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const router = useRouter()

const loading = ref(false)
const stats = reactive({
  total_tasks: 0,
  running_tasks: 0,
  todaySyncs: 0,
  failedTasks: 0,
})

const tasks = ref([
  {
    id: 1,
    name: '主分支同步',
    repo_name: 'git-manage-service',
    sync_mode: 'single',
    sourceRemote: 'origin',
    targetRemote: 'mirror',
    source_branch: 'main',
    target_branch: 'main',
    enabled: true,
    cron: '0 * * * *',
    lastSync: '2024-05-16 14:30:00',
    lastSyncStatus: 'success',
  },
  {
    id: 2,
    name: '全分支镜像同步',
    repo_name: 'frontend',
    sync_mode: 'all-branch',
    sourceRemote: 'origin',
    targetRemote: 'backup',
    enabled: true,
    cron: '30 * * * *',
    lastSync: '2024-05-16 13:30:00',
    lastSyncStatus: 'success',
  },
  {
    id: 3,
    name: '开发分支同步',
    repo_name: 'backend',
    sync_mode: 'single',
    sourceRemote: 'origin',
    targetRemote: 'dev',
    source_branch: 'develop',
    target_branch: 'develop',
    enabled: false,
    cron: '',
    lastSync: '2024-05-15 10:00:00',
    lastSyncStatus: 'failed',
  },
])

const repos = ref([
  { key: 'git-manage-service', name: 'git-manage-service' },
  { key: 'frontend', name: 'frontend' },
  { key: 'backend', name: 'backend' },
])

const filters = reactive({
  keyword: '',
  status: '',
  repo: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 3,
})

onMounted(() => {
  stats.total_tasks = 3
  stats.running_tasks = 1
  stats.todaySyncs = 12
  stats.failedTasks = 1
})

function resetFilters() {
  filters.keyword = ''
  filters.status = ''
  filters.repo = ''
}

function run_task(row: any) {
  console.log('Run task:', row)
}

function viewHistory(_row: any) {
  router.push('/sync/history')
}

function editTask(row: any) {
  console.log('Edit task:', row)
}

function delete_task(row: any) {
  console.log('Delete task:', row)
}

function formatTime(time: string) {
  return time
}

function getLastSyncClass(row: any) {
  return row.lastSyncStatus === 'failed' ? 'text-danger' : 'text-success'
}
</script>

<style scoped lang="scss">
.sync-dashboard {
  .stats-row {
    margin-bottom: 24px;
  }

  .stat-card {
    .stat-content {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 24px;

      &.success {
        background: var(--el-color-success-lighter);
        color: var(--el-color-success);
      }

      &.primary {
        background: var(--el-color-primary-lighter);
        color: var(--el-color-primary);
      }

      &.warning {
        background: var(--el-color-warning-lighter);
        color: var(--el-color-warning);
      }

      &.danger {
        background: var(--el-color-danger-lighter);
        color: var(--el-color-danger);
      }
    }

    .stat-info {
      .stat-value {
        font-size: 28px;
        font-weight: 700;
        color: var(--text-color-primary);
        line-height: 1.2;
      }

      .stat-label {
        font-size: 13px;
        color: var(--text-color-secondary);
        margin-top: 4px;
      }
    }
  }

  .filter-card {
    margin-bottom: 16px;
  }

  .tasks-table {
    .task-name-cell {
      display: flex;
      align-items: center;
      gap: 8px;

      .name {
        font-weight: 500;
      }
    }

    .sync-flow {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 13px;

      .source {
        color: var(--el-color-primary);
        font-weight: 500;
      }

      .arrow {
        color: var(--text-color-placeholder);
        font-size: 12px;
      }

      .target {
        color: var(--el-color-success);
        font-weight: 500;
      }
    }

    .cron-text {
      display: flex;
      align-items: center;
      gap: 4px;
      font-family: 'SF Mono', Monaco, monospace;
      font-size: 12px;
      color: var(--text-color-secondary);
    }

    .text-muted {
      color: var(--text-color-placeholder);
    }

    .text-success {
      color: var(--el-color-success);
    }

    .text-danger {
      color: var(--el-color-danger);
    }
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
