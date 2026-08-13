<template>
  <div class="author-settings-page">
    <PageHeader title="Git 作者管理" subtitle="管理提交身份和别名" show-back back-route="/settings" />

    <div class="content-area">
      <div v-if="activeIdentity" class="active-card">
        <div class="active-header">
          <SectionTitle title="当前激活身份" />
          <StatusBadge variant="success" text="已激活" />
        </div>
        <div class="active-body">
          <div class="avatar" :style="{ background: avatarColor(activeIdentity.canonical_name) }">
            {{ avatarText(activeIdentity.canonical_name) }}
          </div>
          <div class="active-info">
            <div class="active-name">{{ activeIdentity.canonical_name }}</div>
            <div class="active-email">{{ activeIdentity.canonical_email }}</div>
          </div>
        </div>
        <div v-if="activeIdentity.aliases.length > 0" class="active-aliases">
          <span class="alias-label">别名:</span>
          <el-tag v-for="(a, i) in activeIdentity.aliases" :key="i" size="small" type="info" class="alias-tag">
            {{ a.name }} &lt;{{ a.email }}&gt;
          </el-tag>
        </div>
      </div>

      <div class="section-header">
        <div class="section-info">
          <SectionTitle title="所有身份" />
          <span v-if="identities.length > 0" class="count-tag">{{ identities.length }} 个身份</span>
        </div>
        <div class="section-actions">
          <ActionPill variant="outline" :disabled="aiLoading" @click="doSmartSuggest">AI 推荐别名</ActionPill>
          <ActionPill variant="outline" :disabled="aiLoading" @click="doSuggestMerges">AI 分析合并</ActionPill>
          <ActionPill variant="primary" :icon="Plus" @click="openCreateDialog">新建身份</ActionPill>
        </div>
      </div>

      <div v-if="loadingState" class="state-card"><LoadingState /></div>
      <div v-else-if="identities.length === 0" class="state-card">
        <EmptyState title="暂无身份" description="创建 Git 作者身份，用于统一提交记录和历史修复">
          <template #action>
            <ActionPill variant="primary" :icon="Plus" @click="openCreateDialog">创建身份</ActionPill>
          </template>
        </EmptyState>
      </div>
      <DataTable v-else :columns="identityColumns" :data="identities" row-key="id">
        <template #cell-name="{ row }">
          <div class="cell-identity">
            <div class="cell-avatar" :style="{ background: avatarColor(row.canonical_name) }">
              {{ avatarText(row.canonical_name) }}
            </div>
            <div class="cell-info">
              <div class="cell-name">{{ row.canonical_name }}</div>
              <div class="cell-email">{{ row.canonical_email }}</div>
            </div>
          </div>
        </template>
        <template #cell-aliases="{ row }">
          <span class="cell-aliases-count">{{ row.aliases.length }} 个别名</span>
        </template>
        <template #cell-status="{ row }">
          <StatusBadge v-if="row.is_default" variant="success" text="已激活" :show-dot="false" />
          <StatusBadge v-else variant="default" text="未激活" :show-dot="false" />
        </template>
        <template #row-actions="{ row }">
          <button v-if="!row.is_default" class="act-btn act-btn--green" :disabled="activatingId === row.id" @click="handleActivate(row.id)">
            {{ activatingId === row.id ? '激活中...' : '激活' }}
          </button>
          <button class="act-btn act-btn--primary" @click="openEditDialog(row)" aria-label="编辑身份">编辑</button>
          <button v-if="!row.is_default" class="act-btn act-btn--danger" @click="handleDelete(row)" aria-label="删除身份">删除</button>
        </template>
      </DataTable>

      <div v-if="aiSuggestion" class="ai-result-card">
        <div class="ai-result-header">
          <SectionTitle title="AI 别名推荐" />
          <el-button text size="small" @click="aiSuggestion = null">关闭</el-button>
        </div>
        <p class="ai-result-summary">{{ aiSuggestion.summary }}</p>
        <div v-if="aiSuggestion.suggestions.length > 0" class="ai-suggest-list">
          <div v-for="(s, i) in aiSuggestion.suggestions" :key="i" class="ai-suggest-item">
            <div class="ai-suggest-info">
              <span class="ai-suggest-arrow">{{ s.identity_name }} ← <strong>{{ s.alias_name }}</strong> &lt;{{ s.alias_email }}&gt;</span>
              <el-tag :type="s.confidence === 'high' ? 'success' : s.confidence === 'medium' ? 'warning' : 'info'" size="small">{{ s.confidence }}</el-tag>
            </div>
            <div class="ai-suggest-reason">{{ s.reason }}</div>
            <el-button size="small" type="primary" @click="adoptSuggestion(s)">采纳</el-button>
          </div>
        </div>
        <el-empty v-else description="AI 未发现新的别名推荐" :image-size="40" />
      </div>

      <div v-if="aiMerge" class="ai-result-card">
        <div class="ai-result-header">
          <SectionTitle title="AI 身份合并建议" />
          <el-button text size="small" @click="aiMerge = null">关闭</el-button>
        </div>
        <p class="ai-result-summary">{{ aiMerge.summary }}</p>
        <div v-if="aiMerge.merges.length > 0" class="ai-suggest-list">
          <div v-for="(m, i) in aiMerge.merges" :key="i" class="ai-suggest-item">
            <div class="ai-suggest-info">
              <span class="ai-suggest-arrow">保留 <strong>{{ m.keep_name }}</strong> ← 合并 {{ m.merge_names }}</span>
            </div>
            <div class="ai-suggest-reason">{{ m.reason }}</div>
            <el-button size="small" type="primary" @click="executeMerge(m)">执行合并</el-button>
          </div>
        </div>
        <el-empty v-else description="AI 认为当前身份无需合并" :image-size="40" />
      </div>

      <div class="hint-card">
        <el-icon :size="16"><InfoFilled /></el-icon>
        <span>激活身份后，新提交将使用该身份的名称和邮箱，同时更新全局 ~/.gitconfig。别名用于识别历史提交中属于你的记录。</span>
      </div>
    </div>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑身份' : '新建身份'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="dialogForm" :rules="formRules" label-width="100px" @submit.prevent="handleSaveDialog">
        <el-form-item label="主名" prop="canonical_name">
          <el-input v-model="dialogForm.canonical_name" placeholder="如 John Doe" />
        </el-form-item>
        <el-form-item label="主邮箱" prop="canonical_email">
          <el-input v-model="dialogForm.canonical_email" placeholder="如 john@example.com" />
        </el-form-item>
        <el-form-item label="别名">
          <div class="alias-list">
            <div v-for="(a, i) in dialogForm.aliases" :key="i" class="alias-row">
              <el-input v-model="a.name" size="small" placeholder="名称" style="width: 140px" />
              <el-input v-model="a.email" size="small" placeholder="邮箱" style="flex: 1" />
              <el-button size="small" type="danger" :icon="Delete" circle aria-label="删除别名" @click="dialogForm.aliases.splice(i, 1)" />
            </div>
            <el-button size="small" @click="dialogForm.aliases.push({ name: '', email: '' })">+ 添加别名</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveDialog" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Delete, InfoFilled } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import { useAuthorIdentity, useAuthorAI } from '@/composables/useAuthor'
