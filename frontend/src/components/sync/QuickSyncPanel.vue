<template>
  <div v-if="modelValue" class="quick-sync-panel">
    <div class="qp-title">
      <span class="qp-icon">⚡</span>
      <span class="qp-title-text">快速同步</span>
      <button class="qp-close" @click="modelValue = false"><el-icon><Close /></el-icon></button>
    </div>
    <div class="qp-body">
      <div class="qp-field">
        <span class="qp-label">源</span>
        <div class="qp-row">
          <el-select v-model="quickForm.sourceRemote" style="width: 120px">
            <el-option label="Local" value="local" />
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
          <el-input v-model="quickForm.sourceBranch" placeholder="分支" style="width: 120px" />
        </div>
      </div>
      <div class="qp-arrow">
        <el-icon :size="20" color="#64748B"><Right /></el-icon>
      </div>
      <div class="qp-field">
        <span class="qp-label">目标</span>
        <div class="qp-row">
          <el-select v-model="quickForm.targetRemote" style="width: 120px">
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
          <el-input v-model="quickForm.targetBranch" placeholder="分支" style="width: 120px" />
        </div>
      </div>
      <div class="qp-options">
        <el-checkbox v-model="quickForm.gitTags">--tags</el-checkbox>
        <el-checkbox v-model="quickForm.gitForce">--force</el-checkbox>
        <el-checkbox v-model="quickForm.gitPrune">--prune</el-checkbox>
      </div>
      <div class="qp-actions">
        <ActionPill variant="outline" @click="handlePreview" :disabled="previewing">预览</ActionPill>
        <ActionPill variant="primary" @click="handleQuickSync" :disabled="syncing">执行</ActionPill>
      </div>
    </div>
    <div v-if="previewResult" class="qp-preview">
      <el-alert :title="previewResult.command" type="info" :closable="false">
        <div v-if="previewResult.commits_to_push">
          <strong>Commits:</strong> {{ Array.isArray(previewResult.commits_to_push) ? previewResult.commits_to_push.length : previewResult.commits_to_push }}
        </div>
        <div v-if="previewResult.tags_to_push">
          <strong>Tags:</strong> {{ Array.isArray(previewResult.tags_to_push) ? previewResult.tags_to_push.length : previewResult.tags_to_push }}
        </div>
        <div v-if="previewResult.warning" style="color: #F59E0B">{{ previewResult.warning }}</div>
      </el-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Close, Right } from '@element-plus/icons-vue'
import {
  createSyncTask,
  runSyncTask,
  previewSync,
} from '@/api/modules/sync'
import type { PreviewSyncResponse } from '@/types/sync'
import ActionPill from '@/components/common/ActionPill.vue'

const props = defineProps<{
  repoKey: string
  remoteNames: string[]
}>()

const modelValue = defineModel<boolean>({ required: true })

const syncing = ref(false)
const previewing = ref(false)
const previewResult = ref<PreviewSyncResponse | null>(null)
const quickForm = ref({
  sourceRemote: 'local',
  sourceBranch: 'main',
  targetRemote: '',
  targetBranch: 'main',
  gitTags: false,
  gitForce: false,
  gitPrune: false,
})

watch(() => props.remoteNames, (names) => {
  if (names.length > 0 && !quickForm.value.targetRemote) {
    quickForm.value.targetRemote = names[0]!
  }
}, { immediate: true })

watch(modelValue, (val) => {
  if (!val) {
    previewResult.value = null
  }
})

async function handlePreview() {
  previewing.value = true
  try {
    previewResult.value = await previewSync({
      repo_key: props.repoKey,
      source_remote: quickForm.value.sourceRemote,
      source_branch: quickForm.value.sourceBranch,
      target_remote: quickForm.value.targetRemote,
      target_branch: quickForm.value.targetBranch,
      git_tags: quickForm.value.gitTags,
      git_force: quickForm.value.gitForce,
      git_prune: quickForm.value.gitPrune,
    })
  } catch (e: any) {
    ElMessage.error(e.message || '预览失败')
  } finally {
    previewing.value = false
  }
}

async function handleQuickSync() {
  if (quickForm.value.gitForce) {
    try {
      await ElMessageBox.confirm('--force 会覆盖远端提交，确定继续？', '危险操作', { type: 'warning' })
    } catch {
      return
    }
  }

  syncing.value = true
  try {
    const result = await createSyncTask({
      source_repo_key: props.repoKey,
      target_repo_key: props.repoKey,
      source_remote: quickForm.value.sourceRemote,
      source_branch: quickForm.value.sourceBranch,
      target_remote: quickForm.value.targetRemote,
      target_branch: quickForm.value.targetBranch,
      git_tags: quickForm.value.gitTags,
      git_force: quickForm.value.gitForce,
      git_prune: quickForm.value.gitPrune,
      enabled: false,
    }) as any
    if (result?.task_key) {
      await runSyncTask(result.task_key)
      ElMessage.success('同步已触发')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '同步失败')
  } finally {
    syncing.value = false
  }
}
</script>

<style scoped>
.quick-sync-panel {
  background: var(--bg-color-page);
  border: 1px solid var(--primary-color);
  border-radius: var(--border-radius-lg);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.qp-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.qp-icon {
  font-size: 16px;
}

.qp-title-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
  flex: 1;
}

.qp-close {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-color-secondary);
  display: flex;
  align-items: center;
}

.qp-body {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
}

.qp-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.qp-label {
  font-size: 11px;
  color: var(--text-color-secondary);
}

.qp-row {
  display: flex;
  gap: 8px;
}

.qp-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
}

.qp-options {
  display: flex;
  gap: 12px;
  align-items: center;
}

.qp-actions {
  display: flex;
  gap: 8px;
}

.qp-preview {
  margin-top: 0;
}

@media (max-width: 768px) {
  .qp-body {
    flex-direction: column;
    align-items: stretch;
  }

  .qp-arrow {
    transform: rotate(90deg);
    justify-content: center;
  }
}
</style>
