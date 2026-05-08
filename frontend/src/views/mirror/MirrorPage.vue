<template>
  <div class="mirror-page">
    <PageHeader :title="'镜像同步 - ' + repoName" :show-back="true" :back-route="`/local-repos/${repoKey}`" />

    <div class="mirror-content">
      <div class="toolbar">
        <el-button type="primary" @click="showCreateDialog('pull')">创建 Pull Mirror</el-button>
        <el-button @click="showCreateDialog('push')">创建 Push Mirror</el-button>
        <el-button @click="batchTriggerSync" :disabled="selectedMirrors.length === 0" type="success">
          批量同步 ({{ selectedMirrors.length }})
        </el-button>
        <el-button @click="loadMirrors">刷新</el-button>
      </div>

      <el-table :data="mirrors" @selection-change="handleSelectionChange" v-loading="loading" stripe>
        <el-table-column type="selection" width="50" />
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.mirrorType === 'pull' ? 'primary' : 'warning'" size="small">
              {{ row.mirrorType === 'pull' ? 'Pull' : 'Push' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="远程URL" prop="remoteUrl" min-width="250" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="(statusMap[row.status]?.type || 'info') as any" size="small">
              {{ statusMap[row.status]?.label || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分支过滤" prop="branchFilter" width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.branchFilter || '全部分支' }}</template>
        </el-table-column>
        <el-table-column label="间隔" width="100">
          <template #default="{ row }">{{ row.cronExpr || `${row.syncInterval}s` }}</template>
        </el-table-column>
        <el-table-column label="上次同步" width="170">
          <template #default="{ row }">{{ row.lastSyncAt ? formatTime(row.lastSyncAt) : '-' }}</template>
        </el-table-column>
        <el-table-column label="Git选项" width="120">
          <template #default="{ row }">
            <span class="git-opts">
              <span v-if="row.gitForce" title="强制">F</span>
              <span v-if="row.gitPrune" title="清理">P</span>
              <span v-if="row.gitTags" title="标签">T</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="70">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" type="primary" @click="triggerSync(row)" :disabled="row.status === 'syncing'">
                同步
              </el-button>
              <el-button size="small" @click="showLogs(row)">日志</el-button>
              <el-button size="small" @click="editMirror(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="deleteMirror(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingMirror ? '编辑 Mirror' : `创建 ${createType === 'push' ? 'Push' : 'Pull'} Mirror`" width="600px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="远程URL" required>
          <el-input v-model="form.remoteUrl" placeholder="https://github.com/user/repo.git" />
        </el-form-item>
        <el-form-item label="Remote名称">
          <el-input v-model="form.remoteName" placeholder="origin" />
        </el-form-item>
        <el-form-item label="凭据">
          <el-select v-model="form.credentialId" clearable placeholder="选择凭据" style="width: 100%">
            <el-option v-for="c in credentials" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分支过滤">
          <el-input v-model="form.branchFilter" placeholder="main, develop, release/*" />
          <div class="form-hint">glob模式，逗号分隔。留空=全部分支</div>
        </el-form-item>
        <el-form-item label="同步间隔">
          <el-input-number v-model="form.syncInterval" :min="30" :step="30" />
          <span style="margin-left: 8px">秒</span>
        </el-form-item>
        <el-form-item label="Cron表达式">
          <el-input v-model="form.cronExpr" placeholder="0 */5 * * * (可选)" />
        </el-form-item>
        <el-form-item label="Push触发">
          <el-switch v-model="form.syncOnPush" />
          <span class="form-hint" style="margin-left: 8px">Push事件自动触发同步</span>
        </el-form-item>
        <el-form-item label="Git选项">
          <el-checkbox v-model="form.gitForce">强制推送</el-checkbox>
          <el-checkbox v-model="form.gitPrune">清理分支</el-checkbox>
          <el-checkbox v-model="form.gitTags">同步标签</el-checkbox>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logDialogVisible" title="同步日志" width="800px" destroy-on-close>
      <el-table :data="syncLogs" v-loading="logLoading" stripe max-height="500">
        <el-table-column label="触发类型" width="100">
          <template #default="{ row }">{{ triggerTypeMap[row.triggerType] || row.triggerType }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="90">
          <template #default="{ row }">{{ row.durationMs ? `${(row.durationMs / 1000).toFixed(1)}s` : '-' }}</template>
        </el-table-column>
        <el-table-column label="分支数" prop="branchesSynced" width="70" />
        <el-table-column label="提交数" prop="commitsPushed" width="70" />
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ row.startedAt ? formatTime(row.startedAt) : '-' }}</template>
        </el-table-column>
        <el-table-column label="错误" prop="errorMessage" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button size="small" @click="showLogDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="logDetailVisible" title="日志详情" width="700px">
      <div v-if="currentLog">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="状态">{{ currentLog.status }}</el-descriptions-item>
          <el-descriptions-item label="触发类型">{{ triggerTypeMap[currentLog.triggerType] }}</el-descriptions-item>
          <el-descriptions-item label="耗时">{{ currentLog.durationMs ? `${(currentLog.durationMs / 1000).toFixed(2)}s` : '-' }}</el-descriptions-item>
          <el-descriptions-item label="分支数">{{ currentLog.branchesSynced }}</el-descriptions-item>
          <el-descriptions-item label="提交数">{{ currentLog.commitsPushed }}</el-descriptions-item>
          <el-descriptions-item label="错误" :span="2">{{ currentLog.errorMessage || '无' }}</el-descriptions-item>
        </el-descriptions>
        <div v-if="currentLog.detailLog" style="margin-top: 16px">
          <h4>执行日志</h4>
          <el-input type="textarea" :model-value="currentLog.detailLog" :rows="15" readonly style="font-family: monospace" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '@/components/common/PageHeader.vue'
import * as mirrorApi from '@/api/modules/mirror'
import { listCredentials } from '@/api/modules/credential'
import type { MirrorDTO, MirrorSyncLogDTO, CreateMirrorReq, UpdateMirrorReq } from '@/types/mirror'
import { MIRROR_STATUS_MAP, TRIGGER_TYPE_MAP } from '@/types/mirror'

const route = useRoute()
const repoKey = route.params.repoKey as string
const repoName = ref(repoKey)

const loading = ref(false)
const mirrors = ref<MirrorDTO[]>([])
const selectedMirrors = ref<MirrorDTO[]>([])
const statusMap = MIRROR_STATUS_MAP
const triggerTypeMap = TRIGGER_TYPE_MAP

const dialogVisible = ref(false)
const editingMirror = ref<MirrorDTO | null>(null)
const createType = ref<'pull' | 'push'>('pull')
const credentials = ref<{ id: number; name: string }[]>([])

const form = ref<{
  remoteUrl: string
  remoteName: string
  credentialId: number | null
  branchFilter: string
  syncInterval: number
  cronExpr: string
  syncOnPush: boolean
  gitForce: boolean
  gitPrune: boolean
  gitTags: boolean
  enabled: boolean
}>({
  remoteUrl: '',
  remoteName: 'origin',
  credentialId: null,
  branchFilter: '',
  syncInterval: 600,
  cronExpr: '',
  syncOnPush: false,
  gitForce: false,
  gitPrune: true,
  gitTags: true,
  enabled: true,
})

const logDialogVisible = ref(false)
const logLoading = ref(false)
const syncLogs = ref<MirrorSyncLogDTO[]>([])
const currentLogMirrorId = ref<number>(0)

const logDetailVisible = ref(false)
const currentLog = ref<MirrorSyncLogDTO | null>(null)

onMounted(() => {
  loadMirrors()
  loadCredentials()
})

async function loadMirrors() {
  loading.value = true
  try {
    mirrors.value = await mirrorApi.getMirrors()
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

function handleSelectionChange(selection: MirrorDTO[]) {
  selectedMirrors.value = selection
}

function showCreateDialog(type: 'pull' | 'push' = 'pull') {
  editingMirror.value = null
  createType.value = type
  form.value = {
    remoteUrl: '',
    remoteName: 'origin',
    credentialId: null,
    branchFilter: '',
    syncInterval: 600,
    cronExpr: '',
    syncOnPush: false,
    gitForce: false,
    gitPrune: true,
    gitTags: true,
    enabled: true,
  }
  dialogVisible.value = true
}

function editMirror(mirror: MirrorDTO) {
  editingMirror.value = mirror
  createType.value = mirror.mirrorType
  form.value = {
    remoteUrl: mirror.remoteUrl,
    remoteName: mirror.remoteName,
    credentialId: mirror.credentialId,
    branchFilter: mirror.branchFilter,
    syncInterval: mirror.syncInterval,
    cronExpr: mirror.cronExpr,
    syncOnPush: mirror.syncOnPush,
    gitForce: mirror.gitForce,
    gitPrune: mirror.gitPrune,
    gitTags: mirror.gitTags,
    enabled: mirror.enabled,
  }
  dialogVisible.value = true
}

async function submitForm() {
  if (!form.value.remoteUrl) {
    ElMessage.warning('请输入远程URL')
    return
  }

  try {
    if (editingMirror.value) {
      const data: UpdateMirrorReq = { ...form.value }
      await mirrorApi.updateMirror(editingMirror.value.id, data)
      ElMessage.success('更新成功')
    } else {
      const data: CreateMirrorReq = {
        repoId: 0,
        mirrorType: createType.value,
        ...form.value,
      }
      await mirrorApi.createMirror(data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadMirrors()
  } catch (e: any) {
    ElMessage.error('操作失败: ' + e.message)
  }
}

async function deleteMirror(mirror: MirrorDTO) {
  try {
    await ElMessageBox.confirm('确认删除此 Mirror?', '删除确认', { type: 'warning' })
    await mirrorApi.deleteMirror(mirror.id)
    ElMessage.success('已删除')
    loadMirrors()
  } catch {}
}

async function triggerSync(mirror: MirrorDTO) {
  try {
    await mirrorApi.triggerMirrorSync(mirror.id)
    ElMessage.success('同步已触发')
    setTimeout(loadMirrors, 2000)
  } catch (e: any) {
    ElMessage.error('触发失败: ' + e.message)
  }
}

async function batchTriggerSync() {
  try {
    await mirrorApi.batchTriggerMirrorSync(selectedMirrors.value.map(m => m.id))
    ElMessage.success(`已触发 ${selectedMirrors.value.length} 个同步任务`)
    setTimeout(loadMirrors, 2000)
  } catch (e: any) {
    ElMessage.error('批量触发失败: ' + e.message)
  }
}

async function toggleEnabled(mirror: MirrorDTO) {
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
  }
}

async function showLogs(mirror: MirrorDTO) {
  currentLogMirrorId.value = mirror.id
  logDialogVisible.value = true
  logLoading.value = true
  try {
    syncLogs.value = await mirrorApi.getMirrorSyncLogs(mirror.id, 50)
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
.mirror-page {
  padding: 20px;
}

.mirror-content {
  margin-top: 16px;
}

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.git-opts span {
  display: inline-block;
  width: 20px;
  height: 20px;
  line-height: 20px;
  text-align: center;
  border-radius: 4px;
  background: #e6e6e6;
  margin-right: 4px;
  font-size: 12px;
  font-weight: bold;
}

.form-hint {
  color: #999;
  font-size: 12px;
}
</style>