import type { AuthorIdentityDTO, AliasEntry, AliasSuggestion, MergeCandidate } from '@/api/modules/author'

const { identities, loading: loadingState, loadIdentities, handleCreate, handleUpdate, handleDelete: doDelete, handleActivate: doActivate } = useAuthorIdentity()
const { aiLoading, aiSuggestion, aiMerge, smartSuggest, suggestMerges } = useAuthorAI('')

const activeIdentity = computed(() => identities.value.find(i => i.is_default) || null)
const activatingId = ref<number | null>(null)

const identityColumns = [
  { key: 'name', label: '身份', flex: 3 },
  { key: 'aliases', label: '别名', width: '100px' },
  { key: 'status', label: '状态', width: '100px' },
]

function avatarText(name: string) {
  return name.substring(0, 2).toUpperCase()
}

function avatarColor(name: string) {
  const colors = ['#6366F1', '#8B5CF6', '#EC4899', '#F59E0B', '#10B981', '#3B82F6', '#EF4444', '#14B8A6']
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}

const showDialog = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const dialogForm = ref<{ canonical_name: string; canonical_email: string; aliases: AliasEntry[] }>({
  canonical_name: '',
  canonical_email: '',
  aliases: [],
})

const emailRe = /^[^@\s]+@[^@\s]+\.[^@\s]+$/

