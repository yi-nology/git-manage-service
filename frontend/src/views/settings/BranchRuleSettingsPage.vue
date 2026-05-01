<template>
  <div class="branch-rule-page">
    <PageHeader title="分支规则设置" showBack backRoute="/settings" />

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'types' }" @click="activeTab = 'types'">分支类型</button>
      <button class="tab-btn" :class="{ active: activeTab === 'protected' }" @click="activeTab = 'protected'">保护分支</button>
      <button class="tab-btn" :class="{ active: activeTab === 'general' }" @click="activeTab = 'general'">全局开关</button>
    </div>

    <div v-show="activeTab === 'types'" class="tab-content">
      <div class="section-header">
        <div class="section-info">
          <h3>分支类型规则</h3>
          <p class="section-desc">定义分支命名前缀、任务 ID 要求和生命周期规则。新建分支时将自动校验。</p>
        </div>
        <div class="section-actions">
          <ActionPill variant="primary" :icon="Plus" @click="addRule">添加规则</ActionPill>
        </div>
      </div>

      <LoadingState v-if="loading" />

      <EmptyState v-else-if="rules.length === 0" title="暂无分支类型规则" description="添加规则以约束分支命名和生命周期">
        <template #action>
          <ActionPill variant="primary" :icon="Plus" @click="addRule">添加规则</ActionPill>
        </template>
      </EmptyState>

      <div v-else class="rules-list">
        <div v-for="(rule, idx) in rules" :key="idx" class="rule-card">
          <div class="rule-card-header">
            <div class="rule-prefix-row">
              <span class="rule-prefix-badge" :style="{ background: prefixColor(rule.prefix), color: '#fff' }">{{ rule.prefix || '/' }}</span>
              <input v-model="rule.display_name" class="rule-name-input" placeholder="显示名称" />
            </div>
            <div class="rule-actions">
              <button v-if="idx > 0" class="act-btn act-btn--default" @click="moveRule(idx, -1)">上移</button>
              <button v-if="idx < rules.length - 1" class="act-btn act-btn--default" @click="moveRule(idx, 1)">下移</button>
              <button class="act-btn act-btn--danger" @click="removeRule(idx)">删除</button>
            </div>
          </div>

          <div class="rule-card-body">
            <div class="rule-grid">
              <div class="rule-field">
                <label>前缀</label>
                <input v-model="rule.prefix" class="field-input" placeholder="feature/" />
              </div>
              <div class="rule-field">
                <label>任务 ID</label>
                <div class="switch-row">
                  <el-switch v-model="rule.require_task_id" />
                  <span class="toggle-label">{{ rule.require_task_id ? '必需' : '可选' }}</span>
                </div>
              </div>
              <div class="rule-field" v-if="rule.require_task_id">
                <label>任务 ID 正则</label>
                <input v-model="rule.task_id_pattern" class="field-input" placeholder="\d+" />
              </div>
              <div class="rule-field">
                <label>允许直接推送</label>
                <div class="switch-row">
                  <el-switch v-model="rule.allow_direct_push" />
                  <span class="toggle-label">{{ rule.allow_direct_push ? '允许' : '禁止' }}</span>
                </div>
              </div>
              <div class="rule-field">
                <label>需要代码审查</label>
                <div class="switch-row">
                  <el-switch v-model="rule.require_code_review" />
                  <span class="toggle-label">{{ rule.require_code_review ? '必需' : '可选' }}</span>
                </div>
              </div>
              <div class="rule-field">
                <label>合并后自动删除</label>
                <div class="switch-row">
                  <el-switch v-model="rule.auto_delete_on_merge" />
                  <span class="toggle-label">{{ rule.auto_delete_on_merge ? '自动' : '保留' }}</span>
                </div>
              </div>
            </div>

            <div class="rule-grid" style="margin-top:12px">
              <div class="rule-field">
                <label>源分支（可从此处创建）</label>
                <input v-model="rule.source_branches_str" class="field-input" placeholder="main, develop（逗号分隔）" />
              </div>
              <div class="rule-field">
                <label>目标分支（可合并到此处）</label>
                <input v-model="rule.target_branches_str" class="field-input" placeholder="main, develop（逗号分隔）" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="rules.length > 0" class="form-actions">
        <ActionPill variant="outline" @click="loadRules">取消</ActionPill>
        <ActionPill variant="primary" @click="saveRules" :disabled="saving">{{ saving ? '保存中...' : '保存规则' }}</ActionPill>
      </div>
    </div>

    <div v-show="activeTab === 'protected'" class="tab-content">
      <div class="section-header">
        <div class="section-info">
          <h3>保护分支</h3>
          <p class="section-desc">受保护的分支禁止直接推送和强制删除。建议将 main/master/release 等分支设为保护分支。</p>
        </div>
      </div>

      <div class="protected-card">
        <div class="protected-input-row">
          <input v-model="newProtectedBranch" class="field-input" placeholder="输入分支名，如 main、release/*" @keyup.enter="addProtectedBranch" />
          <ActionPill variant="primary" :icon="Plus" @click="addProtectedBranch" :disabled="!newProtectedBranch.trim()">添加</ActionPill>
        </div>
        <div class="protected-tags">
          <div v-for="(name, idx) in ruleSet.protected_branches" :key="idx" class="protected-tag">
            <span class="tag-name">{{ name }}</span>
            <button class="tag-remove" @click="removeProtectedBranch(idx)">&times;</button>
          </div>
          <div v-if="ruleSet.protected_branches.length === 0" class="empty-protected">
            暂无保护分支
          </div>
        </div>
      </div>

      <div v-if="ruleSet.protected_branches.length > 0" class="form-actions">
        <ActionPill variant="primary" @click="saveRules" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</ActionPill>
      </div>
    </div>

    <div v-show="activeTab === 'general'" class="tab-content">
      <div class="section-header">
        <div class="section-info">
          <h3>全局开关</h3>
          <p class="section-desc">全局启用/禁用分支规则。关闭后所有仓库创建分支时不再校验。</p>
        </div>
      </div>
      <div class="settings-form-card">
        <div class="form-row">
          <div class="form-label-col">
            <span class="form-label">启用分支规则</span>
            <span class="form-desc">开启后，创建分支时将按规则校验分支名和源分支</span>
          </div>
          <el-switch v-model="ruleSet.enabled" />
        </div>
        <div class="form-actions">
          <ActionPill variant="primary" @click="saveRules" :disabled="saving">{{ saving ? '保存中...' : '保存设置' }}</ActionPill>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getBranchRules, updateBranchRules } from '@/api/modules/branch-rule'
