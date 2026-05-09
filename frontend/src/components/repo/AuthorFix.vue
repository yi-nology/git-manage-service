<template>
  <div class="author-fix-section">
    <div v-loading="configLoading" class="config-bar">
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
        <template v-if="allIdentities.length > 0">
          <el-select v-model="selectedIdentityId" placeholder="选择身份" size="small" style="width: 200px" :disabled="configSaving" @change="handleConfigChange">
            <el-option label="使用全局默认" :value="(null as any)" />
            <el-option v-for="id in allIdentities" :key="id.id" :label="`${id.canonicalName} (${id.canonicalEmail})`" :value="id.id" />
          </el-select>
        </template>
        <router-link v-else to="/settings/author">
          <el-button size="small" type="primary" link>请先创建 Git 作者身份</el-button>
        </router-link>
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
          <el-button type="primary" size="small" :loading="scanLoading" @click="doScan">扫描仓库</el-button>
        </div>
      </div>

      <div v-if="hasScanned && !scanLoading" class="scan-summary">
        扫描了 {{ totalCommits }} 个提交，发现 <strong>{{ scanResult.length }}</strong> 个作者不匹配
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
          <el-table-column label="匹配方式" width="100">
            <template #default="{ row }">
              <el-tag :type="row.matchType === 'exact' ? 'success' : 'warning'" size="small">
                {{ row.matchType === 'exact' ? '名称+邮箱' : '仅邮箱' }}
              </el-tag>
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
      <el-empty v-else-if="hasScanned && !scanLoading" description="没有发现作者不匹配的提交" :image-size="60" />
      <div v-else-if="!scanLoading" class="empty-scan">
        <el-icon :size="32" color="var(--text-color-placeholder)"><Search /></el-icon>
        <span>点击「扫描仓库」检查作者不匹配的提交</span>
      </div>
    </div>

    <div v-if="selectedCommits.length > 0" class="bottom-bar">
      <div class="bottom-left">
        <span>已选 {{ selectedCommits.length }} 个提交</span>
        <el-button text size="small" @click="handleSelection([])">取消选择</el-button>
      </div>
      <div class="bottom-right">
        <el-select v-model="pushRemote" size="small" placeholder="不推送" style="width: 160px">
          <el-option label="不推送" value="" />
          <el-option v-for="r in remotes" :key="r" :label="r" :value="r" />
        </el-select>
        <el-button type="primary" size="small" @click="handleFixSelected">修复选中</el-button>
      </div>
    </div>

    <div v-if="taskId" class="progress-card">
      <div class="progress-header">
        <SectionTitle title="任务进度" />
        <el-button v-if="taskStatus === 'success' || taskStatus === 'failed'" text size="small" @click="dismissTask" aria-label="关闭">关闭</el-button>
      </div>
      <el-alert v-if="taskStatus === 'failed'" :title="taskError" type="error" show-icon :closable="false" />
      <el-alert v-else-if="taskStatus === 'success'" title="修复完成" type="success" show-icon :closable="false" />
      <template v-else>
        <el-progress :percentage="0" :indeterminate="true" />
        <div class="task-logs" aria-live="polite">
          <div v-for="(log, i) in taskLogs" :key="i" class="log-line">{{ log }}</div>
        </div>
      </template>
    </div>

    <div v-if="showAIPanel" class="ai-panel">
      <div class="ai-panel-header">
        <SectionTitle title="AI 助手" />
        <div class="ai-panel-actions">
          <el-button v-if="hasScanned && scanResult.length > 0" size="small" :loading="aiLoading" @click="doAnalyzeScan">解读扫描结果</el-button>
          <el-button v-if="selectedCommits.length > 0" size="small" :loading="aiLoading" @click="doAssessRisk">风险评估</el-button>
          <el-button text size="small" @click="showAIPanel = false">收起</el-button>
        </div>
      </div>

      <div v-if="aiAnalysis" class="ai-card">
        <div class="ai-card-title">扫描解读</div>
        <div class="ai-card-content" v-html="formatAI(aiAnalysis)"></div>
      </div>

      <div v-if="aiRisk" class="ai-card">
        <div class="ai-card-title">
          风险评估
          <el-tag :type="aiRisk.riskLevel === 'low' ? 'success' : aiRisk.riskLevel === 'medium' ? 'warning' : 'danger'" size="small">
            {{ aiRisk.riskLevel === 'low' ? '低风险' : aiRisk.riskLevel === 'medium' ? '中等风险' : '高风险' }}
          </el-tag>
        </div>
        <div class="ai-card-content">
          <p>{{ aiRisk.summary }}</p>
          <div v-if="aiRisk.recommendations.length > 0" class="ai-recommendations">
            <div v-for="(r, i) in aiRisk.recommendations" :key="i" class="ai-rec-item">{{ r }}</div>
          </div>
        </div>
      </div>

      <div class="ai-chat">
        <div class="ai-chat-messages" ref="chatContainer">
          <div v-if="chatMessages.length === 0" class="ai-chat-empty">问任何关于 Git 作者管理的问题...</div>
          <div v-for="(m, i) in chatMessages" :key="i" class="ai-msg" :class="'ai-msg--' + m.role">
            <div class="ai-msg-content">{{ m.content }}</div>
          </div>
        </div>
        <div class="ai-chat-input">
          <el-input v-model="chatInput" placeholder="输入问题..." size="small" @keyup.enter="doSendChat" :disabled="chatLoading">
            <template #append>
              <el-button :loading="chatLoading" @click="doSendChat" aria-label="发送">发送</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </div>

    <div v-if="!showAIPanel" class="ai-toggle">
      <el-button size="small" type="primary" link @click="showAIPanel = true">
        <el-icon><ChatDotRound /></el-icon> AI 助手
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { User, Search, ChatDotRound } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import SectionTitle from '@/components/common/SectionTitle.vue'
import { useAuthorIdentity, useAuthorFix, useAuthorAI } from '@/composables/useAuthor'
import type { MismatchedCommit as _MismatchedCommit } from '@/api/modules/author'

