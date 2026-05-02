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

      <el-icon class="fe-home-btn" @click="collapseAll"><HomeFilled /></el-icon>

      <template v-if="isWorktree && wsStatus">
        <el-tag size="small" effect="plain">{{ wsStatus.branch }}</el-tag>
        <span v-if="wsStatus.ahead" class="fe-ahead">
          <el-icon><Top /></el-icon>{{ wsStatus.ahead }}
        </span>
        <span v-if="wsStatus.behind" class="fe-behind">
          <el-icon><Bottom /></el-icon>{{ wsStatus.behind }}
        </span>
        <el-tag v-if="!wsStatus.isClean" size="small" type="warning">有变更</el-tag>
        <el-tag v-else size="small" type="success">干净</el-tag>
      </template>

      <div class="fe-spacer" />

      <el-button v-if="isWorktree" size="small" :loading="pulling" @click="handlePull">
        <el-icon><Download /></el-icon>
      </el-button>
      <el-button v-if="isWorktree" size="small" :loading="pushing" @click="handlePush">
        <el-icon><Upload /></el-icon>
      </el-button>
      <el-button size="small" :loading="treeLoading" @click="refreshAll">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>

    <div class="fe-content">
      <div class="fe-sidebar">
        <div v-if="isWorktree" class="fe-view-toggle">
          <span :class="{ active: viewMode === 'all' }" @click="viewMode = 'all'">全部</span>
          <span :class="{ active: viewMode === 'changes' }" @click="viewMode = 'changes'">仅变更</span>
        </div>

        <div v-show="viewMode === 'all'" class="fe-tree" v-loading="treeLoading">
          <div
            v-for="item in flatTreeItems"
            :key="item.path"
            class="fe-tree-item"
            :class="{
              active: selectedFile === item.path && !item.isDir,
              'is-dir': item.isDir,
            }"
            :style="{ paddingLeft: 8 + item.depth * 16 + 'px' }"
            @click="item.isDir ? toggleDir(item.path) : selectTreeFile(item.path)"
          >
            <span v-if="item.isDir" class="fe-tree-toggle">
              <el-icon :size="12">
                <ArrowDown v-if="expandedDirs.has(item.path)" />
                <ArrowRight v-else />
              </el-icon>
            </span>
            <el-icon v-if="item.isDir" :size="14"><Folder /></el-icon>
            <el-icon v-else :size="14"><Document /></el-icon>
            <span class="fe-tree-name" :title="item.name">{{ item.name }}</span>

            <template v-if="isWorktree && getFileStatus(item.path)">
              <el-tag
                size="small"
                :type="statusTagType(getFileStatus(item.path)!)"
                class="fe-tree-badge"
              >
                {{ getFileStatus(item.path)![0].toUpperCase() }}
              </el-tag>
              <div class="fe-tree-actions">
                <template v-if="isFileUntracked(item.path)">
                  <el-button size="small" type="success" link @click.stop="stageFile(item.path)">
                    <el-icon><Plus /></el-icon>
                  </el-button>
                  <el-button size="small" type="info" link @click.stop="handleGitignore([item.path])">
                    <el-icon><Close /></el-icon>
                  </el-button>
                </template>
                <template v-else-if="isFileConflicted(item.path)">
                  <el-button size="small" type="danger" link @click.stop="openConflictResolver(item.path)">
                    <el-icon><WarningFilled /></el-icon>
                  </el-button>
                </template>
                <template v-else>
                  <el-button
                    v-if="isFileUnstaged(item.path)"
                    size="small" type="success" link
                    @click.stop="stageFile(item.path)"
                  >
                    <el-icon><Plus /></el-icon>
                  </el-button>
                  <el-button
                    v-if="isFileStaged(item.path)"
                    size="small" type="warning" link
                    @click.stop="unstageFile(item.path)"
                  >
                    <el-icon><Minus /></el-icon>
                  </el-button>
                </template>
              </div>
            </template>
          </div>
          <el-empty v-if="!treeLoading && !flatTreeItems.length" description="空目录" :image-size="48" />
        </div>

        <div v-if="viewMode === 'changes' && isWorktree" class="fe-changes">
          <template v-if="wsStatus?.staged?.length">
            <div class="fe-section-header">
              <span class="fe-dot fe-dot-green" />
              已暂存 ({{ wsStatus.staged.length }})
              <div class="fe-section-spacer" />
              <el-button size="small" type="warning" link @click.stop="unstageAllFiles">全部取消</el-button>
            </div>
            <div
              v-for="f in wsStatus.staged"
              :key="'s-' + f.path"
              class="fe-change-item"
              :class="{ active: selectedFile === f.path }"
              @click="selectChangedFile(f.path)"
            >
              <el-icon :size="14"><Document /></el-icon>
              <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
              <el-tag size="small" :type="statusTagType(f.status)">{{ f.status[0].toUpperCase() }}</el-tag>
              <el-button size="small" type="warning" @click.stop="unstageFile(f.path)">
                <el-icon><Minus /></el-icon>
              </el-button>
            </div>
          </template>

          <template v-if="wsStatus?.unstaged?.length">
            <div class="fe-section-header">
              <span class="fe-dot fe-dot-orange" />
              未暂存 ({{ wsStatus.unstaged.length }})
              <div class="fe-section-spacer" />
              <el-button size="small" type="success" link @click.stop="stageAllUnstaged">全部暂存</el-button>
            </div>
            <div
              v-for="f in wsStatus.unstaged"
              :key="'u-' + f.path"
              class="fe-change-item"
              :class="{ active: selectedFile === f.path }"
              @click="selectChangedFile(f.path)"
            >
              <el-icon :size="14"><Document /></el-icon>
              <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
              <el-tag size="small" :type="statusTagType(f.status)">{{ f.status[0].toUpperCase() }}</el-tag>
              <el-button size="small" type="success" @click.stop="stageFile(f.path)">
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>
          </template>

          <template v-if="wsStatus?.untracked?.length">
            <div class="fe-section-header">
              <span class="fe-dot fe-dot-gray" />
              未跟踪 ({{ wsStatus.untracked.length }})
            </div>
            <div
              v-for="f in wsStatus.untracked"
              :key="'t-' + f.path"
              class="fe-change-item"
              :class="{ active: selectedFile === f.path }"
              @click="selectChangedFile(f.path)"
            >
              <el-icon :size="14"><Document /></el-icon>
              <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
              <el-tag size="small" type="info">?</el-tag>
              <el-button size="small" type="success" @click.stop="stageFile(f.path)">
                <el-icon><Plus /></el-icon>
              </el-button>
              <el-button size="small" type="info" @click.stop="handleGitignore([f.path])">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
          </template>

          <template v-if="wsStatus?.conflicted?.length">
            <div class="fe-section-header">
              <span class="fe-dot fe-dot-red" />
              冲突 ({{ wsStatus.conflicted.length }})
            </div>
            <div
              v-for="f in wsStatus.conflicted"
              :key="'c-' + f.path"
              class="fe-change-item fe-conflict-item"
              @click="openConflictResolver(f.path)"
            >
              <el-icon :size="14" color="var(--danger-color)"><WarningFilled /></el-icon>
              <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
              <el-tag size="small" type="danger">C</el-tag>
              <el-button size="small" type="primary" @click.stop="openConflictResolver(f.path)">
                <el-icon><MagicStick /></el-icon>
              </el-button>
            </div>
          </template>

          <el-empty v-if="wsStatus?.isClean" description="工作区干净" :image-size="60" />
        </div>
      </div>

      <div class="fe-main">
        <template v-if="selectedFile && (currentDiffText || currentDiffFile?.isBinary)">
          <div class="fe-diff-header">
            <span class="fe-diff-filename">{{ selectedFile }}</span>
            <div v-if="currentDiffFile && !currentDiffFile.isBinary" class="fe-diff-stats">
              <span class="fe-additions">+{{ currentDiffFile.additions }}</span>
              <span class="fe-deletions">-{{ currentDiffFile.deletions }}</span>
            </div>
          </div>
          <div v-if="currentDiffFile?.isBinary" class="fe-binary-notice">
            <el-icon><Warning /></el-icon> 二进制文件，无法显示差异
          </div>
          <div v-else class="fe-diff-body">
            <table class="fe-diff-table">
              <tbody>
                <template v-for="(line, idx) in parsedDiffLines" :key="idx">
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

        <template v-else-if="selectedFile && blobContent">
          <div class="fe-diff-header">
            <span class="fe-diff-filename">{{ selectedFile }}</span>
            <span class="fe-blob-size">{{ formatSize(blobContent.size) }}</span>
          </div>
          <div class="fe-blob-body">
            <div v-if="blobContent.is_binary" class="fe-binary-notice">
              <el-icon><Warning /></el-icon> 二进制文件，无法预览
            </div>
            <pre v-else class="fe-blob-pre">{{ blobContent.content }}</pre>
          </div>
        </template>

        <div v-else-if="blobLoading" class="fe-empty">
          <el-icon :size="24" class="is-loading"><Refresh /></el-icon>
        </div>

        <div v-else class="fe-empty">
          <el-icon :size="48" color="var(--text-color-secondary)"><Document /></el-icon>
          <span>选择文件查看内容</span>
        </div>
      </div>
    </div>

    <div v-if="isWorktree && hasChanges" class="fe-commit-bar">
      <el-input
        v-model="commitMsg"
        placeholder="提交信息 (Ctrl+Enter)..."
        size="small"
        class="fe-commit-input"
        @keydown.ctrl.enter="doCommit"
      />
      <el-select
        v-model="selectedAuthor"
        size="small"
        placeholder="选择作者"
        class="fe-author-select"
        clearable
      >
        <el-option
          v-for="a in authors"
          :key="a.id"
          :label="a.canonicalName + ' <' + a.canonicalEmail + '>'"
          :value="a.id"
        />
      </el-select>
      <div class="fe-push-toggle">
        <span>推送</span>
        <el-switch v-model="pushAfterCommit" size="small" />
      </div>
      <el-button
        type="primary"
        size="small"
        :loading="committing"
        :disabled="!commitMsg"
        @click="doCommit"
      >
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  HomeFilled, Folder, Document, Clock, Warning, Refresh, Download, Upload,
  Top, Bottom, Check, WarningFilled, MagicStick, Plus, Minus, Close,
  ArrowRight, ArrowDown,
} from '@element-plus/icons-vue'
import {
  getFileTree, getFileBlob,
  type TreeEntry, type BlobContent,
} from '@/api/modules/file'
import {
  getWorkspaceStatus, getWorkspaceDiff, stageFiles, unstageFiles,
  commitChanges, pullWithResolve, pushCurrent, removeTracking, addToGitignore,
  type WorkspaceStatus, type WorkspaceDiff,
} from '@/api/modules/workspace'
import {
  listIdentities, getRepoAuthorConfig,
  type AuthorIdentityDTO,
} from '@/api/modules/author'
import { getBranchList, getTagList } from '@/api/modules/branch'
import { showGitError } from '@/utils/git'
import ConflictResolver from './ConflictResolver.vue'

