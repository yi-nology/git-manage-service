<template>
  <div class="cr-review-detail">
    <div class="detail-header">
      <div class="header-left">
        <button class="back-btn" @click="$emit('close')">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h3>MR !{{ currentTask.mr_iid }} 审查详情</h3>
        <StatusBadge :variant="statusVariant(currentTask.status)" :text="statusLabel(currentTask.status)" />
        <StatusBadge v-if="currentTask.risk_level" :variant="riskVariant(currentTask.risk_level)" :text="riskLabel(currentTask.risk_level)" :showDot="false" />
      </div>
      <div class="header-actions">
        <ActionPill variant="outline" small :icon="Refresh" :disabled="retrying" @click="handleRetry">{{ retrying ? '重试中...' : '重新审查' }}</ActionPill>
        <ActionPill variant="outline" small :icon="Refresh" :disabled="loading" @click="loadFindings">刷新</ActionPill>
      </div>
    </div>

    <div class="meta-info">
      <div class="meta-item">
        <span class="meta-label">提交</span>
        <span class="meta-value mono">{{ shortSha(currentTask.commit_sha) }}</span>
      </div>
      <div class="meta-item">
        <span class="meta-label">触发</span>
        <span class="meta-value">{{ triggerLabel(currentTask.trigger_type) }}</span>
      </div>
      <div class="meta-item">
        <span class="meta-label">时间</span>
        <span class="meta-value">{{ timeAgo(currentTask.created_at) }}</span>
      </div>
    </div>

    <div class="summary-section" v-if="currentTask.summary">
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

    <div class="diff-section" v-if="currentTask.raw_diff && parsedDiffFiles.length > 0">
      <div class="diff-section-header">
        <SectionTitle title="代码变更与问题标注" />
        <span class="findings-summary-tag" v-if="findings.length > 0">
          {{ findings.length }} 个问题
          <template v-if="findings.filter(f => f.source === 'llm').length > 0">
            (AI {{ findings.filter(f => f.source === 'llm').length }})
          </template>
        </span>
      </div>
      <div class="global-findings" v-if="globalFindings().length > 0">
        <div v-for="(f, fiIdx) in globalFindings()" :key="'global-'+fiIdx" class="review-comment" :class="'severity-' + f.severity">
          <div class="comment-header">
            <span class="comment-severity-dot"></span>
            <span class="comment-severity-label">{{ severityText(f.severity) }}</span>
            <span class="comment-source">{{ f.source === 'llm' ? 'AI 审查' : '规则检查' }}</span>
            <span class="comment-rule">{{ f.rule_id }}</span>
          </div>
          <div class="comment-body">
            <div class="comment-title">{{ f.title }}</div>
            <div class="comment-message" v-if="f.message">{{ f.message }}</div>
          </div>
        </div>
      </div>
      <div class="diff-file-list">
        <div v-for="(file, fIdx) in displayFiles" :key="fIdx" class="diff-file-card" :class="{ 'has-findings': fileFindings(file.filePath).length > 0 }">
          <div class="diff-file-header" @click="toggleFile(fIdx)">
            <div class="file-header-left">
              <svg class="file-collapse-icon" :class="{ collapsed: !isFileExpanded(fIdx) }" width="12" height="12" viewBox="0 0 12 12" fill="none">
                <path d="M4 2L8 6L4 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              <span class="file-icon">{{ fileIcon(file.filePath) }}</span>
              <span class="file-path-text">{{ file.filePath }}</span>
            </div>
            <div class="file-header-right">
              <span class="file-stat add" v-if="fileAddCount(file) > 0">+{{ fileAddCount(file) }}</span>
              <span class="file-stat del" v-if="fileDelCount(file) > 0">-{{ fileDelCount(file) }}</span>
              <span class="file-finding-badge" v-if="fileFindings(file.filePath).length > 0">
                {{ fileFindings(file.filePath).length }} 个问题
              </span>
              <button v-if="fileFindings(file.filePath).length > 0" class="toggle-diff-btn" @click.stop="toggleFullDiff(fIdx)">
                {{ showFullDiff(fIdx) ? '仅看问题' : '完整 Diff' }}
              </button>
            </div>
          </div>
          <div class="diff-file-body" v-show="isFileExpanded(fIdx)">
            <table class="diff-table" v-if="getVisibleLines(fIdx).lines.length > 0">
              <tbody>
                <template v-for="(line, lIdx) in getVisibleLines(fIdx).lines" :key="lIdx">
                  <tr v-if="line.type === 'hunk' && line.content === '...'" class="dl-hunk dl-ellipsis">
                    <td class="dl-num" colspan="3">
                      <span class="ellipsis-icon">⋯</span>
                    </td>
                  </tr>
                  <tr v-else-if="line.type === 'hunk'" class="dl-hunk">
                    <td class="dl-num" colspan="3">{{ line.content }}</td>
                  </tr>
                  <tr v-else :class="['dl-row', 'dl-' + line.type, { 'dl-flagged': fmapGet(getVisibleLines(fIdx).fmap, line.newNum).length > 0 }]">
                    <td class="dl-num dl-num-old">{{ line.oldNum || '' }}</td>
                    <td class="dl-num dl-num-new">{{ line.newNum || '' }}</td>
                    <td class="dl-code"><pre>{{ line.content }}</pre></td>
                  </tr>
                  <template v-if="line.type !== 'hunk' && line.content !== '...' && fmapGet(getVisibleLines(fIdx).fmap, line.newNum).length > 0">
                    <tr v-for="(f, fiIdx) in fmapGet(getVisibleLines(fIdx).fmap, line.newNum)" :key="'f-'+lIdx+'-'+fiIdx" class="dl-comment-row">
                      <td class="dl-num"></td>
                      <td class="dl-num"></td>
                      <td class="dl-comment-cell">
                        <div class="gh-comment" :class="'gh-' + f.severity">
                          <div class="gh-comment-header">
                            <span class="gh-avatar">AI</span>
                            <span class="gh-comment-meta">
                              <strong>{{ f.source === 'llm' ? 'AI Code Reviewer' : 'Rule Engine' }}</strong>
                              <span class="gh-severity-tag" :class="'tag-' + f.severity">{{ severityText(f.severity) }}</span>
                            </span>
                          </div>
                          <div class="gh-comment-body">
                            <div class="gh-comment-title">{{ f.title }}</div>
                            <div class="gh-comment-desc" v-if="f.message">{{ f.message }}</div>
                          </div>
                          <div class="gh-comment-suggestion" v-if="f.suggestion">
                            <div class="gh-suggestion-header">
                              <svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M1.5 14h13l-1.5-5H3L1.5 14zM3 6l2-4 2 4M8 6l2-4 2 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                              Suggestion
                            </div>
                            <div class="gh-suggestion-body">{{ f.suggestion }}</div>
                          </div>
                        </div>
                      </td>
                    </tr>
                  </template>
                </template>
              </tbody>
            </table>
            <div class="file-level-findings" v-if="fileLevelFindings(file.filePath).length > 0">
              <div v-for="(f, fiIdx) in fileLevelFindings(file.filePath)" :key="'fl-'+fIdx+'-'+fiIdx" class="gh-comment gh-file-level" :class="'gh-' + f.severity">
                <div class="gh-comment-header">
                  <span class="gh-avatar">{{ f.source === 'llm' ? 'AI' : 'RE' }}</span>
                  <span class="gh-comment-meta">
                    <strong>{{ f.source === 'llm' ? 'AI Code Reviewer' : 'Rule Engine' }}</strong>
                    <span class="gh-severity-tag" :class="'tag-' + f.severity">{{ severityText(f.severity) }}</span>
                  </span>
                </div>
                <div class="gh-comment-body">
                  <div class="gh-comment-title">{{ f.title }}</div>
                  <div class="gh-comment-desc" v-if="f.message">{{ f.message }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="show-more-files" v-if="displayFiles.length < parsedDiffFiles.length">
        <button class="show-more-btn" @click="showMoreFiles">
          显示更多文件 ({{ displayFiles.length }}/{{ parsedDiffFiles.length }})
        </button>
      </div>
    </div>

    <div class="error-section" v-if="currentTask.error_message">
      <div class="error-card">
        <span class="error-label">错误信息</span>
        <span class="error-text">{{ currentTask.error_message }}</span>
      </div>
    </div>

    <div class="findings-section">
      <div class="findings-header">
        <SectionTitle :title="`问题列表 (${filteredFindings.length})`" />
        <div class="filter-pills">
          <ActionPill small :variant="!severityFilter && !sourceFilter ? 'primary' : 'outline'" @click="severityFilter = ''; sourceFilter = ''">全部</ActionPill>
          <ActionPill small :variant="sourceFilter === 'llm' ? 'primary' : 'outline'" @click="sourceFilter = 'llm'; severityFilter = ''">AI 审查</ActionPill>
          <ActionPill small :variant="sourceFilter === 'rule' ? 'primary' : 'outline'" @click="sourceFilter = 'rule'; severityFilter = ''">规则</ActionPill>
          <ActionPill small :variant="severityFilter === 'critical' ? 'danger' : 'outline'" @click="severityFilter = 'critical'; sourceFilter = ''">严重</ActionPill>
          <ActionPill small :variant="severityFilter === 'high' ? 'amber' : 'outline'" @click="severityFilter = 'high'; sourceFilter = ''">高危</ActionPill>
          <ActionPill small :variant="severityFilter === 'medium' ? 'amber' : 'outline'" @click="severityFilter = 'medium'; sourceFilter = ''">中等</ActionPill>
          <ActionPill small :variant="severityFilter === 'low' ? 'primary' : 'outline'" @click="severityFilter = 'low'; sourceFilter = ''">低危</ActionPill>
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElIcon } from 'element-plus'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { getReviewTask, listReviewFindings, retryReviewTask, type ReviewTaskDTO, type ReviewFindingDTO } from '@/api/modules/review'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'

