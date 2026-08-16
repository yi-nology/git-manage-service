<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="日志详情"
    width="750px"
    destroy-on-close
    class="detail-dialog"
  >
    <div v-if="currentLog">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog.status === 'success' ? 'success' : 'danger'">{{ currentLog.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="触发类型">{{ getTriggerLabel(currentLog.trigger_type) }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ currentLog.duration_ms ? `${(currentLog.duration_ms / 1000).toFixed(2)}s` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="分支 / 提交">
          {{ currentLog.branches_synced || 0 }} / {{ currentLog.commits_pushed || 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="错误" :span="2" v-if="currentLog.error_message">
          <span class="error-text">{{ currentLog.error_message }}</span>
        </el-descriptions-item>
      </el-descriptions>
      <div v-if="currentLog.detail_log" class="log-section">
        <div class="log-section-title">执行日志</div>
        <pre class="log-content">{{ currentLog.detail_log }}</pre>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import type { MirrorSyncLogDTO } from '@/types/mirror'
import { TRIGGER_TYPE_MAP } from '@/types/mirror'

defineProps<{
  visible: boolean
  currentLog: MirrorSyncLogDTO | null
}>()

defineEmits<{
  'update:visible': [value: boolean]
}>()

function getTriggerLabel(type: string): string {
  return TRIGGER_TYPE_MAP[type] || type
}
</script>

<style scoped>
.log-section {
  margin-top: 20px;
}

.log-section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-color-primary);
}

.log-content {
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-base);
  padding: 12px;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-text {
  color: var(--el-color-danger);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 12px;
}

:deep(.detail-dialog .el-dialog__body) {
  padding-top: 16px;
}
</style>