void Clock
void ArrowRight

const props = defineProps<{ repoKey: string }>()

const currentRef = ref('worktree')
const selectedFile = ref('')
const viewMode = ref<'all' | 'changes'>('all')

const entries = ref<TreeEntry[]>([])
const treeLoading = ref(false)
const blobContent = ref<BlobContent | null>(null)
const blobLoading = ref(false)

const wsStatus = ref<WorkspaceStatus | null>(null)
const diff = ref<WorkspaceDiff | null>(null)
const pulling = ref(false)
const pushing = ref(false)
const committing = ref(false)

const branches = ref<string[]>([])
const tags = ref<string[]>([])
const showConflictResolver = ref(false)
const conflictFile = ref('')

const expandedDirs = ref(new Set<string>())
const subTreeCache = ref<Record<string, TreeEntry[]>>({})

const commitMsg = ref('')
const selectedAuthor = ref<number | null>(null)
const pushAfterCommit = ref(false)
const authors = ref<AuthorIdentityDTO[]>([])

interface FlatTreeItem {
  name: string
  path: string
  isDir: boolean
  depth: number
}

interface DiffLine {
  type: 'hunk' | 'add' | 'del' | 'ctx'
  content: string
  oldNum: string | number
  newNum: string | number
}

const isWorktree = computed(() => currentRef.value === 'worktree')

