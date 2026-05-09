<template>
  <div class="section-header">
    <div class="section-header-left">
      <SectionTitle :title="title" />
      <span v-if="subtitle" class="section-subtitle">{{ subtitle }}</span>
    </div>
    <div class="section-header-right">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import SectionTitle from './SectionTitle.vue'

defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<style scoped>
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.section-header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.section-header-right {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.section-subtitle {
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
}

@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }

  .section-header-right {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>
