<template>
  <div class="cr-settings-page">
    <PageHeader title="代码审查设置" subtitle="配置代码审查行为和规则" show-back back-route="/settings" />

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'config' }" @click="activeTab = 'config'">审查配置</button>
      <button class="tab-btn" :class="{ active: activeTab === 'rules' }" @click="activeTab = 'rules'">规则管理</button>
      <button class="tab-btn" :class="{ active: activeTab === 'prompt' }" @click="activeTab = 'prompt'">Prompt 结构</button>
    </div>

    <div v-show="activeTab === 'config'" class="tab-content">
      <div class="settings-form-card">
        <div class="form-row">
          <div class="form-label-col">
            <span class="form-label">启用代码审查</span>
            <span class="form-desc">全局开关，关闭后所有仓库不再触发审查</span>
          </div>
          <el-switch v-model="settings.enabled" />
        </div>
        <div class="form-row">
          <div class="form-label-col">
            <span class="form-label">自动审查 MR</span>
            <span class="form-desc">MR 创建或更新时自动触发代码审查</span>
          </div>
          <el-switch v-model="settings.auto_review_on_mr" />
        </div>
        <div class="form-row">
          <div class="form-label-col">
            <span class="form-label">高危阻止合并</span>
            <span class="form-desc">检测到严重或高危问题时自动阻止合并</span>
          </div>
          <el-switch v-model="settings.block_on_high" />
        </div>
        <div class="form-row">
          <div class="form-label-col">
            <span class="form-label">启用 RAG 增强</span>
            <span class="form-desc">利用向量检索代码上下文，增强 LLM 审查质量。需在 LLM 配置中启用至少一个 Embedding 提供商。</span>
          </div>
          <el-switch v-model="settings.rag_enabled" />
        </div>
        <div class="form-row-inline">
          <div class="form-field">
            <label>单次最大审查文件数</label>
            <el-input-number v-model="settings.max_files" :min="1" :max="500" />
          </div>
          <div class="form-field">
            <label>最大差异行数</label>
            <el-input-number v-model="settings.max_diff_lines" :min="100" :max="50000" :step="500" />
          </div>
        </div>
        <div class="form-actions">
          <ActionPill variant="primary" @click="saveSettings" :disabled="settingsSaving">{{ settingsSaving ? '保存中...' : '保存设置' }}</ActionPill>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'rules'" class="tab-content">
      <div class="section-header">
        <div class="section-info">
          <h3>规则管理</h3>
          <p class="section-desc">管理代码审查规则。规则按优先级顺序执行，关闭后审查时将跳过该规则。</p>
        </div>
        <div class="section-actions">
          <ActionPill variant="primary" :icon="Plus" @click="openAddRuleDialog">添加规则</ActionPill>
        </div>
      </div>
      <div v-if="rulesLoading" class="state-card">
        <LoadingState />
      </div>
      <div v-else-if="rules.length === 0" class="state-card">
        <EmptyState title="暂无审查规则" description="添加审查规则以启用代码审查检测" />
      </div>
      <div v-else class="rules-list">
        <div v-for="rule in rules" :key="rule.id" class="rule-card">
          <div class="rule-left">
            <el-switch v-model="rule.enabled" size="small" @change="handleRuleToggle(rule)" />
            <div class="rule-info">
              <span class="rule-name">
                {{ rule.name }}
                <span v-if="rule.rule_type === 'prompt'" class="rule-type-tag">Prompt</span>
              </span>
              <span class="rule-desc">{{ rule.description }}</span>
            </div>
          </div>
          <div class="rule-right">
            <span class="severity-pill" :class="'sev-' + rule.severity">{{ severityLabel(rule.severity) }}</span>
            <button class="act-btn act-btn--primary" @click="openEditRuleDialog(rule)">编辑</button>
            <button class="act-btn act-btn--danger" @click="handleDeleteRule(rule)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'prompt'" class="tab-content">
      <div class="section-header">
        <div class="section-info">
          <h3>Prompt 结构预览</h3>
          <p class="section-desc">系统内置的 Prompt 分为前缀、意图分析、后缀三部分，用户自定义规则插入在意图分析和后缀之间。以下为实际发送给 LLM 的完整结构。</p>
        </div>
        <div class="section-actions">
          <ActionPill variant="outline" @click="refreshPromptPreview">刷新预览</ActionPill>
        </div>
      </div>
      <div class="prompt-preview-list">
        <div class="prompt-block">
          <div class="prompt-block-header">
            <span class="prompt-block-badge badge-system">系统内置</span>
            <span class="prompt-block-title">1. 前缀 — 输出格式与行号规则</span>
          </div>
          <pre class="prompt-block-content">{{ promptPreview.prefix }}</pre>
        </div>
        <div class="prompt-block">
          <div class="prompt-block-header">
            <span class="prompt-block-badge badge-system">系统内置</span>
            <span class="prompt-block-title">2. 变更意图分析</span>
          </div>
          <pre class="prompt-block-content">{{ promptPreview.intent }}</pre>
        </div>
        <div class="prompt-block">
          <div class="prompt-block-header">
            <span class="prompt-block-badge badge-custom">用户自定义</span>
            <span class="prompt-block-title">3. 自定义审查规则（来自上方「规则管理」）</span>
          </div>
          <div v-if="promptPreview.customRules.length === 0" class="prompt-block-content prompt-empty">暂无启用的自定义规则，请在「规则管理」中添加</div>
          <pre v-else class="prompt-block-content">{{ promptPreview.customRules }}</pre>
        </div>
        <div class="prompt-block">
          <div class="prompt-block-header">
            <span class="prompt-block-badge badge-system">系统内置</span>
            <span class="prompt-block-title">4. 后缀 — 最终约束</span>
          </div>
          <pre class="prompt-block-content">{{ promptPreview.suffix }}</pre>
        </div>
      </div>
    </div>

    <el-dialog v-model="showRuleDialog" :title="editingRule ? '编辑审查规则' : '添加审查规则'" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="规则 ID">
          <el-input v-model="ruleForm.id" placeholder="如 secret, sql_injection" :disabled="!!editingRule" />
        </el-form-item>
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="如：密钥泄露检测" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="ruleForm.description" type="textarea" :rows="2" placeholder="规则描述" />
        </el-form-item>
        <el-form-item label="严重级别">
          <el-select v-model="ruleForm.severity" style="width:100%">
            <el-option label="严重 (critical)" value="critical" />
            <el-option label="高危 (high)" value="high" />
            <el-option label="中危 (medium)" value="medium" />
            <el-option label="低危 (low)" value="low" />
            <el-option label="信息 (info)" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="ruleForm.category" style="width:100%">
            <el-option label="安全" value="security" />
            <el-option label="质量" value="quality" />
            <el-option label="数据库" value="database" />
            <el-option label="风格" value="style" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-radio-group v-model="ruleForm.rule_type">
            <el-radio value="pattern">模式匹配</el-radio>
            <el-radio value="prompt">自定义 Prompt</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="ruleForm.rule_type === 'prompt'" label="Prompt 内容">
          <el-input v-model="ruleForm.prompt_text" type="textarea" :rows="5" placeholder="输入自定义 Prompt，将在 LLM 审查时注入 system prompt 中。例如：请特别关注以下方面的安全问题..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRuleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRule" :loading="ruleSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import { getCodeReviewSettings, updateCodeReviewSettings } from '@/api/modules/llm-settings'
