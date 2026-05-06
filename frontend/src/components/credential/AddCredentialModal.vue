<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑凭证' : '添加凭证'"
    width="600px"
    :close-on-click-modal="false"
  >
    <div class="form-content">
      <div class="form-row">
        <div class="form-field half">
          <label class="field-label">凭证名称 <span class="required">*</span></label>
          <input
            v-model="form.name"
            type="text"
            class="field-input"
            placeholder="例如：GitHub 个人密钥"
            :disabled="loading"
          />
        </div>
        <div class="form-field half">
          <label class="field-label">凭证类型 <span class="required">*</span></label>
          <select v-model="form.type" class="field-select" :disabled="loading">
            <option value="ssh_key">SSH 密钥</option>
            <option value="http_basic">HTTP 用户名密码</option>
            <option value="http_token">HTTP Token</option>
            <option value="platform_token">平台 API Token</option>
          </select>
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">描述</label>
        <textarea
          v-model="form.description"
          class="field-textarea"
          placeholder="凭证的用途说明（可选）"
          rows="3"
          :disabled="loading"
        />
      </div>

      <div v-if="form.type === 'ssh_key'" class="ssh-key-section">
        <div class="toggle-group">
          <button
            class="toggle-btn"
            :class="{ active: !useLocalFile }"
            @click="useLocalFile = false"
            :disabled="loading"
          >
            选择密钥库
          </button>
          <button
            class="toggle-btn"
            :class="{ active: useLocalFile }"
            @click="useLocalFile = true"
            :disabled="loading"
          >
            本地文件路径
          </button>
        </div>

        <div v-if="!useLocalFile" class="form-field">
          <label class="field-label">选择 SSH 密钥 <span class="required">*</span></label>
          <select v-model="form.ssh_key_id" class="field-select" :disabled="loading">
            <option :value="0">请选择</option>
            <option v-for="key in sshKeys" :key="key.id" :value="key.id">{{ key.name }}</option>
          </select>
        </div>

        <div v-else class="form-field">
          <label class="field-label">密钥文件路径 <span class="required">*</span></label>
          <input
            v-model="form.ssh_key_path"
            type="text"
            class="field-input"
            placeholder="/Users/username/.ssh/id_rsa"
            :disabled="loading"
          />
        </div>
      </div>

      <div v-if="form.type === 'http_basic'" class="http-section">
        <div class="form-field">
          <label class="field-label">用户名 <span class="required">*</span></label>
          <input
            v-model="form.username"
            type="text"
            class="field-input"
            placeholder="Git 用户名"
            :disabled="loading"
          />
        </div>
        <div class="form-field">
          <label class="field-label">密码 / Personal Access Token <span class="required">*</span></label>
          <input
            v-model="form.secret"
            type="password"
            class="field-input"
            placeholder="********"
            :disabled="loading"
          />
        </div>
      </div>

      <div v-if="form.type === 'http_token' || form.type === 'platform_token'" class="token-section">
        <div class="form-field">
          <label class="field-label">
            {{ form.type === 'platform_token' ? 'API Token' : 'Token' }}
            <span class="required">*</span>
          </label>
          <input
            v-model="form.secret"
            type="password"
            class="field-input"
            placeholder="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
            :disabled="loading"
          />
        </div>
        <div v-if="form.type === 'http_token'" class="form-field">
          <label class="field-label">用户名（可选）</label>
          <input
            v-model="form.username"
            type="text"
            class="field-input"
            placeholder="Git 用户名（部分服务端需要）"
            :disabled="loading"
          />
        </div>
      </div>

      <div class="form-field">
        <label class="field-label">URL 匹配模式</label>
        <input
          v-model="form.url_pattern"
          type="text"
          class="field-input"
          placeholder="*.github.com - 配置后将在添加仓库时自动推荐"
          :disabled="loading"
        />
        <p class="field-hint">使用通配符匹配，例如：*.github.com、gitlab.com/*</p>
      </div>
    </div>

    <template #footer>
      <div class="modal-footer">
        <button class="btn btn-secondary" @click="visible = false" :disabled="loading">
          取消
        </button>
        <button class="btn btn-primary" @click="handleSubmit" :disabled="loading || !canSubmit">
          <span v-if="loading" class="loading-spinner"></span>
          <span v-else>{{ isEdit ? '保存修改' : '添加凭证' }}</span>
        </button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import type { CredentialDTO, CreateCredentialReq, UpdateCredentialReq } from '@/types/credential'
import { createCredential, updateCredential, getCredential } from '@/api/modules/credential'
import { listSSHKeys } from '@/api/modules/ssh-key'
import { ElMessage } from 'element-plus'

interface SSHKey {
  id: number
  name: string
}

const props = defineProps<{
  credentialId?: number | null
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  success: []
}>()

const loading = ref(false)
const isEdit = computed(() => !!props.credentialId)
const useLocalFile = ref(false)
const sshKeys = ref<SSHKey[]>([])

const form = ref<CreateCredentialReq>({
  name: '',
  type: 'ssh_key',
  description: '',
  ssh_key_id: 0,
  ssh_key_path: '',
  username: '',
  secret: '',
  url_pattern: '',
})

const canSubmit = computed(() => {
  if (!form.value.name) return false
  if (form.value.type === 'ssh_key') {
    return form.value.ssh_key_id > 0 || !!form.value.ssh_key_path
  }
  return !!form.value.secret
})

watch(
  () => props.visible,
  async (val) => {
    if (val) {
      useLocalFile.value = false
      await loadSSHKeys()

      if (props.credentialId) {
        await loadCredential(props.credentialId)
      } else {
        form.value = {
          name: '',
          type: 'ssh_key',
          description: '',
          ssh_key_id: 0,
          ssh_key_path: '',
          username: '',
          secret: '',
          url_pattern: '',
        }
      }
    }
  }
)

async function loadSSHKeys() {
  try {
    sshKeys.value = await listSSHKeys()
  } catch {
    sshKeys.value = []
  }
}

async function loadCredential(id: number) {
  loading.value = true
  try {
    const cred = await getCredential(id)
    form.value = {
      name: cred.name,
      type: cred.type,
      description: cred.description,
      ssh_key_id: cred.ssh_key_id || 0,
      ssh_key_path: cred.ssh_key_path || '',
      username: cred.username || '',
      secret: '',
      url_pattern: cred.url_pattern || '',
    }
    useLocalFile.value = !!cred.ssh_key_path
  } catch (e: any) {
    ElMessage.error(e?.message || '加载凭证失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!canSubmit.value) return

  loading.value = true
  try {
    const data = {
      name: form.value.name,
      type: form.value.type,
      description: form.value.description,
      ssh_key_id: form.value.ssh_key_id,
      ssh_key_path: useLocalFile.value ? form.value.ssh_key_path : '',
      username: form.value.username,
      secret: form.value.secret,
      url_pattern: form.value.url_pattern,
    }

    if (isEdit.value) {
      await updateCredential(props.credentialId!, data as UpdateCredentialReq)
      ElMessage.success('凭证已更新')
    } else {
      await createCredential(data as CreateCredentialReq)
      ElMessage.success('凭证已创建')
    }

    emit('success')
    emit('update:visible', false)
  } catch (e: any) {
    ElMessage.error(e?.message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.form-content {
  padding: 10px 0;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.form-field.half {
  flex: 1;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.field-label .required {
  color: #ef4444;
}

.field-input,
.field-select,
.field-textarea {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 14px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  transition: border-color 0.2s;
}

.field-input:focus,
.field-select:focus,
.field-textarea:focus {
  border-color: var(--accent-primary);
}

.field-input:disabled,
.field-select:disabled,
.field-textarea:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.field-hint {
  font-size: 12px;
  color: var(--text-color-tertiary);
  margin: 0;
}

.toggle-group {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.toggle-btn {
  flex: 1;
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-color-page);
  color: var(--text-color-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn:hover:not(:disabled) {
  border-color: var(--accent-primary);
}

.toggle-btn.active {
  background: var(--accent-primary);
  color: #fff;
  border-color: var(--accent-primary);
}

.toggle-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn {
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-secondary {
  background: var(--bg-color-tertiary);
  color: var(--text-color-primary);
}

.btn-secondary:hover {
  background: var(--bg-color-hover);
}

.btn-primary {
  background: var(--accent-primary);
  color: #fff;
}

.btn-primary:hover {
  opacity: 0.9;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
