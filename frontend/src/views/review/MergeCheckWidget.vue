<template>
  <FormCard class="merge-check-widget" title="合并检查">
    <template #header>
      <StatusBadge
        v-if="result"
        :variant="result.mergeable ? 'success' : 'danger'"
        :text="result.mergeable ? '通过' : '已阻止'"
      />
    </template>
    <LoadingState v-if="loading && !result" text="检查中..." />
    <template v-else>
      <div class="mc-alert" v-if="result && !result.mergeable">
        <strong>{{ blockedCount }} 个问题阻止了此次合并</strong>
        <p>请解决严重级别的问题后方可继续合并操作。</p>
      </div>
      <div class="mc-pass-info" v-if="result?.mergeable">
        <strong>所有检查已通过</strong>
        <p>此 MR 可以安全合并。</p>
      </div>
      <div class="mc-checks" v-if="result?.checks?.length">
        <div v-for="check in result.checks" :key="check.check_type" class="mc-check-item">
          <el-icon :size="20" :color="check.status === 'success' ? '#10B981' : '#EF4444'">
            <CircleCheckFilled v-if="check.status === 'success'" />
            <CircleCloseFilled v-else />
          </el-icon>
          <div class="mc-check-info">
            <div class="mc-check-title">
              {{ checkLabel(check.check_type) }} — {{ check.status === 'success' ? '通过' : '未通过' }}
            </div>
            <div class="mc-check-detail">{{ check.message }}</div>
          </div>
        </div>
      </div>
    </template>
    <template #footer>
      <ActionPill variant="outline" :disabled="loading" @click="loadData">重新检查</ActionPill>
    </template>
  </FormCard>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { CircleCheckFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import { checkMerge, type MergeCheckDTO } from '@/api/modules/review'
import FormCard from '@/components/common/FormCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import LoadingState from '@/components/common/LoadingState.vue'

const props = defineProps<{
  repoKey: string
  mrIid: string
  commitSha?: string
}>()

const loading = ref(false)
const result = ref<MergeCheckDTO | null>(null)

const blockedCount = computed(() => {
  if (!result.value?.checks) return 0
  return result.value.checks.filter(c => c.status !== 'success').length
})

function checkLabel(type: string) {
  const m: Record<string, string> = { code_review: '代码审查', pipeline: '流水线' }
  return m[type] || type
}

async function loadData() {
  loading.value = true
  try {
    result.value = await checkMerge({
      repo_key: props.repoKey,
      mr_iid: props.mrIid,
      commit_sha: props.commitSha,
    })
  } catch (e) { console.error(e) } finally { loading.value = false }
}

onMounted(loadData)
</script>

<style scoped>
.mc-alert { background: #FEF2F2; padding: 12px; border-radius: 8px; }
.mc-alert strong { font-size: 14px; color: #991B1B; }
.mc-alert p { font-size: 13px; color: #B91C1C; margin: 4px 0 0; }
.mc-pass-info { background: #F0FDF4; padding: 12px; border-radius: 8px; }
.mc-pass-info strong { font-size: 14px; color: #166534; }
.mc-pass-info p { font-size: 13px; color: #166534; margin: 4px 0 0; }
.mc-checks { display: flex; flex-direction: column; gap: 12px; }
.mc-check-item { display: flex; gap: 12px; align-items: flex-start; }
.mc-check-info { display: flex; flex-direction: column; gap: 2px; }
.mc-check-title { font-size: 14px; font-weight: 600; }
.mc-check-detail { font-size: 12px; color: var(--el-text-color-secondary); }
</style>
