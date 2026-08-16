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
import './diff-viewer.css'
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
