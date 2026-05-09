<template>
  <div class="ai-panel">
    <div class="ai-panel-header">
      <div class="ai-panel-title">
        <el-icon :size="16" color="#6366F1"><MagicStick /></el-icon>
        <span>{{ title || 'AI 助手' }}</span>
      </div>
      <div class="ai-panel-actions">
        <slot name="header-actions"></slot>
        <el-button text size="small" @click="clearMessages">
          <el-icon><Delete /></el-icon>
        </el-button>
        <el-button text size="small" @click="emit('update:visible', false); emit('close')">
          <el-icon :size="16"><Close /></el-icon>
        </el-button>
      </div>
    </div>
    
    <div v-if="quickActions.length" class="ai-quick-actions">
      <el-button
        v-for="action in quickActions"
        :key="action.key"
        size="small"
        :type="activeAction === action.key ? 'primary' : 'default'"
        @click="handleQuickAction(action)"
      >{{ action.label }}</el-button>
    </div>
    
    <div v-if="draftContent" class="ai-draft-section">
      <div class="ai-draft-header">
        <span>AI 已生成修改草案</span>
        <div v-if="draftStats" class="ai-draft-stats">
          <span v-if="draftStats.added !== undefined" class="added">+{{ draftStats.added }} 行</span>
          <span v-if="draftStats.removed !== undefined" class="removed">-{{ draftStats.removed }} 行</span>
        </div>
      </div>
      <div class="ai-draft-summary">{{ draftSummary }}</div>
      <el-scrollbar class="ai-draft-scroll">
        <pre class="ai-draft-content">{{ draftContent }}</pre>
      </el-scrollbar>
      <div class="ai-draft-actions">
        <slot name="draft-actions">
          <el-button size="small" type="primary" @click="onAcceptDraft">应用修改</el-button>
          <el-button size="small" @click="rejectDraft">拒绝</el-button>
        </slot>
      </div>
    </div>
    
    <div v-else :ref="setMessagesRef" class="ai-messages">
      <div v-if="messages.length === 0" class="ai-empty">
        <el-icon :size="28" color="#6366F1"><MagicStick /></el-icon>
        <p>{{ emptyHint || '选择快捷操作或输入问题开始' }}</p>
      </div>
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        :class="['ai-message', `ai-message--${msg.role}`]"
      >
        <div class="ai-message-avatar">
          <el-icon v-if="msg.role === 'user'" :size="14"><User /></el-icon>
          <el-icon v-else :size="14"><MagicStick /></el-icon>
        </div>
        <div class="ai-message-body">
          <div class="ai-message-content" v-html="renderMarkdown(msg.content)"></div>
          <div v-if="msg.references && msg.references.length" class="ai-message-refs">
            <div v-for="ref in msg.references" :key="ref.id" class="ai-ref-item">
              <el-icon :size="12"><Link /></el-icon>
              <span>{{ ref.label }}</span>
            </div>
          </div>
          <div class="ai-message-feedback" v-if="msg.role === 'assistant' && !msg.feedback">
            <el-button text size="small" @click="sendFeedback(msg, 'helpful')">有用 👍</el-button>
            <el-button text size="small" @click="sendFeedback(msg, 'not_helpful')">没用 👎</el-button>
          </div>
        </div>
      </div>
      <div v-if="loading" class="ai-message ai-message--assistant">
        <div class="ai-message-avatar">
          <el-icon :size="14"><MagicStick /></el-icon>
        </div>
        <div class="ai-message-body">
          <div class="ai-typing"><span></span><span></span><span></span></div>
        </div>
      </div>
    </div>
    
    <div class="ai-panel-input">
      <el-input
        v-model="input"
        type="textarea"
        :rows="2"
        :disabled="loading"
        :placeholder="inputPlaceholder || '输入指令或问题...'"
        resize="none"
        @keydown.enter.exact.prevent="sendMessage"
      />
      <el-button
        type="primary"
        :loading="loading"
        :disabled="!input.trim()"
        @click="sendMessage"
      >{{ sendLabel || '发送' }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, type VNodeRef } from 'vue'
import { ElMessage } from 'element-plus'
import { MagicStick, User, Delete, Close, Link } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { aiApi } from '@/api/modules/ai'

export interface AIRef {
  type: string
  id: string
  label: string
  filePath?: string
  url?: string
}

 export interface AIMessage {
   role: 'user' | 'assistant'
   content: string
   references?: AIRef[]
   applyContent?: string
   feedback?: string
   invocationId?: number
 }

export interface QuickAction {
  key: string
  label: string
  prompt: string
}

export interface DraftStats {
  added?: number
  removed?: number
}

 const props = withDefaults(defineProps<{
   title?: string
   visible?: boolean
   emptyHint?: string
   inputPlaceholder?: string
   sendLabel?: string
   quickActions?: QuickAction[]
   aiLoading?: boolean
 }>(), {
   quickActions: () => [],
   visible: true,
 })

 const emit = defineEmits<{
   'update:visible': [value: boolean]
   close: []
   send: [message: string]
   apply: [content: string]
   acceptDraft: []
   rejectDraft: []
  }>()

 const messages = ref<AIMessage[]>([])
 const input = ref('')
 const internalLoading = ref(false)
 const loading = computed(() => props.aiLoading || internalLoading.value)
 const activeAction = ref('')
 const draftContent = ref('')
 const draftSummary = ref('')
 const draftStats = ref<DraftStats | null>(null)
 const messagesRef = ref<HTMLElement | null>(null)

 const setDraft = (content: string, summary?: string, stats?: DraftStats) => {
   internalLoading.value = false
   draftContent.value = content
   draftSummary.value = summary || ''
   draftStats.value = stats || null
 }

