import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RepoDTO } from '@/types/repo'
import { getRepoList } from '@/api/modules/repo'

export const useRepoStore = defineStore('repo', () => {
  const repoList = ref<RepoDTO[]>([])
  const loading = ref(false)

  async function fetchRepoList() {
    loading.value = true
    try {
      const data = await getRepoList()
      // 确保 repoList 始终是数组
      repoList.value = Array.isArray(data) ? data : []
    } catch (error) {
      console.error('[RepoStore] Failed to fetch repo list:', error)
      repoList.value = []
    } finally {
      loading.value = false
    }
  }

  return { repoList, loading, fetchRepoList }
})
