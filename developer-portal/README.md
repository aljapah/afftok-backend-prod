# AffTok Developer Portal

Production-grade developer documentation for the AffTok affiliate tracking platform.

## Overview

This documentation portal covers all aspects of integrating with AffTok:

- **Quick Start** - Get up and running in minutes
- **API Reference** - Complete REST API documentation
- **SDK Documentation** - Android, iOS, Flutter, React Native, Web
- **Features** - AI Assistant, Teams & Contests, Advertisers Portal
- **Webhooks** - Real-time event notifications
- **Security** - Best practices and security guidelines
- **Integration Guides** - Step-by-step integration tutorials
- **Testing & QA** - Testing tools and procedures
- **Operations** - Operational guidelines and best practices

## New in v2.0 (December 2024)

- 🤖 **AI Assistant** - Intelligent guided assistant for promoters
- 👥 **Teams & Contests** - Social features and gamification
- 🏢 **Advertisers Portal** - Two-sided marketplace support
- 📱 **Enhanced Mobile App** - Beautiful UI with RTL support

## Documentation Structure

```
docs/
├── index.md                    # Welcome page
├── overview/
│   ├── architecture.md         # System architecture
│   ├── tracking-flow.md        # Click-to-conversion flow
│   └── smart-routing.md        # Smart routing capabilities
├── quickstart/
│   ├── README.md               # Quick start guide
│   └── signed-links.md         # Signed links documentation
├── api/
│   ├── README.md               # API overview
│   ├── authentication.md       # Auth methods
│   ├── offers.md               # Offers API
│   ├── stats.md                # Stats API
│   ├── postbacks.md            # Postback API
│   ├── geo-rules.md            # Geo Rules API
│   ├── webhooks.md             # Webhooks API
│   ├── tenants.md              # Tenants API
│   ├── admin.md                # Admin API
│   └── errors.md               # Error codes
├── sdk/
│   ├── README.md               # SDK overview
│   ├── android.md              # Android SDK
│   ├── ios.md                  # iOS SDK
│   ├── flutter.md              # Flutter SDK
│   ├── react-native.md         # React Native SDK
│   └── web.md                  # Web SDK
├── features/                   # NEW
│   ├── README.md               # Features overview
│   ├── ai-assistant.md         # AI Assistant documentation
│   ├── teams-contests.md       # Teams & Contests
│   └── advertisers.md          # Advertisers Portal
├── webhooks/
│   └── README.md               # Webhook documentation
├── security/
│   └── README.md               # Security documentation
├── guides/
│   └── README.md               # Integration guides
├── testing/
│   └── README.md               # Testing documentation
└── operations/
    └── README.md               # Operational guidelines
```

## Running Locally

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
cd developer-portal
npm install
```

### Development

```bash
npm start
```

This starts a local development server at `http://localhost:3000`.

### Build

```bash
npm run build
```

This generates static content into the `build` directory.

### Deploy

```bash
npm run deploy
```

## Built With

- [Docusaurus 3](https://docusaurus.io/) - Documentation framework
- [React](https://reactjs.org/) - UI library
- [MDX](https://mdxjs.com/) - Markdown with JSX

## Documentation Standards

### Code Examples

All code examples should:
- Be complete and runnable
- Include error handling
- Use real endpoint URLs
- Show expected responses

### API Documentation

Each API endpoint should include:
- HTTP method and path
- Authentication requirements
- Request parameters
- Response format
- Code examples in multiple languages

### SDK Documentation

Each SDK should include:
- Installation instructions
- Initialization
- All available methods
- Error handling
- Complete examples

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

Copyright © 2024 AffTok. All rights reserved.
