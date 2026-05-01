<template>
  <div class="data-table">
    <div class="table-header">
      <div
        v-for="col in columns"
        :key="col.key"
        class="th"
        :style="{ width: col.width || 'auto', flex: col.width ? undefined : (col.flex || 1) }"
      >
        {{ col.label }}
      </div>
    </div>
    <template v-if="!loading && data.length > 0">
      <div v-for="(row, index) in data" :key="rowKey ? row[rowKey] : index" class="table-row">
        <div
          v-for="col in columns"
          :key="col.key"
          class="td"
          :style="{ width: col.width || 'auto', flex: col.width ? undefined : (col.flex || 1) }"
        >
          <slot :name="`cell-${col.key}`" :row="row" :index="index">
            {{ row[col.key] }}
          </slot>
        </div>
        <div v-if="$slots['row-actions']" class="td td-actions">
          <slot name="row-actions" :row="row" :index="index"></slot>
        </div>
      </div>
    </template>
    <div v-if="loading" class="table-loading">
      <slot name="loading">
        <LoadingState text="加载中..." />
      </slot>
    </div>
    <div v-if="!loading && data.length === 0" class="table-empty">
      <slot name="empty">
        <EmptyState icon="Folder" title="暂无数据" description="没有符合条件的记录" />
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import LoadingState from './LoadingState.vue'
import EmptyState from './EmptyState.vue'

export interface TableColumn {
  key: string
  label: string
  width?: string
  flex?: number | string
}

withDefaults(defineProps<{
  columns: TableColumn[]
  data: any[]
  rowKey?: string
  loading?: boolean
}>(), {
  rowKey: 'id',
  loading: false,
})
</script>

<style scoped>
.data-table {
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: var(--surface-card);
}

.th {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--text-color-secondary);
  flex-shrink: 0;
}

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
  transition: background var(--transition-fast);
  min-height: 48px;
}

.table-row:nth-child(even) {
  background: var(--surface-card);
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--border-color-extra-light);
}

.td {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  flex-shrink: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.td-actions {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
  width: auto;
  flex: none;
}

.table-loading,
.table-empty {
  padding: 40px 20px;
}
</style>
