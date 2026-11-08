<template>
  <div class="remote-repo-detail-page">
    <RemoteRepoHeader
      :platform="providerPlatform"
      :repo-full-name="repoFullName"
      :linked-repo-key="linkedRepoKey"
      :repo-data="repoData"
      @clone="handleClone"
    />

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'cr' }" @click="activeTab = 'cr'">CR / MR</button>
      <button class="tab-btn" :class="{ active: activeTab === 'codereview' }" @click="activeTab = 'codereview'">代码审查</button>
      <button class="tab-btn" :class="{ active: activeTab === 'branchrules' }" @click="activeTab = 'branchrules'">分支规则</button>
      <button class="tab-btn" :class="{ active: activeTab === 'webhooks' }" @click="activeTab = 'webhooks'">Webhook 事件</button>
      <button class="tab-btn" :class="{ active: activeTab === 'branches' }" @click="activeTab = 'branches'">远程分支</button>
    </div>

    <div v-show="activeTab === 'cr'" class="tab-content">
      <CRReviewDetail v-if="reviewTask" :task="reviewTask" :repo-owner="repoOwner" :repo-name="repo_name" @close="reviewTask = null" @retried="reviewTask = $event" />
      <CRTab v-else :active="activeTab === 'cr'" :provider-id="provider_id" :repo-owner="repoOwner" :repo-name="repo_name" @show-review="reviewTask = $event" />
    </div>

    <div v-show="activeTab === 'codereview'" class="tab-content">
      <RemoteReviewConfigTab :active="activeTab === 'codereview'" :provider-id="provider_id" :repo-owner="repoOwner" :repo-name="repo_name" />
    </div>

    <div v-show="activeTab === 'branchrules'" class="tab-content">
      <BranchRulesTab :active="activeTab === 'branchrules'" :provider-id="provider_id" :repo-owner="repoOwner" :repo-name="repo_name" />
    </div>

    <div v-show="activeTab === 'webhooks'" class="tab-content">
      <WebhookEventsTab :active="activeTab === 'webhooks'" />
    </div>

    <div v-show="activeTab === 'branches'" class="tab-content">
      <RemoteBranchesTab
        :active="activeTab === 'branches'"
        :provider-id="provider_id"
        :repo-owner="repoOwner"
        :repo-name="repo_name"
        :linked-repo-key="linkedRepoKey"
        :default-branch="repoData?.default_branch || ''"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ROUTES } from '@/router/paths'
import { useRoute, useRouter } from 'vue-router'
import { listProviderRepos } from '@/api/modules/provider'
import { getRepoList } from '@/api/modules/repo'
import { listBindings } from '@/api/modules/binding'
import type { ReviewTaskDTO } from '@/api/modules/review'
import { useProviderStore } from '@/stores/useProviderStore'
import RemoteRepoHeader from '@/components/remote/RemoteRepoHeader.vue'
import CRTab from '@/components/remote/CRTab.vue'
import CRReviewDetail from '@/components/remote/CRReviewDetail.vue'
import RemoteReviewConfigTab from '@/components/remote/RemoteReviewConfigTab.vue'
import BranchRulesTab from '@/components/remote/BranchRulesTab.vue'
import WebhookEventsTab from '@/components/remote/WebhookEventsTab.vue'
import RemoteBranchesTab from '@/components/remote/RemoteBranchesTab.vue'

const providerStore = useProviderStore()

const route = useRoute()
const router = useRouter()
const provider_id = Number(route.params.provider_id)
const repoOwner = route.params.repoOwner as string
const repo_name = route.params.repo_name as string
const repoFullName = computed(() => `${repoOwner}/${repo_name}`)

const providerPlatform = ref('')
const linkedRepoKey = ref<string | null>(null)
const activeTab = ref((route.query.tab as string) || 'cr')
const reviewTask = ref<ReviewTaskDTO | null>(null)
const repoData = ref<{ clone_url?: string; ssh_url?: string; default_branch?: string; private?: boolean } | null>(null)

async function loadInitial() {
  const [, , bindings] = await Promise.all([
    providerStore.fetchProviders(),
    getRepoList().catch(() => []),
    listBindings({ provider_config_id: provider_id }).catch(() => []),
  ])
  const prov = providerStore.getProviderById(provider_id)
  if (prov) providerPlatform.value = prov.platform

  const linked = (bindings || []).find((b: any) =>
    b.platform_owner === repoOwner &&
    b.platform_repo === repo_name &&
    b.status === 'active'
  )
  if (linked) linkedRepoKey.value = linked.repo_key

  const remote_repos = await listProviderRepos(provider_id, { page: 1, per_page: 100 }).catch(() => [])
  const found = (remote_repos || []).find((r: any) => r.full_name === repoFullName.value)
  if (found) repoData.value = found
}

function handleClone() {
  const url = repoData.value?.ssh_url || repoData.value?.clone_url
  if (url) {
    const query: Record<string, string> = { url }
    if (provider_id) query.provider_config_id = String(provider_id)
    if (repoOwner) query.platform_owner = repoOwner
    if (repo_name) query.platform_repo = repo_name
    router.push({ path: ROUTES.RepoClone, query })
  }
}

watch(activeTab, (tab) => {
  if (tab !== 'cr') reviewTask.value = null
})

onMounted(loadInitial)
</script>

<style scoped>
.remote-repo-detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.tab-bar {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0;
}

.tab-btn {
  padding: 10px 20px;
  border: none;
  background: transparent;
  font-size: 14px;
  color: var(--text-color-secondary);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
  font-weight: 500;
}

.tab-btn.active {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
}

.tab-btn:hover {
  color: var(--accent-primary);
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