const formRules: FormRules = {
  canonical_name: [
    { required: true, message: '名称不能为空', trigger: 'blur' },
    { max: 100, message: '名称不能超过100个字符', trigger: 'blur' },
  ],
  canonical_email: [
    { required: true, message: '邮箱不能为空', trigger: 'blur' },
    { pattern: emailRe, message: '邮箱格式不正确', trigger: 'blur' },
  ],
}

function openCreateDialog() {
  editingId.value = null
  dialogForm.value = { canonical_name: '', canonical_email: '', aliases: [] }
  showDialog.value = true
}

function openEditDialog(item: AuthorIdentityDTO) {
  editingId.value = item.id
  dialogForm.value = {
    canonical_name: item.canonical_name,
    canonical_email: item.canonical_email,
    aliases: item.aliases.map(a => ({ ...a })),
  }
  showDialog.value = true
}

async function handleSaveDialog() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch { return }
  for (let i = 0; i < dialogForm.value.aliases.length; i++) {
    const a = dialogForm.value.aliases[i]!
    if (!a.name.trim() || !a.email.trim()) {
      ElMessage.warning(`别名 #${i + 1} 的名称和邮箱不能为空`)
      return
    }
    if (!emailRe.test(a.email)) {
      ElMessage.warning(`别名 #${i + 1} 的邮箱格式不正确: ${a.email}`)
      return
    }
  }
  saving.value = true
  try {
    if (editingId.value) {
      await handleUpdate(editingId.value, dialogForm.value)
    } else {
      await handleCreate(dialogForm.value)
    }
    showDialog.value = false
  } finally {
    saving.value = false
  }
}

async function handleDelete(item: AuthorIdentityDTO) {
  try {
    await ElMessageBox.confirm(
      `确认删除身份「${item.canonical_name}」？关联的仓库将恢复使用全局默认。`,
      '删除身份',
      { confirmButtonText: '确认删除', cancelButtonText: '取消', type: 'warning' }
    )
  } catch { return }
  await doDelete(item.id)
}

async function handleActivate(id: number) {
  try {
    await ElMessageBox.confirm(
      '激活此身份将更新全局 ~/.gitconfig 配置，所有新提交将使用此身份。确认激活？',
      '确认激活',
      { confirmButtonText: '确认激活', cancelButtonText: '取消', type: 'info' }
    )
  } catch { return }
  activatingId.value = id
  try {
    await doActivate(id)
  } finally {
    activatingId.value = null
  }
}

async function doSmartSuggest() {
  if (identities.value.length === 0) {
    ElMessage.warning('请先创建至少一个身份')
    return
  }
  await smartSuggest()
}

async function doSuggestMerges() {
  if (identities.value.length < 2) {
    ElMessage.warning('至少需要两个身份才能分析合并')
    return
  }
  await suggestMerges()
}

async function adoptSuggestion(s: AliasSuggestion) {
  const identity = identities.value.find(i => i.id === s.identity_id)
  if (!identity) {
    ElMessage.error('找不到目标身份')
    return
  }
  const newAliases = [...identity.aliases, { name: s.alias_name, email: s.alias_email }]
  await handleUpdate(identity.id, {
    canonical_name: identity.canonical_name,
    canonical_email: identity.canonical_email,
    aliases: newAliases,
  })
  aiSuggestion.value = null
}

async function executeMerge(m: MergeCandidate) {
  try {
    await ElMessageBox.confirm(
      `将 "${m.merge_names}" 的所有别名合并到 "${m.keep_name}"，并删除被合并的身份？`,
      '确认合并',
      { confirmButtonText: '确认合并', cancelButtonText: '取消', type: 'warning' }
    )
  } catch { return }
  const keepIdentity = identities.value.find(i => i.id === m.keep_id)
  if (!keepIdentity) return
  const mergeAliases: AliasEntry[] = []
  for (const mid of m.merge_ids) {
    const mi = identities.value.find(i => i.id === mid)
    if (mi) {
      mergeAliases.push({ name: mi.canonical_name, email: mi.canonical_email })
      mergeAliases.push(...mi.aliases)
      await doDelete(mid)
    }
  }
  await handleUpdate(keepIdentity.id, {
    canonical_name: keepIdentity.canonical_name,
    canonical_email: keepIdentity.canonical_email,
    aliases: [...keepIdentity.aliases, ...mergeAliases],
  })
  aiMerge.value = null
}

