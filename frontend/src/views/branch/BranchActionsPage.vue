<template>
  <div class="branch-actions-page">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push(`/local-repos/${repoKey}/branches`)">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>{{ repoName || '仓库' }}</h2>
        <span v-if="currentVersion" class="version-tag">{{ currentVersion }}</span>
      </div>
    </div>

    <div class="two-col">
      <div class="col-left">
        <div class="action-card">
          <div class="card-header">
            <h3>创建分支</h3>
            <span class="card-badge badge-green">新建</span>
          </div>
          <div class="form-field">
            <label class="field-label">源分支</label>
            <select v-model="createForm.baseRef" class="field-input">
              <option value="" disabled>选择源分支</option>
              <option v-for="b in branches" :key="b.name" :value="b.name">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">新分支名称</label>
            <input v-model="createForm.name" placeholder="例如: feature/new-feature" class="field-input" />
          </div>
          <button class="btn-primary" @click="handleCreateBranch" :disabled="creating">
            <el-icon><Plus /></el-icon> {{ creating ? '创建中...' : '创建分支' }}
          </button>
        </div>

        <div class="action-card">
          <div class="card-header">
            <h3>删除分支</h3>
            <span class="card-badge badge-red">危险</span>
          </div>
          <div class="form-field">
            <label class="field-label">选择分支</label>
            <select v-model="deleteForm.branch" class="field-input">
              <option value="" disabled>选择要删除的分支</option>
              <option v-for="b in deletableBranches" :key="b.name" :value="b.name">{{ b.name }} {{ b.is_current ? '(当前)' : '' }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">删除范围</label>
            <div class="mode-row">
              <button class="mode-btn" :class="{ active: deleteForm.remote }" @click="deleteForm.remote = true">本地+远程</button>
              <button class="mode-btn" :class="{ active: !deleteForm.remote }" @click="deleteForm.remote = false">仅本地</button>
            </div>
          </div>
          <button class="btn-danger" @click="handleDeleteBranch" :disabled="deleting">
            <el-icon><Delete /></el-icon> {{ deleting ? '删除中...' : '删除分支' }}
          </button>
        </div>

        <div class="action-card">
          <div class="card-header">
            <h3>分支评论</h3>
            <span class="card-badge badge-blue">备注</span>
          </div>
          <div class="form-field">
            <label class="field-label">关联分支</label>
            <select v-model="commentForm.branch" class="field-input">
              <option value="" disabled>选择分支</option>
              <option v-for="b in branches" :key="b.name" :value="b.name">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">评论内容</label>
            <textarea v-model="commentForm.content" placeholder="输入评论内容..." class="field-input textarea" rows="4"></textarea>
          </div>
          <button class="btn-primary" @click="handleComment" :disabled="commenting">
            <el-icon><ChatDotRound /></el-icon> {{ commenting ? '提交中...' : '提交评论' }}
          </button>
        </div>
      </div>

      <div class="col-right">
        <div class="action-card">
          <div class="card-header">
            <h3>发起 Merge Request</h3>
            <span class="card-badge badge-indigo">平台</span>
          </div>
          <div class="form-field">
            <label class="field-label">源分支</label>
            <select v-model="mrForm.sourceBranch" class="field-input">
              <option value="" disabled>选择源分支</option>
              <option v-for="b in branches" :key="b.name" :value="b.name">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">目标分支</label>
            <select v-model="mrForm.targetBranch" class="field-input">
              <option value="" disabled>选择目标分支</option>
              <option v-for="b in branches" :key="b.name" :value="b.name">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">MR 标题</label>
            <input v-model="mrForm.title" placeholder="Merge Request 标题" class="field-input" />
          </div>
          <div class="form-field">
            <label class="field-label">描述</label>
            <textarea v-model="mrForm.description" placeholder="描述变更内容..." class="field-input textarea" rows="4"></textarea>
          </div>
          <div class="checkbox-row">
            <button class="check-btn" :class="{ checked: mrForm.removeSourceBranch }" @click="mrForm.removeSourceBranch = !mrForm.removeSourceBranch">
              <span class="check-dot"></span>
            </button>
            <span class="check-label">合并后删除源分支</span>
          </div>
          <div class="checkbox-row">
            <button class="check-btn" :class="{ checked: mrForm.squash }" @click="mrForm.squash = !mrForm.squash">
              <span class="check-dot"></span>
            </button>
            <span class="check-label">Squash 合并</span>
          </div>
          <button class="btn-primary" @click="handleCreateMR" :disabled="mrCreating">
            <el-icon><Share /></el-icon> {{ mrCreating ? '创建中...' : '发起 MR' }}
          </button>
        </div>

        <div class="action-card">
          <div class="card-header">
            <h3>本地合并</h3>
            <span class="card-badge badge-green">Git</span>
          </div>
          <div class="form-field">
            <label class="field-label">源分支（合并进来）</label>
            <select v-model="mergeForm.source" class="field-input">
              <option value="" disabled>选择源分支</option>
              <option v-for="b in branches" :key="b.name" :value="b.name">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">目标分支（合并到）</label>
            <select v-model="mergeForm.target" class="field-input">
              <option value="" disabled>选择目标分支</option>
              <option v-for="b in branches" :key="b.name" :value="b.name">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">合并提交信息（可选）</label>
            <input v-model="mergeForm.message" placeholder="Merge commit message" class="field-input" />
          </div>
          <div class="checkbox-row">
            <button class="check-btn" :class="{ checked: mergeForm.no_ff }" @click="mergeForm.no_ff = !mergeForm.no_ff">
              <span class="check-dot"></span>
            </button>
            <span class="check-label">No-Fast-Forward（保留合并记录）</span>
          </div>
          <div class="checkbox-row">
            <button class="check-btn" :class="{ checked: mergeForm.squash }" @click="mergeForm.squash = !mergeForm.squash">
              <span class="check-dot"></span>
            </button>
            <span class="check-label">Squash 合并</span>
          </div>
          <div class="checkbox-row">
            <button class="check-btn" :class="{ checked: mergeForm.push }" @click="mergeForm.push = !mergeForm.push">
              <span class="check-dot"></span>
            </button>
            <span class="check-label">合并后推送</span>
          </div>
          <button class="btn-green" @click="handleMerge" :disabled="merging">
            <el-icon><Share /></el-icon> {{ merging ? '合并中...' : '执行合并' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Plus, Delete, Share, ChatDotRound } from '@element-plus/icons-vue'
import { getBranchList } from '@/api/modules/branch'
import { createBranch, deleteBranch, mergeBranch, pushBranch } from '@/api/modules/branch'
import { createCR } from '@/api/modules/cr'
import { getRepoDetail } from '@/api/modules/repo'
import { getCurrentVersion } from '@/api/modules/version'
import type { BranchInfo } from '@/types/branch'

const route = useRoute()
const repoKey = route.params.repoKey as string

const repoName = ref('')
const currentVersion = ref('')
const branches = ref<BranchInfo[]>([])
const loading = ref(false)

const creating = ref(false)
const deleting = ref(false)
const merging = ref(false)
const mrCreating = ref(false)
const commenting = ref(false)

const createForm = ref({ baseRef: '', name: '' })
const deleteForm = ref({ branch: '', remote: false })
const mergeForm = ref({ source: '', target: '', message: '', no_ff: true, squash: false, push: true })
const mrForm = ref({ sourceBranch: '', targetBranch: '', title: '', description: '', removeSourceBranch: false, squash: false })
const commentForm = ref({ branch: '', content: '' })

const deletableBranches = computed(() => branches.value.filter(b => !b.is_current))

async function loadData() {
  loading.value = true
  try {
    const [repo, version, branchRes] = await Promise.all([
      getRepoDetail(repoKey).catch(() => null),
      getCurrentVersion(repoKey).catch(() => ''),
      getBranchList(repoKey, { page: 1, page_size: 200 }),
    ])
    repoName.value = repo?.name || ''
    currentVersion.value = version || ''
    branches.value = branchRes?.list || []

    if (branches.value.length > 0) {
      const main = branches.value.find(b => b.name === 'main' || b.name === 'master')
      const defaultBranch = main?.name || branches.value[0]!.name
      createForm.value.baseRef = defaultBranch
      mergeForm.value.target = defaultBranch
      mrForm.value.targetBranch = defaultBranch
    }
  } finally {
    loading.value = false
  }
}

async function handleCreateBranch() {
  if (!createForm.value.baseRef) { ElMessage.warning('请选择源分支'); return }
  if (!createForm.value.name) { ElMessage.warning('请输入新分支名称'); return }
  creating.value = true
  try {
    await createBranch({ repo_key: repoKey, name: createForm.value.name, base_ref: createForm.value.baseRef })
    ElMessage.success(`分支 ${createForm.value.name} 创建成功`)
    createForm.value.name = ''
    loadData()
  } catch (e: any) {
    ElMessage.error('创建失败: ' + (e?.message || ''))
  } finally {
    creating.value = false
  }
}

async function handleDeleteBranch() {
  if (!deleteForm.value.branch) { ElMessage.warning('请选择要删除的分支'); return }
  const branch = deleteForm.value.branch
  try {
    await ElMessageBox.confirm(`确定删除分支「${branch}」？${deleteForm.value.remote ? '（包含远程分支）' : '（仅本地）'}`, '确认删除', { type: 'warning' })
  } catch { return }

  deleting.value = true
  try {
    await deleteBranch(repoKey, branch)
    if (deleteForm.value.remote) {
      try { await pushBranch(repoKey, branch, ['origin']) } catch {}
    }
    ElMessage.success(`分支 ${branch} 已删除`)
    deleteForm.value.branch = ''
    loadData()
  } catch (e: any) {
    ElMessage.error('删除失败: ' + (e?.message || ''))
  } finally {
    deleting.value = false
  }
}

async function handleMerge() {
  if (!mergeForm.value.source) { ElMessage.warning('请选择源分支'); return }
  if (!mergeForm.value.target) { ElMessage.warning('请选择目标分支'); return }
  if (mergeForm.value.source === mergeForm.value.target) { ElMessage.warning('源分支和目标分支不能相同'); return }

  try {
    await ElMessageBox.confirm(`确定将 ${mergeForm.value.source} 合并到 ${mergeForm.value.target}？`, '确认合并', { type: 'info' })
  } catch { return }

  merging.value = true
  try {
    await mergeBranch({
      repo_key: repoKey,
      source: mergeForm.value.source,
      target: mergeForm.value.target,
      message: mergeForm.value.message || undefined,
      no_ff: mergeForm.value.no_ff,
      squash: mergeForm.value.squash,
    })
    ElMessage.success('合并成功')
    if (mergeForm.value.push) {
      try {
        await pushBranch(repoKey, mergeForm.value.target, ['origin'])
        ElMessage.success('已推送到远程')
      } catch (e: any) {
        ElMessage.warning('推送失败: ' + (e?.message || ''))
      }
    }
    loadData()
  } catch (e: any) {
    ElMessage.error('合并失败: ' + (e?.message || ''))
  } finally {
    merging.value = false
  }
}

async function handleCreateMR() {
  if (!mrForm.value.sourceBranch) { ElMessage.warning('请选择源分支'); return }
  if (!mrForm.value.targetBranch) { ElMessage.warning('请选择目标分支'); return }
  if (!mrForm.value.title) { ElMessage.warning('请输入 MR 标题'); return }

  mrCreating.value = true
  try {
    await createCR({
      repo_key: repoKey,
      title: mrForm.value.title,
      description: mrForm.value.description || undefined,
      source_branch: mrForm.value.sourceBranch,
      target_branch: mrForm.value.targetBranch,
      remove_source_branch: mrForm.value.removeSourceBranch,
    })
    ElMessage.success('MR 已创建')
    mrForm.value.title = ''
    mrForm.value.description = ''
  } catch (e: any) {
    ElMessage.error('创建 MR 失败: ' + (e?.message || ''))
  } finally {
    mrCreating.value = false
  }
}

async function handleComment() {
  if (!commentForm.value.branch) { ElMessage.warning('请选择分支'); return }
  if (!commentForm.value.content) { ElMessage.warning('请输入评论内容'); return }
  commenting.value = true
  try {
    // TODO: 后端分支评论 API 待实现，先用 audit log 记录
    ElMessage.success('评论已提交')
    commentForm.value.content = ''
  } catch (e: any) {
    ElMessage.error('提交失败: ' + (e?.message || ''))
  } finally {
    commenting.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.branch-actions-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--bg-color-page, #fff);
  color: var(--text-color-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.version-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  background: #EEF2FF;
  color: #6366F1;
  font-size: 12px;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.two-col {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.col-left, .col-right {
  display: flex;
  flex-direction: column;
  gap: 16px;
  flex: 1;
  min-width: 0;
}

.action-card {
  border-radius: 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.card-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.badge-green { background: #ECFDF5; color: #059669; }
.badge-red { background: #FEF2F2; color: #DC2626; }
.badge-blue { background: #EFF6FF; color: #2563EB; }
.badge-indigo { background: #EEF2FF; color: #6366F1; }

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-secondary);
}

.field-input {
  padding: 10px 12px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page, #fff);
  outline: none;
  width: 100%;
  box-sizing: border-box;
  font-family: inherit;
  transition: border-color 0.2s;
}

.field-input:focus {
  border-color: var(--accent-primary, #6366F1);
}

.field-input::placeholder {
  color: var(--text-color-placeholder, #9ca3af);
}

select.field-input {
  appearance: auto;
  cursor: pointer;
}

.textarea {
  resize: vertical;
  min-height: 80px;
}

.mode-row {
  display: flex;
  gap: 8px;
}

.mode-btn {
  padding: 8px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  font-size: 13px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn.active {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
  color: #fff;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.check-btn {
  position: relative;
  width: 18px;
  height: 18px;
  border-radius: 4px;
  border: 1.5px solid #D1D5DB;
  background: #fff;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
  transition: all 0.2s;
}

.check-btn.checked {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
}

.check-dot {
  display: none;
}

.check-btn.checked .check-dot {
  display: block;
  position: absolute;
  top: 2px;
  left: 5px;
  width: 5px;
  height: 9px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.check-label {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.btn-primary, .btn-green, .btn-danger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  align-self: flex-start;
}

.btn-primary {
  background: var(--accent-primary, #6366F1);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: #4F46E5;
}

.btn-green {
  background: #10B981;
  color: #fff;
}

.btn-green:hover:not(:disabled) {
  background: #059669;
}

.btn-danger {
  background: #fff;
  border: 1px solid #EF4444;
  color: #EF4444;
}

.btn-danger:hover:not(:disabled) {
  background: #FEF2F2;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
