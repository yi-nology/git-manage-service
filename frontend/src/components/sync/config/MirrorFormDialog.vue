<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="editingMirror ? '编辑镜像' : `创建 ${createType === 'push' ? 'Push' : 'Pull'} 镜像`"
    width="700px"
    destroy-on-close
    class="mirror-dialog"
  >
    <el-form :model="form" label-width="100px">
      <el-form-item label="远程仓库">
        <el-select
          :model-value="selectedRemote"
          @update:model-value="onRemoteSelect"
          placeholder="选择当前仓库远程"
          style="width: 100%"
        >
          <el-option v-for="r in repoRemotes" :key="r.name" :label="`${r.name} (${r.url})`" :value="r">
            <span style="font-weight: 600;">{{ r.name }}</span>
            <span style="color: var(--text-color-secondary); margin-left: 8px; font-size: 12px;">{{ r.url }}</span>
          </el-option>
        </el-select>
        <div class="form-tip">从当前仓库远程列表中选择，或手动填写下方</div>
      </el-form-item>
      <el-form-item label="远程 URL" required>
        <el-input v-model="form.remote_url" placeholder="https://github.com/user/repo.git" />
      </el-form-item>
      <el-form-item label="Remote 名称">
        <el-input v-model="form.remote_name" placeholder="origin" />
      </el-form-item>
      <el-form-item label="凭据">
        <el-select v-model="form.credential_id" clearable placeholder="选择凭据（可选）" style="width: 100%">
          <el-option v-for="c in credentials" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="分支过滤">
        <el-select
          v-model="selectedBranches"
          multiple
          filterable
          allow-create
          placeholder="选择或输入分支"
          style="width: 100%"
          @change="onBranchesChange"
        >
          <el-option v-for="b in repoBranches" :key="b" :label="b" :value="b" />
        </el-select>
        <div class="form-tip">选择要同步的分支，支持手动输入 glob 模式。留空同步全部分支</div>
      </el-form-item>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="同步间隔">
            <el-input-number v-model="form.sync_interval" :min="30" :step="30" style="width: 100%" />
            <span style="margin-left: 8px">秒</span>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="Cron 表达式">
            <el-input v-model="form.cron_expr" placeholder="0 */5 * * *" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="触发设置">
        <el-checkbox v-model="form.sync_on_push">Push 事件自动触发同步</el-checkbox>
      </el-form-item>
      <el-divider content-position="left">Git 选项</el-divider>
      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item>
            <el-checkbox v-model="form.git_force">
              <span class="checkbox-label">强制推送 <span class="warning-text">⚠️</span></span>
            </el-checkbox>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item>
            <el-checkbox v-model="form.git_prune">清理已删除分支</el-checkbox>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item>
            <el-checkbox v-model="form.git_tags">同步标签</el-checkbox>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="启用状态">
        <el-switch v-model="form.enabled" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="$emit('submit')" :loading="saving">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { MirrorDTO } from '@/types/mirror'

interface RemoteItem {
  name: string
  url: string
}

interface CredentialItem {
  id: number
  name: string
}

interface MirrorForm {
  remote_url: string
  remote_name: string
  credential_id: number | null
  branch_filter: string
  sync_interval: number
  cron_expr: string
  sync_on_push: boolean
  git_force: boolean
  git_prune: boolean
  git_tags: boolean
  enabled: boolean
}

defineProps<{
  visible: boolean
  editingMirror: MirrorDTO | null
  createType: 'pull' | 'push'
  form: MirrorForm
  selectedRemote: RemoteItem | null
  selectedBranches: string[]
  credentials: CredentialItem[]
  repoRemotes: RemoteItem[]
  repoBranches: string[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'update:form': [form: MirrorForm]
  'update:selectedRemote': [remote: RemoteItem | null]
  'update:selectedBranches': [branches: string[]]
  submit: []
}>()

function onRemoteSelect(remote: RemoteItem | null) {
  emit('update:selectedRemote', remote)
}

function onBranchesChange(branches: string[]) {
  emit('update:selectedBranches', branches)
}
</script>

<style scoped>
.form-tip {
  font-size: 12px;
  color: var(--text-color-secondary);
  margin-top: 4px;
}

.checkbox-label {
  font-size: 13px;
}

.warning-text {
  color: var(--el-color-warning);
}

:deep(.mirror-dialog .el-dialog__body) {
  padding-top: 16px;
}
</style>