const props = defineProps<{
  task: ReviewTaskDTO
  repoOwner?: string
  repoName?: string
}>()

const emit = defineEmits<{
  close: []
  retried: [task: ReviewTaskDTO]
}>()

const loading = ref(false)
const retrying = ref(false)
const findings = ref<ReviewFindingDTO[]>([])
const severityFilter = ref('')
const sourceFilter = ref('')
const currentTask = ref<ReviewTaskDTO>({ ...props.task })
let pollTimer: ReturnType<typeof setInterval> | null = null

const renderedSummary = computed(() => {
  if (!currentTask.value.summary) return ''
  return marked.parse(currentTask.value.summary) as string
})

const filteredFindings = computed(() => {
  let result = findings.value
  if (severityFilter.value) result = result.filter(f => f.severity === severityFilter.value)
  if (sourceFilter.value) result = result.filter(f => f.source === sourceFilter.value)
  return result
})

interface DiffLine {
  type: 'hunk' | 'add' | 'del' | 'ctx'
  content: string
  oldNum: string | number
  newNum: string | number
}

interface DiffFile {
  filePath: string
  lines: DiffLine[]
}

const parsedDiffFiles = computed<DiffFile[]>(() => {
  if (!currentTask.value.raw_diff) return []
  const text = currentTask.value.raw_diff
  const files: DiffFile[] = []
  let currentFile: DiffFile | null = null
  let oldNum = 0
  let newNum = 0
  let currentPath = ''

  for (const line of text.split('\n')) {
    if (line.startsWith('diff --git')) {
      const m = line.match(/diff --git a\/(.+?) b\/(.+)/)
      if (m) currentPath = m[2] || ''
      if (currentFile) files.push(currentFile)
      currentFile = { filePath: currentPath, lines: [] }
      oldNum = 0
      newNum = 0
      continue
    }
    if (!currentFile) continue
    if (line.startsWith('@@')) {
      const m = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/)
      if (m) { oldNum = parseInt(m[1] || '0'); newNum = parseInt(m[2] || '0') }
      currentFile.lines.push({ type: 'hunk', content: line, oldNum: '', newNum: '' })
    } else if (line.startsWith('+')) {
      currentFile.lines.push({ type: 'add', content: line.slice(1), oldNum: '', newNum: newNum++ })
    } else if (line.startsWith('-')) {
      currentFile.lines.push({ type: 'del', content: line.slice(1), oldNum: oldNum++, newNum: '' })
    } else if (line.startsWith(' ') || line === '') {
      const content = line.startsWith(' ') ? line.slice(1) : ''
      currentFile.lines.push({ type: 'ctx', content, oldNum: oldNum++, newNum: newNum++ })
    } else if (line.startsWith('---') || line.startsWith('+++') || line.startsWith('index') || line.startsWith('new file') || line.startsWith('deleted file') || line.startsWith('rename')) {
      continue
    }
  }
  if (currentFile) files.push(currentFile)
  return files
})

