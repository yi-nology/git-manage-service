<template>
  <h3 class="section-title"><slot>{{ title }}</slot></h3>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
}>(), {
  title: '',
})
</script>

<style scoped>
.section-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0;
  font-family: var(--font-family);
}
</style>
