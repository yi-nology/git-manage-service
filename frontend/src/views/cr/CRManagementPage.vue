<template>
  <div class="cr-page-wrapper">
    <div class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="$router.push(`/local-repos/${repoKey}`)">
          <el-icon><ArrowLeft /></el-icon> 返回
        </button>
        <h2>{{ repoName || '仓库' }}</h2>
        <span v-if="currentVersion" class="version-tag">{{ currentVersion }}</span>
      </div>
      <div class="header-actions">
        <button class="action-pill action-pill--primary" @click="handleSync" :disabled="syncing">
          <el-icon><Refresh /></el-icon> {{ syncing ? '同步中...' : '同步 CR' }}
        </button>
      </div>
    </div>

    <div class="cr-layout">
      <RepoSidebar :repo-key="repoKey" active-key="cr" />
      <div class="cr-content">
        <div class="cr-management-page">
          <div class="filter-bar">
            <button v-for="f in filters" :key="f.key" class="filter-btn" :class="{ active: activeFilter === f.key }" @click="activeFilter = f.key">
              <span class="filter-dot" v-if="f.key === 'opened'" style="background:#10B981"></span>
              <span class="filter-dot" v-if="f.key === 'merged'" style="background:#2563EB"></span>
              <span class="filter-dot" v-if="f.key === 'closed'" style="background:#EF4444"></span>
              {{ f.label }}
            </button>
            <div class="filter-spacer"></div>
            <div class="filter-search">
              <el-icon :size="14"><Search /></el-icon>
              <input v-model="searchText" placeholder="搜索 CR..." class="search-input" />
            </div>
          </div>

        <div class="table-card" v-if="filteredCRs.length > 0">
          <div class="table-header">
            <span class="th" style="width:60px">CR</span>
            <span class="th" style="flex:1">标题</span>
            <span class="th" style="width:80px">平台</span>
            <span class="th" style="width:140px">源分支</span>
            <span class="th" style="width:140px">目标分支</span>
            <span class="th" style="width:80px">状态</span>
            <span class="th" style="width:100px">作者</span>
          </div>
          <div v-for="row in filteredCRs" :key="row.id" class="table-row">
            <span class="td td-id" style="width:60px">#{{ row.cr_number }}</span>
            <span class="td td-title" style="flex:1" :title="row.title">{{ row.title }}</span>
            <span class="td" style="width:80px">
              <span class="platform-tag">
                <span class="platform-dot" :style="{ background: platformColor(row.platform) }"></span>
                {{ platformLabel(row.platform) }}
              </span>
            </span>
            <span class="td td-mono" style="width:140px">{{ row.source_branch }}</span>
            <span class="td td-mono" style="width:140px">{{ row.target_branch }}</span>
            <span class="td" style="width:80px">
              <span class="status-pill" :class="'status-' + row.state">{{ stateLabel(row.state) }}</span>
            </span>
            <span class="td" style="width:100px">{{ row.author_name }}</span>
          </div>
        </div>

        <div v-else-if="!loading" class="empty-state">
          <el-icon class="empty-icon"><Share /></el-icon>
          <h3>暂无 CR</h3>
          <p>点击"同步"从远程平台拉取 Change Request</p>
        </div>
      </div>
    </div>
  </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, Search, Share } from '@element-plus/icons-vue'
import { listCRs, syncCRs } from '@/api/modules/cr'
import type { CRDTO } from '@/api/modules/cr'
import RepoSidebar from '@/components/repo/RepoSidebar.vue'
import { getRepoDetail } from '@/api/modules/repo'
import { getCurrentVersion } from '@/api/modules/version'

const route = useRoute()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const syncing = ref(false)
const activeFilter = ref('all')
const searchText = ref('')
const crs = ref<CRDTO[]>([])
const repoName = ref('')
const currentVersion = ref('')

const filters = [
  { key: 'all', label: '全部' },
  { key: 'opened', label: 'Open' },
  { key: 'merged', label: 'Merged' },
  { key: 'closed', label: 'Closed' },
]

const PLATFORM_COLORS: Record<string, string> = { gitlab: '#FC6D26', github: '#24292F', gitea: '#609926' }
const PLATFORM_LABELS: Record<string, string> = { gitlab: 'GitLab', github: 'GitHub', gitea: 'Gitea' }

