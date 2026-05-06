<template>
  <span class="status-badge" :class="status">
    <span class="status-dot"></span>
    {{ statusLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
}>()

const statusLabel = computed(() => {
  const labels: Record<string, string> = {
    active: '正常可用',
    invalid: '连接失败',
    expiring: '即将过期',
    expired: '已过期',
    disabled: '已禁用',
  }
  return labels[props.status] || props.status
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.5;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-badge.active {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}
.status-badge.active .status-dot {
  background: #22c55e;
}

.status-badge.invalid {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}
.status-badge.invalid .status-dot {
  background: #ef4444;
}

.status-badge.expiring {
  background: rgba(249, 115, 22, 0.1);
  color: #f97316;
}
.status-badge.expiring .status-dot {
  background: #f97316;
}

.status-badge.expired {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}
.status-badge.expired .status-dot {
  background: #6b7280;
}

.status-badge.disabled {
  background: rgba(107, 114, 128, 0.1);
  color: #6b7280;
}
.status-badge.disabled .status-dot {
  background: #6b7280;
}
</style>
