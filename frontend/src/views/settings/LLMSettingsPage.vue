<template>
  <div class="llm-settings-page">
    <PageHeader title="LLM 配置" subtitle="管理大模型服务提供商，用于代码审查、Spec 辅助等 AI 功能" show-back back-route="/settings" />

    <div class="section-header">
      <div class="section-info">
        <h3>提供商管理</h3>
        <p class="section-desc">支持 Coding Plan 订阅套餐、国内外直连 API 和本地部署模型。</p>
      </div>
      <div class="section-actions">
        <span v-if="providers.length > 0" class="count-tag">{{ providers.length }} 个已配置</span>
        <ActionPill variant="primary" :icon="Plus" @click="openAddDialog">添加提供商</ActionPill>
      </div>
    </div>

    <div v-if="loading" class="state-card"><LoadingState /></div>
    <div v-else-if="providers.length === 0" class="state-card">
      <EmptyState title="暂无 LLM 提供商" description="添加 Coding Plan 或直连 API 提供商，启用 AI 功能">
        <template #action>
          <ActionPill variant="primary" :icon="Plus" @click="openAddDialog">添加提供商</ActionPill>
        </template>
      </EmptyState>
    </div>
    <DataTable v-else :columns="providerColumns" :data="providers" row-key="id">
      <template #cell-name="{ row }">
        <span style="font-weight:500">{{ row.name }}</span>
        <span v-if="row.preset_id" class="preset-tag">{{ row.preset_id }}</span>
      </template>
      <template #cell-type="{ row }">
        {{ typeLabel(row.type) }}
      </template>
      <template #cell-model="{ row }">
        <span class="mono">{{ row.model }}</span>
      </template>
      <template #cell-base_url="{ row }">
        <span class="mono" style="font-size:12px">{{ row.base_url }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge v-if="row.is_default" variant="success" text="默认" :show-dot="false" />
        <StatusBadge v-else variant="info" text="可用" :show-dot="false" />
        <StatusBadge v-if="row.is_embedding" variant="warning" text="Embedding" :show-dot="false" style="margin-left:4px" />
      </template>
      <template #row-actions="{ row }">
        <button class="act-btn act-btn--primary" @click="openEditDialog(row)">编辑</button>
        <button class="act-btn act-btn--green" @click="handleTest(row)" :disabled="testingId === row.id">
          {{ testingId === row.id ? '测试中...' : '测试' }}
        </button>
        <button class="act-btn act-btn--purple" @click="handleTestEmbedding(row)" :disabled="testingEmbeddingId === row.id">
          {{ testingEmbeddingId === row.id ? '测试中...' : '测试 Embedding' }}
        </button>
        <button v-if="!row.is_default" class="act-btn act-btn--green" @click="handleSetDefault(row)">设为默认</button>
        <button class="act-btn act-btn--danger" @click="handleDelete(row)">删除</button>
      </template>
    </DataTable>

    <div class="tip-card">
      <el-icon :size="16" style="color:#D97706;flex-shrink:0"><InfoFilled /></el-icon>
      <span>提示：建议至少配置一个本地 LLM（Ollama）作为后备，避免网络故障时 AI 功能中断。API Key 将加密存储。</span>
    </div>

    <!-- Step 1: 预设选择 -->
    <el-dialog v-model="showDialog" title="添加 LLM 提供商" width="720px" destroy-on-close @close="resetDialog">
      <div v-if="dialogStep === 'select'">
        <div v-if="presetsLoading" style="text-align:center;padding:40px 0">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
          <p style="margin-top:12px;color:var(--text-color-secondary)">加载预设列表...</p>
        </div>
        <LLMPresetSelector v-else :presets="presets" @select="onPresetSelected" @custom="onCustomSelected" />
      </div>

      <!-- Step 2: 配置表单 -->
      <div v-else>
        <div class="form-header">
          <span class="form-header__title">配置: {{ selectedPreset?.display_name || '自定义' }}</span>
          <button class="back-btn" @click="dialogStep = 'select'">← 返回选择</button>
        </div>

        <div v-if="selectedPreset?.warning" class="warning-card">
          <el-icon :size="14" style="color:#D97706;flex-shrink:0"><WarningFilled /></el-icon>
          <span>{{ selectedPreset.warning }}</span>
        </div>

        <div v-if="selectedPreset?.subscribe_url" class="subscribe-card">
          <span>还没有订阅？</span>
          <a :href="selectedPreset.subscribe_url" target="_blank" class="subscribe-link">去订阅 →</a>
        </div>

        <el-form label-width="100px" style="margin-top:16px">
          <el-form-item v-if="selectedPreset?.supports_anthropic" label="协议">
            <el-radio-group v-model="form.protocol">
              <el-radio value="openai">OpenAI 兼容</el-radio>
              <el-radio value="anthropic">Anthropic 兼容</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="名称">
            <el-input v-model="form.name" placeholder="唯一标识" :disabled="!!editingProvider" />
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input v-model="form.base_url" :disabled="!!selectedPreset" placeholder="https://..." />
          </el-form-item>
          <div class="dialog-row">
            <el-form-item label="模型" class="dialog-half">
              <div v-if="form.type === 'ollama' && presetModels.length === 0" style="display:flex;gap:8px;width:100%">
                <el-select v-if="ollamaModels.length > 0" v-model="form.model" style="flex:1" filterable allow-create>
                  <el-option v-for="m in ollamaModels" :key="m" :label="m" :value="m" />
                </el-select>
                <el-input v-else v-model="form.model" placeholder="模型名称" style="flex:1" />
                <el-button @click="handleFetchOllamaModels" :loading="ollamaLoading" style="flex-shrink:0">获取模型</el-button>
              </div>
              <el-select v-else-if="presetModels.length > 0" v-model="form.model" style="width:100%">
                <el-option v-for="m in presetModels" :key="m.id" :label="m.display_name" :value="m.id" />
              </el-select>
              <el-input v-else v-model="form.model" placeholder="模型名称" />
            </el-form-item>
            <el-form-item label="最大 Tokens" class="dialog-half">
              <el-input-number v-model="form.max_tokens" :min="256" :max="128000" :step="512" style="width:100%" />
            </el-form-item>
          </div>
          <el-form-item v-if="form.type !== 'ollama'" label="API Key">
            <el-input v-model="form.api_key" type="password" show-password placeholder="sk-..." />
          </el-form-item>
          <el-form-item label="设为默认">
            <el-switch v-model="form.is_default" />
          </el-form-item>
          <el-divider content-position="left">Embedding 配置</el-divider>
          <el-form-item label="用作 Embedding">
            <el-switch v-model="form.is_embedding" />
          </el-form-item>
          <el-form-item v-if="form.is_embedding" label="Embedding 模型">
            <el-input v-model="form.embedding_model" :placeholder="form.type === 'ollama' ? 'nomic-embed-text' : 'text-embedding-3-small'" />
            <div style="font-size:11px;color:var(--text-color-secondary);margin-top:4px">
              留空使用默认模型（Ollama: nomic-embed-text, 其他: text-embedding-3-small）
            </div>
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <template v-if="dialogStep === 'select'">
          <el-button @click="showDialog = false">取消</el-button>
        </template>
        <template v-else>
          <el-button @click="dialogStep = 'select'">返回</el-button>
          <el-button type="success" @click="handleTestForm" :loading="testingForm">测试连接</el-button>
          <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, InfoFilled, WarningFilled, Loading } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import LLMPresetSelector from '@/components/llm/LLMPresetSelector.vue'
