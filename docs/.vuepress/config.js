module.exports = {
  title: 'Git Manage Service',
  description: '轻量级多仓库自动化同步管理系统',
  base: '/git-manage-service/',
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/images/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#6366f1' }],
    ['meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }],
    ['meta', { name: 'apple-mobile-web-app-status-bar-style', content: 'black' }]
  ],
  themeConfig: {
    repo: 'yi-nology/git-manage-service',
    repoLabel: 'GitHub',
    docsDir: 'docs',
    docsBranch: 'main',
    editLinks: true,
    editLinkText: '在 GitHub 上编辑此页',
    lastUpdated: '上次更新',
    nav: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/getting-started' },
      {
        text: '功能指南',
        items: [
          { text: '仓库管理', link: '/features/repo' },
          { text: '分支管理', link: '/features/branch' },
          { text: '同步任务', link: '/features/sync' },
          { text: 'Webhook', link: '/features/webhook' },
          { text: '通知配置', link: '/features/notification' },
          { text: 'SSH 密钥', link: '/features/ssh' },
          { text: '代码度量', link: '/features/metrics' },
          { text: 'Spec 编辑器', link: '/features/spec-editor' },
          { text: 'Patch 管理', link: '/features/patch-manager' }
        ]
      },
      {
        text: '部署方案',
        items: [
          { text: '二进制部署', link: '/deployment/binary' },
          { text: 'Docker 部署', link: '/deployment/docker' },
          { text: 'Kubernetes', link: '/deployment/kubernetes' }
        ]
      },
      { text: '配置参考', link: '/configuration' },
      { text: 'API 文档', link: '/api' }
    ],
    sidebar: {
      '/': [
        {
          title: '📖 概览',
          collapsable: false,
          children: [
            '',
            'getting-started'
          ]
        },
        {
          title: '🚀 功能指南',
          collapsable: true,
          children: [
            'features/repo',
            'features/branch',
            'features/sync',
            'features/webhook',
            'features/notification',
            'features/ssh',
            'features/metrics',
            'features/spec-editor',
            'features/patch-manager'
          ]
        },
        {
          title: '📦 部署方案',
          collapsable: true,
          children: [
            'deployment/binary',
            'deployment/docker',
            'deployment/kubernetes'
          ]
        },
        {
          title: '⚙️ 参考',
          collapsable: true,
          children: [
            'configuration',
            'api'
          ]
        }
      ],
      '/features/': [
        {
          title: '功能指南',
          collapsable: false,
          children: [
            'repo',
            'branch',
            'sync',
            'webhook',
            'notification',
            'ssh',
            'metrics',
            'spec-editor',
            'patch-manager'
          ]
        }
      ],
      '/deployment/': [
        {
          title: '部署方案',
          collapsable: false,
          children: [
            'binary',
            'docker',
            'kubernetes'
          ]
        }
      ],
      '/dev-notes/': [
        {
          title: '开发笔记',
          collapsable: true,
          children: [
            '',
            'AI_API',
            'AI_FOLLOWUP_REVIEW',
            'AI_IMPROVEMENT_CHECKLIST',
            'AI_INTEGRATION_DESIGN',
            'AI_REMAINING_ISSUES',
            'AI_REVIEW_SUMMARY_FOLLOWUP',
            'API_NAMING_CONVENTION',
            'BROWSER_TEST_REPORT',
            'BUGFIX_ISDIR_FIELD',
            'BUGFIX_TREE_STRUCTURE',
            'DESKTOP_BUILD',
            'DESKTOP_STATUS',
            'DEV_SERVER_STATUS',
            'E2E_TEST_REPORT',
            'FRONTEND_TYPE_GENERATION_PLAN',
            'GIT_SYNC_INTEGRATION',
            'OPTIMIZATION_PLAN',
            'OPTIMIZATION_REPORT',
            'PROJECT_REVIEW_AND_IMPROVEMENT_PLAN',
            'PROJECT_REVIEW_FOLLOWUP_PLAN',
            'PROJECT_STRUCTURE',
            'REFACTORING_PLAN',
            'SPEC_EDITOR_AUTO_GUIDE',
            'SPEC_EDITOR_INIT_FEATURE',
            'SPEC_EDITOR_PLAN',
            'SPEC_EDITOR_PROGRESS_REPORT',
            'SPEC_EDITOR_TEST_REPORT',
            'SPEC_EDITOR_TEST_REPORT_FINAL',
            'SSH_TROUBLESHOOTING',
            'SYNC_REDESIGN_PLAN',
            'SYSTEM_IMPROVEMENT_ISSUES',
            'TEST_REPORT',
            'TODO-Git管理服务整体优化-20260308',
            'WAILS_GUIDE'
          ]
        }
      ]
    },
    search: true,
    searchMaxSuggestions: 10,
    smoothScroll: true
  },
  markdown: {
    lineNumbers: true
  },
  plugins: []
}
