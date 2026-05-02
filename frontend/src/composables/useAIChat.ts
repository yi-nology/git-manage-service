import { ref, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

export interface AIChatMessage {
  role: 'user' | 'assistant'
  content: string
  applyContent?: string
}

export function useAIChat(getContent: () => string) {
  const aiMessages = ref<AIChatMessage[]>([])
  const aiInput = ref('')
  const aiLoading = ref(false)
  const activeAction = ref('')
  const aiMode = ref<'chat' | 'agent'>('chat')
  const pendingAgentContent = ref('')
  const messagesRef = ref<HTMLDivElement>()

  const quickActions = [
    { key: 'check', label: '检查问题', prompt: '请检查这个 spec 文件是否存在语法错误、缺失字段或不符合 RPM 打包规范的问题，列出所有发现的问题并给出修改建议。' },
    { key: 'complete', label: '补全字段', prompt: '请分析这个 spec 文件，补全所有缺失的必需字段和可选但推荐的字段（如 URL、Source、BuildRequires 等），直接输出补全后的完整内容。' },
    { key: 'optimize', label: '优化建议', prompt: '请分析这个 spec 文件，给出优化建议，包括：构建流程优化、依赖精简、文件列表完善、宏使用规范等。' },
    { key: 'explain', label: '解释说明', prompt: '请逐段解释这个 spec 文件的内容和作用，对每个 section 和重要指令进行说明。' },
    { key: 'generate', label: '生成构建段', prompt: '请根据当前 spec 文件的 Name、Version、Source 等元信息，帮我生成完整的 %build、%install 和 %files 段内容。直接输出修改后的完整 spec 文件。' },
  ]

  const agentDiffText = computed(() => {
    if (!pendingAgentContent.value) return ''
    const orig = getContent().split('\n')
    const modified = pendingAgentContent.value.split('\n')
    const lines: string[] = []
    const maxLen = Math.max(orig.length, modified.length)
    for (let i = 0; i < maxLen; i++) {
      const o = orig[i] ?? ''
      const m = modified[i] ?? ''
      if (o !== m) {
        if (o && !m) lines.push(`- ${o}`)
        else if (!o && m) lines.push(`+ ${m}`)
        else { lines.push(`- ${o}`); lines.push(`+ ${m}`) }
      }
    }
    return lines.join('\n') || '没有变更'
  })

  const agentDiffStats = computed(() => {
    const text = agentDiffText.value
    return {
      added: (text.match(/^\+/gm) || []).length,
      removed: (text.match(/^-/gm) || []).length,
    }
  })

  function renderMarkdown(t: string): string {
    return t
      .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\n/g, '<br>')
  }

  function scrollAIToBottom() {
    nextTick(() => {
      if (messagesRef.value) messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    })
  }

  async function callAI(prompt: string, action?: string) {
    const content = getContent()
    if (!content) { ElMessage.warning('请先选择一个 Spec 文件'); return }
    aiMessages.value.push({ role: 'user', content: prompt })
    aiLoading.value = true
    activeAction.value = action || ''
    scrollAIToBottom()
    try {
      const effectiveAction = action || (aiMode.value === 'agent' ? 'agent' : 'chat')
      const history = aiMessages.value.slice(-10).map(m => ({ role: m.role, content: m.content }))
      const res = await request.post<unknown, { result: string; apply_content?: string }>('/spec/ai-assist', {
        content,
        prompt,
        action: effectiveAction,
        history,
      })
      if (effectiveAction === 'agent' && res?.apply_content) {
        pendingAgentContent.value = res.apply_content
        aiMessages.value.push({ role: 'assistant', content: '已生成修改，请查看并确认。' })
      } else {
        const msg: AIChatMessage = { role: 'assistant', content: res?.result || '无结果' }
        if (res?.apply_content) msg.applyContent = res.apply_content
        aiMessages.value.push(msg)
      }
    } catch (e: any) {
      aiMessages.value.push({ role: 'assistant', content: `请求失败: ${e?.message || '未知错误'}` })
    } finally {
      aiLoading.value = false
      activeAction.value = ''
      scrollAIToBottom()
    }
  }

  function sendAIMessage() {
    const t = aiInput.value.trim()
    if (!t || aiLoading.value) return
    aiInput.value = ''
    callAI(t)
  }

  function handleQuickAction(a: { key: string; prompt: string }) {
    callAI(a.prompt, a.key)
  }

  function clearMessages() {
    aiMessages.value = []
    pendingAgentContent.value = ''
  }

  function acceptAgentChange(): string | null {
    if (!pendingAgentContent.value) return null
    const result = pendingAgentContent.value
    pendingAgentContent.value = ''
    return result
  }

  function rejectAgentChange() {
    pendingAgentContent.value = ''
    ElMessage.info('已拒绝修改')
  }

  return {
    aiMessages,
    aiInput,
    aiLoading,
    activeAction,
    aiMode,
    pendingAgentContent,
    messagesRef,
    agentDiffText,
    agentDiffStats,
    quickActions,
    renderMarkdown,
    scrollAIToBottom,
    callAI,
    sendAIMessage,
    handleQuickAction,
    clearMessages,
    acceptAgentChange,
    rejectAgentChange,
  }
}
