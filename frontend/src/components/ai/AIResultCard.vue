<template>
  <div :class="['ai-result-card', `ai-result-card--${type}`]">
    <div class="ai-result-icon">
      <el-icon :size="18" v-if="type === 'success'"><SuccessFilled /></el-icon>
      <el-icon :size="18" v-else-if="type === 'warning'"><WarningFilled /></el-icon>
      <el-icon :size="18" v-else-if="type === 'error'"><CircleCloseFilled /></el-icon>
      <el-icon :size="18" v-else-if="type === 'info'"><InfoFilled /></el-icon>
      <el-icon :size="18" v-else><MagicStick /></el-icon>
    </div>
    <div class="ai-result-content">
      <div class="ai-result-title">{{ title }}</div>
      <div v-if="description" class="ai-result-description">{{ description }}</div>
      <div v-if="$slots.default" class="ai-result-body">
        <slot></slot>
      </div>
      <div v-if="actions.length || $slots.actions" class="ai-result-actions">
        <slot name="actions"></slot>
        <el-button
          v-for="action in actions"
          :key="action.key"
          :type="action.type || 'default'"
          size="small"
          @click="onAction(action)"
        >{{ action.label }}</el-button>
      </div>
    </div>
    <div v-if="closable" class="ai-result-close">
      <el-button text size="small" @click="$emit('close')">
        <el-icon :size="14"><Close /></el-icon>
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { SuccessFilled, WarningFilled, CircleCloseFilled, InfoFilled, MagicStick, Close } from '@element-plus/icons-vue'

export interface ResultAction {
  key: string
  label: string
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'default'
  payload?: unknown
}

const props = withDefaults(defineProps<{
  type?: 'success' | 'warning' | 'error' | 'info' | 'ai'
  title: string
  description?: string
  actions?: ResultAction[]
  closable?: boolean
}>(), {
  actions: () => [],
})

const emit = defineEmits<{
  action: [action: ResultAction]
  close: []
}>()

const onAction = (action: ResultAction) => {
  emit('action', action)
}
</script>

<style lang="scss" scoped>
.ai-result-card {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  background: var(--bg-color-secondary);
  border-left: 3px solid var(--el-color-info);
  
  &--success {
    border-left-color: var(--el-color-success);
    .ai-result-icon { color: var(--el-color-success); }
  }
  
  &--warning {
    border-left-color: var(--el-color-warning);
    .ai-result-icon { color: var(--el-color-warning); }
  }
  
  &--error {
    border-left-color: var(--el-color-danger);
    .ai-result-icon { color: var(--el-color-danger); }
  }
  
  &--info {
    border-left-color: var(--el-color-info);
    .ai-result-icon { color: var(--el-color-info); }
  }
  
  &--ai {
    border-left-color: #6366F1;
    .ai-result-icon { color: #6366F1; }
  }
  
  &-icon {
    flex-shrink: 0;
    margin-top: 2px;
  }
  
  &-content {
    flex: 1;
    min-width: 0;
  }
  
  &-title {
    font-weight: 500;
    margin-bottom: 4px;
  }
  
  &-description {
    font-size: 13px;
    color: var(--text-color-secondary);
    margin-bottom: 8px;
  }
  
  &-body {
    margin-bottom: 8px;
  }
  
  &-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  
  &-close {
    flex-shrink: 0;
    margin-left: auto;
  }
}
</style>
