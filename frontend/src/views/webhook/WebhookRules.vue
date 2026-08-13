<template>
  <div class="webhook-rules">
    <PageHeader title="Webhook 规则管理">
      <template #actions>
        <el-button type="primary" @click="showCreateDialog = true" :icon="Plus">新建规则</el-button>
      </template>
    </PageHeader>

    <el-card class="filter-card">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input v-model="filters.keyword" placeholder="搜索规则名称" clearable :prefix-icon="Search" />
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.provider" placeholder="Git 平台" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="GitHub" value="github" />
            <el-option label="GitLab" value="gitlab" />
            <el-option label="Gitee" value="gitee" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: 100%">
            <el-option label="全部" value="" />
            <el-option label="已启用" value="enabled" />
            <el-option label="已禁用" value="disabled" />
          </el-select>
        </el-col>
        <el-col :span="10" style="text-align: right">
          <el-button @click="resetFilters" :icon="Refresh">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="rules" v-loading="loading" class="rules-table">
      <el-table-column prop="name" label="规则名称" min-width="180">
        <template #default="{ row }">
          <div class="name-cell">
            <el-tag v-if="row.enabled" type="success" size="small" effect="plain">启用</el-tag>
            <el-tag v-else size="small" effect="plain">禁用</el-tag>
            <span class="name">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="provider" label="Git 平台" width="120">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.provider }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="eventTypes" label="触发事件" min-width="200">
        <template #default="{ row }">
          <div class="event-tags">
            <el-tag v-for="event in row.eventTypes" :key="event" size="small" effect="plain">{{ event }}</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="repoFilter" label="仓库过滤" width="200">
        <template #default="{ row }">
          <span v-if="row.repoFilter" class="filter-text">{{ row.repoFilter }}</span>
          <span v-else class="text-muted">全部仓库</span>
        </template>
      </el-table-column>
      <el-table-column prop="triggerCount" label="触发次数" width="100" align="center" />
      <el-table-column prop="lastTrigger" label="上次触发" width="180">
        <template #default="{ row }">
          <span v-if="row.lastTrigger">{{ row.lastTrigger }}</span>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" :icon="View" @click="viewDetail(row)">详情</el-button>
          <el-button size="small" :icon="Edit" @click="editRule(row)">编辑</el-button>
          <el-button size="small" type="danger" :icon="Delete" @click="delete_rule(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        :page-sizes="[10, 20, 50, 100]"
      />
    </div>

    <el-dialog v-model="showCreateDialog" :title="editingRule ? '编辑规则' : '新建规则'" width="700px" destroy-on-close>
      <el-form :model="ruleForm" label-width="120px">
        <el-form-item label="规则名称" required>
          <el-input v-model="ruleForm.name" placeholder="输入规则名称" />
        </el-form-item>

        <el-form-item label="Git 平台" required>
          <el-select v-model="ruleForm.provider" placeholder="选择 Git 平台" style="width: 100%">
            <el-option label="GitHub" value="github" />
            <el-option label="GitLab" value="gitlab" />
            <el-option label="Gitee" value="gitee" />
          </el-select>
        </el-form-item>

        <el-form-item label="触发事件" required>
          <el-checkbox-group v-model="ruleForm.eventTypes">
            <el-checkbox label="push">Push 事件</el-checkbox>
            <el-checkbox label="create">分支创建</el-checkbox>
            <el-checkbox label="delete">分支删除</el-checkbox>
            <el-checkbox label="pull_request">PR/MR 事件</el-checkbox>
            <el-checkbox label="tag">Tag 事件</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="仓库过滤">
          <el-input v-model="ruleForm.repoFilter" placeholder="仓库名称过滤，支持通配符" />
          <div class="form-tip">留空表示匹配所有仓库，支持通配符: * 匹配任意字符</div>
        </el-form-item>

        <el-form-item label="分支过滤">
          <el-input v-model="ruleForm.branch_filter" placeholder="分支名称过滤，支持通配符" />
        </el-form-item>

        <el-form-item label="触发动作">
          <el-select v-model="ruleForm.action" placeholder="选择触发动作" style="width: 100%">
            <el-option label="触发同步任务" value="sync" />
            <el-option label="发送通知" value="notify" />
            <el-option label="自定义脚本" value="script" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="ruleForm.action === 'sync'" label="关联同步任务">
          <el-select v-model="ruleForm.syncTaskId" placeholder="选择同步任务" style="width: 100%">
            <el-option label="主分支同步" value="1" />
            <el-option label="全分支镜像同步" value="2" />
          </el-select>
        </el-form-item>

        <el-form-item label="启用规则">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Plus, Search, Refresh, View, Edit, Delete } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'

const loading = ref(false)
const showCreateDialog = ref(false)
const editingRule = ref<any>(null)

const rules = ref([
  {
    id: 1,
    name: 'GitHub Push 同步',
    provider: 'github',
    eventTypes: ['push'],
    repoFilter: '',
    branch_filter: 'main',
    action: 'sync',
    syncTaskId: '1',
    enabled: true,
    triggerCount: 156,
    lastTrigger: '2024-05-16 14:30:00',
  },
  {
    id: 2,
    name: 'GitLab PR 通知',
    provider: 'gitlab',
    eventTypes: ['pull_request'],
    repoFilter: 'frontend/*',
    branch_filter: '',
    action: 'notify',
    enabled: true,
    triggerCount: 42,
    lastTrigger: '2024-05-15 10:20:00',
  },
  {
    id: 3,
    name: 'Tag 发布同步',
    provider: 'github',
    eventTypes: ['tag'],
    repoFilter: 'backend/*',
    branch_filter: '',
    action: 'sync',
    syncTaskId: '2',
    enabled: false,
    triggerCount: 8,
    lastTrigger: '2024-05-10 18:00:00',
  },
])

const filters = reactive({
  keyword: '',
  provider: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 3,
})

const ruleForm = reactive({
  name: '',
  provider: '',
  eventTypes: [] as string[],
  repoFilter: '',
  branch_filter: '',
  action: 'sync',
  syncTaskId: '',
  enabled: true,
})

function resetFilters() {
  filters.keyword = ''
  filters.provider = ''
  filters.status = ''
}

function viewDetail(row: any) {
  console.log('View detail:', row)
}

function editRule(row: any) {
  editingRule.value = row
  Object.assign(ruleForm, row)
  showCreateDialog.value = true
}

function delete_rule(row: any) {
  console.log('Delete rule:', row)
}

function saveRule() {
  console.log('Save rule:', ruleForm)
  showCreateDialog.value = false
}
</script>

<style scoped lang="scss">
.webhook-rules {
  .filter-card {
    margin-bottom: 16px;
  }

  .rules-table {
    .name-cell {
      display: flex;
      align-items: center;
      gap: 8px;

      .name {
        font-weight: 500;
      }
    }

    .event-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }

    .filter-text {
      font-family: 'SF Mono', Monaco, monospace;
      font-size: 12px;
    }

    .text-muted {
      color: var(--text-color-placeholder);
    }
  }

  .form-tip {
    font-size: 12px;
    color: var(--text-color-placeholder);
    margin-top: 4px;
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
