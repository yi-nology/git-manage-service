<template>
  <div class="fe-sidebar">
    <div v-if="isWorktree" class="fe-view-toggle">
      <span :class="{ active: viewMode === 'all' }" @click="$emit('update:viewMode', 'all')">全部</span>
      <span :class="{ active: viewMode === 'changes' }" @click="$emit('update:viewMode', 'changes')">仅变更</span>
    </div>

    <div v-show="viewMode === 'all'" class="fe-tree" v-loading="treeLoading">
      <div
        v-for="item in flatTreeItems"
        :key="item.path"
        class="fe-tree-item"
        :class="{
          active: selectedFile === item.path && !item.isDir,
          'is-dir': item.isDir,
        }"
        :style="{ paddingLeft: 8 + item.depth * 16 + 'px' }"
        @click="item.isDir ? $emit('toggleDir', item.path) : $emit('selectTreeFile', item.path)"
      >
        <span v-if="item.isDir" class="fe-tree-toggle">
          <el-icon :size="12">
            <ArrowDown v-if="expandedDirs.has(item.path)" />
            <ArrowRight v-else />
          </el-icon>
        </span>
        <el-icon v-if="item.isDir" :size="14"><Folder /></el-icon>
        <el-icon v-else :size="14"><Document /></el-icon>
        <span class="fe-tree-name" :title="item.name">{{ item.name }}</span>

        <template v-if="isWorktree && getFileStatus(item.path)">
          <el-tag
            size="small"
            :type="statusTagType(getFileStatus(item.path)!)"
            class="fe-tree-badge"
          >
            {{ getFileStatus(item.path)![0].toUpperCase() }}
          </el-tag>
          <div class="fe-tree-actions">
            <template v-if="isFileUntracked(item.path)">
              <el-button size="small" type="success" link @click.stop="$emit('stageFile', item.path)">
                <el-icon><Plus /></el-icon>
              </el-button>
              <el-button size="small" type="info" link @click.stop="$emit('gitignore', [item.path])">
                <el-icon><Close /></el-icon>
              </el-button>
            </template>
            <template v-else-if="isFileConflicted(item.path)">
              <el-button size="small" type="danger" link @click.stop="$emit('openConflict', item.path)">
                <el-icon><WarningFilled /></el-icon>
              </el-button>
            </template>
            <template v-else>
              <el-button
                v-if="isFileUnstaged(item.path)"
                size="small" type="success" link
                @click.stop="$emit('stageFile', item.path)"
              >
                <el-icon><Plus /></el-icon>
              </el-button>
              <el-button
                v-if="isFileStaged(item.path)"
                size="small" type="warning" link
                @click.stop="$emit('unstageFile', item.path)"
              >
                <el-icon><Minus /></el-icon>
              </el-button>
            </template>
          </div>
        </template>
      </div>
      <el-empty v-if="!treeLoading && !flatTreeItems.length" description="空目录" :image-size="48" />
    </div>

    <div v-if="viewMode === 'changes' && isWorktree" class="fe-changes">
      <template v-if="wsStatus?.staged?.length">
        <div class="fe-section-header">
          <span class="fe-dot fe-dot-green" />
          已暂存 ({{ wsStatus.staged.length }})
          <div class="fe-section-spacer" />
          <el-button size="small" type="warning" link @click.stop="$emit('unstageAll')">全部取消</el-button>
        </div>
        <div
          v-for="f in wsStatus.staged"
          :key="'s-' + f.path"
          class="fe-change-item"
          :class="{ active: selectedFile === f.path }"
          @click="$emit('selectChangedFile', f.path)"
        >
          <el-icon :size="14"><Document /></el-icon>
          <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
          <el-tag size="small" :type="statusTagType(f.status)">{{ f.status[0].toUpperCase() }}</el-tag>
          <el-dropdown trigger="click" @command="(cmd: string) => $emit('fileAction', cmd, f.path)" @click.stop teleported>
            <el-button size="small" type="primary" link @click.stop>
              <el-icon><Operation /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="unstage"><el-icon><Minus /></el-icon> 取消暂存</el-dropdown-item>
                <el-dropdown-item command="untrack"><el-icon><Close /></el-icon> 移除跟踪</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>

      <template v-if="wsStatus?.unstaged?.length">
        <div class="fe-section-header">
          <span class="fe-dot fe-dot-orange" />
          未暂存 ({{ wsStatus.unstaged.length }})
          <div class="fe-section-spacer" />
          <el-button size="small" type="success" link @click.stop="$emit('stageAllUnstaged')">全部暂存</el-button>
        </div>
        <div
          v-for="f in wsStatus.unstaged"
          :key="'u-' + f.path"
          class="fe-change-item"
          :class="{ active: selectedFile === f.path }"
          @click="$emit('selectChangedFile', f.path)"
        >
          <el-icon :size="14"><Document /></el-icon>
          <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
          <el-tag size="small" :type="statusTagType(f.status)">{{ f.status[0].toUpperCase() }}</el-tag>
          <el-dropdown trigger="click" @command="(cmd: string) => $emit('fileAction', cmd, f.path)" @click.stop teleported>
            <el-button size="small" type="primary" link @click.stop>
              <el-icon><Operation /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="stage"><el-icon><Plus /></el-icon> 暂存</el-dropdown-item>
                <el-dropdown-item command="untrack"><el-icon><Close /></el-icon> 移除跟踪</el-dropdown-item>
                <el-dropdown-item command="gitignore"><el-icon><Close /></el-icon> 添加到 .gitignore</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>

      <template v-if="wsStatus?.untracked?.length">
        <div class="fe-section-header">
          <span class="fe-dot fe-dot-gray" />
          未跟踪 ({{ wsStatus.untracked.length }})
          <div class="fe-section-spacer" />
          <el-button size="small" type="info" link @click.stop="$emit('gitignore', wsStatus!.untracked.map(f => f.path))">全部忽略</el-button>
        </div>
        <div
          v-for="f in wsStatus.untracked"
          :key="'t-' + f.path"
          class="fe-change-item"
          :class="{ active: selectedFile === f.path }"
          @click="$emit('selectChangedFile', f.path)"
        >
          <el-icon :size="14"><Document /></el-icon>
          <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
          <el-tag size="small" type="info">?</el-tag>
          <el-dropdown trigger="click" @command="(cmd: string) => $emit('fileAction', cmd, f.path)" @click.stop teleported>
            <el-button size="small" type="primary" link @click.stop>
              <el-icon><Operation /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="stage"><el-icon><Plus /></el-icon> 暂存（跟踪）</el-dropdown-item>
                <el-dropdown-item command="gitignore"><el-icon><Close /></el-icon> 添加到 .gitignore</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>

      <template v-if="wsStatus?.conflicted?.length">
        <div class="fe-section-header">
          <span class="fe-dot fe-dot-red" />
          冲突 ({{ wsStatus.conflicted.length }})
        </div>
        <div
          v-for="f in wsStatus.conflicted"
          :key="'c-' + f.path"
          class="fe-change-item fe-conflict-item"
          @click="$emit('openConflict', f.path)"
        >
          <el-icon :size="14" color="var(--danger-color)"><WarningFilled /></el-icon>
          <span class="fe-change-name" :title="f.path">{{ f.path }}</span>
          <el-tag size="small" type="danger">C</el-tag>
          <el-button size="small" type="primary" @click.stop="$emit('openConflict', f.path)">
            <el-icon><MagicStick /></el-icon>
          </el-button>
        </div>
      </template>

      <el-empty v-if="wsStatus?.isClean" description="工作区干净" :image-size="60" />
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Folder, Document, WarningFilled, MagicStick, Plus, Minus, Close,
  ArrowRight, ArrowDown, Operation,
} from '@element-plus/icons-vue'
import type { FlatTreeItem } from '@/composables/useFileTree'
import type { WorkspaceStatus } from '@/api/modules/workspace'

