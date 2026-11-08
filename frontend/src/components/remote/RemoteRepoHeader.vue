<template>
  <PageHeader showBack backRoute="/remote-repos">
    <template #title-suffix>
      <div class="repo-icon" :style="{ background: platformMeta(platform).iconBg }">
        <el-icon :size="18" :style="{ color: platformMeta(platform).iconColor }"><FolderOpened /></el-icon>
      </div>
      <div class="repo-title-info">
        <h2>{{ repoFullName }}</h2>
        <span class="platform-badge" :style="{ background: platformMeta(platform).iconBg, color: platformMeta(platform).iconColor }">{{ platformMeta(platform).label }}</span>
        <StatusBadge v-if="linkedRepoKey" variant="success" text="已关联本地" />
      </div>
    </template>
    <template #actions>
      <ActionPill v-if="!linkedRepoKey" variant="primary" :icon="Download" @click="$emit('clone')">克隆到本地</ActionPill>
      <ActionPill v-else variant="outline" :icon="FolderOpened" @click="$router.push(ROUTES.RepoDetail(linkedRepoKey))">查看本地仓库</ActionPill>
    </template>
  </PageHeader>

  <div v-if="!linkedRepoKey" class="unlinked-card">
    <div class="unlinked-icon"><el-icon :size="32"><Link /></el-icon></div>
    <h3>该仓库尚未关联本地</h3>
    <div class="unlinked-info">
      <div class="info-row">
        <span class="info-label">HTTP</span>
        <span class="info-value mono">{{ repoData?.clone_url || '-' }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">SSH</span>
        <span class="info-value mono">{{ repoData?.ssh_url || '-' }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">默认分支</span>
        <span class="info-value">{{ repoData?.default_branch || '-' }}</span>
      </div>
      <div class="info-row">
        <span class="info-label">可见性</span>
        <span class="info-value">{{ repoData?.private ? '私有' : '公开' }}</span>
      </div>
    </div>
    <ActionPill variant="primary" :icon="Download" @click="$emit('clone')" style="margin-top:16px">克隆到本地</ActionPill>
  </div>
</template>

<script setup lang="ts">
// 远程仓库详情页头部：平台标识 + 关联状态 + 未关联时的仓库信息卡。
import { ROUTES } from '@/router/paths'
import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { FolderOpened, Download, Link } from '@element-plus/icons-vue'

defineProps<{
  platform: string
  repoFullName: string
  linkedRepoKey: string | null
  repoData: { clone_url?: string; ssh_url?: string; default_branch?: string; private?: boolean } | null
}>()

defineEmits<{
  clone: []
}>()

const PLATFORM_META: Record<string, { label: string; iconBg: string; iconColor: string }> = {
  gitlab: { label: 'GitLab', iconBg: '#FFF4E6', iconColor: '#FC6D26' },
  github: { label: 'GitHub', iconBg: '#F3F4F6', iconColor: '#24292F' },
  gitea: { label: 'Gitea', iconBg: '#ECFDF5', iconColor: '#609926' },
  gitee: { label: 'Gitee', iconBg: '#FEF2F2', iconColor: '#C71D23' },
  tencent_code: { label: '腾讯工蜂', iconBg: '#E8F5E9', iconColor: '#1B5E20' },
  forgejo: { label: 'Forgejo', iconBg: '#FFF7ED', iconColor: '#F97316' },
}

function platformMeta(p: string) {
  return PLATFORM_META[p] || { label: p, iconBg: '#F3F4F6', iconColor: '#6B7280' }
}
</script>

<style scoped>
.repo-title-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.repo-title-info h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.platform-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.repo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
}

.unlinked-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--bg-color-page);
}

.unlinked-icon { color: var(--text-color-placeholder); }
.unlinked-card h3 { margin: 0; font-size: 16px; color: var(--text-color-primary); }

.unlinked-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  max-width: 600px;
  margin-top: 16px;
}

.info-row {
  display: flex;
  gap: 12px;
  align-items: baseline;
}

.info-label {
  width: 70px;
  font-size: 12px;
  color: var(--text-color-placeholder);
  flex-shrink: 0;
}

.info-value {
  font-size: 13px;
  color: var(--text-color-primary);
}

.mono {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
}
</style>
