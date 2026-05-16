<template>
  <div class="webhook-events-log">
    <PageHeader title="Webhook 事件日志">
      <template #actions>
        <el-button @click="loadData" :icon="Refresh">刷新</el-button>
      </template>
    </PageHeader>

    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon primary">
              <el-icon><DataLine /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">总事件数</div>
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
              <div class="stat-label">处理成功</div>
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
              <div class="stat-label">处理失败</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon warning">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.avgDuration }}ms</div>
              <div class="stat-label">平均耗时</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <SectionTitle title="筛选条件" />
    <el-card class="filter-card">
      <el-row :gutter="16">
        <el-col :span="4">
          <el-select v-model="filters.provider" placeholder="Git 平台" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="GitHub" value="github" />
            <el-option label="GitLab" value="gitlab" />
            <el-option label="Gitee" value="gitee" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.eventType" placeholder="事件类型" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="Push" value="push" />
            <el-option label="Create" value="create" />
            <el-option label="Delete" value="delete" />
            <el-option label="Pull Request" value="pull_request" />
            <el-option label="Tag" value="tag" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="处理状态" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-col>
        <el-col :span="12" style="text-align: right">
          <el-button @click="resetFilters" :icon="Refresh">重置</el-button>
          <el-button type="primary" @click="loadData" :icon="Search">搜索</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="events" v-loading="loading" class="events-table" stripe>
      <el-table-column prop="timestamp" label="时间" width="180" />
      <el-table-column prop="provider" label="平台" width="100">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.provider }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="eventType" label="事件类型" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ row.eventType }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="repo" label="仓库" min-width="180" />
      <el-table-column prop="branch" label="分支" width="140" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'success'" size="small" type="success">成功</el-tag>
          <el-tag v-else-if="row.status === 'failed'" size="small" type="danger">失败</el-tag>
          <el-tag v-else size="small" type="warning">处理中</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="duration" label="耗时" width="100" align="right">
        <template #default="{ row }">
          <span>{{ row.duration }}ms</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button size="small" :icon="View" @click="viewDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        :page-sizes="[20, 50, 100]"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh, DataLine, CircleCheck, CircleClose, Timer, Search, View } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const loading = ref(false)

const stats = reactive({
  total: 0,
  success: 0,
  failed: 0,
  avgDuration: 0,
})

const events = ref([
  {
    id: 1,
    timestamp: '2024-05-16 14:30:00',
    provider: 'github',
    eventType: 'push',
    repo: 'git-manage-service',
    branch: 'main',
    status: 'success',
    duration: 1250,
    ip: '192.30.252.1',
  },
  {
    id: 2,
    timestamp: '2024-05-16 13:45:00',
    provider: 'gitlab',
    eventType: 'push',
    repo: 'frontend',
    branch: 'develop',
    status: 'success',
    duration: 890,
    ip: '35.227.12.1',
  },
  {
    id: 3,
    timestamp: '2024-05-16 12:00:00',
    provider: 'github',
    eventType: 'tag',
    repo: 'backend',
    branch: '-',
    status: 'failed',
    duration: 520,
    ip: '192.30.252.2',
  },
])

const filters = reactive({
  provider: '',
  eventType: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 3,
})

onMounted(() => {
  stats.total = 1247
  stats.success = 1235
  stats.failed = 12
  stats.avgDuration = 980
})

function resetFilters() {
  filters.provider = ''
  filters.eventType = ''
  filters.status = ''
}

function loadData() {
  console.log('Load data with filters:', filters)
}

function viewDetail(row: any) {
  console.log('View detail:', row)
}
</script>

<style scoped lang="scss">
.webhook-events-log {
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

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
