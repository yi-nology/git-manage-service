<template>
  <div class="cli-config-page">
    <PageHeader title="CLI 配置管理" subtitle="管理代码审查使用的 CLI 工具配置" show-back back-route="/settings" />

    <div class="section-header">
      <div class="section-info">
        <h3>已配置的 CLI 工具</h3>
        <p class="section-desc">管理已添加的 CLI 工具，点击扫描可自动检测本地安装</p>
      </div>
      <div class="header-actions">
        <ActionPill variant="default" @click="openScanDialog" :loading="scanning">扫描本地安装</ActionPill>
        <ActionPill variant="primary" :icon="Plus" @click="openAddDialog">手动添加</ActionPill>
      </div>
    </div>

    <DataTable :columns="configColumns" :data="configs" :loading="loading" row-key="id">
      <template #cell-name="{ row }">
        <span style="font-weight:500">{{ row.name }}</span>
      </template>
      <template #cell-cli_type="{ row }">
        <StatusBadge :variant="cliTypeVariant(row.cliType)" :text="cliTypeLabel(row.cliType)" />
      </template>
      <template #cell-exec_path="{ row }">
        <code class="path-code">{{ row.execPath }}</code>
      </template>
      <template #cell-is_active="{ row }">
        <StatusBadge :variant="row.isActive ? 'success' : 'default'" :text="row.isActive ? '启用' : '禁用'" />
      </template>
      <template #row-actions="{ row }">
        <button class="act-btn act-btn--green" @click="handleTest(row)">测试</button>
        <button class="act-btn act-btn--primary" @click="openEditDialog(row)">编辑</button>
        <button class="act-btn act-btn--danger" @click="handleDelete(row)">删除</button>
      </template>
    </DataTable>

    <el-dialog v-model="showDialog" :title="editing ? '编辑 CLI 配置' : '添加 CLI 配置'" width="520px" destroy-on-close @close="resetForm">
      <el-form label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="例如: Claude Code Review" />
        </el-form-item>
        <el-form-item label="CLI 类型">
          <el-select v-model="form.cli_type" style="width:100%">
            <el-option label="Claude Code" value="claude_cli" />
            <el-option label="OpenCode" value="opencode_cli" />
            <el-option label="Qoder" value="qoder_cli" />
            <el-option label="Codex CLI" value="codex_cli" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行路径">
          <el-input v-model="form.exec_path" placeholder="例如: /usr/local/bin/claude" />
        </el-form-item>
        <el-form-item label="配置 JSON">
          <el-input v-model="form.config_json" type="textarea" :rows="4" placeholder='可选，例如: {"timeout": 300}' />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showScanDialog" title="扫描本地 CLI 工具" width="600px" destroy-on-close>
      <div v-if="scanning" class="scan-loading">
        <LoadingState text="正在扫描本地安装的 CLI 工具..." />
      </div>
      <div v-else-if="scanResults.length === 0" class="scan-empty">
        <EmptyState title="未检测到 CLI 工具" description="请先安装 Claude Code、OpenCode、Qoder 或 Codex CLI" />
      </div>
      <div v-else class="scan-results">
        <div v-for="item in scanResults" :key="item.cliType" class="scan-item">
          <div class="scan-item-info">
            <div class="scan-item-header">
              <StatusBadge :variant="item.isInstalled ? cliTypeVariant(item.cliType) : 'default'" :text="cliTypeLabel(item.cliType)" />
              <span v-if="item.isInstalled" class="scan-badge scan-badge--ok">已安装</span>
              <span v-else class="scan-badge scan-badge--missing">未安装</span>
            </div>
            <div v-if="item.isInstalled" class="scan-item-detail">
              <code class="path-code">{{ item.execPath }}</code>
              <span v-if="item.version" class="scan-version">{{ item.version }}</span>
            </div>
            <div v-else class="scan-item-detail">
              <span class="scan-hint">系统 PATH 中未找到此工具</span>
            </div>
          </div>
          <button
            v-if="item.isInstalled && !isAlreadyConfigured(item.cliType)"
            class="act-btn act-btn--primary"
            @click="handleQuickAdd(item)"
          >添加</button>
          <span v-else-if="isAlreadyConfigured(item.cliType)" class="scan-badge scan-badge--ok">已配置</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="showScanDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import DataTable from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingState from '@/components/common/LoadingState.vue'
import {
  listCLIConfigs,
  createCLIConfig,
  updateCLIConfig,
  deleteCLIConfig,
  testCLIConfig,
  scanCLIs,
} from '@/api/modules/review'
import type { ReviewCLIConfigDTO, ScannedCLI } from '@/api/modules/review'
import type { TableColumn } from '@/components/common/DataTable.vue'

