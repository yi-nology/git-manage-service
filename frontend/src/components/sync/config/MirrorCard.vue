<template>
  <div class="mirror-card" :class="`type-${mirror.mirror_type}`">
    <div class="card-header">
      <div class="card-title">
        <el-tag :type="mirror.mirror_type === 'pull' ? 'primary' : 'warning'" size="large">
          <el-icon v-if="mirror.mirror_type === 'pull'"><Download /></el-icon>
          <el-icon v-else><Upload /></el-icon>
          {{ mirror.mirror_type.toUpperCase() }}
        </el-tag>
        <el-tag :type="statusType" size="small">
          {{ statusLabel }}
        </el-tag>
      </div>
      <div class="card-actions">
        <el-switch
          :model-value="mirror.enabled"
          @change="$emit('toggle-enabled', mirror)"
          :loading="updating"
        />
      </div>
    </div>

    <div class="card-body">
      <div class="info-row">
        <span class="info-label">远程 URL</span>
        <span class="info-value mono">{{ mirror.remote_url }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Remote 名称</span>
        <span class="info-value">{{ mirror.remote_name }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">分支过滤</span>
        <span class="info-value">{{ mirror.branch_filter || '全部分支' }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">同步间隔</span>
        <span class="info-value">{{ mirror.cron_expr || `${mirror.sync_interval}s` }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Git 选项</span>
        <div class="git-options">
          <el-tag v-if="mirror.git_force" size="small" type="danger" effect="plain">--force</el-tag>
          <el-tag v-if="mirror.git_prune" size="small" effect="plain">--prune</el-tag>
          <el-tag v-if="mirror.git_tags" size="small" effect="plain">--tags</el-tag>
          <el-tag v-if="!mirror.git_force && !mirror.git_prune && !mirror.git_tags" size="small" type="info">无</el-tag>
        </div>
      </div>
      <div class="info-row">
        <span class="info-label">上次同步</span>
        <span class="info-value">{{ mirror.last_sync_at ? formatTime(mirror.last_sync_at) : '从未' }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">下次同步</span>
        <span class="info-value" :class="{ 'syncing': syncing }">
          {{ syncing ? '同步中...' : (mirror.next_sync_at ? formatTime(mirror.next_sync_at) : '-') }}
        </span>
      </div>
      <div v-if="mirror.last_error" class="info-row error-row">
        <span class="info-label">错误</span>
        <span class="info-value">{{ mirror.last_error }}</span>
      </div>
    </div>

    <div class="card-footer">
      <el-button-group>
        <el-button type="primary" size="small" @click="$emit('sync', mirror)" :loading="syncing">
          <el-icon><Refresh /></el-icon>
          同步
        </el-button>
        <el-button size="small" @click="$emit('show-logs', mirror)">
          <el-icon><Document /></el-icon>
          日志
        </el-button>
        <el-button size="small" @click="$emit('edit', mirror)">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button type="danger" size="small" @click="$emit('delete', mirror)">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </el-button-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Refresh, Download, Upload, Document, Edit, Delete } from '@element-plus/icons-vue'
import type { MirrorDTO } from '@/types/mirror'
import { MIRROR_STATUS_MAP } from '@/types/mirror'

const props = defineProps<{
  mirror: MirrorDTO
  syncing?: boolean
  updating?: boolean
}>()

defineEmits<{
  sync: [mirror: MirrorDTO]
  'show-logs': [mirror: MirrorDTO]
  edit: [mirror: MirrorDTO]
  delete: [mirror: MirrorDTO]
  'toggle-enabled': [mirror: MirrorDTO]
}>()

const statusType = computed(() => {
  const type = MIRROR_STATUS_MAP[props.mirror.status]?.type
  return (type as 'success' | 'warning' | 'danger' | 'info') || 'info'
})

const statusLabel = computed(() => {
  return MIRROR_STATUS_MAP[props.mirror.status]?.label || props.mirror.status
})

function formatTime(timeStr: string): string {
  if (!timeStr) return '-'
  const d = new Date(timeStr)
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.mirror-card {
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  overflow: hidden;
  background: var(--fill-color-blank);
  transition: all 0.2s ease;
}

.mirror-card:hover {
  box-shadow: var(--box-shadow-light);
  border-color: var(--primary-color-lighter);
}

.mirror-card.type-pull {
  border-left: 4px solid var(--el-color-primary);
}

.mirror-card.type-push {
  border-left: 4px solid var(--el-color-warning);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: var(--bg-color-page);
  border-bottom: 1px solid var(--border-color-lighter);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-title .el-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
}

.card-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.info-label {
  min-width: 80px;
  font-size: 12px;
  color: var(--text-color-secondary);
  flex-shrink: 0;
}

.info-value {
  flex: 1;
  font-size: 13px;
  color: var(--text-color-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-value.mono {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 12px;
}

.info-value.syncing {
  color: var(--el-color-primary);
  font-weight: 600;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.git-options {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.error-row {
  padding: 8px 12px;
  background: var(--el-color-danger-lighter);
  border-radius: var(--border-radius-base);
  margin: 0 -4px;
}

.error-row .info-label,
.error-row .info-value {
  color: var(--el-color-danger);
}

.card-footer {
  padding: 12px 16px 16px;
  border-top: 1px solid var(--border-color-lighter);
}

.card-footer .el-button-group {
  width: 100%;
}

.card-footer .el-button {
  flex: 1;
}

@media (max-width: 768px) {
  .card-footer .el-button-group {
    flex-wrap: wrap;
  }

  .card-footer .el-button {
    flex: 1 1 calc(50% - 2px);
  }
}
</style>
