<template>
  <div class="meta-info">
    <div class="meta-item">
      <span class="meta-label">提交</span>
      <span class="meta-value mono">{{ shortSha(task.commit_sha) }}</span>
    </div>
    <div class="meta-item">
      <span class="meta-label">触发</span>
      <span class="meta-value">{{ triggerLabel(task.trigger_type) }}</span>
    </div>
    <div class="meta-item">
      <span class="meta-label">时间</span>
      <span class="meta-value">{{ timeAgo(task.created_at) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ReviewTaskDTO } from '@/api/modules/review'
import { shortSha, triggerLabel, timeAgo } from './diff-utils'

defineProps<{
  task: ReviewTaskDTO
}>()
</script>

<style scoped>
.meta-info {
  display: flex;
  gap: 24px;
  padding: 12px 16px;
  background: var(--surface-card);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.meta-item { display: flex; flex-direction: column; gap: 2px; }
.meta-label { font-size: 12px; color: var(--text-color-placeholder); }
.meta-value { font-size: 13px; color: var(--text-color-primary); }
.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; }
</style>
