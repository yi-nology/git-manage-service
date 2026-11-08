<template>
  <div class="config-header">
    <span class="config-title">触发时机</span>
    <div class="toggle-row">
      <button class="toggle-btn" :class="{ active: enabled }" @click="$emit('update:enabled', !enabled)">
        <span class="toggle-dot" :class="{ right: enabled }"></span>
      </button>
      <span class="toggle-label">{{ enabled ? '启用' : '禁用' }}</span>
    </div>
  </div>

  <div class="trigger-grid">
    <div class="trigger-group">
      <div class="trigger-group-title">同步事件</div>
      <label class="checkbox-row" v-for="evt in syncEvents" :key="evt.key">
        <span class="checkbox-box" :class="{ checked: eventsMap[evt.key], 'check-success': evt.key.includes('success'), 'check-danger': evt.key.includes('failure'), 'check-warning': evt.key.includes('conflict') }">
          <span v-if="eventsMap[evt.key]" class="check-icon">&#10003;</span>
        </span>
        <span class="checkbox-label">{{ evt.label }}</span>
      </label>
    </div>
    <div class="trigger-group">
      <div class="trigger-group-title">Webhook 事件</div>
      <label class="checkbox-row" v-for="evt in webhookEvents" :key="evt.key">
        <span class="checkbox-box" :class="{ checked: eventsMap[evt.key], 'check-danger': evt.key.includes('error') }">
          <span v-if="eventsMap[evt.key]" class="check-icon">&#10003;</span>
        </span>
        <span class="checkbox-label">{{ evt.label }}</span>
      </label>
    </div>
    <div class="trigger-group">
      <div class="trigger-group-title">定时任务</div>
      <label class="checkbox-row" v-for="evt in cronEvents" :key="evt.key">
        <span class="checkbox-box" :class="{ checked: eventsMap[evt.key] }">
          <span v-if="eventsMap[evt.key]" class="check-icon">&#10003;</span>
        </span>
        <span class="checkbox-label">{{ evt.label }}</span>
      </label>
    </div>
    <div class="trigger-group">
      <div class="trigger-group-title">备份事件</div>
      <label class="checkbox-row" v-for="evt in backupEvents" :key="evt.key">
        <span class="checkbox-box" :class="{ checked: eventsMap[evt.key], 'check-danger': evt.key.includes('failure') }">
          <span v-if="eventsMap[evt.key]" class="check-icon">&#10003;</span>
        </span>
        <span class="checkbox-label">{{ evt.label }}</span>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
// 通知渠道的触发时机区块：启用开关 + 各事件组勾选展示。
// eventsMap 为页面传入的响应式事件开关表（key -> 是否勾选）。
defineProps<{
  enabled: boolean
  eventsMap: Record<string, boolean>
}>()

defineEmits<{
  'update:enabled': [value: boolean]
}>()

const syncEvents = [
  { key: 'sync_success', label: '同步成功' },
  { key: 'sync_failure', label: '同步失败' },
  { key: 'sync_conflict', label: '同步冲突' },
]
const webhookEvents = [
  { key: 'webhook_received', label: 'Webhook 接收' },
  { key: 'webhook_error', label: 'Webhook 错误' },
]
const cronEvents = [
  { key: 'cron_triggered', label: '定时任务触发' },
]
const backupEvents = [
  { key: 'backup_success', label: '备份成功' },
  { key: 'backup_failure', label: '备份失败' },
]
</script>

<style scoped>
.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-btn {
  position: relative;
  width: 40px;
  height: 22px;
  border-radius: 11px;
  border: none;
  background: var(--border-color, #d1d5db);
  cursor: pointer;
  transition: background 0.2s;
  padding: 0;
}

.toggle-btn.active {
  background: var(--accent-primary);
}

.toggle-dot {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  transition: left 0.2s;
}

.toggle-dot.right {
  left: 20px;
}

.toggle-label {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.trigger-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.trigger-group {
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.trigger-group-title {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-secondary);
  margin-bottom: 2px;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-color-primary);
}

.checkbox-box {
  width: 14px;
  height: 14px;
  border-radius: 3px;
  border: 1.5px solid var(--border-color, #d1d5db);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.15s;
}

.checkbox-box.checked {
  border-color: var(--accent-primary);
  background: var(--accent-primary);
}

.checkbox-box.check-success.checked {
  border-color: #10B981;
  background: #10B981;
}

.checkbox-box.check-danger.checked {
  border-color: #EF4444;
  background: #EF4444;
}

.checkbox-box.check-warning.checked {
  border-color: #F59E0B;
  background: #F59E0B;
}

.check-icon {
  font-size: 10px;
  color: #fff;
  line-height: 1;
}

.checkbox-label {
  font-size: 13px;
}
</style>
