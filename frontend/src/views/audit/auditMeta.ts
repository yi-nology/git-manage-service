// AuditLogPage 的展示元数据与纯格式化函数：无状态、无副作用，
// 独立成模块让页面 SFC 只保留真正的查询逻辑。

export interface ActionGroup {
  label: string
  items: { value: string; label: string }[]
}

export const actionLabelMap: Record<string, string> = {
  CREATE: '创建仓库',
  UPDATE: '更新仓库',
  DELETE: '删除仓库',
  FETCH_REPO: '拉取仓库',
  CLONE_REPO: '克隆仓库',
  SCAN_REPO: '扫描仓库',
  SCAN_DIRECTORY: '扫描目录',
  BATCH_CREATE_REPO: '批量创建仓库',

  CREATE_BRANCH: '创建分支',
  DELETE_BRANCH: '删除分支',
  UPDATE_BRANCH: '更新分支',
  CHECKOUT_BRANCH: '切换分支',
  PUSH_BRANCH: '推送分支',
  PULL_BRANCH: '拉取分支',
  MERGE: '合并分支',
  CHERRY_PICK: 'Cherry-Pick',
  REBASE: '变基',
  REBASE_ABORT: '中止变基',
  REBASE_CONTINUE: '继续变基',

  WORKSPACE_STAGE: '暂存文件',
  WORKSPACE_UNSTAGE: '取消暂存',
  WORKSPACE_COMMIT: '提交变更',
  WORKSPACE_PULL: '拉取代码',
  WORKSPACE_PUSH: '推送代码',
  WORKSPACE_UNTRACK: '移除跟踪',
  WORKSPACE_GITIGNORE: '添加忽略',
  WORKSPACE_RESOLVE: '解决冲突',
  WORKSPACE_AI_RESOLVE: 'AI 解决冲突',
  WORKSPACE_AI_COMMIT_MSG: 'AI 生成提交信息',

  TAG_CREATE: '创建标签',
  TAG_DELETE: '删除标签',
  TAG_PUSH: '推送标签',

  STASH_SAVE: '保存暂存',
  STASH_APPLY: '应用暂存',
  STASH_POP: '弹出暂存',
  STASH_DROP: '丢弃暂存',
  STASH_CLEAR: '清空暂存',

  SUBMODULE_ADD: '添加子模块',
  SUBMODULE_INIT: '初始化子模块',
  SUBMODULE_UPDATE: '更新子模块',
  SUBMODULE_SYNC: '同步子模块',
  SUBMODULE_REMOVE: '删除子模块',

  SYNC_CREATE: '创建同步任务',
  SYNC_UPDATE: '更新同步任务',
  SYNC_DELETE: '删除同步任务',
  SYNC_EXECUTE: '执行同步',
  SYNC_RUN: '运行同步',
  SYNC_BATCH: '批量同步',
  SYNC_PREVIEW: '预览同步',
  SYNC_ADHOC: '手动同步',
  SYNC: '同步',

  WEBHOOK_TRIGGER: 'Webhook 触发',
  WEBHOOK_TRIGGER_BY_TOKEN: 'Token 触发',
  WEBHOOK_RECEIVED: '接收 Webhook',
  WEBHOOK_EVENT_RETRY: '重试 Webhook',

  SUBMIT_CHANGES: '提交变更',
  SYSTEM_CONFIG_UPDATE: '更新系统配置',
  SYSTEM_TEST_CONNECTION: '测试连接',
  SYSTEM_SELECT_DIR: '选择目录',

  CREDENTIAL_CREATE: '创建凭证',
  CREDENTIAL_UPDATE: '更新凭证',
  CREDENTIAL_DELETE: '删除凭证',
  CREDENTIAL_TEST: '测试凭证',

  SSHKEY_CREATE: '创建 SSH 密钥',
  SSHKEY_UPDATE: '更新 SSH 密钥',
  SSHKEY_DELETE: '删除 SSH 密钥',
  SSHKEY_TEST: '测试 SSH 密钥',

  LLM_PROVIDER_CREATE: '创建 LLM 提供商',
  LLM_PROVIDER_UPDATE: '更新 LLM 提供商',
  LLM_PROVIDER_DELETE: '删除 LLM 提供商',
  LLM_PROVIDER_SET_DEFAULT: '设为默认 LLM',
  LLM_PROVIDER_TEST: '测试 LLM 连接',
  CODE_REVIEW_SETTINGS_UPDATE: '更新代码审查设置',
  BRANCH_RULES_UPDATE: '更新分支规则',
  REMOTE_BRANCH_RULES_UPDATE: '更新远程分支规则',

  REVIEW_RULE_CREATE: '创建审查规则',
  REVIEW_RULE_UPDATE: '更新审查规则',
  REVIEW_RULE_DELETE: '删除审查规则',
  REVIEW_RULES_BATCH_UPDATE: '批量更新审查规则',

  PROVIDER_CREATE: '创建平台配置',
  PROVIDER_UPDATE: '更新平台配置',
  PROVIDER_DELETE: '删除平台配置',
  PROVIDER_TEST: '测试平台连接',
  PROVIDER_REMOTE_BRANCH_CREATE: '创建远程分支',
  PROVIDER_REMOTE_BRANCH_DELETE: '删除远程分支',

  BINDING_CREATE: '创建绑定',
  BINDING_UPDATE: '更新绑定',
  BINDING_DELETE: '删除绑定',
  BINDING_SET_PRIMARY: '设为主绑定',
  BINDING_REGISTER_WEBHOOK: '注册 Webhook',
  BINDING_DELETE_WEBHOOK: '删除 Webhook',

  CR_CREATE: '创建 CR',
  CR_MERGE: '合并 CR',
  CR_CLOSE: '关闭 CR',
  CR_CREATE_BY_PROVIDER: '创建远程 CR',
  CR_MERGE_BY_PROVIDER: '合并远程 CR',
  CR_CLOSE_BY_PROVIDER: '关闭远程 CR',
  CR_SYNC: '同步 CR',

  REVIEW_TASK_CREATE: '创建审查任务',
  REVIEW_TASK_RETRY: '重试审查任务',
  REVIEW_CONFIG_UPDATE: '更新审查配置',
  REMOTE_REVIEW_CONFIG_UPDATE: '更新远程审查配置',

  AUTHOR_IDENTITY_CREATE: '创建身份',
  AUTHOR_IDENTITY_UPDATE: '更新身份',
  AUTHOR_IDENTITY_DELETE: '删除身份',
  AUTHOR_IDENTITY_ACTIVATE: '激活身份',
  AUTHOR_REPO_CONFIG_SET: '设置仓库作者',
  AUTHOR_FIX: '修正作者',
  AUTHOR_FIX_ALL: '批量修正作者',

  NOTIFICATION_CHANNEL_CREATE: '创建通知渠道',
  NOTIFICATION_CHANNEL_UPDATE: '更新通知渠道',
  NOTIFICATION_CHANNEL_DELETE: '删除通知渠道',

  PATCH_SAVE: '保存补丁',
  PATCH_APPLY: '应用补丁',
  PATCH_DELETE: '删除补丁',

  REPO_SLIM: '仓库瘦身',
  REPO_GC: '仓库 GC',
  REPO_GITIGNORE: '更新 Gitignore',

  SAVE_SPEC: '保存 Spec',
  CREATE_SPEC: '创建 Spec',
  DELETE_SPEC: '删除 Spec',
  COMMIT_SPEC: '提交 Spec',
  SPEC_AI_ASSIST: 'AI 辅助 Spec',
  SPEC_AI_FIX: 'AI 修复 Spec',
  SPEC_CONFIG_SAVE: '保存 Spec 配置',
  SPEC_LINT_RULE_CREATE: '创建 Lint 规则',
  SPEC_LINT_RULE_UPDATE: '更新 Lint 规则',

  STATS_CONFIG_SAVE: '保存统计配置',
}

