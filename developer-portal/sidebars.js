/**
 * AffTok Developer Portal Sidebar Configuration
 */

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docsSidebar: [
    'index',
    {
      type: 'category',
      label: 'Overview',
      items: [
        'overview/architecture',
        'overview/tracking-flow',
        'overview/smart-routing',
      ],
    },
    {
      type: 'category',
      label: 'Quick Start',
      items: [
        'quickstart/README',
        'quickstart/signed-links',
      ],
    },
    {
      type: 'category',
      label: 'API Reference',
      items: [
        'api/README',
        'api/authentication',
        'api/offers',
        'api/stats',
        'api/postbacks',
        'api/geo-rules',
        'api/webhooks',
        'api/tenants',
        'api/admin',
        'api/errors',
      ],
    },
    {
      type: 'category',
      label: 'SDKs',
      items: [
        'sdk/README',
        'sdk/android',
        'sdk/ios',
        'sdk/flutter',
        'sdk/react-native',
        'sdk/web',
      ],
    },
    {
      type: 'category',
      label: 'Features',
      items: [
        'features/README',
        'features/ai-assistant',
        'features/teams-contests',
        'features/advertisers',
      ],
    },
    {
      type: 'category',
      label: 'Webhooks',
      items: [
        'webhooks/README',
      ],
    },
    {
      type: 'category',
      label: 'Security',
      items: [
        'security/README',
      ],
    },
    {
      type: 'category',
      label: 'Integration Guides',
      items: [
        'guides/README',
      ],
    },
    {
      type: 'category',
      label: 'Testing & QA',
      items: [
        'testing/README',
      ],
    },
    {
      type: 'category',
      label: 'Operations',
      items: [
        'operations/README',
      ],
    },
  ],
};

module.exports = sidebars;

