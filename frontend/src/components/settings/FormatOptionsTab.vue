<template>
  <div class="tab-section">
    <div class="section-desc">配置 Spec 文件格式化器的行为。这些选项控制自动格式化时的转换规则。</div>

    <div class="format-group">
      <h4 class="group-title">宏风格</h4>
      <div class="option-grid">
        <div class="option-item">
          <div class="option-label">
            <span>花括号宏风格 (Curlify)</span>
            <span class="option-desc">将 %{macro} 转换为 ${macro} 风格</span>
          </div>
          <el-switch v-model="formatOptions.curlify" />
        </div>
        <div class="option-item">
          <div class="option-label">
            <span>路径宏替换</span>
            <span class="option-desc">将硬编码路径替换为 RPM 宏</span>
          </div>
          <el-switch v-model="formatOptions.path_macros" />
        </div>
        <div class="option-item">
          <div class="option-label">
            <span>工具宏替换</span>
            <span class="option-desc">使用标准构建工具宏</span>
          </div>
          <el-switch v-model="formatOptions.util_macros" />
        </div>
      </div>
    </div>

    <div class="format-group">
      <h4 class="group-title">清理选项</h4>
      <div class="option-grid">
        <div class="option-item">
          <div class="option-label"><span>移除 %clean 段</span></div>
          <el-switch v-model="formatOptions.remove_clean" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>移除 BuildRoot</span></div>
          <el-switch v-model="formatOptions.remove_build_root" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>移除 Group 标签</span></div>
          <el-switch v-model="formatOptions.remove_group" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>通用清理</span></div>
          <el-switch v-model="formatOptions.common_cleanup" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>条件语句修剪</span></div>
          <el-switch v-model="formatOptions.conditional_trim" />
        </div>
      </div>
    </div>

    <div class="format-group">
      <h4 class="group-title">排版选项</h4>
      <div class="option-grid">
        <div class="option-item">
          <div class="option-label"><span>Tab 转空格</span></div>
          <el-switch v-model="formatOptions.tab_to_spaces" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>缩进大小</span></div>
          <el-input-number v-model="formatOptions.indent_size" :min="2" :max="8" :step="1" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>排序前言标签</span></div>
          <el-switch v-model="formatOptions.preamble_order" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>对齐标签值</span></div>
          <el-switch v-model="formatOptions.align_values" />
        </div>
        <div class="option-item">
          <div class="option-label"><span>排序依赖</span></div>
          <el-switch v-model="formatOptions.sort_deps" />
        </div>
      </div>
    </div>

    <div class="format-group">
      <h4 class="group-title">许可证</h4>
      <div class="option-grid">
        <div class="option-item">
          <div class="option-label">
            <span>License 转 SPDX</span>
            <span class="option-desc">将 License 字段转换为 SPDX 标识符</span>
          </div>
          <el-switch v-model="formatOptions.license_spdx" />
        </div>
      </div>
    </div>

    <div class="form-actions">
      <el-button @click="loadFormatOptions">重置</el-button>
      <el-button type="primary" :loading="saving" @click="saveFormatOptions">保存格式化配置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSpecConfig, saveSpecConfig } from '@/api/modules/spec'

const saving = ref(false)

const formatOptions = reactive({
  curlify: true,
  remove_clean: true,
  remove_build_root: true,
  remove_group: false,
  license_spdx: true,
  sort_deps: true,
  tab_to_spaces: true,
  indent_size: 4,
  preamble_order: true,
  align_values: true,
  path_macros: true,
  util_macros: true,
  common_cleanup: true,
  conditional_trim: true,
})

async function loadFormatOptions() {
  try {
    const data = await getSpecConfig()
    if (data?.formatOptions) {
      Object.assign(formatOptions, data.formatOptions)
    }
  } catch {}
}

async function saveFormatOptions() {
  saving.value = true
  try {
    await saveSpecConfig({ formatOptions: { ...formatOptions } })
    ElMessage.success('格式化配置已保存')
  } catch (e: any) {
    ElMessage.error('保存失败: ' + (e?.message || ''))
  } finally { saving.value = false }
}

onMounted(() => {
  loadFormatOptions()
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

.format-group {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  overflow: hidden;
}

.group-title {
  margin: 0;
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
  background: var(--surface-card);
  border-bottom: 1px solid var(--border-color);
}

.option-grid {
  display: flex;
  flex-direction: column;
}

.option-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color-light);
}

.option-item:last-child {
  border-bottom: none;
}

.option-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.option-label span:first-child {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.option-desc {
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
