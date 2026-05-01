<template>
  <div class="patch-page">
    <PageHeader
      title="Patch 管理"
      :show-back="true"
      :back-route="`/local-repos/${repoKey}`"
    >
      <template #title-suffix>
        <span v-if="repoName" class="repo-tag">{{ repoName }}</span>
      </template>
    </PageHeader>

    <PatchManager :repo-key="repoKey" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '@/components/common/PageHeader.vue'
import PatchManager from '@/components/patch/PatchManager.vue'
import { getRepoDetail } from '@/api/modules/repo'

const route = useRoute()
const repoKey = route.params.repoKey as string
const repoName = ref('')

onMounted(async () => {
  try {
    const repo = await getRepoDetail(repoKey)
    repoName.value = repo.name
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
