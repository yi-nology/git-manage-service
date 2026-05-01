<template>
  <div class="form-card">
    <div v-if="title || $slots.header" class="form-card-header">
      <h3 v-if="title" class="form-card-title">{{ title }}</h3>
      <slot name="header"></slot>
    </div>
    <div class="form-card-body">
      <slot></slot>
    </div>
    <div v-if="$slots.footer" class="form-card-footer">
      <slot name="footer"></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
}>(), {
  title: '',
})
</script>

<style scoped>
.form-card {
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.form-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-card-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0;
}

.form-card-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-card-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 8px;
}
</style>
