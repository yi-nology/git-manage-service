<template>
  <span class="purpose-badge" :class="type">
    {{ purposeLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  type: string
}>()

const purposeLabel = computed(() => {
  const labels: Record<string, string> = {
    ssh_key: 'Git 远程',
    http_basic: 'Git 远程 / API',
    http_token: 'Git 远程 / API',
    platform_token: '平台 API',
  }
  return labels[props.type] || 'Git 远程'
})
</script>

<style scoped>
.purpose-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  background: var(--bg-color-tertiary);
  color: var(--text-color-secondary);
  border: 1px solid var(--border-color);
}
</style>
