<template>
  <div class="sync-config-panel">
    <div class="panel-header">
      <div class="header-left">
        <h2 class="page-title">镜像同步</h2>
        <p class="page-subtitle">配置仓库的双向镜像同步，支持 Pull 和 Push 两种模式</p>
      </div>
      <div class="header-actions">
        <ActionPill variant="outline" :icon="Refresh" @click="loadMirrors">刷新</ActionPill>
        <ActionPill variant="primary" :icon="Plus" @click="showCreateDialog('pull')">
          <span class="btn-content">
            <el-icon><Download /></el-icon>
            新建 Pull Mirror
          </span>
        </ActionPill>
        <ActionPill variant="amber" :icon="Plus" @click="showCreateDialog('push')">
          <span class="btn-content">
            <el-icon><Upload /></el-icon>
            新建 Push Mirror
          </span>
        </ActionPill>
      </div>
    </div>

    <el-empty v-if="!loading && mirrors.length === 0" description="暂无镜像配置" :image-size="120">
      <div class="empty-actions">
        <el-button type="primary" @click="showCreateDialog('pull')">
          <el-icon style="margin-right: 6px;"><Download /></el-icon>
          创建 Pull Mirror
        </el-button>
        <el-button @click="showCreateDialog('push')">
          <el-icon style="margin-right: 6px;"><Upload /></el-icon>
          创建 Push Mirror
        </el-button>
      </div>
    </el-empty>

    <div v-if="mirrors.length > 0" class="mirror-grid">
      <MirrorCard
        v-for="mirror in mirrors"
        :key="mirror.id"
        :mirror="mirror"
        :syncing="syncingId === mirror.id"
        :updating="updatingId === mirror.id"
        @sync="triggerSync"
        @show-logs="showLogs"
        @edit="editMirror"
        @delete="deleteMirror"
        @toggle-enabled="toggleEnabled"
      />
    </div>

    <MirrorFormDialog
      v-model:visible="dialogVisible"
      :editing-mirror="editingMirror"
      :create-type="createType"
      :form="form"
      :selected-remote="selectedRemote"
      :selected-branches="selectedBranches"
      :credentials="credentials"
      :repo-remotes="repoRemotes"
      :repo-branches="repoBranches"
      :saving="saving"
      @update:selected-remote="onRemoteChange"
      @update:selected-branches="onBranchesChange"
      @submit="submitForm"
    />

    <SyncLogDialog
      v-model:visible="logDialogVisible"
      :loading="logLoading"
      :sync-logs="syncLogs"
      @refresh="loadSyncLogs"
      @show-detail="showLogDetail"
    />

    <SyncLogDetailDialog
      v-model:visible="logDetailVisible"
      :current-log="currentLog"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Download, Upload } from '@element-plus/icons-vue'
import * as mirrorApi from '@/api/modules/mirror'
import { listCredentials } from '@/api/modules/credential'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getBranchList } from '@/api/modules/branch'
import type { MirrorDTO, MirrorSyncLogDTO, CreateMirrorReq, UpdateMirrorReq } from '@/types/mirror'
import ActionPill from '@/components/common/ActionPill.vue'
import MirrorCard from './config/MirrorCard.vue'
import MirrorFormDialog from './config/MirrorFormDialog.vue'
import SyncLogDialog from './config/SyncLogDialog.vue'
import SyncLogDetailDialog from './config/SyncLogDetailDialog.vue'

const props = defineProps<{ repoKey: string }>()

const loading = ref(false)
const saving = ref(false)
const syncingId = ref<number | null>(null)
const updatingId = ref<number | null>(null)
const mirrors = ref<MirrorDTO[]>([])
const credentials = ref<{ id: number; name: string }[]>([])
const repoRemotes = ref<{ name: string; url: string }[]>([])
const repoBranches = ref<string[]>([])
const selectedRemote = ref<{ name: string; url: string } | null>(null)
const selectedBranches = ref<string[]>([])
const currentRepoId = ref<number>(0)

const dialogVisible = ref(false)
const editingMirror = ref<MirrorDTO | null>(null)
const createType = ref<'pull' | 'push'>('pull')

const logDialogVisible = ref(false)
const logLoading = ref(false)
const syncLogs = ref<MirrorSyncLogDTO[]>([])
const currentLogMirrorId = ref<number>(0)

const logDetailVisible = ref(false)
const currentLog = ref<MirrorSyncLogDTO | null>(null)

const form = ref({
  remote_url: '',
  remote_name: 'origin',
  credential_id: null as number | null,
  branch_filter: '',
  sync_interval: 600,
  cron_expr: '',
  sync_on_push: false,
  git_force: false,
  git_prune: true,
  git_tags: true,
  enabled: true,
})

onMounted(async () => {
  await loadRepoInfo()
  await Promise.all([loadMirrors(), loadCredentials()])
})

async function loadRepoInfo() {
  try {
    const repo = await getRepoDetail(props.repoKey)
    currentRepoId.value = repo.id
  } catch {}
}

async function loadMirrors() {
  loading.value = true
  try {
    const all = await mirrorApi.getMirrors(currentRepoId.value || undefined)
    mirrors.value = all || []
  } catch (e: any) {
    ElMessage.error('加载镜像列表失败: ' + e.message)
  } finally {
    loading.value = false
  }
}

async function loadCredentials() {
  try {
    const list = await listCredentials()
    credentials.value = (list || []).map((c: any) => ({ id: c.id, name: c.name }))
  } catch {}
}

