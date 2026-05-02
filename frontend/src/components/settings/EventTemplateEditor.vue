<template>
  <template v-if="enabledEvents.length > 0">
    <el-form-item>
      <el-tabs v-model="activeEventTab" type="card" class="event-template-tabs">
        <el-tab-pane
          v-for="event in enabledEvents"
          :key="event"
          :name="event"
          :label="triggerEventLabels[event]"
        />
      </el-tabs>
    </el-form-item>

    <el-form-item>
      <div class="variable-panel">
        <div v-for="category in variableCategories" :key="category.label" class="var-group">
          <div class="var-group-header">
            <el-tag :type="category.type" size="small" effect="dark">{{ category.label }}</el-tag>
          </div>
          <div class="var-buttons">
            <el-tooltip
              v-for="varName in category.vars"
              :key="varName"
              :content="`${formatVar(varName)}\n示例: ${getVarByName(varName)?.example || ''}`"
              placement="top"
              :show-after="300"
              raw-content
            >
              <el-button
                size="small"
                :type="category.type || 'primary'"
                plain
                class="var-btn"
                @click="insertVariable(varName)"
              >{{ getVarByName(varName)?.description || varName }}</el-button>
            </el-tooltip>
          </div>
        </div>
        <div class="active-editor-hint">
          <el-icon :size="14"><InfoFilled /></el-icon>
          <span>点击变量将插入到
            <el-tag :type="activeEditor === 'title' ? 'primary' : 'success'" size="small" effect="plain">
              {{ activeEditor === 'title' ? '标题模板' : '内容模板' }}
            </el-tag>
            <el-tag type="info" size="small" effect="plain" style="margin-left:4px">{{ triggerEventLabels[activeEventTab] }}</el-tag>
          </span>
        </div>
      </div>
    </el-form-item>

    <el-form-item label="标题模板">
      <el-input
        ref="titleInputRef"
        v-model="currentTitleTemplate"
        :placeholder="`留空使用「${triggerEventLabels[activeEventTab]}」的内置默认模板`"
        clearable
        class="template-input"
        :class="{ 'editor-active': activeEditor === 'title' }"
        @focus="handleEditorFocus('title')"
      />
    </el-form-item>

    <el-form-item label="内容模板">
      <el-input
        ref="contentInputRef"
        v-model="currentContentTemplate"
        type="textarea"
        :rows="6"
        :placeholder="`留空使用「${triggerEventLabels[activeEventTab]}」的内置默认模板`"
        clearable
        class="template-input"
        :class="{ 'editor-active': activeEditor === 'content' }"
        @focus="handleEditorFocus('content')"
      />
    </el-form-item>

    <el-form-item>
      <el-collapse v-model="previewCollapse" class="preview-collapse">
        <el-collapse-item title="实时预览" name="preview">
          <div class="template-preview">
            <div class="preview-section">
              <div class="preview-label">标题</div>
              <div class="preview-title">{{ previewTitle }}</div>
            </div>
            <el-divider style="margin: 12px 0" />
            <div class="preview-section">
              <div class="preview-label">内容</div>
              <pre class="preview-content">{{ previewContent }}</pre>
            </div>
            <div class="preview-hint">
              <el-icon :size="12"><InfoFilled /></el-icon>
              预览使用示例数据渲染，留空字段将使用内置默认模板
            </div>
          </div>
        </el-collapse-item>
      </el-collapse>
    </el-form-item>
  </template>

  <template v-else>
    <el-form-item>
      <div class="empty-template-hint">
        <el-icon :size="32" color="var(--el-text-color-placeholder)"><InfoFilled /></el-icon>
        <p>请先在上方选择触发时机，再配置对应的消息模板</p>
      </div>
    </el-form-item>
  </template>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, watch, onMounted } from 'vue'
import { ElMessage, type InputInstance } from 'element-plus'
import { InfoFilled } from '@element-plus/icons-vue'

const props = defineProps<{
  enabledEvents: string[]
  triggerEventLabels: Record<string, string>
  initialTemplatesJson?: string
}>()

const eventTemplates = reactive<Record<string, { title_template: string; content_template: string }>>({})
const activeEventTab = ref('')
const activeEditor = ref<'title' | 'content'>('content')
const titleInputRef = ref<InputInstance | null>(null)
const contentInputRef = ref<InputInstance | null>(null)
const previewCollapse = ref(['preview'])

