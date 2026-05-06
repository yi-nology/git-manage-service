<template>
  <el-dialog
    v-model="localVisible"
    :title="title || '应用 AI 生成的修改'"
    width="60%"
    :close-on-click-modal="false"
  >
    <div class="ai-apply-dialog">
      <div v-if="summary" class="ai-apply-summary">
        <el-alert :title="summary" type="info" :closable="false" />
      </div>
      
      <div class="ai-apply-diff">
        <div class="ai-apply-diff-header">
          <span>变更预览</span>
          <div class="ai-apply-diff-stats" v-if="stats">
            <span class="added">+{{ stats.added }} 行</span>
            <span class="removed">-{{ stats.removed }} 行</span>
          </div>
        </div>
        <el-scrollbar class="ai-apply-diff-scroll">
          <pre class="ai-apply-diff-content"><code>{{ diffContent || applyContent }}</code></pre>
        </el-scrollbar>
      </div>
      
      <div v-if="showCommitMessage" class="ai-apply-commit">
        <el-input
          v-model="localCommitMessage"
          type="textarea"
          :rows="2"
          placeholder="输入提交信息..."
          label="提交信息"
        />
      </div>
      
      <div v-if="riskLevel" class="ai-apply-risk">
        <el-alert :title="`风险等级: ${riskLevelText}`" :type="riskAlertType" :closable="false">
          <template v-if="riskNote" #default>{{ riskNote }}</template>
        </el-alert>
      </div>
    </div>
    
    <template #footer>
      <div class="ai-apply-footer">
        <el-button @click="onReject">拒绝修改</el-button>
        <el-button type="primary" @click="onApply">确认应用</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  title?: string
  summary?: string
  applyContent: string
  diffContent?: string
  stats?: { added?: number; removed?: number }
  showCommitMessage?: boolean
  commitMessage?: string
  riskLevel?: 'low' | 'medium' | 'high' | 'critical'
  riskNote?: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  apply: [options: { commitMessage?: string }]
  reject: []
}>()

const localVisible = ref(props.visible)
const localCommitMessage = ref(props.commitMessage || '')

watch(
  () => props.visible,
  (val) => {
    localVisible.value = val
  }
)

watch(localVisible, (val) => {
  emit('update:visible', val)
})

watch(
  () => props.commitMessage,
  (val) => {
    localCommitMessage.value = val || ''
  }
)

const riskLevelText = computed(() => {
  const map: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重',
  }
  return map[props.riskLevel || 'low']
})

const riskAlertType = computed(() => {
  const map: Record<string, 'success' | 'info' | 'warning' | 'error'> = {
    low: 'success',
    medium: 'info',
    high: 'warning',
    critical: 'error',
  }
  return map[props.riskLevel || 'low']
})

const onApply = () => {
  emit('apply', { commitMessage: localCommitMessage.value })
  localVisible.value = false
}

const onReject = () => {
  emit('reject')
  emit('update:visible', false)
}
</script>

<style lang="scss" scoped>
.ai-apply-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
  
  &-summary {
    flex-shrink: 0;
  }
  
  &-diff {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
    
    &-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 500;
    }
    
    &-stats {
      display: flex;
      gap: 12px;
      font-size: 13px;
      font-weight: normal;
      
      .added { color: var(--el-color-success); }
      .removed { color: var(--el-color-danger); }
    }
    
    &-scroll {
      height: 300px;
      border: 1px solid var(--border-color);
      border-radius: 6px;
      background: var(--bg-color-secondary);
    }
    
    &-content {
      padding: 12px;
      margin: 0;
      font-family: var(--font-mono);
      font-size: 12px;
      white-space: pre-wrap;
      word-break: break-all;
    }
  }
  
  &-commit {
    flex-shrink: 0;
  }
  
  &-risk {
    flex-shrink: 0;
  }
}

.ai-apply-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
