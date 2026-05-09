<template>
  <div class="spec-editor-container">
    <div class="toolbar">
      <div class="toolbar-left">
        <h3>Spec 编辑器</h3>
        <span v-if="currentFile" class="current-file">
          {{ currentFile }}
          <el-tag v-if="isDirty" type="warning" size="small">未保存</el-tag>
        </span>
      </div>
      <div class="toolbar-right">
        <el-button
          size="small"
          :type="showAIPanel ? 'primary' : 'default'"
          :class="{ 'ai-active-btn': showAIPanel }"
          @click="showAIPanel = !showAIPanel"
        >
          <el-icon><MagicStick /></el-icon> AI
        </el-button>
        <el-button size="small" type="primary" @click="showFormatOptions = true">
          <el-icon><MagicStick /></el-icon> 格式化
        </el-button>
        <el-dropdown split-button size="small" class="tb-outline-btn" @click="handleLint">
          <el-icon><DocumentChecked /></el-icon> 检查
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleLintWithMode('rule_only')">仅规则</el-dropdown-item>
              <el-dropdown-item @click="handleLintWithMode('rule_and_ai')">规则 + AI</el-dropdown-item>
              <el-dropdown-item @click="handleLintWithMode('ai_only')">仅 AI</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button size="small" @click="showRuleManager = true" class="tb-outline-btn">
          <el-icon><Setting /></el-icon> 规则
        </el-button>
        <el-button
          size="small"
          :disabled="!isDirty || hasErrors()"
          :loading="savingInProgress"
          @click="saveCurrentFile"
          class="tb-outline-btn"
        >
          <el-icon><Download /></el-icon> 保存
        </el-button>
        <el-button
          size="small"
          type="primary"
          :disabled="!isDirty"
          @click="showCommitDialog = true"
        >
          <el-icon><Promotion /></el-icon> Commit
        </el-button>
      </div>
    </div>

    <div class="editor-layout">
      <AIChatPanel
        v-if="showAIPanel"
        :content="content"
        @apply-content="applyAIContent"
        @close="showAIPanel = false"
      />

      <div class="file-tree-panel">
        <div class="tree-header">
          <el-input
            v-model="filterText"
            placeholder="搜索文件"
            clearable
            size="small"
            :prefix-icon="Search"
          />
        </div>
        <el-scrollbar>
          <el-tree
            ref="treeRef"
            :data="fileTree"
            :props="{ label: 'name', children: 'children' }"
            :filter-node-method="filterNode"
            node-key="path"
            highlight-current
            :expand-on-click-node="false"
            @node-click="handleNodeClick"
          >
            <template #default="{ node, data }">
              <span class="custom-tree-node">
                <el-icon v-if="data.is_dir" class="folder-icon"><Folder /></el-icon>
                <el-icon v-else class="file-icon"><Document /></el-icon>
                <span class="node-label">{{ node.label }}</span>
              </span>
            </template>
          </el-tree>
          <div v-if="fileTree.length === 0 && !loading" class="empty-tree-container">
            <el-empty description="此仓库暂无 .spec 文件">
              <el-button type="primary" @click="showInitDialog = true">
                <el-icon><Plus /></el-icon> 初始化 Spec 文件
              </el-button>
            </el-empty>
          </div>
        </el-scrollbar>
      </div>

      <div class="editor-panel">
        <div class="monaco-container" ref="monacoContainer">
          <div v-if="!content" class="empty-editor">
            <el-empty description="请选择 .spec 文件" />
          </div>
        </div>
        <div class="problems-panel">
          <div class="problems-header">
            <span>问题 ({{ lintIssues.length }})</span>
            <el-tag v-if="errorCount > 0" type="danger" size="small">{{ errorCount }} 错误</el-tag>
            <el-tag v-if="warningCount > 0" type="warning" size="small">{{ warningCount }} 警告</el-tag>
          </div>
          <el-scrollbar>
            <div v-if="lintIssues.length === 0" class="no-problems">
              <el-icon><CircleCheck /></el-icon>
              <span>没有问题</span>
            </div>
            <div
              v-for="(issue, idx) in lintIssues"
              :key="`${issue.line}-${idx}`"
              class="problem-item"
              :class="`problem-${issue.severity}`"
              @click="goToLine(issue.line, issue.column)"
            >
              <el-icon v-if="issue.severity === 'error'" color="#f56c6c"><CircleClose /></el-icon>
              <el-icon v-else-if="issue.severity === 'warning'" color="#e6a23c"><WarningFilled /></el-icon>
              <el-icon v-else color="#909399"><InfoFilled /></el-icon>
              <div class="problem-body">
                <div class="problem-msg">
                  <span>{{ issue.message }}</span>
                  <el-tag v-if="issue.source === 'ai'" size="small" class="ai-badge">AI</el-tag>
                </div>
                <div class="problem-meta">
                  <span>Line {{ issue.line }}</span>
                </div>
                <div v-if="issue.quickFix && issue.source === 'ai'" class="problem-fix">
                  <span class="fix-hint">{{ issue.quickFix }}</span>
                  <el-button
                    size="small"
                    type="primary"
                    :loading="fixingIndex === idx"
                    @click.stop="handleAIFix(issue, idx)"
                  >修复</el-button>
                </div>
              </div>
            </div>
          </el-scrollbar>
        </div>
      </div>
    </div>

    <RuleManagerDialog v-model="showRuleManager" />

    <FormatPreviewDialog
      v-model="showFormatOptions"
      :content="content"
      @apply-format="applyFormatResult"
    />

    <el-dialog v-model="showCommitDialog" title="提交变更" width="600px" :close-on-click-modal="false">
      <el-form label-width="110px">
        <el-form-item label="Commit 消息">
          <el-input v-model="commitMsg" type="textarea" :rows="4" placeholder="输入 commit 消息" />
        </el-form-item>
        <el-form-item label="变更预览">
          <div class="diff-preview">
            <div class="diff-stats">
              <span class="added">+{{ addedLines }}</span>
              <span class="removed">-{{ removedLines }}</span>
            </div>
            <el-scrollbar max-height="250px">
              <pre class="diff-content">{{ diffPreview }}</pre>
            </el-scrollbar>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCommitDialog = false">取消</el-button>
        <el-button type="primary" :loading="committing" @click="handleCommit">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showInitDialog" title="初始化 Spec 文件" width="600px" :close-on-click-modal="false">
      <el-form :model="initForm" label-width="100px" :rules="initFormRules" ref="initFormRef">
        <el-form-item label="文件名" prop="filename">
          <el-input v-model="initForm.filename" placeholder="mypackage"><template #append>.spec</template></el-input>
        </el-form-item>
        <el-form-item label="Name" prop="name"><el-input v-model="initForm.name" /></el-form-item>
        <el-form-item label="Version" prop="version"><el-input v-model="initForm.version" /></el-form-item>
        <el-form-item label="Release" prop="release"><el-input v-model="initForm.release" /></el-form-item>
        <el-form-item label="Summary" prop="summary"><el-input v-model="initForm.summary" /></el-form-item>
        <el-form-item label="License" prop="license">
          <el-select v-model="initForm.license" filterable>
            <el-option label="MIT" value="MIT" />
            <el-option label="Apache-2.0" value="Apache-2.0" />
            <el-option label="GPL-3.0" value="GPL-3.0" />
            <el-option label="BSD-3-Clause" value="BSD-3-Clause" />
          </el-select>
        </el-form-item>
        <el-form-item label="URL"><el-input v-model="initForm.url" /></el-form-item>
        <el-form-item label="描述">
          <el-input v-model="initForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInitDialog = false">取消</el-button>
        <el-button type="primary" :loading="initInProgress" @click="handleInitSpec">创建并打开</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, Folder, Document, DocumentChecked, Download,
  CircleCheck, CircleClose, WarningFilled, InfoFilled, Plus,
  MagicStick, Setting, Promotion,
} from '@element-plus/icons-vue'
import * as monaco from 'monaco-editor'
import {
  getSpecTree, getSpecContent, saveSpecContent, lintSpec,
  createSpecFile, commitSpec, aiFixSpec,
} from '@/api/modules/spec'
import type { SpecFileNode, LintIssue } from '@/types/spec'
import AIChatPanel from './AIChatPanel.vue'
import RuleManagerDialog from './RuleManagerDialog.vue'
import FormatPreviewDialog from './FormatPreviewDialog.vue'
import '../../spec-editor.css'

