<template>
  <div class="ai-chat-panel">
    <div class="ai-chat-header">
      <div class="ai-mode-switch">
        <span :class="['ai-mode-btn', { active: aiMode === 'chat' }]" @click="aiMode = 'chat'">对话</span>
        <span :class="['ai-mode-btn', { active: aiMode === 'agent' }]" @click="aiMode = 'agent'">Agent</span>
      </div>
      <div class="ai-chat-header-actions">
        <el-button text size="small" @click="clearMessages">
          <el-icon><Delete /></el-icon>
        </el-button>
        <el-button text size="small" @click="$emit('close')">
          <el-icon :size="16"><Close /></el-icon>
        </el-button>
      </div>
    </div>
    <div v-if="aiMode === 'chat'" class="ai-quick-actions">
      <el-button
        v-for="action in quickActions"
        :key="action.key"
        size="small"
        :type="activeAction === action.key ? 'primary' : 'default'"
        @click="handleQuickAction(action)"
      >{{ action.label }}</el-button>
    </div>
    <div v-if="aiMode === 'agent' && !pendingAgentContent" class="ai-agent-hint">
      <el-icon :size="16" color="#6366F1"><MagicStick /></el-icon>
      <span>Agent 模式：描述你想要的修改，AI 将直接编辑文件</span>
    </div>
    <div v-if="pendingAgentContent" class="agent-diff-section">
      <div class="agent-diff-header">
        <span>AI 已修改文件</span>
        <div class="agent-diff-stats">
          <span class="added">+{{ agentDiffStats.added }} 行</span>
          <span class="removed">-{{ agentDiffStats.removed }} 行</span>
        </div>
      </div>
      <el-scrollbar class="agent-diff-scroll">
        <pre class="agent-diff-content">{{ agentDiffText }}</pre>
      </el-scrollbar>
      <div class="agent-diff-actions">
        <el-button size="small" type="primary" @click="onAcceptAgentChange">接受修改</el-button>
        <el-button size="small" @click="rejectAgentChange">拒绝</el-button>
      </div>
    </div>
    <div v-else ref="messagesRef" class="ai-messages">
      <div v-if="aiMessages.length === 0" class="ai-empty">
        <el-icon :size="28" color="#6366F1"><MagicStick /></el-icon>
        <p v-if="aiMode === 'agent'">输入修改指令，Agent 将直接编辑文件<br/>例如：补全缺失字段、修改版本号、优化构建脚本</p>
        <p v-else>选择快捷操作或输入问题<br/>AI 将协助你编辑 Spec 文件</p>
      </div>
      <div
        v-for="(msg, idx) in aiMessages"
        :key="idx"
        :class="['ai-message', `ai-message--${msg.role}`]"
      >
        <div class="ai-message-avatar">
          <el-icon v-if="msg.role === 'user'" :size="14"><User /></el-icon>
          <el-icon v-else :size="14"><MagicStick /></el-icon>
        </div>
        <div class="ai-message-body">
          <div class="ai-message-content" v-html="renderMarkdown(msg.content)"></div>
          <div v-if="msg.role === 'assistant' && msg.applyContent" class="ai-message-actions">
            <el-button size="small" type="primary" @click="$emit('apply-content', msg.applyContent)">
              应用到编辑器
            </el-button>
          </div>
        </div>
      </div>
      <div v-if="aiLoading" class="ai-message ai-message--assistant">
        <div class="ai-message-avatar">
          <el-icon :size="14"><MagicStick /></el-icon>
        </div>
        <div class="ai-message-body">
          <div class="ai-typing"><span></span><span></span><span></span></div>
        </div>
      </div>
    </div>
    <div class="ai-chat-input">
      <el-input
        v-model="aiInput"
        type="textarea"
        :rows="2"
        :disabled="aiLoading"
        :placeholder="aiMode === 'agent' ? '描述你想做的修改，如：补全 BuildRequires、修改版本号为 2.0...' : '输入问题或指令...'"
        resize="none"
        @keydown.enter.exact.prevent="sendAIMessage"
      />
      <el-button
        type="primary"
        :loading="aiLoading"
        :disabled="!aiInput.trim()"
        @click="sendAIMessage"
      >{{ aiMode === 'agent' ? '执行' : '发送' }}</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { MagicStick, Close, Delete, User } from '@element-plus/icons-vue'
import { useAIChat } from '@/composables/useAIChat'

const props = defineProps<{ content: string }>()
const emit = defineEmits<{
  'apply-content': [content: string]
  'close': []
}>()

const {
  aiMessages, aiInput, aiLoading, activeAction, aiMode,
  pendingAgentContent, messagesRef, agentDiffText, agentDiffStats,
  quickActions, renderMarkdown, sendAIMessage, handleQuickAction,
  clearMessages, acceptAgentChange, rejectAgentChange,
} = useAIChat(() => props.content)

function onAcceptAgentChange() {
  const result = acceptAgentChange()
  if (result) {
    emit('apply-content', result)
  }
}
</script>

