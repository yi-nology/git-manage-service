<template>
  <div class="register-page">
    <PageHeader title="注册本地仓库" :showBack="true" backRoute="/local-repos" />

    <div class="tab-bar">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="tab-item"
        :class="{ active: activeTab === tab.key }"
        @click="switchTab(tab.key as 'single' | 'batch')"
      >{{ tab.label }}</button>
    </div>

    <FormCard v-if="activeTab === 'single'">
      <div class="form-field">
        <label class="field-label">仓库名称</label>
        <input v-model="singleForm.name" placeholder="my-project" class="field-input" />
      </div>

      <div class="form-field">
        <label class="field-label">本地路径</label>
        <div class="path-row">
          <input v-model="singleForm.path" placeholder="/home/user/local-repos/my-project" class="field-input" readonly />
          <button class="browse-btn" @click="handleSelectDir" :disabled="selectingDir">
            {{ selectingDir ? '选择中...' : '浏览' }}
          </button>
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">远程 URL (可选)</label>
        <input v-model="singleForm.remoteUrl" placeholder="git@github.com:user/my-project.git" class="field-input" />
      </div>

      <div class="form-field" v-if="singleRepoInfo">
        <label class="field-label">仓库状态</label>
        <div class="repo-status-row">
          <span class="status-tag tag-success">有效 Git 仓库</span>
          <span class="status-tag tag-info">{{ singleRepoInfo.current_branch || 'unknown' }}</span>
          <span v-if="singleRepoInfo.remotes?.length" class="status-tag tag-info">
            {{ singleRepoInfo.remotes[0]?.name }}: {{ simplifyUrl(singleRepoInfo.remotes[0]?.fetch_url || '') }}
          </span>
          <span :class="singleRepoInfo.has_changes ? 'status-tag tag-amber' : 'status-tag tag-success'">
            {{ singleRepoInfo.has_changes ? '有更改' : '干净' }}
          </span>
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">认证凭证 (可选)</label>
        <CredentialSelector
          v-model="singleForm.credentialId"
          placeholder="选择认证凭证（可选）"
          style="width: 100%; max-width: 400px"
        />
      </div>

      <div class="form-actions">
        <ActionPill variant="outline" @click="router.push('/local-repos')">取消</ActionPill>
        <ActionPill variant="primary" :icon="Check" @click="handleRegisterSingle" :disabled="registering">
          {{ registering ? '注册中...' : '注册仓库' }}
        </ActionPill>
      </div>
    </FormCard>

    <BatchRegisterPanel v-if="activeTab === 'batch'" @registered="() => {}" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Check } from '@element-plus/icons-vue'
import { selectDirectory, scanDirectory, createRepo } from '@/api/modules/repo'
import type { ScannedRepo } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { useNotification } from '@/composables/useNotification'
import PageHeader from '@/components/common/PageHeader.vue'
import FormCard from '@/components/common/FormCard.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import BatchRegisterPanel from '@/components/repo/BatchRegisterPanel.vue'

const router = useRouter()
const { showSuccess, showError } = useNotification()

const tabs = [
  { key: 'single', label: '单个注册' },
  { key: 'batch', label: '批量注册' },
]
const activeTab = ref<'single' | 'batch'>('single')

const selectingDir = ref(false)
const scanning = ref(false)
const registering = ref(false)

const singleForm = reactive({
  name: '',
  path: '',
  remoteUrl: '',
  credentialId: undefined as number | undefined,
})
const singleRepoInfo = ref<ScannedRepo | null>(null)

function switchTab(key: 'single' | 'batch') {
  activeTab.value = key
}

async function handleSelectDir() {
  selectingDir.value = true
  try {
    const title = '选择 Git 仓库根目录'
    const res = await selectDirectory(title)
    if (res.cancelled !== 'true' && res.path) {
      singleForm.path = res.path
      singleForm.name = res.path.split('/').pop() || ''
      await autoDetectSingleRepo()
    }
  } catch (e: any) {
    const msg = e?.message || String(e)
    if (msg.includes('取消') || msg.includes('-128') || msg.includes('cancelled')) return
    showError('选择目录失败', e)
  } finally {
    selectingDir.value = false
  }
}

async function autoDetectSingleRepo() {
  if (!singleForm.path) return
  scanning.value = true
  try {
    const res = await scanDirectory(singleForm.path, 0, false)
    const repos = res.repos || []
    if (repos.length > 0) {
      singleRepoInfo.value = repos[0]!
      if (!singleForm.name) singleForm.name = repos[0]!.name
      showSuccess('检测到有效的 Git 仓库')
    } else {
      singleRepoInfo.value = null
      showError('该目录不是有效的 Git 仓库')
    }
  } catch (e: any) {
    showError('验证失败', e)
    singleRepoInfo.value = null
  } finally {
    scanning.value = false
  }
}

function simplifyUrl(url: string): string {
  if (!url) return ''
  if (url.startsWith('https://github.com/')) return url.replace('https://github.com/', 'github:')
  if (url.startsWith('git@github.com:')) return url.replace('git@github.com:', 'github:')
  return url
}

async function handleRegisterSingle() {
  const name = singleForm.name.trim()
  if (!name) { showError('请输入仓库名称'); return }
  if (!singleForm.path) { showError('请选择仓库路径'); return }

  registering.value = true
  try {
    await createRepo({
      name,
      path: singleForm.path,
      default_credential_id: singleForm.credentialId,
    })
    showSuccess(`仓库 "${name}" 注册成功`)
    router.push('/local-repos')
  } catch (e: any) {
    showError('注册失败', e)
  } finally {
    registering.value = false
  }
}
</script>

<style scoped>
.register-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.tab-bar {
  display: flex;
  border-bottom: 1px solid var(--border-color);
}

.tab-item {
  padding: 10px 20px;
  font-size: 14px;
  font-weight: normal;
  color: var(--text-color-secondary);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-item:hover {
  color: var(--accent-primary);
}

.tab-item.active {
  color: var(--accent-primary);
  font-weight: 500;
  border-bottom-color: var(--accent-primary);
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.field-input {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  width: 100%;
  box-sizing: border-box;
}

.field-input:focus {
  border-color: var(--accent-primary);
}

.path-row {
  display: flex;
  gap: 8px;
}

.path-row .field-input {
  flex: 1;
}

.browse-btn {
  padding: 10px 16px;
  border-radius: 6px;
  border: none;
  background: var(--accent-bg);
  color: var(--accent-primary);
  font-size: 13px;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.2s;
}

.browse-btn:hover:not(:disabled) {
  opacity: 0.85;
}

.browse-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
}

.tag-success { background: #ECFDF5; color: #10B981; }
.tag-info { background: var(--accent-bg); color: #6366F1; }
.tag-amber { background: #FFFBEB; color: #F59E0B; }

.repo-status-row {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
