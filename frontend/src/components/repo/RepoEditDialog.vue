<template>
  <el-dialog :model-value="visible" @update:model-value="emit('update:visible', $event)" title="编辑仓库" width="750px" destroy-on-close>
    <el-form :model="editForm" label-width="100px">
      <el-form-item label="名称" required>
        <el-input v-model="editForm.name" placeholder="仓库名称" />
      </el-form-item>
      <el-form-item label="本地路径" required>
        <el-input v-model="editForm.path" placeholder="本地仓库路径" />
      </el-form-item>
      <el-form-item label="远程 URL">
        <div class="url-input-group">
          <el-radio-group v-model="editUrlMode" size="small" class="url-mode-switch">
            <el-radio-button value="ssh">SSH</el-radio-button>
            <el-radio-button value="https">HTTPS</el-radio-button>
          </el-radio-group>
          <el-input
            v-model="editForm.remote_url"
            :placeholder="editUrlMode === 'ssh' ? 'git@github.com:user/repo.git' : 'https://github.com/user/repo.git'"
            @blur="validateEditUrl"
            :class="{ 'is-error-input': editUrlError }"
          />
        </div>
        <div v-if="editUrlError" class="field-error">{{ editUrlError }}</div>
      </el-form-item>
      <el-form-item label="默认凭证">
        <CredentialSelector
          v-model="editDefaultCredentialId"
          :url="editForm.remote_url"
          placeholder="选择默认凭证（可选）"
        />
      </el-form-item>

      <el-divider content-position="left">远程仓库配置</el-divider>

      <el-form-item label="">
        <div class="remotes-section">
          <div class="remotes-header">
            <span>配置多个远程仓库及其凭证</span>
            <el-button size="small" type="primary" @click="addEditRemote">+ 新增远程</el-button>
          </div>
          <div v-for="(remote, index) in editRemotes" :key="index" class="edit-remote-item">
            <el-card shadow="hover">
              <div class="edit-remote-row">
                <el-input v-model="remote.name" size="small" placeholder="名称 (如 origin)" style="width: 120px;" />
                <el-radio-group v-model="remoteUrlModes[index]" size="small" class="url-mode-switch-sm">
                  <el-radio-button value="ssh">SSH</el-radio-button>
                  <el-radio-button value="https">HTTPS</el-radio-button>
                </el-radio-group>
                <el-input v-model="remote.fetch_url" size="small" :placeholder="remoteUrlModes[index] === 'ssh' ? 'git@host:user/repo.git' : 'https://host/repo.git'" style="flex: 1;" @blur="validateRemoteUrl(index)" :class="{ 'is-error-input': remoteUrlErrors[index] }" />
                <el-button size="small" :icon="Connection" circle @click="testEditRemote(index)" title="测试连接" :loading="remote._testing" />
                <el-button size="small" :icon="Delete" circle type="danger" @click="removeEditRemote(index)" title="删除" />
              </div>
              <div v-if="remoteUrlErrors[index]" class="field-error" style="margin-left: 128px;">{{ remoteUrlErrors[index] }}</div>
              <div class="edit-remote-cred">
                <span class="cred-label">凭证:</span>
                <CredentialSelector
                  :model-value="editRemoteCredentials?.[remote.name]"
                  :url="remote.fetch_url"
                  placeholder="选择凭证（可选）"
                  @update:model-value="(v: number | undefined) => updateEditRemoteCred(remote.name, v)"
                />
              </div>
            </el-card>
          </div>
          <el-empty v-if="editRemotes.length === 0" description="无远程仓库配置" :image-size="60" />
        </div>
      </el-form-item>

      <el-form-item v-if="editTrackingBranches.length > 0" label="分支追踪">
        <div class="tracking-branches">
          <el-tag v-for="b in editTrackingBranches" :key="b.name" size="small" style="margin: 2px 4px;">
            {{ b.name }} -> {{ b.upstream_ref }}
          </el-tag>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleSaveEdit" :loading="editSaving">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Connection } from '@element-plus/icons-vue'
