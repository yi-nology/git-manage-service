<template>
  <div class="credential-page">
    <PageHeader title="凭证管理" subtitle="统一管理 Git 仓库认证凭证">
      <template #actions>
        <ActionPill variant="primary" :icon="Plus" @click="router.push('/settings/credentials/add')">
          添加凭证
        </ActionPill>
      </template>
    </PageHeader>

    <div class="info-banner">
      <el-icon class="info-icon"><InfoFilled /></el-icon>
      <span>凭证用于 Git 仓库的认证。支持 SSH 密钥（数据库或本地文件）和 HTTP（用户名密码或 Token）。配置 URL 匹配模式后，系统将在配置仓库时自动推荐匹配的凭证。</span>
    </div>

    <LoadingState v-if="loading" />

    <EmptyState
      v-else-if="credentials.length === 0"
      title="暂无凭证"
      description="点击上方按钮创建第一个凭证"
    />

    <div v-else class="cred-grid">
      <CredentialCard
        v-for="cred in credentials"
        :key="cred.id"
        :credential="cred"
        @edit="handleEdit"
        @delete="handleDelete"
      />
    </div>

    <div v-if="credentials.length > 0" class="test-section">
      <SectionTitle title="测试凭证连接" />
      <div class="test-card">
        <div class="test-row">
          <div class="test-field">
            <label class="field-label">凭证</label>
            <select v-model="testCredId" class="field-select">
              <option :value="undefined" disabled>选择凭证</option>
              <option v-for="c in credentials" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div class="test-field" style="flex:1">
            <label class="field-label">远程 URL</label>
            <input v-model="testUrl" placeholder="git@github.com:user/repo.git" class="field-input" />
          </div>
          <button class="test-btn" @click="handleTest" :disabled="!testCredId || !testUrl">
            <span v-if="testing" class="loading-spinner-sm"></span>
            <span v-else>测试连接</span>
          </button>
        </div>
        <div v-if="testResult !== null" class="test-result" :class="testResult.success ? 'success' : 'error'">
          <el-icon v-if="testResult.success"><CircleCheck /></el-icon>
          <el-icon v-else><CircleClose /></el-icon>
          <span>{{ testResult.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, InfoFilled, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { listCredentials, deleteCredential, testCredential } from '@/api/modules/credential'
import type { CredentialDTO } from '@/types/credential'
import CredentialCard from '@/components/credential/CredentialCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const router = useRouter()
const loading = ref(false)
const credentials = ref<CredentialDTO[]>([])

const testCredId = ref<number>()
const testUrl = ref('')
const testing = ref(false)
const testResult = ref<{ success: boolean; message: string } | null>(null)

async function loadCredentials() {
  loading.value = true
  try {
    credentials.value = await listCredentials() || []
  } catch {
    credentials.value = []
  } finally {
    loading.value = false
  }
}

function handleEdit(cred: CredentialDTO) {
  router.push(`/settings/credentials/${cred.id}/edit`)
}

async function handleDelete(cred: CredentialDTO) {
  try {
    await deleteCredential(cred.id)
    ElMessage.success('凭证已删除')
    loadCredentials()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  }
}

async function handleTest() {
  if (!testCredId.value || !testUrl.value) return
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testCredential(testCredId.value, testUrl.value)
  } catch (e: any) {
    testResult.value = { success: false, message: e?.message || '测试失败' }
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadCredentials()
})
</script>

<style scoped>
.credential-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-banner {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 16px;
  border-radius: 8px;
  background: var(--accent-bg);
  border: 1px solid var(--border-color);
  font-size: 13px;
  color: var(--text-color-secondary);
  line-height: 1.6;
}

.info-icon {
  color: var(--accent-primary);
  margin-top: 2px;
  flex-shrink: 0;
}

.cred-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.test-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.test-card {
  border-radius: 12px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 20px;
}

.test-row {
  display: flex;
  gap: 16px;
  align-items: flex-end;
}

.test-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-secondary);
}

.field-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  min-width: 200px;
  outline: none;
}

.field-select:focus {
  border-color: var(--accent-primary);
}

.field-input {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-color-primary);
  background: var(--bg-color-page);
  width: 100%;
  outline: none;
}

.field-input:focus {
  border-color: var(--accent-primary);
}

.test-btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: none;
  background: var(--accent-primary);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  min-width: 90px;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 36px;
}

.test-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.test-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-spinner-sm {
  width: 14px;
  height: 14px;
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

.test-result {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-radius: 8px;
  margin-top: 12px;
  font-size: 13px;
}

.test-result.success {
  background: #ecfdf5;
  color: #10b981;
}

.test-result.error {
  background: #fef2f2;
  color: #ef4444;
}
</style>