const props = withDefaults(defineProps<{
  repoKey: string
  remotes?: string[]
}>(), {
  remotes: () => ['origin'],
})

const { identities: allIdentities, loadIdentities } = useAuthorIdentity()
const {
  repoConfig, configLoading, scanResult, scanLoading, totalCommits, selectedCommits,
  taskId, taskStatus, taskLogs, taskError,
  loadRepoConfig, setConfig, scan, fixAll, fixSelected, handleSelection,
} = useAuthorFix(props.repoKey)

const selectedIdentityId = ref<number | null>(null)
const pushRemote = ref('')
const configSaving = ref(false)
const hasScanned = ref(false)
const dismissed = ref(false)

function handleConfigChange(val: number | null) {
  configSaving.value = true
  setConfig(val).finally(() => { configSaving.value = false })
}

async function doScan() {
  hasScanned.value = false
  await scan()
  hasScanned.value = true
}

async function handleFixAll() {
  try {
    await ElMessageBox.confirm(
      '即将重写所有匹配提交的作者信息。此操作不可恢复！',
      '确认一键修复',
      { confirmButtonText: '确认修复', cancelButtonText: '取消', type: 'warning' }
    )
  } catch { return }
  dismissed.value = false
  fixAll(pushRemote.value || '')
}

async function handleFixSelected() {
  if (selectedCommits.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `即将修复 ${selectedCommits.value.length} 个提交的作者信息。此操作不可恢复！`,
      '确认修复',
      { confirmButtonText: '确认修复', cancelButtonText: '取消', type: 'warning' }
    )
  } catch { return }
  dismissed.value = false
  fixSelected(pushRemote.value || '')
}

function dismissTask() {
  dismissed.value = true
  taskId.value = ''
}

const showAIPanel = ref(false)
const chatInput = ref('')
const chatContainer = ref<HTMLElement | null>(null)
const {
  aiLoading, aiAnalysis, aiRisk,
  chatMessages, chatLoading,
  analyzeScan, assessRisk, sendChat, clearChat: _clearChat,
} = useAuthorAI(props.repoKey)

async function doAnalyzeScan() {
  await analyzeScan({ commits: scanResult.value, totalCommits: totalCommits.value, matchCount: scanResult.value.length })
}

async function doAssessRisk() {
  const commits = selectedCommits.value.length > 0 ? selectedCommits.value : scanResult.value
  await assessRisk(commits)
}

async function doSendChat() {
  if (!chatInput.value.trim()) return
  const msg = chatInput.value
  chatInput.value = ''
  await sendChat(msg)
  await nextTick()
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight
  }
}

function formatAI(text: string) {
  return text.replace(/\n/g, '<br>')
}

onMounted(async () => {
  await Promise.all([loadIdentities(), loadRepoConfig()])
  if (repoConfig.value?.identityId) {
    selectedIdentityId.value = repoConfig.value.identityId
  }
})
</script>

<style scoped>
.author-fix-section { display: flex; flex-direction: column; gap: var(--spacing-lg); }