export const actionGroups: ActionGroup[] = [
  {
    label: '仓库管理',
    items: [
      { value: 'CREATE', label: '创建仓库' },
      { value: 'UPDATE', label: '更新仓库' },
      { value: 'DELETE', label: '删除仓库' },
      { value: 'FETCH_REPO', label: '拉取仓库' },
      { value: 'CLONE_REPO', label: '克隆仓库' },
      { value: 'SCAN_REPO', label: '扫描仓库' },
      { value: 'BATCH_CREATE_REPO', label: '批量创建仓库' },
      { value: 'REPO_SLIM', label: '仓库瘦身' },
      { value: 'REPO_GC', label: '仓库 GC' },
      { value: 'REPO_GITIGNORE', label: '更新 Gitignore' },
    ],
  },
  {
    label: '分支操作',
    items: [
      { value: 'CREATE_BRANCH', label: '创建分支' },
      { value: 'DELETE_BRANCH', label: '删除分支' },
      { value: 'UPDATE_BRANCH', label: '更新分支' },
      { value: 'CHECKOUT_BRANCH', label: '切换分支' },
      { value: 'PUSH_BRANCH', label: '推送分支' },
      { value: 'PULL_BRANCH', label: '拉取分支' },
      { value: 'MERGE', label: '合并分支' },
      { value: 'CHERRY_PICK', label: 'Cherry-Pick' },
      { value: 'REBASE', label: '变基' },
      { value: 'REBASE_ABORT', label: '中止变基' },
      { value: 'REBASE_CONTINUE', label: '继续变基' },
    ],
  },
  {
    label: '工作区',
    items: [
      { value: 'WORKSPACE_STAGE', label: '暂存文件' },
      { value: 'WORKSPACE_UNSTAGE', label: '取消暂存' },
      { value: 'WORKSPACE_COMMIT', label: '提交变更' },
      { value: 'WORKSPACE_PULL', label: '拉取代码' },
      { value: 'WORKSPACE_PUSH', label: '推送代码' },
      { value: 'WORKSPACE_UNTRACK', label: '移除跟踪' },
      { value: 'WORKSPACE_GITIGNORE', label: '添加忽略' },
      { value: 'WORKSPACE_RESOLVE', label: '解决冲突' },
      { value: 'WORKSPACE_AI_RESOLVE', label: 'AI 解决冲突' },
      { value: 'WORKSPACE_AI_COMMIT_MSG', label: 'AI 生成提交信息' },
    ],
  },
  {
    label: '标签',
    items: [
      { value: 'TAG_CREATE', label: '创建标签' },
      { value: 'TAG_DELETE', label: '删除标签' },
      { value: 'TAG_PUSH', label: '推送标签' },
    ],
  },
  {
    label: '凭证 / 密钥',
    items: [
      { value: 'CREDENTIAL_CREATE', label: '创建凭证' },
      { value: 'CREDENTIAL_UPDATE', label: '更新凭证' },
      { value: 'CREDENTIAL_DELETE', label: '删除凭证' },
      { value: 'CREDENTIAL_TEST', label: '测试凭证' },
      { value: 'SSHKEY_CREATE', label: '创建 SSH 密钥' },
      { value: 'SSHKEY_UPDATE', label: '更新 SSH 密钥' },
      { value: 'SSHKEY_DELETE', label: '删除 SSH 密钥' },
      { value: 'SSHKEY_TEST', label: '测试 SSH 密钥' },
    ],
  },
  {
    label: '同步',
    items: [
      { value: 'SYNC_CREATE', label: '创建同步任务' },
      { value: 'SYNC_UPDATE', label: '更新同步任务' },
      { value: 'SYNC_DELETE', label: '删除同步任务' },
      { value: 'SYNC_EXECUTE', label: '执行同步' },
      { value: 'SYNC_RUN', label: '运行同步' },
      { value: 'SYNC_BATCH', label: '批量同步' },
      { value: 'SYNC_PREVIEW', label: '预览同步' },
    ],
  },
  {
    label: 'Webhook',
    items: [
      { value: 'WEBHOOK_TRIGGER', label: 'Webhook 触发' },
      { value: 'WEBHOOK_TRIGGER_BY_TOKEN', label: 'Token 触发' },
      { value: 'WEBHOOK_RECEIVED', label: '接收 Webhook' },
      { value: 'WEBHOOK_EVENT_RETRY', label: '重试 Webhook' },
    ],
  },
  {
    label: '平台绑定',
    items: [
      { value: 'PROVIDER_CREATE', label: '创建平台配置' },
      { value: 'PROVIDER_UPDATE', label: '更新平台配置' },
      { value: 'PROVIDER_DELETE', label: '删除平台配置' },
      { value: 'BINDING_CREATE', label: '创建绑定' },
      { value: 'BINDING_UPDATE', label: '更新绑定' },
      { value: 'BINDING_DELETE', label: '删除绑定' },
      { value: 'BINDING_SET_PRIMARY', label: '设为主绑定' },
    ],
  },
  {
    label: '代码审查',
    items: [
      { value: 'CR_CREATE', label: '创建 CR' },
      { value: 'CR_MERGE', label: '合并 CR' },
      { value: 'CR_CLOSE', label: '关闭 CR' },
      { value: 'REVIEW_TASK_CREATE', label: '创建审查任务' },
      { value: 'REVIEW_TASK_RETRY', label: '重试审查任务' },
      { value: 'REVIEW_CONFIG_UPDATE', label: '更新审查配置' },
    ],
  },
  {
    label: '作者管理',
    items: [
      { value: 'AUTHOR_IDENTITY_CREATE', label: '创建身份' },
      { value: 'AUTHOR_IDENTITY_UPDATE', label: '更新身份' },
      { value: 'AUTHOR_IDENTITY_DELETE', label: '删除身份' },
      { value: 'AUTHOR_IDENTITY_ACTIVATE', label: '激活身份' },
      { value: 'AUTHOR_FIX', label: '修正作者' },
      { value: 'AUTHOR_FIX_ALL', label: '批量修正作者' },
    ],
  },
  {
    label: 'Spec',
    items: [
      { value: 'SAVE_SPEC', label: '保存 Spec' },
      { value: 'CREATE_SPEC', label: '创建 Spec' },
      { value: 'DELETE_SPEC', label: '删除 Spec' },
      { value: 'COMMIT_SPEC', label: '提交 Spec' },
      { value: 'SPEC_AI_ASSIST', label: 'AI 辅助 Spec' },
      { value: 'SPEC_AI_FIX', label: 'AI 修复 Spec' },
      { value: 'SPEC_CONFIG_SAVE', label: '保存 Spec 配置' },
      { value: 'SPEC_LINT_RULE_CREATE', label: '创建 Lint 规则' },
      { value: 'SPEC_LINT_RULE_UPDATE', label: '更新 Lint 规则' },
    ],
  },
  {
    label: '设置',
    items: [
      { value: 'LLM_PROVIDER_CREATE', label: '创建 LLM 提供商' },
      { value: 'LLM_PROVIDER_UPDATE', label: '更新 LLM 提供商' },
      { value: 'LLM_PROVIDER_DELETE', label: '删除 LLM 提供商' },
      { value: 'LLM_PROVIDER_SET_DEFAULT', label: '设为默认 LLM' },
      { value: 'CODE_REVIEW_SETTINGS_UPDATE', label: '更新代码审查设置' },
      { value: 'BRANCH_RULES_UPDATE', label: '更新分支规则' },
      { value: 'REVIEW_RULE_CREATE', label: '创建审查规则' },
      { value: 'REVIEW_RULE_UPDATE', label: '更新审查规则' },
      { value: 'REVIEW_RULE_DELETE', label: '删除审查规则' },
      { value: 'NOTIFICATION_CHANNEL_CREATE', label: '创建通知渠道' },
      { value: 'NOTIFICATION_CHANNEL_UPDATE', label: '更新通知渠道' },
      { value: 'NOTIFICATION_CHANNEL_DELETE', label: '删除通知渠道' },
      { value: 'SYSTEM_CONFIG_UPDATE', label: '更新系统配置' },
    ],
  },
  {
    label: '其他',
    items: [
      { value: 'SUBMIT_CHANGES', label: '提交变更' },
      { value: 'SUBMODULE_ADD', label: '添加子模块' },
      { value: 'SUBMODULE_INIT', label: '初始化子模块' },
      { value: 'SUBMODULE_UPDATE', label: '更新子模块' },
      { value: 'SUBMODULE_SYNC', label: '同步子模块' },
      { value: 'SUBMODULE_REMOVE', label: '删除子模块' },
      { value: 'STASH_SAVE', label: '保存暂存' },
      { value: 'STASH_APPLY', label: '应用暂存' },
      { value: 'STASH_POP', label: '弹出暂存' },
      { value: 'STASH_DROP', label: '丢弃暂存' },
      { value: 'STASH_CLEAR', label: '清空暂存' },
      { value: 'PATCH_SAVE', label: '保存补丁' },
      { value: 'PATCH_APPLY', label: '应用补丁' },
      { value: 'PATCH_DELETE', label: '删除补丁' },
      { value: 'STATS_CONFIG_SAVE', label: '保存统计配置' },
    ],
  },
]

