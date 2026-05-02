import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getWorkspaceStatus,
  getWorkspaceDiff,
  stageFiles,
  unstageFiles,
  removeTracking,
  addToGitignore,
  pullWithResolve,
  pushCurrent,
  type WorkspaceStatus,
  type WorkspaceDiff,
} from '@/api/modules/workspace'
import { showGitError } from '@/utils/git'

export function useWorkspaceStatus(repoKey: string) {
  const wsStatus = ref<WorkspaceStatus | null>(null)
  const diff = ref<WorkspaceDiff | null>(null)
  const pulling = ref(false)
  const pushing = ref(false)

  const fileStatusMap = computed(() => {
    const map: Record<string, string> = {}
    if (!wsStatus.value) return map
    for (const f of wsStatus.value.untracked) map[f.path] = '?'
    for (const f of wsStatus.value.unstaged) map[f.path] = f.status
    for (const f of wsStatus.value.staged) map[f.path] = f.status
    for (const f of wsStatus.value.conflicted) map[f.path] = 'C'
    return map
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

  async function loadStatus() {
    try {
      wsStatus.value = await getWorkspaceStatus(repoKey)
    } catch {
      wsStatus.value = null
    }
  }

  async function loadDiff(path: string): Promise<boolean> {
    try {
      diff.value = await getWorkspaceDiff(repoKey, { file: path })
      return true
    } catch {
      diff.value = null
      return false
    }
  }

  async function stageFile(file: string, onRefresh: () => Promise<void>) {
    try {
      await stageFiles(repoKey, [file])
      await onRefresh()
    } catch (e: any) {
      ElMessage.error(e?.message || '暂存失败')
    }
  }

  async function unstageFile(file: string, onRefresh: () => Promise<void>) {
    try {
      await unstageFiles(repoKey, [file])
      await onRefresh()
    } catch (e: any) {
      ElMessage.error(e?.message || '取消暂存失败')
    }
  }

  async function stageAllUnstaged(onRefresh: () => Promise<void>) {
    try {
      await stageFiles(repoKey, [], true)
      await onRefresh()
      ElMessage.success('全部已暂存')
    } catch (e: any) {
      ElMessage.error(e?.message || '暂存失败')
    }
  }

  async function unstageAllFiles(onRefresh: () => Promise<void>) {
    try {
      await unstageFiles(repoKey, [], true)
      await onRefresh()
      ElMessage.success('已取消全部暂存')
    } catch (e: any) {
      ElMessage.error(e?.message || '取消暂存失败')
    }
  }

  async function handleGitignore(paths: string[], onRefresh: () => Promise<void>) {
    try {
      await ElMessageBox.confirm(
        `将 ${paths.length} 个文件/目录加入 .gitignore？`,
        '确认忽略',
        { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' },
      )
      await addToGitignore(repoKey, paths)
      await onRefresh()
      ElMessage.success('已加入 .gitignore')
    } catch { /* cancelled or error */ }
  }

  async function handleUntrack(paths: string[], onRefresh: () => Promise<void>) {
    try {
      await removeTracking(repoKey, paths)
      await onRefresh()
      ElMessage.success('已取消跟踪')
    } catch (e: any) {
      ElMessage.error(e?.message || '取消跟踪失败')
    }
  }

  async function handlePull(onRefresh: () => Promise<void>) {
    pulling.value = true
    try {
      const result = await pullWithResolve(repoKey)
      if (result.status === 'conflicts') {
        ElMessage.warning(`拉取完成，检测到 ${result.conflicts.length} 个冲突`)
      } else {
        ElMessage.success('拉取成功')
      }
      await onRefresh()
    } catch (e: any) {
      showGitError(e, '拉取')
    } finally {
      pulling.value = false
    }
  }

  async function handlePush(onRefresh: () => Promise<void>) {
    pushing.value = true
    try {
      await pushCurrent(repoKey)
      ElMessage.success('推送成功')
      await onRefresh()
    } catch (e: any) {
      showGitError(e, '推送')
    } finally {
      pushing.value = false
    }
  }

  async function handleFileAction(action: string, path: string, onRefresh: () => Promise<void>) {
    switch (action) {
      case 'stage':
        await stageFile(path, onRefresh)
        break
      case 'unstage':
        await unstageFile(path, onRefresh)
        break
      case 'untrack':
        await handleUntrack([path], onRefresh)
        break
      case 'gitignore':
        await handleGitignore([path], onRefresh)
        break
    }
  }

  return {
    wsStatus,
    diff,
    pulling,
    pushing,
    fileStatusMap,
    hasChanges,
    getFileStatus,
    isFileStaged,
    isFileUnstaged,
    isFileUntracked,
    isFileConflicted,
    loadStatus,
    loadDiff,
    stageFile,
    unstageFile,
    stageAllUnstaged,
    unstageAllFiles,
    handleGitignore,
    handleUntrack,
    handlePull,
    handlePush,
    handleFileAction,
  }
}
