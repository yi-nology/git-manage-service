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
      <div v-if="showAIPanel" class="ai-chat-panel">
        <div class="ai-chat-header">
          <div class="ai-mode-switch">
            <span
              :class="['ai-mode-btn', { active: aiMode === 'chat' }]"
              @click="aiMode = 'chat'"
            >对话</span>
            <span
              :class="['ai-mode-btn', { active: aiMode === 'agent' }]"
              @click="aiMode = 'agent'"
            >Agent</span>
          </div>
          <div class="ai-chat-header-actions">
            <el-button text size="small" @click="aiMessages = []; pendingAgentContent = ''">
              <el-icon><Delete /></el-icon>
            </el-button>
            <el-button text size="small" @click="showAIPanel = false">
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
            <el-button size="small" type="primary" @click="acceptAgentChange">接受修改</el-button>
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
                <el-button size="small" type="primary" @click="applyAIContent(msg.applyContent)">
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

    <!-- 规则管理 -->
    <el-dialog v-model="showRuleManager" title="规则管理" width="700px" :close-on-click-modal="false">
      <div class="rule-manager">
        <div class="rm-toolbar">
          <el-input v-model="ruleSearch" placeholder="搜索规则" clearable :prefix-icon="Search" style="width:260px" />
          <el-button type="primary" size="small" :icon="Plus" @click="showCreateRuleDialog = true">创建规则</el-button>
        </div>
        <el-table :data="filteredRules" style="width:100%" max-height="450">
          <el-table-column prop="name" label="规则名称" width="180" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="severity" label="级别" width="90">
            <template #default="{ row }">
              <el-tag :type="row.severity==='error'?'danger':row.severity==='warning'?'warning':'info'" size="small">
                {{ {error:'错误',warning:'警告',info:'信息'}[row.severity as string] }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="启用" width="70">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" @change="handleToggleRule(row)" />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <!-- 创建规则 -->
    <el-dialog v-model="showCreateRuleDialog" title="创建自定义规则" width="500px" append-to-body>
      <el-form :model="newRule" label-width="90px" ref="newRuleFormRef">
        <el-form-item label="规则 ID" required>
          <el-input v-model="newRule.id" placeholder="my-custom-rule" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="newRule.name" placeholder="规则显示名称" />
        </el-form-item>
        <el-form-item label="描述" required>
          <el-input v-model="newRule.description" placeholder="规则描述" />
        </el-form-item>
        <el-form-item label="严重级别">
          <el-select v-model="newRule.severity" style="width:100%">
            <el-option label="错误" value="error" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配模式">
          <el-input v-model="newRule.pattern" placeholder="正则表达式" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateRuleDialog = false">取消</el-button>
        <el-button type="primary" :loading="creatingRule" @click="handleCreateRule">创建</el-button>
      </template>
    </el-dialog>

    <!-- Commit -->
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

    <!-- 初始化 Spec 文件 -->
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

    <!-- 格式化预览 -->
    <el-dialog v-model="showFormatPreview" width="1000px" :close-on-click-modal="false" :show-close="false" class="fmt-preview-dialog">
      <template #header>
        <div class="fmt-diff-header">
          <span class="fmt-diff-title">格式化预览 — {{ formatChanges.length }} 处变更</span>
          <div class="fmt-diff-stats">
            <span class="added">+{{ formatAddedLines }} 行</span>
            <span class="removed">-{{ formatRemovedLines }} 行</span>
          </div>
        </div>
      </template>
      <div class="fmt-diff-tabs">
        <span :class="['fmt-tab', { active: fmtPreviewTab === 'changes' }]" @click="fmtPreviewTab = 'changes'">变更列表</span>
        <span :class="['fmt-tab', { active: fmtPreviewTab === 'diff' }]" @click="fmtPreviewTab = 'diff'">并排对比</span>
      </div>
      <div v-if="fmtPreviewTab === 'changes'" class="fmt-diff-body">
        <div v-if="formatChanges.length > 0" class="fmt-changes-list">
          <el-scrollbar max-height="420px">
            <div v-for="(ch, idx) in formatChanges" :key="idx" class="fmt-change-card">
              <div class="fmt-change-head">
                <span :class="['fmt-change-tag', `fmt-tag-${ch.type}`]">
                  {{ { removed: '删除', modified: '修改', reordered: '排序' }[ch.type] || ch.type }}
                </span>
                <span class="fmt-change-reason">{{ ch.reason }}</span>
              </div>
              <div v-if="ch.before" class="fmt-change-line fmt-line-removed">- {{ ch.before }}</div>
              <div v-if="ch.after" class="fmt-change-line fmt-line-added">+ {{ ch.after }}</div>
            </div>
          </el-scrollbar>
        </div>
        <div v-else class="fmt-no-changes">
          <el-icon :size="28" color="#10B981"><CircleCheck /></el-icon>
          <span>无需格式化，文件已符合规范</span>
        </div>
      </div>
      <div v-else class="fmt-diff-body">
        <el-scrollbar max-height="420px">
          <div class="fmt-side-by-side">
            <div class="fmt-side-col">
              <div class="fmt-side-title">原始文件</div>
              <pre class="fmt-side-content">{{ content }}</pre>
            </div>
            <div class="fmt-side-col">
              <div class="fmt-side-title">格式化后</div>
              <pre class="fmt-side-content">{{ formatResult }}</pre>
            </div>
          </div>
        </el-scrollbar>
      </div>
      <template #footer>
        <div class="fmt-diff-footer">
          <el-button class="fmt-btn-outline" @click="showFormatPreview = false">取消</el-button>
          <el-button type="primary" :disabled="formatChanges.length === 0" @click="applyFormatResult">应用格式化</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 格式化选项 -->
    <el-dialog v-model="showFormatOptions" width="520px" :close-on-click-modal="false" :show-close="false" class="fmt-opt-dialog">
      <template #header>
        <div class="fmt-opt-header">
          <span class="fmt-opt-title">格式化选项</span>
          <el-button text size="small" @click="showFormatOptions = false">
            <el-icon :size="16" color="var(--text-color-secondary, #64748B)"><Close /></el-icon>
          </el-button>
        </div>
      </template>
      <div class="fmt-opt-body">
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">宏括号规范化</span>
            <span class="fmt-opt-desc">将 %macro 转为 %{macro}，排除白名单关键字</span>
          </div>
          <el-switch v-model="formatOpts.curlify" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">删除 %clean 段</span>
            <span class="fmt-opt-desc">现代 RPM 不再需要 %clean section</span>
          </div>
          <el-switch v-model="formatOpts.removeClean" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">删除 BuildRoot</span>
            <span class="fmt-opt-desc">现代 RPM 已废弃 BuildRoot 字段</span>
          </div>
          <el-switch v-model="formatOpts.removeBuildRoot" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">删除 Group 字段</span>
            <span class="fmt-opt-desc">现代 RPM 不再使用 Group 分类</span>
          </div>
          <el-switch v-model="formatOpts.removeGroup" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">License SPDX 修正</span>
            <span class="fmt-opt-desc">将非标准 License 名修正为 SPDX 标识符</span>
          </div>
          <el-switch v-model="formatOpts.licenseSpdx" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">依赖排序去重</span>
            <span class="fmt-opt-desc">BuildRequires/Requires 按字母排序并去重</span>
          </div>
          <el-switch v-model="formatOpts.sortDeps" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">Tab 转空格</span>
            <span class="fmt-opt-desc">所有 Tab 替换为空格</span>
          </div>
          <el-switch v-model="formatOpts.tabToSpaces" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">Preamble 标签排序</span>
            <span class="fmt-opt-desc">按 RPM 规范顺序重排 Name/Version/BuildRequires 等</span>
          </div>
          <el-switch v-model="formatOpts.preambleOrder" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">标签值对齐</span>
            <span class="fmt-opt-desc">统一冒号后缩进，提升可读性</span>
          </div>
          <el-switch v-model="formatOpts.alignValues" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">路径宏替换</span>
            <span class="fmt-opt-desc">硬编码路径 → RPM 宏 (如 /usr/bin → %{_bindir})</span>
          </div>
          <el-switch v-model="formatOpts.pathMacros" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">工具宏展开</span>
            <span class="fmt-opt-desc">%{__make} → make, %{__rm} → rm 等</span>
          </div>
          <el-switch v-model="formatOpts.utilMacros" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">通用清理</span>
            <span class="fmt-opt-desc">$RPM_BUILD_ROOT → %{buildroot}, egrep → grep -E</span>
          </div>
          <el-switch v-model="formatOpts.commonCleanup" />
        </div>
        <div class="fmt-opt-row">
          <div class="fmt-opt-label">
            <span class="fmt-opt-name">条件块空行压缩</span>
            <span class="fmt-opt-desc">移除 %if/%else/%endif 后的多余空行</span>
          </div>
          <el-switch v-model="formatOpts.conditionalTrim" />
        </div>
      </div>
      <template #footer>
        <div class="fmt-opt-footer">
          <el-button class="fmt-btn-outline" @click="showFormatOptions = false">关闭</el-button>
          <el-button type="primary" @click="showFormatOptions = false; handleFormat()">应用并格式化</el-button>
        </div>
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
  MagicStick, Close, Delete, User, Setting, Promotion,
} from '@element-plus/icons-vue'
import * as monaco from 'monaco-editor'
import {
  getSpecTree, getSpecContent, saveSpecContent, lintSpec,
  createSpecFile, getLintRules, updateLintRule, createLintRule,
  commitSpec, aiFixSpec, formatSpec,
} from '@/api/modules/spec'
import request from '@/api/request'
import type { SpecFileNode, LintIssue, LintRule, FormatChange } from '@/types/spec'

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
const lintingInProgress = ref(false)
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
    lintingInProgress.value = true
    const result = await lintSpec(content.value, undefined, mode as any)
    lintIssues.value = (result.issues || []).map(i => ({ ...i, ruleName: i.ruleId || '', source: i.source || 'rule' }))
    updateMonacoMarkers()
  } catch { ElMessage.error('Linting 失败') }
  finally { lintingInProgress.value = false }
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

