import { defineConfig } from 'vitepress'
import { mermaidSkeletonPlugin } from './mermaid-skeleton'

// HAR Skills 文档站配置 —— 中英双语，基于 VitePress 1.x
// 视觉系统：深海军蓝 + 能力树六色，等宽显示字体呼应 HAR 的 JSON 文本本质

const ZH_SIDEBAR = [
  {
    text: '开始',
    collapsed: false,
    items: [
      { text: '概览', link: '/zh/' },
      { text: '快速开始', link: '/zh/quick-start' },
      { text: '安装', link: '/zh/install' },
      { text: 'HAR 格式入门', link: '/zh/har-basics' }
    ]
  },
  {
    text: '接入方式',
    collapsed: false,
    items: [
      { text: 'AI Agent Skill', link: '/zh/access/skill' },
      { text: 'CLI 命令行', link: '/zh/access/cli' },
      { text: 'Go SDK', link: '/zh/access/sdk' },
      { text: 'MCP 封装', link: '/zh/access/mcp' }
    ]
  },
  {
    text: 'CLI 命令参考',
    collapsed: true,
    items: [
      { text: '全局参数', link: '/zh/cli/global-flags' },
      { text: '基础操作', link: '/zh/cli/basic' },
      { text: '文件操作', link: '/zh/cli/files' },
      { text: '安全与隐私', link: '/zh/cli/security' },
      { text: '深度分析', link: '/zh/cli/analysis' },
      { text: '转换与导出', link: '/zh/cli/transform' }
    ]
  },
  {
    text: 'SDK 指南',
    collapsed: true,
    items: [
      { text: '数据结构', link: '/zh/sdk/data-structures' },
      { text: '四种解析策略', link: '/zh/sdk/parsing-strategies' },
      { text: 'Provider 接口', link: '/zh/sdk/providers' },
      { text: '函数式选项', link: '/zh/sdk/functional-options' },
      { text: '过滤与链式结果', link: '/zh/sdk/filtering' },
      { text: '转换与脱敏', link: '/zh/sdk/transform' },
      { text: '导出能力', link: '/zh/sdk/export' },
      { text: '差异·合并·拆分', link: '/zh/sdk/diff-merge-split' },
      { text: 'API 速查', link: '/zh/sdk/api-reference' },
      { text: '请求录制归档', link: '/zh/sdk/recording-requests' },
    ]
  },
  {
    text: '实现原理',
    collapsed: true,
    items: [
      { text: '内存优化原理', link: '/zh/internals/memory-optimized' },
      { text: '懒加载原理', link: '/zh/internals/lazy-loading' },
      { text: '流式解析原理', link: '/zh/internals/streaming' },
      { text: '宽松解析与错误体系', link: '/zh/internals/lenient-parsing' },
      { text: '扩展字段保真', link: '/zh/internals/custom-fields' },
      { text: 'Waterfall 分层算法', link: '/zh/internals/waterfall' }
    ]
  },
  {
    text: '示例与工作流',
    collapsed: true,
    items: [
      { text: '安全审计工作流', link: '/zh/workflows/security-audit' },
      { text: '性能优化工作流', link: '/zh/workflows/performance' },
      { text: 'API 迁移测试', link: '/zh/workflows/api-migration' },
      { text: '数据清洗与分享', link: '/zh/workflows/data-cleaning' },
      { text: '示例代码集', link: '/zh/examples/' }
    ]
  },
  {
    text: '贡献',
    collapsed: true,
    items: [
      { text: '架构总览', link: '/zh/contributing/architecture' },
      { text: '贡献指南', link: '/zh/contributing/' }
    ]
  }
]