async function loadRepoRemoteInfo() {
  try {
    const repo = await getRepoDetail(props.repoKey)
    if (repo?.path) {
      const scan = await scanRepo(repo.path)
      repoRemotes.value = (scan.remotes || []).map((r: any) => ({
        name: r.name,
        url: r.fetch_url || r.url,
      }))
    }
    const branches = await getBranchList(props.repoKey, { page_size: 200 })
    repoBranches.value = (branches?.list || []).map((b: any) => b.name)
  } catch {}
}

function onRemoteChange(remote: { name: string; url: string } | null) {
  selectedRemote.value = remote
  if (remote) {
    form.value.remote_url = remote.url
    form.value.remote_name = remote.name
  }
}

function onBranchesChange(branches: string[]) {
  selectedBranches.value = branches
  form.value.branch_filter = branches.join(', ')
}

async function showCreateDialog(type: 'pull' | 'push' = 'pull') {
  editingMirror.value = null
  createType.value = type
  selectedRemote.value = null
  selectedBranches.value = []
  form.value = {
    remote_url: '',
    remote_name: 'origin',
    credential_id: null,
    branch_filter: '',
    sync_interval: 600,
    cron_expr: '',
    sync_on_push: false,
    git_force: false,
    git_prune: true,
    git_tags: true,
    enabled: true,
  }
  await loadRepoRemoteInfo()
  if (repoRemotes.value.length > 0 && type === 'pull') {
    const remote = repoRemotes.value[0]
    if (remote) {
      selectedRemote.value = remote
      form.value.remote_url = remote.url
      form.value.remote_name = remote.name
    }
  }
  dialogVisible.value = true
}

async function editMirror(mirror: MirrorDTO) {
  editingMirror.value = mirror
  createType.value = mirror.mirror_type
  await loadRepoRemoteInfo()
  selectedBranches.value = mirror.branch_filter
    ? mirror.branch_filter.split(',').map((b: string) => b.trim()).filter(Boolean)
    : []
  const matchedRemote = repoRemotes.value.find((r: any) => r.name === mirror.remote_name)
  selectedRemote.value = matchedRemote || null
  form.value = {
    remote_url: mirror.remote_url,
    remote_name: mirror.remote_name,
    credential_id: mirror.credential_id,
    branch_filter: mirror.branch_filter,
    sync_interval: mirror.sync_interval,
    cron_expr: mirror.cron_expr,
    sync_on_push: mirror.sync_on_push,
    git_force: mirror.git_force,
    git_prune: mirror.git_prune,
    git_tags: mirror.git_tags,
    enabled: mirror.enabled,
  }
  dialogVisible.value = true
}

async function submitForm() {
  if (!form.value.remote_url) {
    ElMessage.warning('请输入远程 URL')
    return
  }
  saving.value = true
  try {
    if (editingMirror.value) {
      const data: UpdateMirrorReq = { ...form.value }
      await mirrorApi.updateMirror(editingMirror.value.id, data)
      ElMessage.success('更新成功')
    } else {
      const data: CreateMirrorReq = {
        repo_id: currentRepoId.value,
        mirror_type: createType.value,
        ...form.value,
      }
      await mirrorApi.createMirror(data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadMirrors()
  } catch (e: any) {
    ElMessage.error('操作失败: ' + e.message)
  } finally {
    saving.value = false
  }
}

async function deleteMirror(mirror: MirrorDTO) {
  try {
    await ElMessageBox.confirm('确认删除此镜像？', '删除确认', { type: 'warning' })
    await mirrorApi.deleteMirror(mirror.id)
    ElMessage.success('已删除')
    loadMirrors()
  } catch {}
}

async function triggerSync(mirror: MirrorDTO) {
  syncingId.value = mirror.id
  try {
    await mirrorApi.triggerMirrorSync(mirror.id)
    ElMessage.success('同步已触发，任务正在后台执行')
    setTimeout(() => {
      loadMirrors()
      if (logDialogVisible.value && currentLogMirrorId.value === mirror.id) {
        loadSyncLogs()
      }
    }, 2000)
  } catch (e: any) {
    ElMessage.error('触发失败: ' + e.message)
  } finally {
    setTimeout(() => {
      syncingId.value = null
    }, 3000)
  }
}

async function toggleEnabled(mirror: MirrorDTO) {
  updatingId.value = mirror.id
  try {
    if (mirror.status === 'paused' && mirror.enabled) {
      await mirrorApi.resumeMirror(mirror.id)
    } else if (!mirror.enabled) {
      await mirrorApi.pauseMirror(mirror.id)
    } else {
      await mirrorApi.updateMirror(mirror.id, { enabled: mirror.enabled })
    }
    ElMessage.success('状态已更新')
    loadMirrors()
  } catch (e: any) {
    ElMessage.error('更新失败: ' + e.message)
    mirror.enabled = !mirror.enabled
  } finally {
    updatingId.value = null
  }
}

async function showLogs(mirror: MirrorDTO) {
  currentLogMirrorId.value = mirror.id
  logDialogVisible.value = true
  await loadSyncLogs()
}

async function loadSyncLogs() {
  if (!currentLogMirrorId.value) return
  logLoading.value = true
  try {
    syncLogs.value = await mirrorApi.getMirrorSyncLogs(currentLogMirrorId.value, 50)
  } catch (e: any) {
    ElMessage.error('加载日志失败')
  } finally {
    logLoading.value = false
  }
}

async function showLogDetail(log: MirrorSyncLogDTO) {
  try {
    currentLog.value = await mirrorApi.getSyncLogDetail(log.id)
    logDetailVisible.value = true
  } catch (e: any) {
    currentLog.value = log
    logDetailVisible.value = true
  }
}
</script>

<style scoped>
.sync-config-panel {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 8px 0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.btn-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.empty-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-top: 16px;
}

.mirror-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: 16px;
}

@media (max-width: 768px) {
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .mirror-grid {
    grid-template-columns: 1fr;
  }
}
</style>