const activeAuthor = computed(() => {
  if (!selectedAuthor.value) return null
  return authors.value.find(a => a.id === selectedAuthor.value) || null
})

const fileStatusMap = computed(() => {
  const map: Record<string, string> = {}
  if (!wsStatus.value) return map
  for (const f of wsStatus.value.untracked) map[f.path] = '?'
  for (const f of wsStatus.value.unstaged) map[f.path] = f.status
  for (const f of wsStatus.value.staged) map[f.path] = f.status
  for (const f of wsStatus.value.conflicted) map[f.path] = 'C'
  return map
})

const flatTreeItems = computed<FlatTreeItem[]>(() => {
  const items: FlatTreeItem[] = []
  function walk(list: TreeEntry[], depth: number) {
    const sorted = [...list].sort((a, b) => {
      if (a.type !== b.type) return a.type === 'dir' ? -1 : 1
      return a.name.localeCompare(b.name)
    })
    for (const e of sorted) {
      items.push({ name: e.name, path: e.path, isDir: e.type === 'dir', depth })
      if (e.type === 'dir' && expandedDirs.value.has(e.path)) {
        const children = subTreeCache.value[e.path]
        if (children) walk(children, depth + 1)
      }
    }
  }
  walk(entries.value, 0)
  return items
})

const currentDiffFile = computed(() => {
  if (!diff.value?.files?.length) return null
  return diff.value.files.find(f => f.file === selectedFile.value) || diff.value.files[0]
})

