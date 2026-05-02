<template>
  <div class="tab-section">
    <div class="section-desc">配置 Spec 文件 AI 辅助功能的默认行为。</div>

    <div class="ai-form-card">
      <div class="ai-field">
        <label class="ai-label">默认操作模式</label>
        <p class="ai-desc">选择 AI 辅助的默认交互模式</p>
        <el-select v-model="aiConfig.defaultAction" style="width: 300px">
          <el-option label="对话 (Chat)" value="chat" />
          <el-option label="补全 (Complete)" value="complete" />
          <el-option label="生成 (Generate)" value="generate" />
          <el-option label="代理 (Agent)" value="agent" />
        </el-select>
      </div>

      <div class="ai-field">
        <label class="ai-label">系统提示词</label>
        <p class="ai-desc">自定义 AI 辅助的系统提示词，留空使用默认提示词</p>
        <el-input
          v-model="aiConfig.systemPrompt"
          type="textarea"
          :rows="6"
          placeholder="输入自定义系统提示词..."
        />
      </div>

      <div class="ai-field">
        <label class="ai-label">自动修复</label>
        <p class="ai-desc">检测到问题时自动尝试修复</p>
        <el-switch v-model="aiConfig.autoFix" />
      </div>
    </div>

    <div class="form-actions">
      <el-button @click="loadAIConfig">重置</el-button>
      <el-button type="primary" :loading="saving" @click="saveAIConfig">保存 AI 配置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSpecConfig, saveSpecConfig } from '@/api/modules/spec'

const saving = ref(false)

const aiConfig = reactive({
  defaultAction: 'chat',
  systemPrompt: '',
  autoFix: false,
})

async function loadAIConfig() {
  try {
    const data = await getSpecConfig()
    if (data?.aiConfig) {
      Object.assign(aiConfig, data.aiConfig)
    }
  } catch {}
}

async function saveAIConfig() {
  saving.value = true
  try {
    await saveSpecConfig({ aiConfig: { ...aiConfig } })
    ElMessage.success('AI 配置已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally { saving.value = false }
}

onMounted(() => {
  loadAIConfig()
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

.ai-form-card {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  background: var(--bg-color-page);
}

.ai-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ai-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.ai-desc {
  margin: 0;
  font-size: 12px;
  color: var(--text-color-placeholder);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 8px;
}
</style>
