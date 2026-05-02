<template>
  <div>
    <div class="content-header">
      <SectionTitle title="分支规则配置" />
      <div class="content-actions">
        <ActionPill variant="outline" :icon="Refresh" @click="loadBranchRules">刷新</ActionPill>
      </div>
    </div>

    <LoadingState v-if="brLoading" />

    <div v-else class="config-panel">
      <div class="config-sidebar">
        <button class="cfg-nav-btn active">分支规则</button>
      </div>

      <div class="config-form-area">
        <div class="form-section">
          <div class="form-row">
            <div class="form-label">
              <span>使用自定义规则</span>
              <span class="form-desc">开启后将覆盖全局分支规则，仅对此远端仓库生效</span>
            </div>
            <el-switch v-model="branchRuleCfg.use_custom_rules" />
          </div>

          <template v-if="branchRuleCfg.use_custom_rules">
            <div class="form-section" style="margin-top:12px">
              <h4 style="margin:0 0 12px;font-size:14px;color:var(--text-color-primary)">分支类型规则</h4>
              <div v-for="(rule, idx) in branchRuleCfg.rules" :key="idx" class="br-rule-card">
                <div class="br-rule-header">
                  <span class="br-rule-prefix">{{ rule.prefix || '/' }}</span>
                  <input v-model="rule.display_name" class="br-rule-name-input" placeholder="显示名称" />
                  <ActionPill variant="danger" small @click="branchRuleCfg.rules.splice(idx, 1)">删除</ActionPill>
                </div>
                <div class="br-rule-grid">
                  <div class="form-field">
                    <label>前缀</label>
                    <input v-model="rule.prefix" class="field-input" placeholder="feature/" />
                  </div>
                  <div class="form-field">
                    <label>任务 ID</label>
                    <div class="switch-row">
                      <el-switch v-model="rule.require_task_id" />
                      <span class="toggle-label-sm">{{ rule.require_task_id ? '必需' : '可选' }}</span>
                    </div>
                  </div>
                  <div class="form-field">
                    <label>允许直接推送</label>
                    <div class="switch-row">
                      <el-switch v-model="rule.allow_direct_push" />
                      <span class="toggle-label-sm">{{ rule.allow_direct_push ? '允许' : '禁止' }}</span>
                    </div>
                  </div>
                  <div class="form-field">
                    <label>需要代码审查</label>
                    <div class="switch-row">
                      <el-switch v-model="rule.require_code_review" />
                      <span class="toggle-label-sm">{{ rule.require_code_review ? '必需' : '可选' }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <ActionPill variant="outline" :icon="Plus" @click="branchRuleCfg.rules.push({ id:0, prefix:'', display_name:'', source_branches:[], target_branches:[], require_task_id:false, task_id_pattern:'', auto_delete_on_merge:false, allow_direct_push:true, require_code_review:false, sort_order:branchRuleCfg.rules.length })" style="margin-top:8px">添加规则</ActionPill>
            </div>

            <div class="form-section" style="margin-top:16px">
              <h4 style="margin:0 0 8px;font-size:14px;color:var(--text-color-primary)">保护分支</h4>
              <div class="protected-tags">
                <div v-for="(name, idx) in branchRuleCfg.protected_branches" :key="idx" class="protected-tag-sm">
                  <span>{{ name }}</span>
                  <button class="tag-remove" @click="branchRuleCfg.protected_branches.splice(idx, 1)">&times;</button>
                </div>
                <span v-if="branchRuleCfg.protected_branches.length === 0" class="text-muted">暂无</span>
              </div>
            </div>
          </template>
        </div>

        <div v-if="branchRuleCfg.linked_repos && branchRuleCfg.linked_repos.length > 0" class="scope-card">
          <h4>生效范围</h4>
          <p class="scope-desc">以下本地仓库将使用此远端仓库的分支规则配置：</p>
          <div class="scope-repos">
            <div v-for="r in branchRuleCfg.linked_repos" :key="r.id" class="scope-repo-item">
              <el-icon :size="14" style="color:#6366F1"><FolderOpened /></el-icon>
              <span class="scope-repo-name">{{ r.name }}</span>
              <span class="scope-repo-key">{{ r.key }}</span>
            </div>
          </div>
        </div>

        <div class="form-actions">
          <ActionPill variant="outline" @click="loadBranchRules">取消</ActionPill>
          <ActionPill variant="primary" @click="saveBranchRules" :disabled="brSaving">{{ brSaving ? '保存中...' : '保存' }}</ActionPill>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, FolderOpened } from '@element-plus/icons-vue'
import { getRemoteRepoBranchRules, updateRemoteRepoBranchRules } from '@/api/modules/branch-rule'
import type { BranchRuleDTO } from '@/api/modules/branch-rule'
import SectionTitle from '@/components/common/SectionTitle.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import LoadingState from '@/components/common/LoadingState.vue'