function pathMatches(findingPath: string, diffPath: string): boolean {
  if (findingPath === diffPath) return true
  const fb = findingPath.split('/').pop() || ''
  const db = diffPath.split('/').pop() || ''
  if (fb && db && fb === db) return true
  if (diffPath.endsWith('/' + findingPath) || findingPath.endsWith('/' + diffPath)) return true
  return false
}

function fileFindings(filePath: string): ReviewFindingDTO[] {
  return findings.value.filter(f => f.file_path && pathMatches(f.file_path, filePath))
}

function fileLevelFindings(filePath: string): ReviewFindingDTO[] {
  return findings.value.filter(f => f.file_path && pathMatches(f.file_path, filePath) && (!f.new_line || f.new_line === 0))
}

function globalFindings(): ReviewFindingDTO[] {
  return findings.value.filter(f => !f.file_path)
}

function fmapGet(fmap: Map<number, ReviewFindingDTO[]>, lineNum: number | string): ReviewFindingDTO[] {
  if (!lineNum || lineNum === '') return []
  const n = typeof lineNum === 'string' ? parseInt(lineNum) : lineNum
  if (isNaN(n)) return []
  return fmap.get(n) || []
}

function buildFindingLineMap(file: DiffFile): Map<number, ReviewFindingDTO[]> {
  const ff = fileFindings(file.filePath).filter(f => f.new_line && f.new_line > 0)
  if (ff.length === 0) return new Map()

  const codeLines: { idx: number; newNum: number; content: string; type: string }[] = []
  for (let i = 0; i < file.lines.length; i++) {
    const l = file.lines[i]
    if (!l || l.type === 'hunk') continue
    const n = typeof l.newNum === 'number' ? l.newNum : parseInt(String(l.newNum))
    if (!isNaN(n) && n > 0) {
      codeLines.push({ idx: i, newNum: n, content: l.content.toLowerCase(), type: l.type })
    }
  }

  const result = new Map<number, ReviewFindingDTO[]>()
  for (const f of ff) {
    const ln = f.new_line
    const exact = codeLines.find(c => c.newNum === ln)
    if (exact) {
      const arr = result.get(exact.newNum) || []
      arr.push(f)
      result.set(exact.newNum, arr)
      continue
    }

    const matched = findBestCodeLine(f, codeLines)
    if (matched) {
      const arr = result.get(matched.newNum) || []
      arr.push(f)
      result.set(matched.newNum, arr)
    }
  }
  return result
}