.config-bar { display: flex; justify-content: space-between; align-items: center; padding: var(--spacing-md) var(--spacing-lg); background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); }
.config-left { display: flex; gap: var(--spacing-sm); align-items: center; }
.config-icon { width: 36px; height: 36px; border-radius: var(--border-radius-md); background: var(--accent-bg); display: flex; align-items: center; justify-content: center; color: var(--primary-color); }
.config-info { display: flex; flex-direction: column; gap: 2px; }
.config-label { font-size: var(--font-size-sm); font-weight: 600; color: var(--text-color-primary); }
.config-value { font-size: var(--font-size-xs); color: var(--text-color-secondary); }
.config-right { display: flex; align-items: center; }

.scan-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); overflow: hidden; }
.scan-header { display: flex; justify-content: space-between; align-items: center; padding: var(--spacing-md) var(--spacing-lg); border-bottom: 1px solid var(--border-color); }
.scan-title-row { display: flex; gap: var(--spacing-sm); align-items: center; }
.scan-actions { display: flex; gap: var(--spacing-sm); }
.scan-summary { padding: var(--spacing-sm) var(--spacing-lg); font-size: var(--font-size-xs); color: var(--text-color-secondary); border-bottom: 1px solid var(--border-color-light); }
.scan-table { padding: 0; }
.empty-scan { display: flex; flex-direction: column; gap: var(--spacing-sm); align-items: center; padding: 48px var(--spacing-lg); color: var(--text-color-placeholder); font-size: var(--font-size-sm); }

.bottom-bar { display: flex; justify-content: space-between; align-items: center; padding: var(--spacing-sm) var(--spacing-lg); background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); font-size: var(--font-size-sm); color: var(--text-color-primary); }
.bottom-left { display: flex; gap: var(--spacing-sm); align-items: center; }
.bottom-right { display: flex; gap: var(--spacing-sm); align-items: center; }

.progress-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); padding: var(--spacing-lg); display: flex; flex-direction: column; gap: var(--spacing-sm); }
.progress-header { display: flex; justify-content: space-between; align-items: center; }
.task-logs { max-height: 200px; overflow-y: auto; background: var(--bg-color-page); border: 1px solid var(--border-color); border-radius: var(--border-radius-md); padding: var(--spacing-sm); font-family: var(--font-family-mono, monospace); font-size: var(--font-size-xs); }
.log-line { padding: 2px 0; color: var(--text-color-secondary); }

.mono { font-family: var(--font-family-mono, 'SF Mono', Monaco, Menlo, Consolas, monospace); font-size: var(--font-size-xs); }

.ai-toggle { padding: var(--spacing-xs) 0; }
.ai-panel { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: var(--border-radius-lg); padding: var(--spacing-lg); display: flex; flex-direction: column; gap: var(--spacing-md); }
.ai-panel-header { display: flex; justify-content: space-between; align-items: center; }
.ai-panel-actions { display: flex; gap: var(--spacing-sm); align-items: center; }

.ai-card { padding: var(--spacing-md); background: var(--bg-color-page); border: 1px solid var(--border-color-light); border-radius: var(--border-radius-md); }
.ai-card-title { display: flex; gap: var(--spacing-sm); align-items: center; font-size: var(--font-size-sm); font-weight: 600; color: var(--text-color-primary); margin-bottom: var(--spacing-sm); }
.ai-card-content { font-size: var(--font-size-sm); color: var(--text-color-secondary); line-height: 1.6; }
.ai-recommendations { margin-top: var(--spacing-sm); display: flex; flex-direction: column; gap: 4px; }
.ai-rec-item { padding-left: 12px; position: relative; font-size: var(--font-size-xs); }
.ai-rec-item::before { content: '•'; position: absolute; left: 0; color: var(--primary-color); }

.ai-chat { display: flex; flex-direction: column; gap: var(--spacing-sm); }
.ai-chat-messages { max-height: 300px; overflow-y: auto; padding: var(--spacing-sm); background: var(--bg-color-page); border: 1px solid var(--border-color-light); border-radius: var(--border-radius-md); min-height: 60px; }
.ai-chat-empty { text-align: center; padding: var(--spacing-lg); color: var(--text-color-placeholder); font-size: var(--font-size-sm); }
.ai-msg { margin-bottom: var(--spacing-sm); }
.ai-msg--user { text-align: right; }
.ai-msg--user .ai-msg-content { display: inline-block; background: var(--primary-color); color: #fff; padding: 6px 12px; border-radius: var(--border-radius-md); font-size: var(--font-size-sm); max-width: 80%; text-align: left; }
.ai-msg--assistant .ai-msg-content { background: var(--surface-card); border: 1px solid var(--border-color); padding: 8px 12px; border-radius: var(--border-radius-md); font-size: var(--font-size-sm); line-height: 1.6; white-space: pre-wrap; }
.ai-chat-input { display: flex; gap: var(--spacing-sm); }
</style>
