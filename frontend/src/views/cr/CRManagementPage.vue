<template>
  <div class="cr-page-wrapper">
    <PageHeader :title="repoName || '仓库'" showBack :backRoute="`/local-repos/${repoKey}`">
      <template #title-suffix>
        <span v-if="currentVersion" class="version-tag">{{ currentVersion }}</span>
      </template>
      <template #actions>
        <ActionPill variant="primary" :icon="Refresh" @click="handleSync" :disabled="syncing">
          {{ syncing ? '同步中...' : '同步 CR' }}
        </ActionPill>
      </template>
    </PageHeader>

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
            <SearchBar v-model="searchText" placeholder="搜索 CR..." />
          </div>

          <DataTable v-if="filteredCRs.length > 0 || loading" :columns="crColumns" :data="filteredCRs" :loading="loading" row-key="id">
            <template #cell-cr_number="{ row }">
              <span class="td-id">#{{ row.cr_number }}</span>
            </template>
            <template #cell-title="{ row }">
              <span class="td-title" :title="row.title">{{ row.title }}</span>
            </template>
            <template #cell-platform="{ row }">
              <span class="platform-tag">
                <span class="platform-dot" :style="{ background: platformColor(row.platform) }"></span>
                {{ platformLabel(row.platform) }}
              </span>
            </template>
            <template #cell-source_branch="{ row }">
              <span class="mono">{{ row.source_branch }}</span>
            </template>
            <template #cell-target_branch="{ row }">
              <span class="mono">{{ row.target_branch }}</span>
            </template>
            <template #cell-state="{ row }">
              <StatusBadge :variant="crStateVariant(row.state)" :text="stateLabel(row.state)" :showDot="false" />
            </template>
          </DataTable>

          <EmptyState v-else-if="!loading" title="暂无 CR" description="点击「同步」从远程平台拉取 Change Request" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { listCRs, syncCRs } from '@/api/modules/cr'
import type { CRDTO } from '@/api/modules/cr'
import RepoSidebar from '@/components/repo/RepoSidebar.vue'
import { getRepoDetail } from '@/api/modules/repo'
import { getCurrentVersion } from '@/api/modules/version'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SearchBar from '@/components/common/SearchBar.vue'

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
  { key: 'opened', label: '进行中' },
  { key: 'merged', label: '已合并' },
  { key: 'closed', label: '已关闭' },
]

const PLATFORM_COLORS: Record<string, string> = { gitlab: '#FC6D26', github: '#24292F', gitea: '#609926' }
const PLATFORM_LABELS: Record<string, string> = { gitlab: 'GitLab', github: 'GitHub', gitea: 'Gitea' }

function platformColor(p?: string) { return PLATFORM_COLORS[p || ''] || '#6B7280' }
function platformLabel(p?: string) { return PLATFORM_LABELS[p || ''] || p || '-' }
function stateLabel(s: string) {
  if (s === 'opened') return '进行中'
  if (s === 'merged') return '已合并'
  if (s === 'closed') return '已关闭'
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function crStateVariant(s: string): 'success' | 'info' | 'danger' | 'default' {
  if (s === 'opened') return 'success'
  if (s === 'merged') return 'info'
  if (s === 'closed') return 'danger'
  return 'default'
}

const crColumns: TableColumn[] = [
  { key: 'cr_number', label: 'CR', width: '60px' },
  { key: 'title', label: '标题' },
  { key: 'platform', label: '平台', width: '80px' },
  { key: 'source_branch', label: '源分支', width: '140px' },
  { key: 'target_branch', label: '目标分支', width: '140px' },
  { key: 'state', label: '状态', width: '80px' },
  { key: 'author_name', label: '作者', width: '100px' },
]

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

.version-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  background: var(--accent-bg);
  color: #6366F1;
  font-size: 12px;
  font-weight: 500;
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
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
  border: none; border-bottom: 2px solid transparent;
  border-radius: 0; background: transparent; font-size: 13px;
  color: var(--text-color-secondary); cursor: pointer; transition: all 0.2s;
}
.filter-btn.active { background: transparent; border-bottom: 2px solid var(--accent-primary); color: var(--accent-primary); font-weight: 500; }
.filter-btn:hover { color: var(--accent-primary); }
.filter-btn.active .filter-dot { }
.filter-dot { width: 8px; height: 8px; border-radius: 50%; }
.filter-spacer { flex: 1; }

.td-id { color: var(--accent-primary); font-weight: 600; }
.td-title { color: var(--text-color-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.platform-tag { display: flex; align-items: center; gap: 4px; font-size: 11px; }
.platform-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}
</style>
