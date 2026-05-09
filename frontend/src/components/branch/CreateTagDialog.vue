<template>
  <el-dialog v-model="dialogVisible" title="打标签 (Tag)" width="550px" destroy-on-close>
    <el-form :model="tagForm" label-width="100px">
      <el-form-item label="目标引用">
        <el-input :model-value="tagForm.ref" disabled />
      </el-form-item>
      <el-form-item label="版本类型">
        <el-radio-group v-model="tagForm.versionType" @change="handleTagVersionTypeChange">
          <el-radio-button value="patch">Patch</el-radio-button>
          <el-radio-button value="minor">Minor</el-radio-button>
          <el-radio-button value="major">Major</el-radio-button>
          <el-radio-button value="custom">自定义</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="当前版本" v-if="tagNextVersion">
        <el-tag type="info" size="small">{{ tagNextVersion.current || '无' }}</el-tag>
      </el-form-item>
      <el-form-item label="标签名" required>
        <el-input v-model="tagForm.name" :disabled="tagForm.versionType !== 'custom'" placeholder="v1.0.0" />
      </el-form-item>
      <el-form-item label="说明">
        <el-input v-model="tagForm.message" type="textarea" :rows="2" />
      </el-form-item>
      <el-form-item label="推送到远端">
        <el-select v-model="tagForm.push_remote" placeholder="不推送" clearable>
          <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleCreateTag">创建</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createTag } from '@/api/modules/branch'
import { getNextVersion } from '@/api/modules/version'
import type { NextVersionInfo } from '@/api/modules/version'

const props = defineProps<{
  visible: boolean
  repoKey: string
  remoteNames: string[]
  branchName: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'created': []
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
})

const tagForm = ref({ ref: '', name: '', message: '', push_remote: '', versionType: 'patch' as 'patch' | 'minor' | 'major' | 'custom' })
const tagNextVersion = ref<NextVersionInfo | null>(null)

watch(() => props.visible, (v) => {
  if (v && props.branchName) {
    tagForm.value = { ref: props.branchName, name: '', message: '', push_remote: '', versionType: 'patch' }
    tagNextVersion.value = null
    getNextVersion(props.repoKey).then(info => {
      tagNextVersion.value = info
      handleTagVersionTypeChange('patch')
    }).catch(() => {})
  }
})

function handleTagVersionTypeChange(type: string | number | boolean | undefined) {
  if (!tagNextVersion.value) return
  switch (type) {
    case 'patch':
      tagForm.value.name = tagNextVersion.value.next_patch
      break
    case 'minor':
      tagForm.value.name = tagNextVersion.value.next_minor
      break
    case 'major':
      tagForm.value.name = tagNextVersion.value.next_major
      break
    case 'custom':
      tagForm.value.name = ''
      break
  }
}

async function handleCreateTag() {
  if (!tagForm.value.name) {
    ElMessage.warning('请输入标签名')
    return
  }
  try {
    await createTag({
      repo_key: props.repoKey,
      name: tagForm.value.name,
      ref: tagForm.value.ref,
      message: tagForm.value.message || undefined,
      push_remote: tagForm.value.push_remote || undefined,
    })
    ElMessage.success('标签创建成功')
    dialogVisible.value = false
    emit('created')
  } catch {}
}
</script>
