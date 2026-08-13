 <template>
   <div class="review-page-wrapper">
     <div class="header-bar">
       <PageHeader :showBack="true" :backRoute="`/local-repos/${repo_key}/review/tasks`">
         <template #actions>
           <ActionPill variant="ai" :icon="MagicStick" :disabled="aiLoading" @click="showAIPanel = !showAIPanel">
             AI 审查助手
           </ActionPill>
           <ActionPill variant="outline" :icon="Refresh" :disabled="retrying" @click="handleRetry">
             {{ retrying ? '重试中...' : '重试' }}
           </ActionPill>
         </template>
       </PageHeader>
     </div>

    <div class="detail-layout">
      <LoadingState v-if="loading && !task" />
      <div class="detail-content" v-else>
        <div class="breadcrumb">
          <router-link :to="`/local-repos/${repo_key}/review`">总览</router-link>
          <span class="sep">/</span>
          <router-link :to="`/local-repos/${repo_key}/review/tasks`">任务</router-link>
          <span class="sep">/</span>
          <span class="current">审查任务 #{{ task_id }}</span>
        </div>

        <div class="title-row">
          <div>
            <div class="repo-label">{{ task?.repo_name || repo_key }} · MR #{{ task?.mr_iid }}</div>
            <h2>审查任务 #{{ task_id }}</h2>
          </div>
        </div>

        <div class="meta-row" v-if="task">
          <StatusBadge :variant="statusVariant(task.status)" :text="statusLabel(task.status)" />
          <StatusBadge v-if="task.risk_level" :variant="riskVariant(task.risk_level)" :text="riskLabel(task.risk_level)" />
          <span class="meta-text">提交: {{ shortSha(task.commit_sha) }} · 触发: {{ triggerLabel(task.trigger_type) }} · {{ timeAgo(task.created_at) }}</span>
        </div>

        <div class="info-grid" v-if="task">
          <div class="info-item">
            <div class="info-label">MR 编号</div>
            <div class="info-value">!{{ task.mr_iid }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">Commit SHA</div>
            <div class="info-value mono">{{ shortSha(task.commit_sha) }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">触发方式</div>
            <div class="info-value">{{ triggerLabel(task.trigger_type) }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">风险等级</div>
            <div class="info-value">
              <StatusBadge v-if="task.risk_level" :variant="riskVariant(task.risk_level)" :text="riskLabel(task.risk_level)" :showDot="false" />
              <span v-else>—</span>
            </div>
          </div>
        </div>

        <div class="summary-card" v-if="task?.summary">
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
          <p>{{ task.summary }}</p>
        </div>

        <div class="error-card" v-if="task?.error_message">
          <SectionTitle title="错误信息" />
          <p>{{ task.error_message }}</p>
        </div>

        <div class="findings-section">
          <div class="findings-header">
            <SectionTitle title="问题列表" />
            <div class="filter-pills">
              <ActionPill small :variant="!severityFilter ? 'primary' : 'outline'" @click="severityFilter = ''">全部 ({{ findings.length }})</ActionPill>
              <ActionPill small :variant="severityFilter === 'critical' ? 'danger' : 'outline'" @click="severityFilter = 'critical'">严重 ({{ countBySeverity('critical') }})</ActionPill>
              <ActionPill small :variant="severityFilter === 'high' ? 'amber' : 'outline'" @click="severityFilter = 'high'">高危 ({{ countBySeverity('high') }})</ActionPill>
              <ActionPill small :variant="severityFilter === 'medium' ? 'amber' : 'outline'" @click="severityFilter = 'medium'">中等 ({{ countBySeverity('medium') }})</ActionPill>
              <ActionPill small :variant="severityFilter === 'low' ? 'primary' : 'outline'" @click="severityFilter = 'low'">低危 ({{ countBySeverity('low') }})</ActionPill>
            </div>
          </div>

          <div class="finding-cards">
            <div v-for="f in filteredFindings" :key="f.id" class="finding-card" :class="'finding-' + f.severity">
              <div class="finding-head">
                <div class="finding-badges">
                  <StatusBadge :variant="severityVariant(f.severity)" :text="severityText(f.severity)" :showDot="false" />
                  <span class="rule-badge">{{ f.rule_id }}</span>
                  <span class="source-badge">{{ f.source === 'llm' ? 'AI' : '规则' }}</span>
                </div>
              </div>
              <div class="finding-title">{{ f.title }}</div>
              <div class="finding-file" v-if="f.file_path">{{ f.file_path }}<template v-if="f.new_line"> : 第 {{ f.new_line }} 行</template></div>
              <div class="finding-message">{{ f.message }}</div>
              <div class="finding-suggestion" v-if="f.suggestion">
                <span class="suggestion-label">建议:</span> {{ f.suggestion }}
              </div>
            </div>
          </div>

           <EmptyState v-if="filteredFindings.length === 0 && !loading" :title="severityFilter ? '暂无该级别的问题' : '暂无问题'" />
         </div>
       </div>

       <AIPanel
         ref="aiPanelRef"
         v-if="showAIPanel"
         title="AI 审查助手"
          v-model:visible="showAIPanel"
         :quick-actions="quickActions"
         empty-hint="AI 可以帮助分析代码变更、识别潜在问题并提供改进建议。输入您的问题开始使用。"
         :ai-loading="aiLoading"
         @send="handleAIReview"
         @close="showAIPanel = false"
       />
     </div>
   </div>
 </template>

 <script setup lang="ts">
 import { ref, computed, onMounted } from 'vue'
 import { useRoute } from 'vue-router'
 import { ElMessage } from 'element-plus'
 import { Refresh, MagicStick } from '@element-plus/icons-vue'
 import { getReviewTask, listReviewFindings, retryReviewTask, type ReviewTaskDTO, type ReviewFindingDTO } from '@/api/modules/review'
 import { aiApi } from '@/api/modules/ai'
 import PageHeader from '@/components/common/PageHeader.vue'
 import LoadingState from '@/components/common/LoadingState.vue'
 import ActionPill from '@/components/common/ActionPill.vue'
 import StatusBadge from '@/components/common/StatusBadge.vue'
 import SectionTitle from '@/components/common/SectionTitle.vue'
 import EmptyState from '@/components/common/EmptyState.vue'
 import AIPanel from '@/components/ai/AIPanel.vue'
 import type { QuickAction } from '@/types/ai'

const route = useRoute()
const repo_key = route.params.repo_key as string
const task_id = Number(route.params.task_id)

 const loading = ref(false)
 const retrying = ref(false)
 const task = ref<ReviewTaskDTO | null>(null)
const findings = ref<ReviewFindingDTO[]>([])
const severityFilter = ref('')
const showAIPanel = ref(false)
const aiLoading = ref(false)
const aiPanelRef = ref<{
  addResponse: (content: string) => void
} | null>(null)

 const quickActions: QuickAction[] = [
   { key: 'analyze', label: '深度分析', prompt: '请对本次代码审查进行深度分析，总结主要问题并提供改进建议。' },
   { key: 'risk', label: '风险评估', prompt: '请评估本次代码变更的整体风险等级，重点关注安全漏洞、性能问题和兼容性影响。' },
   { key: 'suggest', label: '优化建议', prompt: '请针对发现的问题提供具体的代码优化建议和最佳实践。' }
 ]

const filteredFindings = computed(() => {
  if (!severityFilter.value) return findings.value
  return findings.value.filter(f => f.severity === severityFilter.value)
})

function countBySeverity(s: string) {
  return findings.value.filter(f => f.severity === s).length
}

function severityText(s: string) { const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }; return m[s] || s }
function statusLabel(s: string) { const m: Record<string, string> = { pending: '等待中', running: '运行中', success: '通过', failed: '失败', blocked: '已阻止' }; return m[s] || s }
function riskLabel(r: string) { const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中等', low: '低危', info: '提示' }; return m[r] || r }
function triggerLabel(t: string) { const m: Record<string, string> = { manual: '手动', webhook: 'Webhook', api: 'API' }; return m[t] || t }
function shortSha(sha: string) { return sha ? sha.substring(0, 7) : '—' }
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

  function statusVariant(s: string): StatusVariant { const m: Record<string, StatusVariant> = { pending: 'warning', running: 'running', success: 'success', failed: 'danger', blocked: 'danger' }; return m[s] || 'default' }

  function riskVariant(r: string): StatusVariant { const m: Record<string, StatusVariant> = { critical: 'danger', high: 'warning', medium: 'warning', low: 'info', info: 'default' }; return m[r] || 'default' }

  function severityVariant(s: string): StatusVariant { const m: Record<string, StatusVariant> = { critical: 'danger', high: 'warning', medium: 'warning', low: 'info', info: 'default' }; return m[s] || 'default' }

async function loadData() {
  loading.value = true
  try {
    const [t, f] = await Promise.all([getReviewTask(task_id), listReviewFindings(task_id)])
    task.value = t
    findings.value = f || []
  } catch (e) { console.error(e) } finally { loading.value = false }
}

 async function handleRetry() {
   retrying.value = true
   try {
     await retryReviewTask(task_id)
     ElMessage.success('审查任务已重新开始')
     loadData()
   } catch (e) { console.error(e) } finally { retrying.value = false }
 }

function formatReviewFindings(title: string, items: Array<{ message: string; file_path?: string; start_line?: number; suggestion?: string }> = []) {
  if (items.length === 0) return ''
  return [
    title,
    ...items.map((item, index) => {
      const location = item.file_path ? ` (${item.file_path}${item.start_line ? `:${item.start_line}` : ''})` : ''
      const suggestion = item.suggestion ? `\n   建议：${item.suggestion}` : ''
      return `${index + 1}. ${item.message}${location}${suggestion}`
    }),
  ].join('\n')
}

function formatReviewAIResponse(response: Awaited<ReturnType<typeof aiApi.code_review>>) {
  const sections = [response.summary]
  if (response.risk_level) {
    sections.push(``, `风险等级：${response.risk_level}`)
  }
  const blocking = formatReviewFindings('阻断问题：', response.blocking || [])
  const high = formatReviewFindings('高风险问题：', response.high_risk || [])
  const optional = formatReviewFindings('可选改进：', response.optional || [])
  for (const section of [blocking, high, optional]) {
    if (section) {
      sections.push('', section)
    }
  }
  if (response.merge_notes) {
    sections.push('', `合并前说明：${response.merge_notes}`)
  }
  return sections.join('\n')
}

 async function handleAIReview(message: string) {
   aiLoading.value = true
   try {
      const response = await aiApi.review_summary({
        repo_key,
        task_id: String(task_id),
        task_status: task.value?.status || 'unknown',
       findings: findings.value.map(f => ({
         severity: f.severity,
         file_path: f.file_path || 'unknown',
         title: f.title,
         message: f.message,
         rule_id: f.rule_id || f.id.toString(),
       })),
       changed_files: [...new Set(findings.value.map(f => f.file_path).filter(Boolean))] as string[],
       user_instruction: message,
     })

     const contextInfo = `> 📝 **分析对象：** 审查任务 #${task_id} (${task.value?.status || 'unknown'})`
     aiPanelRef.value?.addResponse(contextInfo + '\n\n' + formatReviewAIResponse(response))
   } catch (e) {
     aiPanelRef.value?.addResponse('AI 审查分析失败，请稍后重试。')
     ElMessage.error('AI 审查分析失败，请稍后重试')
   } finally {
     aiLoading.value = false
   }
 }

 onMounted(loadData)
</script>

<style scoped>
.review-page-wrapper { min-height: 100%; }
.header-bar { padding: 16px 24px; border-bottom: 1px solid var(--el-border-color-lighter); }
.detail-layout { padding: 20px 24px; }
.detail-content { max-width: 900px; }
.breadcrumb { font-size: 13px; margin-bottom: 16px; display: flex; align-items: center; gap: 8px; }
.breadcrumb a { color: var(--el-color-primary); text-decoration: none; }
.breadcrumb .sep { color: #ccc; }
.breadcrumb .current { font-weight: 600; }
.title-row { margin-bottom: 12px; }
.repo-label { font-size: 14px; color: var(--el-text-color-secondary); font-family: 'IBM Plex Mono', monospace; margin-bottom: 4px; }
.title-row h2 { font-size: 22px; font-weight: 700; margin: 0; }
.meta-row { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
.meta-text { font-size: 12px; color: var(--el-text-color-secondary); font-family: 'Geist Mono', monospace; }
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding: 16px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  margin-bottom: 20px;
}
.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.info-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.info-value {
  font-size: 14px;
  color: var(--el-text-color-primary);
}
.info-value.mono {
  font-family: 'Geist Mono', monospace;
}
.summary-card { background: #0A0A0A; color: #fff; border-radius: 8px; padding: 20px; margin-bottom: 20px; }
.summary-card p { margin: 8px 0 0; font-size: 14px; color: #ccc; line-height: 1.6; }
.summary-stats {
  display: flex;
  gap: 16px;
  margin-top: 12px;
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
.error-card { background: #FEF2F2; border-radius: 8px; padding: 16px; margin-bottom: 20px; }
.error-card p { margin: 8px 0 0; font-size: 13px; color: #B91C1C; }
.findings-section { margin-top: 24px; }
.findings-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; flex-wrap: wrap; gap: 12px; }
.filter-pills { display: flex; gap: 8px; flex-wrap: wrap; }
.finding-cards { display: flex; flex-direction: column; gap: 12px; }
.finding-card { padding: 16px; border-radius: 8px; background: var(--el-fill-color-light); border: 1px solid var(--el-border-color-lighter); border-left: 3px solid var(--el-border-color); }
.finding-critical { border-left-color: #DC2626; }
.finding-high { border-left-color: #EA580C; }
.finding-medium { border-left-color: #D97706; }
.finding-low { border-left-color: #2563EB; }
.finding-info { border-left-color: #9CA3AF; }
.finding-head { margin-bottom: 8px; }
.finding-badges { display: flex; gap: 8px; align-items: center; }
.rule-badge { font-size: 11px; font-family: 'Geist Mono', monospace; color: var(--el-text-color-secondary); }
.source-badge { font-size: 11px; padding: 1px 6px; border-radius: 3px; background: var(--el-color-primary-light-9); color: var(--el-color-primary); }
.finding-title { font-size: 15px; font-weight: 600; margin-bottom: 6px; }
.finding-file { font-size: 12px; font-family: 'IBM Plex Mono', monospace; color: var(--el-color-primary); margin-bottom: 8px; }
.finding-message { font-size: 13px; color: var(--el-text-color-secondary); line-height: 1.5; margin-bottom: 8px; }
.finding-suggestion { font-size: 12px; padding: 8px 12px; background: #F0FDF4; border-radius: 4px; color: #166534; line-height: 1.4; }
.suggestion-label { font-weight: 600; }
</style>
