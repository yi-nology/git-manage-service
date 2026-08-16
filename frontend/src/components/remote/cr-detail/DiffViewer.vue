<template>
  <div class="diff-section" v-if="task.raw_diff && parsedDiffFiles.length > 0">
    <div class="diff-section-header">
      <SectionTitle title="代码变更与问题标注" />
      <span class="findings-summary-tag" v-if="findings.length > 0">
        {{ findings.length }} 个问题
        <template v-if="findings.filter(f => f.source === 'llm').length > 0">
          (AI {{ findings.filter(f => f.source === 'llm').length }})
        </template>
      </span>
    </div>
    <div class="global-findings" v-if="globalFindingsList.length > 0">
      <div v-for="(f, fiIdx) in globalFindingsList" :key="'global-'+fiIdx" class="review-comment" :class="'severity-' + f.severity">
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
      <div v-for="(file, fIdx) in displayFiles" :key="fIdx" class="diff-file-card" :class="{ 'has-findings': getFFindings(file.file_path).length > 0 }">
        <div class="diff-file-header" @click="toggleFile(fIdx)">
          <div class="file-header-left">
            <svg class="file-collapse-icon" :class="{ collapsed: !isFileExpanded(fIdx) }" width="12" height="12" viewBox="0 0 12 12" fill="none">
              <path d="M4 2L8 6L4 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <span class="file-icon">{{ fileIcon(file.file_path) }}</span>
            <span class="file-path-text">{{ file.file_path }}</span>
          </div>
          <div class="file-header-right">
            <span class="file-stat add" v-if="fileAddCount(file) > 0">+{{ fileAddCount(file) }}</span>
            <span class="file-stat del" v-if="fileDelCount(file) > 0">-{{ fileDelCount(file) }}</span>
            <span class="file-finding-badge" v-if="getFFindings(file.file_path).length > 0">
              {{ getFFindings(file.file_path).length }} 个问题
            </span>
            <button v-if="getFFindings(file.file_path).length > 0" class="toggle-diff-btn" @click.stop="toggleFullDiff(fIdx)">
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
          <div class="file-level-findings" v-if="getFileLevelFindings(file.file_path).length > 0">
            <div v-for="(f, fiIdx) in getFileLevelFindings(file.file_path)" :key="'fl-'+fIdx+'-'+fiIdx" class="gh-comment gh-file-level" :class="'gh-' + f.severity">
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
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ReviewTaskDTO, ReviewFindingDTO } from '@/api/modules/review'
import SectionTitle from '@/components/common/SectionTitle.vue'
import {
  parseDiff, fileFindings as getFindings, fileLevelFindings as getFileLevelFindingsList,
  globalFindings as getGlobalFindings, fmapGet, buildFindingLineMap, fileIcon,
  fileAddCount, fileDelCount, severityText,
  type DiffFile, type VisibleLinesResult,
} from './diff-utils'

const props = defineProps<{
  task: ReviewTaskDTO
  findings: ReviewFindingDTO[]
}>()

const expandedFiles = ref<Record<number, boolean>>({})
const fullDiffFiles = ref<Record<number, boolean>>({})
const visibleFileLimit = ref(20)

const parsedDiffFiles = computed<DiffFile[]>(() => parseDiff(props.task.raw_diff))

const globalFindingsList = computed(() => getGlobalFindings(props.findings))

function getFFindings(file_path: string): ReviewFindingDTO[] {
  return getFindings(props.findings, file_path)
}

function getFileLevelFindings(file_path: string): ReviewFindingDTO[] {
  return getFileLevelFindingsList(props.findings, file_path)
}

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
  if (idx < files.length && getFFindings(files[idx]!.file_path).length > 0) return true
  return false
}

const CONTEXT_LINES = 3

const fileVisibleCache = computed<Map<number, VisibleLinesResult>>(() => {
  const cache = new Map<number, VisibleLinesResult>()
  const files = displayFiles.value
  for (let i = 0; i < files.length; i++) {
    cache.set(i, computeVisibleLines(files[i]!, i))
  }
  return cache
})

function computeVisibleLines(file: DiffFile, fIdx: number): VisibleLinesResult {
  const fmap = buildFindingLineMap(file, props.findings)
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
  const withFindings = all.filter(f => getFFindings(f.file_path).length > 0)
  const without = all.filter(f => getFFindings(f.file_path).length === 0)
  const shown = [...withFindings, ...without].slice(0, visibleFileLimit.value)
  return shown
})
</script>

<style scoped>
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

.review-comment {
  padding: 12px 16px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  border-left: 3px solid var(--border-color);
}
.review-comment.severity-critical { border-left-color: #DC2626; }
.review-comment.severity-high { border-left-color: #EA580C; }
.review-comment.severity-medium { border-left-color: #D97706; }
.review-comment.severity-low { border-left-color: #2563EB; }

.comment-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.comment-severity-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.comment-severity-label {
  font-size: 12px;
  font-weight: 600;
}
.comment-source {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--accent-bg);
  color: var(--accent-primary);
}
.comment-rule {
  font-size: 11px;
  font-family: 'SF Mono', monospace;
  color: var(--text-color-secondary);
}

.comment-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}
.comment-message {
  font-size: 12px;
  color: var(--text-color-secondary);
  line-height: 1.5;
}
</style>
