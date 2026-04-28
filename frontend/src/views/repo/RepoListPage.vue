<template>
  <div class="repo-list-page">
    <div class="title-row">
      <div class="title-left">
        <h2 class="page-title">仓库列表</h2>
        <p class="page-subtitle">管理和监控所有 Git 仓库</p>
      </div>
      <el-dropdown @command="handleAddCommand">
        <button class="add-btn">
          <el-icon><Plus /></el-icon> 添加仓库
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="register">
              <el-icon><FolderOpened /></el-icon> 注册本地仓库
            </el-dropdown-item>
            <el-dropdown-item command="clone">
              <el-icon><Download /></el-icon> 克隆远程仓库
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div class="search-card" v-if="repoStore.repoList.length > 0 || searchText">
      <el-icon class="search-icon"><SearchIcon /></el-icon>
      <input
        v-model="searchText"
        placeholder="搜索仓库名称或路径..."
        class="search-input"
      />
      <el-icon v-if="searchText" class="clear-icon" @click="searchText = ''"><Close /></el-icon>
    </div>

    <TableSkeleton
      v-if="repoStore.loading"
      :rows="5"
      :columns="6"
      :column-widths="['60px', '150px', '250px', '200px', '120px', '120px']"
    />

    <div class="repo-table-card" v-else-if="filteredRepos.length > 0">
      <div class="table-header">
        <span class="th" style="width:60px">ID</span>
        <span class="th" style="width:150px">名称</span>
        <span class="th" style="width:280px">路径</span>
        <span class="th" style="width:250px">远程地址</span>
        <span class="th" style="width:120px">远端平台</span>
        <span class="th" style="flex:1">操作</span>
      </div>
      <div
        v-for="row in paginatedData"
        :key="row.key"
        class="table-row"
      >
        <span class="td" style="width:60px">{{ row.id }}</span>
        <span class="td td-name" style="width:150px" @click="router.push(`/local-repos/${row.key}`)">{{ row.name }}</span>
        <span class="td td-mono" style="width:280px" :title="row.path">{{ row.path }}</span>
        <span class="td td-mono" style="width:250px">{{ row.remote_url || '无远程仓库' }}</span>
        <span class="td" style="width:120px">
          <template v-if="repoPlatformBadges[row.key]?.length">
            <el-tag v-for="b in repoPlatformBadges[row.key]" :key="b" size="small" effect="plain" style="margin-right:2px">{{ b }}</el-tag>
          </template>
          <template v-else><span style="color:var(--el-text-color-placeholder)">-</span></template>
        </span>
        <span class="td td-actions" style="flex:1">
          <button class="act-btn" @click="router.push(`/local-repos/${row.key}`)">
            <el-icon><View /></el-icon> 详情
          </button>
          <button class="act-btn" @click="router.push(`/local-repos/${row.key}/branches`)">
            <el-icon><Share /></el-icon> 分支
          </button>
          <button class="act-btn" @click="router.push(`/local-repos/${row.key}/sync`)">
            <el-icon><Refresh /></el-icon> 同步
          </button>
          <button class="act-btn act-btn--danger" @click="handleDelete(row.key, row.name)">
            <el-icon><Delete /></el-icon> 删除
          </button>
        </span>
      </div>
    </div>

    <div v-else class="empty-state">
      <el-icon class="empty-icon"><Folder /></el-icon>
      <h3>暂无仓库</h3>
      <p>添加您的第一个仓库开始管理</p>
      <button class="add-btn" @click="router.push('/local-repos/register')">添加第一个仓库</button>
    </div>

    <div class="pagination-row" v-if="filteredRepos.length > 0">
      <span class="pag-info">共 {{ filteredRepos.length }} 条</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  Delete,
  View,
  Share,
  Refresh,
  FolderOpened,
  Download,
  Folder,
  Plus,
  Search as SearchIcon,
  Close,
} from '@element-plus/icons-vue'
import { useRepoStore } from '@/stores/useRepoStore'
import { deleteRepo } from '@/api/modules/repo'
import { listBindings } from '@/api/modules/binding'
import type { RepoProviderBindingDTO } from '@/types/binding'
import { useNotification } from '@/composables/useNotification'
import TableSkeleton from '@/components/common/TableSkeleton.vue'

const router = useRouter()
const repoStore = useRepoStore()
const { showSuccess, showError } = useNotification()

const repoPlatformBadges = ref<Record<string, string[]>>({})

