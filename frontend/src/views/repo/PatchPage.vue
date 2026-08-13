<template>
  <div class="patch-page">
    <PageHeader
      title="Patch 管理"
      :show-back="true"
      :back-route="`/local-repos/${repo_key}`"
    >
      <template #title-suffix>
        <span v-if="repo_name" class="repo-tag">{{ repo_name }}</span>
      </template>
    </PageHeader>

    <PatchManager :repo-key="repo_key" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '@/components/common/PageHeader.vue'
import PatchManager from '@/components/patch/PatchManager.vue'
import { getRepoDetail } from '@/api/modules/repo'

const route = useRoute()
const repo_key = route.params.repo_key as string
const repo_name = ref('')

onMounted(async () => {
  try {
    const repo = await getRepoDetail(repo_key)
    repo_name.value = repo.name
  } catch (e) {
    console.error('Failed to load repo info:', e)
  }
})
</script>

<style scoped>
.patch-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.repo-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  background: var(--accent-bg);
  color: #6366F1;
  font-size: 12px;
  font-weight: 500;
}
</style>