import {
  listLLMProviders, createLLMProvider, updateLLMProvider,
  deleteLLMProvider, setDefaultLLMProvider, testLLMProvider, fetchLLMPresets,
  fetchOllamaModels, testEmbedding,
} from '@/api/modules/llm-settings'
import type { LLMProviderDTO, LLMPresetDTO } from '@/api/modules/llm-settings'

const providerColumns: TableColumn[] = [
  { key: 'name', label: '名称', width: '180px' },
  { key: 'type', label: '类型', width: '120px' },
  { key: 'model', label: '模型', flex: 1 },
  { key: 'base_url', label: 'Base URL', width: '260px' },
  { key: 'status', label: '状态', width: '80px' },
]

const loading = ref(false)
const providers = ref<LLMProviderDTO[]>([])
const testingId = ref<number | null>(null)
const testingEmbeddingId = ref<number | null>(null)
const showDialog = ref(false)
const editingProvider = ref<LLMProviderDTO | null>(null)
const saving = ref(false)
const testingForm = ref(false)
const dialogStep = ref<'select' | 'configure'>('select')

const presets = ref<LLMPresetDTO[]>([])
const presetsLoading = ref(false)
const selectedPreset = ref<LLMPresetDTO | null>(null)
const ollamaModels = ref<string[]>([])
const ollamaLoading = ref(false)

const form = ref({
  type: 'openai_compatible',
  name: '',
  model: '',
  max_tokens: 4096,
  base_url: '',
  api_key: '',
  is_default: false,
  is_embedding: false,
  embedding_model: '',
  preset_id: '',
  protocol: 'openai',
})

const presetModels = computed(() => {
  if (!selectedPreset.value) return []
  return selectedPreset.value.models || []
})

function typeLabel(t: string) {
  const m: Record<string, string> = {
    openai_compatible: 'OpenAI 兼容',
    ollama: 'Ollama',
    anthropic: 'Anthropic',
    gemini: 'Gemini',
  }
  return m[t] || t
}