function ensureEventTemplate(event: string) {
  if (!eventTemplates[event]) {
    eventTemplates[event] = { title_template: '', content_template: '' }
  }
}

watch(() => props.enabledEvents, (events) => {
  for (const event of events) {
    ensureEventTemplate(event)
  }
  if (!events.includes(activeEventTab.value) && events.length > 0) {
    activeEventTab.value = events[0]!
  }
  if (events.length === 0) {
    activeEventTab.value = ''
  }
}, { immediate: true })

const currentTitleTemplate = computed({
  get: () => eventTemplates[activeEventTab.value]?.title_template || '',
  set: (val: string) => {
    ensureEventTemplate(activeEventTab.value)
    eventTemplates[activeEventTab.value]!.title_template = val
  }
})

const currentContentTemplate = computed({
  get: () => eventTemplates[activeEventTab.value]?.content_template || '',
  set: (val: string) => {
    ensureEventTemplate(activeEventTab.value)
    eventTemplates[activeEventTab.value]!.content_template = val
  }
})

const templateVariables = [
  { name: 'TaskKey', description: '任务标识', example: 'my-sync-task', events: '全部' },
  { name: 'Status', description: '状态码', example: 'success', events: '全部' },
  { name: 'StatusText', description: '状态文字', example: '成功', events: '全部' },
  { name: 'EventType', description: '事件类型', example: 'sync_success', events: '全部' },
  { name: 'EventLabel', description: '事件名称', example: '同步成功', events: '全部' },
  { name: 'Timestamp', description: '时间', example: '2026-02-16 10:30:00', events: '全部' },
  { name: 'RepoKey', description: '仓库标识', example: 'my-repo', events: '全部' },
  { name: 'SourceRemote', description: '源远程仓库', example: 'origin', events: '同步事件' },
  { name: 'SourceBranch', description: '源分支', example: 'main', events: '同步事件' },
  { name: 'TargetRemote', description: '目标远程仓库', example: 'backup', events: '同步事件' },
  { name: 'TargetBranch', description: '目标分支', example: 'main', events: '同步事件' },
  { name: 'CommitRange', description: '提交范围', example: 'abc..def', events: '同步成功' },
  { name: 'Duration', description: '执行耗时', example: '3.2s', events: '同步/备份' },
  { name: 'SyncMode', description: '同步模式', example: 'all-branch', events: '同步事件' },
  { name: 'BranchCount', description: '总分支数', example: '10', events: '同步事件' },
  { name: 'SuccessCount', description: '成功数', example: '8', events: '同步事件' },
  { name: 'FailedCount', description: '失败数', example: '2', events: '同步事件' },
  { name: 'ErrorMessage', description: '错误信息', example: 'push failed', events: '失败/错误/冲突' },
  { name: 'CronExpression', description: 'Cron表达式', example: '0 2 * * *', events: '定时任务' },
  { name: 'WebhookSource', description: 'Webhook来源', example: 'github', events: 'Webhook事件' },
  { name: 'BackupPath', description: '备份路径', example: '/backups/repo.tar.gz', events: '备份事件' },
]

const variableCategories = [
  {
    label: '通用',
    type: '' as const,
    vars: ['TaskKey', 'Status', 'StatusText', 'EventType', 'EventLabel', 'Timestamp', 'RepoKey']
  },
  {
    label: '同步事件',
    type: 'success' as const,
    vars: ['SourceRemote', 'SourceBranch', 'TargetRemote', 'TargetBranch', 'CommitRange', 'Duration', 'SyncMode', 'BranchCount', 'SuccessCount', 'FailedCount']
  },
  {
    label: '错误信息',
    type: 'danger' as const,
    vars: ['ErrorMessage']
  },
  {
    label: '特殊变量',
    type: 'warning' as const,
    vars: ['CronExpression', 'WebhookSource', 'BackupPath']
  }
]

const basePreviewData: Record<string, string> = {
  TaskKey: 'my-sync-task', RepoKey: 'my-repo',
  SourceRemote: 'origin', SourceBranch: 'main',
  TargetRemote: 'backup', TargetBranch: 'main',
  CommitRange: 'abc123..def456', Duration: '3.2s',
  SyncMode: 'single-branch', BranchCount: '10',
  SuccessCount: '8', FailedCount: '2',
  ErrorMessage: 'push failed: connection refused',
  CronExpression: '0 2 * * *', WebhookSource: 'github',
  BackupPath: '/backups/repo.tar.gz', Timestamp: '2026-02-20 10:30:00'
}

