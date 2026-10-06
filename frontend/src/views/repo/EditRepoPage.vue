<template>
  <div class="edit-repo-page">
    <PageHeader title="编辑仓库" :showBack="true" :back-route="ROUTES.RepoDetail(repo_key)" />

    <FormCard v-if="!pageLoading">
      <div class="form-field">
        <label class="field-label">仓库名称</label>
        <input v-model="editForm.name" placeholder="仓库名称" class="field-input" />
      </div>

      <div class="form-field">
        <label class="field-label">本地路径</label>
        <input v-model="editForm.path" placeholder="本地仓库路径" class="field-input" />
      </div>

      <div class="form-field">
        <label class="field-label">远程 URL</label>
        <div class="url-row">
          <div class="proto-toggle">
            <button class="proto-btn" :class="{ active: editUrlMode === 'ssh' }" @click="switchMainProto('ssh')">SSH</button>
            <button class="proto-btn" :class="{ active: editUrlMode === 'https' }" @click="switchMainProto('https')">HTTPS</button>
          </div>
          <input v-model="editForm.remote_url" :placeholder="editUrlMode === 'ssh' ? 'git@github.com:user/repo.git' : 'https://github.com/user/repo.git'" class="field-input url-input" @blur="validateMainUrl" />
        </div>
        <div v-if="editUrlError" class="field-error">{{ editUrlError }}</div>
      </div>

      <div class="form-field">
        <label class="field-label">默认凭证</label>
        <CredentialSelector v-model="editDefaultCredentialId" :url="editForm.remote_url" placeholder="选择默认凭证（可选）" />
      </div>

      <div class="spacer"></div>

      <div class="config-header">
        <span class="config-title">远程仓库配置</span>
        <div class="header-tags" v-if="editRemotes.length > 0">
          <span class="type-badge" v-for="r in editRemotes" :key="r.name">{{ r.name }}</span>
        </div>
      </div>

      <div class="remote-list">
        <div v-for="(remote, index) in editRemotes" :key="index" class="remote-card">
          <div class="remote-row">
            <input v-model="remote.name" placeholder="名称" class="field-input remote-name-input" />
            <div class="proto-toggle proto-toggle--sm">
              <button class="proto-btn proto-btn--sm" :class="{ active: remoteUrlModes[index] === 'ssh' }" @click="switchRemoteProto(index, 'ssh')">SSH</button>
              <button class="proto-btn proto-btn--sm" :class="{ active: remoteUrlModes[index] === 'https' }" @click="switchRemoteProto(index, 'https')">HTTPS</button>
            </div>
            <input v-model="remote.fetch_url" :placeholder="remoteUrlModes[index] === 'ssh' ? 'git@host:user/repo.git' : 'https://host/repo.git'" class="field-input remote-url-input" @blur="validateRemoteUrl(index)" />
            <button class="icon-btn icon-btn--primary" @click="testEditRemote(index)" :disabled="remote._testing" title="测试连接">
              <el-icon><Connection /><span v-if="remote._testing" class="btn-spinner btn-spinner--sm"></span></el-icon>
            </button>
            <button class="icon-btn icon-btn--danger" @click="removeEditRemote(index)" title="删除">
              <el-icon><Delete /></el-icon>
            </button>
          </div>
          <div v-if="remoteUrlErrors[index]" class="field-error">{{ remoteUrlErrors[index] }}</div>
          <div class="remote-cred-row">
            <span class="cred-label">凭证:</span>
            <CredentialSelector :model-value="editRemoteCredentials[remote.name]" :url="remote.fetch_url" placeholder="选择凭证（可选）" @update:model-value="(v) => updateEditRemoteCred(remote.name, v)" />
          </div>
        </div>
      </div>

      <button class="add-remote-btn" @click="addEditRemote">
        <el-icon><Plus /></el-icon> 新增远程仓库
      </button>

      <template v-if="editTrackingBranches.length > 0">
        <div class="spacer"></div>
        <div class="config-header">
          <span class="config-title">分支追踪</span>
        </div>
        <div class="tracking-tags">
          <span class="track-tag" v-for="b in editTrackingBranches" :key="b.name">{{ b.name }} -> {{ b.upstream_ref }}</span>
        </div>
      </template>

      <div class="spacer"></div>

      <div class="config-header">
        <span class="config-title">远端平台关联</span>
      </div>
      <BindingPanel
        :bindings="bindings"
        @add="openBindingDialogWithProviders"
        @delete="handleDeleteBinding"
        @set-primary="handleSetPrimary"
        @register-webhook="handleRegisterWebhook"
        @delete-webhook="handleDeleteWebhook"
      />

      <BindingDialog
        v-model:visible="showBindingDialog"
        :repo-key="repo_key"
        :providers="availableProviders"
        @created="loadBindings"
      />

      <template #footer>
        <ActionPill variant="outline" @click="router.push(ROUTES.RepoDetail(repo_key))">取消</ActionPill>
        <ActionPill variant="primary" :icon="Check" :disabled="editSaving" @click="handleSaveEdit">
          {{ editSaving ? '保存中...' : '保存' }}
        </ActionPill>
      </template>
    </FormCard>

    <LoadingState v-else />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ROUTES } from '@/router/paths'
