<template>
  <span class="type-badge" :class="type">
    <svg v-if="type === 'ssh_key'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <rect x="2" y="5" width="20" height="14" rx="2"/>
      <line x1="6" y1="9" x2="6" y2="15"/>
      <line x1="10" y1="9" x2="10" y2="15"/>
      <line x1="14" y1="9" x2="14" y2="15"/>
      <line x1="18" y1="9" x2="18" y2="15"/>
    </svg>
    <svg v-else-if="type === 'http_basic'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
      <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
    </svg>
    <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M12 2L2 7l10 5 10-5-10-5z"/>
      <path d="M2 17l10 5 10-5"/>
      <path d="M2 12l10 5 10-5"/>
    </svg>
    {{ typeLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  type: string
}>()

const typeLabel = computed(() => {
  const labels: Record<string, string> = {
    ssh_key: 'SSH 密钥',
    http_basic: 'HTTP 密码',
    http_token: 'HTTP Token',
    platform_token: '平台 Token',
  }
  return labels[props.type] || props.type
})
</script>

<style scoped>
.type-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.type-badge svg {
  width: 14px;
  height: 14px;
}

.type-badge.ssh_key {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.type-badge.http_basic {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.type-badge.http_token,
.type-badge.platform_token {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}
</style>
