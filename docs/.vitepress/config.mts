import { defineConfig } from 'vitepress'

const repo = 'https://github.com/777genius/proxykit'
const docsBase = process.env.GITHUB_ACTIONS === 'true' ? '/proxykit/' : '/'

export default defineConfig({
  base: docsBase,
  title: 'proxykit',
  description: 'Composable Go proxy foundation for reverse, forward, CONNECT, WebSocket, and runtime-aware proxy workflows.',
  cleanUrls: true,
  ignoreDeadLinks: true,
  lastUpdated: true,
  head: [
    ['meta', { name: 'theme-color', content: '#0b1016' }],
    ['meta', { property: 'og:title', content: 'proxykit' }],
    ['meta', { property: 'og:description', content: 'Composable Go proxy foundation for reverse, forward, CONNECT, WebSocket, and runtime-aware proxy workflows.' }]
  ],
  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Cookbook', link: '/guide/cookbook' },
      { text: 'Reference', link: '/reference/reverse' },
      { text: 'Compatibility', link: '/guide/compatibility' },
      { text: 'Comparisons', link: '/guide/comparisons' },
      { text: 'FAQ', link: '/guide/faq' }
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Use Cases', link: '/guide/use-cases' },
            { text: 'Package Matrix', link: '/guide/package-matrix' },
            { text: 'Architecture', link: '/guide/architecture' },
            { text: 'Packages', link: '/guide/packages' },
            { text: 'Cookbook', link: '/guide/cookbook' },
            { text: 'Observation Hooks', link: '/guide/observation-hooks' },
            { text: 'Compatibility and Versioning', link: '/guide/compatibility' },
            { text: 'Migration', link: '/guide/migration' },
            { text: 'Limits and Non-Goals', link: '/guide/limits' },
            { text: 'Comparisons', link: '/guide/comparisons' },
            { text: 'FAQ', link: '/guide/faq' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'Core Transports',
          items: [
            { text: 'reverse', link: '/reference/reverse' },
            { text: 'forward', link: '/reference/forward' },
            { text: 'connect', link: '/reference/connect' },
            { text: 'wsproxy', link: '/reference/wsproxy' },
            { text: 'proxyruntime', link: '/reference/proxyruntime' }
          ]
        },
        {
          text: 'Shared Building Blocks',
          items: [
            { text: 'observe', link: '/reference/observe' },
            { text: 'cookies', link: '/reference/cookies' },
            { text: 'Utilities', link: '/reference/utilities' }
          ]
        }
      ]
    },
    socialLinks: [
      { icon: 'github', link: repo }
    ],
    search: {
      provider: 'local'
    },
    outline: {
      level: [2, 3]
    },
    editLink: {
      pattern: `${repo}/edit/main/docs/:path`
    },
    footer: {
      message: 'Released under the Apache 2.0 License.',
      copyright: 'Copyright 2025 Belief'
    }
  }
})