const setMessagesRef: VNodeRef = (el) => {
  messagesRef.value = el as HTMLElement | null
}

const renderMarkdown = (content: string) => {
  return marked.parse(content)
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

const clearMessages = () => {
  messages.value = []
  draftContent.value = ''
  draftSummary.value = ''
  draftStats.value = null
}

const handleQuickAction = (action: QuickAction) => {
  activeAction.value = action.key
  input.value = action.prompt
  sendMessage()
}

 const sendMessage = () => {
   if (!input.value.trim() || internalLoading.value) return
   
   const userMessage = input.value.trim()
   messages.value.push({ role: 'user', content: userMessage })
   input.value = ''
   scrollToBottom()
   
   internalLoading.value = true
   emit('send', userMessage)
 }

const addResponse = (content: string, references?: AIRef[], applyContent?: string, invocationId?: number) => {
   internalLoading.value = false
   messages.value.push({
     role: 'assistant',
     content,
     references,
     applyContent,
     invocationId,
   })
   scrollToBottom()
 }

 const rejectDraft = () => {
   draftContent.value = ''
   draftSummary.value = ''
   draftStats.value = null
   emit('rejectDraft')
 }

 const onAcceptDraft = () => {
   emit('acceptDraft')
 }

const sendFeedback = async (msg: AIMessage, feedback: string) => {
  if (!msg.invocationId) {
    ElMessage.warning('该 AI 结果暂不支持反馈')
    return
  }
  if (msg.feedback) {
    return
  }
  try {
    await aiApi.submitFeedback({ invocationId: msg.invocationId, feedback })
    msg.feedback = feedback
    ElMessage.success('反馈已提交')
  } catch (e) {
    ElMessage.error('反馈提交失败，请重试')
  }
}

defineExpose({
  addResponse,
  setDraft,
  clearMessages,
  loading,
})
</script>

<style lang="scss" scoped>
.ai-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-color);
  border-left: 1px solid var(--border-color);
  
  &-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
  }
  
  &-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
  }
  
  &-actions {
    display: flex;
    gap: 4px;
  }
}

.ai-quick-actions {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--border-color);
}

.ai-draft-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
  
  &-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    font-weight: 500;
  }
  
  &-stats {
    display: flex;
    gap: 12px;
    font-size: 13px;
    
    .added { color: var(--el-color-success); }
    .removed { color: var(--el-color-danger); }
  }
  
  &-summary {
    font-size: 13px;
    color: var(--text-color-secondary);
    margin-bottom: 12px;
  }
  
  &-scroll {
    flex: 1;
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
  
  &-actions {
    display: flex;
    gap: 8px;
    margin-top: 12px;
  }
}

.ai-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ai-empty {
  text-align: center;
  padding: 48px 16px;
  color: var(--text-color-secondary);
  
  p {
    margin: 12px 0 0;
    font-size: 13px;
    line-height: 1.6;
  }
}

.ai-message {
  display: flex;
  gap: 12px;
  
  &--user {
    flex-direction: row-reverse;
    
    .ai-message-body {
      background: var(--el-color-primary-light-9);
      border-radius: 12px 12px 4px 12px;
    }
  }
  
  &--assistant {
    .ai-message-body {
      background: var(--bg-color-secondary);
      border-radius: 12px 12px 12px 4px;
    }
  }
  
  &-avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-color-tertiary);
    flex-shrink: 0;
    margin-top: 4px;
  }
  
  &-body {
    flex: 1;
    min-width: 0;
    padding: 10px 12px;
    font-size: 14px;
    line-height: 1.6;
  }
  
  &-content {
    :deep(p) { margin: 0 0 8px; &:last-child { margin: 0; } }
    :deep(code) { background: rgba(0,0,0,0.05); padding: 2px 6px; border-radius: 4px; font-size: 13px; }
    :deep(pre) { background: rgba(0,0,0,0.05); padding: 12px; border-radius: 6px; overflow-x: auto; margin: 8px 0; }
    :deep(pre code) { padding: 0; }
  }
  
  &-refs {
    margin-top: 8px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    font-size: 12px;
    color: var(--text-color-secondary);
  }
  
  &-feedback {
    margin-top: 8px;
    display: flex;
    gap: 8px;
  }
}

.ai-ref-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: var(--bg-color);
  border-radius: 4px;
}

.ai-typing {
  display: flex;
  gap: 4px;
  padding: 8px 0;
  
  span {
    width: 6px;
    height: 6px;
    background: var(--text-color-secondary);
    border-radius: 50%;
    animation: typing 1.4s infinite;
    
    &:nth-child(2) { animation-delay: 0.2s; }
    &:nth-child(3) { animation-delay: 0.4s; }
  }
}

@keyframes typing {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
  30% { transform: translateY(-4px); opacity: 1; }
}

.ai-panel-input {
  padding: 12px 16px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 8px;
  align-items: flex-end;
}
</style>
