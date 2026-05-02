<template>
  <div class="spec-settings-page">
    <PageHeader title="Spec 全局配置" subtitle="管理 Spec 文件的默认模板、Lint 规则、格式化选项和 AI 辅助配置" show-back back-route="/settings" />

    <el-tabs v-model="activeTab" class="settings-tabs">
      <el-tab-pane label="默认模板" name="template">
        <div class="tab-section">
          <div class="section-desc">新建 Spec 文件时使用的默认模板内容。留空则使用系统内置模板。</div>
          <el-input
            v-model="templateContent"
            type="textarea"
            :rows="18"
            placeholder="输入默认 Spec 模板内容..."
            class="template-editor"
          />
          <div class="form-actions">
            <el-button @click="resetTemplate">恢复默认</el-button>
            <el-button type="primary" :loading="saving" @click="saveTemplate">保存模板</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="Lint 规则管理" name="lint">
        <LintRulesTab />
      </el-tab-pane>

      <el-tab-pane label="格式化配置" name="format">
        <FormatOptionsTab />
      </el-tab-pane>

      <el-tab-pane label="AI 辅助配置" name="ai">
        <AISettingsTab />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import PageHeader from '@/components/common/PageHeader.vue'
import LintRulesTab from '@/components/settings/LintRulesTab.vue'
import FormatOptionsTab from '@/components/settings/FormatOptionsTab.vue'
import AISettingsTab from '@/components/settings/AISettingsTab.vue'
import { getSpecConfig, saveSpecConfig } from '@/api/modules/spec'

const activeTab = ref('template')
const saving = ref(false)

const builtinTemplate = ref('')
const templateContent = ref('')

async function loadConfig() {
  try {
    const data = await getSpecConfig()
    if (data?.defaultTemplate) {
      templateContent.value = data.defaultTemplate
      builtinTemplate.value = data.defaultTemplate
    }
  } catch {}
}

function resetTemplate() {
  templateContent.value = builtinTemplate.value
  ElMessage.success('已恢复默认模板')
}

async function saveTemplate() {
  saving.value = true
  try {
    await saveSpecConfig({ defaultTemplate: templateContent.value })
    ElMessage.success('模板已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally { saving.value = false }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.spec-settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.settings-tabs :deep(.el-tabs__header) {
  margin-bottom: 20px;
}

.settings-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  font-weight: 500;
}

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

.template-editor :deep(.el-textarea__inner) {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 8px;
}
</style>
