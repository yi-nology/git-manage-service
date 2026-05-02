<template>
  <div>
    <div class="content-header">
      <SectionTitle title="Change Requests" />
      <div class="content-actions">
        <ActionPill variant="primary" :icon="Plus" @click="showCreateCRDialog = true">创建 CR</ActionPill>
        <ActionPill variant="outline" :icon="Refresh" @click="handleSyncCRs" :disabled="crSyncing">{{ crSyncing ? '刷新中...' : '刷新' }}</ActionPill>
      </div>
    </div>

    <DataTable :columns="crColumns" :data="crs" :loading="crLoading" row-key="id">
      <template #cell-cr_number="{ row }">
        <span class="mono">{{ row.cr_number }}</span>
      </template>
      <template #cell-source_branch="{ row }">
        <span class="mono">{{ row.source_branch }}</span>
      </template>
      <template #cell-target_branch="{ row }">
        <span class="mono">{{ row.target_branch }}</span>
      </template>
      <template #cell-state="{ row }">
        <StatusBadge :variant="crStateVariant(row.state)" :text="crStateLabel(row.state)" :showDot="false" />
      </template>
      <template #empty>
        <EmptyState title="暂无 CR" />
      </template>
      <template #row-actions="{ row }">
        <ActionPill v-if="row.state === 'opened'" variant="primary" small @click="handleMergeCR(row)">合并</ActionPill>
        <ActionPill v-if="row.state === 'opened'" variant="danger" small @click="handleCloseCR(row)">关闭</ActionPill>
      </template>
    </DataTable>

    <el-dialog v-model="showCreateCRDialog" title="创建 CR / MR" width="520px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="标题"><el-input v-model="createCRForm.title" placeholder="CR 标题" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="createCRForm.description" type="textarea" :rows="3" placeholder="可选描述" /></el-form-item>
        <el-form-item label="源分支"><el-input v-model="createCRForm.source_branch" placeholder="feature-branch" /></el-form-item>
        <el-form-item label="目标分支"><el-input v-model="createCRForm.target_branch" placeholder="main" /></el-form-item>
        <el-form-item label="标签"><el-input v-model="createCRForm.labels" placeholder="逗号分隔，如: bugfix,feature" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateCRDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCR" :loading="createCRLoading">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { listRemoteCRs, createRemoteCR, mergeRemoteCR, closeRemoteCR } from '@/api/modules/cr'
import type { CRDTO } from '@/api/modules/cr'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const props = defineProps<{
  active: boolean
  providerId: number
  repoOwner: string
  repoName: string
}>()

const crLoading = ref(false)
const crSyncing = ref(false)
const crs = ref<CRDTO[]>([])
const showCreateCRDialog = ref(false)
const createCRForm = ref({ title: '', description: '', source_branch: '', target_branch: '', labels: '' })
const createCRLoading = ref(false)
const loaded = ref(false)

const crColumns: TableColumn[] = [
  { key: 'cr_number', label: '#', width: '60px' },
  { key: 'title', label: '标题' },
  { key: 'source_branch', label: '源分支', width: '100px' },
  { key: 'target_branch', label: '目标分支', width: '100px' },
  { key: 'state', label: '状态', width: '80px' },
]

function crStateLabel(s: string) {
  if (s === 'opened') return '开启'
  if (s === 'merged') return '已合并'
  if (s === 'closed') return '已关闭'
  return s
}

function crStateVariant(s: string): 'success' | 'info' | 'danger' | 'default' {
  if (s === 'opened') return 'success'
  if (s === 'merged') return 'info'
  if (s === 'closed') return 'danger'
  return 'default'
}

async function loadCRs() {
  crLoading.value = true
  try {
    const res = await listRemoteCRs({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName, page: 1, per_page: 100 })
    crs.value = res?.items || []
  } catch { crs.value = [] }
  finally { crLoading.value = false }
}

async function handleSyncCRs() {
  crSyncing.value = true
  try {
    const res = await listRemoteCRs({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName, page: 1, per_page: 100 })
    crs.value = res?.items || []
    ElMessage.success(`已刷新，共 ${crs.value.length} 个 CR`)
  } catch (e: any) { ElMessage.error('刷新失败: ' + (e?.message || '')) }
  finally { crSyncing.value = false }
}

async function handleMergeCR(cr: CRDTO) {
  try { await ElMessageBox.confirm(`确定合并 CR #${cr.cr_number}？`, '确认合并', { type: 'info' }) } catch { return }
  try {
    await mergeRemoteCR(props.providerId, props.repoOwner, props.repoName, cr.cr_number)
    ElMessage.success('合并成功')
    loadCRs()
  } catch (e: any) { ElMessage.error('合并失败: ' + (e?.message || '')) }
}

async function handleCloseCR(cr: CRDTO) {
  try { await ElMessageBox.confirm(`确定关闭 CR #${cr.cr_number}？`, '确认关闭', { type: 'warning' }) } catch { return }
  try {
    await closeRemoteCR(props.providerId, props.repoOwner, props.repoName, cr.cr_number)
    ElMessage.success('已关闭')
    loadCRs()
  } catch (e: any) { ElMessage.error('关闭失败: ' + (e?.message || '')) }
}

async function handleCreateCR() {
  const f = createCRForm.value
  if (!f.title || !f.source_branch || !f.target_branch) {
    ElMessage.warning('请填写标题、源分支和目标分支')
    return
  }
  createCRLoading.value = true
  try {
    await createRemoteCR({
      provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName,
      title: f.title, description: f.description,
      source_branch: f.source_branch, target_branch: f.target_branch,
      labels: f.labels ? f.labels.split(',').map(s => s.trim()).filter(Boolean) : undefined,
    })
    ElMessage.success('CR 创建成功')
    showCreateCRDialog.value = false
    createCRForm.value = { title: '', description: '', source_branch: '', target_branch: '', labels: '' }
    loadCRs()
  } catch (e: any) {
    ElMessage.error('创建失败: ' + (e?.message || ''))
  } finally {
    createCRLoading.value = false
  }
}

watch(() => props.active, (val) => {
  if (val && !loaded.value) {
    loadCRs()
    loaded.value = true
  }
}, { immediate: true })
</script>

<style scoped>
.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.content-actions {
  display: flex;
  gap: 8px;
}
.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}
</style>
