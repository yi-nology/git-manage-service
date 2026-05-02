import { ref, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listIdentities,
  createIdentity,
  updateIdentity,
  deleteIdentity,
  activateIdentity,
  getRepoAuthorConfig,
  setRepoAuthorConfig,
  scanAuthor,
  fixAuthorAll,
  fixAuthor,
  authorAI,
  authorChat,
} from '@/api/modules/author'
import type { AuthorIdentityDTO, AliasEntry, RepoAuthorConfigDTO, MismatchedCommit, AliasSuggestionResult, MergeSuggestionResult, RiskAssessmentResult, ChatMessageDTO } from '@/api/modules/author'
import { getTaskStatus } from '@/api/modules/maintenance'

export function useAuthorIdentity() {
  const identities = ref<AuthorIdentityDTO[]>([])
  const loading = ref(false)

  async function loadIdentities() {
    loading.value = true
    try {
      identities.value = (await listIdentities()) as AuthorIdentityDTO[] || []
    } catch { identities.value = [] }
    finally { loading.value = false }
  }

  async function handleCreate(data: { canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }) {
    try {
      await createIdentity(data)
      ElMessage.success('身份创建成功')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('创建失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleUpdate(id: number, data: { canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }) {
    try {
      await updateIdentity(id, data)
      ElMessage.success('身份更新成功')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('更新失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleDelete(id: number) {
    try {
      await ElMessageBox.confirm('确认删除此身份？关联的仓库将恢复使用全局默认。', '删除身份', { type: 'warning' })
    } catch { return }
    try {
      await deleteIdentity(id)
      ElMessage.success('删除成功')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('删除失败: ' + (e.message || '未知错误'))
    }
  }

  async function handleActivate(id: number) {
    try {
      await activateIdentity(id)
      ElMessage.success('已激活并更新 ~/.gitconfig')
      await loadIdentities()
    } catch (e: any) {
      ElMessage.error('激活失败: ' + (e.message || '未知错误'))
    }
  }

  return {
    identities,
    loading,
    loadIdentities,
    handleCreate,
    handleUpdate,
    handleDelete,
    handleActivate,
  }
}

export function useAuthorFix(repoKey: string) {
  const repoConfig = ref<RepoAuthorConfigDTO | null>(null)
  const configLoading = ref(false)

  const scanResult = ref<MismatchedCommit[]>([])
  const scanLoading = ref(false)
  const totalCommits = ref(0)
  const selectedCommits = ref<MismatchedCommit[]>([])

  const taskId = ref('')
  const taskStatus = ref('')
  const taskLogs = ref<string[]>([])
  const taskError = ref('')

  let pollTimer: ReturnType<typeof setInterval> | null = null

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  async function loadRepoConfig() {
    configLoading.value = true
    try {
      repoConfig.value = (await getRepoAuthorConfig(repoKey)) as RepoAuthorConfigDTO || null
    } catch { repoConfig.value = null }
    finally { configLoading.value = false }
  }

  async function setConfig(identityId: number | null) {
    try {
      if (identityId === null) {
        await setRepoAuthorConfig(repoKey, null, true)
      } else {
        await setRepoAuthorConfig(repoKey, identityId)
      }
      ElMessage.success('仓库作者身份已更新')
      await loadRepoConfig()
    } catch (e: any) {
      ElMessage.error('设置失败: ' + (e.message || '未知错误'))
    }
  }

  async function scan() {
    scanLoading.value = true
    try {
      const result = (await scanAuthor(repoKey)) as any
      scanResult.value = result?.commits || []
      totalCommits.value = result?.totalCommits || 0
    } catch (e: any) {
      ElMessage.error('扫描失败: ' + (e.message || '未知错误'))
    } finally {
      scanLoading.value = false
    }
  }

  async function fixAll(pushRemote = '') {
    try {
      const res = (await fixAuthorAll(repoKey, pushRemote)) as any
      taskId.value = res?.taskId || ''
      taskStatus.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('修复失败: ' + (e.message || '未知错误'))
    }
  }

  async function fixSelected(pushRemote = '') {
    if (selectedCommits.value.length === 0) {
      ElMessage.warning('请先选择要修复的提交')
      return
    }
    try {
      const hashes = selectedCommits.value.map(c => c.hash)
      const res = (await fixAuthor(repoKey, hashes, pushRemote)) as any
      taskId.value = res?.taskId || ''
      taskStatus.value = 'running'
      taskLogs.value = []
      taskError.value = ''
      startPolling()
    } catch (e: any) {
      ElMessage.error('修复失败: ' + (e.message || '未知错误'))
    }
  }

  function startPolling() {
    stopPolling()
    pollTimer = setInterval(async () => {
      try {
        const task = (await getTaskStatus(taskId.value)) as any
        taskStatus.value = task?.status || ''
        taskLogs.value = task?.progress || []
        taskError.value = task?.error || ''
        if (task?.status === 'success' || task?.status === 'failed') {
          stopPolling()
          if (task.status === 'success') {
            ElMessage.success('作者修复完成')
            scan()
          }
        }
      } catch {
        stopPolling()
      }
    }, 2000)
  }

  function handleSelection(rows: MismatchedCommit[]) {
    selectedCommits.value = rows
  }

  onUnmounted(() => {
    stopPolling()
  })

  return {
    repoConfig, configLoading, scanResult, scanLoading, totalCommits, selectedCommits,
    taskId, taskStatus, taskLogs, taskError,
    loadRepoConfig, setConfig, scan, fixAll, fixSelected, handleSelection,
  }
}

export function useAuthorAI(repoKey: string) {
  const aiLoading = ref(false)
  const aiAnalysis = ref('')
  const aiSuggestion = ref<AliasSuggestionResult | null>(null)
  const aiMerge = ref<MergeSuggestionResult | null>(null)
  const aiRisk = ref<RiskAssessmentResult | null>(null)

  const chatMessages = ref<ChatMessageDTO[]>([])
  const chatLoading = ref(false)

  async function smartSuggest() {
    aiLoading.value = true
    aiSuggestion.value = null
    try {
      const res = (await authorAI(repoKey, 'suggest')) as any
      aiSuggestion.value = res?.suggest || null
    } catch (e: any) {
      ElMessage.error('AI 推荐失败: ' + (e.message || '请检查 LLM 配置'))
    } finally {
      aiLoading.value = false
    }
  }

  async function analyzeScan(scanData: { commits: MismatchedCommit[]; totalCommits: number; matchCount: number }) {
    aiLoading.value = true
    aiAnalysis.value = ''
    try {
      const res = (await authorAI(repoKey, 'analyze', { scan: scanData })) as any
      aiAnalysis.value = res?.result || ''
    } catch (e: any) {
      ElMessage.error('AI 分析失败: ' + (e.message || '请检查 LLM 配置'))
    } finally {
      aiLoading.value = false
    }
  }

  async function suggestMerges() {
    aiLoading.value = true
    aiMerge.value = null
    try {
      const res = (await authorAI(repoKey, 'merge')) as any
      aiMerge.value = res?.merge || null
    } catch (e: any) {
      ElMessage.error('AI 分析失败: ' + (e.message || '请检查 LLM 配置'))
    } finally {
      aiLoading.value = false
    }
  }

  async function assessRisk(commits: MismatchedCommit[]) {
    aiLoading.value = true
    aiRisk.value = null
    try {
      const res = (await authorAI(repoKey, 'risk', { commits })) as any
      aiRisk.value = res?.risk || null
    } catch (e: any) {
      ElMessage.error('AI 风险评估失败: ' + (e.message || '请检查 LLM 配置'))
    } finally {
      aiLoading.value = false
    }
  }

  async function sendChat(prompt: string) {
    chatLoading.value = true
    chatMessages.value.push({ role: 'user', content: prompt })
    try {
      const res = (await authorChat(repoKey, prompt, chatMessages.value.slice(-10))) as any
      const answer = res?.result || '无回复'
      chatMessages.value.push({ role: 'assistant', content: answer })
    } catch (e: any) {
      chatMessages.value.push({ role: 'assistant', content: '抱歉，AI 调用失败: ' + (e.message || '请检查 LLM 配置') })
    } finally {
      chatLoading.value = false
    }
  }

  function clearChat() {
    chatMessages.value = []
  }

  return {
    aiLoading, aiAnalysis, aiSuggestion, aiMerge, aiRisk,
    chatMessages, chatLoading,
    smartSuggest, analyzeScan, suggestMerges, assessRisk, sendChat, clearChat,
  }
}