async function loadPlatformBadges() {
  try {
    const allBindings: RepoProviderBindingDTO[] = await listBindings() || []
    const map: Record<string, string[]> = {}
    for (const b of allBindings) {
      const key = b.repo_key
      if (!map[key]) map[key] = []
      const label = { gitlab: 'GitLab', github: 'GitHub', gitea: 'Gitea' }[b.platform] || b.platform
      if (!map[key].includes(label)) map[key].push(label)
    }
    repoPlatformBadges.value = map
  } catch {}
}

// 搜索
const searchText = ref('')

// 分页
const currentPage = ref(1)
const pageSize = ref(10)

// 排序
const sortProp = ref('id')
const sortOrder = ref<'ascending' | 'descending' | null>('ascending')

// @ts-ignore
const handleSortChange = ({ prop, order }: { prop: string; order: string | null }) => {
  sortProp.value = prop
  sortOrder.value = order as 'ascending' | 'descending' | null
}

onMounted(async () => {
  await repoStore.fetchRepoList()
  loadPlatformBadges()
})

// 过滤后的仓库列表
const filteredRepos = computed(() => {
  let list = [...repoStore.repoList]

  // 搜索过滤
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    list = list.filter(
      (repo) =>
        repo.name.toLowerCase().includes(search) ||
        repo.path.toLowerCase().includes(search) ||
        (repo.remote_url && repo.remote_url.toLowerCase().includes(search))
    )
  }

  // 排序
  if (sortProp.value && sortOrder.value) {
    list.sort((a: any, b: any) => {
      const aVal = a[sortProp.value]
      const bVal = b[sortProp.value]
      const modifier = sortOrder.value === 'ascending' ? 1 : -1

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return (aVal - bVal) * modifier
      }

      return String(aVal).localeCompare(String(bVal)) * modifier
    })
  }

  return list
})

// 分页数据
const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredRepos.value.slice(start, end)
})

// 添加命令
function handleAddCommand(command: string) {
  if (command === 'register') {
    router.push('/local-repos/register')
  } else if (command === 'clone') {
    router.push('/local-repos/clone')
  }
}

// 添加命令
async function handleDelete(key: string, name: string) {
  try {
    await ElMessageBox.confirm(
      `确定要删除仓库 "${name}" 吗？如果被同步任务使用将无法删除。`,
      '确认删除',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消',
      }
    )

    await deleteRepo(key)
    showSuccess('仓库已删除')
    await repoStore.fetchRepoList()

    // 如果当前页没有数据了，返回上一页
    if (paginatedData.value.length === 0 && currentPage.value > 1) {
      currentPage.value--
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      showError('删除失败', error)
    }
  }
}
</script>

<style scoped>
.repo-list-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.add-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--primary-color);
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 500;
  padding: 10px 20px;
  border-radius: var(--border-radius-md);
  border: none;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.add-btn:hover {
  background: var(--primary-color-hover);
}

.search-card {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 12px 16px;
}

.search-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--text-color-primary);
  font-family: var(--font-family);
}

.search-input::placeholder {
  color: var(--text-color-placeholder);
}

.clear-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
  cursor: pointer;
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.clear-icon:hover {
  color: var(--text-color-primary);
}

.repo-table-card {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: var(--accent-bg);
}

.th {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-secondary);
}

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color);
  transition: background var(--transition-fast);
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--border-color-extra-light);
}

.td {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.td-name {
  color: var(--primary-color);
  font-weight: 500;
  cursor: pointer;
  transition: opacity var(--transition-fast);
}

.td-name:hover {
  opacity: 0.8;
}

.td-mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.td-actions {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
}

.act-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: transparent;
  font-size: 12px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.act-btn:hover {
  border-color: var(--accent-primary, #6366F1);
  color: var(--accent-primary, #6366F1);
}

.act-btn--danger:hover {
  border-color: #EF4444;
  color: #EF4444;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 12px;
}

.empty-icon {
  font-size: 48px;
  color: var(--text-color-placeholder);
}

.empty-state h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-color-primary);
}

.empty-state p {
  margin: 0;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.pagination-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 8px 0;
}

.pag-info {
  font-size: 12px;
  color: var(--text-color-secondary);
}

@media (max-width: 768px) {
  .repo-list-page {
    padding: var(--spacing-md);
  }

  .title-row {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .repo-table-card {
    overflow-x: auto;
  }

  .table-header,
  .table-row {
    min-width: 700px;
  }
}
</style>