interface Props { repoKey: string }
const props = defineProps<Props>()

const filterText = ref('')
const fileTree = ref<SpecFileNode[]>([])
const currentFile = ref('')
const content = ref('')
const originalContent = ref('')
const isDirty = ref(false)
const lintIssues = ref<LintIssue[]>([])
const loading = ref(false)
const savingInProgress = ref(false)
const committing = ref(false)
const treeRef = ref()
const monacoContainer = ref<HTMLElement>()
let editorInstance: monaco.editor.IStandaloneCodeEditor | null = null
const errorCount = computed(() => lintIssues.value.filter(i => i.severity === 'error').length)
const warningCount = computed(() => lintIssues.value.filter(i => i.severity === 'warning').length)
let lintDebounceTimer: ReturnType<typeof setTimeout> | null = null
const lintMode = ref<'rule_only' | 'rule_and_ai' | 'ai_only'>('rule_only')
const fixingIndex = ref<number | null>(null)

const showAIPanel = ref(false)
const showRuleManager = ref(false)
const showFormatOptions = ref(false)

watch(showAIPanel, () => {
  nextTick(() => { if (editorInstance) editorInstance.layout() })
})

defineExpose({
  refresh: loadFileTree,
  clearEditor: () => {
    currentFile.value = ''
    content.value = ''
    originalContent.value = ''
    isDirty.value = false
    lintIssues.value = []
    if (editorInstance) editorInstance.setValue('')
  }
})