import type { CodeReviewGlobalSettingsDTO } from '@/api/modules/llm-settings'
import { listReviewRules, createReviewRule, updateReviewRule, deleteReviewRule, getPromptStructure } from '@/api/modules/review-rules'
import type { ReviewRuleDTO, PromptStructureDTO } from '@/api/modules/review-rules'

const activeTab = ref('config')
const settingsSaving = ref(false)
const settings = ref<CodeReviewGlobalSettingsDTO>({
  enabled: true,
  auto_review_on_mr: true,
  block_on_high: true,
  max_files: 50,
  max_diff_lines: 3000,
  rag_enabled: false,
})

const rules = ref<ReviewRuleDTO[]>([])
const rulesLoading = ref(false)
const showRuleDialog = ref(false)
const editingRule = ref<ReviewRuleDTO | null>(null)
const ruleSaving = ref(false)
const ruleForm = ref({ id: '', name: '', description: '', severity: 'medium', category: 'other', enabled: true, sort_order: 0, rule_type: 'pattern', prompt_text: '' })

function severityLabel(s: string) {
  const m: Record<string, string> = { critical: '严重', high: '高危', medium: '中危', low: '低危', info: '信息' }
  return m[s] || s
}

async function loadSettings() {
  try { const data = await getCodeReviewSettings(); if (data) settings.value = data } catch { /* use defaults */ }
}

