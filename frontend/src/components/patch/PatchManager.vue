<template>
  <div class="patch-manager">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>Patch 管理</span>
            <el-tag v-if="seriesStats" type="info" size="small">
              {{ seriesStats.applied_count }}/{{ seriesStats.total_patches }} 已应用
            </el-tag>
          </div>
          <div class="header-actions">
            <el-button
              v-if="seriesStats && seriesStats.can_apply_next"
              type="success"
              size="small"
              @click="applyNextPatch"
            >
              <el-icon><ArrowRight /></el-icon> 应用下一个 ({{ getNextPatchName() }})
            </el-button>
            <el-button
              v-if="seriesStats && seriesStats.pending_count > 1"
              type="warning"
              size="small"
              @click="applyAllPending"
            >
              <el-icon><Finished /></el-icon> 批量应用 ({{ seriesStats.pending_count }}个)
            </el-button>
            <el-button type="primary" size="small" @click="openGenerateDialog">
              <el-icon><Plus /></el-icon> 生成 Patch
            </el-button>
            <el-button size="small" @click="loadPatches">
              <el-icon><Refresh /></el-icon> 刷新
            </el-button>
          </div>
        </div>
      </template>

      <div v-if="seriesStats && seriesStats.total_patches > 0" class="progress-section">
        <el-progress
          :percentage="getProgress()"
          :status="seriesStats.conflict_count > 0 ? 'exception' : 'success'"
        />
        <div class="progress-text">
          已应用 {{ seriesStats.applied_count }} / 共 {{ seriesStats.total_patches }} 个 patch
          <span v-if="seriesStats.conflict_count > 0" class="error-text">
            ({{ seriesStats.conflict_count }} 个冲突)
          </span>
        </div>
      </div>

      <el-table :data="patches" v-loading="loading" stripe border size="small">
        <el-table-column prop="sequence" label="序号" width="70" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ String(row.sequence).padStart(3, '0') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="文件名" min-width="200">
          <template #default="{ row }">
            <div class="patch-name">
              <el-icon v-if="row.is_applied" color="#67C23A"><CircleCheck /></el-icon>
              <el-icon v-else-if="row.can_apply" color="#E6A23C"><Clock /></el-icon>
              <el-icon v-else color="#F56C6C"><Warning /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_applied" type="success" size="small">已应用</el-tag>
            <el-tag v-else-if="row.can_apply" type="warning" size="small">待应用</el-tag>
            <el-tag v-else type="danger" size="small">冲突</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="100">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mod_time" label="修改时间" width="160" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewPatch(row)">查看</el-button>
            <el-button size="small" type="primary" @click="downloadPatch(row)">下载</el-button>
            <el-button
              size="small"
              type="success"
              @click="openApplyDialog(row)"
              :disabled="!row.can_apply && !row.is_applied"
            >
              {{ row.is_applied ? '已应用' : '应用' }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无 Patch 文件" :image-size="60">
            <el-button type="primary" size="small" @click="openGenerateDialog">创建第一个 Patch</el-button>
          </el-empty>
        </template>
      </el-table>
    </el-card>

    <PatchGenerateDialog
      v-model="showGenerateDialog"
      :repo-key="repoKey"
      :patches="patches"
      @generated="loadPatches"
    />

    <el-dialog v-model="showViewDialog" title="Patch 内容" width="800px" destroy-on-close>
      <el-input
        v-model="patchContent"
        type="textarea"
        :rows="20"
        readonly
        class="patch-content"
      />
      <template #footer>
        <el-button @click="showViewDialog = false">关闭</el-button>
        <el-button type="primary" @click="copyContent">复制内容</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showApplyDialog" title="应用 Patch" width="700px" destroy-on-close>
      <el-alert v-if="!patchStats" type="info" :closable="false" class="mb-4">
        应用 Patch 将修改工作区文件，请确保已提交或暂存当前更改。
      </el-alert>

      <el-alert v-else-if="!patchStats.can_apply" type="error" :closable="false" class="mb-4">
        <template #title>此 Patch 无法应用</template>
        <div class="error-detail">{{ patchStats.error }}</div>
      </el-alert>

      <el-alert v-else type="success" :closable="false" class="mb-4">
        <template #title>Patch 可以应用</template>
        <pre class="stat-output">{{ patchStats.stat }}</pre>
      </el-alert>

      <el-form :model="applyForm" label-width="100px">
        <el-form-item label="Patch 文件">
          <el-input :value="applyForm.patchName" readonly />
        </el-form-item>

        <el-form-item label="提交消息">
          <el-input
            v-model="applyForm.commit_message"
            type="textarea"
            :rows="3"
            placeholder="留空则不自动提交，仅应用到工作区"
          />
          <div class="hint">快捷选项：
            <el-button size="small" link @click="applyForm.commit_message = 'feat: apply patch ' + applyForm.patchName">feat: apply patch</el-button>
            <el-button size="small" link @click="applyForm.commit_message = 'fix: apply patch'">fix: apply patch</el-button>
          </div>
        </el-form-item>

        <el-form-item v-if="applyForm.commit_message">
          <el-checkbox v-model="applyForm.sign_off">添加 Signed-off-by</el-checkbox>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showApplyDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleApply"
          :loading="applying"
          :disabled="!!patchStats && !patchStats.can_apply"
        >
          应用
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, CircleCheck, Clock, Warning, ArrowRight, Finished } from '@element-plus/icons-vue'
import {
  listPatches,
  getPatchContent,
  getPatchDownloadUrl,
  applyPatch,
  checkPatch,
  deletePatch,
} from '@/api/modules/patch'
import type { PatchInfoDTO, PatchStatsDTO } from '@/types/patch'
import { useNotification } from '@/composables/useNotification'
import PatchGenerateDialog from './PatchGenerateDialog.vue'