import type { BranchRuleDTO, BranchRuleSetDTO } from '@/api/modules/branch-rule'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'

const activeTab = ref('types')
const loading = ref(false)
const saving = ref(false)
const newProtectedBranch = ref('')

const ruleSet = ref<BranchRuleSetDTO>({
  enabled: true,
  rules: [],
  protected_branches: [],
})

interface EditableRule extends BranchRuleDTO {
  source_branches_str: string
  target_branches_str: string
}

const rules = ref<EditableRule[]>([])

const PREFIX_COLORS: Record<string, string> = {
  'feature/': '#6366F1',
  'bugfix/': '#EF4444',
  'hotfix/': '#F59E0B',
  'release/': '#10B981',
  'support/': '#3B82F6',
  'chore/': '#8B5CF6',
}

function prefixColor(prefix: string) {
  return PREFIX_COLORS[prefix] || '#6B7280'
}

function ruleToEditable(r: BranchRuleDTO): EditableRule {
  return {
    ...r,
    source_branches_str: (r.source_branches || []).join(', '),
    target_branches_str: (r.target_branches || []).join(', '),
  }
}

function editableToRule(r: EditableRule): BranchRuleDTO {
  return {
    id: r.id,
    prefix: r.prefix,
    display_name: r.display_name,
    source_branches: r.source_branches_str.split(',').map(s => s.trim()).filter(Boolean),
    target_branches: r.target_branches_str.split(',').map(s => s.trim()).filter(Boolean),
    require_task_id: r.require_task_id,
    task_id_pattern: r.task_id_pattern,
    auto_delete_on_merge: r.auto_delete_on_merge,
    allow_direct_push: r.allow_direct_push,
    require_code_review: r.require_code_review,
    sort_order: r.sort_order,
  }
}

async function loadRules() {
  loading.value = true
  try {
    const data = await getBranchRules()
    if (data) {
      ruleSet.value = data
      rules.value = (data.rules || []).map(ruleToEditable)
    }
  } catch { /* use defaults */ }
  finally { loading.value = false }
}

function addRule() {
  rules.value.push({
    id: 0,
    prefix: '',
    display_name: '',
    source_branches: [],
    target_branches: [],
    require_task_id: false,
    task_id_pattern: '',
    auto_delete_on_merge: false,
    allow_direct_push: true,
    require_code_review: false,
    sort_order: rules.value.length,
    source_branches_str: '',
    target_branches_str: '',
  })
}

function removeRule(idx: number) {
  rules.value.splice(idx, 1)
}

function moveRule(idx: number, dir: number) {
  const target = idx + dir
  if (target < 0 || target >= rules.value.length) return
  const arr = [...rules.value]
  const temp = arr[idx]!
  arr[idx] = arr[target]!
  arr[target] = temp
  rules.value = arr
}

