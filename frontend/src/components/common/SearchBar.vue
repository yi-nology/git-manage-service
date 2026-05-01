<template>
  <div class="search-bar">
    <el-icon class="search-icon"><Search /></el-icon>
    <input
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      :placeholder="placeholder"
      class="search-input"
    />
    <el-icon v-if="modelValue" class="clear-icon" @click="$emit('update:modelValue', '')"><Close /></el-icon>
  </div>
</template>

<script setup lang="ts">
import { Search, Close } from '@element-plus/icons-vue'

withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
}>(), {
  modelValue: '',
  placeholder: '搜索...',
})

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<style scoped>
.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 12px 16px;
}

.search-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: var(--font-size-sm);
  color: var(--text-color-primary);
  font-family: var(--font-family);
}

.search-input::placeholder {
  color: var(--text-color-placeholder);
}

.clear-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
  cursor: pointer;
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.clear-icon:hover {
  color: var(--text-color-primary);
}
</style>
