<template>
  <div>
    <div class="content-header">
      <SectionTitle title="远程分支" />
      <div class="content-actions">
        <ActionPill variant="primary" :icon="Plus" @click="showCreateBranchDialog = true">创建分支</ActionPill>
        <ActionPill variant="outline" :icon="Refresh" @click="loadRemoteBranches">刷新</ActionPill>
      </div>
    </div>

    <DataTable :columns="branchColumns" :data="remoteBranches" :loading="rbLoading" row-key="name">
      <template #cell-name="{ row }">
        <span class="mono"><el-icon :size="14" style="color:#10B981"><Share /></el-icon> {{ row.name }}</span>
      </template>
      <template #empty>
        <EmptyState title="暂无远程分支数据" />
      </template>
      <template #row-actions="{ row }">
        <ActionPill variant="primary" small @click="handleCheckoutRemote(row.name)">检出本地</ActionPill>
        <ActionPill variant="danger" small @click="handleDeleteRemoteBranch(row.name)">删除</ActionPill>
      </template>
    </DataTable>

    <el-dialog v-model="showCreateBranchDialog" title="创建远程分支" width="480px" destroy-on-close @open="loadDialogBranches">
      <el-form label-width="80px">
        <el-form-item label="分支名"><el-input v-model="createBranchForm.branch" placeholder="如: feature/new-api" /></el-form-item>
        <el-form-item label="基于">
          <el-select v-model="createBranchForm.ref" filterable :placeholder="defaultBranch || '选择基准分支'" style="width: 100%" :loading="dialogBranchesLoading">
            <el-option v-for="b in dialogBranches" :key="b.name" :label="b.name" :value="b.name" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateBranchDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateBranch" :loading="createBranchLoading">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Share } from '@element-plus/icons-vue'
import { listRemoteBranches, createRemoteBranch, deleteRemoteBranch } from '@/api/modules/provider'
import { createBranch } from '@/api/modules/branch'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const props = defineProps<{
  active: boolean
  providerId: number
  repoOwner: string
  repoName: string
  linkedRepoKey: string | null
  defaultBranch: string
}>()

const rbLoading = ref(false)
const remoteBranches = ref<{ name: string }[]>([])
const showCreateBranchDialog = ref(false)
const createBranchForm = ref({ branch: '', ref: '' })
const createBranchLoading = ref(false)
const loaded = ref(false)
const dialogBranches = ref<{ name: string }[]>([])
const dialogBranchesLoading = ref(false)

const branchColumns: TableColumn[] = [
  { key: 'name', label: '分支名' },
]

async function loadRemoteBranches() {
  rbLoading.value = true
  remoteBranches.value = []
  try {
    const res = await listRemoteBranches({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName })
    remoteBranches.value = (res || []) as any[]
  } catch { remoteBranches.value = [] }
  finally { rbLoading.value = false }
}

async function loadDialogBranches() {
  dialogBranchesLoading.value = true
  try {
    const res = await listRemoteBranches({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName })
    dialogBranches.value = (res || []) as any[]
  } catch { dialogBranches.value = [] }
  finally { dialogBranchesLoading.value = false }
}

async function handleCheckoutRemote(branchName: string) {
  if (!props.linkedRepoKey) {
    ElMessage.warning('请先克隆到本地再检出分支')
    return
  }
  try {
    await createBranch({ repo_key: props.linkedRepoKey!, name: branchName, base_ref: `origin/${branchName}` })
    ElMessage.success(`已检出分支 ${branchName}`)
  } catch (e: any) { ElMessage.error('检出失败: ' + (e?.message || '')) }
}

async function handleDeleteRemoteBranch(branchName: string) {
  try { await ElMessageBox.confirm(`确定删除远程分支 ${branchName}？此操作不可恢复！`, '确认删除', { type: 'warning' }) } catch { return }
  try {
    await deleteRemoteBranch({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName, branch: branchName })
    ElMessage.success(`已删除分支 ${branchName}`)
    loadRemoteBranches()
  } catch (e: any) { ElMessage.error('删除失败: ' + (e?.message || '')) }
}

async function handleCreateBranch() {
  const f = createBranchForm.value
  if (!f.branch) { ElMessage.warning('请输入分支名'); return }
  if (!f.ref) { ElMessage.warning('请输入基于的分支/标签'); return }
  createBranchLoading.value = true
  try {
    await createRemoteBranch({ provider_id: props.providerId, owner: props.repoOwner, repo: props.repoName, branch: f.branch, ref: f.ref })
    ElMessage.success(`分支 ${f.branch} 创建成功`)
    showCreateBranchDialog.value = false
    createBranchForm.value = { branch: '', ref: '' }
    loadRemoteBranches()
  } catch (e: any) { ElMessage.error('创建失败: ' + (e?.message || '')) }
  finally { createBranchLoading.value = false }
}

watch(() => props.active, (val) => {
  if (val && !loaded.value) {
    loadRemoteBranches()
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