// Commit
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

// AI Fix from problems panel
async function handleAIFix(issue: LintIssue, idx: number) {
  fixingIndex.value = idx
  try {
    const res = await aiFixSpec(content.value, issue.message, issue.line, issue.severity)
    if (res?.content) { applyAIContent(res.content); ElMessage.success('修复已应用') }
    else ElMessage.warning('AI 未返回修复内容')
  } catch (e: any) { ElMessage.error('修复失败: ' + (e?.message || '')) }
  finally { fixingIndex.value = null }
}

// Rule Manager
const showRuleManager = ref(false)
const rules = ref<LintRule[]>([])
const ruleSearch = ref('')
const showCreateRuleDialog = ref(false)
const creatingRule = ref(false)
const newRuleFormRef = ref()
const newRule = ref({ id: '', name: '', description: '', severity: 'warning', pattern: '', enabled: true })
const filteredRules = computed(() => {
  if (!ruleSearch.value) return rules.value
  const s = ruleSearch.value.toLowerCase()
  return rules.value.filter(r => r.name.toLowerCase().includes(s) || r.description.toLowerCase().includes(s))
})

watch(showRuleManager, async (v) => { if (v) await loadRules() })

async function loadRules() {
  try { rules.value = await getLintRules() } catch { rules.value = [] }
}

