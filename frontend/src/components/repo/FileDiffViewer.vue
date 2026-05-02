<template>
  <div class="fe-main">
    <template v-if="selectedFile && (diffText || diffFile?.isBinary)">
      <div class="fe-diff-header">
        <span class="fe-diff-filename">{{ selectedFile }}</span>
        <div v-if="diffFile && !diffFile.isBinary" class="fe-diff-stats">
          <span class="fe-additions">+{{ diffFile.additions }}</span>
          <span class="fe-deletions">-{{ diffFile.deletions }}</span>
        </div>
      </div>
      <div v-if="diffFile?.isBinary" class="fe-binary-notice">
        <el-icon><Warning /></el-icon> 二进制文件，无法显示差异
      </div>
      <div v-else class="fe-diff-body">
        <table class="fe-diff-table">
          <tbody>
            <template v-for="(line, idx) in parsedLines" :key="idx">
              <tr v-if="line.type === 'hunk'" class="fe-dl-hunk">
                <td class="fe-dl-num" colspan="3">{{ line.content }}</td>
              </tr>
              <tr v-else :class="'fe-dl-' + line.type">
                <td class="fe-dl-num fe-dl-old">{{ line.oldNum }}</td>
                <td class="fe-dl-num fe-dl-new">{{ line.newNum }}</td>
                <td class="fe-dl-content"><pre>{{ line.content }}</pre></td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </template>

    <template v-else-if="selectedFile && blob">
      <div class="fe-diff-header">
        <span class="fe-diff-filename">{{ selectedFile }}</span>
        <span class="fe-blob-size">{{ formatSize(blob.size) }}</span>
      </div>
      <div class="fe-blob-body">
        <div v-if="blob.is_binary" class="fe-binary-notice">
          <el-icon><Warning /></el-icon> 二进制文件，无法预览
        </div>
        <pre v-else class="fe-blob-pre">{{ blob.content }}</pre>
      </div>
    </template>

    <div v-else-if="loading" class="fe-empty">
      <el-icon :size="24" class="is-loading"><Refresh /></el-icon>
    </div>

    <div v-else class="fe-empty">
      <el-icon :size="48" color="var(--text-color-secondary)"><Document /></el-icon>
      <span>选择文件查看内容</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Document, Warning, Refresh } from '@element-plus/icons-vue'
import type { BlobContent } from '@/api/modules/file'
import type { WorkspaceDiffFile } from '@/api/modules/workspace'

interface DiffLine {
  type: 'hunk' | 'add' | 'del' | 'ctx'
  content: string
  oldNum: string | number
  newNum: string | number
}

const props = defineProps<{
  selectedFile: string
  diffText: string
  diffFile: WorkspaceDiffFile | null
  blob: BlobContent | null
  loading: boolean
}>()

function parseDiffLines(text: string): DiffLine[] {
  const lines = text.split('\n')
  const result: DiffLine[] = []
  let oldNum = 0
  let newNum = 0
  for (const line of lines) {
    if (line.startsWith('@@')) {
      const m = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/)
      if (m) {
        oldNum = parseInt(m[1])
        newNum = parseInt(m[2])
      }
      result.push({ type: 'hunk', content: line, oldNum: '', newNum: '' })
    } else if (line.startsWith('+')) {
      result.push({ type: 'add', content: line.slice(1), oldNum: '', newNum: newNum++ })
    } else if (line.startsWith('-')) {
      result.push({ type: 'del', content: line.slice(1), oldNum: oldNum++, newNum: '' })
    } else if (line.startsWith(' ') || line === '') {
      const content = line.startsWith(' ') ? line.slice(1) : ''
      result.push({ type: 'ctx', content, oldNum: oldNum++, newNum: newNum++ })
    }
  }
  return result
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

const parsedLines = computed(() => {
  if (!props.diffText) return [] as DiffLine[]
  return parseDiffLines(props.diffText)
})
</script>

<style scoped>
.fe-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.fe-diff-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-color-page);
  flex-shrink: 0;
}

.fe-diff-filename {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fe-diff-stats {
  display: flex;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.fe-additions { color: var(--success-color); }
.fe-deletions { color: var(--danger-color); }

.fe-blob-size {
  font-size: 12px;
  color: var(--text-color-secondary);
  flex-shrink: 0;
}

.fe-diff-body {
  flex: 1;
  overflow: auto;
  background: var(--bg-color-page);
}

.fe-diff-table {
  width: 100%;
  border-collapse: collapse;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}

.fe-diff-table .fe-dl-num {
  width: 50px;
  min-width: 50px;
  max-width: 50px;
  padding: 0 8px;
  text-align: right;
  color: var(--diff-num-color);
  background: var(--diff-num-bg);
  border-right: 1px solid var(--diff-num-border);
  user-select: none;
  vertical-align: top;
  font-size: 12px;
  line-height: 1.5;
}

.fe-diff-table .fe-dl-content {
  padding: 0 12px;
  vertical-align: top;
}

.fe-diff-table .fe-dl-content pre {
  margin: 0;
  font-family: inherit;
  font-size: inherit;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.fe-diff-table .fe-dl-hunk {
  background: var(--diff-hunk-bg);
  color: var(--diff-hunk-color);
}

.fe-diff-table .fe-dl-hunk .fe-dl-num {
  background: var(--diff-hunk-bg);
}

.fe-diff-table .fe-dl-add {
  background: var(--diff-add-bg);
}

.fe-diff-table .fe-dl-add .fe-dl-num {
  background: var(--diff-add-num-bg);
}

.fe-diff-table .fe-dl-add .fe-dl-content {
  color: var(--diff-add-marker-color);
}

.fe-diff-table .fe-dl-del {
  background: var(--diff-del-bg);
}

.fe-diff-table .fe-dl-del .fe-dl-num {
  background: var(--diff-del-num-bg);
}

.fe-diff-table .fe-dl-del .fe-dl-content {
  color: var(--diff-del-marker-color);
}

.fe-diff-table .fe-dl-ctx .fe-dl-content {
  color: var(--diff-ctx-color);
}

.fe-blob-body {
  flex: 1;
  overflow: auto;
}

.fe-blob-pre {
  margin: 0;
  padding: 12px 16px;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
}

.fe-binary-notice {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-color-secondary);
  font-size: 14px;
}

.fe-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-color-secondary);
  font-size: 14px;
}
</style>
