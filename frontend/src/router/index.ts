import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/home/HomePage.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'local-repos',
        name: 'RepoList',
        component: () => import('@/views/repo/RepoListPage.vue'),
        meta: { title: '本地仓库' },
      },
      {
        path: 'local-repos/register',
        name: 'RepoRegister',
        component: () => import('@/views/repo/RepoRegisterPage.vue'),
        meta: { title: '注册仓库' },
      },
      {
        path: 'local-repos/clone',
        name: 'RepoClone',
        component: () => import('@/views/repo/RepoClonePage.vue'),
        meta: { title: '克隆仓库' },
      },
      {
        path: 'local-repos/:repoKey',
        name: 'RepoDetail',
        component: () => import('@/views/repo/RepoDetailPage.vue'),
        meta: { title: '仓库详情' },
      },
      {
        path: 'local-repos/:repoKey/edit',
        name: 'EditRepo',
        component: () => import('@/views/repo/EditRepoPage.vue'),
        meta: { title: '编辑仓库' },
      },
      {
        path: 'local-repos/:repoKey/branches',
        name: 'BranchList',
        component: () => import('@/views/branch/BranchListPage.vue'),
        meta: { title: '分支管理' },
      },
      {
        path: 'local-repos/:repoKey/branch-actions',
        name: 'BranchActions',
        component: () => import('@/views/branch/BranchActionsPage.vue'),
        meta: { title: '分支操作' },
      },
      {
        path: 'local-repos/:repoKey/branches/:branchName',
        name: 'BranchDetail',
        component: () => import('@/views/branch/BranchDetailPage.vue'),
        meta: { title: '分支详情' },
      },
      {
        path: 'local-repos/:repoKey/compare',
        name: 'BranchCompare',
        component: () => import('@/views/branch/BranchComparePage.vue'),
        meta: { title: '分支对比' },
      },
      {
        path: 'local-repos/:repoKey/patches',
        name: 'RepoPatches',
        component: () => import('@/views/repo/PatchPage.vue'),
        meta: { title: 'Patch 管理' },
      },
      {
        path: 'local-repos/:repoKey/mirrors',
        name: 'RepoMirrors',
        component: () => import('@/views/mirror/MirrorPage.vue'),
        meta: { title: '镜像同步' },
      },

      {
        path: 'local-repos/:repoKey/cr',
        name: 'CRManagement',
        component: () => import('@/views/cr/CRManagementPage.vue'),
        meta: { title: 'CR 管理' },
      },
      {
        path: 'local-repos/:repoKey/review',
        name: 'ReviewDashboard',
        component: () => import('@/views/review/ReviewDashboardPage.vue'),
        meta: { title: '代码审查' },
      },
      {
        path: 'local-repos/:repoKey/review/tasks',
        name: 'ReviewTaskList',
        component: () => import('@/views/review/ReviewTaskListPage.vue'),
        meta: { title: '审查任务列表' },
      },
      {
        path: 'local-repos/:repoKey/review/tasks/:taskId',
        name: 'ReviewTaskDetail',
        component: () => import('@/views/review/ReviewTaskDetailPage.vue'),
        meta: { title: '审查任务详情' },
      },
      {
        path: 'local-repos/:repoKey/review/config',
        name: 'ReviewConfig',
        component: () => import('@/views/review/ReviewConfigPage.vue'),
        meta: { title: '审查配置' },
      },
      {
        path: 'local-repos/:repoKey/webhooks',
        name: 'WebhookEvents',
        component: () => import('@/views/webhook/WebhookEventsPage.vue'),
        meta: { title: 'Webhook 事件' },
      },
      {
        path: 'audit',
        name: 'AuditLog',
        component: () => import('@/views/audit/AuditLogPage.vue'),
        meta: { title: '审计日志' },
      },
      {
        path: 'remote-repos',
        name: 'RemoteRepos',
        component: () => import('@/views/remote/RemoteReposPage.vue'),
        meta: { title: '远端仓库管理' },
      },
      {
        path: 'remote-repos/:providerId/:repoOwner/:repoName',
        name: 'RemoteRepoDetail',
        component: () => import('@/views/remote/RemoteRepoDetailPage.vue'),
        meta: { title: '远端仓库详情' },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/SettingsPage.vue'),
        meta: { title: '系统设置' },
      },
      {
        path: 'settings/ssh-keys',
        name: 'SSHKeys',
        component: () => import('@/views/settings/SSHKeysPage.vue'),
        meta: { title: 'SSH 密钥管理' },
      },
      {
        path: 'settings/credentials',
        name: 'Credentials',
        component: () => import('@/views/settings/CredentialPage.vue'),
        meta: { title: '凭证管理' },
      },
      {
        path: 'settings/credentials/add',
        name: 'AddCredential',
        component: () => import('@/views/settings/AddCredentialPage.vue'),
        meta: { title: '添加凭证' },
      },
      {
        path: 'settings/platforms',
        name: 'PlatformConfig',
        component: () => import('@/views/settings/PlatformConfigPage.vue'),
        meta: { title: '平台配置' },
      },
      {
        path: 'settings/credentials/:id/edit',
        name: 'EditCredential',
        component: () => import('@/views/settings/AddCredentialPage.vue'),
        meta: { title: '编辑凭证' },
      },
      {
        path: 'settings/notification-channels',
        name: 'NotificationChannels',
        component: () => import('@/views/settings/NotificationChannelsPage.vue'),
        meta: { title: '通知渠道管理' },
      },
      {
        path: 'settings/notification-channels/add',
        name: 'AddChannel',
        component: () => import('@/views/settings/AddChannelPage.vue'),
        meta: { title: '添加通知渠道' },
      },
      {
        path: 'settings/notification-channels/:id/edit',
        name: 'EditChannel',
        component: () => import('@/views/settings/AddChannelPage.vue'),
        meta: { title: '编辑通知渠道' },
      },
      {
        path: 'settings/llm',
        name: 'LLMSettings',
        component: () => import('@/views/settings/LLMSettingsPage.vue'),
        meta: { title: 'LLM 配置' },
      },
      {
        path: 'settings/code-review',
        name: 'CodeReviewSettings',
        component: () => import('@/views/settings/CodeReviewSettingsPage.vue'),
        meta: { title: '代码审查设置' },
      },
      {
        path: 'settings/branch-rules',
        name: 'BranchRuleSettings',
        component: () => import('@/views/settings/BranchRuleSettingsPage.vue'),
        meta: { title: '分支规则管理' },
      },
      {
        path: 'settings/author',
        name: 'AuthorSettings',
        component: () => import('@/views/settings/AuthorSettingsPage.vue'),
        meta: { title: 'Git 作者管理' },
      },
      {
        path: 'settings/spec',
        name: 'SpecSettings',
        component: () => import('@/views/settings/SpecSettingsPage.vue'),
        meta: { title: 'Spec 全局配置' },
      },
      {
        path: 'mcp',
        name: 'MCP',
        component: () => import('@/views/mcp/McpPage.vue'),
        meta: { title: 'MCP 配置' },
      },

    ],
  },
  // 404 页面
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/NotFoundPage.vue'),
    meta: { title: '页面未找到' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string
  document.title = title ? `${title} - Git Branch Manager` : 'Git Branch Manager'
  next()
})

export default router
