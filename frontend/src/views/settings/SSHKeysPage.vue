<template>
  <div class="ssh-keys-page">
    <div class="title-row">
      <div class="title-left">
        <h2 class="page-title">SSH 密钥管理</h2>
        <p class="page-subtitle">管理用于 Git 仓库认证的 SSH 密钥</p>
      </div>
      <button class="add-btn" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        添加密钥
      </button>
    </div>

    <div v-if="loading" class="loading-card">
      <div class="loading-spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="sshKeys.length === 0" class="empty-card">
      <div class="empty-icon">
        <el-icon :size="32"><Unlock /></el-icon>
      </div>
      <div class="empty-text">暂无 SSH 密钥</div>
      <div class="empty-sub">点击上方按钮添加第一把密钥</div>
    </div>

    <div v-else class="table-card">
      <div class="table-header">
        <span class="th" style="width:160px">名称</span>
        <span class="th" style="width:100px">类型</span>
        <span class="th" style="flex:1">指纹 (Fingerprint)</span>
        <span class="th" style="width:160px">创建时间</span>
        <span class="th" style="width:200px">操作</span>
      </div>
      <div v-for="key in sshKeys" :key="key.id" class="table-row">
        <span class="td name-cell" style="width:160px">{{ key.name }}</span>
        <span class="td" style="width:100px">
          <span class="type-tag" :class="'type-' + keyTypeClass(key.key_type)">
            {{ keyTypeLabel(key.key_type) }}
          </span>
        </span>
        <span class="td fingerprint-cell" style="flex:1">{{ key.key_type ? key.key_type.toUpperCase() : '-' }}</span>
        <span class="td" style="width:160px">{{ formatDate(key.created_at) }}</span>
        <span class="td" style="width:200px">
          <div class="action-btns">
            <button class="action-btn btn-view" @click="showDetailDialog(key)">查看</button>
            <button class="action-btn btn-edit" @click="showEditDialog(key)">编辑</button>
            <button class="action-btn btn-test" @click="showTestDialog(key)">测试</button>
            <button class="action-btn btn-delete" @click="handleDelete(key)">删除</button>
          </div>
        </span>
      </div>
    </div>

    <!-- 创建密钥对话框 -->
    <el-dialog v-model="createDialogVisible" title="新增 SSH 密钥" width="600px" destroy-on-close>
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="例如: GitHub Personal Key" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="createForm.description" placeholder="可选，用于备注密钥用途" />
        </el-form-item>
        <el-form-item label="添加方式">
          <div class="key-input-mode">
            <button type="button" class="mode-btn" :class="{ active: keyInputMode === 'paste' }" @click="keyInputMode = 'paste'">粘贴内容</button>
            <button type="button" class="mode-btn" :class="{ active: keyInputMode === 'file' }" @click="keyInputMode = 'file'">选择文件</button>
          </div>
        </el-form-item>
        <el-form-item v-show="keyInputMode === 'paste'" label="私钥" prop="private_key">
          <el-input
            v-model="createForm.private_key"
            type="textarea"
            :rows="8"
            placeholder="粘贴 SSH 私钥内容（以 -----BEGIN 开头）"
          />
        </el-form-item>
        <el-form-item v-show="keyInputMode === 'file'" label="私钥文件">
            <input
              ref="fileInputRef"
              type="file"
              style="display:none"
              @change="handleFileSelect"
            />
          <div v-if="!selectedFileName" class="file-drop-zone" @click="triggerFileInput">
            <el-icon :size="24" style="color:var(--text-color-placeholder)"><Upload /></el-icon>
            <span>点击选择私钥文件</span>
            <span class="file-hint">支持 id_rsa、id_ed25519、id_ecdsa 等格式</span>
          </div>
          <div v-else class="file-selected" @click="triggerFileInput">
            <el-icon style="color:#6366F1"><Document /></el-icon>
            <span class="file-name">{{ selectedFileName }}</span>
            <button type="button" class="file-remove" @click.stop="clearFile">移除</button>
          </div>
        </el-form-item>
        <el-form-item label="密码短语" prop="passphrase">
          <el-input
            v-model="createForm.passphrase"
            type="password"
            show-password
            placeholder="如果私钥有密码保护，请输入"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑密钥对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑 SSH 密钥" width="600px" destroy-on-close>
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="名称">
          <el-input :model-value="editForm.name" disabled />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" placeholder="可选，用于备注密钥用途" />
        </el-form-item>
        <el-form-item label="更换私钥">
          <el-input
            v-model="editForm.private_key"
            type="textarea"
            :rows="6"
            placeholder="留空则不修改私钥；填入新私钥将自动重新提取公钥和密钥类型"
          />
        </el-form-item>
        <el-form-item label="密码短语">
          <el-input
            v-model="editForm.passphrase"
            type="password"
            show-password
            placeholder="留空则不修改密码短语"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEdit" :loading="editing">保存</el-button>
      </template>
    </el-dialog>

    <!-- 查看密钥详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="密钥详情" width="600px">
      <el-descriptions :column="1" border v-if="currentKey">
        <el-descriptions-item label="名称">{{ currentKey.name }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ currentKey.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <span class="type-tag" :class="'type-' + keyTypeClass(currentKey.key_type)">
            {{ keyTypeLabel(currentKey.key_type) }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="密码保护">
          {{ currentKey.has_passphrase ? '是' : '否' }}
        </el-descriptions-item>
        <el-descriptions-item label="公钥">
          <el-input
            :model-value="currentKey.public_key"
            type="textarea"
            :rows="4"
            readonly
          />
          <el-button size="small" @click="copyPublicKey" style="margin-top: 8px;">
            <el-icon><CopyDocument /></el-icon>
            复制公钥
          </el-button>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentKey.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(currentKey.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 测试连接对话框 -->
    <el-dialog v-model="testDialogVisible" title="测试 SSH 连接" width="500px">
      <el-form :model="testForm" label-width="80px">
        <el-form-item label="Git URL">
          <el-input
            v-model="testForm.url"
            placeholder="例如: git@github.com:user/repo.git"
          />
        </el-form-item>
      </el-form>
      <div v-if="testResult" class="test-result" :class="testResult.success ? 'success' : 'error'">
        <el-icon v-if="testResult.success"><CircleCheck /></el-icon>
        <el-icon v-else><CircleClose /></el-icon>
        <span>{{ testResult.message }}</span>
      </div>
      <template #footer>
        <el-button @click="testDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleTest" :loading="testing">测试连接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormRules } from 'element-plus'
import { Plus, Unlock, CopyDocument, CircleCheck, CircleClose, Upload, Document } from '@element-plus/icons-vue'
import {
  listDBSSHKeys,
  createDBSSHKey,
  updateDBSSHKey,
  deleteDBSSHKey,
  testDBSSHKey,
  type DBSSHKey,
  type CreateDBSSHKeyReq,
  type UpdateDBSSHKeyReq,
  type TestDBSSHKeyResp
} from '@/api/modules/sshkey'

const loading = ref(false)
const sshKeys = ref<DBSSHKey[]>([])

const createDialogVisible = ref(false)
const creating = ref(false)
const keyInputMode = ref<'paste' | 'file'>('paste')
const fileInputRef = ref<HTMLInputElement>()
const selectedFileName = ref('')
const createForm = ref<CreateDBSSHKeyReq>({
  name: '',
  description: '',
  private_key: '',
  passphrase: '',
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入密钥名称', trigger: 'blur' }],
  private_key: [{ required: true, message: '请输入私钥内容', trigger: 'blur' }],
}

const editDialogVisible = ref(false)
const editing = ref(false)
const editingKeyId = ref<number>(0)
const editForm = ref<UpdateDBSSHKeyReq & { name: string }>({
  name: '',
  description: '',
  private_key: '',
  passphrase: '',
})

const detailDialogVisible = ref(false)
const currentKey = ref<DBSSHKey | null>(null)

const testDialogVisible = ref(false)
const testing = ref(false)
const testForm = ref({ url: '' })
const testResult = ref<TestDBSSHKeyResp | null>(null)
const testKeyId = ref<number>(0)

const KEY_TYPE_LABELS: Record<string, string> = {
  rsa: 'RSA',
  ed25519: 'Ed25519',
  ecdsa: 'ECDSA',
  dsa: 'DSA',
  unknown: '未知',
}

function keyTypeLabel(t: string): string {
  if (!t) return '未知'
  return KEY_TYPE_LABELS[t.toLowerCase()] ?? t.toUpperCase()
}

function keyTypeClass(t: string): string {
  if (!t) return 'info'
  const lower = t.toLowerCase()
  if (lower === 'ed25519') return 'success'
  if (lower === 'rsa') return 'info'
  if (lower === 'ecdsa') return 'warning'
  if (lower === 'dsa') return 'default'
  return 'default'
}

onMounted(() => {
  fetchSSHKeys()
})

async function fetchSSHKeys() {
  loading.value = true
  try {
    sshKeys.value = await listDBSSHKeys() || []
  } catch {
    ElMessage.error('获取 SSH 密钥列表失败')
    sshKeys.value = []
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  createForm.value = { name: '', description: '', private_key: '', passphrase: '' }
  keyInputMode.value = 'paste'
  selectedFileName.value = ''
  createDialogVisible.value = true
}

function triggerFileInput() {
  fileInputRef.value?.click()
}

function clearFile() {
  selectedFileName.value = ''
  createForm.value.private_key = ''
  if (fileInputRef.value) fileInputRef.value.value = ''
}

function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  selectedFileName.value = file.name
  const reader = new FileReader()
  reader.onload = (e) => {
    createForm.value.private_key = (e.target?.result as string) || ''
  }
  reader.onerror = () => {
    ElMessage.error('读取文件失败')
    clearFile()
  }
  reader.readAsText(file)
}

async function handleCreate() {
  if (!createForm.value.name) {
    ElMessage.warning('请输入密钥名称')
    return
  }
  if (keyInputMode.value === 'paste' && !createForm.value.private_key) {
    ElMessage.warning('请输入私钥内容')
    return
  }
  if (keyInputMode.value === 'file' && !createForm.value.private_key) {
    ElMessage.warning('请选择私钥文件')
    return
  }

  creating.value = true
  try {
    await createDBSSHKey(createForm.value)
    ElMessage.success('SSH 密钥创建成功')
    createDialogVisible.value = false
    fetchSSHKeys()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || '创建失败')
  } finally {
    creating.value = false
  }
}

function showEditDialog(key: DBSSHKey) {
  editingKeyId.value = key.id
  editForm.value = {
    name: key.name,
    description: key.description || '',
    private_key: '',
    passphrase: '',
  }
  editDialogVisible.value = true
}

async function handleEdit() {
  editing.value = true
  try {
    const payload: UpdateDBSSHKeyReq = {
      description: editForm.value.description,
    }
    if (editForm.value.private_key) payload.private_key = editForm.value.private_key
    if (editForm.value.passphrase) payload.passphrase = editForm.value.passphrase

    await updateDBSSHKey(editingKeyId.value, payload)
    ElMessage.success('SSH 密钥已更新')
    editDialogVisible.value = false
    fetchSSHKeys()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || '更新失败')
  } finally {
    editing.value = false
  }
}

function showDetailDialog(key: DBSSHKey) {
  currentKey.value = key
  detailDialogVisible.value = true
}

function showTestDialog(key: DBSSHKey) {
  testKeyId.value = key.id
  testForm.value.url = ''
  testResult.value = null
  testDialogVisible.value = true
}

async function handleTest() {
  if (!testForm.value.url) {
    ElMessage.warning('请输入 Git URL')
    return
  }

  testing.value = true
  testResult.value = null
  try {
    testResult.value = await testDBSSHKey(testKeyId.value, { url: testForm.value.url })
  } catch (e: any) {
    testResult.value = {
      success: false,
      message: e?.response?.data?.msg || '测试失败'
    }
  } finally {
    testing.value = false
  }
}

async function handleDelete(key: DBSSHKey) {
  await ElMessageBox.confirm(`确定要删除密钥 "${key.name}" 吗？此操作不可恢复。`, '确认删除', {
    type: 'warning',
  })

  try {
    await deleteDBSSHKey(key.id)
    ElMessage.success('删除成功')
    fetchSSHKeys()
  } catch {
    ElMessage.error('删除失败')
  }
}

function copyPublicKey() {
  if (!currentKey.value) return
  navigator.clipboard.writeText(currentKey.value.public_key)
  ElMessage.success('公钥已复制到剪贴板')
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<style scoped>
.ssh-keys-page {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.page-subtitle {
  margin: 0;
  font-size: 13px;
  font-weight: normal;
  color: var(--text-color-secondary);
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 8px;
  border: none;
  background: var(--accent-primary, #6366F1);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.add-btn:hover {
  opacity: 0.9;
}

.loading-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  color: var(--text-color-secondary);
  font-size: 13px;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid var(--border-color, #e5e7eb);
  border-top-color: var(--accent-primary, #6366F1);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 24px;
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
}

.empty-icon {
  color: var(--text-color-placeholder, #9ca3af);
  margin-bottom: 4px;
}

.empty-text {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.empty-sub {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.table-card {
  border-radius: 12px;
  background: var(--bg-color-page, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: var(--accent-bg, #EEF2FF);
}

.th {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color-secondary);
}

.table-row {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.table-row:last-child {
  border-bottom: none;
}

.td {
  font-size: 13px;
  color: var(--text-color-regular);
}

.name-cell {
  font-weight: 500;
  color: var(--text-color-primary);
}

.fingerprint-cell {
  font-size: 12px;
  color: var(--text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: normal;
}

.type-tag.type-success {
  background: #ECFDF5;
  color: #10B981;
}

.type-tag.type-info {
  background: #EEF2FF;
  color: #6366F1;
}

.type-tag.type-warning {
  background: #FFFBEB;
  color: #F59E0B;
}

.type-tag.type-default {
  background: var(--accent-bg, #EEF2FF);
  color: var(--text-color-secondary);
}

.action-btns {
  display: flex;
  gap: 4px;
}

.action-btn {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  border: 1px solid transparent;
  cursor: pointer;
  background: none;
  transition: all 0.2s;
}

.btn-view {
  color: var(--text-color-secondary);
  border-color: var(--border-color, #e5e7eb);
}

.btn-view:hover {
  background: var(--accent-bg, #EEF2FF);
  color: var(--accent-primary, #6366F1);
}

.btn-edit {
  color: #6366F1;
  border-color: #6366F1;
}

.btn-edit:hover {
  background: #EEF2FF;
}

.btn-test {
  color: #10B981;
  border-color: #10B981;
}

.btn-test:hover {
  background: #ECFDF5;
}

.btn-delete {
  color: #EF4444;
  border-color: #EF4444;
}

.btn-delete:hover {
  background: #FEF2F2;
}

.test-result {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-radius: 8px;
  margin-top: 12px;
}

.test-result.success {
  background-color: #ECFDF5;
  color: #10B981;
}

.test-result.error {
  background-color: #FEF2F2;
  color: #EF4444;
}

.key-input-mode {
  display: flex;
  gap: 8px;
}

.mode-btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #e5e7eb);
  background: var(--bg-color-page, #fff);
  font-size: 12px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn.active {
  background: var(--accent-primary, #6366F1);
  border-color: var(--accent-primary, #6366F1);
  color: #fff;
}

.file-drop-zone {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 28px 20px;
  border: 2px dashed var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--bg-color-page, #fff);
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
  color: var(--text-color-secondary);
}

.file-drop-zone:hover {
  border-color: var(--accent-primary, #6366F1);
  background: #FAFAFE;
}

.file-hint {
  font-size: 11px;
  color: var(--text-color-placeholder);
}

.file-selected {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--bg-color-page, #fff);
  cursor: pointer;
  transition: all 0.2s;
}

.file-selected:hover {
  border-color: var(--accent-primary, #6366F1);
}

.file-name {
  flex: 1;
  font-size: 13px;
  color: var(--text-color-primary);
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
}

.file-remove {
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid #FCA5A5;
  background: transparent;
  font-size: 11px;
  color: #EF4444;
  cursor: pointer;
  transition: all 0.2s;
}

.file-remove:hover {
  background: #FEF2F2;
}
</style>
