<template>
  <div class="file-explorer">
    <div class="fe-toolbar">
      <el-select v-model="currentRef" size="small" class="fe-ref-select">
        <el-option label="工作区" value="worktree" />
        <el-option-group label="分支">
          <el-option v-for="b in branches" :key="b" :label="b" :value="b" />
        </el-option-group>
        <el-option-group v-if="tags.length" label="标签">
          <el-option v-for="t in tags" :key="t" :label="t" :value="t" />
        </el-option-group>
      </el-select>

      <el-icon class="fe-home-btn" @click="treeApi.collapseAll()"><HomeFilled /></el-icon>

      <template v-if="isWorktree && wsStatus">
        <el-tag size="small" effect="plain">{{ wsStatus.branch }}</el-tag>
        <span v-if="wsStatus.ahead" class="fe-ahead">
          <el-icon><Top /></el-icon>{{ wsStatus.ahead }}
        </span>
        <span v-if="wsStatus.behind" class="fe-behind">
          <el-icon><Bottom /></el-icon>{{ wsStatus.behind }}
        </span>
        <el-tag v-if="!wsStatus.is_clean" size="small" type="warning">有变更</el-tag>
        <el-tag v-else size="small" type="success">干净</el-tag>
      </template>

      <div class="fe-spacer" />

      <el-button v-if="isWorktree" size="small" :loading="pulling" @click="wsApi.handlePull(refreshAll)">
        <el-icon><Download /></el-icon>
      </el-button>
      <el-button v-if="isWorktree" size="small" :loading="pushing" @click="wsApi.handlePush(refreshAll)">
        <el-icon><Upload /></el-icon>
      </el-button>
      <el-button size="small" :loading="treeLoading" @click="refreshAll">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>

    <div class="fe-content">
      <FileTree
        :is-worktree="isWorktree"
        :view-mode="viewMode"
        :selected-file="selectedFile"
        :flat-tree-items="flatTreeItems"
        :tree-loading="treeLoading"
        :expanded-dirs="expandedDirs"
        :ws-status="wsStatus"
        :get-file-status="wsApi.getFileStatus"
        :is-file-staged="wsApi.isFileStaged"
        :is-file-unstaged="wsApi.isFileUnstaged"
        :is-file-untracked="wsApi.isFileUntracked"
        :is-file-conflicted="wsApi.isFileConflicted"
        @update:view-mode="viewMode = $event"
        @toggle-dir="handleToggleDir"
        @select-tree-file="selectTreeFile"
        @select-changed-file="selectChangedFile"
        @stage-file="wsApi.stageFile($event, refreshAll)"
        @unstage-file="wsApi.unstageFile($event, refreshAll)"
        @stage-all-unstaged="wsApi.stageAllUnstaged(refreshAll)"
        @unstage-all="wsApi.unstageAllFiles(refreshAll)"
        @gitignore="wsApi.handleGitignore($event, refreshAll)"
        @file-action="(action: string, path: string) => wsApi.handleFileAction(action, path, refreshAll)"
        @open-conflict="openConflictResolver"
      />

      <FileDiffViewer
        :selected-file="selectedFile"
        :diff-text="currentDiffText"
        :diff-file="currentDiffFile"
        :blob="blobContent"
        :loading="blobLoading"
      />
    </div>

    <CommitBar
      v-if="isWorktree && hasChanges"
      :commit-msg="commitMsg"
      :selected-author="selectedAuthor"
      :push-after-commit="pushAfterCommit"
      :authors="authors"
      :committing="committing"
      :generating-msg="generatingMsg"
      @update:commit-msg="commitMsg = $event"
      @update:selected-author="selectedAuthor = $event"
      @update:push-after-commit="pushAfterCommit = $event"
      @commit="doCommit"
      @generate-msg="handleGenerateMsg"
    />

    <ConflictResolver
      v-if="showConflictResolver"
      :repo-key="repoKey"
      :file-path="conflictFile"
      @resolved="onConflictResolved"
      @close="showConflictResolver = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  HomeFilled, Warning, Refresh, Download, Upload,
  Top, Bottom,
} from '@element-plus/icons-vue'
import { getFileBlob, type BlobContent } from '@/api/modules/file'
import {
  commitChanges, generateCommitMessage,
  type WorkspaceDiff,
} from '@/api/modules/workspace'
import {
  listIdentities, getRepoAuthorConfig,
  type AuthorIdentityDTO,
} from '@/api/modules/author'
import { getBranchList, getTagList } from '@/api/modules/branch'

import ConflictResolver from './ConflictResolver.vue'
import FileTree from './FileTree.vue'
import FileDiffViewer from './FileDiffViewer.vue'
import CommitBar from './CommitBar.vue'
import { useFileTree } from '@/composables/useFileTree'
import { useWorkspaceStatus } from '@/composables/useWorkspaceStatus'

void Warning

const props = defineProps<{ repoKey: string }>()

const currentRef = ref('worktree')
const selectedFile = ref('')
const viewMode = ref<'all' | 'changes'>('all')

const blobContent = ref<BlobContent | null>(null)
const blobLoading = ref(false)
const diff = ref<WorkspaceDiff | null>(null)