const defaultEventTitleTemplates: Record<string, string> = {
  sync_success: '[成功] 同步任务 {{.TaskKey}}',
  sync_failure: '[失败] 同步任务 {{.TaskKey}}',
  sync_conflict: '[冲突] 同步任务 {{.TaskKey}}',
  webhook_received: '[Webhook] 收到请求: {{.TaskKey}}',
  webhook_error: '[Webhook错误] {{.TaskKey}}',
  cron_triggered: '[定时] 任务触发: {{.TaskKey}}',
  backup_success: '[备份成功] {{.RepoKey}}',
  backup_failure: '[备份失败] {{.RepoKey}}'
}

const defaultEventContentTemplates: Record<string, string> = {
  sync_success: '任务: {{.TaskKey}}\n状态: {{.StatusText}}\n源: {{.SourceRemote}}/{{.SourceBranch}}\n目标: {{.TargetRemote}}/{{.TargetBranch}}\n时间: {{.Timestamp}}',
  sync_failure: '任务: {{.TaskKey}}\n状态: {{.StatusText}}\n源: {{.SourceRemote}}/{{.SourceBranch}}\n目标: {{.TargetRemote}}/{{.TargetBranch}}\n错误: {{.ErrorMessage}}\n时间: {{.Timestamp}}',
  sync_conflict: '任务: {{.TaskKey}}\n状态: 同步冲突\n源: {{.SourceRemote}}/{{.SourceBranch}}\n目标: {{.TargetRemote}}/{{.TargetBranch}}\n时间: {{.Timestamp}}',
  webhook_received: '任务: {{.TaskKey}}\n状态: Webhook 请求已接收\n来源: {{.WebhookSource}}\n时间: {{.Timestamp}}',
  webhook_error: '任务: {{.TaskKey}}\n状态: Webhook 处理失败\n错误: {{.ErrorMessage}}\n时间: {{.Timestamp}}',
  cron_triggered: '任务: {{.TaskKey}}\n状态: 定时任务已触发\nCron: {{.CronExpression}}\n时间: {{.Timestamp}}',
  backup_success: '仓库: {{.RepoKey}}\n状态: 备份成功\n备份路径: {{.BackupPath}}\n时间: {{.Timestamp}}',
  backup_failure: '仓库: {{.RepoKey}}\n状态: 备份失败\n错误: {{.ErrorMessage}}\n时间: {{.Timestamp}}'
}

function getPreviewData(): Record<string, string> {
  const event = activeEventTab.value
  const isSuccess = event.includes('success')
  const isFailure = event.includes('failure') || event.includes('error') || event.includes('conflict')
  return {
    ...basePreviewData,
    EventType: event,
    EventLabel: props.triggerEventLabels[event] || event,
    Status: isSuccess ? 'success' : isFailure ? 'failure' : 'info',
    StatusText: isSuccess ? '成功' : isFailure ? '失败' : '通知',
    ErrorMessage: isFailure ? 'push failed: connection refused' : ''
  }
}

function formatVar(name: string): string {
  return '{{.' + name + '}}'
}

function getVarByName(name: string) {
  return templateVariables.find(v => v.name === name)
}

function handleEditorFocus(editorType: 'title' | 'content') {
  activeEditor.value = editorType
}

function insertVariable(varName: string) {
  const varText = `{{.${varName}}}`
  const targetRef = activeEditor.value === 'title' ? titleInputRef.value : contentInputRef.value

  if (!targetRef) {
    navigator.clipboard.writeText(varText).then(() => {
      ElMessage.success(`已复制 ${varText}`)
    }).catch(() => {
      ElMessage.info(`请手动复制: ${varText}`)
    })
    return
  }

  const inputEl = (targetRef as any).textarea || (targetRef as any).input
  if (!inputEl) return

  const startPos = inputEl.selectionStart ?? 0
  const endPos = inputEl.selectionEnd ?? 0
  const currentValue = activeEditor.value === 'title' ? currentTitleTemplate.value : currentContentTemplate.value

  const newValue = currentValue.slice(0, startPos) + varText + currentValue.slice(endPos)

  if (activeEditor.value === 'title') {
    currentTitleTemplate.value = newValue
  } else {
    currentContentTemplate.value = newValue
  }

  nextTick(() => {
    const newCursorPos = startPos + varText.length
    inputEl.focus()
    inputEl.setSelectionRange(newCursorPos, newCursorPos)
  })

  ElMessage.success({ message: `已插入 ${varText}`, duration: 1000 })
}