import { toastApiError } from '@/composables/useNotification'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Check, Delete, Connection } from '@element-plus/icons-vue'
import { getRepoDetail } from '@/api/modules/repo'
import type { RepoDTO } from '@/types/repo'
import type { RepoProviderBindingDTO } from '@/types/binding'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { useRepoEditForm } from '@/composables/useRepoEditForm'
import BindingPanel from '@/components/binding/BindingPanel.vue'
import BindingDialog from '@/components/binding/BindingDialog.vue'
import { listBindings, deleteBinding, setPrimaryBinding, registerBindingWebhook, deleteBindingWebhook } from '@/api/modules/binding'
import { useProviderStore } from '@/stores/useProviderStore'
import PageHeader from '@/components/common/PageHeader.vue'
import FormCard from '@/components/common/FormCard.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import LoadingState from '@/components/common/LoadingState.vue'

const providerStore = useProviderStore()

const router = useRouter()
const route = useRoute()
const repo_key = route.params.repo_key as string

const pageLoading = ref(false)

// 表单状态与远端 URL 编辑逻辑统一来自 useRepoEditForm（与 RepoEditDialog 共享）
const {
  editSaving, editForm, editUrlError, editRemotes, editTrackingBranches,
  editDefaultCredentialId, editRemoteCredentials, editUrlMode,
  remoteUrlModes, remoteUrlErrors,
  switchMainProto, switchRemoteProto,
  validateMainUrl, validateRemoteUrl,
  addEditRemote, removeEditRemote, updateEditRemoteCred, testEditRemote,
  fillFromRepo, save,
} = useRepoEditForm(() => repo_key)

const bindings = ref<RepoProviderBindingDTO[]>([])
const showBindingDialog = ref(false)
const availableProviders = ref<any[]>([])

async function handleSaveEdit() {
  if (await save()) router.push(ROUTES.RepoDetail(repo_key))
}

async function loadBindings() {
  try { bindings.value = (await listBindings({ repo_key: repo_key })) || [] } catch { bindings.value = [] }
}

async function handleDeleteBinding(id: number) {
  try {
    await ElMessageBox.confirm('确认取消此关联？', '取消关联', { type: 'warning' })
    await deleteBinding(id, true)
    ElMessage.success('关联已取消')
    loadBindings()
  } catch {}
}

async function handleSetPrimary(id: number) {
  try { await setPrimaryBinding(id); ElMessage.success('已设为主关联'); loadBindings() } catch (e) { toastApiError(e, '操作失败') }
}

async function handleRegisterWebhook(id: number) {
  try { await registerBindingWebhook(id); ElMessage.success('Webhook 已注册'); loadBindings() } catch (e: any) { toastApiError(e, '注册失败: ', '') }
}

async function handleDeleteWebhook(id: number) {
  try { await deleteBindingWebhook(id); ElMessage.success('Webhook 已删除'); loadBindings() } catch (e: any) { toastApiError(e, '删除失败: ', '') }
}

async function openBindingDialogWithProviders() {
  try { await providerStore.fetchProviders(); availableProviders.value = providerStore.providers } catch { availableProviders.value = [] }
  showBindingDialog.value = true
}

onMounted(async () => {
  pageLoading.value = true
  try {
    const repo: RepoDTO = await getRepoDetail(repo_key)
    fillFromRepo(repo)
    loadBindings()
  } catch (e) {
    toastApiError(e, '加载仓库失败')
    router.push(ROUTES.RepoDetail(repo_key))
  } finally {
    pageLoading.value = false
  }
})
</script>

<style scoped>
.edit-repo-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.field-error {
  color: #EF4444;
  font-size: 12px;
}

.url-row {
  display: flex;
  align-items: stretch;
}

.proto-toggle {
  display: flex;
  border: 1px solid var(--border-color);
  border-radius: 6px 0 0 6px;
  overflow: hidden;
  flex-shrink: 0;
}

.proto-btn {
  padding: 10px 14px;
  border: none;
  background: var(--bg-color-page);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
  border-right: 1px solid var(--border-color);
}

.proto-btn:last-child {
  border-right: none;
}

.proto-btn.active {
  background: var(--accent-primary);
  color: #fff;
}

.proto-btn--sm {
  padding: 8px 12px;
  font-size: 11px;
}

.proto-toggle--sm {
  border-radius: 6px;
  flex-shrink: 0;
}

.url-input {
  border-radius: 0 6px 6px 0;
  border-left: none;
}

.spacer {
  height: 16px;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.header-tags {
  display: flex;
  gap: 6px;
}

.type-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  background: var(--accent-bg);
  color: #6366F1;
}

.remote-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.remote-card {
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.remote-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.remote-name-input {
  width: 120px;
  flex-shrink: 0;
}

.remote-url-input {
  flex: 1;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: transparent;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  position: relative;
}

.icon-btn--primary {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.icon-btn--primary:hover:not(:disabled) {
  background: var(--accent-primary);
  color: #fff;
}

.icon-btn--danger {
  border-color: #EF4444;
  color: #EF4444;
}

.icon-btn--danger:hover {
  background: #EF4444;
  color: #fff;
}

.icon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.remote-cred-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.cred-label {
  font-size: 12px;
  color: var(--text-color-secondary);
  flex-shrink: 0;
}

.add-remote-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
  align-self: flex-start;
}

.add-remote-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.tracking-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.track-tag {
  padding: 4px 8px;
  border-radius: 4px;
  background: var(--accent-bg);
  color: #6366F1;
  font-size: 12px;
}

</style>
