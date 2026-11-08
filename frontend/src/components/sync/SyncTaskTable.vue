<template>
  <el-card class="task-list-card">
    <template #header>
      <span class="list-title">任务列表</span>
      <span class="list-count">({{ tasks.length }})</span>
    </template>

    <el-table :data="tasks" v-loading="loading" stripe>
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
              {{ row.source_repo_key }}:{{ row.source_branch }}
            </div>
            <el-icon class="arrow-icon"><Right /></el-icon>
            <div class="repo-badge target">
              <el-icon><Top /></el-icon>
              {{ row.target_repo_key }}:{{ row.target_branch }}
            </div>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="sync_mode" label="同步模式" width="100">
        <template #default="{ row }">
          <el-tag type="primary" size="small">
            {{ row.sync_mode === 'single' ? '单向' : '双向' }}
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
          <div v-if="row.last_run_at" class="last-run-cell">
            <span :class="['status-dot', row.last_status]"></span>
            <span class="run-time">{{ formatTime(row.last_run_at) }}</span>
          </div>
          <span v-else class="no-run">未执行</span>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="$emit('run', row)" :loading="runningKeys.has(row.key)">
            <el-icon><VideoPlay /></el-icon>
            运行
          </el-button>
          <el-button size="small" @click="$emit('edit', row)">
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
          <el-button size="small" type="danger" @click="$emit('delete', row)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && tasks.length === 0" description="暂无同步任务">
      <el-button type="primary" @click="$emit('create')">创建第一个任务</el-button>
    </el-empty>
  </el-card>
</template>

<script setup lang="ts">
import type { SyncTask } from '@/types/sync_v2'
import { VideoPlay, Edit, Delete, Clock, Right, Top, Bottom } from '@element-plus/icons-vue'

// 同步任务列表卡片：表格 + 空态，操作通过事件交回页面处理。
defineProps<{
  tasks: SyncTask[]
  loading: boolean
  runningKeys: Set<string>
}>()

defineEmits<{
  run: [row: SyncTask]
  edit: [row: SyncTask]
  delete: [row: SyncTask]
  create: []
}>()

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
</script>

<style scoped lang="less">
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

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
