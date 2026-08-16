<template>
  <div class="findings-section">
    <div class="findings-header">
      <SectionTitle :title="`问题列表 (${filteredFindings.length})`" />
      <div class="filter-pills">
        <ActionPill small :variant="!severityFilter && !sourceFilter ? 'primary' : 'outline'" @click="clearFilters">全部</ActionPill>
        <ActionPill small :variant="sourceFilter === 'llm' ? 'primary' : 'outline'" @click="setSourceFilter('llm')">AI 审查</ActionPill>
        <ActionPill small :variant="sourceFilter === 'rule' ? 'primary' : 'outline'" @click="setSourceFilter('rule')">规则</ActionPill>
        <ActionPill small :variant="severityFilter === 'critical' ? 'danger' : 'outline'" @click="setSeverityFilter('critical')">严重</ActionPill>
        <ActionPill small :variant="severityFilter === 'high' ? 'amber' : 'outline'" @click="setSeverityFilter('high')">高危</ActionPill>
        <ActionPill small :variant="severityFilter === 'medium' ? 'amber' : 'outline'" @click="setSeverityFilter('medium')">中等</ActionPill>
        <ActionPill small :variant="severityFilter === 'low' ? 'primary' : 'outline'" @click="setSeverityFilter('low')">低危</ActionPill>
      </div>
    </div>

    <LoadingState v-if="loading" />

    <div v-else-if="filteredFindings.length === 0" class="empty-findings">
      <EmptyState :title="severityFilter ? '暂无该级别的问题' : '暂无问题'" />
    </div>

    <div v-else class="finding-cards">
      <div v-for="f in filteredFindings" :key="f.id" class="finding-card" :class="'finding-' + f.severity">
        <div class="finding-head">
          <StatusBadge :variant="severityVariant(f.severity)" :text="severityText(f.severity)" :showDot="false" />
          <span class="rule-badge">{{ f.rule_id }}</span>
          <span class="source-badge">{{ f.source === 'llm' ? 'AI' : '规则' }}</span>
        </div>
        <div class="finding-title">{{ f.title }}</div>
        <div class="finding-file" v-if="f.file_path">{{ f.file_path }}<template v-if="f.new_line"> : 第 {{ f.new_line }} 行</template></div>
        <div class="finding-message">{{ f.message }}</div>
        <div class="finding-suggestion" v-if="f.suggestion">
          <span class="suggestion-label">建议:</span> {{ f.suggestion }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ReviewFindingDTO } from '@/api/modules/review'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import { severityText, severityVariant } from './diff-utils'

const props = defineProps<{
  findings: ReviewFindingDTO[]
  loading: boolean
}>()

const severityFilter = ref('')
const sourceFilter = ref('')

const filteredFindings = computed(() => {
  let result = props.findings
  if (severityFilter.value) result = result.filter(f => f.severity === severityFilter.value)
  if (sourceFilter.value) result = result.filter(f => f.source === sourceFilter.value)
  return result
})

function clearFilters() {
  severityFilter.value = ''
  sourceFilter.value = ''
}

function setSourceFilter(source: string) {
  sourceFilter.value = source
  severityFilter.value = ''
}

function setSeverityFilter(severity: string) {
  severityFilter.value = severity
  sourceFilter.value = ''
}
</script>

<style scoped>
.findings-section { margin-top: 8px; }

.findings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-pills { display: flex; gap: 6px; flex-wrap: wrap; }

.finding-cards { display: flex; flex-direction: column; gap: 12px; }

.finding-card {
  padding: 16px;
  border-radius: 8px;
  background: var(--surface-card);
  border: 1px solid var(--border-color);
  border-left: 3px solid var(--border-color);
}
.finding-critical { border-left-color: #DC2626; }
.finding-high { border-left-color: #EA580C; }
.finding-medium { border-left-color: #D97706; }
.finding-low { border-left-color: #2563EB; }
.finding-info { border-left-color: #9CA3AF; }

.finding-head { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
.rule-badge { font-size: 11px; font-family: 'SF Mono', monospace; color: var(--text-color-secondary); }
.source-badge { font-size: 11px; padding: 1px 6px; border-radius: 3px; background: var(--accent-bg); color: var(--accent-primary); }

.finding-title { font-size: 14px; font-weight: 600; margin-bottom: 6px; }
.finding-file { font-size: 12px; font-family: 'SF Mono', monospace; color: var(--accent-primary); margin-bottom: 8px; }
.finding-message { font-size: 13px; color: var(--text-color-secondary); line-height: 1.5; margin-bottom: 8px; }
.finding-suggestion { font-size: 12px; padding: 8px 12px; background: #F0FDF4; border-radius: 4px; color: #166534; line-height: 1.4; }
.suggestion-label { font-weight: 600; }

.empty-findings { padding: 20px; }
</style>
