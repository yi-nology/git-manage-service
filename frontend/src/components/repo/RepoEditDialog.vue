<template>
  <el-dialog :model-value="visible" @update:model-value="emit('update:visible', $event)" title="编辑仓库" width="750px" destroy-on-close>
    <el-form :model="editForm" label-width="100px">
      <el-form-item label="名称" required>
        <el-input v-model="editForm.name" placeholder="仓库名称" />
      </el-form-item>
      <el-form-item label="本地路径" required>
        <el-input v-model="editForm.path" placeholder="本地仓库路径" />
      </el-form-item>
      <el-form-item label="远程 URL">
        <div class="url-input-group">
          <el-radio-group v-model="editUrlMode" size="small" class="url-mode-switch">
            <el-radio-button value="ssh">SSH</el-radio-button>
            <el-radio-button value="https">HTTPS</el-radio-button>
          </el-radio-group>
          <el-input
            v-model="editForm.remote_url"
            :placeholder="editUrlMode === 'ssh' ? 'git@github.com:user/repo.git' : 'https://github.com/user/repo.git'"
            @blur="validateMainUrl"
            :class="{ 'is-error-input': editUrlError }"
          />
        </div>
        <div v-if="editUrlError" class="field-error">{{ editUrlError }}</div>
      </el-form-item>
      <el-form-item label="默认凭证">
        <CredentialSelector
          v-model="editDefaultCredentialId"
          :url="editForm.remote_url"
          placeholder="选择默认凭证（可选）"
        />
      </el-form-item>

      <el-divider content-position="left">远程仓库配置</el-divider>

      <el-form-item label="">
        <div class="remotes-section">
          <div class="remotes-header">
            <span>配置多个远程仓库及其凭证</span>
            <el-button size="small" type="primary" @click="addEditRemote">+ 新增远程</el-button>
          </div>
          <div v-for="(remote, index) in editRemotes" :key="index" class="edit-remote-item">
            <el-card shadow="hover">
              <div class="edit-remote-row">
                <el-input v-model="remote.name" size="small" placeholder="名称 (如 origin)" style="width: 120px;" />
                <el-radio-group v-model="remoteUrlModes[index]" size="small" class="url-mode-switch-sm">
                  <el-radio-button value="ssh">SSH</el-radio-button>
                  <el-radio-button value="https">HTTPS</el-radio-button>
                </el-radio-group>
                <el-input v-model="remote.fetch_url" size="small" :placeholder="remoteUrlModes[index] === 'ssh' ? 'git@host:user/repo.git' : 'https://host/repo.git'" style="flex: 1;" @blur="validateRemoteUrl(index)" :class="{ 'is-error-input': remoteUrlErrors[index] }" />
                <el-button size="small" :icon="Connection" circle @click="testEditRemote(index)" title="测试连接" :loading="remote._testing" />
                <el-button size="small" :icon="Delete" circle type="danger" @click="removeEditRemote(index)" title="删除" />
              </div>
              <div v-if="remoteUrlErrors[index]" class="field-error" style="margin-left: 128px;">{{ remoteUrlErrors[index] }}</div>
              <div class="edit-remote-cred">
                <span class="cred-label">凭证:</span>
                <CredentialSelector
                  :model-value="editRemoteCredentials?.[remote.name]"
                  :url="remote.fetch_url"
                  placeholder="选择凭证（可选）"
                  @update:model-value="(v: number | undefined) => updateEditRemoteCred(remote.name, v)"
                />
              </div>
            </el-card>
          </div>
          <el-empty v-if="editRemotes.length === 0" description="无远程仓库配置" :image-size="60" />
        </div>
      </el-form-item>

      <el-form-item v-if="editTrackingBranches.length > 0" label="分支追踪">
        <div class="tracking-branches">
          <el-tag v-for="b in editTrackingBranches" :key="b.name" size="small" style="margin: 2px 4px;">
            {{ b.name }} -> {{ b.upstream_ref }}
          </el-tag>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleSaveEdit" :loading="editSaving">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { Delete, Connection } from '@element-plus/icons-vue'
import type { RepoDTO } from '@/types/repo'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import { useRepoEditForm } from '@/composables/useRepoEditForm'

const props = defineProps<{
  visible: boolean
  repo: RepoDTO | null
  repoKey: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

// 表单状态与远端 URL 编辑逻辑统一来自 useRepoEditForm（与 EditRepoPage 共享）
const {
  editSaving, editForm, editUrlError, editRemotes, editTrackingBranches,
  editDefaultCredentialId, editRemoteCredentials, editUrlMode,
  remoteUrlModes, remoteUrlErrors,
  validateMainUrl, validateRemoteUrl,
  addEditRemote, removeEditRemote, updateEditRemoteCred, testEditRemote,
  fillFromRepo, save,
} = useRepoEditForm(() => props.repoKey)

watch(() => props.visible, (val) => {
  if (!val || !props.repo) return
  fillFromRepo(props.repo)
})

async function handleSaveEdit() {
  if (await save()) {
    emit('update:visible', false)
    emit('saved')
  }
}
</script>

<style scoped>
.url-input-group {
  display: flex;
  gap: var(--spacing-sm);
  width: 100%;
}
.url-input-group .el-input {
  flex: 1;
}
.url-mode-switch {
  flex-shrink: 0;
}
.url-mode-switch-sm {
  flex-shrink: 0;
}
.edit-remote-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}
.edit-remote-cred {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}
.edit-remote-cred .cred-label {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  flex-shrink: 0;
}
.remotes-section {
  width: 100%;
}
.remotes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: var(--font-size-md);
  color: var(--text-color-regular);
}
.tracking-branches {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}
.field-error {
  color: var(--danger-color);
  font-size: var(--font-size-xs);
  margin-top: var(--spacing-xs);
}
.is-error-input :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--danger-color) inset;
}
.edit-remote-item {
  margin-bottom: 8px;
}
</style>
