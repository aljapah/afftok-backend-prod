# AffTok Developer Documentation

Welcome to the AffTok Developer Portal. This documentation provides everything you need to integrate with the AffTok affiliate tracking platform.

## What is AffTok?

AffTok is an enterprise-grade affiliate tracking platform that provides:

- **Real-time Click Tracking** - Track every click with millisecond precision
- **Conversion Attribution** - Accurate attribution with multi-touch support
- **Smart Routing** - Intelligent traffic distribution based on geo, device, and performance
- **Fraud Prevention** - Advanced bot detection and click validation
- **Multi-Tenant Architecture** - Complete isolation for enterprise deployments
- **Zero-Drop Tracking** - Never lose a click or conversion, even during outages
- **Two-Sided Marketplace** - Connect promoters with advertisers
- **AI Assistant** - Intelligent guidance for promoters
- **Teams & Contests** - Social features and gamification

## Platform Overview

### For Promoters (المروّجين)
- Browse and join offers
- Get unique tracking links
- Track clicks and conversions
- AI-powered performance analysis
- Join teams and compete in contests
- Earn points and climb leaderboards

### For Advertisers (المعلنين)
- Register company account
- Create and manage offers
- Track promoter performance
- Real-time analytics dashboard

### For Developers
- RESTful API with full documentation
- SDKs for all major platforms
- Webhook integrations
- Comprehensive testing tools

## Documentation Sections

| Section | Description |
|---------|-------------|
| [Quick Start](./quickstart/README.md) | Get up and running in 5 minutes |
| [API Reference](./api/README.md) | Complete REST API documentation |
| [SDKs](./sdk/README.md) | Mobile and Web SDK guides |
| [Features](./features/README.md) | AI Assistant, Teams, Advertisers |
| [Webhooks](./webhooks/README.md) | Real-time event notifications |
| [Security](./security/README.md) | Authentication and best practices |
| [Integration Guides](./guides/README.md) | Step-by-step integration tutorials |
| [Testing](./testing/README.md) | Testing tools and QA guides |
| [Operations](./operations/README.md) | Operational guidelines |

## New Features (December 2024)

### 🤖 AI Assistant
Intelligent guided assistant for promoters:
- Personal statistics dashboard
- Smart offer suggestions
- Performance analysis with recommendations
- Success tips and strategies
- Points & levels system
- BYOK (Bring Your Own Key) for advanced chat

[Learn more →](./features/ai-assistant.md)

### 👥 Teams & Contests
Social and gamification features:
- Create and join teams
- Global leaderboard rankings
- Competitive contests with prizes
- Team-based challenges

[Learn more →](./features/teams-contests.md)

### 🏢 Advertisers Portal
Two-sided marketplace:
- Advertiser registration and onboarding
- Offer creation with approval workflow
- Performance analytics
- Promoter management

[Learn more →](./features/advertisers.md)

## Quick Links

- [Generate API Key](./api/authentication.md)
- [Track Your First Click](./quickstart/README.md)
- [Send a Postback](./api/postbacks.md)
- [SDK Installation](./sdk/README.md)
- [AI Assistant Guide](./features/ai-assistant.md)
- [Teams & Contests](./features/teams-contests.md)

## Business Model

AffTok is **100% free** for both promoters and advertisers:
- No subscription fees
- No transaction fees
- Platform focuses on **tracking only**
- Companies pay promoters directly
- AffTok earns commission from companies

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      AffTok Platform                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │   Mobile    │    │   Backend   │    │   Admin     │     │
│  │   App       │◄──►│   (Go)      │◄──►│   Panel     │     │
│  │  (Flutter)  │    │             │    │  (React)    │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│                            │                                │
│                     ┌──────┴──────┐                        │
│                     │             │                         │
│               ┌─────▼─────┐ ┌─────▼─────┐                  │
│               │ PostgreSQL│ │   Redis   │                  │
│               │    DB     │ │   Cache   │                  │
│               └───────────┘ └───────────┘                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## System Status

Check the current system status at [status.afftokapp.com](https://status.afftokapp.com)

## Support

- **Email**: support@afftokapp.com
- **Website**: [afftokapp.com](https://afftokapp.com)
- **Twitter**: [@afftokapp](https://twitter.com/afftokapp)
- **Instagram**: [@afftokapp](https://instagram.com/afftokapp)

---

## Version

- **API Version**: v1.0
- **SDK Version**: 1.0.0
- **Documentation Version**: 2.0.0
- **Last Updated**: December 2024

## Changelog

### v2.0.0 (December 2024)
- Added AI Assistant documentation
- Added Teams & Contests documentation
- Added Advertisers Portal documentation
- Updated platform overview
- Added business model section
- Updated architecture diagram
