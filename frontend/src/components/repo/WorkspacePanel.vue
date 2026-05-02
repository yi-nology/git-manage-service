<template>
  <div class="workspace-panel">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-tag v-if="ws.status?.branch" type="primary" effect="plain" class="branch-tag">
          <el-icon><svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M11 11L13 13V7H11M13 1C7.5 1 3 5.5 3 11C3 16.5 7.5 21 13 21C14.8 21 16.5 20.5 18 19.6L22.1 23.7L23.7 22.1L19.6 18C20.5 16.5 21 14.8 21 13C21 7.5 16.5 1 11 1Z"/></svg></el-icon>
          {{ ws.status?.branch }}
        </el-tag>
        <span v-if="ws.status?.ahead" class="ahead-badge">
          <el-icon><Top /></el-icon> {{ ws.status.ahead }}
        </span>
        <span v-if="ws.status?.behind" class="behind-badge">
          <el-icon><Bottom /></el-icon> {{ ws.status.behind }}
        </span>
        <el-tag v-if="ws.status && !ws.status.isClean" type="warning" size="small">有变更</el-tag>
        <el-tag v-if="ws.status?.isClean" type="success" size="small">干净</el-tag>
      </div>
      <div class="toolbar-right">
        <el-button @click="ws.handlePull()" :loading="ws.pulling.value" size="small">
          <el-icon><Download /></el-icon> 拉取
        </el-button>
        <el-button @click="ws.loadStatus()" :loading="ws.loading.value" size="small">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
      </div>
    </div>

    <div class="body">
      <div class="file-list">
        <div class="file-list-header">
          <span class="file-list-title">变更文件</span>
          <div class="file-list-badges">
            <el-tag v-if="ws.status?.staged.length" type="success" size="small">{{ ws.status.staged.length }} 暂存</el-tag>
            <el-tag v-if="ws.status?.unstaged.length" type="warning" size="small">{{ ws.status.unstaged.length }} 未暂存</el-tag>
          </div>
        </div>

        <div class="file-list-body" v-loading="ws.loading.value">
          <template v-if="ws.status?.staged.length">
            <div class="file-section-label">
              <span class="dot dot-green"></span> 已暂存
            </div>
            <div v-for="f in ws.status.staged" :key="'s-'+f.path"
                 class="file-item" :class="{ active: ws.selectedFile.value === f.path }"
                 @click="selectFile(f.path)">
              <el-icon><Document /></el-icon>
              <span class="file-name">{{ f.path }}</span>
              <el-tag size="small" :type="statusTagType(f.status)">{{ f.status[0].toUpperCase() }}</el-tag>
            </div>
          </template>

          <template v-if="ws.status?.unstaged.length">
            <div class="file-section-label">
              <span class="dot dot-orange"></span> 未暂存
            </div>
            <div v-for="f in ws.status.unstaged" :key="'u-'+f.path"
                 class="file-item" :class="{ active: ws.selectedFile.value === f.path }"
                 @click="selectFile(f.path)">
              <el-icon><Document /></el-icon>
              <span class="file-name">{{ f.path }}</span>
              <el-tag size="small" :type="statusTagType(f.status)">{{ f.status[0].toUpperCase() }}</el-tag>
            </div>
          </template>

          <template v-if="ws.status?.untracked.length">
            <div class="file-section-label">
              <span class="dot dot-gray"></span> 未跟踪
            </div>
            <div v-for="f in ws.status.untracked" :key="'t-'+f.path"
                 class="file-item" :class="{ active: ws.selectedFile.value === f.path }"
                 @click="selectFile(f.path)">
              <el-icon><Document /></el-icon>
              <span class="file-name">{{ f.path }}</span>
              <el-tag size="small" type="info">?</el-tag>
            </div>
          </template>

          <template v-if="ws.status?.conflicted.length">
            <div class="file-section-label">
              <span class="dot dot-red"></span> 冲突
            </div>
            <div v-for="f in ws.status.conflicted" :key="'c-'+f.path"
                 class="file-item conflict-item"
                 @click="openConflictResolver(f.path)">
              <el-icon color="var(--el-color-danger)"><WarningFilled /></el-icon>
              <span class="file-name">{{ f.path }}</span>
              <el-tag size="small" type="danger">C</el-tag>
              <el-button size="small" type="primary" @click.stop="aiResolveFile(f.path)">
                <el-icon><MagicStick /></el-icon>
              </el-button>
            </div>
          </template>

          <el-empty v-if="ws.status?.isClean" description="工作区干净" :image-size="60" />
        </div>
      </div>

      <div class="diff-area">
        <div class="diff-toolbar" v-if="ws.selectedFile.value">
          <span class="diff-file">{{ ws.selectedFile.value }}</span>
          <div class="diff-stats" v-if="currentDiffFile">
            <span class="additions">+{{ currentDiffFile.additions }}</span>
            <span class="deletions">-{{ currentDiffFile.deletions }}</span>
          </div>
        </div>
        <pre class="diff-content" v-if="currentDiffFile?.diff">{{ currentDiffFile.diff }}</pre>
        <el-empty v-else description="选择文件查看变更" :image-size="80" />
      </div>
    </div>

    <div class="commit-bar">
      <el-input v-model="commitMessage" placeholder="输入提交信息 (Ctrl+Enter 提交)..." @keydown.ctrl.enter="doCommit" />
      <el-button @click="showAuthorDialog = true" plain size="small">
        <el-icon><User /></el-icon> {{ authorName || '作者' }}
      </el-button>
      <div class="push-switch">
        <span class="push-label">推送</span>
        <el-switch v-model="pushAfterCommit" />
      </div>
      <el-button type="primary" @click="doCommit" :loading="ws.committing.value" :disabled="!commitMessage || ws.status?.isClean">
        <el-icon><Check /></el-icon> 提交
      </el-button>
    </div>

    <ConflictResolver
      v-if="showConflictResolver"
      :repo-key="repoKey"
      :file-path="conflictFile"
      @resolved="onConflictResolved"
      @close="showConflictResolver = false"
    />

    <el-dialog v-model="showAuthorDialog" title="作者信息" width="400px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="authorName" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="authorEmail" />
        </el-form-item>
      </el-form>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useWorkspace } from '@/composables/useWorkspace'
