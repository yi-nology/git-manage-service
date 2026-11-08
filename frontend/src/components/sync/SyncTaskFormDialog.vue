<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="isEdit ? '编辑同步任务' : '新建同步任务'"
    width="600px"
    @close="handleClose"
  >
    <el-form :model="form" :rules="formRules" label-width="100px" ref="formRef">
      <el-form-item label="任务名称" prop="name">
        <el-input v-model="form.name" placeholder="输入任务名称" />
      </el-form-item>
      <el-form-item label="源仓库" prop="source_repo_key">
        <el-input v-model="form.source_repo_key" placeholder="源仓库标识" />
      </el-form-item>
      <el-form-item label="源分支" prop="source_branch">
        <el-input v-model="form.source_branch" placeholder="如: main, develop" />
      </el-form-item>
      <el-form-item label="目标仓库" prop="target_repo_key">
        <el-input v-model="form.target_repo_key" placeholder="目标仓库标识" />
      </el-form-item>
      <el-form-item label="目标分支" prop="target_branch">
        <el-input v-model="form.target_branch" placeholder="如: master, main" />
      </el-form-item>
      <el-form-item label="同步模式" prop="sync_mode">
        <el-radio-group v-model="form.sync_mode">
          <el-radio value="single">单向同步</el-radio>
          <el-radio value="bidirectional">双向同步</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="定时规则" prop="cron">
        <el-input v-model="form.cron" placeholder="Cron 表达式，如: 0 * * * *" />
        <div class="form-tip">留空则仅支持手动触发</div>
      </el-form-item>
      <el-form-item label="同步选项">
        <el-checkbox v-model="form.git_tags">同步标签</el-checkbox>
        <el-checkbox v-model="form.git_force">强制推送</el-checkbox>
        <el-checkbox v-model="form.git_prune">清理远程已删除分支</el-checkbox>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleConfirm" :loading="submitting">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { SyncTask } from '@/types/sync_v2'

// 同步任务新建/编辑弹窗：表单字段与校验在组件内，提交交回页面处理（form 由页面传入并回写）。
defineProps<{
  visible: boolean
  isEdit: boolean
  form: Partial<SyncTask>
  submitting: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  submit: []
}>()

const formRef = ref()

const formRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  source_repo_key: [{ required: true, message: '请输入源仓库标识', trigger: 'blur' }],
  source_branch: [{ required: true, message: '请输入源分支', trigger: 'blur' }],
  target_repo_key: [{ required: true, message: '请输入目标仓库标识', trigger: 'blur' }],
  target_branch: [{ required: true, message: '请输入目标分支', trigger: 'blur' }],
}

function handleClose() {
  formRef.value?.resetFields()
}

async function handleConfirm() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  emit('submit')
}
</script>

<style scoped lang="less">
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
