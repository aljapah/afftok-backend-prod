// @ts-check
// Docusaurus configuration for AffTok Developer Portal

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'AffTok Developer Portal',
  tagline: 'Enterprise-grade affiliate tracking platform',
  favicon: 'img/favicon.ico',

  url: 'https://docs.afftok.com',
  baseUrl: '/',

  organizationName: 'afftok',
  projectName: 'developer-portal',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'ar'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/afftok/developer-portal/tree/main/',
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: 'img/afftok-social-card.jpg',
      navbar: {
        title: 'AffTok',
        logo: {
          alt: 'AffTok Logo',
          src: 'img/logo.svg',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docsSidebar',
            position: 'left',
            label: 'Documentation',
          },
          {
            to: '/docs/api',
            label: 'API Reference',
            position: 'left',
          },
          {
            to: '/docs/sdk',
            label: 'SDKs',
            position: 'left',
          },
          {
            href: 'https://github.com/afftok',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              { label: 'Quick Start', to: '/docs/quickstart' },
              { label: 'API Reference', to: '/docs/api' },
              { label: 'SDKs', to: '/docs/sdk' },
            ],
          },
          {
            title: 'Resources',
            items: [
              { label: 'Security', to: '/docs/security' },
              { label: 'Webhooks', to: '/docs/webhooks' },
              { label: 'Testing', to: '/docs/testing' },
            ],
          },
          {
            title: 'Features',
            items: [
              { label: 'AI Assistant', to: '/docs/features/ai-assistant' },
              { label: 'Teams & Contests', to: '/docs/features/teams-contests' },
              { label: 'Advertisers', to: '/docs/features/advertisers' },
            ],
          },
          {
            title: 'Community',
            items: [
              { label: 'Website', href: 'https://afftokapp.com' },
              { label: 'Twitter', href: 'https://twitter.com/afftokapp' },
              { label: 'Instagram', href: 'https://instagram.com/afftokapp' },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} AffTok. All rights reserved.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['php', 'kotlin', 'swift', 'dart', 'go', 'bash', 'json'],
      },
      algolia: {
        appId: 'YOUR_APP_ID',
        apiKey: 'YOUR_SEARCH_API_KEY',
        indexName: 'afftok',
        contextualSearch: true,
      },
    }),

  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'api',
        path: 'docs/api',
        routeBasePath: 'api',
        sidebarPath: require.resolve('./sidebarsApi.js'),
      },
    ],
  ],
};

module.exports = config;

