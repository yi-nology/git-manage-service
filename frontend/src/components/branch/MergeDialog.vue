<template>
  <el-dialog v-model="dialogVisible" title="合并分支" width="550px" destroy-on-close>
    <p>
      即将合并 <strong>{{ sourceBranch }}</strong> 到 <strong>{{ targetBranch }}</strong>
    </p>

    <div v-if="mergeChecking" class="mb-3">
      <el-icon class="is-loading"><Loading /></el-icon> 正在检测冲突...
    </div>

    <div v-if="mergeCheckResult && !mergeChecking">
      <el-alert
        v-if="mergeCheckResult.success"
        title="可以自动合并"
        type="success"
        :closable="false"
        show-icon
        class="mb-3"
      />
      <el-alert
        v-else
        title="检测到冲突"
        type="error"
        :closable="false"
        show-icon
        class="mb-3"
      >
        <p>无法自动合并。以下文件存在冲突：</p>
        <div class="conflict-list">
          <div v-for="c in mergeCheckResult.conflicts" :key="c" class="conflict-row">
            <span class="conflict-path">{{ c }}</span>
            <div class="conflict-actions">
              <el-button type="primary" size="small" @click="openConflictResolver(c)">
                <el-icon><MagicStick /></el-icon> 解决冲突
              </el-button>
            </div>
          </div>
        </div>
        <div class="conflict-batch-bar">
          <el-button type="primary" @click="batchAIResolve" :loading="batchResolving">
            <el-icon><MagicStick /></el-icon> AI 批量解决全部 ({{ mergeCheckResult.conflicts.length }} 文件)
          </el-button>
        </div>
      </el-alert>
    </div>

    <el-form v-if="mergeCheckResult?.success" :model="mergeForm" label-width="100px">
      <el-form-item label="合并信息">
        <el-input v-model="mergeForm.message" type="textarea" :rows="3" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="success" @click="handleMerge" :disabled="!mergeCheckResult?.success" :loading="merging">
        确认合并
      </el-button>
      <el-button v-if="mergeCheckResult && !mergeCheckResult.success" @click="recheckMerge" :loading="mergeChecking">
        重新检测
      </el-button>
    </template>
  </el-dialog>

  <ConflictResolver
    v-if="showConflictResolver"
    :repo-key="repoKey"
    :file-path="conflictFile"
    @resolved="onConflictResolved"
    @close="showConflictResolver = false"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, MagicStick } from '@element-plus/icons-vue'
import { checkMerge, mergeBranch } from '@/api/modules/branch'
import type { MergeCheckResult } from '@/types/branch'
import ConflictResolver from '@/components/repo/ConflictResolver.vue'
import { getConflictDetail, aiResolveConflict, markConflictResolved } from '@/api/modules/workspace'

const props = defineProps<{
  visible: boolean
  repoKey: string
  sourceBranch: string
  targetBranch: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'merged': []
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const mergeChecking = ref(false)
const mergeCheckResult = ref<MergeCheckResult | null>(null)
const merging = ref(false)
const mergeForm = ref({ message: '' })

const showConflictResolver = ref(false)
const conflictFile = ref('')
const batchResolving = ref(false)

watch(() => props.visible, (v) => {
  if (v && props.sourceBranch && props.targetBranch) {
    mergeCheckResult.value = null
    mergeForm.value.message = `Merge ${props.sourceBranch} into ${props.targetBranch}`
    doCheckMerge()
  }
})

async function doCheckMerge() {
  mergeChecking.value = true
  try {
    mergeCheckResult.value = await checkMerge(props.repoKey, props.sourceBranch, props.targetBranch)
  } finally {
    mergeChecking.value = false
  }
}

async function recheckMerge() {
  await doCheckMerge()
}

async function handleMerge() {
  merging.value = true
  try {
    await mergeBranch({
      repo_key: props.repoKey,
      source: props.sourceBranch,
      target: props.targetBranch,
      message: mergeForm.value.message,
    })
    ElMessage.success('合并成功')
    dialogVisible.value = false
    emit('merged')
  } finally {
    merging.value = false
  }
}

function openConflictResolver(file: string) {
  conflictFile.value = file
  showConflictResolver.value = true
}

function onConflictResolved() {
  showConflictResolver.value = false
  conflictFile.value = ''
  recheckMerge()
}

async function batchAIResolve() {
  if (!mergeCheckResult.value?.conflicts.length) return
  batchResolving.value = true
  let resolved = 0
  for (const file of mergeCheckResult.value.conflicts) {
    try {
      const detail = await getConflictDetail(props.repoKey, file)
      if (!detail) continue
      const result = await aiResolveConflict(props.repoKey, file, detail.ours_content, detail.theirs_content, detail.base_content)
      if (result) {
        await markConflictResolved(props.repoKey, file, result.resolved_content, true)
        resolved++
      }
    } catch {}
  }
  batchResolving.value = false
  ElMessage.success(`已解决 ${resolved}/${mergeCheckResult.value.conflicts.length} 个冲突`)
  await recheckMerge()
}
</script>

<style scoped>
.mb-3 {
  margin-bottom: 12px;
}

.conflict-list {
  margin-top: 8px;
}

.conflict-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
}

.conflict-row:last-child {
  border-bottom: none;
}

.conflict-path {
  font-family: 'Menlo', 'Monaco', monospace;
  font-size: 13px;
  color: var(--el-color-danger);
}

.conflict-actions {
  display: flex;
  gap: 8px;
}

.conflict-batch-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}
</style>