async function saveSettings() {
  settingsSaving.value = true
  try { await updateCodeReviewSettings(settings.value); ElMessage.success('设置已保存') }
  catch (e: any) { ElMessage.error('保存失败: ' + (e?.message || '')) }
  finally { settingsSaving.value = false }
}

async function loadRules(): Promise<void> {
  rulesLoading.value = true
  try { rules.value = await listReviewRules() || [] } catch { rules.value = [] }
  finally { rulesLoading.value = false }
}

function openAddRuleDialog() {
  editingRule.value = null
  ruleForm.value = { id: '', name: '', description: '', severity: 'medium', category: 'other', enabled: true, sort_order: 0, rule_type: 'pattern', prompt_text: '' }
  showRuleDialog.value = true
}

function openEditRuleDialog(rule: ReviewRuleDTO) {
  editingRule.value = rule
  ruleForm.value = { id: rule.id, name: rule.name, description: rule.description, severity: rule.severity, category: rule.category, enabled: rule.enabled, sort_order: rule.sort_order, rule_type: rule.rule_type || 'pattern', prompt_text: rule.prompt_text || '' }
  showRuleDialog.value = true
}

async function handleSaveRule() {
  const f = ruleForm.value
  if (!f.id || !f.name) { ElMessage.warning('请填写规则 ID 和名称'); return }
  ruleSaving.value = true
  try {
    if (editingRule.value) { await updateReviewRule(f.id, f) } else { await createReviewRule(f) }
    ElMessage.success(editingRule.value ? '更新成功' : '添加成功')
    showRuleDialog.value = false
    loadRules()
  } catch (e: any) { ElMessage.error((editingRule.value ? '更新' : '添加') + '失败: ' + (e?.message || '')) }
  finally { ruleSaving.value = false }
}

async function handleRuleToggle(rule: ReviewRuleDTO) {
  try { await updateReviewRule(rule.id, { enabled: rule.enabled }) }
  catch (e: any) { rule.enabled = !rule.enabled; ElMessage.error('更新失败: ' + (e?.message || '')) }
}

async function handleDeleteRule(rule: ReviewRuleDTO) {
  try { await ElMessageBox.confirm(`确定删除规则 "${rule.name}"？`, '确认删除', { type: 'warning' }) } catch { return }
  try { await deleteReviewRule(rule.id); ElMessage.success('已删除'); loadRules() }
  catch (e: any) { ElMessage.error('删除失败: ' + (e?.message || '')) }
}

const promptPreview = ref<{ prefix: string; intent: string; suffix: string; customRules: string }>({
  prefix: '', intent: '', suffix: '', customRules: '',
})

async function loadPromptPreview() {
  try {
    const data = await getPromptStructure()
    if (data) {
      promptPreview.value.prefix = data.prefix || ''
      promptPreview.value.intent = data.intent || ''
      promptPreview.value.suffix = data.suffix || ''
    }
    const enabledRules = rules.value.filter(r => r.enabled && r.rule_type === 'prompt')
    if (enabledRules.length > 0) {
      promptPreview.value.customRules = enabledRules.map((r, i) => `${i + 1}. [${r.name}] ${r.prompt_text}`).join('\n\n')
    } else {
      promptPreview.value.customRules = ''
    }
  } catch { /* ignore */ }
}

function refreshPromptPreview() {
  loadPromptPreview()
}

onMounted(() => { loadSettings(); loadRules().then(() => loadPromptPreview()) })
</script>