async function handleToggleRule(rule: LintRule) {
  try { await updateLintRule(rule.id, { enabled: rule.enabled }) } catch { ElMessage.error('更新规则失败') }
}

async function handleCreateRule() {
  try {
    creatingRule.value = true
    await createLintRule({ ...newRule.value })
    ElMessage.success('规则创建成功')
    showCreateRuleDialog.value = false
    newRule.value = { id: '', name: '', description: '', severity: 'warning', pattern: '', enabled: true }
    await loadRules()
  } catch { ElMessage.error('创建规则失败') }
  finally { creatingRule.value = false }
}

// Init Spec
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

// AI Chat
interface AIChatMessage { role: 'user' | 'assistant'; content: string; applyContent?: string }
const showAIPanel = ref(false)
const aiMessages = ref<AIChatMessage[]>([])
const aiInput = ref('')
const aiLoading = ref(false)
const activeAction = ref('')
const messagesRef = ref<HTMLDivElement>()
const aiMode = ref<'chat' | 'agent'>('chat')

watch(showAIPanel, () => {
  nextTick(() => { if (editorInstance) editorInstance.layout() })
})
const pendingAgentContent = ref('')

const agentDiffText = computed(() => {
  if (!pendingAgentContent.value) return ''
  const orig = content.value.split('\n')
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

const quickActions = [
  { key: 'check', label: '检查问题', prompt: '请检查这个 spec 文件是否存在语法错误、缺失字段或不符合 RPM 打包规范的问题，列出所有发现的问题并给出修改建议。' },
  { key: 'complete', label: '补全字段', prompt: '请分析这个 spec 文件，补全所有缺失的必需字段和可选但推荐的字段（如 URL、Source、BuildRequires 等），直接输出补全后的完整内容。' },
  { key: 'optimize', label: '优化建议', prompt: '请分析这个 spec 文件，给出优化建议，包括：构建流程优化、依赖精简、文件列表完善、宏使用规范等。' },
  { key: 'explain', label: '解释说明', prompt: '请逐段解释这个 spec 文件的内容和作用，对每个 section 和重要指令进行说明。' },
  { key: 'generate', label: '生成构建段', prompt: '请根据当前 spec 文件的 Name、Version、Source 等元信息，帮我生成完整的 %build、%install 和 %files 段内容。直接输出修改后的完整 spec 文件。' },
]

function renderMarkdown(t: string): string {
  return t.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>').replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>').replace(/\n/g, '<br>')
}
function scrollAIToBottom() { nextTick(() => { if (messagesRef.value) messagesRef.value.scrollTop = messagesRef.value.scrollHeight }) }

async function callAI(prompt: string, action?: string) {
  if (!content.value) { ElMessage.warning('请先选择一个 Spec 文件'); return }
  aiMessages.value.push({ role: 'user', content: prompt })
  aiLoading.value = true; activeAction.value = action || ''; scrollAIToBottom()
  try {
    const effectiveAction = action || (aiMode.value === 'agent' ? 'agent' : 'chat')
    const history = aiMessages.value.slice(-10).map(m => ({ role: m.role, content: m.content }))
    const res = await request.post<unknown, { result: string; apply_content?: string }>('/spec/ai-assist', { content: content.value, prompt, action: effectiveAction, history })
    if (effectiveAction === 'agent' && res?.apply_content) {
      pendingAgentContent.value = res.apply_content
      aiMessages.value.push({ role: 'assistant', content: '已生成修改，请查看并确认。' })
    } else {
      const msg: AIChatMessage = { role: 'assistant', content: res?.result || '无结果' }
      if (res?.apply_content) msg.applyContent = res.apply_content
      aiMessages.value.push(msg)
    }
  } catch (e: any) { aiMessages.value.push({ role: 'assistant', content: `请求失败: ${e?.message || '未知错误'}` }) }
  finally { aiLoading.value = false; activeAction.value = ''; scrollAIToBottom() }
}

function sendAIMessage() { const t = aiInput.value.trim(); if (!t || aiLoading.value) return; aiInput.value = ''; callAI(t) }
function handleQuickAction(a: { key: string; prompt: string }) { callAI(a.prompt, a.key) }
function applyAIContent(c: string) {
  content.value = c; isDirty.value = c !== originalContent.value
  if (editorInstance) editorInstance.setValue(c)
  ElMessage.success('AI 内容已应用到编辑器')
}

function acceptAgentChange() {
  if (!pendingAgentContent.value) return
  content.value = pendingAgentContent.value
  isDirty.value = content.value !== originalContent.value
  if (editorInstance) editorInstance.setValue(content.value)
  pendingAgentContent.value = ''
  ElMessage.success('已接受 AI 修改')
}

function rejectAgentChange() {
  pendingAgentContent.value = ''
  ElMessage.info('已拒绝修改')
}

// Format
const showFormatPreview = ref(false)
const showFormatOptions = ref(false)
const formatChanges = ref<FormatChange[]>([])
const formatResult = ref('')
const fmtPreviewTab = ref<'changes' | 'diff'>('changes')
const formatAddedLines = computed(() => formatChanges.value.filter(c => c.after).length)
const formatRemovedLines = computed(() => formatChanges.value.filter(c => c.before).length)
const formatOpts = ref({
  curlify: true,
  removeClean: true,
  removeBuildRoot: true,
  removeGroup: false,
  licenseSpdx: true,
  sortDeps: true,
  tabToSpaces: true,
  indentSize: 4,
  preambleOrder: true,
  alignValues: true,
  pathMacros: true,
  utilMacros: true,
  commonCleanup: true,
  conditionalTrim: true,
})

async function handleFormat() {
  if (!content.value) { ElMessage.warning('请先选择一个 Spec 文件'); return }
  try {
    const result = await formatSpec(content.value, { ...formatOpts.value })
    formatResult.value = result.content
    formatChanges.value = result.changes || []
    if (formatChanges.value.length === 0 && result.content === content.value) {
      ElMessage.success('无需格式化，文件已符合规范')
      return
    }
    showFormatPreview.value = true
  } catch (e: any) {
    ElMessage.error('格式化失败: ' + (e?.message || ''))
  }
}

function applyFormatResult() {
  if (!formatResult.value) return
  content.value = formatResult.value
  isDirty.value = content.value !== originalContent.value
  if (editorInstance) editorInstance.setValue(content.value)
  showFormatPreview.value = false
  ElMessage.success('格式化已应用')
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

/* AI Chat Panel */
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

/* Rule Manager */
.rule-manager { min-height: 400px; }
.rm-toolbar { display: flex; gap: 12px; margin-bottom: 16px; }

/* Commit Dialog */
.diff-preview { width: 100%; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; overflow: hidden; }
.diff-stats { padding: 8px 12px; border-bottom: 1px solid var(--border-color, #E2E8F0); font-size: 13px; color: var(--text-color-regular, #475569); }
.added { color: #22c55e; margin-right: 12px; }
.removed { color: #ef4444; }
.diff-content { padding: 12px; margin: 0; font-size: 12px; font-family: 'Consolas', monospace; color: var(--text-color-regular, #475569); line-height: 1.5; }

/* Format Options Dialog */
.fmt-opt-dialog :deep(.el-dialog) { background: var(--bg-color-page, #FFFFFF); border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-opt-dialog :deep(.el-dialog__header) { margin: 0; padding: 0; }
.fmt-opt-dialog :deep(.el-dialog__body) { padding: 0; }
.fmt-opt-dialog :deep(.el-dialog__footer) { padding: 0; }

/* Format Preview Dialog */
.fmt-preview-dialog :deep(.el-dialog) { background: var(--bg-color-page, #FFFFFF); border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-preview-dialog :deep(.el-dialog__header) { margin: 0; padding: 0; }
.fmt-preview-dialog :deep(.el-dialog__body) { padding: 0; }
.fmt-preview-dialog :deep(.el-dialog__footer) { padding: 0; }
</style>

<style>
.fmt-opt-dialog .el-dialog { background: var(--bg-color-page, #FFFFFF) !important; border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-opt-dialog .el-dialog__header { margin: 0; padding: 0; }
.fmt-opt-dialog .el-dialog__body { padding: 0; }
.fmt-opt-dialog .el-dialog__footer { padding: 0; }
.fmt-opt-header { height: 48px; padding: 0 20px; display: flex; align-items: center; justify-content: space-between; background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); }
.fmt-opt-title { font-size: 14px; font-weight: 600; color: var(--text-color-primary, #1E293B); }
.fmt-opt-body { padding: 20px 24px; display: flex; flex-direction: column; gap: 20px; }
.fmt-opt-row { display: flex; align-items: center; justify-content: space-between; }
.fmt-opt-label { display: flex; flex-direction: column; gap: 2px; max-width: 380px; }
.fmt-opt-name { font-size: 13px; font-weight: 500; color: var(--text-color-primary, #1E293B); }
.fmt-opt-desc { font-size: 11px; color: var(--text-color-secondary, #64748B); }
.fmt-opt-row .el-switch { --el-switch-on-color: #6366F1; --el-switch-off-color: #CBD5E1; }
.fmt-opt-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 0 20px; height: 56px; align-items: center; background: var(--bg-color, #F8F9FC); border-top: 1px solid var(--border-color, #E2E8F0); border-radius: 0 0 12px 12px; }
.fmt-btn-outline { color: var(--text-color-regular, #475569) !important; border-color: var(--border-color, #E2E8F0) !important; background: transparent !important; }
.fmt-btn-outline:hover { border-color: var(--primary-color, #6366F1) !important; color: var(--primary-color, #6366F1) !important; }

.fmt-preview-dialog .el-dialog { background: var(--bg-color-page, #FFFFFF) !important; border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-preview-dialog .el-dialog__header { margin: 0; padding: 0; }
.fmt-preview-dialog .el-dialog__body { padding: 0; }
.fmt-preview-dialog .el-dialog__footer { padding: 0; }
.fmt-diff-header { height: 48px; padding: 0 20px; display: flex; align-items: center; justify-content: space-between; background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); border-radius: 12px 12px 0 0; }
.fmt-diff-title { font-size: 14px; font-weight: 600; color: var(--text-color-primary, #1E293B); }
.fmt-diff-stats { display: flex; gap: 12px; font-size: 12px; font-weight: 500; }
.fmt-diff-stats .added { color: #22C55E; }
.fmt-diff-stats .removed { color: #EF4444; }
.fmt-diff-tabs { height: 36px; padding: 0 20px; display: flex; gap: 16px; align-items: center; background: var(--bg-color-page, #FFFFFF); border-bottom: 1px solid var(--border-color, #E2E8F0); }
.fmt-tab { font-size: 12px; color: var(--text-color-secondary, #64748B); cursor: pointer; transition: color 0.15s; }
.fmt-tab:hover { color: var(--text-color-primary, #1E293B); }
.fmt-tab.active { color: #6366F1; font-weight: 500; }
.fmt-diff-body { padding: 16px; min-height: 200px; }
.fmt-changes-list { display: flex; flex-direction: column; gap: 10px; }
.fmt-change-card { padding: 10px 14px; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; display: flex; flex-direction: column; gap: 6px; }
.fmt-change-head { display: flex; align-items: center; gap: 8px; }
.fmt-change-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; color: #fff; }
.fmt-tag-removed { background: #EF4444; }
.fmt-tag-modified { background: #F59E0B; }
.fmt-tag-reordered { background: #6366F1; }
.fmt-change-reason { font-size: 11px; color: var(--text-color-secondary, #64748B); }
.fmt-change-line { font-family: 'Consolas', 'Courier New', monospace; font-size: 11px; white-space: pre-wrap; word-break: break-all; }
.fmt-line-removed { color: #EF4444; }
.fmt-line-added { color: #22C55E; }
.fmt-no-changes { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; padding: 48px; color: #10B981; font-size: 14px; }
.fmt-diff-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 0 20px; height: 56px; align-items: center; background: var(--bg-color, #F8F9FC); border-top: 1px solid var(--border-color, #E2E8F0); border-radius: 0 0 12px 12px; }
.fmt-side-by-side { display: flex; gap: 12px; }
.fmt-side-col { flex: 1; min-width: 0; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; overflow: hidden; }
.fmt-side-title { padding: 8px 14px; font-size: 12px; font-weight: 500; color: var(--text-color-secondary, #64748B); background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); }
.fmt-side-content { margin: 0; padding: 12px 14px; font-size: 11px; font-family: 'Consolas', 'Courier New', monospace; color: var(--text-color-regular, #475569); line-height: 1.6; white-space: pre-wrap; word-break: break-all; }
</style>
