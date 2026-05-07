<template>
  <div class="cr-review-detail">
    <div class="detail-header">
      <div class="header-left">
        <button class="back-btn" @click="$emit('close')">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h3>MR !{{ task.mr_iid }} 审查详情</h3>
        <StatusBadge :variant="statusVariant(task.status)" :text="statusLabel(task.status)" />
        <StatusBadge v-if="task.risk_level" :variant="riskVariant(task.risk_level)" :text="riskLabel(task.risk_level)" :showDot="false" />
      </div>
      <div class="header-actions">
        <ActionPill variant="outline" small :icon="Refresh" :disabled="loading" @click="loadFindings">刷新</ActionPill>
      </div>
    </div>

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
      <p class="summary-text">{{ task.summary }}</p>
    </div>

    <div class="findings-section">
      <div class="findings-header">
        <SectionTitle :title="`问题列表 (${filteredFindings.length})`" />
        <div class="filter-pills">
          <ActionPill small :variant="!severityFilter ? 'primary' : 'outline'" @click="severityFilter = ''">全部</ActionPill>
          <ActionPill small :variant="severityFilter === 'critical' ? 'danger' : 'outline'" @click="severityFilter = 'critical'">严重</ActionPill>
          <ActionPill small :variant="severityFilter === 'high' ? 'amber' : 'outline'" @click="severityFilter = 'high'">高危</ActionPill>
          <ActionPill small :variant="severityFilter === 'medium' ? 'amber' : 'outline'" @click="severityFilter = 'medium'">中等</ActionPill>
          <ActionPill small :variant="severityFilter === 'low' ? 'primary' : 'outline'" @click="severityFilter = 'low'">低危</ActionPill>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElIcon } from 'element-plus'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { getReviewTask, listReviewFindings, type ReviewTaskDTO, type ReviewFindingDTO } from '@/api/modules/review'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'

const props = defineProps<{
  task: ReviewTaskDTO
}>()

const emit = defineEmits<{
  close: []
}>()

const loading = ref(false)
const findings = ref<ReviewFindingDTO[]>([])
const severityFilter = ref('')

const filteredFindings = computed(() => {
  if (!severityFilter.value) return findings.value
  return findings.value.filter(f => f.severity === severityFilter.value)
})

function countBySeverity(s: string) {
  return findings.value.filter(f => f.severity === s).length
}

function severityText(s: string) {
  const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }
  return m[s] || s
}

function statusLabel(s: string) {
  const m: Record<string, string> = { pending: '等待中', running: '运行中', success: '通过', failed: '失败', blocked: '已阻止' }
  return m[s] || s
}

function riskLabel(r: string) {
  const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }
  return m[r] || r
}

function triggerLabel(t: string) {
  const m: Record<string, string> = { manual: '手动', webhook: 'Webhook', api: 'API' }
  return m[t] || t
}

function shortSha(sha: string) {
  return sha ? sha.substring(0, 7) : '—'
}

function timeAgo(d: string) {
  if (!d) return '—'
  const diff = Date.now() - new Date(d).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const h = Math.floor(mins / 60)
  if (h < 24) return `${h} 小时前`
  return `${Math.floor(h / 24)} 天前`
}

type StatusVariant = 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'

function statusVariant(s: string): StatusVariant {
  const m: Record<string, StatusVariant> = { pending: 'warning', running: 'running', success: 'success', failed: 'danger', blocked: 'danger' }
  return m[s] || 'default'
}

function riskVariant(r: string): StatusVariant {
  const m: Record<string, StatusVariant> = { critical: 'danger', high: 'warning', medium: 'warning', low: 'info', info: 'default' }
  return m[r] || 'default'
}

function severityVariant(s: string): StatusVariant {
  const m: Record<string, StatusVariant> = { critical: 'danger', high: 'warning', medium: 'warning', low: 'info', info: 'default' }
  return m[s] || 'default'
}

async function loadFindings() {
  loading.value = true
  try {
    const res = await listReviewFindings(props.task.id)
    findings.value = res || []
  } catch {
    findings.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadFindings)
</script>

<style scoped>
.cr-review-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

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