const props = defineProps<{
  repoKey: string
}>()

const { showSuccess, showError } = useNotification()

const loading = ref(false)
const patches = ref<PatchInfoDTO[]>([])
const seriesStats = ref<any>(null)

const showGenerateDialog = ref(false)

const showViewDialog = ref(false)
const patchContent = ref('')

const showApplyDialog = ref(false)
const applyForm = ref({
  patchPath: '',
  patchName: '',
  commit_message: '',
  sign_off: false,
})
const applying = ref(false)
const patchStats = ref<PatchStatsDTO | null>(null)

onMounted(() => {
  loadPatches()
})

async function loadPatches() {
  loading.value = true
  try {
    const result = await listPatches(props.repoKey)
    patches.value = Array.isArray(result) ? result : []

    const applied = patches.value.filter(p => p.is_applied).length
    const pending = patches.value.filter(p => !p.is_applied && p.can_apply).length
    const conflict = patches.value.filter(p => !p.is_applied && !p.can_apply).length

    const nextIndex = patches.value.findIndex(p => !p.is_applied && p.can_apply)

    seriesStats.value = {
      total_patches: patches.value.length,
      applied_count: applied,
      pending_count: pending,
      conflict_count: conflict,
      can_apply_next: nextIndex >= 0,
      next_patch_index: nextIndex,
    }
  } catch (e: any) {
    showError('加载失败', e)
    patches.value = []
    seriesStats.value = null
  } finally {
    loading.value = false
  }
}

function openGenerateDialog() {
  showGenerateDialog.value = true
}

async function viewPatch(patch: PatchInfoDTO) {
  try {
    const result = await getPatchContent(patch.path)
    patchContent.value = result.content
    showViewDialog.value = true
  } catch (e: any) {
    showError('读取失败', e)
  }
}

function downloadPatch(patch: PatchInfoDTO) {
  const url = getPatchDownloadUrl(patch.path)
  window.open(url, '_blank')
}

