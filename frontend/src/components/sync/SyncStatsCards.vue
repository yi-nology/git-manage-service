<template>
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-icon total">
        <el-icon :size="24"><List /></el-icon>
      </div>
      <div class="stat-content">
        <div class="stat-value">{{ stats.total_tasks }}</div>
        <div class="stat-label">总任务数</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon running">
        <el-icon :size="24"><VideoPlay /></el-icon>
      </div>
      <div class="stat-content">
        <div class="stat-value">{{ stats.running_tasks }}</div>
        <div class="stat-label">运行中</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon success">
        <el-icon :size="24"><CircleCheck /></el-icon>
      </div>
      <div class="stat-content">
        <div class="stat-value">{{ stats.today_runs }}</div>
        <div class="stat-label">今日执行</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon failed">
        <el-icon :size="24"><WarningFilled /></el-icon>
      </div>
      <div class="stat-content">
        <div class="stat-value">{{ stats.failed_runs }}</div>
        <div class="stat-label">执行失败</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SyncStats } from '@/types/sync_v2'
import { List, VideoPlay, CircleCheck, WarningFilled } from '@element-plus/icons-vue'

// 同步任务页顶部统计卡片。
defineProps<{
  stats: SyncStats
}>()
</script>

<style scoped lang="less">
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;

  .stat-card {
    background: #fff;
    border-radius: 12px;
    padding: 20px;
    display: flex;
    align-items: center;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
    transition: all 0.2s;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
    }

    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 16px;

      &.total {
        background: linear-gradient(135deg, #ecf5ff 0%, #d9ecff 100%);
        color: #409eff;
      }

      &.running {
        background: linear-gradient(135deg, #f0f9eb 0%, #e1f3d8 100%);
        color: #67c23a;
      }

      &.success {
        background: linear-gradient(135deg, #fdf6ec 0%, #faecd8 100%);
        color: #e6a23c;
      }

      &.failed {
        background: linear-gradient(135deg, #fef0f0 0%, #fde2e2 100%);
        color: #f56c6c;
      }
    }

    .stat-content {
      .stat-value {
        font-size: 28px;
        font-weight: 600;
        color: #303133;
        line-height: 1.2;
      }

      .stat-label {
        font-size: 14px;
        color: #909399;
        margin-top: 4px;
      }
    }
  }
}
</style>