function findBestCodeLine(
  f: ReviewFindingDTO,
  codeLines: { idx: number; newNum: number; content: string; type: string }[]
): { idx: number; newNum: number; content: string; type: string } | null {
  const keywords = extractKeywords(f.title + ' ' + f.message + ' ' + (f.suggestion || ''))
  if (keywords.length === 0) {
    return codeLines.length > 0 ? codeLines[0]! : null
  }

  let bestLine: typeof codeLines[0] | null = null
  let bestScore = -1

  for (const cl of codeLines) {
    if (cl.type === 'del') continue
    let score = 0
    for (const kw of keywords) {
      if (cl.content.includes(kw.toLowerCase())) score += kw.length
    }
    if (cl.type === 'add') score += 1
    if (score > bestScore) {
      bestScore = score
      bestLine = cl
    }
  }

  return bestLine
}

function extractKeywords(text: string): string[] {
  const stopWords = new Set([
    'the', 'a', 'an', 'is', 'are', 'was', 'were', 'be', 'been', 'being',
    'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could',
    'should', 'may', 'might', 'can', 'shall', 'to', 'of', 'in', 'for',
    'on', 'with', 'at', 'by', 'from', 'as', 'into', 'through', 'during',
    'before', 'after', 'above', 'below', 'between', 'out', 'off', 'over',
    'under', 'again', 'further', 'then', 'once', 'here', 'there', 'when',
    'where', 'why', 'how', 'all', 'both', 'each', 'few', 'more', 'most',
    'other', 'some', 'such', 'no', 'nor', 'not', 'only', 'own', 'same',
    'so', 'than', 'too', 'very', 'just', 'because', 'but', 'and', 'or',
    'if', 'while', 'this', 'that', 'these', 'those', 'it', 'its',
    'code', 'file', 'line', 'use', 'used', 'using', 'also', 'which',
    'about', 'up', 'like', 'what', 'get', 'set', 'new', 'add', 'make',
  ])
  const words = text.match(/[a-zA-Z_][a-zA-Z0-9_.]{2,}/g) || []
  const unique = [...new Set(words.filter(w => !stopWords.has(w.toLowerCase())))]
  return unique.filter(w => w.length >= 3).slice(0, 15)
}

