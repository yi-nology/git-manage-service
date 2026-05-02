<template>
  <div class="author-fix-section">
    <div class="config-bar">
      <div class="config-left">
        <div class="config-icon"><el-icon :size="20"><User /></el-icon></div>
        <div class="config-info">
          <div class="config-label">仓库作者身份</div>
          <div class="config-value">
            <template v-if="repoConfig?.identity">
              {{ repoConfig.identity.canonicalName }} ({{ repoConfig.identity.canonicalEmail }})
              — {{ repoConfig.source === 'repo' ? '仓库覆盖' : '使用全局默认' }}
            </template>
            <template v-else>未配置</template>
          </div>
        </div>
      </div>
      <div class="config-right">
        <el-select v-model="selectedIdentityId" placeholder="选择身份" size="small" style="width: 200px" clearable @change="handleConfigChange">
          <el-option label="使用全局默认" :value="null" />
          <el-option v-for="id in allIdentities" :key="id.id" :label="`${id.canonicalName} (${id.canonicalEmail})`" :value="id.id" />
        </el-select>
      </div>
    </div>

    <div class="scan-card">
      <div class="scan-header">
        <div class="scan-title-row">
          <SectionTitle title="作者扫描结果" />
          <el-tag v-if="scanResult.length > 0" type="warning" size="small">{{ scanResult.length }} 条待修复</el-tag>
        </div>
        <div class="scan-actions">
          <el-button v-if="scanResult.length > 0" type="danger" size="small" @click="handleFixAll" :disabled="!!taskId">一键修复全部</el-button>
          <el-button type="primary" size="small" :loading="scanLoading" @click="scan">扫描仓库</el-button>
        </div>
      </div>

      <div v-if="scanResult.length > 0" class="scan-table">
        <el-table :data="scanResult" @selection-change="handleSelection" border size="small" max-height="400">
          <el-table-column type="selection" width="45" />
          <el-table-column label="Hash" width="80">
            <template #default="{ row }"><span class="mono">{{ row.shortHash }}</span></template>
          </el-table-column>
          <el-table-column prop="message" label="提交信息" min-width="200" show-overflow-tooltip />
          <el-table-column prop="authorName" label="当前作者" width="120" />
          <el-table-column label="当前邮箱" width="200">
            <template #default="{ row }"><span class="mono">{{ row.authorEmail }}</span></template>
          </el-table-column>
          <el-table-column label="匹配" width="80">
            <template #default="{ row }">
              <el-tag :type="row.matchType === 'exact' ? 'success' : 'warning'" size="small">{{ row.matchType === 'exact' ? '精确' : 'Email' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="目标" width="200">
            <template #default="{ row }"><span class="mono">{{ row.targetName }} &lt;{{ row.targetEmail }}&gt;</span></template>
          </el-table-column>
          <el-table-column label="日期" width="110">
            <template #default="{ row }">{{ row.date?.substring(0, 10) }}</template>
          </el-table-column>
        </el-table>
      </div>
      <el-empty v-else-if="!scanLoading" description="点击「扫描仓库」检查非主名提交" :image-size="60" />
    </div>

    <div v-if="selectedCommits.length > 0" class="bottom-bar">
      <span>已选 {{ selectedCommits.length }} 个提交</span>
      <div class="bottom-right">
        <el-select v-model="pushRemote" size="small" placeholder="不推送" clearable style="width: 160px">
          <el-option label="不推送" value="" />
          <el-option label="origin" value="origin" />
        </el-select>
        <el-button type="primary" size="small" @click="handleFixSelected">修复选中</el-button>
      </div>
    </div>

    <div v-if="taskId" class="progress-card">
      <SectionTitle title="任务进度" />
      <el-alert v-if="taskStatus === 'failed'" :title="taskError" type="error" show-icon :closable="false" />
      <el-alert v-else-if="taskStatus === 'success'" title="修复完成" type="success" show-icon :closable="false" />
      <template v-else>
        <el-progress :percentage="100" :indeterminate="true" />
        <div class="task-logs">
          <div v-for="(log, i) in taskLogs" :key="i" class="log-line">{{ log }}</div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { User } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import SectionTitle from '@/components/common/SectionTitle.vue'
import { useAuthorIdentity, useAuthorFix } from '@/composables/useAuthor'
import type { MismatchedCommit } from '@/api/modules/author'

const props = defineProps<{ repoKey: string }>()

const { identities: allIdentities, loadIdentities } = useAuthorIdentity()
const {
  repoConfig, configLoading, scanResult, scanLoading, totalCommits, selectedCommits,
  taskId, taskStatus, taskLogs, taskError,
  loadRepoConfig, setConfig, scan: doScan, fixAll, fixSelected, handleSelection,
} = useAuthorFix(props.repoKey)

const selectedIdentityId = ref<number | null>(null)
const pushRemote = ref('')

function handleConfigChange(val: number | null) {
  setConfig(val)
}

async function scan() {
  await doScan()
}

async function handleFixAll() {
  try {
    await ElMessageBox.confirm(
      '即将重写所有匹配提交的作者信息，此操作不可恢复！',
      '确认一键修复',
      { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' }
    )
  } catch { return }
  fixAll(pushRemote.value || '')
}

async function handleFixSelected() {
  if (selectedCommits.value.length === 0) {
    ElMessage.warning('请先选择要修复的提交')
    return
  }
  try {
    await ElMessageBox.confirm(
      `即将修复 ${selectedCommits.value.length} 个提交，此操作不可恢复！`,
      '确认修复',
      { confirmButtonText: '确认', cancelButtonText: '取消', type: 'warning' }
    )
  } catch { return }
  fixSelected(pushRemote.value || '')
}

onMounted(async () => {
  await loadIdentities()
  await loadRepoConfig()
  if (repoConfig.value?.identityId) {
    selectedIdentityId.value = repoConfig.value.identityId
  }
})
</script>

<style scoped>
.author-fix-section { display: flex; flex-direction: column; gap: 20px; }

.config-bar { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: 12px; }
.config-left { display: flex; gap: 12px; align-items: center; }
.config-icon { width: 36px; height: 36px; border-radius: 8px; background: #EEF2FF; display: flex; align-items: center; justify-content: center; color: #6366F1; }
.config-info { display: flex; flex-direction: column; gap: 2px; }
.config-label { font-size: 13px; font-weight: 600; color: var(--text-color-primary); }
.config-value { font-size: 12px; color: var(--text-color-secondary); }

.scan-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: 12px; overflow: hidden; }
.scan-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.scan-title-row { display: flex; gap: 8px; align-items: center; }
.scan-actions { display: flex; gap: 8px; }
.scan-table { padding: 0; }

.bottom-bar { display: flex; justify-content: space-between; align-items: center; padding: 12px 20px; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: 12px; font-size: 13px; color: var(--text-color-primary); }
.bottom-right { display: flex; gap: 8px; align-items: center; }

.progress-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: 12px; padding: 20px; display: flex; flex-direction: column; gap: 12px; }
.task-logs { max-height: 200px; overflow-y: auto; background: #F8FAFC; border: 1px solid var(--border-color); border-radius: 8px; padding: 12px; font-family: monospace; font-size: 12px; }
.log-line { padding: 2px 0; color: var(--text-color-secondary); }

.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 12px; }
</style>
