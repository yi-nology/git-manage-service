<template>
  <div class="page-header">
    <div class="header-left">
      <button v-if="showBack" class="back-btn" @click="handleBack">
        <el-icon><ArrowLeft /></el-icon> 返回
      </button>
      <div v-if="title" class="title-group">
        <h2 class="page-title">{{ title }}</h2>
        <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
      </div>
      <slot name="title-suffix"></slot>
    </div>
    <div class="header-actions">
      <slot name="actions"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  title?: string
  subtitle?: string
  backRoute?: string
  showBack?: boolean
}>(), {
  title: '',
  subtitle: '',
  backRoute: '',
  showBack: false,
})

const router = useRouter()

function handleBack() {
  if (props.backRoute) {
    router.push(props.backRoute)
  } else {
    router.back()
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-color-secondary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.back-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.title-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
