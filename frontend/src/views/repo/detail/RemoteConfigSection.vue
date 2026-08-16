<template>
  <div class="scan-section">
    <div class="info-section-header" style="margin-bottom:12px">
      <SectionTitle title="远程配置" />
      <span class="info-subtitle">来自 .git/config</span>
    </div>
    <div class="scan-remote-list">
      <div v-for="r in scanData.remotes" :key="r.name" class="scan-remote-row">
        <span class="remote-name">{{ r.name }}</span>
        <span class="remote-url">{{ r.fetch_url }}</span>
        <StatusBadge v-if="r.is_mirror" variant="warning" text="Mirror" :show-dot="false" />
      </div>
    </div>
    <div v-if="scanData.branches?.length" class="tracking-tags">
      <StatusBadge v-for="b in scanData.branches" :key="b.name" variant="info" :text="`${b.name} -> ${b.upstream_ref}`" :show-dot="false" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ScanResult } from '@/types/repo'
import SectionTitle from '@/components/common/SectionTitle.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  scanData: ScanResult
}>()
</script>

<style scoped>
.scan-section {
  display: flex;
  flex-direction: column;
}

.info-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-subtitle {
  font-size: 12px;
  color: var(--text-color-secondary, #94A3B8);
}

.scan-remote-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.scan-remote-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: var(--border-radius-sm);
  background: var(--surface-card);
  font-size: 13px;
}

.remote-name {
  font-weight: 600;
  color: var(--accent-primary);
  min-width: 60px;
}

.remote-url {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-color-secondary);
  flex: 1;
}

.tracking-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}
</style>