export function getActionLabel(action: string): string {
  return actionLabelMap[action] || action
}

export const targetTypeMap: Record<string, string> = {
  repo: '仓库',
  task: '同步任务',
  task_key: '同步任务',
  channel: '通知渠道',
  credential: '凭证',
  sshkey: 'SSH 密钥',
  llm_provider: 'LLM 提供商',
  provider: '平台配置',
  binding: '绑定',
  review_rule: '审查规则',
  review_task: '审查任务',
  author: '作者身份',
  lint_rule: 'Lint 规则',
  webhook: 'Webhook',
  webhook_event: 'Webhook 事件',
  settings: '设置',
  system: '系统',
}

export function truncateDetails(details: string): string {
  if (!details || details === '-') return '-'
  try {
    const parsed = JSON.parse(details)
    const keys = Object.keys(parsed)
    const summary = keys.map(k => `${k}=${parsed[k]}`).join(', ')
    return summary.length > 60 ? summary.substring(0, 60) + '...' : summary
  } catch {
    return details.length > 60 ? details.substring(0, 60) + '...' : details
  }
}

export function formatDetailsJSON(details: string): string {
  if (!details || details === '-') return '-'
  try {
    const parsed = JSON.parse(details)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return details
  }
}

export function getActionClass(action: string): string {
  if (action.includes('DELETE') || action.includes('REMOVE') || action.includes('DROP') || action.includes('CLEAR') || action === 'MERGE_CONFLICT') return 'danger'
  if (action.includes('CREATE') || action === 'MERGE_SUCCESS' || action.includes('SUBMODULE_ADD') || action.includes('ACTIVATE')) return 'success'
  if (action.includes('UPDATE') || action.includes('PUSH') || action.includes('COMMIT') || action.includes('SAVE') || action.includes('SUBMIT')) return 'warning'
  return 'info'
}

export function getStatusClass(_action: string): string {
  return 'success'
}

