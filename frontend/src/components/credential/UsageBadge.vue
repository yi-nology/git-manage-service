<template>
  <div class="usage-badge" @click="$emit('click', $event)">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
      <polyline points="22,4 12,14.01 9,11.01"/>
    </svg>
    <span>
      {{ usages.totalRepoCount + usages.totalProviderCount }} 处引用
    </span>
  </div>
</template>

<script setup lang="ts">
import type { CredentialUsages } from '@/types/credential'

defineProps<{
  usages: CredentialUsages
}>()

defineEmits<{
  click: [e: MouseEvent]
}>()
</script>

<style scoped>
.usage-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
  cursor: pointer;
  transition: background 0.2s;
}

.usage-badge:hover {
  background: rgba(139, 92, 246, 0.2);
}

.usage-badge svg {
  width: 14px;
  height: 14px;
}
</style>
