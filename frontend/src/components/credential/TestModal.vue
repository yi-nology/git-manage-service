<template>
  <el-dialog
    v-model="visible"
    title="测试连接"
    width="500px"
    :close-on-click-modal="!loading"
  >
    <div class="test-modal-content">
      <div class="form-field">
        <label class="field-label">远程仓库 URL</label>
        <input
          v-model="testUrl"
          type="text"
          class="field-input"
          placeholder="git@github.com:user/repo.git 或 https://..."
          :disabled="loading"
        />
        <p class="field-hint">输入要测试连接的 Git 仓库地址</p>
      </div>

      <div v-if="result" class="test-result" :class="result.success ? 'success' : 'error'">
        <div class="result-icon">
          <svg v-if="result.success" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22,4 12,14.01 9,11.01"/>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
        </div>
        <div class="result-text">
          <h4>{{ result.success ? '连接成功' : '连接失败' }}</h4>
          <p>{{ result.message }}</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="modal-footer">
        <button class="btn btn-secondary" @click="visible = false" :disabled="loading">
          关闭
        </button>
        <button class="btn btn-primary" @click="handleTest" :disabled="loading || !testUrl">
          <span v-if="loading" class="loading-spinner"></span>
          <span v-else>开始测试</span>
        </button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { CredentialDTO } from '@/types/credential'
import { testCredential } from '@/api/modules/credential'

const props = defineProps<{
  credential: CredentialDTO | null
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const testUrl = ref('')
const loading = ref(false)
const result = ref<{ success: boolean; message: string } | null>(null)

watch(
  () => props.visible,
  (val) => {
    if (!val) {
      testUrl.value = ''
      result.value = null
    }
  }
)

async function handleTest() {
  if (!props.credential || !testUrl.value) return

  loading.value = true
  result.value = null

  try {
    result.value = await testCredential(props.credential.id, testUrl.value)
  } catch (e: any) {
    result.value = {
      success: false,
      message: e?.message || '测试失败',
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.test-modal-content {
  padding: 10px 0;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.field-input {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 14px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  outline: none;
  transition: border-color 0.2s;
}

.field-input:focus {
  border-color: var(--accent-primary);
}

.field-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.field-hint {
  font-size: 12px;
  color: var(--text-color-tertiary);
  margin: 0;
}

.test-result {
  margin-top: 20px;
  padding: 16px;
  border-radius: 8px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.test-result.success {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}

.test-result.error {
  background: #fef2f2;
  border: 1px solid #fecaca;
}

.result-icon {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
}

.test-result.success .result-icon {
  color: #22c55e;
}

.test-result.error .result-icon {
  color: #ef4444;
}

.result-text h4 {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 4px;
}

.test-result.success h4 {
  color: #16a34a;
}

.test-result.error h4 {
  color: #dc2626;
}

.result-text p {
  font-size: 13px;
  color: var(--text-color-secondary);
  margin: 0;
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
