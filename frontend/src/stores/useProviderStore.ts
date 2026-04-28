import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ProviderConfigDTO } from '@/api/modules/provider'
import { listProviders } from '@/api/modules/provider'

export const useProviderStore = defineStore('provider', () => {
  const providers = ref<ProviderConfigDTO[]>([])
  const loading = ref(false)
  const loaded = ref(false)

  async function fetchProviders(force = false) {
    if (loaded.value && !force) return providers.value
    loading.value = true
    try {
      const data = await listProviders()
      providers.value = Array.isArray(data) ? data : []
      loaded.value = true
    } catch (error) {
      console.error('[ProviderStore] Failed to fetch providers:', error)
      providers.value = []
    } finally {
      loading.value = false
    }
    return providers.value
  }

  function getProviderById(id: number) {
    return providers.value.find((p) => p.id === id)
  }

  function invalidate() {
    loaded.value = false
  }

  return { providers, loading, loaded, fetchProviders, getProviderById, invalidate }
})