const currentDiffText = computed(() => currentDiffFile.value?.diff || '')

const parsedDiffLines = computed<DiffLine[]>(() => {
  if (!currentDiffText.value) return []
  return parseDiffLines(currentDiffText.value)
})

const hasChanges = computed(() => {
  if (!wsStatus.value) return false
  return !wsStatus.value.isClean
})

function getFileStatus(path: string): string | undefined {
  return fileStatusMap.value[path]
}

function isFileStaged(path: string): boolean {
  return wsStatus.value?.staged.some(f => f.path === path) || false
}

function isFileUnstaged(path: string): boolean {
  return wsStatus.value?.unstaged.some(f => f.path === path) || false
}

function isFileUntracked(path: string): boolean {
  return wsStatus.value?.untracked.some(f => f.path === path) || false
}

function isFileConflicted(path: string): boolean {
  return wsStatus.value?.conflicted.some(f => f.path === path) || false
}

async function loadBranches() {
  try {
    const res = await getBranchList(props.repoKey, { page_size: 500 })
    branches.value = (res.list || []).map(b => b.name)
    const t = await getTagList(props.repoKey)
    tags.value = t || []
  } catch { /* ignore */ }
}

async function loadTree() {
  treeLoading.value = true
  try {
    const res = await getFileTree(props.repoKey, {
      ref: currentRef.value || undefined,
    })
    entries.value = res.entries || []
  } catch {
    entries.value = []
  } finally {
    treeLoading.value = false
  }
}

async function loadSubTree(path: string) {
  try {
    const res = await getFileTree(props.repoKey, {
      ref: currentRef.value || undefined,
      path,
    })
    subTreeCache.value = { ...subTreeCache.value, [path]: res.entries || [] }
  } catch { /* ignore */ }
}

async function loadStatus() {
  try {
    wsStatus.value = await getWorkspaceStatus(props.repoKey)
  } catch {
    wsStatus.value = null
  }
}

async function loadDiff(path: string) {
  blobLoading.value = true
  try {
    diff.value = await getWorkspaceDiff(props.repoKey, { file: path })
  } catch {
    diff.value = null
  } finally {
    blobLoading.value = false
  }
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
  await loadTree()
  if (isWorktree.value) {
    await loadStatus()
  }
}

async function toggleDir(path: string) {
  const s = new Set(expandedDirs.value)
  if (s.has(path)) {
    s.delete(path)
  } else {
    s.add(path)
    if (!subTreeCache.value[path]) {
      await loadSubTree(path)
    }
  }
  expandedDirs.value = s
}

function collapseAll() {
  expandedDirs.value = new Set()
  subTreeCache.value = {}
}

async function selectTreeFile(path: string) {
  selectedFile.value = path
  blobContent.value = null
  diff.value = null
  if (isWorktree.value && fileStatusMap.value[path]) {
    await loadDiff(path)
  } else {
    await loadBlob(path)
  }
}

async function selectChangedFile(path: string) {
  selectedFile.value = path
  blobContent.value = null
  diff.value = null
  await loadDiff(path)
}

async function stageFile(file: string) {
  try {
    await stageFiles(props.repoKey, [file])
    await refreshAll()
  } catch (e: any) {
    ElMessage.error(e?.message || '暂存失败')
  }
}

async function unstageFile(file: string) {
  try {
    await unstageFiles(props.repoKey, [file])
    await refreshAll()
  } catch (e: any) {
    ElMessage.error(e?.message || '取消暂存失败')
  }
}

async function stageAllUnstaged() {
  try {
    await stageFiles(props.repoKey, [], true)
    await refreshAll()
    ElMessage.success('全部已暂存')
  } catch (e: any) {
    ElMessage.error(e?.message || '暂存失败')
  }
}

async function unstageAllFiles() {
  try {
    await unstageFiles(props.repoKey, [], true)
    await refreshAll()
    ElMessage.success('已取消全部暂存')
  } catch (e: any) {
    ElMessage.error(e?.message || '取消暂存失败')
  }
}