function platformColor(p?: string) { return PLATFORM_COLORS[p || ''] || '#6B7280' }
function platformLabel(p?: string) { return PLATFORM_LABELS[p || ''] || p || '-' }
function stateLabel(s: string) {
  if (s === 'opened') return 'Open'
  if (s === 'merged') return 'Merged'
  return s.charAt(0).toUpperCase() + s.slice(1)
}

const filteredCRs = computed(() => {
  let list = crs.value
  if (activeFilter.value !== 'all') {
    list = list.filter(cr => cr.state === activeFilter.value)
  }
  if (searchText.value) {
    const q = searchText.value.toLowerCase()
    list = list.filter(cr => cr.title.toLowerCase().includes(q) || String(cr.cr_number).includes(q))
  }
  return list
})

async function loadCRs() {
  loading.value = true
  try {
    const res = await listCRs({ repo_key: repoKey, page: 1, page_size: 100 })
    crs.value = res?.items || []
  } catch { crs.value = [] }
  finally { loading.value = false }
}

async function handleSync() {
  syncing.value = true
  try {
    const res = await syncCRs(repoKey)
    ElMessage.success(`同步完成，共 ${res?.synced_count || 0} 个 CR`)
    loadCRs()
  } catch (e: any) {
    ElMessage.error('同步失败: ' + (e?.message || ''))
  } finally {
    syncing.value = false
  }
}

onMounted(async () => {
  loadCRs()
  try {
    const r = await getRepoDetail(repoKey)
    repoName.value = r?.name || ''
  } catch {}
  try { currentVersion.value = (await getCurrentVersion(repoKey)) || '' } catch {}
})
</script>

<style scoped>
.cr-page-wrapper {
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

.header-actions {
  display: flex;
  gap: 10px;
}

.action-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.action-pill--primary {
  background: var(--accent-primary, #6366F1);
  color: #fff;
}

.action-pill--primary:hover:not(:disabled) {
  background: #4F46E5;
}

.action-pill:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cr-layout {
  display: flex;
  gap: 20px;
}

.cr-content {
  flex: 1;
  min-height: calc(100vh - 180px);
}

.cr-management-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.filter-bar { display: flex; align-items: center; gap: 12px; }

.filter-btn {
  display: flex; align-items: center; gap: 4px; padding: 8px 16px;
  border-radius: 8px; border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff); font-size: 13px;
  color: var(--text-color-secondary); cursor: pointer; transition: all 0.2s;
}
.filter-btn.active { background: var(--accent-primary, #6366F1); border-color: var(--accent-primary, #6366F1); color: #fff; }
.filter-btn.active .filter-dot { background: #fff !important; }
.filter-dot { width: 8px; height: 8px; border-radius: 50%; }
.filter-spacer { flex: 1; }

.filter-search {
  display: flex; align-items: center; gap: 8px; padding: 8px 12px;
  border-radius: 8px; border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff); color: var(--text-color-secondary, #94A3B8); font-size: 13px;
}
.search-input { border: none; outline: none; background: transparent; font-size: 13px; color: var(--text-color-primary); width: 120px; }

.table-card {
  border-radius: 12px; border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff); overflow: hidden;
}

.table-header { display: flex; align-items: center; padding: 12px 20px; background: #EEF2FF; }
.th { font-size: 12px; font-weight: 600; color: var(--text-color-secondary); }

.table-row {
  display: flex; align-items: center; padding: 12px 20px;
  border-bottom: 1px solid var(--border-color, #e5e7eb); transition: background 0.15s;
}
.table-row:last-child { border-bottom: none; }
.table-row:hover { background: #F8FAFC; }

.td { font-size: 13px; color: var(--text-color-secondary); }
.td-id { color: var(--accent-primary, #6366F1); font-weight: 600; }
.td-title { color: var(--text-color-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.td-mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 12px; }

.platform-tag { display: flex; align-items: center; gap: 4px; font-size: 11px; }
.platform-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

.status-pill {
  display: inline-block; padding: 4px 8px; border-radius: 9999px;
  font-size: 11px; font-weight: 500; text-align: center; width: 64px;
}
.status-opened { background: #ECFDF5; color: #059669; }
.status-merged { background: #EFF6FF; color: #2563EB; }
.status-closed { background: #FEF2F2; color: #DC2626; }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px 0; gap: 12px; }
.empty-icon { font-size: 48px; color: var(--text-color-placeholder); }
.empty-state h3 { margin: 0; font-size: 16px; color: var(--text-color-primary); }
.empty-state p { margin: 0; font-size: 13px; color: var(--text-color-secondary); }
</style>