function addProtectedBranch() {
  const name = newProtectedBranch.value.trim()
  if (!name) return
  if (ruleSet.value.protected_branches.includes(name)) {
    ElMessage.warning('该分支已存在')
    return
  }
  ruleSet.value.protected_branches.push(name)
  newProtectedBranch.value = ''
}

function removeProtectedBranch(idx: number) {
  ruleSet.value.protected_branches.splice(idx, 1)
}

async function saveRules() {
  saving.value = true
  try {
    const payload: BranchRuleSetDTO = {
      enabled: ruleSet.value.enabled,
      rules: rules.value.map(editableToRule),
      protected_branches: ruleSet.value.protected_branches,
    }
    const data = await updateBranchRules(payload)
    if (data) {
      ruleSet.value = data
      rules.value = (data.rules || []).map(ruleToEditable)
    }
    ElMessage.success('规则已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally {
    saving.value = false
  }
}

onMounted(loadRules)
</script>

<style scoped>
.branch-rule-page { display: flex; flex-direction: column; gap: 20px; }

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

.rules-list { display: flex; flex-direction: column; gap: 12px; }

.rule-card {
  border-radius: 12px; border: 1px solid var(--border-color);
  background: var(--bg-color-page); overflow: hidden;
}
.rule-card-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px 20px; background: var(--surface-card); border-bottom: 1px solid var(--border-color);
}
.rule-prefix-row { display: flex; align-items: center; gap: 10px; }
.rule-prefix-badge {
  display: inline-flex; align-items: center; padding: 4px 12px;
  border-radius: 6px; font-size: 13px; font-weight: 600; font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}
.rule-name-input {
  border: none; background: transparent; font-size: 14px; font-weight: 500;
  color: var(--text-color-primary); outline: none; width: 200px;
}
.rule-name-input::placeholder { color: var(--text-color-placeholder); }
.rule-actions { display: flex; gap: 6px; }

.act-btn {
  display: inline-flex; align-items: center; gap: 4px; padding: 4px 10px;
  border-radius: 4px; font-size: 11px; cursor: pointer; transition: all 0.2s;
  border: 1px solid var(--border-color); background: transparent; color: var(--text-color-secondary);
}
.act-btn--default:hover { background: var(--surface-card); }
.act-btn--danger { border-color: #EF4444; color: #EF4444; }
.act-btn--danger:hover { background: #FEF2F2; }

.rule-card-body { padding: 16px 20px; }

.rule-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px 20px; }
.rule-field { display: flex; flex-direction: column; gap: 4px; }
.rule-field label { font-size: 12px; font-weight: 500; color: var(--text-color-secondary); }

.field-input {
  border: 1px solid var(--border-color); border-radius: 6px;
  padding: 8px 12px; font-size: 13px; color: var(--text-color-primary);
  background: var(--bg-color-page); outline: none; transition: border-color 0.2s;
  width: 100%; box-sizing: border-box;
}
.field-input:focus { border-color: var(--accent-primary); }
.field-input::placeholder { color: var(--text-color-placeholder, #9ca3af); }

.switch-row { display: flex; align-items: center; gap: 8px; height: 36px; }
.toggle-label { font-size: 12px; color: var(--text-color-secondary); }

.protected-card {
  padding: 20px; border-radius: 12px; border: 1px solid var(--border-color);
  background: var(--bg-color-page); display: flex; flex-direction: column; gap: 16px;
}
.protected-input-row { display: flex; gap: 10px; }
.protected-input-row .field-input { flex: 1; }

.protected-tags { display: flex; flex-wrap: wrap; gap: 8px; min-height: 36px; }
.protected-tag {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 12px; border-radius: 6px; background: #FEF2F2;
  border: 1px solid #FECACA; font-size: 13px; color: #DC2626; font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}
.tag-remove {
  border: none; background: none; color: #DC2626; cursor: pointer;
  font-size: 16px; padding: 0; line-height: 1; opacity: 0.6; transition: opacity 0.2s;
}
.tag-remove:hover { opacity: 1; }
.empty-protected { font-size: 13px; color: var(--text-color-placeholder); padding: 8px 0; }

.settings-form-card {
  display: flex; flex-direction: column; gap: 20px; padding: 24px;
  border-radius: 12px; border: 1px solid var(--border-color);
  background: var(--bg-color-page);
}
.form-row {
  display: flex; justify-content: space-between; align-items: center;
  padding-bottom: 12px; border-bottom: 1px solid var(--border-color);
}
.form-label-col { display: flex; flex-direction: column; gap: 2px; }
.form-label { font-size: 14px; font-weight: 500; color: var(--text-color-primary); }
.form-desc { font-size: 12px; color: var(--text-color-placeholder); }

.form-actions { display: flex; justify-content: flex-end; gap: 10px; padding-top: 4px; }
</style>
