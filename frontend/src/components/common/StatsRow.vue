<template>
  <div class="stats-row">
    <div v-for="(stat, i) in stats" :key="i" class="stat-card">
      <div class="stat-top">
        <el-icon v-if="stat.icon" :size="20" :color="stat.color || 'var(--primary-color)'" class="stat-icon">
          <component :is="stat.icon" />
        </el-icon>
        <span class="stat-label">{{ stat.label }}</span>
      </div>
      <span class="stat-value">{{ stat.value }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface StatItem {
  label: string
  value: string | number
  icon?: any
  color?: string
}

defineProps<{
  stats: StatItem[]
}>()
</script>

<style scoped>
.stats-row {
  display: flex;
  gap: 16px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 20px;
  background: var(--surface-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
  min-width: 140px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.stat-top {
  display: flex;
  align-items: center;
  gap: 6px;
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-color-primary);
}
</style>