<style scoped>
.cr-settings-page { display: flex; flex-direction: column; gap: 20px; }
.tab-bar { display: flex; gap: 4px; border-bottom: 1px solid var(--border-color); }
.tab-btn {
  padding: 10px 20px; border: none; background: transparent; font-size: 14px;
  color: var(--text-color-secondary); cursor: pointer; border-bottom: 2px solid transparent;
  transition: all 0.2s; font-weight: 500;
}
.tab-btn.active { color: var(--accent-primary); border-bottom-color: var(--accent-primary); }
.tab-btn:hover { color: var(--accent-primary); }
.tab-content { display: flex; flex-direction: column; gap: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: flex-start; }
.section-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-color-primary); }
.section-desc { margin: 4px 0 0; font-size: 13px; color: var(--text-color-secondary); }
.section-actions { display: flex; align-items: center; gap: 12px; }
.settings-form-card {
  display: flex; flex-direction: column; gap: 20px; padding: 24px;
  border-radius: 12px; border: 1px solid var(--border-color);
  background: var(--bg-color-page);
}
.form-row { display: flex; justify-content: space-between; align-items: center; padding-bottom: 12px; border-bottom: 1px solid var(--border-color); }
.form-label-col { display: flex; flex-direction: column; gap: 2px; }
.form-label { font-size: 14px; font-weight: 500; color: var(--text-color-primary); }
.form-desc { font-size: 12px; color: var(--text-color-placeholder); }
.form-row-inline { display: flex; gap: 16px; }
.form-field { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.form-field label { font-size: 13px; font-weight: 500; color: var(--text-color-secondary); }
.form-actions { display: flex; justify-content: flex-end; padding-top: 12px; border-top: 1px solid var(--border-color); }
.state-card { border-radius: 12px; border: 1px solid var(--border-color); background: var(--bg-color-page); }
.rules-list { display: flex; flex-direction: column; gap: 8px; }
.rule-card {
  display: flex; justify-content: space-between; align-items: center; padding: 16px 20px;
  border-radius: 8px; border: 1px solid var(--border-color); background: var(--bg-color-page);
}
.rule-left { display: flex; align-items: center; gap: 12px; }
.rule-right { display: flex; align-items: center; gap: 8px; }
.rule-info { display: flex; flex-direction: column; gap: 2px; }
.rule-name { font-size: 14px; font-weight: 500; color: var(--text-color-primary); }
.rule-type-tag {
  margin-left: 6px; padding: 1px 6px; border-radius: 4px; font-size: 10px;
  background: #F5F3FF; color: #7C3AED; font-weight: 500;
}
.rule-desc { font-size: 12px; color: var(--text-color-placeholder); }
.severity-pill { display: inline-block; padding: 3px 10px; border-radius: 9999px; font-size: 11px; font-weight: 500; }
.sev-critical { background: #FEF2F2; color: #DC2626; }
.sev-high { background: #FFF7ED; color: #EA580C; }
.sev-medium { background: #FFFBEB; color: #D97706; }
.sev-low { background: #EFF6FF; color: #2563EB; }
.sev-info { background: #F3F4F6; color: #6B7280; }
.act-btn {
  display: inline-flex; align-items: center; gap: 4px; padding: 4px 10px;
  border-radius: 4px; font-size: 11px; cursor: pointer; transition: all 0.2s;
  border: 1px solid var(--border-color); background: transparent; color: var(--text-color-secondary);
}
.act-btn--primary { border-color: #6366F1; color: #6366F1; }
.act-btn--primary:hover { background: var(--accent-bg); }
.act-btn--danger { border-color: #EF4444; color: #EF4444; }
.act-btn--danger:hover { background: #FEF2F2; }
.prompt-preview-list { display: flex; flex-direction: column; gap: 16px; }
.prompt-block {
  border-radius: 8px; border: 1px solid var(--border-color); background: var(--bg-color-page); overflow: hidden;
}
.prompt-block-header {
  display: flex; align-items: center; gap: 10px; padding: 12px 16px;
  background: var(--bg-color-secondary, #F9FAFB); border-bottom: 1px solid var(--border-color);
}
.prompt-block-title { font-size: 13px; font-weight: 600; color: var(--text-color-primary); }
.prompt-block-badge {
  display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; letter-spacing: 0.5px;
}
.badge-system { background: #EFF6FF; color: #2563EB; }
.badge-custom { background: #F0FDF4; color: #16A34A; }
.prompt-block-content {
  margin: 0; padding: 16px; font-size: 12px; line-height: 1.7; color: #374151;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  white-space: pre-wrap; word-break: break-word;
}
.prompt-empty { color: var(--text-color-placeholder); font-family: inherit; font-style: italic; }
</style>