function renderTemplate(tmplStr: string): string {
  if (!tmplStr) return ''
  const data = getPreviewData()
  let result = tmplStr
  for (const [key, value] of Object.entries(data)) {
    result = result.replace(new RegExp(`\\{\\{\\.${key}\\}\\}`, 'g'), value)
  }
  return result
}

const previewTitle = computed(() => {
  const event = activeEventTab.value
  const tmpl = currentTitleTemplate.value || defaultEventTitleTemplates[event] || '[通知] {{.TaskKey}}'
  return renderTemplate(tmpl)
})

const previewContent = computed(() => {
  const event = activeEventTab.value
  const tmpl = currentContentTemplate.value || defaultEventContentTemplates[event] || '任务: {{.TaskKey}}\n状态: {{.StatusText}}\n时间: {{.Timestamp}}'
  return renderTemplate(tmpl)
})

function loadTemplates(json: string) {
  Object.keys(eventTemplates).forEach(k => delete eventTemplates[k])
  activeEventTab.value = ''
  if (json) {
    try {
      const templates: Array<{ event_type: string; title_template: string; content_template: string }> = JSON.parse(json)
      for (const t of templates) {
        eventTemplates[t.event_type] = {
          title_template: t.title_template || '',
          content_template: t.content_template || ''
        }
      }
    } catch { /* ignore */ }
  }
}

function clearTemplates() {
  Object.keys(eventTemplates).forEach(k => delete eventTemplates[k])
  activeEventTab.value = ''
}

function buildJson(): string {
  const list: Array<{ event_type: string; title_template: string; content_template: string }> = []
  for (const event of props.enabledEvents) {
    const tmpl = eventTemplates[event]
    if (tmpl && (tmpl.title_template || tmpl.content_template)) {
      list.push({
        event_type: event,
        title_template: tmpl.title_template,
        content_template: tmpl.content_template
      })
    }
  }
  return list.length > 0 ? JSON.stringify(list) : ''
}

onMounted(() => {
  if (props.initialTemplatesJson) {
    loadTemplates(props.initialTemplatesJson)
  }
})

defineExpose({ loadTemplates, clearTemplates, buildJson })
</script>

<style scoped>
.event-template-tabs {
  width: 100%;
}
.event-template-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}
.event-template-tabs :deep(.el-tabs__item) {
  font-size: 13px;
  padding: 0 16px;
}

.variable-panel {
  width: 100%;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 14px;
  background: var(--el-fill-color-blank);
}
.var-group {
  margin-bottom: 12px;
}
.var-group:last-of-type {
  margin-bottom: 0;
}
.var-group-header {
  margin-bottom: 8px;
}
.var-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.var-btn {
  font-size: 12px;
  padding: 4px 10px;
  height: auto;
  border-radius: 4px;
  transition: all 0.2s ease;
}
.var-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}
.active-editor-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 6px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.template-input {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}
.template-input.editor-active :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset, 0 0 0 3px var(--el-color-primary-light-8);
}
.template-input.editor-active :deep(.el-textarea__inner) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset, 0 0 0 3px var(--el-color-primary-light-8);
}

.preview-collapse {
  width: 100%;
  border: none;
}
.preview-collapse :deep(.el-collapse-item__header) {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
  height: 36px;
  line-height: 36px;
}
.template-preview {
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
  padding: 14px;
}
.preview-section {
  margin-bottom: 4px;
}
.preview-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.preview-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  padding: 8px 12px;
  background: var(--el-bg-color);
  border-radius: 4px;
  border-left: 3px solid var(--el-color-primary);
}
.preview-content {
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  padding: 10px 12px;
  background: var(--el-bg-color);
  border-radius: 4px;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: inherit;
}
.preview-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  margin-top: 10px;
}

.empty-template-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  width: 100%;
  text-align: center;
}
.empty-template-hint p {
  margin-top: 12px;
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}
</style>
