<template>
  <el-dialog v-model="visible" title="添加远端关联" width="600px" @close="reset">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="自动检测" name="auto">
        <div v-if="loading" style="text-align: center; padding: 20px">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>检测中...</span>
        </div>
        <div v-else-if="suggestions.length === 0" style="text-align: center; padding: 20px">
          <el-text type="info">未检测到可关联的远端平台。请确认已配置对应的 Provider 且本地仓库有匹配的 remote。</el-text>
        </div>
        <div v-else>
          <div
            v-for="s in suggestions"
            :key="`${s.provider_config_id}-${s.platform_owner}-${s.platform_repo}`"
            class="suggestion-card"
            :class="{ selected: selected === s }"
            @click="selected = s"
          >
            <div class="suggestion-header">
              <el-tag :type="platformTagType(s.platform)" size="small" effect="dark">
                {{ platformLabel(s.platform) }}
              </el-tag>
              <span>{{ s.platform_owner }}/{{ s.platform_repo }}</span>
              <el-tag size="small" type="info">{{ s.confidence }}</el-tag>
            </div>
            <div class="suggestion-body">
              <el-text size="small" type="info">
                remote: {{ s.remote_name }} | {{ s.remote_url }}
              </el-text>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="手动选择" name="manual">
        <el-form :model="manualForm" label-width="100px" size="default">
          <el-form-item label="Provider">
            <el-select v-model="manualForm.provider_config_id" placeholder="选择平台" style="width: 100%">
              <el-option
                v-for="p in providers"
                :key="p.id"
                :label="`${p.name} (${platformLabel(p.platform)})`"
                :value="p.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Owner">
            <el-input v-model="manualForm.platform_owner" placeholder="如 yi-nology" />
          </el-form-item>
          <el-form-item label="仓库名">
            <el-input v-model="manualForm.platform_repo" placeholder="如 my-project" />
          </el-form-item>
          <el-form-item label="Remote">
            <el-input v-model="manualForm.remote_name" placeholder="如 origin（可选）" />
          </el-form-item>
          <el-form-item label="设为主关联">
            <el-switch v-model="manualForm.is_primary" />
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :disabled="!canSubmit" :loading="submitting" @click="handleSubmit">
        确认关联
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { ProviderConfigDTO } from '@/api/modules/provider'
import { autoDetectBindings, createBinding } from '@/api/modules/binding'
import type { BindingSuggestion } from '@/types/binding'

const props = defineProps<{
  repoKey: string
  providers: ProviderConfigDTO[]
}>()

const emit = defineEmits<{
  created: []
}>()

const visible = defineModel<boolean>('visible', { default: false })

const activeTab = ref('auto')
const loading = ref(false)
const submitting = ref(false)
const suggestions = ref<BindingSuggestion[]>([])
const selected = ref<BindingSuggestion | null>(null)

const manualForm = ref({
  provider_config_id: undefined as number | undefined,
  platform_owner: '',
  platform_repo: '',
  remote_name: '',
  is_primary: true,
})

const canSubmit = computed(() => {
  if (activeTab.value === 'auto') return selected.value !== null
  return manualForm.value.provider_config_id !== undefined &&
    manualForm.value.platform_owner !== '' &&
    manualForm.value.platform_repo !== ''
})

watch(visible, (v) => {
  if (v && props.repoKey) {
    doAutoDetect()
  }
})

async function doAutoDetect() {
  loading.value = true
  suggestions.value = []
  selected.value = null
  try {
    const resp = await autoDetectBindings(props.repoKey)
    suggestions.value = resp?.suggestions || []
  } catch {
    suggestions.value = []
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (activeTab.value === 'auto' && selected.value) {
      await createBinding({
        repo_key: props.repoKey,
        provider_config_id: selected.value.provider_config_id,
        platform_owner: selected.value.platform_owner,
        platform_repo: selected.value.platform_repo,
        remote_name: selected.value.remote_name,
        is_primary: true,
      })
    } else {
      await createBinding({
        repo_key: props.repoKey,
        provider_config_id: manualForm.value.provider_config_id!,
        platform_owner: manualForm.value.platform_owner,
        platform_repo: manualForm.value.platform_repo,
        remote_name: manualForm.value.remote_name || undefined,
        is_primary: manualForm.value.is_primary,
      })
    }
    ElMessage.success('关联创建成功')
    visible.value = false
    emit('created')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建关联失败')
  } finally {
    submitting.value = false
  }
}

function reset() {
  suggestions.value = []
  selected.value = null
  manualForm.value = {
    provider_config_id: undefined,
    platform_owner: '',
    platform_repo: '',
    remote_name: '',
    is_primary: true,
  }
}

function platformLabel(platform: string) {
  const map: Record<string, string> = { gitlab: 'GitLab', github: 'GitHub', gitea: 'Gitea', tencent_code: '腾讯工蜂' }
  return map[platform] || platform
}

function platformTagType(platform: string) {
  const map: Record<string, string> = { gitlab: 'danger', github: '', gitea: 'success', tencent_code: 'success' }
  return (map[platform] || 'info') as '' | 'success' | 'warning' | 'danger' | 'info'
}
</script>

<style scoped>
.suggestion-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
}
.suggestion-card:hover {
  border-color: var(--el-color-primary-light-3);
}
.suggestion-card.selected {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}
.suggestion-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}
.suggestion-body {
  margin-top: 4px;
}
</style>
