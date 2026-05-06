<template>
  <el-dialog
    v-model="visible"
    title="确认删除"
    width="500px"
    :close-on-click-modal="false"
    :show-close="false"
  >
    <div class="delete-modal-content">
      <div class="warning-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
          <line x1="12" y1="9" x2="12" y2="13"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
      </div>

      <div class="warning-text">
        <h3>确定要删除凭证「{{ credential?.name }}」吗？</h3>
        <p class="subtitle">此操作不可撤销，删除后所有相关配置将无法正常使用。</p>
      </div>

      <div v-if="usages && (usages.total_repo_count > 0 || usages.total_provider_count > 0)" class="usage-warning">
        <h4>该凭证当前正在被使用：</h4>
        <div class="usage-list">
          <div v-if="usages.total_repo_count > 0" class="usage-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/>
              <polyline points="13,2 13,9 20,9"/>
            </svg>
            <span>{{ usages.total_repo_count }} 个仓库</span>
          </div>
          <div v-if="usages.total_provider_count > 0" class="usage-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
            <span>{{ usages.total_provider_count }} 个平台配置</span>
          </div>
        </div>
        <p class="danger-text">删除将导致上述引用该凭证的功能失效，请谨慎操作。</p>
      </div>
    </div>

    <template #footer>
      <div class="modal-footer">
        <button class="btn btn-secondary" @click="handleCancel" :disabled="loading">
          取消
        </button>
        <button class="btn btn-danger" @click="handleConfirm" :disabled="loading">
          <span v-if="loading" class="loading-spinner"></span>
          <span v-else>确认删除</span>
        </button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { CredentialDTO, CredentialUsages } from '@/types/credential'
import { getCredentialUsages, deleteCredential } from '@/api/modules/credential'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  credential: CredentialDTO | null
  visible: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'update:loading': [value: boolean]
  confirm: []
  cancel: []
}>()

const usages = ref<CredentialUsages | null>(null)

watch(
  () => props.visible,
  async (val) => {
    if (val && props.credential) {
      try {
        usages.value = await getCredentialUsages(props.credential.id)
      } catch {
        usages.value = null
      }
    } else {
      usages.value = null
    }
  }
)

function handleCancel() {
  emit('cancel')
}

async function handleConfirm() {
  if (!props.credential) return

  emit('update:loading', true)
  try {
    await deleteCredential(props.credential.id)
    ElMessage.success('凭证已删除')
    emit('confirm')
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  } finally {
    emit('update:loading', false)
  }
}
</script>

<style scoped>
.delete-modal-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 20px 0;
}

.warning-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.warning-icon svg {
  width: 32px;
  height: 32px;
}

.warning-text h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color-primary);
  margin: 0 0 8px;
}

.warning-text .subtitle {
  font-size: 14px;
  color: var(--text-color-secondary);
  margin: 0;
}

.usage-warning {
  margin-top: 24px;
  padding: 16px;
  background: #fef2f2;
  border-radius: 8px;
  text-align: left;
  width: 100%;
}

.usage-warning h4 {
  font-size: 14px;
  font-weight: 600;
  color: #ef4444;
  margin: 0 0 12px;
}

.usage-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.usage-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.usage-item svg {
  width: 16px;
  height: 16px;
}

.danger-text {
  font-size: 12px;
  color: #ef4444;
  margin: 0;
  font-weight: 500;
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

.btn-danger {
  background: #ef4444;
  color: #fff;
}

.btn-danger:hover {
  background: #dc2626;
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
