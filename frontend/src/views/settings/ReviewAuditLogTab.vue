<template>
  <div class="audit-log-tab">
    <div class="filter-bar">
      <div class="filter-fields">
        <el-input v-model="filters.task_id" placeholder="任务 ID" clearable style="width:140px" @clear="loadLogs" @keyup.enter="loadLogs" />
        <el-select v-model="filters.action" placeholder="操作类型" clearable style="width:140px" @change="loadLogs">
          <el-option label="创建任务" value="create_task" />
          <el-option label="执行审查" value="execute_review" />
          <el-option label="重试任务" value="retry_task" />
          <el-option label="更新状态" value="update_status" />
        </el-select>
        <el-select v-model="filters.status" placeholder="状态" clearable style="width:120px" @change="loadLogs">
          <el-option label="成功" value="success" />
          <el-option label="失败" value="failed" />
          <el-option label="进行中" value="running" />
        </el-select>
      </div>
      <ActionPill variant="primary" @click="loadLogs">刷新</ActionPill>
    </div>

    <DataTable :columns="columns" :data="logs" :loading="loading" row-key="id">
      <template #cell-task_id="{ row }">
        <span style="font-weight:500">#{{ row.task_id }}</span>
      </template>
      <template #cell-action="{ row }">
        <StatusBadge :variant="actionVariant(row.action)" :text="row.action" />
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :variant="statusVariant(row.status)" :text="row.status" />
      </template>
      <template #cell-duration="{ row }">
        <span v-if="row.duration">{{ row.duration }}ms</span>
        <span v-else>-</span>
      </template>
      <template #cell-error_message="{ row }">
        <span v-if="row.error_message" class="error-text">{{ row.error_message }}</span>
        <span v-else>-</span>
      </template>
      <template #cell-created_at="{ row }">
        {{ formatTime(row.created_at) }}
      </template>
    </DataTable>

    <div v-if="total > pageSize" class="pagination-bar">
      <PaginationBar :total="total" :page="page" :page-size="pageSize" @change="handlePageChange" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import DataTable from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'
import { listReviewAuditLogs } from '@/api/modules/review'
import type { ReviewAuditLogDTO } from '@/api/modules/review'
import type { TableColumn } from '@/components/common/DataTable.vue'

const loading = ref(false)
const logs = ref<ReviewAuditLogDTO[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filters = reactive({
  task_id: '',
  action: '',
  status: '',
})

const columns: TableColumn[] = [
  { key: 'task_id', label: '任务 ID', width: '90px' },
  { key: 'action', label: '操作', width: '130px' },
  { key: 'status', label: '状态', width: '90px' },
  { key: 'duration', label: '耗时', width: '90px' },
  { key: 'error_message', label: '错误信息', flex: 1 },
  { key: 'created_at', label: '时间', width: '170px' },
]

function actionVariant(action: string): string {
  const map: Record<string, string> = {
    create_task: 'blue',
    execute_review: 'purple',
    retry_task: 'amber',
    update_status: 'teal',
  }
  return map[action] || 'default'
}

function statusVariant(status: string): string {
  const map: Record<string, string> = {
    success: 'success',
    failed: 'danger',
    running: 'warning',
  }
  return map[status] || 'default'
}

function formatTime(t: string): string {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN', { hour12: false })
}

async function loadLogs() {
  loading.value = true
  try {
    const taskIdNum = filters.task_id ? parseInt(filters.task_id) : undefined
    const res = await listReviewAuditLogs({
      task_id: taskIdNum,
      action: filters.action || undefined,
      status: filters.status || undefined,
      page: page.value,
      page_size: pageSize.value,
    }) as any
    logs.value = res.logs || []
    total.value = res.total || res.pagination?.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number) {
  page.value = p
  loadLogs()
}

onMounted(() => { loadLogs() })
</script>

<style scoped>
.audit-log-tab {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.filter-fields {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.error-text {
  color: var(--danger-color, #F56C6C);
  font-size: var(--font-size-xs);
}

.pagination-bar {
  display: flex;
  justify-content: center;
}
</style>