watch(filterText, (val) => { treeRef.value?.filter(val) })

onMounted(async () => {
  await loadFileTree()
  await nextTick()
  initMonaco()
})

onBeforeUnmount(() => {
  if (editorInstance) editorInstance.dispose()
  if (lintDebounceTimer) clearTimeout(lintDebounceTimer)
})

async function loadFileTree() {
  try {
    loading.value = true
    const tree = await getSpecTree(props.repoKey)
    fileTree.value = Array.isArray(tree) ? tree : []
  } catch (e: any) {
    if (e?.code === 'ERR_CANCELED') return
    ElMessage.error('加载文件树失败')
    fileTree.value = []
  } finally { loading.value = false }
}

async function loadFile(path: string) {
  if (isDirty.value) {
    const ok = await ElMessageBox.confirm('当前文件未保存，是否继续？', '提示', { type: 'warning' }).catch(() => false)
    if (!ok) return
  }
  try {
    loading.value = true
    const { content: fc } = await getSpecContent(path, props.repoKey)
    currentFile.value = path
    content.value = fc
    originalContent.value = fc
    isDirty.value = false
    lintIssues.value = []
    if (editorInstance) editorInstance.setValue(fc)
  } catch (e: any) {
    if (e?.code === 'ERR_CANCELED') return
    ElMessage.error('加载文件失败')
  } finally { loading.value = false }
}

function initMonaco() {
  if (!monacoContainer.value) return
  monaco.languages.register({ id: 'rpmspec' })
  monaco.languages.setMonarchTokensProvider('rpmspec', {
    keywords: ['Name','Version','Release','Summary','License','URL','Source0','Patch0','BuildArch','BuildRoot','BuildRequires','Requires','Provides','Obsoletes','Conflicts','%description','%prep','%build','%install','%clean','%files','%changelog','%package','%post','%postun','%pre','%preun'],
    tokenizer: {
      root: [
        [/#.*/, 'comment'], [/%\w+/, 'keyword'], [/\$\w+/, 'variable'],
        [/\$\{\w+\}/, 'variable'], [/%\{\w+\}/, 'variable'],
        [/[<>=]+/, 'operator'], [/\d+\.\d+\.\d+/, 'number'], [/\d+/, 'number'],
        [/"([^"]*)"/, 'string'], [/'([^']*)'/, 'string'],
      ]
    }
  })
  editorInstance = monaco.editor.create(monacoContainer.value, {
    value: content.value, language: 'rpmspec', theme: 'vs',
    automaticLayout: true, fontSize: 14, lineNumbers: 'on',
    minimap: { enabled: true }, scrollBeyondLastLine: false,
  })
  editorInstance.onDidChangeModelContent(() => {
    content.value = editorInstance?.getValue() || ''
    isDirty.value = content.value !== originalContent.value
    if (lintDebounceTimer) clearTimeout(lintDebounceTimer)
    lintDebounceTimer = setTimeout(() => { if (content.value) doLint('rule_only') }, 500)
  })
}

