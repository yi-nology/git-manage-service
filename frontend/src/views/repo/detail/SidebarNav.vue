<template>
  <div class="left-nav">
    <div class="sidebar-card">
      <div
        v-for="item in items"
        :key="item.key"
        class="sidebar-item"
        :class="{ active: activeTab === item.key && !(item as any).route }"
        @click="$emit('select', item.key)"
      >
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

interface SidebarItem {
  key: string
  label: string
  icon: Component
  route?: boolean
}

defineProps<{
  items: SidebarItem[]
  activeTab: string
}>()

defineEmits<{
  select: [key: string]
}>()
</script>

<style scoped>
.left-nav {
  width: 200px;
  flex-shrink: 0;
}

.sidebar-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--surface-sidebar);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  padding: 8px;
  height: calc(100vh - 156px);
  overflow-y: auto;
  position: sticky;
  top: calc(var(--header-height) + 16px);
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 10px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--text-color-primary);
  font-size: var(--font-size-sm);
  min-height: 34px;
  white-space: nowrap;
}

.sidebar-item:hover {
  background: var(--bg-color-page);
}

.sidebar-item.active {
  background: var(--accent-bg);
  color: var(--accent-primary);
}

.sidebar-item.active .el-icon {
  color: var(--accent-primary);
}

.sidebar-item .el-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
}

@media (max-width: 1024px) {
  .left-nav {
    width: 200px;
  }
}

@media (max-width: 768px) {
  .left-nav {
    width: 100%;
  }

  .sidebar-card {
    height: auto;
    max-height: 300px;
    flex-direction: row;
    flex-wrap: wrap;
    position: static;
  }

  .sidebar-item {
    flex-shrink: 0;
  }
}
</style>
