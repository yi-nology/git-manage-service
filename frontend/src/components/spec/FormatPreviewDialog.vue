<template>
  <el-dialog v-model="optionsVisible" width="520px" :close-on-click-modal="false" :show-close="false" append-to-body class="fmt-opt-dialog">
    <template #header>
      <div class="fmt-opt-header">
        <span class="fmt-opt-title">格式化选项</span>
        <el-button text size="small" @click="optionsVisible = false">
          <el-icon :size="16" color="var(--text-color-secondary, #64748B)"><Close /></el-icon>
        </el-button>
      </div>
    </template>
    <div class="fmt-opt-body">
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">宏括号规范化</span>
          <span class="fmt-opt-desc">将 %macro 转为 %{macro}，排除白名单关键字</span>
        </div>
        <el-switch v-model="formatOpts.curlify" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">删除 %clean 段</span>
          <span class="fmt-opt-desc">现代 RPM 不再需要 %clean section</span>
        </div>
        <el-switch v-model="formatOpts.remove_clean" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">删除 BuildRoot</span>
          <span class="fmt-opt-desc">现代 RPM 已废弃 BuildRoot 字段</span>
        </div>
        <el-switch v-model="formatOpts.remove_build_root" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">删除 Group 字段</span>
          <span class="fmt-opt-desc">现代 RPM 不再使用 Group 分类</span>
        </div>
        <el-switch v-model="formatOpts.remove_group" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">License SPDX 修正</span>
          <span class="fmt-opt-desc">将非标准 License 名修正为 SPDX 标识符</span>
        </div>
        <el-switch v-model="formatOpts.license_spdx" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">依赖排序去重</span>
          <span class="fmt-opt-desc">BuildRequires/Requires 按字母排序并去重</span>
        </div>
        <el-switch v-model="formatOpts.sort_deps" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">Tab 转空格</span>
          <span class="fmt-opt-desc">所有 Tab 替换为空格</span>
        </div>
        <el-switch v-model="formatOpts.tab_to_spaces" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">Preamble 标签排序</span>
          <span class="fmt-opt-desc">按 RPM 规范顺序重排 Name/Version/BuildRequires 等</span>
        </div>
        <el-switch v-model="formatOpts.preamble_order" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">标签值对齐</span>
          <span class="fmt-opt-desc">统一冒号后缩进，提升可读性</span>
        </div>
        <el-switch v-model="formatOpts.align_values" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">路径宏替换</span>
          <span class="fmt-opt-desc">硬编码路径 → RPM 宏 (如 /usr/bin → %{_bindir})</span>
        </div>
        <el-switch v-model="formatOpts.path_macros" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">工具宏展开</span>
          <span class="fmt-opt-desc">%{__make} → make, %{__rm} → rm 等</span>
        </div>
        <el-switch v-model="formatOpts.util_macros" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">通用清理</span>
          <span class="fmt-opt-desc">$RPM_BUILD_ROOT → %{buildroot}, egrep → grep -E</span>
        </div>
        <el-switch v-model="formatOpts.common_cleanup" />
      </div>
      <div class="fmt-opt-row">
        <div class="fmt-opt-label">
          <span class="fmt-opt-name">条件块空行压缩</span>
          <span class="fmt-opt-desc">移除 %if/%else/%endif 后的多余空行</span>
        </div>
        <el-switch v-model="formatOpts.conditional_trim" />
      </div>
    </div>
    <template #footer>
      <div class="fmt-opt-footer">
        <el-button class="fmt-btn-outline" @click="optionsVisible = false">关闭</el-button>
        <el-button type="primary" @click="optionsVisible = false; handleFormat()">应用并格式化</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="previewVisible" width="1000px" :close-on-click-modal="false" :show-close="false" append-to-body class="fmt-preview-dialog">
    <template #header>
      <div class="fmt-diff-header">
        <span class="fmt-diff-title">格式化预览 — {{ formatChanges.length }} 处变更</span>
        <div class="fmt-diff-stats">
          <span class="added">+{{ formatAddedLines }} 行</span>
          <span class="removed">-{{ formatRemovedLines }} 行</span>
        </div>
      </div>
    </template>
    <div class="fmt-diff-tabs">
      <span :class="['fmt-tab', { active: fmtPreviewTab === 'changes' }]" @click="fmtPreviewTab = 'changes'">变更列表</span>
      <span :class="['fmt-tab', { active: fmtPreviewTab === 'diff' }]" @click="fmtPreviewTab = 'diff'">并排对比</span>
    </div>
    <div v-if="fmtPreviewTab === 'changes'" class="fmt-diff-body">
      <div v-if="formatChanges.length > 0" class="fmt-changes-list">
        <el-scrollbar max-height="420px">
          <div v-for="(ch, idx) in formatChanges" :key="idx" class="fmt-change-card">
            <div class="fmt-change-head">
              <span :class="['fmt-change-tag', `fmt-tag-${ch.type}`]">
                {{ { removed: '删除', modified: '修改', reordered: '排序' }[ch.type] || ch.type }}
              </span>
              <span class="fmt-change-reason">{{ ch.reason }}</span>
            </div>
            <div v-if="ch.before" class="fmt-change-line fmt-line-removed">- {{ ch.before }}</div>
            <div v-if="ch.after" class="fmt-change-line fmt-line-added">+ {{ ch.after }}</div>
          </div>
        </el-scrollbar>
      </div>
      <div v-else class="fmt-no-changes">
        <el-icon :size="28" color="#10B981"><CircleCheck /></el-icon>
        <span>无需格式化，文件已符合规范</span>
      </div>
    </div>
    <div v-else class="fmt-diff-body">
      <el-scrollbar max-height="420px">
        <div class="fmt-side-by-side">
          <div class="fmt-side-col">
            <div class="fmt-side-title">原始文件</div>
            <pre class="fmt-side-content">{{ content }}</pre>
          </div>
          <div class="fmt-side-col">
            <div class="fmt-side-title">格式化后</div>
            <pre class="fmt-side-content">{{ formatResult }}</pre>
          </div>
        </div>
      </el-scrollbar>
    </div>
    <template #footer>
      <div class="fmt-diff-footer">
        <el-button class="fmt-btn-outline" @click="previewVisible = false">取消</el-button>
        <el-button type="primary" :disabled="formatChanges.length === 0" @click="onApply">应用格式化</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Close, CircleCheck } from '@element-plus/icons-vue'