import { scanRepo, updateRepo } from '@/api/modules/repo'
import { testConnection } from '@/api/modules/system'
import { testCredential } from '@/api/modules/credential'
import type { RepoDTO, GitRemote, TrackingBranch } from '@/types/repo'
import { validateGitRemoteUrl, detectGitProtocol, convertGitUrl } from '@/utils/git'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'

interface EditRemoteRow extends GitRemote {
  _testing?: boolean
}

const props = defineProps<{
  visible: boolean
  repo: RepoDTO | null
  repoKey: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const editSaving = ref(false)
const editForm = ref({ name: '', path: '', remote_url: '' })
const editRemotes = ref<EditRemoteRow[]>([])
const editTrackingBranches = ref<TrackingBranch[]>([])
const editDefaultCredentialId = ref<number | undefined>()
const editRemoteCredentials = ref<Record<string, number | undefined>>()
const editUrlError = ref('')
const remoteUrlErrors = ref<Record<number, string>>({})
const editUrlMode = ref<'ssh' | 'https'>('ssh')
const remoteUrlModes = ref<Record<number, 'ssh' | 'https'>>({})

watch(() => props.visible, (val) => {
  if (!val || !props.repo) return
  editForm.value = {
    name: props.repo.name,
    path: props.repo.path,
    remote_url: props.repo.remote_url || '',
  }
  editRemotes.value = []
  editTrackingBranches.value = []
  editDefaultCredentialId.value = props.repo.default_credential_id
  editRemoteCredentials.value = { ...(props.repo.remote_credentials || {}) }
  editUrlError.value = ''
  remoteUrlErrors.value = {}
  remoteUrlModes.value = {}
  const mainProto = detectGitProtocol(props.repo.remote_url || '')
  editUrlMode.value = mainProto === 'http' ? 'https' : 'ssh'

  if (props.repo.path) {
    scanRepo(props.repo.path).then(result => {
      editRemotes.value = (result.remotes || []).map(r => ({
        ...r,
        _testing: false,
      }))
      editTrackingBranches.value = result.branches || []
      editRemotes.value.forEach((r, i) => {
        const p = detectGitProtocol(r.fetch_url || '')
        remoteUrlModes.value[i] = p === 'http' ? 'https' : 'ssh'
      })
      if (!editForm.value.remote_url && editRemotes.value.length > 0) {
        editForm.value.remote_url = editRemotes.value[0]!.fetch_url
        const p = detectGitProtocol(editForm.value.remote_url)
        editUrlMode.value = p === 'http' ? 'https' : 'ssh'
      }
    }).catch(() => {})
  }
})

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

async function handleSaveEdit() {
  if (!editForm.value.name || !editForm.value.path) {
    ElMessage.warning('名称和路径不能为空')
    return
  }
  if (editForm.value.remote_url) {
    const err = validateGitRemoteUrl(editForm.value.remote_url)
    if (err) {
      editUrlError.value = err
      return
    }
  }
  for (let i = 0; i < editRemotes.value.length; i++) {
    const r = editRemotes.value[i]!
    if (r.fetch_url) {
      const err = validateGitRemoteUrl(r.fetch_url)
      if (err) {
        remoteUrlErrors.value[i] = err
        ElMessage.warning(`远程 "${r.name || 'unnamed'}" 的 URL 格式不正确`)
        return
      }
    }
  }
  editSaving.value = true
  try {
    const remotes: GitRemote[] = editRemotes.value
      .filter(r => r.name && r.fetch_url)
      .map(r => ({
        name: r.name,
        fetch_url: r.fetch_url,
        push_url: r.push_url || r.fetch_url,
        is_mirror: r.is_mirror,
      }))
    const rc: Record<string, number> = {}
    for (const [k, v] of Object.entries(editRemoteCredentials.value ?? {})) {
      if (v) rc[k] = v
    }
    await updateRepo({
      key: props.repoKey,
      name: editForm.value.name,
      path: editForm.value.path,
      remote_url: editForm.value.remote_url || undefined,
      remotes,
      default_credential_id: editDefaultCredentialId.value,
      remote_credentials: Object.keys(rc).length > 0 ? rc : undefined,
    })
    ElMessage.success('保存成功')
    emit('update:visible', false)
    emit('saved')
  } finally {
    editSaving.value = false
  }
}

function addEditRemote() {
  editRemotes.value.push({
    name: '',
    fetch_url: '',
    push_url: '',
    is_mirror: false,
    _testing: false,
  })
}

function updateEditRemoteCred(name: string, val: number | undefined) {
  if (val) {
    editRemoteCredentials.value![name] = val
  } else {
    delete editRemoteCredentials.value![name]
  }
}

function validateEditUrl() {
  const url = editForm.value.remote_url
  if (!url) {
    editUrlError.value = ''
    return
  }
  const proto = detectGitProtocol(url)
  if (proto === 'ssh') editUrlMode.value = 'ssh'
  else if (proto === 'http') editUrlMode.value = 'https'
  editUrlError.value = validateGitRemoteUrl(url)
}

function validateRemoteUrl(index: number) {
  const remote = editRemotes.value[index]
  if (!remote) return
  if (!remote.fetch_url) {
    delete remoteUrlErrors.value[index]
    return
  }
  const proto = detectGitProtocol(remote.fetch_url)
  if (proto === 'ssh') remoteUrlModes.value[index] = 'ssh'
  else if (proto === 'http') remoteUrlModes.value[index] = 'https'
  const err = validateGitRemoteUrl(remote.fetch_url)
  if (err) {
    remoteUrlErrors.value[index] = err
  } else {
    delete remoteUrlErrors.value[index]
  }
}

function removeEditRemote(index: number) {
  editRemotes.value.splice(index, 1)
  delete remoteUrlErrors.value[index]
  delete remoteUrlModes.value[index]
}

async function testEditRemote(index: number) {
  const row = editRemotes.value[index]
  if (!row || !row.fetch_url) {
    ElMessage.warning('请输入 Fetch URL')
    return
  }
  row._testing = true
  try {
    const credential_id = editRemoteCredentials.value?.[row.name]
    if (credential_id) {
      const result = await testCredential(credential_id, row.fetch_url)
      if (result.success) {
        ElMessage.success(`${row.name || 'Remote'} 连接成功`)
      } else {
        ElMessage.error('连接失败: ' + (result.message || '未知错误'))
      }
    } else if (editDefaultCredentialId.value) {
      const result = await testCredential(editDefaultCredentialId.value, row.fetch_url)
      if (result.success) {
        ElMessage.success(`${row.name || 'Remote'} 连接成功`)
      } else {
        ElMessage.error('连接失败: ' + (result.message || '未知错误'))
      }
    } else {
      const result = await testConnection(row.fetch_url)
      if (result.status === 'success') {
        ElMessage.success(`${row.name || 'Remote'} 连接成功`)
      } else {
        ElMessage.error('连接失败: ' + (result.error || '未知错误'))
      }
    }
  } catch (e: any) {
    ElMessage.error('连接测试请求失败: ' + (e?.message || ''))
  } finally {
    row._testing = false
  }
}
</script>

<style scoped>
.url-input-group {
  display: flex;
  gap: var(--spacing-sm);
  width: 100%;
}
.url-input-group .el-input {
  flex: 1;
}
.url-mode-switch {
  flex-shrink: 0;
}
.url-mode-switch-sm {
  flex-shrink: 0;
}
.edit-remote-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}
.edit-remote-cred {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}
.edit-remote-cred .cred-label {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  flex-shrink: 0;
}
.remotes-section {
  width: 100%;
}
.remotes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: var(--font-size-md);
  color: var(--text-color-regular);
}
.tracking-branches {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}
.field-error {
  color: var(--danger-color);
  font-size: var(--font-size-xs);
  margin-top: var(--spacing-xs);
}
.is-error-input :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--danger-color) inset;
}
.edit-remote-item {
  margin-bottom: 8px;
}
</style>