<style scoped>
.ai-chat-panel { width: 360px; flex-shrink: 0; display: flex; flex-direction: column; border-right: 1px solid var(--border-color, #E2E8F0); background: var(--bg-color-page, #fff); overflow: hidden; min-height: 0; }
.ai-chat-header { height: 42px; padding: 0 12px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--border-color, #E2E8F0); background: var(--bg-color, #F8F9FC); }
.ai-mode-switch { display: flex; background: var(--bg-color-page, #fff); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; padding: 2px; gap: 2px; }
.ai-mode-btn { padding: 3px 12px; border-radius: 4px; font-size: 12px; cursor: pointer; color: var(--text-color-secondary, #64748B); transition: all 0.15s; user-select: none; }
.ai-mode-btn:hover { color: var(--primary-color, #6366F1); }
.ai-mode-btn.active { background: var(--primary-color, #6366F1); color: #fff; }
.ai-chat-header-actions { display: flex; gap: 4px; }
.ai-quick-actions { padding: 8px 12px; display: flex; flex-wrap: wrap; gap: 6px; border-bottom: 1px solid var(--border-color, #E2E8F0); }
.ai-quick-actions .el-button { font-size: 11px; border-radius: 12px; padding: 2px 10px; height: 24px; }
.ai-agent-hint { padding: 8px 12px; display: flex; align-items: center; gap: 6px; border-bottom: 1px solid var(--border-color, #E2E8F0); font-size: 11px; color: var(--text-color-secondary, #64748B); background: var(--accent-bg, #EEF2FF); }
.agent-diff-section { flex: 1; display: flex; flex-direction: column; min-height: 0; overflow: hidden; }
.agent-diff-header { padding: 8px 12px; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border-color, #E2E8F0); font-size: 13px; font-weight: 500; background: var(--bg-color, #F8F9FC); color: var(--text-color-primary, #1E293B); }
.agent-diff-stats { display: flex; gap: 12px; font-size: 12px; font-weight: 400; }
.agent-diff-scroll { flex: 1; min-height: 0; }
.agent-diff-content { padding: 12px; margin: 0; font-size: 12px; font-family: 'Consolas', monospace; line-height: 1.6; white-space: pre-wrap; word-break: break-all; color: var(--text-color-regular, #475569); }
.agent-diff-actions { padding: 8px 12px; display: flex; gap: 8px; border-top: 1px solid var(--border-color, #E2E8F0); background: var(--bg-color, #F8F9FC); }
.ai-messages { flex: 1; overflow-y: auto; padding: 12px; display: flex; flex-direction: column; gap: 12px; min-height: 0; }
.ai-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; padding: 30px 16px; text-align: center; color: var(--text-color-secondary, #64748B); font-size: 12px; line-height: 1.6; }
.ai-message { display: flex; gap: 8px; }
.ai-message--user { flex-direction: row-reverse; }
.ai-message-avatar { width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; background: var(--border-color, #E2E8F0); color: var(--text-color-secondary, #64748B); }
.ai-message--user .ai-message-avatar { background: var(--primary-color, #6366F1); color: #fff; }
.ai-message--assistant .ai-message-avatar { background: #7c3aed; color: #fff; }
.ai-message-body { max-width: 260px; }
.ai-message-content { font-size: 12px; line-height: 1.5; padding: 8px 10px; border-radius: 10px; word-break: break-word; }
.ai-message--user .ai-message-content { background: var(--primary-color, #6366F1); color: #fff; border-top-right-radius: 3px; }
.ai-message--assistant .ai-message-content { background: var(--bg-color, #F8F9FC); color: var(--text-color-regular, #475569); border: 1px solid var(--border-color-light, #F1F5F9); border-top-left-radius: 3px; }
.ai-message-content :deep(pre) { background: var(--bg-color, #F8F9FC); color: var(--text-color-regular, #475569); padding: 6px 10px; border-radius: 4px; overflow-x: auto; margin: 6px 0; font-size: 11px; border: 1px solid var(--border-color-light, #F1F5F9); }
.ai-message-content :deep(code) { font-family: 'Consolas', monospace; font-size: 11px; background: var(--bg-color, #F8F9FC); padding: 1px 4px; border-radius: 3px; }
.ai-message-content :deep(strong) { color: var(--primary-color, #6366F1); }
.ai-message-actions { margin-top: 6px; }
.ai-message-actions .el-button { font-size: 11px; height: 22px; padding: 0 8px; }
.ai-typing { display: flex; gap: 4px; padding: 10px; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color-light, #F1F5F9); border-radius: 10px; border-top-left-radius: 3px; }
.ai-typing span { width: 5px; height: 5px; border-radius: 50%; background: var(--text-color-placeholder, #94A3B8); animation: typing 1.4s infinite; }
.ai-typing span:nth-child(2) { animation-delay: .2s; }
.ai-typing span:nth-child(3) { animation-delay: .4s; }
@keyframes typing { 0%,60%,100%{opacity:.3;transform:translateY(0)} 30%{opacity:1;transform:translateY(-3px)} }
.ai-chat-input { padding: 8px 12px; display: flex; gap: 6px; align-items: flex-end; border-top: 1px solid var(--border-color, #E2E8F0); background: var(--bg-color, #F8F9FC); }
.ai-chat-input :deep(.el-textarea__inner) { background: var(--bg-color-page, #fff); border-color: var(--border-color, #E2E8F0); color: var(--text-color-primary, #1E293B); font-size: 12px; border-radius: 6px; }
.ai-chat-input :deep(.el-textarea__inner):focus { border-color: var(--primary-color, #6366F1); }
.ai-chat-input .el-button { height: 36px; flex-shrink: 0; }
</style>