const loading = ref(false)
const configs = ref<ReviewCLIConfigDTO[]>([])
const showDialog = ref(false)
const saving = ref(false)
const editing = ref<ReviewCLIConfigDTO | null>(null)
const showScanDialog = ref(false)
const scanning = ref(false)
const scanResults = ref<ScannedCLI[]>([])

const form = ref({
  name: '',
  cli_type: 'claude_cli',
  exec_path: '',
  config_json: '',
  is_active: true,
})

const configColumns: TableColumn[] = [
  { key: 'name', label: '名称', width: '160px' },
  { key: 'cli_type', label: '类型', width: '130px' },
  { key: 'exec_path', label: '执行路径', flex: 1 },
  { key: 'is_active', label: '状态', width: '80px' },
]

const cliTypeLabels: Record<string, string> = {
  claude_cli: 'Claude Code',
  opencode_cli: 'OpenCode',
  qoder_cli: 'Qoder',
  codex_cli: 'Codex CLI',
}

function cliTypeLabel(type: string): string {
  return cliTypeLabels[type] || type
}

function cliTypeVariant(type: string): string {
  const map: Record<string, string> = {
    claude_cli: 'purple',
    opencode_cli: 'blue',
    qoder_cli: 'green',
    codex_cli: 'teal',
  }
  return map[type] || 'default'
}

function isAlreadyConfigured(cliType: string): boolean {
  return configs.value.some(c => c.cliType === cliType)
}

async function loadConfigs() {
  loading.value = true
  try {
    const res = await listCLIConfigs()
    configs.value = Array.isArray(res) ? res : (res as any)?.configs || []
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function openScanDialog() {
  showScanDialog.value = true
  scanning.value = true
  try {
    const res = await scanCLIs()
    scanResults.value = Array.isArray(res) ? res : []
  } catch (e: any) {
    ElMessage.error(e.message || '扫描失败')
    scanResults.value = []
  } finally {
    scanning.value = false
  }
}

async function handleQuickAdd(item: ScannedCLI) {
  saving.value = true
  try {
    await createCLIConfig({
      name: item.name,
      cli_type: item.cliType,
      exec_path: item.execPath,
      is_active: true,
    })
    ElMessage.success(`${item.name} 添加成功`)
    await loadConfigs()
  } catch (e: any) {
    ElMessage.error(e.message || '添加失败')
  } finally {
    saving.value = false
  }
}

function openAddDialog() {
  editing.value = null
  form.value = { name: '', cli_type: 'claude_cli', exec_path: '', config_json: '', is_active: true }
  showDialog.value = true
}

function openEditDialog(row: ReviewCLIConfigDTO) {
  editing.value = row
  form.value = {
    name: row.name,
    cli_type: row.cliType,
    exec_path: row.execPath,
    config_json: row.configJson || '',
    is_active: row.isActive,
  }
  showDialog.value = true
}

function resetForm() {
  editing.value = null
}

async function handleSave() {
  if (!form.value.name || !form.value.cli_type || !form.value.exec_path) {
    ElMessage.warning('请填写名称、CLI 类型和执行路径')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateCLIConfig(editing.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await createCLIConfig(form.value)
      ElMessage.success('添加成功')
    }
    showDialog.value = false
    await loadConfigs()
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: ReviewCLIConfigDTO) {
  try {
    await ElMessageBox.confirm(`确定删除 CLI 配置 "${row.name}"？`, '确认删除', { type: 'warning' })
  } catch { return }
  try {
    await deleteCLIConfig(row.id)
    ElMessage.success('已删除')
    await loadConfigs()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

async function handleTest(row: ReviewCLIConfigDTO) {
  try {
    const res = await testCLIConfig(row.id) as any
    if (res.success) {
      ElMessage.success(`CLI 可用${res.version ? ` (版本: ${res.version})` : ''}`)
    } else {
      ElMessage.warning(`CLI 不可用: ${res.message}`)
    }
  } catch (e: any) {
    ElMessage.error(e.message || '测试失败')
  }
}

onMounted(() => { loadConfigs() })
</script>

<style scoped>
.cli-config-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.section-info h3 {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-color-primary);
}

.section-desc {
  margin: 4px 0 0;
  font-size: var(--font-size-sm);
  color: var(--text-color-secondary);
}

.path-code {
  font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
  font-size: var(--font-size-xs);
  background: var(--surface-card);
  padding: 2px 6px;
  border-radius: 4px;
  color: var(--text-color-regular);
}

.scan-loading,
.scan-empty {
  padding: 32px 0;
}

.scan-results {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.scan-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--surface-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-md);
}

.scan-item-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.scan-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.scan-item-detail {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-left: 4px;
}

.scan-version {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
}

.scan-hint {
  font-size: var(--font-size-xs);
  color: var(--text-color-placeholder);
}

.scan-badge {
  font-size: var(--font-size-xs);
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.scan-badge--ok {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.scan-badge--missing {
  background: rgba(107, 114, 128, 0.1);
  color: #6B7280;
}
</style>
