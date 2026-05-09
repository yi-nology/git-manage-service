<template>
  <button
    class="text-btn"
    :class="[
      variant ? `text-btn--${variant}` : '',
      { 'text-btn--small': small, 'text-btn--disabled': disabled }
    ]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <el-icon v-if="icon" class="text-btn-icon"><component :is="icon" /></el-icon>
    <slot />
  </button>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  variant?: 'primary' | 'success' | 'warning' | 'danger' | 'outline' | 'purple'
  icon?: Component
  small?: boolean
  disabled?: boolean
}>()

defineEmits<{ click: [] }>()
</script>

<style scoped>
.text-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-color-regular);
  transition: all var(--transition-fast);
  white-space: nowrap;
  line-height: 1;
}

.text-btn:hover {
  background: var(--border-color-extra-light);
}

.text-btn--small {
  padding: 4px 8px;
  font-size: var(--font-size-xs);
}

.text-btn--primary {
  color: var(--primary-color);
}
.text-btn--primary:hover {
  background: rgba(99, 102, 241, 0.08);
}

.text-btn--success {
  color: var(--success-color);
}
.text-btn--success:hover {
  background: rgba(16, 185, 129, 0.08);
}

.text-btn--warning {
  color: var(--warning-color);
}
.text-btn--warning:hover {
  background: rgba(245, 158, 11, 0.08);
}

.text-btn--danger {
  color: var(--danger-color);
}
.text-btn--danger:hover {
  background: rgba(239, 68, 68, 0.08);
}

.text-btn--outline {
  border-color: var(--border-color);
  color: var(--text-color-regular);
}
.text-btn--outline:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.text-btn--purple {
  color: #8B5CF6;
}
.text-btn--purple:hover {
  background: rgba(139, 92, 246, 0.08);
}

.text-btn--disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.text-btn-icon {
  font-size: 14px;
}
</style>