import ConflictResolver from './ConflictResolver.vue'
import { Document, Refresh, Download, Top, Bottom, Check, User, WarningFilled, MagicStick } from '@element-plus/icons-vue'

const props = defineProps<{ repoKey: string }>()

const {
  loading, status, diff, selectedFile, pulling, committing,
  loadStatus, loadDiff,
  handleStageAll, handleStageFile, handleUnstageFile, handleUnstageAll,
  handleCommit, handlePull,
  handleGetConflictDetail, handleResolveConflict, handleAIResolve,
} = useWorkspace(props.repoKey)

const commitMessage = ref('')
const pushAfterCommit = ref(false)
const authorName = ref('')
const authorEmail = ref('')
const showAuthorDialog = ref(false)
const showConflictResolver = ref(false)
const conflictFile = ref('')

const currentDiffFile = computed(() => {
  if (!diff.value?.files?.length) return null
  return diff.value.files.find(f => f.file === selectedFile.value) || diff.value.files[0]
})

function statusTagType(status: string) {
  switch (status) {
    case 'added': return 'success'
    case 'modified': return 'warning'
    case 'deleted': return 'danger'
    case 'renamed': return 'info'
    default: return 'info'
  }
}

function selectFile(path: string) {
  selectedFile.value = path
  loadDiff(path)
}

function openConflictResolver(path: string) {
  conflictFile.value = path
  showConflictResolver.value = true
}

async function aiResolveFile(path: string) {
  const detail = await handleGetConflictDetail(path)
  if (!detail) return
  const resolved = await handleAIResolve(path, detail.oursContent, detail.theirsContent, detail.baseContent)
  if (resolved) {
    await handleResolveConflict(path, resolved.resolvedContent)
  }
}

async function doCommit() {
  const result = await handleCommit(commitMessage.value, authorName.value, authorEmail.value, pushAfterCommit.value, '')
  if (result) {
    commitMessage.value = ''
  }
}

function onConflictResolved() {
  showConflictResolver.value = false
  conflictFile.value = ''
}

onMounted(() => {
  loadStatus()
})
</script>

<style scoped>
.workspace-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  gap: 12px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.branch-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.ahead-badge {
  display: flex;
  align-items: center;
  gap: 2px;
  color: var(--el-color-success);
  font-size: 12px;
  font-weight: 500;
}

.behind-badge {
  display: flex;
  align-items: center;
  gap: 2px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  font-weight: 500;
}

.body {
  display: flex;
  flex: 1;
  gap: 12px;
  min-height: 0;
}

.file-list {
  width: 280px;
  min-width: 280px;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}

.file-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.file-list-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.file-list-badges {
  display: flex;
  gap: 4px;
}

.file-list-body {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.file-section-label {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  font-size: 12px;
  font-weight: 600;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot-green { background: var(--el-color-success); }
.dot-orange { background: var(--el-color-warning); }
.dot-gray { background: var(--el-text-color-secondary); }
.dot-red { background: var(--el-color-danger); }

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  cursor: pointer;
  transition: background 0.15s;
}

.file-item:hover { background: var(--el-fill-color); }
.file-item.active { background: var(--el-color-primary-light-9); }

.conflict-item {
  background: var(--el-color-danger-light-9);
  border-left: 2px solid var(--el-color-danger);
}

.file-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.diff-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}

.diff-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.diff-file {
  font-size: 13px;
  font-weight: 600;
}

.diff-stats {
  display: flex;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
}

.additions { color: var(--el-color-success); }
.deletions { color: var(--el-color-danger); }

.diff-content {
  flex: 1;
  padding: 12px;
  margin: 0;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
  white-space: pre-wrap;
  background: var(--el-bg-color-page);
}

.commit-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.commit-bar .el-input {
  flex: 1;
}

.push-switch {
  display: flex;
  align-items: center;
  gap: 6px;
}

.push-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
