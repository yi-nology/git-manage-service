<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="同步日志"
    width="900px"
    destroy-on-close
    class="log-dialog"
  >
    <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
      <span style="color: var(--text-color-secondary); font-size: 12px;">
        提示：同步任务在后台异步执行，如无数据请点击刷新按钮
      </span>
      <el-button size="small" @click="$emit('refresh')" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>
    <el-empty v-if="!loading && syncLogs.length === 0" description="暂无同步日志" :image-size="120" />
    <el-table v-if="syncLogs.length > 0" :data="syncLogs" v-loading="loading" stripe max-height="500">
      <el-table-column label="触发类型" width="100">
        <template #default="{ row }">
          <el-tag :type="getTriggerType(row.trigger_type)" size="small">
            {{ getTriggerLabel(row.trigger_type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="耗时" width="90">
        <template #default="{ row }">{{ row.duration_ms ? `${(row.duration_ms / 1000).toFixed(1)}s` : '-' }}</template>
      </el-table-column>
      <el-table-column label="分支" prop="branches_synced" width="70" align="center" />
      <el-table-column label="提交" prop="commits_pushed" width="70" align="center" />
      <el-table-column label="开始时间" width="170">
        <template #default="{ row }">{{ row.started_at ? formatTime(row.started_at) : '-' }}</template>
      </el-table-column>
      <el-table-column label="错误" prop="error_message" min-width="200" show-overflow-tooltip />
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="$emit('show-detail', row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>

<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'
import type { MirrorSyncLogDTO } from '@/types/mirror'
import { TRIGGER_TYPE_MAP } from '@/types/mirror'

defineProps<{
  visible: boolean
  loading: boolean
  syncLogs: MirrorSyncLogDTO[]
}>()

defineEmits<{
  'update:visible': [value: boolean]
  refresh: []
  'show-detail': [log: MirrorSyncLogDTO]
}>()

function getTriggerType(type: string): 'success' | 'warning' | 'danger' | 'primary' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'primary' | 'info'> = {
    manual: 'primary',
    cron: 'warning',
    webhook: 'success',
    push_event: 'info',
  }
  return map[type] || 'info'
}

function getTriggerLabel(type: string): string {
  return TRIGGER_TYPE_MAP[type] || type
}

function formatTime(timeStr: string): string {
  if (!timeStr) return '-'
  const d = new Date(timeStr)
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
:deep(.log-dialog .el-dialog__body) {
  padding-top: 16px;
}
</style>