const expandedFiles = ref<Record<number, boolean>>({})
const fullDiffFiles = ref<Record<number, boolean>>({})
const visibleFileLimit = ref(20)

function toggleFullDiff(idx: number) {
  fullDiffFiles.value[idx] = !fullDiffFiles.value[idx]
}

function showFullDiff(idx: number): boolean {
  return !!fullDiffFiles.value[idx]
}

function toggleFile(idx: number) {
  expandedFiles.value[idx] = !isFileExpanded(idx)
}

function isFileExpanded(idx: number): boolean {
  if (expandedFiles.value[idx] !== undefined) return expandedFiles.value[idx]
  const files = parsedDiffFiles.value
  if (idx < files.length && fileFindings(files[idx]!.filePath).length > 0) return true
  return false
}

const CONTEXT_LINES = 3

interface VisibleLinesResult {
  lines: DiffFile['lines']
  fmap: Map<number, ReviewFindingDTO[]>
}

const fileVisibleCache = computed<Map<number, VisibleLinesResult>>(() => {
  const cache = new Map<number, VisibleLinesResult>()
  const files = displayFiles.value
  for (let i = 0; i < files.length; i++) {
    cache.set(i, computeVisibleLines(files[i]!, i))
  }
  return cache
})

function computeVisibleLines(file: DiffFile, fIdx: number): VisibleLinesResult {
  const fmap = buildFindingLineMap(file)
  if (showFullDiff(fIdx)) return { lines: file.lines, fmap }
  if (fmap.size === 0) return { lines: file.lines, fmap }
  const flaggedLines = new Set<number>()
  for (const [lineNum] of fmap) flaggedLines.add(lineNum)
  const showIdx = new Set<number>()
  for (let i = 0; i < file.lines.length; i++) {
    const ln = file.lines[i]
    if (!ln || ln.type === 'hunk') continue
    const n = typeof ln.newNum === 'number' ? ln.newNum : parseInt(String(ln.newNum))
    if (!isNaN(n) && flaggedLines.has(n)) {
      let start = i - CONTEXT_LINES
      let foundHunk = false
      for (let j = i - 1; j >= Math.max(0, start); j--) {
        if (file.lines[j]?.type === 'hunk') { start = j; foundHunk = true; break }
      }
      if (!foundHunk && i > 0 && file.lines[0]?.type === 'hunk') start = 0
      const end = Math.min(file.lines.length - 1, i + CONTEXT_LINES)
      for (let k = Math.max(0, start); k <= end; k++) showIdx.add(k)
    }
  }
  const result: DiffFile['lines'] = []
  let lastShown = -1
  const sorted = [...showIdx].sort((a, b) => a - b)
  for (const idx of sorted) {
    if (lastShown >= 0 && idx > lastShown + 1) {
      result.push({ type: 'hunk' as const, content: '...', oldNum: '', newNum: '' })
    }
    result.push(file.lines[idx]!)
    lastShown = idx
  }
  return { lines: result, fmap }
}

