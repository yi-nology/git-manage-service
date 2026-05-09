<template>
  <div class="tab-bar" :class="[`tab-bar--${style}`]">
    <div
      v-for="tab in tabs"
      :key="tab.key"
      class="tab-item"
      :class="{ active: modelValue === tab.key }"
      @click="$emit('update:modelValue', tab.key)"
    >
      <el-icon v-if="tab.icon" class="tab-icon"><component :is="tab.icon" /></el-icon>
      <span class="tab-label">{{ tab.label }}</span>
      <span v-if="tab.count !== undefined" class="tab-count">{{ tab.count }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

export interface TabItem {
  key: string
  label: string
  icon?: Component
  count?: number
}

defineProps<{
  tabs: TabItem[]
  modelValue: string
  style?: 'underline' | 'pill'
}>()

defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<style scoped>
.tab-bar {
  display: flex;
  gap: 2px;
  margin-bottom: var(--spacing-md);
}

.tab-bar--underline {
  border-bottom: 1px solid var(--border-color);
  gap: 0;
}

.tab-bar--underline .tab-item {
  padding: 10px 16px;
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-color-secondary);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all var(--transition-fast);
  margin-bottom: -1px;
}

.tab-bar--underline .tab-item:hover {
  color: var(--text-color-primary);
}

.tab-bar--underline .tab-item.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

.tab-bar--pill {
  gap: 4px;
  padding: 4px;
  background: var(--bg-color);
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
}

.tab-bar--pill .tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--border-radius-sm);
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.tab-bar--pill .tab-item:hover {
  color: var(--text-color-primary);
  background: var(--border-color-extra-light);
}

.tab-bar--pill .tab-item.active {
  color: var(--primary-color);
  background: var(--bg-color-page);
  box-shadow: var(--box-shadow-sm);
}

.tab-icon {
  font-size: 14px;
}

.tab-count {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--border-color);
  color: var(--text-color-secondary);
  font-weight: 600;
}

.tab-bar--pill .tab-item.active .tab-count {
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-color);
}

@media (max-width: 768px) {
  .tab-bar--underline,
  .tab-bar--pill {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .tab-bar--pill .tab-item {
    padding: 6px 12px;
    font-size: var(--font-size-xs);
  }
}
</style>