async function doLint(mode: string) {
  if (!content.value) return
  try {
    const result = await lintSpec(content.value, undefined, mode as any)
    lintIssues.value = (result.issues || []).map(i => ({ ...i, ruleName: i.ruleId || '', source: i.source || 'rule' }))
    updateMonacoMarkers()
  } catch { ElMessage.error('Linting 失败') }
}

function updateMonacoMarkers() {
  if (!editorInstance) return
  const model = editorInstance.getModel()
  if (!model) return
  monaco.editor.setModelMarkers(model, 'rpmspec', lintIssues.value.map(i => ({
    severity: i.severity === 'error' ? monaco.MarkerSeverity.Error : i.severity === 'warning' ? monaco.MarkerSeverity.Warning : monaco.MarkerSeverity.Info,
    message: i.message, startLineNumber: i.line, startColumn: i.column || 1,
    endLineNumber: i.endLine || i.line, endColumn: i.endColumn || (i.column || 1) + 10,
  })))
}

function handleLint() { doLint(lintMode.value) }
function handleLintWithMode(mode: 'rule_only' | 'rule_and_ai' | 'ai_only') {
  lintMode.value = mode
  doLint(mode)
}

function hasErrors() { return errorCount.value > 0 }

async function saveCurrentFile() {
  if (!currentFile.value) return
  if (errorCount.value > 0) { ElMessage.warning(`发现 ${errorCount.value} 个错误，请先修复`); return }
  try {
    savingInProgress.value = true
    await saveSpecContent(currentFile.value, { content: content.value, message: `chore(spec): update ${currentFile.value}` }, props.repoKey)
    originalContent.value = content.value
    isDirty.value = false
    ElMessage.success('保存成功')
  } catch { ElMessage.error('保存失败') }
  finally { savingInProgress.value = false }
}

const showCommitDialog = ref(false)
const commitMsg = ref('')
const diffPreview = computed(() => {
  const orig = originalContent.value.split('\n'), cur = content.value.split('\n')
  const d: string[] = []
  for (let i = 0; i < Math.max(orig.length, cur.length); i++) {
    if ((orig[i]||'') !== (cur[i]||'')) {
      if (orig[i]) d.push(`- ${orig[i]}`)
      if (cur[i]) d.push(`+ ${cur[i]}`)
    }
  }
  return d.join('\n') || '没有变更'
})
const addedLines = computed(() => (diffPreview.value.match(/^\+/gm) || []).length)
const removedLines = computed(() => (diffPreview.value.match(/^-/gm) || []).length)

async function handleCommit() {
  if (!commitMsg.value.trim() || !currentFile.value) return
  try {
    committing.value = true
    await commitSpec(currentFile.value, { message: commitMsg.value, content: content.value }, props.repoKey)
    originalContent.value = content.value
    isDirty.value = false
    showCommitDialog.value = false
    commitMsg.value = ''
    ElMessage.success('提交成功')
  } catch { ElMessage.error('提交失败') }
  finally { committing.value = false }
}

async function handleAIFix(issue: LintIssue, idx: number) {
  fixingIndex.value = idx
  try {
    const res = await aiFixSpec(content.value, issue.message, issue.line, issue.severity)
    if (res?.content) { applyAIContent(res.content); ElMessage.success('修复已应用') }
    else ElMessage.warning('AI 未返回修复内容')
  } catch (e: any) { ElMessage.error('修复失败: ' + (e?.message || '')) }
  finally { fixingIndex.value = null }
}

