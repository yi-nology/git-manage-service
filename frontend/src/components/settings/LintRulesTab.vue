<template>
  <div class="tab-section">
    <div class="section-header">
      <div class="section-info">
        <h3>Lint 规则</h3>
        <p class="section-desc">管理 Spec 文件的静态检查规则，启用/禁用规则或添加自定义规则。</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="openAddRuleDialog">添加规则</el-button>
    </div>

    <div v-if="lintLoading" style="padding: 40px 0; text-align: center">
      <el-icon class="is-loading" :size="20"><Loading /></el-icon>
      <p style="margin-top: 8px; color: var(--text-color-secondary)">加载规则中...</p>
    </div>

    <el-table v-else :data="lintRules" stripe border class="rules-table" empty-text="暂无 Lint 规则">
      <el-table-column label="启用" width="70" align="center">
        <template #default="{ row }">
          <el-switch v-model="row.enabled" @change="toggleRule(row)" />
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column label="类别" width="120" align="center">
        <template #default="{ row }">
          <el-tag :type="categoryTagType(row.category)" size="small">{{ categoryLabel(row.category) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="严重级别" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="severityTagType(row.severity)" size="small">{{ severityLabel(row.severity) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="pattern" label="模式" min-width="160" show-overflow-tooltip />
      <el-table-column label="操作" width="140" align="center">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEditRuleDialog(row)">编辑</el-button>
          <el-button link type="danger" size="small" @click="handleDeleteRule(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="ruleDialogVisible" :title="editingRule ? '编辑规则' : '添加规则'" width="560px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="ID" v-if="!editingRule">
          <el-input v-model="ruleForm.id" placeholder="唯一标识，如 my_custom_rule" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="ruleForm.name" placeholder="规则显示名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="ruleForm.description" type="textarea" :rows="2" placeholder="规则说明" />
        </el-form-item>
        <el-form-item label="类别">
          <el-select v-model="ruleForm.category" style="width: 100%">
            <el-option label="语法" value="syntax" />
            <el-option label="最佳实践" value="best_practice" />
            <el-option label="自定义" value="custom" />
            <el-option label="必选项" value="required" />
            <el-option label="风格" value="style" />
          </el-select>
        </el-form-item>
        <el-form-item label="严重级别">
          <el-select v-model="ruleForm.severity" style="width: 100%">
            <el-option label="错误" value="error" />
            <el-option label="警告" value="warning" />
            <el-option label="信息" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配模式">
          <el-input v-model="ruleForm.pattern" placeholder="正则表达式" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="ruleForm.priority" :min="0" :max="100" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSaving" @click="handleSaveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Loading } from '@element-plus/icons-vue'
import { getLintRules, createLintRule, updateLintRule } from '@/api/modules/spec'

const lintLoading = ref(false)
const lintRules = ref<any[]>([])
const ruleDialogVisible = ref(false)
const ruleSaving = ref(false)
const editingRule = ref<any>(null)
const ruleForm = reactive({
  id: '',
  name: '',
  description: '',
  category: 'custom' as const,
  severity: 'warning' as const,
  pattern: '',
  enabled: true,
  priority: 50,
})

function categoryLabel(c: string) {
  const m: Record<string, string> = { syntax: '语法', best_practice: '最佳实践', custom: '自定义', required: '必选项', style: '风格' }
  return m[c] || c
}

function categoryTagType(c: string): 'primary' | 'danger' | 'success' | 'info' | 'warning' | undefined {
  const m: Record<string, 'primary' | 'danger' | 'success' | 'info' | 'warning'> = { syntax: 'primary', best_practice: 'success', custom: 'info', required: 'danger', style: 'warning' }
  return m[c]
}

function severityLabel(s: string) {
  const m: Record<string, string> = { error: '错误', warning: '警告', info: '信息' }
  return m[s] || s
}

function severityTagType(s: string): 'primary' | 'danger' | 'success' | 'info' | 'warning' | undefined {
  const m: Record<string, 'primary' | 'danger' | 'success' | 'info' | 'warning'> = { error: 'danger', warning: 'warning', info: 'info' }
  return m[s]
}

async function loadLintRules() {
  lintLoading.value = true
  try {
    lintRules.value = await getLintRules() || []
  } catch { lintRules.value = [] }
  finally { lintLoading.value = false }
}

async function toggleRule(rule: any) {
  try {
    await updateLintRule(rule.id, { enabled: rule.enabled })
  } catch (e: any) {
    rule.enabled = !rule.enabled
    ElMessage.error('更新失败: ' + (e?.message || ''))
  }
}

function openAddRuleDialog() {
  editingRule.value = null
  Object.assign(ruleForm, { id: '', name: '', description: '', category: 'custom', severity: 'warning', pattern: '', enabled: true, priority: 50 })
  ruleDialogVisible.value = true
}

function openEditRuleDialog(rule: any) {
  editingRule.value = rule
  Object.assign(ruleForm, {
    id: rule.id,
    name: rule.name,
    description: rule.description || '',
    category: rule.category,
    severity: rule.severity,
    pattern: rule.pattern || '',
    enabled: rule.enabled,
    priority: rule.priority || 0,
  })
  ruleDialogVisible.value = true
}

async function handleSaveRule() {
  if (!ruleForm.name) {
    ElMessage.warning('请填写规则名称')
    return
  }
  if (!editingRule.value && !ruleForm.id) {
    ElMessage.warning('请填写规则 ID')
    return
  }
  ruleSaving.value = true
  try {
    if (editingRule.value) {
      await updateLintRule(editingRule.value.id, {
        name: ruleForm.name,
        description: ruleForm.description,
        category: ruleForm.category,
        severity: ruleForm.severity,
        pattern: ruleForm.pattern,
        enabled: ruleForm.enabled,
        priority: ruleForm.priority,
      })
      ElMessage.success('规则已更新')
    } else {
      await createLintRule({
        id: ruleForm.id,
        name: ruleForm.name,
        description: ruleForm.description,
        category: ruleForm.category,
        severity: ruleForm.severity,
        pattern: ruleForm.pattern,
        enabled: ruleForm.enabled,
        priority: ruleForm.priority,
      })
      ElMessage.success('规则已创建')
    }
    ruleDialogVisible.value = false
    loadLintRules()
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally { ruleSaving.value = false }
}

async function handleDeleteRule(rule: any) {
  try {
    await ElMessageBox.confirm(`确定删除规则 "${rule.name}"？`, '确认删除', { type: 'warning' })
  } catch { return }
  try {
    await updateLintRule(rule.id, { enabled: false })
    lintRules.value = lintRules.value.filter((r: any) => r.id !== rule.id)
    ElMessage.success('规则已禁用')
  } catch (e: any) {
    ElMessage.error('操作失败: ' + (e?.message || ''))
  }
}

onMounted(() => {
  loadLintRules()
})
</script>

<style scoped>
.tab-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-desc {
  font-size: 13px;
  color: var(--text-color-secondary);
  line-height: 1.6;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.rules-table :deep(.el-table__cell) {
  padding: 8px 0;
}
</style>