import { formatSpec } from '@/api/modules/spec'
import type { FormatChange } from '@/types/spec'

const props = defineProps<{ modelValue: boolean; content: string }>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'apply-format': [content: string]
}>()

const optionsVisible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const previewVisible = ref(false)
const formatChanges = ref<FormatChange[]>([])
const formatResult = ref('')
const fmtPreviewTab = ref<'changes' | 'diff'>('changes')
const formatAddedLines = computed(() => formatChanges.value.filter(c => c.after).length)
const formatRemovedLines = computed(() => formatChanges.value.filter(c => c.before).length)
const formatOpts = ref({
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

async function handleFormat() {
  if (!props.content) { ElMessage.warning('请先选择一个 Spec 文件'); return }
  try {
    const result = await formatSpec(props.content, { ...formatOpts.value })
    formatResult.value = result.content
    formatChanges.value = result.changes || []
    if (formatChanges.value.length === 0 && result.content === props.content) {
      ElMessage.success('无需格式化，文件已符合规范')
      return
    }
    previewVisible.value = true
  } catch (e: any) {
    ElMessage.error('格式化失败: ' + (e?.message || ''))
  }
}

function onApply() {
  if (!formatResult.value) return
  emit('apply-format', formatResult.value)
  previewVisible.value = false
  ElMessage.success('格式化已应用')
}
</script>

<style scoped>
.fmt-opt-dialog :deep(.el-dialog) { background: var(--bg-color-page, #FFFFFF); border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-opt-dialog :deep(.el-dialog__header) { margin: 0; padding: 0; }
.fmt-opt-dialog :deep(.el-dialog__body) { padding: 0; }
.fmt-opt-dialog :deep(.el-dialog__footer) { padding: 0; }
.fmt-preview-dialog :deep(.el-dialog) { background: var(--bg-color-page, #FFFFFF); border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-preview-dialog :deep(.el-dialog__header) { margin: 0; padding: 0; }
.fmt-preview-dialog :deep(.el-dialog__body) { padding: 0; }
.fmt-preview-dialog :deep(.el-dialog__footer) { padding: 0; }
</style>

<style>
.fmt-opt-dialog .el-dialog { background: var(--bg-color-page, #FFFFFF) !important; border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-opt-dialog .el-dialog__header { margin: 0; padding: 0; }
.fmt-opt-dialog .el-dialog__body { padding: 0; }
.fmt-opt-dialog .el-dialog__footer { padding: 0; }
.fmt-opt-header { height: 48px; padding: 0 20px; display: flex; align-items: center; justify-content: space-between; background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); }
.fmt-opt-title { font-size: 14px; font-weight: 600; color: var(--text-color-primary, #1E293B); }
.fmt-opt-body { padding: 20px 24px; display: flex; flex-direction: column; gap: 20px; }
.fmt-opt-row { display: flex; align-items: center; justify-content: space-between; }
.fmt-opt-label { display: flex; flex-direction: column; gap: 2px; max-width: 380px; }
.fmt-opt-name { font-size: 13px; font-weight: 500; color: var(--text-color-primary, #1E293B); }
.fmt-opt-desc { font-size: 11px; color: var(--text-color-secondary, #64748B); }
.fmt-opt-row .el-switch { --el-switch-on-color: #6366F1; --el-switch-off-color: #CBD5E1; }
.fmt-opt-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 0 20px; height: 56px; align-items: center; background: var(--bg-color, #F8F9FC); border-top: 1px solid var(--border-color, #E2E8F0); border-radius: 0 0 12px 12px; }
.fmt-btn-outline { color: var(--text-color-regular, #475569) !important; border-color: var(--border-color, #E2E8F0) !important; background: transparent !important; }
.fmt-btn-outline:hover { border-color: var(--primary-color, #6366F1) !important; color: var(--primary-color, #6366F1) !important; }

.fmt-preview-dialog .el-dialog { background: var(--bg-color-page, #FFFFFF) !important; border: 1px solid var(--border-color, #E2E8F0); border-radius: 12px; overflow: hidden; }
.fmt-preview-dialog .el-dialog__header { margin: 0; padding: 0; }
.fmt-preview-dialog .el-dialog__body { padding: 0; }
.fmt-preview-dialog .el-dialog__footer { padding: 0; }
.fmt-diff-header { height: 48px; padding: 0 20px; display: flex; align-items: center; justify-content: space-between; background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); border-radius: 12px 12px 0 0; }
.fmt-diff-title { font-size: 14px; font-weight: 600; color: var(--text-color-primary, #1E293B); }
.fmt-diff-stats { display: flex; gap: 12px; font-size: 12px; font-weight: 500; }
.fmt-diff-stats .added { color: #22C55E; }
.fmt-diff-stats .removed { color: #EF4444; }
.fmt-diff-tabs { height: 36px; padding: 0 20px; display: flex; gap: 16px; align-items: center; background: var(--bg-color-page, #FFFFFF); border-bottom: 1px solid var(--border-color, #E2E8F0); }
.fmt-tab { font-size: 12px; color: var(--text-color-secondary, #64748B); cursor: pointer; transition: color 0.15s; }
.fmt-tab:hover { color: var(--text-color-primary, #1E293B); }
.fmt-tab.active { color: #6366F1; font-weight: 500; }
.fmt-diff-body { padding: 16px; min-height: 200px; }
.fmt-changes-list { display: flex; flex-direction: column; gap: 10px; }
.fmt-change-card { padding: 10px 14px; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; display: flex; flex-direction: column; gap: 6px; }
.fmt-change-head { display: flex; align-items: center; gap: 8px; }
.fmt-change-tag { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; color: #fff; }
.fmt-tag-removed { background: #EF4444; }
.fmt-tag-modified { background: #F59E0B; }
.fmt-tag-reordered { background: #6366F1; }
.fmt-change-reason { font-size: 11px; color: var(--text-color-secondary, #64748B); }
.fmt-change-line { font-family: 'Consolas', 'Courier New', monospace; font-size: 11px; white-space: pre-wrap; word-break: break-all; }
.fmt-line-removed { color: #EF4444; }
.fmt-line-added { color: #22C55E; }
.fmt-no-changes { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; padding: 48px; color: #10B981; font-size: 14px; }
.fmt-diff-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 0 20px; height: 56px; align-items: center; background: var(--bg-color, #F8F9FC); border-top: 1px solid var(--border-color, #E2E8F0); border-radius: 0 0 12px 12px; }
.fmt-side-by-side { display: flex; gap: 12px; }
.fmt-side-col { flex: 1; min-width: 0; background: var(--bg-color, #F8F9FC); border: 1px solid var(--border-color, #E2E8F0); border-radius: 6px; overflow: hidden; }
.fmt-side-title { padding: 8px 14px; font-size: 12px; font-weight: 500; color: var(--text-color-secondary, #64748B); background: var(--bg-color, #F8F9FC); border-bottom: 1px solid var(--border-color, #E2E8F0); }
.fmt-side-content { margin: 0; padding: 12px 14px; font-size: 11px; font-family: 'Consolas', 'Courier New', monospace; color: var(--text-color-regular, #475569); line-height: 1.6; white-space: pre-wrap; word-break: break-all; }
</style>