function applyAIContent(c: string) {
  content.value = c
  isDirty.value = c !== originalContent.value
  if (editorInstance) editorInstance.setValue(c)
  ElMessage.success('AI 内容已应用到编辑器')
}

function applyFormatResult(c: string) {
  content.value = c
  isDirty.value = c !== originalContent.value
  if (editorInstance) editorInstance.setValue(c)
}

const showInitDialog = ref(false)
const initInProgress = ref(false)
const initFormRef = ref()
const initForm = ref({ filename: '', name: '', version: '1.0.0', release: '1', summary: '', license: 'MIT', url: '', description: '' })
const initFormRules = {
  filename: [{ required: true, message: '请输入文件名', trigger: 'blur' }],
  name: [{ required: true, message: '请输入软件包名称', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
  summary: [{ required: true, message: '请输入简要描述', trigger: 'blur' }],
}

async function handleInitSpec() {
  if (!initFormRef.value) return
  try { await initFormRef.value.validate() } catch { return }
  try {
    initInProgress.value = true
    const fn = initForm.value.filename.endsWith('.spec') ? initForm.value.filename : `${initForm.value.filename}.spec`
    const res = await createSpecFile({ repo_key: props.repoKey, path: '.', name: fn, content: generateSpecTemplate(initForm.value) })
    ElMessage.success('Spec 文件创建成功')
    showInitDialog.value = false
    await loadFileTree()
    await loadFile(res.path)
    initForm.value = { filename: '', name: '', version: '1.0.0', release: '1', summary: '', license: 'MIT', url: '', description: '' }
  } catch { ElMessage.error('创建失败') }
  finally { initInProgress.value = false }
}

function generateSpecTemplate(f: typeof initForm.value): string {
  return `Name:           ${f.name}
Version:        ${f.version}
Release:        ${f.release}%{?dist}
Summary:        ${f.summary}

License:        ${f.license}
URL:            ${f.url || 'https://example.com'}
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  make

%description
${f.description || f.name}

%prep
%setup -q

%build
%configure
make %{?_smp_mflags}

%install
rm -rf %{buildroot}
%make_install

%files
%doc README.md
%license LICENSE
%{_bindir}/%{name}

%changelog
* $(date '+%a %b %d %Y') Your Name <your.email@example.com> - ${f.version}-${f.release}
- Initial package
`
}

function filterNode(value: string, data: Record<string, unknown>) {
  if (!value) return true
  return (data.name as string).toLowerCase().includes(value.toLowerCase())
}
function handleNodeClick(data: SpecFileNode) { if (!data.is_dir) loadFile(data.path) }
function goToLine(line: number, column?: number) {
  if (!editorInstance) return
  editorInstance.revealLineInCenter(line)
  editorInstance.setPosition({ lineNumber: line, column: column || 1 })
  editorInstance.focus()
}
</script>

<style scoped>
.spec-editor-container { height: 100%; display: flex; flex-direction: column; background: var(--bg-color-page, #fff); color: var(--text-color-primary, #1E293B); border: 1px solid var(--border-color, #E2E8F0); border-radius: var(--border-radius-lg, 12px); overflow: hidden; }
.toolbar { height: 50px; padding: 0 16px; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid var(--border-color, #E2E8F0); background: var(--bg-color-page, #fff); }
.toolbar-left { display: flex; align-items: center; gap: 12px; }
.toolbar-left h3 { margin: 0; font-size: 15px; font-weight: 600; color: var(--text-color-primary, #1E293B); }
.current-file { display: flex; align-items: center; gap: 8px; color: var(--text-color-secondary, #64748B); font-size: 13px; }
.toolbar-right { display: flex; gap: 8px; align-items: center; }
.toolbar-right .ai-active-btn, .toolbar-right .ai-active-btn:focus { background: #7C3AED !important; border-color: #7C3AED !important; color: #fff !important; }
.tb-outline-btn { color: var(--text-color-regular, #475569) !important; border-color: var(--border-color, #E2E8F0) !important; background: transparent !important; }
.tb-outline-btn:hover { border-color: var(--primary-color, #6366F1) !important; color: var(--primary-color, #6366F1) !important; }
.tb-outline-btn.is-disabled { color: var(--text-color-placeholder, #94A3B8) !important; border-color: var(--border-color-light, #F1F5F9) !important; }
.editor-layout { flex: 1; display: flex; overflow: hidden; min-height: 0; min-width: 0; }

.file-tree-panel { width: 220px; min-width: 160px; flex: 0 0 auto; display: flex; flex-direction: column; border-right: 1px solid var(--border-color, #E2E8F0); background: var(--bg-color-page, #fff); overflow: hidden; }
.tree-header { padding: 12px; border-bottom: 1px solid var(--border-color, #E2E8F0); }
.file-tree-panel :deep(.el-tree) { background: transparent; color: var(--text-color-regular, #475569); }
.file-tree-panel :deep(.el-tree-node__content:hover) { background-color: var(--surface-hover, #F8F9FC); }
.file-tree-panel :deep(.el-tree-node.is-current > .el-tree-node__content) { background-color: var(--accent-bg, #EEF2FF); }
.custom-tree-node { display: flex; align-items: center; gap: 6px; font-size: 13px; }
.folder-icon { color: #F59E0B; }
.file-icon { color: #6366F1; }
.node-label { flex: 1; }

.editor-panel { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; }
.monaco-container { flex: 1; min-height: 0; position: relative; }
.empty-editor { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: var(--bg-color-page, #fff); }

.problems-panel { height: 200px; border-top: 1px solid var(--border-color, #E2E8F0); display: flex; flex-direction: column; background: var(--bg-color-page, #fff); }
.problems-header { height: 36px; padding: 0 12px; display: flex; align-items: center; gap: 8px; background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); font-size: 13px; font-weight: 500; color: var(--text-color-primary, #1E293B); }
.problems-panel .el-scrollbar { flex: 1; }
.no-problems { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 24px; color: var(--success-color, #10B981); font-size: 14px; }
.problem-item { padding: 6px 12px; display: flex; align-items: flex-start; gap: 8px; cursor: pointer; font-size: 13px; border-bottom: 1px solid var(--border-color-light, #F1F5F9); color: var(--text-color-regular, #475569); }
.problem-item:hover { background: var(--surface-hover, #F8F9FC); }
.problem-error { border-left: 3px solid var(--danger-color, #EF4444); }
.problem-warning { border-left: 3px solid var(--warning-color, #F59E0B); }
.problem-info { border-left: 3px solid var(--info-color, #94A3B8); }
.problem-body { flex: 1; min-width: 0; }
.problem-msg { display: flex; align-items: center; gap: 6px; }
.ai-badge { font-size: 10px; height: 18px; padding: 0 4px; line-height: 16px; background: #7c3aed; border-color: #7c3aed; color: #fff; }
.problem-meta { font-size: 12px; color: var(--text-color-secondary, #64748B); margin-top: 2px; }
.problem-fix { margin-top: 6px; display: flex; align-items: flex-start; gap: 8px; }
.fix-hint { flex: 1; font-size: 12px; color: var(--text-color-secondary, #64748B); background: var(--bg-color, #F8F9FC); padding: 4px 8px; border-radius: 4px; }
.problem-fix .el-button { flex-shrink: 0; font-size: 12px; height: 24px; padding: 0 8px; }

.empty-tree-container { padding: 24px; text-align: center; }
.empty-tree-container :deep(.el-empty__description) { color: var(--text-color-secondary, #64748B); }

.diff-preview { width: 100%; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; overflow: hidden; }
.diff-stats { padding: 8px 12px; border-bottom: 1px solid var(--border-color, #E2E8F0); font-size: 13px; color: var(--text-color-regular, #475569); }
.added { color: #22c55e; margin-right: 12px; }
.removed { color: #ef4444; }
.diff-content { padding: 12px; margin: 0; font-size: 12px; font-family: 'Consolas', monospace; color: var(--text-color-regular, #475569); line-height: 1.5; }
</style>
