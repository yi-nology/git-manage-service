<template>
  <FormCard>
    <div class="form-field">
      <label class="field-label">父目录路径</label>
      <div class="path-row">
        <input v-model="batchForm.path" placeholder="选择包含多个仓库的父目录" class="field-input" readonly />
        <button class="browse-btn" @click="handleSelectDir" :disabled="selectingDir">
          {{ selectingDir ? '选择中...' : '浏览' }}
        </button>
      </div>
    </div>

    <div class="form-actions" style="margin-top: -8px">
      <ActionPill
        variant="primary"
        :icon="Search"
        @click="handleScan"
        :disabled="!batchForm.path || scanning"
      >
        {{ scanning ? '扫描中...' : '扫描仓库' }}
      </ActionPill>
    </div>

    <div v-if="scannedRepos.length > 0" class="scanned-section">
      <div class="section-header">
        <span class="section-title">选择要注册的仓库 (共 {{ scannedRepos.length }} 个)</span>
        <div class="header-actions">
          <ActionPill variant="outline" small @click="selectAll">全选</ActionPill>
          <ActionPill variant="outline" small @click="selectNone">取消全选</ActionPill>
        </div>
      </div>

      <div class="repo-list">
        <div
          v-for="repo in scannedRepos"
          :key="repo.path"
          class="repo-item"
          :class="{ selected: selectedRepos.includes(repo.path) }"
          @click="toggleRepo(repo.path)"
        >
          <el-checkbox
            :model-value="selectedRepos.includes(repo.path)"
            @click.stop
            @change="toggleRepo(repo.path)"
          />
          <div class="repo-info">
            <div class="repo-name">
              <el-icon><FolderChecked /></el-icon>
              {{ repo.name }}
            </div>
            <div class="repo-meta">
              <span class="status-tag tag-info">{{ repo.current_branch || 'unknown' }}</span>
              <span class="repo-path">{{ repo.path }}</span>
            </div>
            <div class="repo-remote" v-if="repo.remotes.length > 0">
              <el-icon><Link /></el-icon>
              {{ getMainRemote(repo) }}
            </div>
          </div>
          <div class="repo-status">
            <span v-if="repo.has_changes" class="status-tag tag-amber">有更改</span>
            <span v-else class="status-tag tag-success">干净</span>
          </div>
        </div>
      </div>

      <div class="selection-info">
        已选择 <strong>{{ selectedRepos.length }}</strong> 个仓库
      </div>
    </div>

    <div v-if="selectedRepos.length > 0" class="form-field">
      <label class="field-label">默认凭证 (可选)</label>
      <CredentialSelector
        v-model="batchForm.credential_id"
        placeholder="选择默认认证凭证（可选）"
        style="width: 100%"
      />
    </div>

    <div v-if="selectedRepos.length > 0" class="form-actions">
      <ActionPill variant="outline" @click="router.push('/local-repos')">取消</ActionPill>
      <ActionPill variant="primary" :icon="Check" @click="handleRegister" :disabled="registering">
        {{ registering ? '注册中...' : `注册 ${selectedRepos.length} 个仓库` }}
      </ActionPill>
    </div>

    <div class="empty-card" v-if="!scanning && scannedRepos.length === 0 && batchForm.path && hasScanned">
      <span class="empty-text">在该目录下未找到 Git 仓库</span>
      <ActionPill variant="outline" @click="handleSelectDir">选择其他目录</ActionPill>
    </div>
  </FormCard>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import {
  FolderChecked,
  Search,
  Link,
  Check,
} from '@element-plus/icons-vue'
import { selectDirectory, scanDirectory, batchCreateRepos } from '@/api/modules/repo'
import type { ScannedRepo } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { useNotification } from '@/composables/useNotification'
import FormCard from '@/components/common/FormCard.vue'
import ActionPill from '@/components/common/ActionPill.vue'

const emit = defineEmits<{
  'registered': []
}>()

const router = useRouter()
const { showSuccess, showError } = useNotification()

const selectingDir = ref(false)
const scanning = ref(false)
const registering = ref(false)

