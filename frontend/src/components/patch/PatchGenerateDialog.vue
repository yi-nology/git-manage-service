<template>
  <el-dialog v-model="visible" title="生成 Patch" width="700px" destroy-on-close @open="loadDialogData">
    <el-form :model="generateForm" label-width="120px">
      <el-form-item label="生成方式">
        <el-radio-group v-model="generateMode">
          <el-radio value="range">分支/Tag/Commit 范围</el-radio>
          <el-radio value="commits">选择 Commits</el-radio>
        </el-radio-group>
      </el-form-item>

      <template v-if="generateMode === 'range'">
        <el-form-item label="基准（起点）">
          <el-select
            v-model="generateForm.base"
            filterable
            allow-create
            placeholder="选择或输入基准（分支/Tag/Commit）"
            style="width: 100%"
          >
            <el-option-group label="分支">
              <el-option
                v-for="branch in branches"
                :key="branch.name"
                :label="branch.name"
                :value="branch.name"
              />
            </el-option-group>
            <el-option-group label="Tags">
              <el-option
                v-for="tag in tags"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-option-group>
            <el-option-group label="最近 Commits">
              <el-option
                v-for="commit in recentCommits"
                :key="commit.hash"
                :label="`${commit.shortHash} - ${commit.message.slice(0, 50)}`"
                :value="commit.hash"
              />
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="目标（终点）">
          <el-select
            v-model="generateForm.target"
            filterable
            allow-create
            placeholder="选择或输入目标（分支/Tag/Commit）"
            style="width: 100%"
          >
            <el-option-group label="分支">
              <el-option
                v-for="branch in branches"
                :key="branch.name"
                :label="branch.name"
                :value="branch.name"
              />
            </el-option-group>
            <el-option-group label="Tags">
              <el-option
                v-for="tag in tags"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-option-group>
            <el-option-group label="最近 Commits">
              <el-option
                v-for="commit in recentCommits"
                :key="commit.hash"
                :label="`${commit.shortHash} - ${commit.message.slice(0, 50)}`"
                :value="commit.hash"
              />
            </el-option-group>
          </el-select>
        </el-form-item>
      </template>

      <template v-else>
        <el-form-item label="选择 Commits">
          <el-select
            v-model="selectedCommits"
            multiple
            filterable
            placeholder="选择要生成 patch 的 commits（可多选）"
            style="width: 100%"
          >
            <el-option
              v-for="commit in recentCommits"
              :key="commit.hash"
              :label="`${commit.shortHash} - ${commit.message.slice(0, 60)} (${commit.authorName})`"
              :value="commit.hash"
            />
          </el-select>
          <div class="hint">提示：可多选，按 Ctrl/Cmd 点击选择多个</div>
        </el-form-item>
      </template>

      <el-divider />

      <el-form-item label="保存选项">
        <el-checkbox v-model="savePatch">保存到项目</el-checkbox>
      </el-form-item>

      <template v-if="savePatch">
        <el-form-item label="文件名" required>
          <el-input v-model="patchName" placeholder="输入描述，如: feature-login、fix-bug">
            <template #prepend>{{ getNextPatchPrefix() }}</template>
            <template #append>.patch</template>
          </el-input>
          <div class="hint">系统自动生成序号，你只需输入描述部分</div>
        </el-form-item>
        <el-form-item label="保存路径">
          <el-select
            v-model="selectedPath"
            filterable
            allow-create
            placeholder="选择或输入保存路径（默认: patches/）"
            style="width: 100%"
          >
            <el-option label="patches/ (默认)" value="patches" />
            <el-option
              v-for="dir in repoDirs"
              :key="dir.path"
              :label="dir.path + '/'"
              :value="dir.path"
            />
          </el-select>
          <div class="path-hint">
            💡 建议使用默认的 <code>patches/</code> 目录
            <el-button size="small" link @click="selectedPath = 'patches'">重置为默认</el-button>
          </div>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="autoCommit">立即提交到 Git</el-checkbox>
        </el-form-item>
        <el-form-item v-if="autoCommit" label="提交消息">
          <el-input
            v-model="commitMessage"
            placeholder="如: chore: add feature-xxx patch"
          />
          <div class="hint">快捷选项：
            <el-button size="small" link @click="commitMessage = 'chore: add patch for ' + (patchName || 'feature')">chore: add patch</el-button>
            <el-button size="small" link @click="commitMessage = 'feat: add patch'">feat: add patch</el-button>
          </div>
        </el-form-item>
      </template>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleGenerate" :loading="generating">
        {{ savePatch ? '生成并保存' : '生成并下载' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  generatePatch,
  savePatch as savePatchApi,
} from '@/api/modules/patch'
import { getBranchList } from '@/api/modules/branch'
import { getTagList } from '@/api/modules/branch'
import { searchCommits } from '@/api/modules/commit'
import { getFileTree } from '@/api/modules/file'
import type { PatchInfoDTO } from '@/types/patch'
import type { BranchInfo } from '@/types/branch'
import type { CommitDetail } from '@/api/modules/commit'
import { useNotification } from '@/composables/useNotification'

const props = defineProps<{
  repoKey: string
  patches: PatchInfoDTO[]
}>()

const emit = defineEmits<{
  generated: []
}>()

const visible = defineModel<boolean>({ required: true })

const { showSuccess, showError } = useNotification()

const generateMode = ref<'range' | 'commits'>('range')
const generateForm = ref({
  base: '',
  target: '',
  commits: [] as string[],
})
const selectedCommits = ref<string[]>([])
const savePatch = ref(true)
const patchName = ref('')
const selectedPath = ref('')
const autoCommit = ref(false)
const commitMessage = ref('')
const generating = ref(false)

const branches = ref<BranchInfo[]>([])
const tags = ref<string[]>([])
const recentCommits = ref<CommitDetail[]>([])
const repoDirs = ref<any[]>([])

async function loadDialogData() {
  try {
    const res = await getBranchList(props.repoKey, { page_size: 100 })
    branches.value = res.list || []
  } catch (e) {
    console.error('Failed to load branches:', e)
  }

  try {
    tags.value = await getTagList(props.repoKey)
  } catch (e) {
    console.error('Failed to load tags:', e)
  }

  try {
    const res = await searchCommits(props.repoKey, { page_size: 50 })
    recentCommits.value = res.commits
  } catch (e) {
    console.error('Failed to load commits:', e)
  }

  try {
    const res = await getFileTree(props.repoKey, { recursive: true })
    repoDirs.value = buildPathTree(res.entries)
    if (!selectedPath.value) {
      selectedPath.value = 'patches'
    }
  } catch (e) {
    console.error('Failed to load file tree:', e)
  }
}

function buildPathTree(entries: any[]): any[] {
  const dirs: any[] = []

  entries.forEach((e: any) => {
    if (e.type === 'dir') {
      const path = e.path
      if (!path.startsWith('.') &&
          !path.includes('node_modules') &&
          !path.includes('vendor') &&
          !path.includes('dist') &&
          !path.includes('build')) {
        dirs.push({
          path: path,
          name: path.split('/').pop() || path,
        })
      }
    }
  })

  dirs.sort((a, b) => a.path.localeCompare(b.path))

  return dirs
}

function getNextPatchPrefix(): string {
  if (!props.patches || props.patches.length === 0) {
    return '001-'
  }

  let maxSeq = 0
  props.patches.forEach(p => {
    if (p.sequence > maxSeq) {
      maxSeq = p.sequence
    }
  })

  const nextSeq = maxSeq + 1
  return String(nextSeq).padStart(3, '0') + '-'
}

function downloadContent(content: string, filename: string) {
  const blob = new Blob([content], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

async function handleGenerate() {
  if (generateMode.value === 'range') {
    if (!generateForm.value.base || !generateForm.value.target) {
      ElMessage.warning('请选择基准和目标')
      return
    }
  } else {
    if (selectedCommits.value.length === 0) {
      ElMessage.warning('请选择至少一个 Commit')
      return
    }
    generateForm.value.commits = selectedCommits.value
  }

  if (savePatch.value && !patchName.value.trim()) {
    ElMessage.warning('请填写文件名描述')
    return
  }

  if (savePatch.value && autoCommit.value && !commitMessage.value.trim()) {
    ElMessage.warning('请填写提交消息')
    return
  }

  generating.value = true
  try {
    const req: any = { repo_key: props.repoKey }
    if (generateMode.value === 'range') {
      req.base = generateForm.value.base
      req.target = generateForm.value.target
    } else {
      req.commits = generateForm.value.commits
    }

    const result = await generatePatch(req)
    const content = result.content

    if (savePatch.value) {
      const prefix = getNextPatchPrefix()
      const fullName = prefix + (patchName.value.endsWith('.patch') ? patchName.value : patchName.value + '.patch')

      await savePatchApi({
        repo_key: props.repoKey,
        patch_name: fullName,
        patch_content: content,
        custom_path: selectedPath.value || undefined,
        commit_message: autoCommit.value ? commitMessage.value : undefined,
      })
      showSuccess('Patch 已保存' + (autoCommit.value ? '并提交到 Git' : ''))
    } else {
      downloadContent(content, patchName.value || 'patch.patch')
    }

    visible.value = false
    emit('generated')
  } catch (e: any) {
    showError('生成失败', e)
  } finally {
    generating.value = false
  }
}
</script>

<style scoped>
.path-hint {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-xs);
}

.path-hint code {
  padding: 2px 6px;
  background: var(--border-color-light);
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', monospace;
}

.hint {
  font-size: var(--font-size-xs);
  color: var(--text-color-secondary);
  margin-top: var(--spacing-xs);
}
</style>