defineProps<{
  isWorktree: boolean
  viewMode: 'all' | 'changes'
  selectedFile: string
  flatTreeItems: FlatTreeItem[]
  treeLoading: boolean
  expandedDirs: Set<string>
  wsStatus: WorkspaceStatus | null
  getFileStatus: (path: string) => string | undefined
  isFileStaged: (path: string) => boolean
  isFileUnstaged: (path: string) => boolean
  isFileUntracked: (path: string) => boolean
  isFileConflicted: (path: string) => boolean
}>()

defineEmits<{
  'update:viewMode': [value: 'all' | 'changes']
  toggleDir: [path: string]
  selectTreeFile: [path: string]
  selectChangedFile: [path: string]
  stageFile: [path: string]
  unstageFile: [path: string]
  stageAllUnstaged: []
  unstageAll: []
  gitignore: [paths: string[]]
  fileAction: [action: string, path: string]
  openConflict: [path: string]
}>()

function statusTagType(status: string): 'primary' | 'danger' | 'success' | 'info' | 'warning' {
  switch (status) {
    case 'added': return 'success'
    case 'modified': return 'warning'
    case 'deleted': return 'danger'
    case 'renamed': return 'info'
    case 'copied': return 'info'
    case '?': return 'info'
    case 'C': return 'danger'
    default: return 'info'
  }
}
</script>

<style scoped>
.fe-sidebar {
  width: 280px;
  min-width: 280px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-color);
  overflow: hidden;
}

.fe-view-toggle {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.fe-view-toggle span {
  flex: 1;
  text-align: center;
  padding: 6px 0;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  color: var(--text-color-secondary);
  border-bottom: 2px solid transparent;
  transition: all var(--transition-fast);
}

.fe-view-toggle span:hover {
  color: var(--text-color-primary);
}

.fe-view-toggle span.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

.fe-tree {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.fe-tree-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  cursor: pointer;
  transition: background var(--transition-fast);
  position: relative;
  white-space: nowrap;
}

.fe-tree-item:hover {
  background: var(--surface-hover);
}

.fe-tree-item.active {
  background: var(--accent-bg);
}

.fe-tree-item.is-dir {
  color: var(--text-color-primary);
}

.fe-tree-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.fe-tree-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  color: var(--text-color-primary);
}

.fe-tree-badge {
  flex-shrink: 0;
  margin-left: 4px;
}

.fe-tree-actions {
  display: none;
  align-items: center;
  gap: 2px;
  margin-left: auto;
  flex-shrink: 0;
}

.fe-tree-item:hover .fe-tree-actions {
  display: flex;
}

.fe-changes {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.fe-section-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-color-regular);
  position: sticky;
  top: 0;
  background: var(--bg-color-page);
  z-index: 1;
}

.fe-section-spacer {
  flex: 1;
}

.fe-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.fe-dot-green { background: var(--success-color); }
.fe-dot-orange { background: var(--warning-color); }
.fe-dot-gray { background: var(--text-color-secondary); }
.fe-dot-red { background: var(--danger-color); }

.fe-change-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.fe-change-item:hover {
  background: var(--surface-hover);
}

.fe-change-item.active {
  background: var(--accent-bg);
}

.fe-conflict-item {
  background: rgba(239, 68, 68, 0.05);
  border-left: 2px solid var(--danger-color);
}

.fe-change-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-color-primary);
}
</style>
