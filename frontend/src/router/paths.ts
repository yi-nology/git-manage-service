/**
 * 路由路径常量，与 router/index.ts 的 routes 定义一一对应（键 = 路由 name）。
 *
 * 代码里的跳转一律引用此处，不要手写路径字符串：新增或改名路由时只需同步
 * router/index.ts 与本文件，类型系统会把其余引用一并揪出来。
 * 参数化路由用构建函数，参数顺序与路径段一一对应。
 */
export const ROUTES = {
  Home: '/',
  SyncTask: '/sync',
  SyncHistory: '/sync/history',
  NewSyncTask: '/sync/new',
  SyncConfig: '/sync/config',
  WebhookRules: '/webhook/rules',
  WebhookEventsLog: '/webhook/events',
  RepoList: '/local-repos',
  RepoRegister: '/local-repos/register',
  RepoClone: '/local-repos/clone',
  AuditLog: '/audit',
  RemoteRepos: '/remote-repos',
  Settings: '/settings',
  SSHKeys: '/settings/ssh-keys',
  Credentials: '/settings/credentials',
  AddCredential: '/settings/credentials/add',
  PlatformConfig: '/settings/platforms',
  NotificationChannels: '/settings/notification-channels',
  AddChannel: '/settings/notification-channels/add',
  LLMSettings: '/settings/llm',
  CodeReviewSettings: '/settings/code-review',
  BranchRuleSettings: '/settings/branch-rules',
  AuthorSettings: '/settings/author',
  SpecSettings: '/settings/spec',
  MCP: '/mcp',

  // ---- 参数化路由 ----
  RepoDetail: (repoKey: string) => `/local-repos/${repoKey}`,
  EditRepo: (repoKey: string) => `/local-repos/${repoKey}/edit`,
  BranchList: (repoKey: string) => `/local-repos/${repoKey}/branches`,
  BranchDetail: (repoKey: string, branch: string) => `/local-repos/${repoKey}/branches/${branch}`,
  BranchActions: (repoKey: string) => `/local-repos/${repoKey}/branch-actions`,
  BranchCompare: (repoKey: string) => `/local-repos/${repoKey}/compare`,
  RepoPatches: (repoKey: string) => `/local-repos/${repoKey}/patches`,
  RepoMirrors: (repoKey: string) => `/local-repos/${repoKey}/mirrors`,
  CRManagement: (repoKey: string) => `/local-repos/${repoKey}/cr`,
  ReviewDashboard: (repoKey: string) => `/local-repos/${repoKey}/review`,
  ReviewTaskList: (repoKey: string) => `/local-repos/${repoKey}/review/tasks`,
  ReviewTaskDetail: (repoKey: string, taskId: string | number) => `/local-repos/${repoKey}/review/tasks/${taskId}`,
  ReviewConfig: (repoKey: string) => `/local-repos/${repoKey}/review/config`,
  RemoteRepoDetail: (providerId: string | number, owner: string, name: string) =>
    `/remote-repos/${providerId}/${owner}/${name}`,
  EditCredential: (id: string | number) => `/settings/credentials/${id}/edit`,
  EditChannel: (id: string | number) => `/settings/notification-channels/${id}/edit`,
} as const