async function openApplyDialog(patch: PatchInfoDTO) {
  applyForm.value = {
    patchPath: patch.path,
    patchName: patch.name,
    commit_message: '',
    sign_off: false,
  }
  patchStats.value = null
  showApplyDialog.value = true

  try {
    patchStats.value = await checkPatch(props.repoKey, patch.path)
  } catch (e: any) {
    console.error('Failed to check patch:', e)
  }
}

async function handleApply() {
  applying.value = true
  try {
    await applyPatch({
      repo_key: props.repoKey,
      patch_path: applyForm.value.patchPath,
      commit_message: applyForm.value.commit_message || undefined,
      sign_off: applyForm.value.sign_off,
    })
    showSuccess('Patch 已应用')
    showApplyDialog.value = false
  } catch (e: any) {
    showError('应用失败', e)
  } finally {
    applying.value = false
  }
}

async function handleDelete(patch: PatchInfoDTO) {
  try {
    await ElMessageBox.confirm(`确定要删除 "${patch.name}" 吗？`, '确认删除', {
      type: 'warning',
    })
    await deletePatch(props.repoKey, patch.path)
    showSuccess('已删除')
    loadPatches()
  } catch (e: any) {
    if (e !== 'cancel') {
      showError('删除失败', e)
    }
  }
}

function copyContent() {
  navigator.clipboard.writeText(patchContent.value)
  showSuccess('已复制到剪贴板')
}

function getProgress(): number {
  if (!seriesStats.value || seriesStats.value.total_patches === 0) return 0
  return Math.round((seriesStats.value.applied_count / seriesStats.value.total_patches) * 100)
}

function getNextPatchName(): string {
  if (!seriesStats.value || seriesStats.value.next_patch_index < 0) return ''
  const patch = patches.value[seriesStats.value.next_patch_index]
  return patch ? patch.name : ''
}

async function applyNextPatch() {
  if (!seriesStats.value || seriesStats.value.next_patch_index < 0) {
    ElMessage.warning('没有待应用的 patch')
    return
  }

  const patch = patches.value[seriesStats.value.next_patch_index]
  if (patch) {
    await openApplyDialog(patch)
  }
}

async function applyAllPending() {
  if (!seriesStats.value || seriesStats.value.pending_count === 0) {
    ElMessage.warning('没有待应用的 patch')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要批量应用 ${seriesStats.value.pending_count} 个待应用的 patch 吗？`,
      '批量应用 Patch',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      }
    )

    if (!patches.value || !Array.isArray(patches.value)) {
      ElMessage.error('Patch 列表数据异常')
      return
    }

    const pendingPatches = patches.value.filter(p => !p.is_applied && p.can_apply)
    for (const patch of pendingPatches) {
      try {
        await applyPatch({
          repo_key: props.repoKey,
          patch_path: patch.path,
          commit_message: `feat: apply patch ${patch.name}`,
        })
      } catch (e: any) {
        ElMessage.error(`应用 ${patch.name} 失败: ${e.message || e}`)
        break
      }
    }

    ElMessage.success(`已成功应用 ${pendingPatches.length} 个 patch`)
    loadPatches()
  } catch (e) {
    if (e !== 'cancel') {
      console.error('Batch apply failed:', e)
    }
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}
</script>

<style scoped>
.patch-manager {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.patch-name {
  display: flex;
  align-items: center;
  gap: 6px;
}

.patch-name .el-icon {
  flex-shrink: 0;
}

.progress-section {
  margin-bottom: var(--spacing-md);
  padding: 12px;
  background: var(--bg-color);
  border-radius: var(--border-radius-sm);
}

.progress-text {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
}

.error-text {
  color: #F56C6C;
  margin-left: 8px;
}

.patch-content :deep(textarea) {
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}

.mb-4 {
  margin-bottom: 16px;
}

.error-detail {
  margin-top: 8px;
  font-size: 12px;
  white-space: pre-wrap;
}

.stat-output {
  margin-top: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--bg-color);
  border-radius: var(--border-radius-sm);
  font-size: var(--font-size-xs);
  max-height: 200px;
  overflow: auto;
}

.hint {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-xs);
}
</style>
