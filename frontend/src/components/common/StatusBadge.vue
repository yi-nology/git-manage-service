<template>
  <span class="status-badge" :class="`status-badge--${variant}`">
    <span v-if="showDot" class="status-dot"></span>
    {{ text || label }}
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'
  text?: string
  label?: string
  showDot?: boolean
}>(), {
  variant: 'default',
  text: '',
  label: '',
  showDot: true,
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: var(--font-size-xs);
  font-weight: normal;
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-badge--success {
  background: #ECFDF5;
  color: #059669;
}
.status-badge--success .status-dot {
  background: var(--success-color);
}

.status-badge--danger {
  background: #FEF2F2;
  color: #DC2626;
}
.status-badge--danger .status-dot {
  background: var(--danger-color);
}

.status-badge--warning {
  background: #FFFBEB;
  color: #D97706;
}
.status-badge--warning .status-dot {
  background: var(--warning-color);
}

.status-badge--info {
  background: var(--accent-bg);
  color: var(--primary-color);
}
.status-badge--info .status-dot {
  background: var(--primary-color);
}

.status-badge--running {
  background: #EFF6FF;
  color: #3B82F6;
}
.status-badge--running .status-dot {
  background: #3B82F6;
}

.status-badge--default {
  background: var(--accent-bg);
  color: var(--text-color-secondary);
}
.status-badge--default .status-dot {
  background: var(--text-color-placeholder);
}
</style>