async function handleGitignore(paths: string[]) {
  try {
    await ElMessageBox.confirm(
      `将 ${paths.length} 个文件/目录加入 .gitignore？`,
      '确认忽略',
      { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' },
    )
    await addToGitignore(props.repoKey, paths)
    await refreshAll()
    ElMessage.success('已加入 .gitignore')
  } catch { /* cancelled or error */ }
}

async function handleUntrack(paths: string[]) {
  try {
    await removeTracking(props.repoKey, paths)
    await refreshAll()
    ElMessage.success('已取消跟踪')
  } catch (e: any) {
    ElMessage.error(e?.message || '取消跟踪失败')
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
      author_name: activeAuthor.value?.canonicalName || undefined,
      author_email: activeAuthor.value?.canonicalEmail || undefined,
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

async function handlePull() {
  pulling.value = true
  try {
    const result = await pullWithResolve(props.repoKey)
    if (result.status === 'conflicts') {
      ElMessage.warning(`拉取完成，检测到 ${result.conflicts.length} 个冲突`)
    } else {
      ElMessage.success('拉取成功')
    }
    await refreshAll()
  } catch (e: any) {
    showGitError(e, '拉取')
  } finally {
    pulling.value = false
  }
}

async function handlePush() {
  pushing.value = true
  try {
    await pushCurrent(props.repoKey)
    ElMessage.success('推送成功')
    await refreshAll()
  } catch (e: any) {
    showGitError(e, '推送')
  } finally {
    pushing.value = false
  }
}

function openConflictResolver(path: string) {
  conflictFile.value = path
  showConflictResolverRef.value = true
}

function onConflictResolved() {
  showConflictResolverRef.value = false
  conflictFile.value = ''
  refreshAll()
}

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

function statusTagType(status: string): string {
  switch (status) {
    case 'added': return 'success'
    case 'modified': return 'warning'
    case 'deleted': return 'danger'
    case 'renamed': return 'info'
    case 'copied': return 'info'
    case '?': return 'info'
    case 'C': return 'danger'
    default: return 'info'
  }
}

async function loadAuthors() {
  try {
    authors.value = await listIdentities()
    const config = await getRepoAuthorConfig(props.repoKey)
    if (config.identityId) {
      selectedAuthor.value = config.identityId
    }
  } catch { /* ignore */ }
}

watch(currentRef, () => {
  selectedFile.value = ''
  blobContent.value = null
  diff.value = null
  expandedDirs.value = new Set()
  subTreeCache.value = {}
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

.fe-sidebar {
  width: 280px;
  min-width: 280px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-color);
  overflow: hidden;
}

.fe-view-toggle {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.fe-view-toggle span {
  flex: 1;
  text-align: center;
  padding: 6px 0;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  color: var(--text-color-secondary);
  border-bottom: 2px solid transparent;
  transition: all var(--transition-fast);
}

.fe-view-toggle span:hover {
  color: var(--text-color-primary);
}

.fe-view-toggle span.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

.fe-tree {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.fe-tree-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  cursor: pointer;
  transition: background var(--transition-fast);
  position: relative;
  white-space: nowrap;
}

.fe-tree-item:hover {
  background: var(--surface-hover);
}

.fe-tree-item.active {
  background: var(--accent-bg);
}

.fe-tree-item.is-dir {
  color: var(--text-color-primary);
}

.fe-tree-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.fe-tree-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  color: var(--text-color-primary);
}

.fe-tree-badge {
  flex-shrink: 0;
  margin-left: 4px;
}

.fe-tree-actions {
  display: none;
  align-items: center;
  gap: 2px;
  margin-left: auto;
  flex-shrink: 0;
}

.fe-tree-item:hover .fe-tree-actions {
  display: flex;
}

.fe-changes {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.fe-section-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-color-regular);
  position: sticky;
  top: 0;
  background: var(--bg-color-page);
  z-index: 1;
}

.fe-section-spacer {
  flex: 1;
}

.fe-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.fe-dot-green { background: var(--success-color); }
.fe-dot-orange { background: var(--warning-color); }
.fe-dot-gray { background: var(--text-color-secondary); }
.fe-dot-red { background: var(--danger-color); }

.fe-change-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.fe-change-item:hover {
  background: var(--surface-hover);
}

.fe-change-item.active {
  background: var(--accent-bg);
}

.fe-conflict-item {
  background: rgba(239, 68, 68, 0.05);
  border-left: 2px solid var(--danger-color);
}

.fe-change-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-color-primary);
}

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

.fe-commit-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-color-page);
  flex-shrink: 0;
}

.fe-commit-input {
  flex: 1;
}

.fe-author-select {
  width: 200px;
}

.fe-push-toggle {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
  white-space: nowrap;
}
</style>