async function loadProviders() {
  loading.value = true
  try { providers.value = await listLLMProviders() || [] } catch { providers.value = [] }
  finally { loading.value = false }
}

async function loadPresets() {
  presetsLoading.value = true
  try { presets.value = await fetchLLMPresets() || [] } catch { presets.value = [] }
  finally { presetsLoading.value = false }
}

function resetDialog() {
  dialogStep.value = 'select'
  selectedPreset.value = null
  editingProvider.value = null
  ollamaModels.value = []
  form.value = { type: 'openai_compatible', name: '', model: '', max_tokens: 4096, base_url: '', api_key: '', is_default: false, is_embedding: false, embedding_model: '', preset_id: '', protocol: 'openai' }
}

async function handleFetchOllamaModels() {
  ollamaLoading.value = true
  try {
    const models = await fetchOllamaModels(form.value.base_url || undefined)
    ollamaModels.value = models || []
    if (ollamaModels.value.length === 0) {
      ElMessage.warning('未发现已安装的模型')
    } else if (!form.value.model && ollamaModels.value.length > 0) {
      form.value.model = ollamaModels.value[0]!
    }
  } catch (e: any) {
    ElMessage.error('获取 Ollama 模型失败: ' + (e?.message || ''))
  } finally {
    ollamaLoading.value = false
  }
}

function openAddDialog() {
  resetDialog()
  showDialog.value = true
  loadPresets()
}

function onPresetSelected(preset: LLMPresetDTO) {
  selectedPreset.value = preset
  form.value.type = preset.type
  form.value.base_url = preset.base_url
  form.value.model = preset.default_model
  form.value.max_tokens = preset.max_tokens
  form.value.preset_id = preset.id
  form.value.protocol = 'openai'
  form.value.name = preset.id.replace(/-/g, '_')
  dialogStep.value = 'configure'
}

function onCustomSelected() {
  selectedPreset.value = null
  form.value = { type: 'openai_compatible', name: '', model: '', max_tokens: 4096, base_url: '', api_key: '', is_default: false, is_embedding: false, embedding_model: '', preset_id: '', protocol: 'openai' }
  dialogStep.value = 'configure'
}

function openEditDialog(p: LLMProviderDTO) {
  editingProvider.value = p
  selectedPreset.value = null
  form.value = {
    type: p.type, name: p.name, model: p.model, max_tokens: p.max_tokens,
    base_url: p.base_url, api_key: '', is_default: p.is_default,
    is_embedding: p.is_embedding || false, embedding_model: p.embedding_model || '',
    preset_id: p.preset_id || '', protocol: p.protocol || 'openai',
  }
  dialogStep.value = 'configure'
  showDialog.value = true
}

async function handleSave() {
  const f = form.value
  if (!f.name || !f.type || !f.model) {
    ElMessage.warning('请填写名称、类型和模型')
    return
  }
  if (f.type !== 'ollama' && !f.base_url) {
    ElMessage.warning('请填写 Base URL')
    return
  }

  let effectiveType = f.type
  let effectiveBaseURL = f.base_url
  if (selectedPreset.value?.supports_anthropic && f.protocol === 'anthropic') {
    effectiveType = 'anthropic'
    effectiveBaseURL = selectedPreset.value.anthropic_url || f.base_url
  }

  saving.value = true
  try {
    const payload: any = {
      name: f.name, type: effectiveType, model: f.model,
      max_tokens: f.max_tokens, base_url: effectiveBaseURL,
      is_default: f.is_default, is_embedding: f.is_embedding, embedding_model: f.embedding_model,
      preset_id: f.preset_id, protocol: f.protocol,
    }
    if (f.api_key) payload.api_key = f.api_key
    if (editingProvider.value) {
      await updateLLMProvider(editingProvider.value.id, payload)
    } else {
      await createLLMProvider(payload)
    }
    ElMessage.success(editingProvider.value ? '更新成功' : '添加成功')
    showDialog.value = false
    loadProviders()
  } catch (e: any) { ElMessage.error((editingProvider.value ? '更新' : '添加') + '失败: ' + (e?.message || '')) }
  finally { saving.value = false }
}

async function handleTest(p: LLMProviderDTO) {
  testingId.value = p.id
  try { await testLLMProvider(p.id); ElMessage.success('连接测试成功') }
  catch (e: any) { ElMessage.error('连接测试失败: ' + (e?.message || '')) }
  finally { testingId.value = null }
}

async function handleTestEmbedding(p: LLMProviderDTO) {
  testingEmbeddingId.value = p.id
  try {
    const res = await testEmbedding(p.id)
    ElMessage.success(`Embedding 测试成功 (模型: ${res?.model || p.embedding_model || 'default'})`)
  } catch (e: any) { ElMessage.error('Embedding 测试失败: ' + (e?.message || '')) }
  finally { testingEmbeddingId.value = null }
}

