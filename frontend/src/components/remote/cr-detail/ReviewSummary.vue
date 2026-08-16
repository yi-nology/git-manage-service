<template>
  <div class="summary-section" v-if="task.summary">
    <SectionTitle title="审查摘要" />
    <div class="summary-stats">
      <div class="stat-item" v-if="countBySeverity('critical') > 0">
        <span class="stat-dot" style="background: #DC2626"></span>
        <span>Critical: {{ countBySeverity('critical') }}</span>
      </div>
      <div class="stat-item" v-if="countBySeverity('high') > 0">
        <span class="stat-dot" style="background: #EA580C"></span>
        <span>High: {{ countBySeverity('high') }}</span>
      </div>
      <div class="stat-item" v-if="countBySeverity('medium') > 0">
        <span class="stat-dot" style="background: #D97706"></span>
        <span>Medium: {{ countBySeverity('medium') }}</span>
      </div>
      <div class="stat-item" v-if="countBySeverity('low') > 0">
        <span class="stat-dot" style="background: #2563EB"></span>
        <span>Low: {{ countBySeverity('low') }}</span>
      </div>
    </div>
    <div class="summary-text markdown-body" v-html="renderedSummary"></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import type { ReviewTaskDTO, ReviewFindingDTO } from '@/api/modules/review'
import SectionTitle from '@/components/common/SectionTitle.vue'

const props = defineProps<{
  task: ReviewTaskDTO
  findings: ReviewFindingDTO[]
}>()

const renderedSummary = computed(() => {
  if (!props.task.summary) return ''
  return marked.parse(props.task.summary) as string
})

function countBySeverity(s: string) {
  return props.findings.filter(f => f.severity === s).length
}
</script>

<style scoped>
.summary-section {
  padding: 16px;
  background: #0A0A0A;
  border-radius: 8px;
  color: #fff;
}

.summary-stats {
  display: flex;
  gap: 16px;
  margin: 8px 0 12px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #ccc;
}

.stat-dot {
  width: 8px;
  height: 8px;
  border-radius: 4px;
}

.summary-text {
  margin: 0;
  font-size: 13px;
  color: #ccc;
  line-height: 1.6;
}

.summary-text.markdown-body :deep(h2) { font-size: 18px; margin: 0 0 8px; color: #fff; border: none; }
.summary-text.markdown-body :deep(h3) { font-size: 15px; margin: 12px 0 6px; color: #eee; border: none; }
.summary-text.markdown-body :deep(p) { margin: 4px 0; }
.summary-text.markdown-body :deep(ul) { margin: 4px 0; padding-left: 20px; }
.summary-text.markdown-body :deep(li) { margin: 2px 0; }
.summary-text.markdown-body :deep(code) { background: rgba(255,255,255,0.1); padding: 1px 5px; border-radius: 3px; font-size: 12px; }
.summary-text.markdown-body :deep(table) { border-collapse: collapse; margin: 8px 0; font-size: 12px; width: 100%; }
.summary-text.markdown-body :deep(th) { background: rgba(255,255,255,0.08); text-align: left; padding: 6px 10px; border: 1px solid rgba(255,255,255,0.12); color: #ccc; }
.summary-text.markdown-body :deep(td) { padding: 5px 10px; border: 1px solid rgba(255,255,255,0.08); color: #bbb; }
.summary-text.markdown-body :deep(details) { margin: 8px 0; }
.summary-text.markdown-body :deep(summary) { cursor: pointer; color: #999; font-size: 12px; padding: 4px 0; }
.summary-text.markdown-body :deep(summary:hover) { color: #fff; }
.summary-text.markdown-body :deep(blockquote) { border-left: 3px solid rgba(255,255,255,0.15); margin: 4px 0; padding: 2px 12px; color: #999; }
.summary-text.markdown-body :deep(hr) { border: none; border-top: 1px solid rgba(255,255,255,0.1); margin: 8px 0; }
</style>
