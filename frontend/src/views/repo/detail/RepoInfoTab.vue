<template>
  <div v-if="repo" class="info-card">
    <div class="info-top-row">
      <div class="info-left-col">
        <div class="info-section-header">
          <SectionTitle title="基本信息" />
          <ActionPill variant="outline" :icon="Edit" @click="$emit('open-edit')">
            编辑仓库
          </ActionPill>
        </div>
        <div class="info-row">
          <div class="info-field"><span class="info-label">名称</span><span class="info-value info-value--bold">{{ repo.name }}</span></div>
          <div class="info-field"><span class="info-label">当前版本</span><StatusBadge v-if="currentVersion" variant="success" :text="currentVersion" :show-dot="false" /><span v-else class="info-value">-</span></div>
        </div>
        <div class="info-field"><span class="info-label">本地路径</span><span class="info-value mono">{{ repo.path }}</span></div>
        <div class="info-row">
          <div class="info-field"><span class="info-label">Repo Key</span><span class="info-value info-value--accent">{{ repo.key }}<button class="copy-btn-sm" @click="copyKey">复制</button></span></div>
          <div class="info-field"><span class="info-label">远程 URL</span><span class="info-value">{{ repo.remote_url || '-' }}</span></div>
        </div>
        <div class="info-row">
          <div class="info-field"><span class="info-label">创建时间</span><span class="info-value">{{ formatDate(repo.created_at) }}</span></div>
          <div class="info-field"><span class="info-label">更新时间</span><span class="info-value">{{ formatDate(repo.updated_at) }}</span></div>
        </div>
      </div>

      <div class="info-v-divider"></div>

      <div class="info-right-col">
        <BindingPanel
          :bindings="bindings"
          @add="$emit('open-binding-dialog')"
          @delete="(id: number) => $emit('delete-binding', id)"
          @set-primary="(id: number) => $emit('set-primary-binding', id)"
          @register-webhook="(id: number) => $emit('register-webhook', id)"
          @delete-webhook="(id: number) => $emit('delete-webhook', id)"
        />
      </div>
    </div>

    <template v-if="scanData">
      <div class="info-divider"></div>
      <RemoteConfigSection :scan-data="scanData" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { Edit } from '@element-plus/icons-vue'
import type { RepoDTO, ScanResult } from '@/types/repo'
import type { RepoProviderBindingDTO } from '@/types/binding'
import { formatDate } from '@/utils/format'
import ActionPill from '@/components/common/ActionPill.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import BindingPanel from '@/components/binding/BindingPanel.vue'
import RemoteConfigSection from './RemoteConfigSection.vue'

const props = defineProps<{
  repo: RepoDTO
  currentVersion: string
  scanData: ScanResult | null
  bindings: RepoProviderBindingDTO[]
}>()

defineEmits<{
  'open-edit': []
  'open-binding-dialog': []
  'delete-binding': [id: number]
  'set-primary-binding': [id: number]
  'register-webhook': [id: number]
  'delete-webhook': [id: number]
}>()

function copyKey() {
  if (props.repo?.key) {
    navigator.clipboard.writeText(props.repo.key)
    ElMessage.success('已复制 Repo Key')
  }
}
</script>

<style scoped>
.info-card {
  border-radius: var(--border-radius-md);
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 0;
  box-shadow: var(--box-shadow-sm);
}

.info-top-row {
  display: flex;
  gap: 0;
}

.info-left-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-right: 20px;
}

.info-v-divider {
  width: 1px;
  background: var(--border-color);
  align-self: stretch;
}

.info-right-col {
  width: 340px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-left: 20px;
}

.info-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row {
  display: flex;
  gap: 20px;
  min-width: 0;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.info-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.info-value {
  font-size: 14px;
  color: var(--text-color-primary);
  min-width: 0;
  overflow-wrap: anywhere;
}

.info-value--bold { font-weight: 500; }

.info-value--accent {
  color: var(--accent-primary);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-value.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 13px;
}

.copy-btn-sm {
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  background: transparent;
  font-size: 11px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.copy-btn-sm:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.info-divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
}

@media (max-width: 768px) {
  .info-top-row,
  .info-row {
    flex-direction: column;
  }

  .info-left-col,
  .info-right-col {
    width: auto;
    padding: 0;
  }

  .info-v-divider {
    width: auto;
    height: 1px;
    margin: 16px 0;
  }
}
</style>
