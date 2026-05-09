<template>
  <div class="app-breadcrumb">
    <span v-for="(item, idx) in items" :key="idx">
      <router-link v-if="item.path" :to="item.path">{{ item.label }}</router-link>
      <span v-else>{{ item.label }}</span>
      <span v-if="idx < items.length - 1" class="separator"> / </span>
    </span>
  </div>
</template>

<script setup lang="ts">
export interface BreadcrumbItem {
  label: string
  path?: string
}

defineProps<{
  items: BreadcrumbItem[]
}>()
</script>