async function handleTestForm() {
  testingForm.value = true
  try {
    const f = form.value
    let effectiveType = f.type
    let effectiveBaseURL = f.base_url
    if (selectedPreset.value?.supports_anthropic && f.protocol === 'anthropic') {
      effectiveType = 'anthropic'
      effectiveBaseURL = selectedPreset.value.anthropic_url || f.base_url
    }
    const tempData: any = {
      name: f.name || '__test__', type: effectiveType, model: f.model,
      base_url: effectiveBaseURL, max_tokens: f.max_tokens,
      is_default: false, is_embedding: f.is_embedding, embedding_model: f.embedding_model,
      preset_id: f.preset_id, protocol: f.protocol,
    }
    if (f.api_key) tempData.api_key = f.api_key
    const created = await createLLMProvider(tempData)
    if (created) {
      try { await testLLMProvider(created.id) } finally { await deleteLLMProvider(created.id) }
    }
    ElMessage.success('连接测试成功')
  } catch (e: any) { ElMessage.error('连接测试失败: ' + (e?.message || '')) }
  finally { testingForm.value = false }
}

async function handleSetDefault(p: LLMProviderDTO) {
  try { await setDefaultLLMProvider(p.id); ElMessage.success('已设为默认'); loadProviders() }
  catch (e: any) { ElMessage.error('操作失败: ' + (e?.message || '')) }
}

async function handleDelete(p: LLMProviderDTO) {
  try { await ElMessageBox.confirm(`确定删除提供商 "${p.name}"？`, '确认删除', { type: 'warning' }) } catch { return }
  try { await deleteLLMProvider(p.id); ElMessage.success('已删除'); loadProviders() }
  catch (e: any) { ElMessage.error('删除失败: ' + (e?.message || '')) }
}

onMounted(() => { loadProviders() })
</script>

<style scoped>
.llm-settings-page { display: flex; flex-direction: column; gap: 20px; }
.section-header { display: flex; justify-content: space-between; align-items: flex-start; }
.section-header h3 { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-color-primary); }
.section-desc { margin: 4px 0 0; font-size: 13px; color: var(--text-color-secondary); }
.section-actions { display: flex; align-items: center; gap: 12px; }
.count-tag { padding: 4px 10px; border-radius: 4px; background: #ECFDF5; color: #059669; font-size: 12px; font-weight: 500; }
.state-card { border-radius: 12px; border: 1px solid var(--border-color); background: var(--bg-color-page); }
.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 12px; }
.preset-tag {
  margin-left: 8px; padding: 1px 6px; border-radius: 4px; font-size: 10px;
  background: #EFF6FF; color: #2563EB; font-weight: 500;
}
.act-btn {
  display: inline-flex; align-items: center; gap: 4px; padding: 4px 10px;
  border-radius: 4px; font-size: 11px; cursor: pointer; transition: all 0.2s;
  border: 1px solid var(--border-color); background: transparent; color: var(--text-color-secondary);
}
.act-btn--primary { border-color: #6366F1; color: #6366F1; }
.act-btn--primary:hover { background: var(--accent-bg); }
.act-btn--green { border-color: #10B981; color: #10B981; }
.act-btn--green:hover { background: #ECFDF5; }
.act-btn--purple { border-color: #8B5CF6; color: #8B5CF6; }
.act-btn--purple:hover { background: #F5F3FF; }
.act-btn--danger { border-color: #EF4444; color: #EF4444; }
.act-btn--danger:hover { background: #FEF2F2; }
.act-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.tip-card {
  display: flex; align-items: center; gap: 10px; padding: 14px 16px;
  border-radius: 8px; background: #FFFBEB; border: 1px solid #F59E0B;
  font-size: 13px; color: #92400E; line-height: 1.5;
}
.dialog-row { display: flex; gap: 16px; }
.dialog-half { flex: 1; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.form-header__title { font-size: 14px; font-weight: 600; color: var(--text-color-primary); }
.back-btn { background: none; border: none; color: var(--accent-primary); cursor: pointer; font-size: 13px; }
.back-btn:hover { text-decoration: underline; }
.warning-card {
  display: flex; align-items: flex-start; gap: 8px; padding: 10px 14px;
  border-radius: 8px; background: #FFFBEB; border: 1px solid #F59E0B;
  font-size: 12px; color: #92400E; line-height: 1.5;
}
.subscribe-card {
  display: flex; align-items: center; gap: 8px; padding: 8px 14px;
  border-radius: 8px; background: #EFF6FF; border: 1px solid #93C5FD;
  font-size: 12px; color: #1E40AF;
}
.subscribe-link { color: #2563EB; font-weight: 500; text-decoration: none; }
.subscribe-link:hover { text-decoration: underline; }
</style>
