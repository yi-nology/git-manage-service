<template>
  <div class="sync-history">
    <PageHeader title="同步历史">
      <template #actions>
        <el-button :icon="Refresh" @click="loadData">刷新</el-button>
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
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">总同步次数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon success">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.success }}</div>
              <div class="stat-label">成功</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon danger">
              <el-icon><CircleClose /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.failed }}</div>
              <div class="stat-label">失败</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon primary">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.avgDuration }}s</div>
              <div class="stat-label">平均耗时</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <SectionTitle title="筛选条件" />
    <el-card class="filter-card">
      <el-row :gutter="16">
        <el-col :span="5">
          <el-input v-model="filters.taskName" placeholder="任务名称" clearable :prefix-icon="Search" />
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="进行中" value="running" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.repo" placeholder="仓库" clearable style="width: 100%">
            <el-option label="全部仓库" value="" />
            <el-option v-for="repo in repos" :key="repo.key" :label="repo.name" :value="repo.key" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-date-picker
            v-model="filters.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            style="width: 100%"
          />
        </el-col>
        <el-col :span="5" style="text-align: right">
          <el-button @click="resetFilters" :icon="Refresh">重置</el-button>
          <el-button type="primary" @click="loadData" :icon="Search">搜索</el-button>
        </el-col>
      </el-row>
    </el-card>

    <div class="history-list">
      <el-card v-for="item in historyList" :key="item.id" class="history-card" :class="item.status">
        <div class="history-header">
          <div class="task-info">
            <div class="status-icon">
              <el-icon v-if="item.status === 'success'" class="success"><CircleCheck /></el-icon>
              <el-icon v-else-if="item.status === 'failed'" class="danger"><CircleClose /></el-icon>
              <el-icon v-else class="primary"><Loading /></el-icon>
            </div>
            <div>
              <div class="task-name">{{ item.taskName }}</div>
              <div class="task-repo">{{ item.repo_name }}</div>
            </div>
          </div>
          <div class="time-info">
            <div class="start-time">{{ item.start_time }}</div>
            <div class="duration">耗时: {{ item.duration }}s</div>
          </div>
        </div>

        <div class="history-detail">
          <div class="sync-flow">
            <div class="endpoint">
              <span class="label">源</span>
              <span class="value">{{ item.sourceRemote }}/{{ item.source_branch }}</span>
            </div>
            <el-icon class="arrow"><Right /></el-icon>
            <div class="endpoint">
              <span class="label">目标</span>
              <span class="value">{{ item.targetRemote }}/{{ item.target_branch }}</span>
            </div>
          </div>
        </div>

        <div v-if="item.error_message" class="error-message">
          <el-icon><Warning /></el-icon>
          <span>{{ item.error_message }}</span>
        </div>

        <div class="history-actions">
          <el-button size="small" type="primary" :icon="Document" @click="viewLogs(item)">查看日志</el-button>
          <el-button size="small" :icon="RefreshRight" @click="retrySync(item)">重新同步</el-button>
        </div>
      </el-card>
    </div>

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
import {
  Refresh,
  Check,
  CircleCheck,
  CircleClose,
  Timer,
  Search,
  Right,
  Loading,
  Warning,
  Document,
  RefreshRight,
} from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const loading = ref(false)
const stats = reactive({
  total: 0,
  success: 0,
  failed: 0,
  avgDuration: 0,
})

const repos = ref([
  { key: 'git-manage-service', name: 'git-manage-service' },
  { key: 'frontend', name: 'frontend' },
  { key: 'backend', name: 'backend' },
])

const filters = reactive({
  taskName: '',
  status: '',
  repo: '',
  dateRange: null as [Date, Date] | null,
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 3,
})

const historyList = ref([
  {
    id: 1,
    taskName: '主分支同步',
    repo_name: 'git-manage-service',
    status: 'success',
    sourceRemote: 'origin',
    source_branch: 'main',
    targetRemote: 'mirror',
    target_branch: 'main',
    start_time: '2024-05-16 14:30:00',
    duration: 12,
    commit_count: 5,
    error_message: '',
  },
  {
    id: 2,
    taskName: '全分支镜像同步',
    repo_name: 'frontend',
    status: 'success',
    sourceRemote: 'origin',
    source_branch: 'all',
    targetRemote: 'backup',
    target_branch: 'all',
    start_time: '2024-05-16 13:30:00',
    duration: 45,
    commit_count: 12,
    error_message: '',
  },
  {
    id: 3,
    taskName: '开发分支同步',
    repo_name: 'backend',
    status: 'failed',
    sourceRemote: 'origin',
    source_branch: 'develop',
    targetRemote: 'dev',
    target_branch: 'develop',
    start_time: '2024-05-15 10:00:00',
    duration: 8,
    commit_count: 0,
    error_message: '认证失败: 无法连接到远端仓库 dev',
  },
])

onMounted(() => {
  stats.total = 156
  stats.success = 152
  stats.failed = 4
  stats.avgDuration = 18
})

function resetFilters() {
  filters.taskName = ''
  filters.status = ''
  filters.repo = ''
  filters.dateRange = null
}

function loadData() {
  console.log('Load data with filters:', filters)
}

function viewLogs(item: any) {
  console.log('View logs:', item)
}

function retrySync(item: any) {
  console.log('Retry sync:', item)
}
</script>

<style scoped lang="scss">
.sync-history {
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

  .history-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .history-card {
    &.success {
      border-left: 4px solid var(--el-color-success);
    }

    &.failed {
      border-left: 4px solid var(--el-color-danger);
    }

    &.running {
      border-left: 4px solid var(--el-color-primary);
    }

    .history-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 16px;

      .task-info {
        display: flex;
        align-items: center;
        gap: 12px;

        .status-icon {
          font-size: 24px;

          .success {
            color: var(--el-color-success);
          }

          .danger {
            color: var(--el-color-danger);
          }

          .primary {
            color: var(--el-color-primary);
          }
        }

        .task-name {
          font-size: 15px;
          font-weight: 600;
          color: var(--text-color-primary);
        }

        .task-repo {
          font-size: 12px;
          color: var(--text-color-secondary);
          margin-top: 2px;
        }
      }

      .time-info {
        text-align: right;

        .start-time {
          font-size: 13px;
          color: var(--text-color-secondary);
        }

        .duration {
          font-size: 12px;
          color: var(--text-color-placeholder);
          margin-top: 4px;
        }
      }
    }

    .history-detail {
      padding: 12px 16px;
      background: var(--bg-color-page);
      border-radius: 8px;
      margin-bottom: 12px;

      .sync-flow {
        display: flex;
        align-items: center;
        gap: 16px;

        .endpoint {
          display: flex;
          flex-direction: column;
          gap: 4px;

          .label {
            font-size: 11px;
            color: var(--text-color-placeholder);
            text-transform: uppercase;
          }

          .value {
            font-size: 14px;
            font-weight: 500;
            font-family: 'SF Mono', Monaco, monospace;
          }
        }

        .arrow {
          color: var(--text-color-placeholder);
          font-size: 18px;
        }
      }
    }

    .error-message {
      display: flex;
      align-items: flex-start;
      gap: 8px;
      padding: 10px 14px;
      background: var(--el-color-danger-lighter);
      border-radius: 6px;
      color: var(--el-color-danger);
      font-size: 13px;
      margin-bottom: 12px;

      .el-icon {
        font-size: 16px;
        flex-shrink: 0;
        margin-top: 1px;
      }
    }

    .history-actions {
      display: flex;
      gap: 8px;
      padding-top: 8px;
      border-top: 1px solid var(--border-color-light);
    }
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
