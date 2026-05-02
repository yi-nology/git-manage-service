<template>
  <div class="ssh-keys-page">
    <PageHeader title="SSH 密钥管理" subtitle="管理用于 Git 仓库认证的 SSH 密钥">
      <template #actions>
        <ActionPill variant="primary" :icon="Plus" @click="handleCreate">
          添加密钥
        </ActionPill>
      </template>
    </PageHeader>

    <DataTable :columns="columns" :data="sshKeys" row-key="id" :loading="loading">
      <template #cell-name="{ row }">
        <span class="name-cell">{{ row.name }}</span>
      </template>
      <template #cell-key_type="{ row }">
        <StatusBadge
          :variant="keyTypeClass(row.key_type)"
          :text="keyTypeLabel(row.key_type)"
          :show-dot="false"
        />
      </template>
      <template #cell-fingerprint="{ row }">
        <span class="fingerprint-cell">{{ row.key_type ? row.key_type.toUpperCase() : '-' }}</span>
      </template>
      <template #cell-created_at="{ row }">
        {{ formatDate(row.created_at) }}
      </template>
      <template #row-actions="{ row }">
        <ActionPill variant="outline" small @click="handleDetail(row)">查看</ActionPill>
        <ActionPill variant="primary" small @click="handleEdit(row)">编辑</ActionPill>
        <ActionPill variant="green" small @click="handleTest(row)">测试</ActionPill>
        <ActionPill variant="danger" small @click="handleDelete(row)">删除</ActionPill>
      </template>
      <template #empty>
        <EmptyState title="暂无 SSH 密钥" description="点击上方按钮添加第一把密钥" />
      </template>
    </DataTable>

    <SSHKeyDialogs ref="dialogsRef" @changed="fetchSSHKeys" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  listDBSSHKeys,
  type DBSSHKey,
} from '@/api/modules/sshkey'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { TableColumn } from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SSHKeyDialogs from '@/components/settings/SSHKeyDialogs.vue'

type StatusVariant = 'success' | 'danger' | 'warning' | 'info' | 'running' | 'default'

const columns: TableColumn[] = [
  { key: 'name', label: '名称', width: '160px' },
  { key: 'key_type', label: '类型', width: '100px' },
  { key: 'fingerprint', label: '指纹 (Fingerprint)', flex: 1 },
  { key: 'created_at', label: '创建时间', width: '160px' },
]

const loading = ref(false)
const sshKeys = ref<DBSSHKey[]>([])
const dialogsRef = ref<InstanceType<typeof SSHKeyDialogs>>()

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

function keyTypeClass(t: string): StatusVariant {
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

function handleCreate() {
  dialogsRef.value?.openCreateDialog()
}

function handleEdit(key: DBSSHKey) {
  dialogsRef.value?.openEditDialog(key)
}

function handleDetail(key: DBSSHKey) {
  dialogsRef.value?.openDetailDialog(key)
}

function handleTest(key: DBSSHKey) {
  dialogsRef.value?.openTestDialog(key)
}

async function handleDelete(key: DBSSHKey) {
  await dialogsRef.value?.handleDelete(key)
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
</style>
