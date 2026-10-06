import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { scanRepo, updateRepo } from '@/api/modules/repo'
import { testConnection } from '@/api/modules/system'
import { testCredential } from '@/api/modules/credential'
import type { GitRemote, TrackingBranch } from '@/types/repo'
import { validateGitRemoteUrl, detectGitProtocol, convertGitUrl } from '@/utils/git'
import { toastApiError } from '@/composables/useNotification'

export interface EditRemoteRow extends GitRemote {
  _testing?: boolean
}

interface RepoEditSource {
  name: string
  path: string
  remote_url?: string
  default_credential_id?: number
  remote_credentials?: Record<string, number>
}

/**
 * 仓库编辑表单的共享逻辑：EditRepoPage 与 RepoEditDialog 原先各持一份
 * 80-120 行的逐字重复实现（URL 协议切换/校验、远程列表、凭证映射、
 * 测试连接、保存载荷），收敛于此。
 */
export function useRepoEditForm(repoKey: () => string) {
  const editSaving = ref(false)
  const editForm = ref({ name: '', path: '', remote_url: '' })
  const editUrlError = ref('')
  const editRemotes = ref<EditRemoteRow[]>([])
  const editTrackingBranches = ref<TrackingBranch[]>([])
  const editDefaultCredentialId = ref<number | undefined>()
  const editRemoteCredentials = ref<Record<string, number | undefined>>({})
  const editUrlMode = ref<'ssh' | 'https'>('ssh')
  const remoteUrlModes = ref<Record<number, 'ssh' | 'https'>>({})
  const remoteUrlErrors = ref<Record<number, string>>({})

  const protoToMode = (url: string): 'ssh' | 'https' =>
    detectGitProtocol(url) === 'http' ? 'https' : 'ssh'

  // 协议切换统一走 watch：无论来自模板 v-model、switch 函数还是 URL 校验里的
  // 自动识别，都把已有 URL 转换到新模式（convertGitUrl 幂等，格式相同则原样返回）。
  watch(editUrlMode, (newMode, oldMode) => {
    if (oldMode && newMode !== oldMode && editForm.value.remote_url) {
      editForm.value.remote_url = convertGitUrl(editForm.value.remote_url, newMode)
    }
  })
  watch(remoteUrlModes, (newModes, oldModes) => {
    if (!oldModes) return
    for (const [idxStr, newMode] of Object.entries(newModes)) {
      const idx = parseInt(idxStr)
      const oldMode = oldModes[idx]
      if (oldMode && newMode !== oldMode && editRemotes.value[idx]?.fetch_url) {
        editRemotes.value[idx]!.fetch_url = convertGitUrl(editRemotes.value[idx]!.fetch_url, newMode)
      }
    }
  }, { deep: true })

  function switchMainProto(mode: 'ssh' | 'https') {
    editUrlMode.value = mode
  }

  function switchRemoteProto(index: number, mode: 'ssh' | 'https') {
    remoteUrlModes.value[index] = mode
  }

  function validateMainUrl() {
    const url = editForm.value.remote_url
    if (!url) { editUrlError.value = ''; return }
    editUrlMode.value = protoToMode(url)
    editUrlError.value = validateGitRemoteUrl(url)
  }

  function validateRemoteUrl(index: number) {
    const remote = editRemotes.value[index]
    if (!remote) return
    if (!remote.fetch_url) { delete remoteUrlErrors.value[index]; return }
    remoteUrlModes.value[index] = protoToMode(remote.fetch_url)
    const err = validateGitRemoteUrl(remote.fetch_url)
    if (err) remoteUrlErrors.value[index] = err
    else delete remoteUrlErrors.value[index]
  }

  function addEditRemote() {
    editRemotes.value.push({ name: '', fetch_url: '', push_url: '', is_mirror: false, _testing: false })
  }

  function removeEditRemote(index: number) {
    editRemotes.value.splice(index, 1)
    delete remoteUrlErrors.value[index]
    delete remoteUrlModes.value[index]
  }

  function updateEditRemoteCred(name: string, val: number | undefined) {
    if (val) editRemoteCredentials.value[name] = val
    else delete editRemoteCredentials.value[name]
  }

  async function testEditRemote(index: number) {
    const row = editRemotes.value[index]
    if (!row || !row.fetch_url) { ElMessage.warning('请输入 Fetch URL'); return }
    row._testing = true
    try {
      const credential_id = editRemoteCredentials.value[row.name] || editDefaultCredentialId.value
      if (credential_id) {
        const result = await testCredential(credential_id, row.fetch_url)
        if (result.success) ElMessage.success(`${row.name || 'Remote'} 连接成功`)
        else toastApiError(result, '连接失败', '未知错误')
      } else {
        const result = await testConnection(row.fetch_url)
        if (result.status === 'success') ElMessage.success(`${row.name || 'Remote'} 连接成功`)
        else ElMessage.error('连接失败: ' + (result.error || '未知错误'))
      }
    } catch (e) {
      toastApiError(e, '连接测试请求失败')
    } finally {
      row._testing = false
    }
  }

  /** 用仓库数据填充表单，并异步扫描本地远程/追踪分支（扫描失败静默忽略）。 */
  function fillFromRepo(repo: RepoEditSource) {
    editForm.value = { name: repo.name, path: repo.path, remote_url: repo.remote_url || '' }
    editRemotes.value = []
    editTrackingBranches.value = []
    editDefaultCredentialId.value = repo.default_credential_id
    editRemoteCredentials.value = { ...(repo.remote_credentials || {}) }
    editUrlError.value = ''
    remoteUrlErrors.value = {}
    remoteUrlModes.value = {}
    editUrlMode.value = protoToMode(repo.remote_url || '')

    if (!repo.path) return
    scanRepo(repo.path).then(result => {
      editRemotes.value = (result.remotes || []).map(r => ({ ...r, _testing: false }))
      editTrackingBranches.value = result.branches || []
      editRemotes.value.forEach((r, i) => {
        remoteUrlModes.value[i] = protoToMode(r.fetch_url || '')
      })
      if (!editForm.value.remote_url && editRemotes.value.length > 0) {
        editForm.value.remote_url = editRemotes.value[0]!.fetch_url
        editUrlMode.value = protoToMode(editForm.value.remote_url)
      }
    }).catch(() => { /* 扫描失败不阻断编辑 */ })
  }

  /** 保存前的统一校验；失败时已写好对应错误状态并提示。 */
  function validateForSave(): boolean {
    if (!editForm.value.name || !editForm.value.path) {
      ElMessage.warning('名称和路径不能为空')
      return false
    }
    if (editForm.value.remote_url) {
      const err = validateGitRemoteUrl(editForm.value.remote_url)
      if (err) { editUrlError.value = err; return false }
    }
    for (let i = 0; i < editRemotes.value.length; i++) {
      const r = editRemotes.value[i]!
      if (r.fetch_url) {
        const err = validateGitRemoteUrl(r.fetch_url)
        if (err) {
          remoteUrlErrors.value[i] = err
          ElMessage.warning(`远程 "${r.name || 'unnamed'}" 的 URL 格式不正确`)
          return false
        }
      }
    }
    return true
  }

  /** 构造 updateRepo 的请求载荷。 */
  function buildSavePayload() {
    const remotes: GitRemote[] = editRemotes.value
      .filter(r => r.name && r.fetch_url)
      .map(r => ({ name: r.name, fetch_url: r.fetch_url, push_url: r.push_url || r.fetch_url, is_mirror: r.is_mirror }))
    const remote_credentials: Record<string, number> = {}
    for (const [k, v] of Object.entries(editRemoteCredentials.value)) {
      if (v) remote_credentials[k] = v
    }
    return {
      key: repoKey(),
      name: editForm.value.name,
      path: editForm.value.path,
      remote_url: editForm.value.remote_url || undefined,
      remotes,
      default_credential_id: editDefaultCredentialId.value,
      remote_credentials: Object.keys(remote_credentials).length > 0 ? remote_credentials : undefined,
    }
  }

  async function save() {
    if (!validateForSave()) return
    editSaving.value = true
    try {
      await updateRepo(buildSavePayload())
      ElMessage.success('保存成功')
      return true
    } finally {
      editSaving.value = false
    }
  }

  return {
    editSaving, editForm, editUrlError, editRemotes, editTrackingBranches,
    editDefaultCredentialId, editRemoteCredentials, editUrlMode,
    remoteUrlModes, remoteUrlErrors,
    switchMainProto, switchRemoteProto,
    validateMainUrl, validateRemoteUrl,
    addEditRemote, removeEditRemote, updateEditRemoteCred, testEditRemote,
    fillFromRepo, validateForSave, buildSavePayload, save,
  }
}
