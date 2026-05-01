<template>
  <button
    class="action-pill"
    :class="[`action-pill--${variant}`, { 'action-pill--small': small }]"
    :disabled="disabled"
    @click="$emit('click', $event)"
  >
    <el-icon v-if="icon" :size="small ? 12 : 14"><component :is="icon" /></el-icon>
    <slot></slot>
  </button>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  variant?: 'primary' | 'green' | 'amber' | 'danger' | 'outline'
  icon?: any
  disabled?: boolean
  small?: boolean
}>(), {
  variant: 'primary',
  disabled: false,
  small: false,
})

defineEmits<{
  click: [event: MouseEvent]
}>()
</script>

<style scoped>
.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--border-radius-md);
  border: none;
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
  font-family: var(--font-family);
}

.action-pill--small {
  padding: 4px 10px;
  font-size: var(--font-size-xs);
  gap: 4px;
}

.action-pill--primary {
  background: var(--primary-color);
  color: #fff;
}
.action-pill--primary:hover:not(:disabled) {
  background: var(--primary-color-hover);
}

.action-pill--green {
  background: var(--success-color);
  color: #fff;
}
.action-pill--green:hover:not(:disabled) {
  opacity: 0.9;
}

.action-pill--amber {
  background: var(--warning-color);
  color: #fff;
}
.action-pill--amber:hover:not(:disabled) {
  opacity: 0.9;
}

.action-pill--danger {
  background: var(--danger-color);
  color: #fff;
}
.action-pill--danger:hover:not(:disabled) {
  opacity: 0.9;
}

.action-pill--outline {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-color-secondary);
}
.action-pill--outline:hover:not(:disabled) {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.action-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