function getVisibleLines(fIdx: number): VisibleLinesResult {
  return fileVisibleCache.value.get(fIdx) || { lines: [], fmap: new Map() }
}

function showMoreFiles() {
  visibleFileLimit.value += 20
}

const displayFiles = computed(() => {
  const all = parsedDiffFiles.value
  if (all.length <= visibleFileLimit.value) return all
  const withFindings = all.filter(f => fileFindings(f.filePath).length > 0)
  const without = all.filter(f => fileFindings(f.filePath).length === 0)
  const shown = [...withFindings, ...without].slice(0, visibleFileLimit.value)
  return shown
})

function fileIcon(path: string): string {
  if (path.endsWith('.go')) return '🔷'
  if (path.endsWith('.ts') || path.endsWith('.tsx')) return '🔷'
  if (path.endsWith('.js') || path.endsWith('.jsx')) return '🟨'
  if (path.endsWith('.vue')) return '💚'
  if (path.endsWith('.py')) return '🐍'
  if (path.endsWith('.rs')) return '🦀'
  if (path.endsWith('.css') || path.endsWith('.scss')) return '🎨'
  if (path.endsWith('.html')) return '📄'
  if (path.endsWith('.json') || path.endsWith('.yaml') || path.endsWith('.yml')) return '📋'
  if (path.endsWith('.md')) return '📝'
  if (path.endsWith('.sql')) return '🗃️'
  return '📄'
}

function fileAddCount(file: DiffFile): number {
  return file.lines.filter(l => l.type === 'add').length
}