const EN_SIDEBAR = [
  {
    text: 'Getting Started',
    collapsed: false,
    items: [
      { text: 'Overview', link: '/en/' },
      { text: 'Quick Start', link: '/en/quick-start' },
      { text: 'Installation', link: '/en/install' },
      { text: 'HAR Format Primer', link: '/en/har-basics' }
    ]
  },
  {
    text: 'Access Methods',
    collapsed: false,
    items: [
      { text: 'AI Agent Skill', link: '/en/access/skill' },
      { text: 'CLI', link: '/en/access/cli' },
      { text: 'Go SDK', link: '/en/access/sdk' },
      { text: 'MCP Wrapper', link: '/en/access/mcp' }
    ]
  },
  {
    text: 'CLI Reference',
    collapsed: true,
    items: [
      { text: 'Global Flags', link: '/en/cli/global-flags' },
      { text: 'Basic Operations', link: '/en/cli/basic' },
      { text: 'File Operations', link: '/en/cli/files' },
      { text: 'Security & Privacy', link: '/en/cli/security' },
      { text: 'Deep Analysis', link: '/en/cli/analysis' },
      { text: 'Transform & Export', link: '/en/cli/transform' }
    ]
  },
  {
    text: 'SDK Guide',
    collapsed: true,
    items: [
      { text: 'Data Structures', link: '/en/sdk/data-structures' },
      { text: 'Parsing Strategies', link: '/en/sdk/parsing-strategies' },
      { text: 'Provider Interfaces', link: '/en/sdk/providers' },
      { text: 'Functional Options', link: '/en/sdk/functional-options' },
      { text: 'Filtering & Chaining', link: '/en/sdk/filtering' },
      { text: 'Transform & Redact', link: '/en/sdk/transform' },
      { text: 'Export', link: '/en/sdk/export' },
      { text: 'Diff · Merge · Split', link: '/en/sdk/diff-merge-split' },
      { text: 'API Reference', link: '/en/sdk/api-reference' },
      { text: 'Recording Requests', link: '/en/sdk/recording-requests' },
    ]
  },
  {
    text: 'Internals',
    collapsed: true,
    items: [
      { text: 'Memory Optimization', link: '/en/internals/memory-optimized' },
      { text: 'Lazy Loading', link: '/en/internals/lazy-loading' },
      { text: 'Streaming Parsing', link: '/en/internals/streaming' },
      { text: 'Lenient Parsing & Errors', link: '/en/internals/lenient-parsing' },
      { text: 'Custom Field Fidelity', link: '/en/internals/custom-fields' },
      { text: 'Waterfall Layering', link: '/en/internals/waterfall' }
    ]
  },
  {
    text: 'Workflows & Examples',
    collapsed: true,
    items: [
      { text: 'Security Audit', link: '/en/workflows/security-audit' },
      { text: 'Performance Tuning', link: '/en/workflows/performance' },
      { text: 'API Migration Testing', link: '/en/workflows/api-migration' },
      { text: 'Data Cleaning & Sharing', link: '/en/workflows/data-cleaning' },
      { text: 'Examples', link: '/en/examples/' }
    ]
  },
  {
    text: 'Contributing',
    collapsed: true,
    items: [
      { text: 'Architecture', link: '/en/contributing/architecture' },
      { text: 'Contributing Guide', link: '/en/contributing/' }
    ]
  }
]

export default defineConfig({
  lang: 'zh-CN',
  title: 'HAR Skills',
  description: 'AI 原生的 HAR 文件分析工具箱 · AI-native HAR analysis toolkit',
  lastUpdated: true,
  cleanDist: true,
  srcDir: '.',
  outDir: '.vitepress/dist',

  // markdown-it 层注册骨架屏插件：把空 <div class="mermaid"> 换成带占位的容器，
  // 消除 mermaid 客户端懒渲染导致的首屏空白
  markdown: {
    config: (md) => {
      md.use(mermaidSkeletonPlugin)
    }
  },

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#0b1220' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    [
      'link',
      {
        rel: 'stylesheet',
        href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;700&display=swap'
      }
    ]
  ],

  locales: {
    root: {
      label: '简体中文',
      lang: 'zh-CN',
      themeConfig: {
        nav: [
          { text: '概览', link: '/zh/' },
          { text: '快速开始', link: '/zh/quick-start' },
          { text: 'CLI', link: '/zh/cli/global-flags' },
          { text: 'SDK', link: '/zh/sdk/data-structures' },
          { text: '原理', link: '/zh/internals/memory-optimized' },
          { text: '示例', link: '/zh/examples/' },
          {
            text: 'GitHub',
            link: 'https://github.com/cyberspacesec/har-skills'
          }
        ],
        sidebar: ZH_SIDEBAR,
        docFooter: { prev: '上一页', next: '下一页' },
        outline: { label: '本页导航', level: [2, 3] },
        lastUpdatedText: '最后更新',
        returnToTopLabel: '回到顶部',
        sidebarTitle: '目录',
        editLink: {
          text: '在 GitHub 上编辑此页',
          link: 'https://github.com/cyberspacesec/har-skills/edit/main/website'
        },
        search: { provider: 'local', options: { translations: { button: { buttonText: '搜索文档', buttonAriaLabel: '搜索' } } } }
      }
    },
    en: {
      label: 'English',
      lang: 'en-US',
      themeConfig: {
        nav: [
          { text: 'Overview', link: '/en/' },
          { text: 'Quick Start', link: '/en/quick-start' },
          { text: 'CLI', link: '/en/cli/global-flags' },
          { text: 'SDK', link: '/en/sdk/data-structures' },
          { text: 'Internals', link: '/en/internals/memory-optimized' },
          { text: 'Examples', link: '/en/examples/' },
          {
            text: 'GitHub',
            link: 'https://github.com/cyberspacesec/har-skills'
          }
        ],
        sidebar: EN_SIDEBAR,
        outline: { label: 'On this page', level: [2, 3] },
        lastUpdatedText: 'Last updated',
        editLink: {
          text: 'Edit this page on GitHub',
          link: 'https://github.com/cyberspacesec/har-skills/edit/main/website'
        },
        search: { provider: 'local' }
      }
    }
  },

  themeConfig: {
    logo: '/favicon.svg',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/cyberspacesec/har-skills' }
    ],
    footer: {
      message: '基于 MIT 协议发布 · Released under the MIT License',
      copyright: 'Copyright © 2024-present CyberspaceSec'
    }
  }
})
