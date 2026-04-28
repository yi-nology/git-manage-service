<template>
  <div class="patch-page">
    <div class="title-row">
      <button class="back-btn" @click="$router.push(`/local-repos/${repoKey}`)">
        <el-icon><ArrowLeft /></el-icon> 返回
      </button>
      <h2 class="page-title">Patch 管理</h2>
      <span v-if="repoName" class="repo-tag">{{ repoName }}</span>
    </div>

    <PatchManager :repo-key="repoKey" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
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

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-color-page);
  color: var(--text-color-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0;
}

.repo-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 6px;
  background: #EEF2FF;
  color: #6366F1;
  font-size: 12px;
  font-weight: 500;
}
</style>
