<template>
  <div class="repo-list-page">
    <PageHeader title="仓库列表" subtitle="管理和监控所有 Git 仓库">
      <template #actions>
        <el-dropdown @command="handleAddCommand">
          <ActionPill :icon="Plus">添加仓库</ActionPill>
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
      </template>
    </PageHeader>

    <SearchBar v-if="repoStore.repoList.length > 0 || searchText" v-model="searchText" placeholder="搜索仓库名称或路径..." />

    <TableSkeleton
      v-if="repoStore.loading"
      :rows="5"
      :columns="6"
      :column-widths="['60px', '150px', '250px', '200px', '120px', '120px']"
    />

    <DataTable
      v-else-if="filteredRepos.length > 0"
      :columns="columns"
      :data="paginatedData"
      row-key="key"
    >
      <template #cell-name="{ row }">
        <span class="cell-name" @click="router.push(`/local-repos/${row.key}`)">{{ row.name }}</span>
      </template>
      <template #cell-path="{ row }">
        <span class="cell-mono" :title="row.path">{{ row.path }}</span>
      </template>
      <template #cell-remote_url="{ row }">
        <span class="cell-mono">{{ row.remote_url || '无远程仓库' }}</span>
      </template>
      <template #cell-platforms="{ row }">
        <template v-if="repoPlatformBadges[row.key]?.length">
          <el-tag v-for="b in repoPlatformBadges[row.key]" :key="b" size="small" effect="plain" style="margin-right:2px">{{ b }}</el-tag>
        </template>
        <span v-else style="color:var(--text-color-placeholder)">-</span>
      </template>
      <template #row-actions="{ row }">
        <ActionPill variant="outline" small @click="router.push(`/local-repos/${row.key}`)">
          <el-icon><View /></el-icon> 详情
        </ActionPill>
        <ActionPill variant="outline" small @click="router.push(`/local-repos/${row.key}/branches`)">
          <el-icon><Share /></el-icon> 分支
        </ActionPill>
        <ActionPill variant="outline" small @click="router.push(`/local-repos/${row.key}/sync`)">
          <el-icon><Refresh /></el-icon> 同步
        </ActionPill>
        <ActionPill variant="danger" small @click="handleDelete(row.key, row.name)">
          <el-icon><Delete /></el-icon> 删除
        </ActionPill>
      </template>
    </DataTable>

    <EmptyState v-else icon="Folder" title="暂无仓库" description="添加您的第一个仓库开始管理">
      <template #action>
        <ActionPill :icon="Plus" @click="router.push('/local-repos/register')">添加第一个仓库</ActionPill>
      </template>
    </EmptyState>

    <PaginationBar v-if="filteredRepos.length > 0" :total="filteredRepos.length" v-model:current-page="currentPage" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  Delete, View, Share, Refresh, FolderOpened, Download, Plus,
} from '@element-plus/icons-vue'
import { useRepoStore } from '@/stores/useRepoStore'
import { deleteRepo } from '@/api/modules/repo'
import { listBindings } from '@/api/modules/binding'
import type { RepoProviderBindingDTO } from '@/types/binding'
import { useNotification } from '@/composables/useNotification'
import TableSkeleton from '@/components/common/TableSkeleton.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchBar from '@/components/common/SearchBar.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'

const router = useRouter()
const repoStore = useRepoStore()
const { showSuccess, showError } = useNotification()

const columns: TableColumn[] = [
  { key: 'id', label: 'ID', width: '60px' },
  { key: 'name', label: '名称', width: '150px' },
  { key: 'path', label: '路径', width: '280px' },
  { key: 'remote_url', label: '远程地址', width: '250px' },
  { key: 'platforms', label: '远端平台', width: '120px' },
  { key: 'actions', label: '操作', flex: 1 },
]

const repoPlatformBadges = ref<Record<string, string[]>>({})

async function loadPlatformBadges() {
  try {
    const allBindings: RepoProviderBindingDTO[] = await listBindings() || []
    const map: Record<string, string[]> = {}
    for (const b of allBindings) {
      const key = b.repo_key
      if (!map[key]) map[key] = []
      const label = { gitlab: 'GitLab', github: 'GitHub', gitea: 'Gitea', gitee: 'Gitee' }[b.platform] || b.platform
      if (!map[key].includes(label)) map[key].push(label)
    }
    repoPlatformBadges.value = map
  } catch {}
}

const searchText = ref('')
const currentPage = ref(1)
const page_size = ref(10)

onMounted(async () => {
  await repoStore.fetchRepoList()
  loadPlatformBadges()
})

const filteredRepos = computed(() => {
  let list = [...repoStore.repoList]

  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    list = list.filter(
      (repo) =>
        repo.name.toLowerCase().includes(search) ||
        repo.path.toLowerCase().includes(search) ||
        (repo.remote_url && repo.remote_url.toLowerCase().includes(search))
    )
  }

  return list
})

const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * page_size.value
  const end = start + page_size.value
  return filteredRepos.value.slice(start, end)
})

function handleAddCommand(command: string) {
  if (command === 'register') {
    router.push('/local-repos/register')
  } else if (command === 'clone') {
    router.push('/local-repos/clone')
  }
}

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

.cell-name {
  color: var(--primary-color);
  font-weight: 500;
  cursor: pointer;
  transition: opacity var(--transition-fast);
}

.cell-name:hover {
  opacity: 0.8;
}

.cell-mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: var(--font-size-xs);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
