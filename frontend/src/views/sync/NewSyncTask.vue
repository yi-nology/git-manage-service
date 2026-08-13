<template>
  <div class="new-sync-task">
    <PageHeader title="创建同步任务">
      <template #actions>
        <el-button @click="$router.back()" :icon="Back">取消</el-button>
      </template>
    </PageHeader>

    <el-steps :active="currentStep" class="steps-bar" finish-status="success">
      <el-step title="基本配置" />
      <el-step title="同步规则" />
      <el-step title="高级选项" />
    </el-steps>

    <el-card class="form-card">
      <div v-if="currentStep === 0" class="step-content">
        <el-form :model="form" label-width="120px" label-position="left">
          <el-form-item label="任务名称" required>
            <el-input v-model="form.name" placeholder="输入任务名称" />
          </el-form-item>

          <el-form-item label="选择仓库" required>
            <el-select v-model="form.repo_key" placeholder="选择仓库" style="width: 100%">
              <el-option v-for="repo in repos" :key="repo.key" :label="repo.name" :value="repo.key" />
            </el-select>
          </el-form-item>

          <el-form-item label="同步模式" required>
            <el-radio-group v-model="form.sync_mode">
              <el-radio value="single">单分支同步</el-radio>
              <el-radio value="all-branch">全分支同步</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="任务描述">
            <el-input v-model="form.description" type="textarea" :rows="3" placeholder="输入任务描述（可选）" />
          </el-form-item>

          <el-form-item label="启用任务">
            <el-switch v-model="form.enabled" />
          </el-form-item>
        </el-form>
      </div>

      <div v-if="currentStep === 1" class="step-content">
        <el-form :model="form" label-width="120px" label-position="left">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="源远端" required>
                <el-select v-model="form.sourceRemote" placeholder="选择源远端" style="width: 100%">
                  <el-option label="Local (本地)" value="local" />
                  <el-option v-for="r in remotes" :key="r" :label="r" :value="r" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="目标远端" required>
                <el-select v-model="form.targetRemote" placeholder="选择目标远端" style="width: 100%">
                  <el-option v-for="r in remotes" :key="r" :label="r" :value="r" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20" v-if="form.sync_mode === 'single'">
            <el-col :span="12">
              <el-form-item label="源分支" required>
                <el-select v-model="form.source_branch" filterable placeholder="选择源分支" style="width: 100%">
                  <el-option v-for="b in branches" :key="b" :label="b" :value="b" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="目标分支" required>
                <el-select v-model="form.target_branch" filterable placeholder="选择目标分支" style="width: 100%">
                  <el-option v-for="b in branches" :key="b" :label="b" :value="b" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="触发方式">
            <el-radio-group v-model="form.triggerMode">
              <el-radio value="manual">手动触发</el-radio>
              <el-radio value="cron">定时调度</el-radio>
              <el-radio value="webhook">Webhook 触发</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="form.triggerMode === 'cron'" label="Cron 表达式" required>
            <el-input v-model="form.cron" placeholder="例如: 0 * * * *" />
            <div class="cron-hint">
              <span class="hint-item">每分钟: * * * * *</span>
              <span class="hint-item">每小时: 0 * * * *</span>
              <span class="hint-item">每天: 0 0 * * *</span>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <div v-if="currentStep === 2" class="step-content">
        <el-form :model="form" label-width="160px" label-position="left">
          <el-form-item label="同步所有标签">
            <el-switch v-model="form.git_tags" />
            <div class="form-tip">同步时包含所有 Git 标签</div>
          </el-form-item>

          <el-form-item label="强制推送">
            <el-switch v-model="form.git_force" />
            <div class="form-tip">使用 --force 参数强制覆盖远端分支</div>
          </el-form-item>

          <el-form-item label="清理分支">
            <el-switch v-model="form.git_prune" />
            <div class="form-tip">同步后删除本地不存在的远端分支</div>
          </el-form-item>

          <el-form-item label="跳过验证">
            <el-switch v-model="form.git_no_verify" />
            <div class="form-tip">跳过 pre-push 和 pre-receive 钩子验证</div>
          </el-form-item>

          <el-form-item label="超时时间（秒）">
            <el-input-number v-model="form.timeout" :min="30" :max="3600" style="width: 200px" />
          </el-form-item>

          <el-form-item label="重试次数">
            <el-input-number v-model="form.retry_count" :min="0" :max="5" style="width: 200px" />
          </el-form-item>

          <el-form-item label="通知设置">
            <el-checkbox-group v-model="form.notifications">
              <el-checkbox label="sync_success">同步成功通知</el-checkbox>
              <el-checkbox label="sync_failed">同步失败通知</el-checkbox>
              <el-checkbox label="sync_timeout">同步超时通知</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </el-form>
      </div>
    </el-card>

    <div class="form-actions">
      <el-button v-if="currentStep > 0" @click="prevStep" :icon="ArrowLeft">上一步</el-button>
      <el-button v-if="currentStep < 2" type="primary" @click="nextStep" :icon="ArrowRight">下一步</el-button>
      <el-button v-if="currentStep === 2" type="success" @click="submitForm" :icon="Check">创建任务</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Back, ArrowLeft, ArrowRight, Check } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'

const router = useRouter()
const currentStep = ref(0)

const repos = ref([
  { key: 'git-manage-service', name: 'git-manage-service' },
  { key: 'frontend', name: 'frontend' },
  { key: 'backend', name: 'backend' },
])

const remotes = ref(['origin', 'mirror', 'backup', 'upstream'])
const branches = ref(['main', 'master', 'develop', 'feature/*'])

const form = reactive({
  name: '',
  repo_key: '',
  sync_mode: 'single',
  description: '',
  enabled: true,
  sourceRemote: 'origin',
  targetRemote: '',
  source_branch: '',
  target_branch: '',
  triggerMode: 'manual',
  cron: '',
  git_tags: false,
  git_force: false,
  git_prune: false,
  git_no_verify: false,
  timeout: 300,
  retry_count: 2,
  notifications: ['sync_failed', 'sync_timeout'],
})

function nextStep() {
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

function submitForm() {
  console.log('Submit form:', form)
  router.push('/sync')
}
</script>

<style scoped lang="scss">
.new-sync-task {
  .steps-bar {
    max-width: 600px;
    margin: 0 auto 32px;
  }

  .form-card {
    max-width: 800px;
    margin: 0 auto;

    .step-content {
      padding: 20px 0;
    }

    .cron-hint {
      display: flex;
      gap: 16px;
      margin-top: 8px;

      .hint-item {
        font-size: 12px;
        color: var(--text-color-placeholder);
        padding: 2px 8px;
        background: var(--bg-color-page);
        border-radius: 4px;
        cursor: pointer;

        &:hover {
          color: var(--el-color-primary);
        }
      }
    }

    .form-tip {
      font-size: 12px;
      color: var(--text-color-placeholder);
      margin-top: 4px;
    }
  }

  .form-actions {
    display: flex;
    justify-content: center;
    gap: 12px;
    margin-top: 24px;
  }
}
</style>
