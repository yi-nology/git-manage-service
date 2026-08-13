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
      <div v-for="mirror in mirrors" :key="mirror.id" class="mirror-card" :class="`type-${mirror.mirror_type}`">
        <div class="card-header">
          <div class="card-title">
            <el-tag :type="mirror.mirror_type === 'pull' ? 'primary' : 'warning'" size="large">
              <el-icon v-if="mirror.mirror_type === 'pull'"><Download /></el-icon>
              <el-icon v-else><Upload /></el-icon>
              {{ mirror.mirror_type.toUpperCase() }}
            </el-tag>
            <el-tag :type="getStatusType(mirror.status)" size="small">
              {{ getStatusLabel(mirror.status) }}
            </el-tag>
          </div>
          <div class="card-actions">
            <el-switch v-model="mirror.enabled" @change="toggleEnabled(mirror)" :loading="updatingId === mirror.id" />
          </div>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="info-label">远程 URL</span>
            <span class="info-value mono">{{ mirror.remote_url }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Remote 名称</span>
            <span class="info-value">{{ mirror.remote_name }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">分支过滤</span>
            <span class="info-value">{{ mirror.branch_filter || '全部分支' }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">同步间隔</span>
            <span class="info-value">{{ mirror.cron_expr || `${mirror.sync_interval}s` }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Git 选项</span>
            <div class="git-options">
              <el-tag v-if="mirror.git_force" size="small" type="danger" effect="plain">--force</el-tag>
              <el-tag v-if="mirror.git_prune" size="small" effect="plain">--prune</el-tag>
              <el-tag v-if="mirror.git_tags" size="small" effect="plain">--tags</el-tag>
              <el-tag v-if="!mirror.git_force && !mirror.git_prune && !mirror.git_tags" size="small" type="info">无</el-tag>
            </div>
          </div>
           <div class="info-row">
             <span class="info-label">上次同步</span>
             <span class="info-value">{{ mirror.last_sync_at ? formatTime(mirror.last_sync_at) : '从未' }}</span>
           </div>
           <div class="info-row">
             <span class="info-label">下次同步</span>
             <span class="info-value" :class="{ 'syncing': syncingId === mirror.id }">
               {{ syncingId === mirror.id ? '同步中...' : (mirror.next_sync_at ? formatTime(mirror.next_sync_at) : '-') }}
             </span>
           </div>
          <div v-if="mirror.last_error" class="info-row error-row">
            <span class="info-label">错误</span>
            <span class="info-value">{{ mirror.last_error }}</span>
          </div>
        </div>

        <div class="card-footer">
          <el-button-group>
            <el-button type="primary" size="small" @click="triggerSync(mirror)" :loading="syncingId === mirror.id">
              <el-icon><Refresh /></el-icon>
              同步
            </el-button>
            <el-button size="small" @click="showLogs(mirror)">
              <el-icon><Document /></el-icon>
              日志
            </el-button>
            <el-button size="small" @click="editMirror(mirror)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button type="danger" size="small" @click="deleteMirror(mirror)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </el-button-group>
        </div>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingMirror ? '编辑镜像' : `创建 ${createType === 'push' ? 'Push' : 'Pull'} 镜像`" width="700px" destroy-on-close class="mirror-dialog">
      <el-form :model="form" label-width="100px">
        <el-form-item label="远程仓库">
          <el-select v-model="selectedRemote" placeholder="选择当前仓库远程" style="width: 100%" @change="onRemoteChange">
            <el-option v-for="r in repoRemotes" :key="r.name" :label="`${r.name} (${r.url})`" :value="r">
              <span style="font-weight: 600;">{{ r.name }}</span>
              <span style="color: var(--text-color-secondary); margin-left: 8px; font-size: 12px;">{{ r.url }}</span>
            </el-option>
          </el-select>
          <div class="form-tip">从当前仓库远程列表中选择，或手动填写下方</div>
        </el-form-item>
        <el-form-item label="远程 URL" required>
          <el-input v-model="form.remote_url" placeholder="https://github.com/user/repo.git" />
        </el-form-item>
        <el-form-item label="Remote 名称">
          <el-input v-model="form.remote_name" placeholder="origin" />
        </el-form-item>
        <el-form-item label="凭据">
          <el-select v-model="form.credential_id" clearable placeholder="选择凭据（可选）" style="width: 100%">
            <el-option v-for="c in credentials" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分支过滤">
          <el-select
            v-model="selectedBranches"
            multiple
            filterable
            allow-create
            placeholder="选择或输入分支"
            style="width: 100%"
            @change="onBranchesChange"
          >
            <el-option v-for="b in repoBranches" :key="b" :label="b" :value="b" />
          </el-select>
          <div class="form-tip">选择要同步的分支，支持手动输入 glob 模式。留空同步全部分支</div>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="同步间隔">
              <el-input-number v-model="form.sync_interval" :min="30" :step="30" style="width: 100%" />
              <span style="margin-left: 8px">秒</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Cron 表达式">
              <el-input v-model="form.cron_expr" placeholder="0 */5 * * *" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="触发设置">
          <el-checkbox v-model="form.sync_on_push">Push 事件自动触发同步</el-checkbox>
        </el-form-item>
        <el-divider content-position="left">Git 选项</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item>
              <el-checkbox v-model="form.git_force">
                <span class="checkbox-label">强制推送 <span class="warning-text">⚠️</span></span>
              </el-checkbox>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item>
              <el-checkbox v-model="form.git_prune">清理已删除分支</el-checkbox>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item>
              <el-checkbox v-model="form.git_tags">同步标签</el-checkbox>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="启用状态">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="saving">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logDialogVisible" title="同步日志" width="900px" destroy-on-close class="log-dialog">
      <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
        <span style="color: var(--text-color-secondary); font-size: 12px;">
          提示：同步任务在后台异步执行，如无数据请点击刷新按钮
        </span>
        <el-button size="small" @click="loadSyncLogs" :loading="logLoading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
      <el-empty v-if="!logLoading && syncLogs.length === 0" description="暂无同步日志" :image-size="120" />
      <el-table v-if="syncLogs.length > 0" :data="syncLogs" v-loading="logLoading" stripe max-height="500">
        <el-table-column label="触发类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTriggerType(row.trigger_type)" size="small">
              {{ getTriggerLabel(row.trigger_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="90">
          <template #default="{ row }">{{ row.duration_ms ? `${(row.duration_ms / 1000).toFixed(1)}s` : '-' }}</template>
        </el-table-column>
        <el-table-column label="分支" prop="branches_synced" width="70" align="center" />
        <el-table-column label="提交" prop="commits_pushed" width="70" align="center" />
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ row.started_at ? formatTime(row.started_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="错误" prop="error_message" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="showLogDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="logDetailVisible" title="日志详情" width="750px" destroy-on-close class="detail-dialog">
      <div v-if="currentLog">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="状态">
            <el-tag :type="currentLog.status === 'success' ? 'success' : 'danger'">{{ currentLog.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="触发类型">{{ getTriggerLabel(currentLog.trigger_type) }}</el-descriptions-item>
          <el-descriptions-item label="耗时">{{ currentLog.duration_ms ? `${(currentLog.duration_ms / 1000).toFixed(2)}s` : '-' }}</el-descriptions-item>
          <el-descriptions-item label="分支 / 提交">
            {{ currentLog.branches_synced || 0 }} / {{ currentLog.commits_pushed || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="错误" :span="2" v-if="currentLog.error_message">
            <span class="error-text">{{ currentLog.error_message }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="currentLog.detail_log" class="log-section">
          <div class="log-section-title">执行日志</div>
          <pre class="log-content">{{ currentLog.detail_log }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Download, Upload, Document, Edit, Delete } from '@element-plus/icons-vue'
import * as mirrorApi from '@/api/modules/mirror'
import { listCredentials } from '@/api/modules/credential'
import { getRepoDetail, scanRepo } from '@/api/modules/repo'
import { getBranchList } from '@/api/modules/branch'
import type { MirrorDTO, MirrorSyncLogDTO, CreateMirrorReq, UpdateMirrorReq } from '@/types/mirror'
import { MIRROR_STATUS_MAP, TRIGGER_TYPE_MAP } from '@/types/mirror'
import ActionPill from '@/components/common/ActionPill.vue'

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

function getStatusType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  const type = MIRROR_STATUS_MAP[status]?.type
  return (type as 'success' | 'warning' | 'danger' | 'info') || 'info'
}

function getStatusLabel(status: string): string {
  return MIRROR_STATUS_MAP[status]?.label || status
}

function getTriggerType(type: string): 'success' | 'warning' | 'danger' | 'primary' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'primary' | 'info'> = {
    manual: 'primary',
    cron: 'warning',
    webhook: 'success',
    push_event: 'info',
  }
  return map[type] || 'info'
}

function getTriggerLabel(type: string): string {
  return TRIGGER_TYPE_MAP[type] || type
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

function onRemoteChange(remote: { name: string; url: string }) {
  if (remote) {
    form.value.remote_url = remote.url
    form.value.remote_name = remote.name
  }
}

function onBranchesChange(branches: string[]) {
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
    // 刷新状态
    setTimeout(() => {
      loadMirrors()
      if (logDialogVisible.value && currentLogMirrorId.value === mirror.id) {
        loadSyncLogs()
      }
    }, 2000)
  } catch (e: any) {
    ElMessage.error('触发失败: ' + e.message)
  } finally {
    // 保持 loading 状态一段时间
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

function formatTime(timeStr: string): string {
  if (!timeStr) return '-'
  const d = new Date(timeStr)
  return d.toLocaleString('zh-CN')
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

.mirror-card {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  overflow: hidden;
  background: var(--fill-color-blank);
  transition: all 0.2s ease;
}

.mirror-card:hover {
  box-shadow: var(--box-shadow-light);
  border-color: var(--primary-color-lighter);
}

.mirror-card.type-pull {
  border-left: 4px solid var(--el-color-primary);
}

.mirror-card.type-push {
  border-left: 4px solid var(--el-color-warning);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: var(--bg-color-page);
  border-bottom: 1px solid var(--border-color-lighter);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-title .el-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
}

.card-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.info-label {
  min-width: 80px;
  font-size: 12px;
  color: var(--text-color-secondary);
  flex-shrink: 0;
}

.info-value {
  flex: 1;
  font-size: 13px;
  color: var(--text-color-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-value.mono {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 12px;
}

.info-value.syncing {
  color: var(--el-color-primary);
  font-weight: 600;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.git-options {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.error-row {
  padding: 8px 12px;
  background: var(--el-color-danger-lighter);
  border-radius: var(--border-radius-base);
  margin: 0 -4px;
}

.error-row .info-label,
.error-row .info-value {
  color: var(--el-color-danger);
}

.card-footer {
  padding: 12px 16px 16px;
  border-top: 1px solid var(--border-color-lighter);
}

.card-footer .el-button-group {
  width: 100%;
}

.card-footer .el-button {
  flex: 1;
}

.form-tip {
  font-size: 12px;
  color: var(--text-color-secondary);
  margin-top: 4px;
}

.checkbox-label {
  font-size: 13px;
}

.warning-text {
  color: var(--el-color-warning);
}

.log-section {
  margin-top: 20px;
}

.log-section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-color-primary);
}

.log-content {
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-base);
  padding: 12px;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.error-text {
  color: var(--el-color-danger);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 12px;
}

:deep(.mirror-dialog .el-dialog__body),
:deep(.log-dialog .el-dialog__body),
:deep(.detail-dialog .el-dialog__body) {
  padding-top: 16px;
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

  .card-footer .el-button-group {
    flex-wrap: wrap;
  }

  .card-footer .el-button {
    flex: 1 1 calc(50% - 2px);
  }
}
</style>