onMounted(() => {
  loadIdentities()
})
</script>

<style scoped>
.author-settings-page { display: flex; flex-direction: column; gap: var(--spacing-lg); }
.content-area { display: flex; flex-direction: column; gap: var(--spacing-lg); padding: 0 0 var(--spacing-xl); }

.active-card { display: flex; flex-direction: column; gap: var(--spacing-md); padding: var(--spacing-lg); background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); }
.active-header { display: flex; justify-content: space-between; align-items: center; }
.active-body { display: flex; gap: var(--spacing-md); align-items: center; }
.avatar { width: 56px; height: 56px; border-radius: 28px; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 600; flex-shrink: 0; }
.active-info { display: flex; flex-direction: column; gap: var(--spacing-xs); }
.active-name { font-size: var(--font-size-lg); font-weight: 600; color: var(--text-color-primary); }
.active-email { font-size: var(--font-size-sm); color: var(--text-color-secondary); }
.active-aliases { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
.alias-label { font-size: var(--font-size-xs); color: var(--text-color-secondary); font-weight: 500; }
.alias-tag { font-size: var(--font-size-xs); max-width: 240px; overflow: hidden; text-overflow: ellipsis; }

.section-header { display: flex; justify-content: space-between; align-items: center; }
.section-info { display: flex; gap: var(--spacing-sm); align-items: center; }
.section-actions { display: flex; gap: var(--spacing-sm); align-items: center; }
.count-tag { font-size: var(--font-size-xs); color: var(--text-color-secondary); }
.state-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); }

.cell-identity { display: flex; gap: var(--spacing-sm); align-items: center; }
.cell-avatar { width: 32px; height: 32px; border-radius: 16px; color: #fff; display: flex; align-items: center; justify-content: center; font-size: var(--font-size-xs); font-weight: 600; flex-shrink: 0; }
.cell-info { display: flex; flex-direction: column; gap: 2px; }
.cell-name { font-size: var(--font-size-sm); font-weight: 500; color: var(--text-color-primary); }
.cell-email { font-size: var(--font-size-xs); color: var(--text-color-secondary); font-family: var(--font-family-mono, monospace); }
.cell-aliases-count { font-size: var(--font-size-sm); color: var(--text-color-secondary); }

.act-btn { display: inline-flex; align-items: center; gap: 4px; padding: 4px 10px; border-radius: var(--border-radius-sm); border: 1px solid transparent; font-size: var(--font-size-xs); font-weight: 500; cursor: pointer; transition: all var(--transition-fast); background: transparent; }
.act-btn--primary { color: var(--primary-color); border-color: var(--primary-color); }
.act-btn--primary:hover { background: var(--accent-bg); }
.act-btn--green { color: var(--success-color); border-color: var(--success-color); }
.act-btn--green:hover { background: var(--surface-hover); }
.act-btn--danger { color: var(--danger-color); border-color: var(--danger-color); }
.act-btn--danger:hover { background: var(--surface-hover); }
.act-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.hint-card { display: flex; gap: 10px; align-items: center; padding: var(--spacing-sm) var(--spacing-md); border-radius: var(--border-radius-md); background: var(--accent-bg); border: 1px solid var(--border-color-light); font-size: var(--font-size-xs); color: var(--primary-color); line-height: 1.5; }

.alias-list { display: flex; flex-direction: column; gap: var(--spacing-sm); width: 100%; }
.alias-row { display: flex; gap: var(--spacing-sm); align-items: center; }

.ai-result-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); padding: var(--spacing-lg); }
.ai-result-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-sm); }
.ai-result-summary { font-size: var(--font-size-sm); color: var(--text-color-secondary); line-height: 1.6; margin-bottom: var(--spacing-md); }
.ai-suggest-list { display: flex; flex-direction: column; gap: var(--spacing-sm); }
.ai-suggest-item { display: flex; flex-direction: column; gap: 4px; padding: var(--spacing-sm) var(--spacing-md); background: var(--bg-color-page); border: 1px solid var(--border-color-light); border-radius: var(--border-radius-md); }
.ai-suggest-info { display: flex; gap: var(--spacing-sm); align-items: center; justify-content: space-between; }
.ai-suggest-arrow { font-size: var(--font-size-sm); color: var(--text-color-primary); }
.ai-suggest-reason { font-size: var(--font-size-xs); color: var(--text-color-secondary); }
</style>
