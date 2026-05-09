<template>
  <el-card>
    <template #header>
      <div class="card-header-row">
        <span>版本标签管理</span>
        <div class="header-actions">
          <el-button size="small" @click="handleFetchTags" :loading="fetchTagsLoading">
            <el-icon><Download /></el-icon> 拉取远端 Tags
          </el-button>
          <el-button size="small" type="primary" @click="openCreateTagDialog">
            <el-icon><Plus /></el-icon> 创建 Tag
          </el-button>
          <el-button size="small" @click="emit('reload')">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </div>
    </template>
    <div v-if="versionList.length === 0 && !versionsLoading">
      <el-empty description="暂无版本标签">
        <el-button type="primary" @click="openCreateTagDialog">创建第一个 Tag</el-button>
      </el-empty>
    </div>
    <el-table v-else :data="versionList" v-loading="versionsLoading" stripe border size="small">
      <el-table-column prop="name" label="标签名称" width="160">
        <template #default="{ row }">
          <el-tag type="success" size="small">{{ row.name }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="hash" label="Commit" width="120">
        <template #default="{ row }">
          <el-text class="mono-text" size="small">{{ row.hash?.substring(0, 8) }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="tagger" label="作者" width="120" />
      <el-table-column prop="date" label="日期" width="160">
        <template #default="{ row }">{{ formatDate(row.date) }}</template>
      </el-table-column>
      <el-table-column prop="message" label="说明" min-width="200" show-overflow-tooltip />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button-group size="small">
            <el-button @click="handlePushTag(row.name)">
              <el-icon><Top /></el-icon> 推送
            </el-button>
            <el-button @click="handleCopyHash(row.hash)">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
            <el-button type="danger" @click="handleDeleteTag(row.name)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="showCreateTagDialog" title="创建版本标签" width="550px" destroy-on-close>
    <el-form :model="createTagForm" label-width="100px">
      <el-form-item label="版本类型">
        <el-radio-group v-model="createTagForm.versionType" @change="handleVersionTypeChange">
          <el-radio-button value="patch">Patch (修复)</el-radio-button>
          <el-radio-button value="minor">Minor (功能)</el-radio-button>
          <el-radio-button value="major">Major (大版本)</el-radio-button>
          <el-radio-button value="custom">自定义</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="当前版本" v-if="nextVersionInfo">
        <el-tag type="info">{{ nextVersionInfo.current || '无' }}</el-tag>
      </el-form-item>
      <el-form-item label="标签名称" required>
        <el-input v-model="createTagForm.name" :disabled="createTagForm.versionType !== 'custom'" placeholder="v1.0.0" />
      </el-form-item>
      <el-form-item label="目标引用">
        <el-input v-model="createTagForm.ref" placeholder="HEAD (默认当前分支最新提交)" />
      </el-form-item>
      <el-form-item label="说明">
        <el-input v-model="createTagForm.message" type="textarea" :rows="2" placeholder="版本说明" />
      </el-form-item>
      <el-form-item label="推送到远端">
        <el-select v-model="createTagForm.push_remote" placeholder="不推送" clearable>
          <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showCreateTagDialog = false">取消</el-button>
      <el-button type="primary" @click="handleCreateTag" :loading="createTagLoading">创建</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="showPushTagDialog" :title="`推送标签: ${pushTagName}`" width="480px" destroy-on-close>
    <el-form label-width="90px">
      <el-form-item label="目标远端">
        <el-select v-model="pushTagRemote" placeholder="选择远端">
          <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showPushTagDialog = false">取消</el-button>
      <el-button type="primary" @click="handleSubmitPushTag" :loading="pushTagLoading">确认推送</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Plus, Refresh, Top, Delete, CopyDocument } from '@element-plus/icons-vue'
import { fetchRepo } from '@/api/modules/repo'
import { getNextVersion } from '@/api/modules/version'
import type { VersionTag, NextVersionInfo } from '@/api/modules/version'
import { createTag, deleteTag, pushTag } from '@/api/modules/branch'
import { showGitError } from '@/utils/git'
import { formatDate } from '@/utils/format'

const props = defineProps<{
  repoKey: string
  remoteNames: string[]
  versionList: VersionTag[]
  versionsLoading: boolean
}>()

const emit = defineEmits<{
  reload: []
  versionChanged: []
}>()

const fetchTagsLoading = ref(false)
const showCreateTagDialog = ref(false)
const createTagLoading = ref(false)
const nextVersionInfo = ref<NextVersionInfo | null>(null)
const createTagForm = ref({
  versionType: 'patch' as 'patch' | 'minor' | 'major' | 'custom',
  name: '',
  ref: 'HEAD',
  message: '',
  push_remote: '',
})
const showPushTagDialog = ref(false)
const pushTagName = ref('')
const pushTagRemote = ref('origin')
const pushTagLoading = ref(false)

async function handleFetchTags() {
  fetchTagsLoading.value = true
  try {
    await fetchRepo(props.repoKey)
    ElMessage.success('远端 Tags 拉取成功')
    emit('reload')
  } catch {
    ElMessage.error('拉取远端 Tags 失败')
  } finally {
    fetchTagsLoading.value = false
  }
}

async function openCreateTagDialog() {
  createTagForm.value = { versionType: 'patch', name: '', ref: 'HEAD', message: '', push_remote: '' }
  nextVersionInfo.value = null
  showCreateTagDialog.value = true
  try {
    nextVersionInfo.value = await getNextVersion(props.repoKey)
    handleVersionTypeChange(createTagForm.value.versionType)
  } catch { /* ignore */ }
}

function handleVersionTypeChange(type: string | number | boolean | undefined) {
  if (!nextVersionInfo.value) return
  switch (type) {
    case 'patch':
      createTagForm.value.name = nextVersionInfo.value.next_patch
      break
    case 'minor':
      createTagForm.value.name = nextVersionInfo.value.next_minor
      break
    case 'major':
      createTagForm.value.name = nextVersionInfo.value.next_major
      break
    case 'custom':
      createTagForm.value.name = ''
      break
  }
}

async function handleCreateTag() {
  if (!createTagForm.value.name) {
    ElMessage.warning('标签名称不能为空')
    return
  }
  createTagLoading.value = true
  try {
    await createTag({
      repo_key: props.repoKey,
      name: createTagForm.value.name,
      ref: createTagForm.value.ref || 'HEAD',
      message: createTagForm.value.message,
      push_remote: createTagForm.value.push_remote || undefined,
    })
    ElMessage.success(`标签 ${createTagForm.value.name} 创建成功`)
    showCreateTagDialog.value = false
    emit('reload')
    emit('versionChanged')
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('创建标签失败: ' + (err.message || '未知错误'))
  } finally {
    createTagLoading.value = false
  }
}

function handlePushTag(tagName: string) {
  pushTagName.value = tagName
  pushTagRemote.value = props.remoteNames[0] || 'origin'
  showPushTagDialog.value = true
}

async function handleSubmitPushTag() {
  pushTagLoading.value = true
  try {
    await pushTag({ repo_key: props.repoKey, tag_name: pushTagName.value, remote_name: pushTagRemote.value })
    ElMessage.success(`标签 ${pushTagName.value} 已推送到 ${pushTagRemote.value}`)
    showPushTagDialog.value = false
  } catch (e: unknown) {
    showGitError(e, '推送标签')
  } finally {
    pushTagLoading.value = false
  }
}

async function handleDeleteTag(tagName: string) {
  try {
    await ElMessageBox.confirm(`确定要删除标签 "${tagName}" 吗？`, '删除标签', {
      confirmButtonText: '仅删除本地',
      cancelButtonText: '取消',
      type: 'warning',
      distinguishCancelAndClose: true,
    })
    await deleteTag({ repo_key: props.repoKey, name: tagName })
    ElMessage.success(`标签 ${tagName} 已删除`)
    emit('reload')
    emit('versionChanged')
  } catch (action) {
    if (action === 'cancel' || action === 'close') return
    const err = action as { message?: string }
    ElMessage.error('删除标签失败: ' + (err.message || '未知错误'))
  }
}

function handleCopyHash(hash: string) {
  if (hash) {
    navigator.clipboard.writeText(hash)
    ElMessage.success('已复制 Commit Hash')
  }
}
</script>

<style scoped>
.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.mono-text {
  font-family: monospace;
}
</style>
