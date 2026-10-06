<template>
  <div class="cr-review-detail">
    <DetailHeader
      :task="currentTask"
      :retrying="retrying"
      :loading="loading"
      @close="emit('close')"
      @retry="handleRetry"
      @refresh="loadFindings"
    />

    <TaskMetaInfo :task="currentTask" />

    <ReviewSummary :task="currentTask" :findings="findings" />

    <DiffViewer :task="currentTask" :findings="findings" />

    <div class="error-section" v-if="currentTask.error_message">
      <div class="error-card">
        <span class="error-label">错误信息</span>
        <span class="error-text">{{ currentTask.error_message }}</span>
      </div>
    </div>

    <FindingsList :findings="findings" :loading="loading" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toastApiError } from '@/composables/useNotification'
import { ElMessage } from 'element-plus'
import { getReviewTask, listReviewFindings, retryReviewTask, type ReviewTaskDTO, type ReviewFindingDTO } from '@/api/modules/review'
import { usePolling } from '@/composables/usePolling'
import DetailHeader from './cr-detail/DetailHeader.vue'
import TaskMetaInfo from './cr-detail/TaskMetaInfo.vue'
import ReviewSummary from './cr-detail/ReviewSummary.vue'
import DiffViewer from './cr-detail/DiffViewer.vue'
import FindingsList from './cr-detail/FindingsList.vue'

const props = defineProps<{
  task: ReviewTaskDTO
  repoOwner?: string
  repoName?: string
}>()

const emit = defineEmits<{
  close: []
  retried: [task: ReviewTaskDTO]
}>()

const loading = ref(false)
const retrying = ref(false)
const findings = ref<ReviewFindingDTO[]>([])
const currentTask = ref<ReviewTaskDTO>({ ...props.task })

async function loadFindings() {
  loading.value = true
  try {
    const [taskRes, findingsRes] = await Promise.all([
      getReviewTask(currentTask.value.id),
      listReviewFindings(currentTask.value.id),
    ])
    if (taskRes) currentTask.value = taskRes
    findings.value = findingsRes || []
  } catch {
    findings.value = []
  } finally {
    loading.value = false
  }
}

const poll = usePolling({
  interval: 3000,
  tick: async () => {
    if (currentTask.value.status !== 'pending' && currentTask.value.status !== 'running') {
      return false
    }
    const taskRes = await getReviewTask(currentTask.value.id)
    if (taskRes) currentTask.value = taskRes
    if (taskRes && taskRes.status !== 'pending' && taskRes.status !== 'running') {
      const findingsRes = await listReviewFindings(currentTask.value.id)
      findings.value = findingsRes || []
      return false
    }
    return true
  },
  // 查询任务状态偶发失败不终止轮询（保持原行为），由下一轮重试。
  onError: () => true,
})

function startPolling() {
  poll.start()
}

async function handleRetry() {
  retrying.value = true
  try {
    const data: { owner?: string; repo?: string } = {}
    if (props.repoOwner) data.owner = props.repoOwner
    if (props.repoName) data.repo = props.repoName
    const updated = await retryReviewTask(currentTask.value.id, data)
    if (updated) currentTask.value = updated
    ElMessage.success('已重新触发审查')
    emit('retried', currentTask.value)
    startPolling()
  } catch (e: any) {
    toastApiError(e, '重试失败: ', '')
  } finally {
    retrying.value = false
  }
}

onMounted(() => {
  loadFindings()
  startPolling()
})

// usePolling 在组件卸载时自动停止
</script>

<style scoped>
.cr-review-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.error-section { margin-top: 0; }

.error-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 16px;
  background: #FEF2F2;
  border: 1px solid #FECACA;
  border-radius: 8px;
  border-left: 3px solid #DC2626;
}

.error-label { font-size: 12px; font-weight: 600; color: #991B1B; }
.error-text { font-size: 13px; color: #7F1D1D; line-height: 1.5; word-break: break-all; }
</style>