const batchForm = reactive({
  path: '',
  credential_id: undefined as number | undefined,
})
const scannedRepos = ref<ScannedRepo[]>([])
const selectedRepos = ref<string[]>([])
const hasScanned = ref(false)

async function handleSelectDir() {
  selectingDir.value = true
  try {
    const res = await selectDirectory('选择包含 Git 仓库的父目录')
    if (res.cancelled !== 'true' && res.path) {
      batchForm.path = res.path
      hasScanned.value = false
      scannedRepos.value = []
      selectedRepos.value = []
    }
  } catch (e: any) {
    const msg = e?.message || String(e)
    if (msg.includes('取消') || msg.includes('-128') || msg.includes('cancelled')) return
    showError('选择目录失败', e)
  } finally {
    selectingDir.value = false
  }
}

async function handleScan() {
  if (!batchForm.path) return
  scanning.value = true
  try {
    const res = await scanDirectory(batchForm.path, 2, true)
    scannedRepos.value = res.repos || []
    selectedRepos.value = scannedRepos.value.map(r => r.path)
    hasScanned.value = true
    if (res.total > 0) showSuccess(`找到 ${res.total} 个 Git 仓库`)
  } catch (e: any) {
    showError('扫描失败', e)
  } finally {
    scanning.value = false
  }
}

function toggleRepo(path: string) {
  const idx = selectedRepos.value.indexOf(path)
  if (idx >= 0) selectedRepos.value.splice(idx, 1)
  else selectedRepos.value.push(path)
}

function selectAll() { selectedRepos.value = scannedRepos.value.map(r => r.path) }
function selectNone() { selectedRepos.value = [] }

function getMainRemote(repo: ScannedRepo): string {
  const origin = repo.remotes.find(r => r.name === 'origin')
  const remote = origin || repo.remotes[0]
  if (!remote) return '无远程'
  return simplifyUrl(remote.fetch_url)
}

function simplifyUrl(url: string): string {
  if (!url) return ''
  if (url.startsWith('https://github.com/')) return url.replace('https://github.com/', 'github:')
  if (url.startsWith('git@github.com:')) return url.replace('git@github.com:', 'github:')
  return url
}

async function handleRegister() {
  if (selectedRepos.value.length === 0) { showError('请至少选择一个仓库'); return }

  registering.value = true
  try {
    const repos = selectedRepos.value.map(path => {
      const repo = scannedRepos.value.find(r => r.path === path)!
      return { name: repo.name, path: repo.path, default_credential_id: batchForm.credential_id }
    })
    const res = await batchCreateRepos({ repos })
    const failedList = res.failed || []
    const successList = res.success || []
    if (failedList.length > 0) {
      showError(`${failedList.length} 个仓库注册失败: ${failedList.map(f => f.name).join(', ')}`)
    }
    if (successList.length > 0) {
      showSuccess(`成功注册 ${successList.length} 个仓库`)
      router.push('/local-repos')
      emit('registered')
    }
  } catch (e: any) {
    showError('注册失败', e)
  } finally {
    registering.value = false
  }
}
</script>

<style scoped>
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

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.header-actions {
  display: flex;
  gap: 4px;
}

.scanned-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.repo-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.repo-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
}

.repo-item:hover {
  border-color: var(--accent-primary);
  background: var(--accent-bg);
}

.repo-item.selected {
  border-color: var(--accent-primary);
  background: var(--accent-bg);
}

.repo-info { flex: 1; min-width: 0; }

.repo-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-color-primary);
  margin-bottom: 4px;
}

.repo-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.repo-path {
  color: var(--text-color-secondary);
  font-size: 12px;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.repo-remote {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.repo-status {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.selection-info {
  text-align: center;
  padding: 10px;
  background: var(--accent-bg);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
}

.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 40px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: 12px;
}

.empty-text {
  font-size: 14px;
  color: var(--text-color-secondary);
}

@media (max-width: 768px) {
  .path-row {
    flex-direction: column;
  }
}
</style>
