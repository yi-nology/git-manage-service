<template>
  <div class="detail-header">
    <div class="header-left">
      <button class="back-btn" @click="emit('close')">
        <el-icon><ArrowLeft /></el-icon> 返回
      </button>
      <h3>MR !{{ task.mr_iid }} 审查详情</h3>
      <StatusBadge :variant="statusVariant(task.status)" :text="statusLabel(task.status)" />
      <StatusBadge v-if="task.risk_level" :variant="riskVariant(task.risk_level)" :text="riskLabel(task.risk_level)" :showDot="false" />
    </div>
    <div class="header-actions">
      <ActionPill variant="outline" small :icon="Refresh" :disabled="retrying" @click="emit('retry')">{{ retrying ? '重试中...' : '重新审查' }}</ActionPill>
      <ActionPill variant="outline" small :icon="Refresh" :disabled="loading" @click="emit('refresh')">刷新</ActionPill>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElIcon } from 'element-plus'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import type { ReviewTaskDTO } from '@/api/modules/review'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import { statusLabel, statusVariant, riskLabel, riskVariant } from './diff-utils'

defineProps<{
  task: ReviewTaskDTO
  retrying: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  close: []
  retry: []
  refresh: []
}>()
</script>

<style scoped>
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: transparent;
  color: var(--text-color-secondary);
  cursor: pointer;
  font-size: 13px;
}
.back-btn:hover { color: var(--accent-primary); border-color: var(--accent-primary); }

.header-left h3 { margin: 0; font-size: 16px; font-weight: 600; }
</style>
