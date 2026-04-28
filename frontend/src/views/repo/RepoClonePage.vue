<template>
  <div class="clone-page">
    <div class="title-row">
      <div class="title-left">
        <button class="back-btn" @click="$router.push('/local-repos')">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2 class="page-title">克隆远程仓库</h2>
      </div>
    </div>

    <div class="steps-bar">
      <div class="step" :class="{ active: step >= 0, current: step === 0 }">
        <span class="step-num">1</span>
        <span class="step-text">仓库地址</span>
      </div>
      <div class="step-line" :class="{ filled: step >= 1 }"></div>
      <div class="step" :class="{ active: step >= 1, current: step === 1 }">
        <span class="step-num">2</span>
        <span class="step-text">认证配置</span>
      </div>
      <div class="step-line" :class="{ filled: step >= 2 }"></div>
      <div class="step" :class="{ active: step >= 2, current: step === 2 }">
        <span class="step-num">3</span>
        <span class="step-text">确认克隆</span>
      </div>
    </div>

    <div class="form-card" v-if="!taskId">
      <div class="form-field">
        <label class="field-label">远程仓库地址</label>
        <div class="proto-row">
          <button class="proto-btn" :class="{ active: urlMode === 'ssh' }" @click="switchMode('ssh')">SSH</button>
          <button class="proto-btn" :class="{ active: urlMode === 'https' }" @click="switchMode('https')">HTTPS</button>
        </div>
        <div class="url-input-row">
          <span class="url-tag">{{ urlMode === 'ssh' ? 'SSH' : 'HTTPS' }}</span>
          <input v-model="form.remote_url" :placeholder="urlPlaceholder" @blur="onUrlBlur" class="field-input url-field" :class="{ 'is-error': urlError }" />
        </div>
        <span v-if="urlError" class="field-error">{{ urlError }}</span>
        <span v-else class="field-hint">格式: {{ urlMode === 'ssh' ? 'git@host:user/repo.git' : 'https://host/user/repo.git' }}</span>
      </div>

      <div class="form-field">
        <label class="field-label">认证凭证</label>
        <CredentialSelector v-model="form.credential_id" :url="form.remote_url" placeholder="选择凭证（公开仓库可不选）" />
      </div>

      <div class="form-field">
        <label class="field-label">本地路径</label>
        <div class="input-with-btn">
          <input v-model="form.local_path" placeholder="/path/to/clone/destination" class="field-input" />
          <button class="browse-btn" @click="handleBrowse">浏览</button>
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">仓库名称</label>
        <input v-model="form.name" placeholder="可选，默认从 URL 推断" class="field-input" />
      </div>

      <div class="form-actions">
        <button class="action-pill pill-outline" @click="$router.push('/local-repos')">取消</button>
        <button class="action-pill pill-primary" @click="handleClone" :disabled="cloning || !form.remote_url || !form.local_path">
          <el-icon><ArrowRight /></el-icon>
          {{ cloning ? '克隆中...' : '开始克隆' }}
        </button>
      </div>
    </div>

    <div class="form-card" v-if="taskId">
      <div class="section-header">
        <span class="section-title">克隆进度</span>
        <span class="status-pill" :class="'status-' + taskStatus">{{ statusLabel }}</span>
      </div>

      <div v-if="progressLines.length" class="progress-logs">
        <div v-for="(line, i) in progressLines" :key="i" class="log-line">{{ line }}</div>
      </div>

      <div v-if="taskStatus === 'done'" class="result-row">
        <span class="result-text">克隆成功！</span>
        <button class="action-pill pill-primary" @click="$router.push('/local-repos')">查看仓库列表</button>
      </div>

      <div v-if="taskStatus === 'failed'" class="result-row">
        <span class="result-text result-text--error">克隆失败: {{ taskError }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { cloneRepo, getCloneTask, selectDirectory } from '@/api/modules/repo'
import type { CloneRepoReq } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { validateGitRemoteUrl, detectGitProtocol, extractRepoName, convertGitUrl } from '@/utils/git'

type UrlMode = 'ssh' | 'https'

const route = useRoute()

const step = ref(0)
const cloning = ref(false)
const taskId = ref('')
const taskStatus = ref('')
const taskError = ref('')
const progressLines = ref<string[]>([])
let pollTimer: ReturnType<typeof setInterval> | null = null
const urlError = ref('')
const urlMode = ref<UrlMode>('ssh')

const form = ref<CloneRepoReq>({
  remote_url: '',
  local_path: '',
  name: '',
  credential_id: undefined,
})

onMounted(() => {
  const q = route.query
  if (q.url) {
    form.value.remote_url = q.url as string
    const proto = detectGitProtocol(form.value.remote_url)
    if (proto === 'ssh') urlMode.value = 'ssh'
    else if (proto === 'http') urlMode.value = 'https'
    if (!form.value.name) {
      const name = extractRepoName(form.value.remote_url)
      if (name) form.value.name = name
    }
  }
  if (q.provider_config_id) form.value.provider_config_id = Number(q.provider_config_id)
  if (q.platform_owner) form.value.platform_owner = q.platform_owner as string
  if (q.platform_repo) form.value.platform_repo = q.platform_repo as string
})

const urlPlaceholder = computed(() => {
  return urlMode.value === 'ssh'
    ? 'git@github.com:user/repo.git'
    : 'https://github.com/user/repo.git'
})

const statusLabel = computed(() => {
  const map: Record<string, string> = { running: '克隆中...', done: '已完成', failed: '失败' }
  return map[taskStatus.value] || taskStatus.value || '等待中'
})

function switchMode(mode: UrlMode) {
  if (urlMode.value !== mode && form.value.remote_url) {
    form.value.remote_url = convertGitUrl(form.value.remote_url, mode)
  }
  urlMode.value = mode
  urlError.value = ''
}

function onUrlBlur() {
  const url = form.value.remote_url
  if (!url) { urlError.value = ''; return }
  const proto = detectGitProtocol(url)
  if (proto === 'ssh') urlMode.value = 'ssh'
  else if (proto === 'http') urlMode.value = 'https'
  urlError.value = validateGitRemoteUrl(url)
  if (!form.value.name) {
    const name = extractRepoName(url)
    if (name) form.value.name = name
  }
}

async function handleBrowse() {
  try {
    const result = await selectDirectory('选择克隆目录')
    if (!result.cancelled && result.path) {
      form.value.local_path = result.path
    }
  } catch { /* ignore */ }
}

async function handleClone() {
  if (!form.value.remote_url || !form.value.local_path) {
    ElMessage.warning('请填写远程 URL 和本地路径')
    return
  }
  const err = validateGitRemoteUrl(form.value.remote_url)
  if (err) { urlError.value = err; return }

  cloning.value = true
  progressLines.value = []
  taskError.value = ''
  taskStatus.value = 'running'

  try {
    const result = await cloneRepo(form.value)
    taskId.value = result.task_id
    startPolling()
  } catch (e: any) {
    taskStatus.value = 'failed'
    taskError.value = e?.message || '克隆启动失败'
  } finally {
    cloning.value = false
  }
}

function startPolling() {
  pollTimer = setInterval(async () => {
    if (!taskId.value) return
    try {
      const task = await getCloneTask(taskId.value)
      taskStatus.value = task.status
      progressLines.value = task.progress || []
      if (task.error) taskError.value = task.error
      if (task.status === 'done' || task.status === 'failed') {
        stopPolling()
        if (task.status === 'done') ElMessage.success('仓库克隆成功')
      }
    } catch { /* ignore */ }
  }, 1500)
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
}

onUnmounted(() => { stopPolling() })
</script>

<style scoped>
.clone-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.title-row { display: flex; align-items: center; justify-content: space-between; }
.title-left { display: flex; align-items: center; gap: 12px; }

.back-btn {
  display: inline-flex; align-items: center; gap: 6px; font-size: 13px;
  color: var(--text-color-secondary); background: none;
  border: 1px solid var(--border-color, #e5e7eb); border-radius: 6px;
  padding: 6px 12px; cursor: pointer; transition: all 0.2s;
}
.back-btn:hover { border-color: var(--accent-primary, #6366F1); color: var(--accent-primary, #6366F1); }

.page-title { margin: 0; font-size: 24px; font-weight: 600; color: var(--text-color-primary); }

.steps-bar {
  display: flex; align-items: center; justify-content: center; gap: 0;
  padding: 16px 24px; background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb); border-radius: 12px;
}

.step { display: flex; align-items: center; gap: 8px; opacity: 0.4; transition: opacity 0.2s; }
.step.active { opacity: 1; }

.step-num {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: 12px;
  font-size: 12px; font-weight: 600; background: var(--border-color, #e5e7eb); color: var(--text-color-secondary);
}
.step.active .step-num { background: var(--accent-primary, #6366F1); color: #fff; }

.step-text { font-size: 13px; font-weight: 500; color: var(--text-color-primary); }

.step-line { width: 80px; height: 2px; background: var(--border-color, #e5e7eb); margin: 0 8px; transition: background 0.2s; }
.step-line.filled { background: var(--accent-primary, #6366F1); }

.form-card {
  border-radius: 12px; background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb); padding: 24px;
  display: flex; flex-direction: column; gap: 16px;
}

.form-field { display: flex; flex-direction: column; gap: 8px; }

.field-label { font-size: 13px; font-weight: 500; color: var(--text-color-primary); }

.field-input {
  padding: 10px 12px; border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px; font-size: 13px; color: var(--text-color-primary);
  background: var(--bg-color-page, #fff); outline: none; width: 100%; box-sizing: border-box;
  transition: border-color 0.2s;
}
.field-input:focus { border-color: var(--accent-primary, #6366F1); }
.field-input.is-error { border-color: #EF4444; }
.field-input::placeholder { color: var(--text-color-placeholder, #9ca3af); }

.proto-row { display: flex; gap: 8px; }

.proto-btn {
  padding: 8px 14px; border-radius: 8px; border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff); font-size: 13px; color: var(--text-color-secondary);
  cursor: pointer; transition: all 0.2s;
}
.proto-btn.active { background: var(--accent-primary, #6366F1); border-color: var(--accent-primary, #6366F1); color: #fff; }

.url-input-row { display: flex; align-items: center; gap: 8px; }

.url-tag {
  padding: 4px 10px; border-radius: 6px; background: #ECFDF5; color: #10B981;
  font-size: 12px; flex-shrink: 0;
}

.url-field { flex: 1; }

.input-with-btn { display: flex; gap: 8px; }
.input-with-btn .field-input { flex: 1; }

.browse-btn {
  padding: 10px 16px; border-radius: 6px; background: #EEF2FF;
  border: none; font-size: 13px; color: var(--accent-primary, #6366F1);
  cursor: pointer; transition: opacity 0.2s; white-space: nowrap;
}
.browse-btn:hover { opacity: 0.8; }

.field-error { font-size: 12px; color: #EF4444; }
.field-hint { font-size: 12px; color: var(--text-color-secondary, #94A3B8); }

.section-header { display: flex; align-items: center; justify-content: space-between; }
.section-title { font-size: 16px; font-weight: 600; color: var(--text-color-primary); }

.status-pill {
  display: inline-block; padding: 4px 8px; border-radius: 9999px;
  font-size: 11px; font-weight: 500; text-align: center;
}
.status-running { background: #EEF2FF; color: #6366F1; }
.status-done { background: #ECFDF5; color: #059669; }
.status-failed { background: #FEF2F2; color: #DC2626; }

.form-actions { display: flex; justify-content: flex-end; gap: 8px; }

.action-pill {
  display: inline-flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 500;
  padding: 12px 24px; border-radius: 8px; border: none; cursor: pointer; transition: all 0.2s;
}
.action-pill:disabled { opacity: 0.5; cursor: not-allowed; }

.pill-primary { background: var(--accent-primary, #6366F1); color: #fff; }
.pill-primary:hover:not(:disabled) { opacity: 0.9; }

.pill-outline { background: transparent; border: 1px solid var(--border-color, #e5e7eb); color: var(--text-color-primary); }
.pill-outline:hover { border-color: var(--accent-primary, #6366F1); color: var(--accent-primary, #6366F1); }

.progress-logs {
  background: #F8F9FC; border-radius: 6px; padding: 12px;
  max-height: 300px; overflow-y: auto;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 12px;
}

.log-line { line-height: 1.6; color: var(--text-color-primary); }

.result-row { display: flex; align-items: center; justify-content: space-between; }
.result-text { font-size: 14px; font-weight: 500; color: #10B981; }
.result-text--error { color: #EF4444; }

@media (max-width: 768px) {
  .steps-bar { padding: 12px; }
  .step-text { display: none; }
}
</style>