function fileDelCount(file: DiffFile): number {
  return file.lines.filter(l => l.type === 'del').length
}

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
    const [taskRes, findingsRes] = await Promise.all([
      getReviewTask(currentTask.value.id),
      listReviewFindings(currentTask.value.id),
    ])
    if (taskRes) currentTask.value = taskRes
    findings.value = findingsRes || []
  } catch {
    findings.value = []
  } finally {
    loading.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(async () => {
    if (currentTask.value.status === 'pending' || currentTask.value.status === 'running') {
      try {
        const taskRes = await getReviewTask(currentTask.value.id)
        if (taskRes) currentTask.value = taskRes
        if (taskRes && taskRes.status !== 'pending' && taskRes.status !== 'running') {
          const findingsRes = await listReviewFindings(currentTask.value.id)
          findings.value = findingsRes || []
        }
      } catch { /* ignore */ }
    }
  }, 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function handleRetry() {
  retrying.value = true
  try {
    const data: { owner?: string; repo?: string } = {}
    if (props.repoOwner) data.owner = props.repoOwner
    if (props.repoName) data.repo = props.repoName
    const updated = await retryReviewTask(currentTask.value.id, data)
    if (updated) currentTask.value = updated
    ElMessage.success('已重新触发审查')
    emit('retried', currentTask.value)
    startPolling()
  } catch (e: any) {
    ElMessage.error('重试失败: ' + (e?.message || ''))
  } finally {
    retrying.value = false
  }
}

onMounted(() => {
  loadFindings()
  startPolling()
})

onUnmounted(stopPolling)
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

.error-section { margin-top: 0; }

.error-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 16px;
  background: #FEF2F2;
  border: 1px solid #FECACA;
  border-radius: 8px;
  border-left: 3px solid #DC2626;
}

.error-label { font-size: 12px; font-weight: 600; color: #991B1B; }
.error-text { font-size: 13px; color: #7F1D1D; line-height: 1.5; word-break: break-all; }

.diff-section { margin-top: 8px; }

.diff-section-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.findings-summary-tag {
  font-size: 13px;
  color: #57606A;
  background: #F6F8FA;
  border: 1px solid #D0D7DE;
  border-radius: 12px;
  padding: 2px 10px;
}

.show-more-files {
  text-align: center;
  padding: 12px;
}

.show-more-btn {
  background: #F6F8FA;
  border: 1px solid #D0D7DE;
  border-radius: 6px;
  padding: 6px 16px;
  color: #0969DA;
  cursor: pointer;
  font-size: 13px;
}

.show-more-btn:hover {
  background: #EAEFF2;
}

.diff-file-list { display: flex; flex-direction: column; gap: 16px; }

.diff-file-card {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  overflow: hidden;
  background: #fff;
}

.diff-file-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #F6F8FA;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  user-select: none;
  transition: background 0.15s;
}

.diff-file-header:hover { background: #EDF0F3; }

.file-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.file-collapse-icon {
  color: #6B7280;
  transition: transform 0.15s;
  flex-shrink: 0;
}

.file-collapse-icon.collapsed { transform: rotate(0deg); }
.file-collapse-icon:not(.collapsed) { transform: rotate(90deg); }

.file-icon { font-size: 14px; flex-shrink: 0; }

.file-path-text {
  font-size: 13px;
  font-weight: 600;
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-header-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }

.file-stat {
  font-size: 11px;
  font-weight: 600;
  font-family: 'SF Mono', monospace;
}
.file-stat.add { color: #1A7F37; }
.file-stat.del { color: #CF222E; }

.file-finding-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: #FFF1E5;
  color: #BC4C00;
  font-weight: 500;
}

.diff-file-body { overflow-x: auto; }

.global-findings {
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.file-level-findings {
  display: flex;
  flex-direction: column;
  gap: 6px;
  border-bottom: 1px solid #D0D7DE;
  margin-bottom: 4px;
}

.diff-table {
  width: 100%;
  border-collapse: collapse;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  line-height: 20px;
}

.diff-table .dl-num {
  width: 50px;
  min-width: 50px;
  max-width: 50px;
  padding: 0 8px;
  text-align: right;
  color: rgba(27,31,35,0.35);
  background: #F6F8FA;
  border-right: 1px solid #EFF2F5;
  user-select: none;
  vertical-align: top;
  font-size: 11px;
}

.diff-table .dl-code {
  padding: 0 12px;
  vertical-align: top;
}

.diff-table .dl-code pre {
  margin: 0;
  font-family: inherit;
  font-size: inherit;
  line-height: 20px;
  white-space: pre-wrap;
  word-break: break-all;
  color: #24292F;
}

.diff-table .dl-ellipsis {
  background: #F6F8FA;
  cursor: default;
}
.diff-table .dl-ellipsis .dl-num {
  background: #F6F8FA;
  color: #6E7781;
  text-align: center;
}
.ellipsis-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 24px;
  color: #6E7781;
  font-size: 16px;
  letter-spacing: 4px;
  user-select: none;
}

.diff-table .dl-hunk {
  background: var(--diff-hunk-bg, #F1F8FF);
}
.diff-table .dl-hunk .dl-num {
  background: var(--diff-hunk-bg, #F1F8FF);
  color: rgba(27,31,35,0.5);
}
.diff-table .dl-hunk .dl-code pre { color: rgba(27,31,35,0.6); }

.diff-table .dl-row.dl-add { background: var(--diff-add-bg, #E6FFED); }
.diff-table .dl-row.dl-add .dl-num { background: var(--diff-add-num-bg, #CDFFD8); }
.diff-table .dl-row.dl-add .dl-code pre { color: #116329; }

.diff-table .dl-row.dl-del { background: var(--diff-del-bg, #FFEEF0); }
.diff-table .dl-row.dl-del .dl-num { background: var(--diff-del-num-bg, #FFD7D5); }
.diff-table .dl-row.dl-del .dl-code pre { color: #82071E; }

.diff-table .dl-row.dl-ctx .dl-code pre { color: #57606A; }

.diff-table .dl-row.dl-flagged {
  background: #FFF8C5 !important;
}
.diff-table .dl-row.dl-flagged .dl-num {
  background: #F5EEC8 !important;
}

.diff-table .dl-comment-row {
  background: #FFFFFF;
  border-top: 1px solid #D0D7DE;
  border-bottom: 1px solid #D0D7DE;
}

.diff-table .dl-comment-cell { padding: 0; border: none; }

.diff-table .dl-comment-row {
  background: #F6F8FA;
  border-left: 3px solid #D0D7DE;
}
.diff-table .dl-comment-cell { padding: 0; border: none; }

.gh-comment {
  padding: 16px;
  border-left: 3px solid #D0D7DE;
  background: #FFFFFF;
  margin: 0;
}
.gh-comment.gh-file-level { margin: 8px 0; border-radius: 6px; border: 1px solid #D0D7DE; border-left-width: 3px; }
.gh-comment.gh-critical { border-left-color: #CF222E; }
.gh-comment.gh-high { border-left-color: #BF8700; }
.gh-comment.gh-medium { border-left-color: #D29922; }
.gh-comment.gh-low { border-left-color: #0969DA; }
.gh-comment.gh-info { border-left-color: #6E7781; }

.gh-comment-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.gh-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #0969DA;
  color: #FFF;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}
.gh-comment-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.gh-comment-meta strong {
  font-size: 13px;
  color: #24292F;
}
.gh-severity-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 600;
  line-height: 1;
}
.gh-severity-tag.tag-critical { background: #FFEBE9; color: #CF222E; }
.gh-severity-tag.tag-high { background: #FFF8C5; color: #9A6700; }
.gh-severity-tag.tag-medium { background: #FFF8C5; color: #9A6700; }
.gh-severity-tag.tag-low { background: #DDF4FF; color: #0969DA; }
.gh-severity-tag.tag-info { background: #EFF2F5; color: #6E7781; }

.gh-comment-body { margin-bottom: 0; }
.gh-comment-title {
  font-size: 14px;
  font-weight: 600;
  color: #24292F;
  margin-bottom: 6px;
}
.gh-comment-desc {
  font-size: 13px;
  color: #57606A;
  line-height: 1.6;
  white-space: pre-wrap;
}

.gh-comment-suggestion {
  margin-top: 10px;
  border: 1px solid #D0D7DE;
  border-radius: 6px;
  overflow: hidden;
}
.gh-suggestion-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #F6F8FA;
  border-bottom: 1px solid #D0D7DE;
  font-size: 12px;
  font-weight: 600;
  color: #57606A;
}
.gh-suggestion-header svg { color: #1A7F37; }
.gh-suggestion-body {
  padding: 10px 12px;
  font-size: 13px;
  color: #1F2328;
  line-height: 1.6;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  white-space: pre-wrap;
  background: #F0FFF4;
}

.diff-file-card.has-findings {
  border-color: #BF8700;
  box-shadow: 0 0 0 1px #FFF8C5;
}

.toggle-diff-btn {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid #D0D7DE;
  background: #F6F8FA;
  color: #57606A;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}
.toggle-diff-btn:hover {
  background: #FFFFFF;
  border-color: #0969DA;
  color: #0969DA;
}
</style>
