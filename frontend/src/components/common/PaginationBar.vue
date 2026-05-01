<template>
  <div class="pagination-bar">
    <span class="pag-info">共 {{ total }} 条</span>
    <div class="pag-controls">
      <button class="pag-btn" :disabled="currentPage <= 1" @click="$emit('update:currentPage', currentPage - 1)">
        <el-icon><ArrowLeft /></el-icon>
      </button>
      <span class="pag-page">{{ currentPage }} / {{ totalPages }}</span>
      <button class="pag-btn" :disabled="currentPage >= totalPages" @click="$emit('update:currentPage', currentPage + 1)">
        <el-icon><ArrowRight /></el-icon>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  total: number
  currentPage: number
  pageSize?: number
}>(), {
  pageSize: 10,
})

defineEmits<{
  'update:currentPage': [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
</script>

<style scoped>
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  padding: 8px 0;
}

.pag-info {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
}

.pag-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pag-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--border-radius-sm);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  cursor: pointer;
  color: var(--text-color-regular);
  transition: all var(--transition-fast);
}

.pag-btn:hover:not(:disabled) {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.pag-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pag-page {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  min-width: 60px;
  text-align: center;
}
</style>
