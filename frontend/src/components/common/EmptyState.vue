<template>
  <div class="empty-state">
    <div class="empty-icon">
      <el-icon :size="32"><component :is="iconComponent" /></el-icon>
    </div>
    <div class="empty-text">{{ title }}</div>
    <div v-if="description" class="empty-desc">{{ description }}</div>
    <slot name="action"></slot>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Folder } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  icon?: string
  title: string
  description?: string
}>(), {
  icon: 'Folder',
  description: '',
})

const iconMap: Record<string, any> = {
  Folder,
}

const iconComponent = computed(() => iconMap[props.icon] || Folder)
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 24px;
  gap: 12px;
}

.empty-icon {
  color: var(--text-color-placeholder);
  margin-bottom: 4px;
}

.empty-text {
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--text-color-primary);
}

.empty-desc {
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
}
</style>
