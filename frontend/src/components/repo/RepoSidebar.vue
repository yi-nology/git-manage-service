<template>
  <div class="repo-sidebar">
    <div class="sidebar-card">
      <div
        v-for="item in sidebarItems"
        :key="item.key"
        class="sidebar-item"
        :class="{ active: activeKey === item.key }"
        @click="handleClick(item)"
      >
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { InfoFilled, Document, DataAnalysis, Files, Timer, Folder, Search, Box, Link, DocumentCopy, Checked } from '@element-plus/icons-vue'

const props = defineProps<{
  repoKey: string
  activeKey: string
}>()

const router = useRouter()

const sidebarItems = [
  { key: 'info', label: '基本信息', icon: InfoFilled },
  { key: 'spec', label: 'Spec 编辑器', icon: Document },
  { key: 'stats', label: 'Git 有效提交度量', icon: DataAnalysis },
  { key: 'lines', label: '真实工程代码度量', icon: Files },
  { key: 'versions', label: '版本历史', icon: Timer },
  { key: 'files', label: '文件浏览', icon: Folder },
  { key: 'commits', label: 'Commit 搜索', icon: Search },
  { key: 'stash', label: 'Stash 管理', icon: Box },
  { key: 'submodules', label: 'Submodule', icon: Link },
  { key: 'patches', label: 'Patch 管理', icon: DocumentCopy },
  { key: 'review', label: '代码审查', icon: Checked },
]

function handleClick(item: { key: string }) {
  const routeMap: Record<string, string> = {
    info: `/local-repos/${props.repoKey}`,
    spec: `/local-repos/${props.repoKey}`,
    stats: `/local-repos/${props.repoKey}`,
    lines: `/local-repos/${props.repoKey}`,
    versions: `/local-repos/${props.repoKey}`,
    files: `/local-repos/${props.repoKey}`,
    commits: `/local-repos/${props.repoKey}`,
    stash: `/local-repos/${props.repoKey}`,
    submodules: `/local-repos/${props.repoKey}`,
    patches: `/local-repos/${props.repoKey}`,
    review: `/local-repos/${props.repoKey}/review`,
  }
  const target = routeMap[item.key]
  if (target) router.push(target)
}
</script>

<style scoped>
.repo-sidebar {
  width: 220px;
  flex-shrink: 0;
}

.sidebar-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 8px;
  height: calc(100vh - 180px);
  overflow-y: auto;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--text-color-primary);
  font-size: 13px;
}

.sidebar-item:hover {
  background: #F3F4F6;
}

.sidebar-item.active {
  background: var(--accent-bg);
  color: var(--accent-primary);
}

.sidebar-item.active .el-icon {
  color: var(--accent-primary);
}

.sidebar-item .el-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
}
</style>