const props = defineProps<{
  active: boolean
  providerId: number
  repoOwner: string
  repoName: string
}>()

const brLoading = ref(false)
const brSaving = ref(false)
const loaded = ref(false)
const branchRuleCfg = ref<{
  use_custom_rules: boolean
  rules: BranchRuleDTO[]
  protected_branches: string[]
  linked_repos: { id: number; key: string; name: string }[]
}>({
  use_custom_rules: false,
  rules: [],
  protected_branches: [],
  linked_repos: [],
})

async function loadBranchRules() {
  brLoading.value = true
  try {
    const res = await getRemoteRepoBranchRules(props.providerId, props.repoOwner, props.repoName)
    if (res) {
      branchRuleCfg.value = {
        use_custom_rules: res.use_custom_rules,
        rules: res.rules || [],
        protected_branches: res.protected_branches || [],
        linked_repos: res.linked_repos || [],
      }
    }
  } catch { /* use defaults */ }
  finally { brLoading.value = false }
}

async function saveBranchRules() {
  brSaving.value = true
  try {
    const res = await updateRemoteRepoBranchRules(props.providerId, props.repoOwner, props.repoName, {
      use_custom_rules: branchRuleCfg.value.use_custom_rules,
      rules: branchRuleCfg.value.rules,
      protected_branches: branchRuleCfg.value.protected_branches,
    })
    if (res) {
      branchRuleCfg.value = {
        use_custom_rules: res.use_custom_rules,
        rules: res.rules || [],
        protected_branches: res.protected_branches || [],
        linked_repos: res.linked_repos || [],
      }
    }
    ElMessage.success('分支规则已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally {
    brSaving.value = false
  }
}

watch(() => props.active, (val) => {
  if (val && !loaded.value) {
    loadBranchRules()
    loaded.value = true
  }
}, { immediate: true })
</script>

<style scoped>
.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.content-actions {
  display: flex;
  gap: 8px;
}
.config-panel {
  display: flex;
  gap: 20px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}
.config-sidebar {
  width: 180px;
  border-right: 1px solid var(--border-color);
  padding: 16px 0;
  flex-shrink: 0;
}
.cfg-nav-btn {
  display: block;
  width: 100%;
  padding: 10px 20px;
  border: none;
  background: transparent;
  text-align: left;
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  border-left: 3px solid transparent;
  transition: all 0.2s;
}
.cfg-nav-btn.active {
  color: var(--accent-primary);
  border-left-color: var(--accent-primary);
  background: var(--accent-bg);
  font-weight: 500;
}
.cfg-nav-btn:hover { color: var(--accent-primary); }
.config-form-area {
  flex: 1;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.form-section { display: flex; flex-direction: column; gap: 16px; }
.form-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}
.form-label { display: flex; flex-direction: column; gap: 2px; }
.form-label span:first-child { font-size: 14px; font-weight: 500; color: var(--text-color-primary); }
.form-desc { font-size: 12px; color: var(--text-color-placeholder); }
.form-field { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.form-field label { font-size: 13px; font-weight: 500; color: var(--text-color-secondary); }
.scope-card {
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #EEF2FF;
  background: #F8F9FF;
}
.scope-card h4 { margin: 0 0 4px 0; font-size: 14px; color: var(--text-color-primary); }
.scope-desc { margin: 0 0 12px 0; font-size: 12px; color: var(--text-color-placeholder); }
.scope-repos { display: flex; flex-direction: column; gap: 8px; }
.scope-repo-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid var(--border-color);
}
.scope-repo-name { font-size: 13px; font-weight: 500; color: var(--text-color-primary); }
.scope-repo-key { font-size: 11px; color: var(--text-color-placeholder); font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; }
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}
.br-rule-card {
  padding: 12px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  margin-bottom: 8px;
}
.br-rule-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.br-rule-prefix {
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--accent-bg);
  color: #6366F1;
  font-size: 12px;
  font-weight: 600;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}
.br-rule-name-input {
  border: none;
  background: transparent;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
  outline: none;
  flex: 1;
}
.br-rule-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;
}
.switch-row {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 32px;
}
.toggle-label-sm { font-size: 11px; color: var(--text-color-secondary); }
.field-input {
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 6px 8px;
  font-size: 13px;
  outline: none;
  background: var(--bg-color-page);
  color: var(--text-color-primary);
}
.field-input:focus {
  border-color: var(--accent-primary);
}
.protected-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.protected-tag-sm {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 4px;
  background: #FEF2F2;
  border: 1px solid #FECACA;
  font-size: 12px;
  color: #DC2626;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}
.tag-remove {
  border: none;
  background: none;
  color: #DC2626;
  cursor: pointer;
  font-size: 14px;
  padding: 0;
  line-height: 1;
  opacity: 0.6;
}
.tag-remove:hover { opacity: 1; }
.text-muted { font-size: 12px; color: var(--text-color-placeholder); }
</style>
