<template>
  <div class="review-page-wrapper">
    <PageHeader
      :title="`审查配置 · ${repo_name}`"
      show-back
      :back-route="`/local-repos/${repo_key}/review`"
    />

    <div class="config-main">
      <SectionTitle title="通用设置" />
      <p class="config-desc">配置此仓库的代码审查行为。未设置的项将继承全局默认值。</p>

      <FormCard>
        <div class="form-row">
          <div class="form-info">
            <div class="form-label">启用代码审查</div>
            <div class="form-desc">MR 创建或更新时自动触发审查</div>
          </div>
          <el-switch v-model="form.enabled" />
        </div>

        <div class="form-row">
          <div class="form-info">
            <div class="form-label">高危级别阻止合并</div>
            <div class="form-desc">检测到严重或高危问题时阻止合并</div>
          </div>
          <el-switch v-model="form.blockOnHigh" />
        </div>

        <div class="form-field">
          <label class="form-label">LLM 提供商</label>
          <el-select v-model="form.defaultLLM" placeholder="跟随全局默认" clearable style="width: 100%">
            <el-option
              v-for="p in llmProviders"
              :key="p.name"
              :label="`${p.name} (${p.model})`"
              :value="p.name"
            >
              <span>{{ p.name }}</span>
              <span style="float:right;color:var(--el-text-color-secondary);font-size:12px">{{ p.model }}</span>
            </el-option>
          </el-select>
          <span v-if="!form.defaultLLM" class="form-hint">未设置时使用全局默认提供商</span>
        </div>

        <div class="form-row-2col">
          <div class="form-field">
            <label class="form-label">单次最大审查文件数</label>
            <el-input-number v-model="form.maxFiles" :min="1" :max="500" style="width: 100%" />
          </div>
          <div class="form-field">
            <label class="form-label">最大差异行数</label>
            <el-input-number v-model="form.maxDiffLines" :min="100" :max="10000" :step="100" style="width: 100%" />
          </div>
        </div>

        <template #footer>
          <ActionPill variant="outline" @click="loadConfig">取消</ActionPill>
          <ActionPill variant="primary" :disabled="saving" @click="handleSave">保存设置</ActionPill>
        </template>
      </FormCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getReviewConfig, updateReviewConfig } from '@/api/modules/review'
import { listLLMProviders } from '@/api/modules/llm-settings'
import type { LLMProviderDTO } from '@/api/modules/llm-settings'
import PageHeader from '@/components/common/PageHeader.vue'
import FormCard from '@/components/common/FormCard.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const route = useRoute()
const repo_key = route.params.repo_key as string
const repo_name = ref(repo_key)
const saving = ref(false)

const form = ref({
  enabled: true,
  blockOnHigh: true,
  defaultLLM: '',
  maxFiles: 50,
  maxDiffLines: 3000,
})

const llmProviders = ref<LLMProviderDTO[]>([])

async function loadConfig() {
  try {
    const res = await getReviewConfig(repo_key)
    if (res?.config_yaml) {
      try {
        const parsed = JSON.parse(res.config_yaml)
        form.value = { ...form.value, ...parsed }
      } catch {}
    }
  } catch (e) { console.error(e) }
}

async function loadLLMProviders() {
  try { llmProviders.value = await listLLMProviders() || [] } catch { llmProviders.value = [] }
}

async function handleSave() {
  saving.value = true
  try {
    await updateReviewConfig(repo_key, JSON.stringify(form.value, null, 2))
    ElMessage.success('配置已保存')
  } catch (e: any) { ElMessage.error('保存失败: ' + (e?.message || '')) }
  finally { saving.value = false }
}

onMounted(() => { loadConfig(); loadLLMProviders() })
</script>

<style scoped>
.review-page-wrapper { min-height: 100%; }
.config-main { padding: 24px; }
.config-desc { font-size: 14px; color: var(--el-text-color-secondary); margin: 0 0 20px; }
.form-row { display: flex; justify-content: space-between; align-items: center; }
.form-info { display: flex; flex-direction: column; gap: 4px; }
.form-label { font-size: 14px; font-weight: 600; }
.form-desc { font-size: 12px; color: var(--el-text-color-secondary); }
.form-field { display: flex; flex-direction: column; gap: 8px; }
.form-hint { font-size: 12px; color: var(--el-text-color-placeholder); }
.form-row-2col { display: flex; gap: 16px; }
.form-row-2col > .form-field { flex: 1; }
</style>
