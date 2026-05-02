<template>
  <el-dialog :model-value="true" :title="`冲突解决 — ${filePath}`" width="90vw" top="3vh" destroy-on-close append-to-body @close="emit('close')">
    <div class="conflict-resolver">
      <div class="cr-toolbar">
        <div class="cr-toolbar-left">
          <el-button @click="keepOurs" plain size="small">保留 Ours</el-button>
          <el-button @click="keepTheirs" plain size="small">保留 Theirs</el-button>
          <el-button type="primary" @click="doAIResolve" :loading="aiLoading" size="small">
            <el-icon><MagicStick /></el-icon> AI 自动解决
          </el-button>
        </div>
      </div>

      <div class="cr-editor">
        <div class="cr-pane">
          <div class="cr-pane-header ours-header">Ours (HEAD)</div>
          <pre class="cr-pane-content">{{ conflictDetail?.oursContent || '加载中...' }}</pre>
        </div>
        <div class="cr-pane">
          <div class="cr-pane-header merged-header">
            合并结果 (可编辑)
            <el-tag v-if="aiResolved" type="primary" size="small" style="margin-left: 8px">
              AI 建议 · 置信度 {{ (aiResolved.confidence * 100).toFixed(0) }}%
            </el-tag>
          </div>
          <textarea class="cr-pane-content cr-editable" v-model="mergedContent" spellcheck="false" />
        </div>
      </div>

      <div v-if="aiResolved?.explanation" class="ai-explanation">
        <el-icon><InfoFilled /></el-icon>
        <span>AI 说明: {{ aiResolved.explanation }}</span>
      </div>
    </div>

    <template #footer>
      <div class="cr-footer">
        <el-button @click="emit('close')">取消</el-button>
        <el-button type="primary" @click="applyResolved" :disabled="!mergedContent">
          <el-icon><Check /></el-icon> 应用解决
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getConflictDetail, markConflictResolved, aiResolveConflict, type ConflictDetail, type AIResolvedFile } from '@/api/modules/workspace'
import { ElMessage } from 'element-plus'
import { MagicStick, Check, InfoFilled } from '@element-plus/icons-vue'

const props = defineProps<{
  repoKey: string
  filePath: string
}>()

const emit = defineEmits<{
  resolved: []
  close: []
}>()

const conflictDetail = ref<ConflictDetail | null>(null)
const mergedContent = ref('')
const aiLoading = ref(false)
const aiResolved = ref<AIResolvedFile | null>(null)

onMounted(async () => {
  try {
    conflictDetail.value = await getConflictDetail(props.repoKey, props.filePath)
    if (conflictDetail.value?.conflictMarker) {
      mergedContent.value = conflictDetail.value.conflictMarker
    }
  } catch (e: any) {
    ElMessage.error('加载冲突详情失败')
  }
})

function keepOurs() {
  mergedContent.value = conflictDetail.value?.oursContent || ''
}

function keepTheirs() {
  mergedContent.value = conflictDetail.value?.theirsContent || ''
}

async function doAIResolve() {
  if (!conflictDetail.value) return
  aiLoading.value = true
  try {
    const result = await aiResolveConflict(
      props.repoKey,
      props.filePath,
      conflictDetail.value.oursContent,
      conflictDetail.value.theirsContent,
      conflictDetail.value.baseContent,
    )
    if (result) {
      aiResolved.value = result
      mergedContent.value = result.resolvedContent
      ElMessage.success('AI 解决完成')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'AI 解决失败')
  } finally {
    aiLoading.value = false
  }
}

async function applyResolved() {
  try {
    await markConflictResolved(props.repoKey, props.filePath, mergedContent.value, true)
    ElMessage.success('冲突已解决')
    emit('resolved')
  } catch (e: any) {
    ElMessage.error(e?.message || '应用失败')
  }
}
</script>

<style scoped>
.conflict-resolver {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 65vh;
}

.cr-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.cr-toolbar-left {
  display: flex;
  gap: 8px;
}

.cr-editor {
  display: flex;
  flex: 1;
  gap: 1px;
  background: #333;
  border-radius: 8px;
  overflow: hidden;
}

.cr-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  min-width: 0;
}

.cr-pane-header {
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  text-align: center;
  border-bottom: 1px solid #333;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ours-header { color: #F87171; }
.merged-header { color: #10B981; }

.cr-pane-content {
  flex: 1;
  margin: 0;
  padding: 12px;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  color: #d4d4d4;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.cr-editable {
  background: transparent;
  border: none;
  outline: none;
  resize: none;
  width: 100%;
  height: 100%;
}

.ai-explanation {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #1A1A3C;
  border-radius: 6px;
  color: #C4B5FD;
  font-size: 12px;
}

.cr-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
