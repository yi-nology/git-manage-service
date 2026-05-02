<template>
  <el-dialog v-model="visible" title="规则管理" width="700px" :close-on-click-modal="false">
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
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import { getLintRules, updateLintRule, createLintRule } from '@/api/modules/spec'
import type { LintRule } from '@/types/spec'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const rules = ref<LintRule[]>([])
const ruleSearch = ref('')
const showCreateRuleDialog = ref(false)
const creatingRule = ref(false)
const newRuleFormRef = ref()
const newRule = ref({ id: '', name: '', description: '', category: 'custom' as const, severity: 'warning' as const, pattern: '', enabled: true })

const filteredRules = computed(() => {
  if (!ruleSearch.value) return rules.value
  const s = ruleSearch.value.toLowerCase()
  return rules.value.filter(r => r.name.toLowerCase().includes(s) || r.description.toLowerCase().includes(s))
})

watch(() => props.modelValue, async (v) => { if (v) await loadRules() })

async function loadRules() {
  try { rules.value = await getLintRules() } catch { rules.value = [] }
}

async function handleToggleRule(rule: LintRule) {
  try { await updateLintRule(rule.id, { enabled: rule.enabled }) } catch { ElMessage.error('更新规则失败') }
}

async function handleCreateRule() {
  if (newRuleFormRef.value) {
    await newRuleFormRef.value.validate().catch(() => { throw new Error('validation failed') })
  }
  try {
    creatingRule.value = true
    await createLintRule({ ...newRule.value })
    ElMessage.success('规则创建成功')
    showCreateRuleDialog.value = false
    newRule.value = { id: '', name: '', description: '', category: 'custom' as const, severity: 'warning' as const, pattern: '', enabled: true }
    await loadRules()
  } catch { ElMessage.error('创建规则失败') }
  finally { creatingRule.value = false }
}
</script>

<style scoped>
.rule-manager { min-height: 400px; }
.rm-toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
</style>