const committing = ref(false)
const generatingMsg = ref(false)

const branches = ref<string[]>([])
const tags = ref<string[]>([])
const showConflictResolver = ref(false)
const conflictFile = ref('')

const commitMsg = ref('')
const selectedAuthor = ref<number | null>(null)
const pushAfterCommit = ref(false)
const authors = ref<AuthorIdentityDTO[]>([])

const treeApi = useFileTree(props.repoKey)
const wsApi = useWorkspaceStatus(props.repoKey)

const { flatTreeItems, treeLoading, expandedDirs } = treeApi
const { wsStatus, pulling, pushing, hasChanges } = wsApi

const isWorktree = computed(() => currentRef.value === 'worktree')

const activeAuthor = computed(() => {
  if (!selectedAuthor.value) return null
  return authors.value.find(a => a.id === selectedAuthor.value) || null
})

const currentDiffFile = computed(() => {
  if (!diff.value?.files?.length) return null
  return (diff.value.files.find(f => f.file === selectedFile.value) || diff.value.files[0]) ?? null
})

const currentDiffText = computed(() => currentDiffFile.value?.diff || '')

async function handleToggleDir(path: string) {
  await treeApi.toggleDir(path)
}

async function loadBlob(path: string) {
  blobLoading.value = true
  try {
    blobContent.value = await getFileBlob(props.repoKey, {
      ref: currentRef.value || undefined,
      path,
    })
  } catch {
    blobContent.value = null
  } finally {
    blobLoading.value = false
  }
}

async function refreshAll() {
  await treeApi.loadTree(currentRef.value)
  if (isWorktree.value) {
    await wsApi.loadStatus()
  }
}

async function selectTreeFile(path: string) {
  selectedFile.value = path
  blobContent.value = null
  diff.value = null
  if (isWorktree.value && wsApi.getFileStatus(path)) {
    await wsApi.loadDiff(path)
    diff.value = wsApi.diff.value
  } else {
    await loadBlob(path)
  }
}

async function selectChangedFile(path: string) {
  selectedFile.value = path
  blobContent.value = null
  diff.value = null
  await wsApi.loadDiff(path)
  diff.value = wsApi.diff.value
}

async function handleGenerateMsg() {
  generatingMsg.value = true
  try {
    const res = await generateCommitMessage(props.repoKey)
    if (res.message) {
      commitMsg.value = res.message.replace(/^["']|["']$/g, '').trim()
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'AI 生成提交信息失败')
  } finally {
    generatingMsg.value = false
  }
}

async function doCommit() {
  if (!commitMsg.value) {
    ElMessage.warning('请输入提交信息')
    return
  }
  committing.value = true
  try {
    await commitChanges({
      repo_key: props.repoKey,
      stage_all: true,
      message: commitMsg.value,
      author_name: activeAuthor.value?.canonical_name || undefined,
      author_email: activeAuthor.value?.canonical_email || undefined,
      push: pushAfterCommit.value,
    })
    ElMessage.success(pushAfterCommit.value ? '提交并推送成功' : '提交成功')
    commitMsg.value = ''
    selectedFile.value = ''
    blobContent.value = null
    diff.value = null
    await refreshAll()
  } catch (e: any) {
    ElMessage.error(e?.message || '提交失败')
  } finally {
    committing.value = false
  }
}

function openConflictResolver(path: string) {
  conflictFile.value = path
  showConflictResolver.value = true
}

function onConflictResolved() {
  showConflictResolver.value = false
  conflictFile.value = ''
  refreshAll()
}

async function loadBranches() {
  try {
    const res = await getBranchList(props.repoKey, { page_size: 500 })
    branches.value = (res.list || []).map(b => b.name)
    const t = await getTagList(props.repoKey)
    tags.value = t || []
  } catch { /* ignore */ }
}

async function loadAuthors() {
  try {
    authors.value = await listIdentities()
    const config = await getRepoAuthorConfig(props.repoKey)
    if (config.identity_id) {
      selectedAuthor.value = config.identity_id
    }
  } catch { /* ignore */ }
}

watch(currentRef, () => {
  selectedFile.value = ''
  blobContent.value = null
  diff.value = null
  treeApi.resetExpanded()
  if (!isWorktree.value) {
    viewMode.value = 'all'
  }
  refreshAll()
})

onMounted(async () => {
  await loadBranches()
  await refreshAll()
  loadAuthors()
})
</script>

<style scoped>
.file-explorer {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  background: var(--bg-color-page);
}

.fe-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-color-page);
  flex-shrink: 0;
}

.fe-ref-select {
  width: 180px;
}

.fe-home-btn {
  cursor: pointer;
  color: var(--text-color-secondary);
  font-size: 16px;
  transition: color var(--transition-fast);
}

.fe-home-btn:hover {
  color: var(--primary-color);
}

.fe-spacer {
  flex: 1;
}

.fe-ahead {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: var(--success-color);
  font-size: 12px;
  font-weight: 600;
}

.fe-behind {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: var(--text-color-secondary);
  font-size: 12px;
  font-weight: 600;
}

.fe-content {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
</style>
